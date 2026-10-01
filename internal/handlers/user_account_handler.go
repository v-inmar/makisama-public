package handlers

import (
	"app/internal/constants"
	email_ban_models "app/internal/models/email_ban"
	sessions_utils "app/internal/utilities/sessions"
	strings_utils "app/internal/utilities/strings"
	template_utils "app/internal/utilities/templates"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
)

func (repo *HandlerRepo) GetUserAccount(w http.ResponseWriter, r *http.Request) {
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

	name := fmt.Sprintf("%s %s", strings_utils.Capitalize(currentUser.Firstname), strings_utils.Capitalize(currentUser.Lastname))

	td.Data["Name"] = strings_utils.Capitalize(name)
	td.Data["Email"] = strings_utils.ObfuscateEmail(currentUser.Email)

	template_utils.RenderTemplate(
		w,
		repo.Templates["account.html"],
		td,
	)
}

// func (repo *UserHandlerRepo) Account(w http.ResponseWriter, r *http.Request) {
// 	sessionService := sessions_utils.NewSessionService(r, repo.Session)
// 	rndr := template_utils.NewTemplateRenderer(repo.Session)
// 	emailBanDbService := email_ban_models.NewEmailBanDbService(repo.DB)

// 	td := r.Context().Value(constants.TemplateDataContextKey).(*template_utils.TemplateData)

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

// 	if r.Method == http.MethodGet {
// 		name := fmt.Sprintf("%s %s", strings_utils.Capitalize(currentUser.Firstname), strings_utils.Capitalize(currentUser.Lastname))

// 		td.Data["Name"] = strings_utils.Capitalize(name)
// 		td.Data["Email"] = strings_utils.ObfuscateEmail(currentUser.Email)

// 		rndr.RenderTemplate(
// 			w, r,
// 			repo.Templates["account.html"],
// 			td,
// 		)
// 		return
// 	}

// 	rndr.Abort(w, r, repo.Templates["error.html"], http.StatusMethodNotAllowed, constants.Error405Title, constants.Error405Message)

// }
