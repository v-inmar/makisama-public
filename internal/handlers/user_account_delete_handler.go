package handlers

import (
	"app/internal/constants"
	authid_models "app/internal/models/authid"
	board_models "app/internal/models/board"
	email_ban_models "app/internal/models/email_ban"
	feedback_models "app/internal/models/feedback"
	image_models "app/internal/models/image"
	user_models "app/internal/models/user"
	image_utils "app/internal/utilities/image"
	sessions_utils "app/internal/utilities/sessions"
	template_utils "app/internal/utilities/templates"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func (repo *HandlerRepo) GetUserAccountDelete(w http.ResponseWriter, r *http.Request) {
	td := r.Context().Value(constants.TemplateDataContextKey).(*template_utils.TemplateData)
	sessionService := sessions_utils.NewSessionService(r, repo.Session)

	currentUser, err := sessionService.GetUserFromSession(repo.DB)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) && !errors.Is(err, sessions_utils.ErrNoUserInSession) {
			repo.Logger.Errorln(err)
			template_utils.Abort(
				w,
				repo.Templates["error.html"],
				http.StatusInternalServerError,
				constants.Error500Title,
				constants.Error500Message,
			)

			return
		}
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	emailBanDbService := email_ban_models.NewEmailBanDbService(repo.DB)
	ban, err := emailBanDbService.ReadByEmail(currentUser.Email)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			repo.Logger.Errorln(err)
			template_utils.Abort(
				w,
				repo.Templates["error.html"],
				http.StatusInternalServerError,
				constants.Error500Title,
				constants.Error500Message,
			)
			return
		}
	}
	if ban != nil {
		sessionService.DestroySession()
		template_utils.Abort(
			w,
			repo.Templates["error.html"],
			http.StatusForbidden,
			constants.Error403Title,
			constants.Error403Message,
		)
		return
	}

	td.Data["ErrorPassword"] = ""
	template_utils.RenderTemplate(w, repo.Templates["delete.html"], td)
}

