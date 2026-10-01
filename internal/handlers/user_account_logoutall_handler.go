package handlers

import (
	"app/internal/constants"
	authid_models "app/internal/models/authid"
	email_ban_models "app/internal/models/email_ban"
	user_models "app/internal/models/user"
	randoms_utils "app/internal/utilities/randoms"
	sessions_utils "app/internal/utilities/sessions"
	template_utils "app/internal/utilities/templates"
	"database/sql"
	"errors"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func (repo *HandlerRepo) GetUserAccountLogoutAll(w http.ResponseWriter, r *http.Request) {
	td := r.Context().Value(constants.TemplateDataContextKey).(*template_utils.TemplateData)
	sessionService := sessions_utils.NewSessionService(r, repo.Session)

	currentUser, err := sessionService.GetUserFromSession(repo.DB)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) && !errors.Is(err, sessions_utils.ErrNoUserInSession) {
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

	template_utils.RenderTemplate(w, repo.Templates["logout_all.html"], td)
}

func (repo *HandlerRepo) PostUserAccountLogoutAll(w http.ResponseWriter, r *http.Request) {
	td := r.Context().Value(constants.TemplateDataContextKey).(*template_utils.TemplateData)
	sessionService := sessions_utils.NewSessionService(r, repo.Session)

	currentUser, err := sessionService.GetUserFromSession(repo.DB)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) && !errors.Is(err, sessions_utils.ErrNoUserInSession) {
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
		template_utils.Abort(
			w,
			repo.Templates["error.html"],
			http.StatusBadRequest,
			constants.Error400Title,
			constants.Error400Message,
		)
		return
	}

	password := r.FormValue("password")
	if err := bcrypt.CompareHashAndPassword([]byte(currentUser.Password), []byte(password)); err != nil {
		td.Data["ErrorPassword"] = "Password not recognized"
		template_utils.RenderTemplate(w, repo.Templates["logout_all.html"], td)
		return
	}

	authidDbService := authid_models.NewAuthidDbService(repo.DB)
	userDbService := user_models.NewUserDbService(repo.DB)
	if err := func() error {
		tx, err := repo.DB.Beginx()
		if err != nil {
			return err
		}
		defer tx.Rollback()

		// create new authid
		var newAuthid *authid_models.Authid
		for range 5 {
			value, err := randoms_utils.GenerateAuthidValue()
			if err != nil {
				continue
			}

			authid, err := authidDbService.CreateWithTx(tx, value)
			if err != nil {
				continue
			}

			newAuthid = authid
			break

		}

		if newAuthid == nil {
			return errors.New("Unable to generate new authid")
		}

		// get current authid
		currentAuthid, err := authidDbService.ReadByIdWithTx(tx, currentUser.AuthidId)
		if err != nil {
			return err
		}

		// set ttl for current authid
		now := time.Now().UTC()
		ttl := now.AddDate(0, 0, 365)
		if err := authidDbService.UpdateTTLWithTx(tx, currentAuthid, &ttl); err != nil {
			return err
		}

		// change current user authid
		if err := userDbService.UpdateAuthidIdWithTx(tx, currentUser, newAuthid); err != nil {
			return err
		}

		if err := tx.Commit(); err != nil {
			return err
		}

		return nil
	}(); err != nil {
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

	sessionService.ToFlash("Successfully logged out all connected devices using your account.")
	sessionService.DestroySession()
	http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
}

// func (repo *UserHandlerRepo) AccountLogoutAll(w http.ResponseWriter, r *http.Request) {
// 	sessionService := sessions_utils.NewSessionService(r, repo.Session)
// 	userDbService := user_models.NewUserDbService(repo.DB)
// 	authidDbService := authid_models.NewAuthidDbService(repo.DB)
// 	rndr := template_utils.NewTemplateRenderer(repo.Session)
// 	emailBanDbService := email_ban_models.NewEmailBanDbService(repo.DB)

// 	td := r.Context().Value(constants.TemplateDataContextKey).(*template_utils.TemplateData)

// 	currentUser, err := sessionService.GetUserFromSession(repo.DB)
// 	if err != nil {
// 		if !errors.Is(err, sql.ErrNoRows) && !errors.Is(err, sessions_utils.ErrNoUserInSession) {
// 			rndr.Abort(
// 				w, r,
// 				repo.Templates["error.html"],
// 				http.StatusInternalServerError,
// 				constants.Error500Title,
// 				constants.Error500Message,
// 			)
// 			return
// 		}
// 		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
// 		return
// 	}

// 	ban, err := emailBanDbService.ReadByEmail(currentUser.Email)
// 	if err != nil {
// 		if !errors.Is(err, sql.ErrNoRows) {
// 			repo.Logger.Errorln(err)
// 			rndr.Abort(
// 				w, r,
// 				repo.Templates["error.html"],
// 				http.StatusInternalServerError,
// 				constants.Error500Title,
// 				constants.Error500Message,
// 			)
// 			return
// 		}
// 	}
// 	if ban != nil {
// 		sessionService.DestroySession()
// 		rndr.Abort(
// 			w, r,
// 			repo.Templates["error.html"],
// 			http.StatusForbidden,
// 			constants.Error403Title,
// 			constants.Error403Message,
// 		)
// 		return
// 	}

// 	// =========== GET ========== //
// 	if r.Method == http.MethodGet {
// 		rndr.RenderTemplate(w, r, repo.Templates["logout_all.html"], td)
// 		return
// 	}

// 	// =========== POST ========= //
// 	if r.Method == http.MethodPost {
// 		password := r.FormValue("password")
// 		if err := bcrypt.CompareHashAndPassword([]byte(currentUser.Password), []byte(password)); err != nil {
// 			td.Data["ErrorPassword"] = "Password not recognized"
// 			rndr.RenderTemplate(w, r, repo.Templates["logout_all.html"], td)
// 			return
// 		}

// 		if err := func() error {
// 			tx, err := repo.DB.Beginx()
// 			if err != nil {
// 				return err
// 			}
// 			defer tx.Rollback()

// 			// create new authid
// 			var newAuthid *authid_models.Authid
// 			for range 5 {
// 				value, err := randoms_utils.GenerateAuthidValue()
// 				if err != nil {
// 					continue
// 				}

// 				authid, err := authidDbService.CreateWithTx(tx, value)
// 				if err != nil {
// 					continue
// 				}

// 				newAuthid = authid
// 				break

// 			}

// 			if newAuthid == nil {
// 				return errors.New("Unable to generate new authid")
// 			}

// 			// get current authid
// 			currentAuthid, err := authidDbService.ReadByIdWithTx(tx, currentUser.AuthidId)
// 			if err != nil {
// 				return err
// 			}

// 			// set ttl for current authid
// 			now := time.Now().UTC()
// 			ttl := now.AddDate(0, 0, 365)
// 			if err := authidDbService.UpdateTTLWithTx(tx, currentAuthid, &ttl); err != nil {
// 				return err
// 			}

// 			// change current user authid
// 			if err := userDbService.UpdateAuthidIdWithTx(tx, currentUser, newAuthid); err != nil {
// 				return err
// 			}

// 			if err := tx.Commit(); err != nil {
// 				return err
// 			}

// 			return nil
// 		}(); err != nil {
// 			repo.Logger.Errorln(err)
// 			rndr.Abort(
// 				w, r,
// 				repo.Templates["error.html"],
// 				http.StatusInternalServerError,
// 				constants.Error500Title,
// 				constants.Error500Message,
// 			)
// 			return
// 		}

// 		sessionService.ToFlash("Successfully logged out all connected devices using your account.")
// 		sessionService.DestroySession()
// 		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
// 		return

// 	}

// 	rndr.Abort(
// 		w, r,
// 		repo.Templates["error.html"],
// 		http.StatusMethodNotAllowed,
// 		constants.Error405Title,
// 		constants.Error405Message,
// 	)

// }
