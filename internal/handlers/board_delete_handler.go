package handlers

import (
	"app/internal/constants"
	board_models "app/internal/models/board"
	email_ban_models "app/internal/models/email_ban"
	image_models "app/internal/models/image"
	sessions_utils "app/internal/utilities/sessions"
	template_utils "app/internal/utilities/templates"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func (repo *HandlerRepo) PostBoardDeleteBoard(w http.ResponseWriter, r *http.Request) {
	// td := r.Context().Value(constants.TemplateDataContextKey).(*template_utils.TemplateData)

	sessionService := sessions_utils.NewSessionService(r, repo.Session)
	currentUser, err := sessionService.GetUserFromSession(repo.DB)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, sessions_utils.ErrNoUserInSession) {
			http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
			return
		}
		repo.Logger.Errorln(err)
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
		return
	}

	emailBanDbService := email_ban_models.NewEmailBanDbService(repo.DB)
	ban, err := emailBanDbService.ReadByEmail(currentUser.Email)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			repo.Logger.Errorln(err)
			template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
			return
		}
	}
	if ban != nil {
		sessionService.DestroySession()
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusForbidden, constants.Error403Title, constants.Error403Message)
		return
	}

	boardDbService := board_models.NewBoardDbService(repo.DB)
	slug := r.PathValue("slug")
	board, err := boardDbService.ReadBySlug(slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			template_utils.Abort(w, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
			return

		}

		repo.Logger.Errorln(err)
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
		return
	}

	if board.OwnerID != currentUser.ID {
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
		return
	}

	boardImgDbService := image_models.NewBoardImageDbService(repo.DB)
	if err := func() error {
		tx, err := repo.DB.Beginx()
		if err != nil {
			return err
		}
		defer tx.Rollback()

		// get image if it exist to be deleted
		bim, err := boardImgDbService.ReadByBoardIdWithTx(tx, board.ID)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return err
			}
		}

		var imgFileToDelete string

		if bim != nil {
			imgDbService := image_models.NewImageDbService(repo.DB)
			img, err := imgDbService.ReadByIdWithTx(tx, bim.ImageId)
			if err != nil {
				return err
			}

			// assign ttl
			now := time.Now().UTC()
			ttl := now.AddDate(0, 0, 365)
			if err := imgDbService.UpdateTTLWithTx(tx, img, &ttl); err != nil {
				return err
			}

			// delete the bim
			if err := boardImgDbService.DeleteWithTx(tx, bim); err != nil {
				return err
			}

			// get the image full filename + ext
			// so it can be deleted later on
			// because deleting this now and then after this some db error
			// this file remains deleted while db actions rolled back
			imgFileToDelete = fmt.Sprintf("%s.%s", img.Filename, img.Extension)
		}

		if err := boardDbService.DeleteWithTx(tx, board); err != nil {
			return err
		}

		// now delete the image file
		// because no other db actions can fail from this point on
		// apart from the commit
		if imgFileToDelete != "" {
			if err := os.Remove(filepath.Join(constants.BoardUploadImageDirectory, imgFileToDelete)); err != nil {
				return err
			}
		}

		if err := tx.Commit(); err != nil {
			return err
		}

		return nil
	}(); err != nil {
		repo.Logger.Errorln(err)
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
		return
	}

	sessionService.ToFlash("Board deleted")
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}