func (repo *HandlerRepo) PostUserAccountDelete(w http.ResponseWriter, r *http.Request) {
	td := r.Context().Value(constants.TemplateDataContextKey).(*template_utils.TemplateData)
	sessionService := sessions_utils.NewSessionService(r, repo.Session)

	currentUser, err := sessionService.GetUserFromSession(repo.DB)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) && !errors.Is(err, sessions_utils.ErrNoUserInSession) {
			repo.Logger.Errorln(err)
			template_utils.Abort(
				w,
				repo.Templates["error.html"],
				http.StatusInternalServerError,
				constants.Error500Title,
				constants.Error500Message,
			)

			return
		}
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	emailBanDbService := email_ban_models.NewEmailBanDbService(repo.DB)
	ban, err := emailBanDbService.ReadByEmail(currentUser.Email)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			repo.Logger.Errorln(err)
			template_utils.Abort(
				w,
				repo.Templates["error.html"],
				http.StatusInternalServerError,
				constants.Error500Title,
				constants.Error500Message,
			)
			return
		}
	}
	if ban != nil {
		sessionService.DestroySession()
		template_utils.Abort(
			w,
			repo.Templates["error.html"],
			http.StatusForbidden,
			constants.Error403Title,
			constants.Error403Message,
		)
		return
	}

	if err := r.ParseForm(); err != nil {
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusBadRequest, constants.Error400Title, constants.Error400Message)
		return
	}

	password := r.FormValue("password")
	if err := bcrypt.CompareHashAndPassword([]byte(currentUser.Password), []byte(password)); err != nil {
		td.Data["ErrorPassword"] = "Password not recognized"
		template_utils.RenderTemplate(w, repo.Templates["delete.html"], td)
		return
	}

	// get all boards of the current user
	boardDbService := board_models.NewBoardDbService(repo.DB)
	boards, err := boardDbService.ReadAllByOwnerId(currentUser.ID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			repo.Logger.Errorln(err)
			template_utils.Abort(
				w,
				repo.Templates["error.html"],
				http.StatusInternalServerError,
				constants.Error500Title,
				constants.Error500Message,
			)
			return
		}
	}

	// get all feedbacks of the current user
	feedbackDbService := feedback_models.NewFeedbackDbService(repo.DB)
	feedbacks, err := feedbackDbService.ReadAllByOwnerId(currentUser.ID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			repo.Logger.Errorln(err)
			template_utils.Abort(
				w,
				repo.Templates["error.html"],
				http.StatusInternalServerError,
				constants.Error500Title,
				constants.Error500Message,
			)
			return
		}
	}

	boardImageDbService := image_models.NewBoardImageDbService(repo.DB)
	imageDbService := image_models.NewImageDbService(repo.DB)
	authidDbService := authid_models.NewAuthidDbService(repo.DB)
	userDbService := user_models.NewUserDbService(repo.DB)
	filesToDelete := make([]string, 0)
	errDel := func() error {
		tx, err := repo.DB.Beginx()
		if err != nil {
			return err
		}
		defer tx.Rollback()

		// anonymize feedbacks
		if len(feedbacks) > 0 {
			for _, f := range feedbacks {
				if err := feedbackDbService.UpdateOwnerIdWithTx(tx, &f, int64(0)); err != nil {
					return err
				}

				if err := feedbackDbService.UpdateEmailWithTx(tx, &f, ""); err != nil {
					return err
				}
			}
		}

		// delete all user boards and its feedbacks
		// feedbacks are soft delete
		if len(boards) > 0 {
			for _, b := range boards {

				// get all feedback associated with the board
				fs, err := feedbackDbService.ReadAllByBoardIdWithTx(tx, b.ID)
				if err != nil {
					if !errors.Is(err, sql.ErrNoRows) {
						return err
					}
				}

				// soft delete all feedbacks that are member of the board
				if len(fs) > 0 {
					for _, f := range fs {
						if err := feedbackDbService.DeleteFeedbackWithTx(tx, &f); err != nil {
							continue
						}
					}
				}

				// delete board image
				bi, _ := boardImageDbService.ReadByBoardIdWithTx(tx, b.ID)
				if bi != nil {
					img, _ := imageDbService.ReadByIdWithTx(tx, bi.ImageId)
					if img != nil {
						filenameWithExt := fmt.Sprintf("%s.%s", img.Filename, img.Extension)
						now := time.Now().UTC()
						ttl := now.AddDate(0, 0, 365)
						imageDbService.UpdateTTLWithTx(tx, img, &ttl)

						// grab the filename+ext
						filesToDelete = append(filesToDelete, filenameWithExt)
					}

					boardImageDbService.DeleteWithTx(tx, bi)
				}

				// delete the board
				if err := boardDbService.DeleteBoardWithTx(tx, &b); err != nil {
					return err
				}

			}
		}

		// get current user authid and assign ttl
		authid, err := authidDbService.ReadByIdWithTx(tx, currentUser.AuthidId)
		if err != nil {
			return err
		}

		now := time.Now().UTC()
		ttl := now.AddDate(0, 0, 365)
		if err := authidDbService.UpdateTTLWithTx(tx, authid, &ttl); err != nil {
			return err
		}

		// delete user
		if err := userDbService.DeleteUserWithTx(tx, currentUser); err != nil {

			return err
		}

		// commit
		if err := tx.Commit(); err != nil {
			return err
		}

		return nil
	}()

	if errDel != nil {
		repo.Logger.Errorln(errDel)
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
		return
	}

	// delete the actual images
	// async
	if len(filesToDelete) > 0 {
		for _, f := range filesToDelete {
			go func() {
				image_utils.DeleteImageFile(filepath.Join(constants.BoardUploadImageDirectory, f))
			}()
		}
	}

	sessionService.DestroySession()
	sessionService.ToFlash("Account and Boards deleted")
	http.Redirect(w, r, "/", http.StatusSeeOther)

}
