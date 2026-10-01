package handlers

import (
	"app/internal/constants"
	sessions_utils "app/internal/utilities/sessions"
	strings_utils "app/internal/utilities/strings"
	template_utils "app/internal/utilities/templates"
	"database/sql"
	"errors"
	"net/http"
)

func (repo *HandlerRepo) GetHomeHome(w http.ResponseWriter, r *http.Request) {
	td := r.Context().Value(constants.TemplateDataContextKey).(*template_utils.TemplateData)
	sessionService := sessions_utils.NewSessionService(r, repo.Session)

	currentUser, err := sessionService.GetUserFromSession(repo.DB)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) && !errors.Is(err, sessions_utils.ErrNoUserInSession) {
			repo.Logger.Errorln(err)
			template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
			return
		}
	}

	if currentUser != nil {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	td.Data["AppName"] = strings_utils.Capitalize(constants.AppName)

	template_utils.RenderTemplate(w, repo.Templates["home.html"], td)
}

// func (repo *HomeHandlerRepo) GetHome(w http.ResponseWriter, r *http.Request) {

// 	sessionService := sessions_utils.NewSessionService(r, repo.Session)
// 	rndr := template_utils.NewTemplateRenderer(repo.Session)

// 	currentUser, err := sessionService.GetUserFromSession(repo.DB)
// 	if err != nil {
// 		if !errors.Is(err, sql.ErrNoRows) && !errors.Is(err, sessions_utils.ErrNoUserInSession) {
// 			repo.Logger.Errorln(err)
// 			rndr.Abort(w, r, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
// 			return
// 		}
// 	}

// 	if currentUser != nil {
// 		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
// 		return
// 	}

// 	td := r.Context().Value(constants.TemplateDataContextKey).(*template_utils.TemplateData)
// 	td.Data["AppName"] = strings_utils.Capitalize(constants.AppName)

// 	rndr.RenderTemplate(w, r, repo.Templates["home.html"], td)
// }
