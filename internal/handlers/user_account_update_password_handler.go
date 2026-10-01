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
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func (repo *HandlerRepo) GetUserAccountUpdatePassword(w http.ResponseWriter, r *http.Request) {
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

	template_utils.RenderTemplate(w, repo.Templates["update_password.html"], td)
}

func (repo *HandlerRepo) PostUserAccountUpdatePassword(w http.ResponseWriter, r *http.Request) {
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
		template_utils.Abort(
			w,
			repo.Templates["error.html"],
			http.StatusForbidden,
			constants.Error403Title,
			constants.Error403Message,
		)
		return
	}

	password := r.Form.Get("current")
	if err := bcrypt.CompareHashAndPassword([]byte(currentUser.Password), []byte(password)); err != nil {
		td.Data["ErrorCurrent"] = "Password not recognized"
		template_utils.RenderTemplate(w, repo.Templates["update_password.html"], td)
		return
	}

	newPassword := r.Form.Get("new")
	if len(strings.TrimSpace(newPassword)) < 1 { // check that new password is not just all spaces
		td.Data["ErrorNew"] = "All spaces is not accepted"
		template_utils.RenderTemplate(w, repo.Templates["update_password.html"], td)
		return
	}

	if len(newPassword) < 8 {
		td.Data["ErrorNew"] = "Must be atleast 8 characters length"
		template_utils.RenderTemplate(w, repo.Templates["update_password.html"], td)
		return
	}

	repeat := r.Form.Get("repeat")
	if newPassword != repeat {
		td.Data["ErrorRepeat"] = "Must match new password"
		template_utils.RenderTemplate(w, repo.Templates["update_password.html"], td)
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

		// get current authid
		authid, err := authidDbService.ReadByIdWithTx(tx, currentUser.AuthidId)
		if err != nil {
			return err
		}

		// set current authid ttl
		now := time.Now().UTC()
		ttl := now.AddDate(0, 0, 365)
		if err := authidDbService.UpdateTTLWithTx(tx, authid, &ttl); err != nil {
			return err
		}

		// create new authid
		var newAuthid *authid_models.Authid
		for range 5 {
			value, err := randoms_utils.GenerateAuthidValue()
			if err != nil {
				return err
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

		// replace the current user authid
		if err := userDbService.UpdateAuthidIdWithTx(tx, currentUser, newAuthid); err != nil {
			return err
		}

		// replace the password
		if err := userDbService.UpdatePasswordWithTx(tx, currentUser, newPassword); err != nil {
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

	// send email here
	// but not now, limited amount of email sending

	sessionService.ToFlash("Password change. Please re-login using your new password")
	sessionService.DestroySession()
	http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
}

// func (repo *UserHandlerRepo) AccountUpdatePassword(w http.ResponseWriter, r *http.Request) {
// 	sessionService := sessions_utils.NewSessionService(r, repo.Session)
// 	authidDbService := authid_models.NewAuthidDbService(repo.DB)
// 	userDbService := user_models.NewUserDbService(repo.DB)
// 	rndr := template_utils.NewTemplateRenderer(repo.Session)
// 	emailBanDbService := email_ban_models.NewEmailBanDbService(repo.DB)

// 	td := r.Context().Value(constants.TemplateDataContextKey).(*template_utils.TemplateData)

// 	// td.Data["ErrorCurrent"] = ""
// 	// td.Data["ErrorNew"] = ""
// 	// td.Data["ErrorRepeat"] = ""

// 	currentUser, err := sessionService.GetUserFromSession(repo.DB)
// 	if err != nil {
// 		if !errors.Is(err, sql.ErrNoRows) && !errors.Is(err, sessions_utils.ErrNoUserInSession) {
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

// 	// ======== GET ========= //
// 	if r.Method == http.MethodGet {
// 		rndr.RenderTemplate(w, r, repo.Templates["update_password.html"], td)
// 		return
// 	}

// 	// ======== POST ======== //
// 	if r.Method == http.MethodPost {

// 		password := r.FormValue("current")
// 		if err := bcrypt.CompareHashAndPassword([]byte(currentUser.Password), []byte(password)); err != nil {
// 			td.Data["ErrorCurrent"] = "Password not recognized"
// 			rndr.RenderTemplate(w, r, repo.Templates["update_password.html"], td)
// 			return
// 		}

// 		newPassword := r.FormValue("new")
// 		if len(strings.TrimSpace(newPassword)) < 1 { // check that new password is not just all spaces
// 			td.Data["ErrorNew"] = "All spaces is not accepted"
// 			rndr.RenderTemplate(w, r, repo.Templates["update_password.html"], td)
// 			return
// 		}

// 		if len(newPassword) < 8 {
// 			td.Data["ErrorNew"] = "Must be atleast 8 characters length"
// 			rndr.RenderTemplate(w, r, repo.Templates["update_password.html"], td)
// 			return
// 		}

// 		repeat := r.FormValue("repeat")
// 		if newPassword != repeat {
// 			td.Data["ErrorRepeat"] = "Must match new password"
// 			rndr.RenderTemplate(w, r, repo.Templates["update_password.html"], td)
// 			return
// 		}

// 		if err := func() error {
// 			tx, err := repo.DB.Beginx()
// 			if err != nil {
// 				return err
// 			}
// 			defer tx.Rollback()

// 			// get current authid
// 			authid, err := authidDbService.ReadByIdWithTx(tx, currentUser.AuthidId)
// 			if err != nil {
// 				return err
// 			}

// 			// set current authid ttl
// 			now := time.Now().UTC()
// 			ttl := now.AddDate(0, 0, 365)
// 			if err := authidDbService.UpdateTTLWithTx(tx, authid, &ttl); err != nil {
// 				return err
// 			}

// 			// create new authid
// 			var newAuthid *authid_models.Authid
// 			for range 5 {
// 				value, err := randoms_utils.GenerateAuthidValue()
// 				if err != nil {
// 					return err
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

// 			// replace the current user authid
// 			if err := userDbService.UpdateAuthidIdWithTx(tx, currentUser, newAuthid); err != nil {
// 				return err
// 			}

// 			// replace the password
// 			if err := userDbService.UpdatePasswordWithTx(tx, currentUser, newPassword); err != nil {
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

// 		// send email here
// 		// but not now, limited amount of email sending

// 		sessionService.ToFlash("Password change. Please re-login using your new password")
// 		sessionService.DestroySession()
// 		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
// 		return
// 	}

// 	rndr.Abort(w, r, repo.Templates["error.html"], http.StatusMethodNotAllowed, constants.Error405Title, constants.Error405Message)
// }
