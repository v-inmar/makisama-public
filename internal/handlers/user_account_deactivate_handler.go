package handlers

import (
	"app/internal/constants"
	email_ban_models "app/internal/models/email_ban"
	user_models "app/internal/models/user"
	sessions_utils "app/internal/utilities/sessions"
	template_utils "app/internal/utilities/templates"
	"database/sql"
	"errors"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

func (repo *HandlerRepo) GetUserAccountDeactivate(w http.ResponseWriter, r *http.Request) {
	td := r.Context().Value(constants.TemplateDataContextKey).(*template_utils.TemplateData)

	sessionService := sessions_utils.NewSessionService(r, repo.Session)
	currentUser, err := sessionService.GetUserFromSession(repo.DB)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) && !errors.Is(err, sessions_utils.ErrNoUserInSession) {
			template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
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
	template_utils.RenderTemplate(w, repo.Templates["deactivate.html"], td)
}

func (repo *HandlerRepo) PostUserAccountDeactivate(w http.ResponseWriter, r *http.Request) {
	td := r.Context().Value(constants.TemplateDataContextKey).(*template_utils.TemplateData)

	sessionService := sessions_utils.NewSessionService(r, repo.Session)
	currentUser, err := sessionService.GetUserFromSession(repo.DB)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) && !errors.Is(err, sessions_utils.ErrNoUserInSession) {
			template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
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

	password := r.FormValue("password")
	if err := bcrypt.CompareHashAndPassword([]byte(currentUser.Password), []byte(password)); err != nil {
		td.Data["ErrorPassword"] = "Password not recognized"
		template_utils.RenderTemplate(w, repo.Templates["deactivate.html"], td)
		return
	}

	userDbService := user_models.NewUserDbService(repo.DB)
	// just deactivate user
	// because each route for any boards or feedback will
	// check if the board owner is active or not
	if err := userDbService.UpdateDeactivate(currentUser); err != nil {
		repo.Logger.Errorln(err)
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
		return
	}

	// send email here letting user know about deactivated account
	// NOTE: NOT NOW, NOT ENOUGH DAILY ALLOCATED EMAIL SENDING

	sessionService.DestroySession()
	sessionService.ToFlash("Account deactivated. Log back in to reactivate")
	http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
}

// func (repo *UserHandlerRepo) AccountDeactivate(w http.ResponseWriter, r *http.Request) {
// 	sessionService := sessions_utils.NewSessionService(r, repo.Session)
// 	rndr := template_utils.NewTemplateRenderer(repo.Session)
// 	userDbService := user_models.NewUserDbService(repo.DB)
// 	emailBanDbService := email_ban_models.NewEmailBanDbService(repo.DB)

// 	currentUser, err := sessionService.GetUserFromSession(repo.DB)
// 	if err != nil {
// 		if !errors.Is(err, sql.ErrNoRows) && !errors.Is(err, sessions_utils.ErrNoUserInSession) {
// 			rndr.Abort(w, r, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
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

// 	td := r.Context().Value(constants.TemplateDataContextKey).(*template_utils.TemplateData)
// 	if r.Method == http.MethodGet {
// 		td.Data["ErrorPassword"] = ""
// 		rndr.RenderTemplate(w, r, repo.Templates["deactivate.html"], td)
// 		return
// 	}

// 	if r.Method == http.MethodPost {
// 		password := r.FormValue("password")
// 		if err := bcrypt.CompareHashAndPassword([]byte(currentUser.Password), []byte(password)); err != nil {
// 			td.Data["ErrorPassword"] = "Password not recognized"
// 			rndr.RenderTemplate(w, r, repo.Templates["deactivate.html"], td)
// 			return
// 		}

// 		// just deactivate user
// 		// because each route for any boards or feedback will
// 		// check if the board owner is active or not
// 		if err := userDbService.UpdateDeactivate(currentUser); err != nil {
// 			repo.Logger.Errorln(err)
// 			rndr.Abort(w, r, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
// 			return
// 		}

// 		// send email here letting user know about deactivated account
// 		// NOTE: NOT NOW, NOT ENOUGH DAILY ALLOCATED EMAIL SENDING

// 		sessionService.DestroySession()
// 		sessionService.ToFlash("Account deactivated. Log back in to reactivate")
// 		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)

// 	}

// 	rndr.Abort(w, r, repo.Templates["error.html"], http.StatusMethodNotAllowed, constants.Error405Title, constants.Error405Message)

// }
