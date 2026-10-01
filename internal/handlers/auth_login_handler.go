package handlers

import (
	"app/internal/constants"
	authid_models "app/internal/models/authid"
	email_ban_models "app/internal/models/email_ban"
	user_models "app/internal/models/user"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	sessions_utils "app/internal/utilities/sessions"
	strings_utils "app/internal/utilities/strings"
	template_utils "app/internal/utilities/templates"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func (repo *HandlerRepo) GetAuthLogin(w http.ResponseWriter, r *http.Request) {
	td := r.Context().Value(constants.TemplateDataContextKey).(*template_utils.TemplateData)

	sessionService := sessions_utils.NewSessionService(r, repo.Session)

	currentUser, err := sessionService.GetUserFromSession(repo.DB)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) && !errors.Is(err, sessions_utils.ErrNoUserInSession) {
			repo.Logger.Errorln(err)
			template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
		}
	}

	if currentUser != nil {
		// user is in session
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	template_utils.RenderTemplate(w, repo.Templates["login.html"], td)

}

func (repo *HandlerRepo) PostAuthLogin(w http.ResponseWriter, r *http.Request) {
	td := r.Context().Value(constants.TemplateDataContextKey).(*template_utils.TemplateData)

	sessionService := sessions_utils.NewSessionService(r, repo.Session)

	currentUser, err := sessionService.GetUserFromSession(repo.DB)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) && !errors.Is(err, sessions_utils.ErrNoUserInSession) {
			repo.Logger.Errorln(err)
			template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
		}
	}

	if currentUser != nil {
		// user is in session
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusBadRequest, constants.Error400Title, constants.Error400Message)
		return
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")
	remember := r.Form.Get("remember") != ""
	td.Data["Email"] = email

	email = strings.TrimSpace(strings.ToLower(email))

	// check ban
	emailBanDbService := email_ban_models.NewEmailBanDbService(repo.DB)
	ban, err := emailBanDbService.ReadByEmail(email)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			repo.Logger.Errorln(err)
			template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
		}
	}
	if ban != nil {
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusForbidden, constants.Error403Title, constants.Error403Message)
		return
	}

	userDbService := user_models.NewUserDbService(repo.DB)
	user, err := userDbService.ReadByEmail(email)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			repo.Logger.Errorln(err)
			template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
			return
		}

		td.Data["ErrorEmail"] = "Invalid email and/or password"
		td.Data["ErrorPassword"] = "Invalid email and/or password"
		template_utils.RenderTemplate(w, repo.Templates["login.html"], td)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		td.Data["ErrorEmail"] = "Invalid email and/or password"
		td.Data["ErrorPassword"] = "Invalid email and/or password"
		template_utils.RenderTemplate(w, repo.Templates["login.html"], td)
		return
	}

	var flashMsg string

	if user.DeactivatedAt.Valid {
		if err := userDbService.UpdateReactivate(user); err != nil {
			repo.Logger.Errorln(err)
			template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
			return
		}
		flashMsg = fmt.Sprintf("Welcome back, %s. You account has been recativated.", user.Firstname)
	} else {
		flashMsg = fmt.Sprintf("%s, you are now logged in.", strings_utils.Capitalize(user.Firstname))
	}

	authDbService := authid_models.NewAuthidDbService(repo.DB)
	authid, err := authDbService.ReadById(user.AuthidId)
	if err != nil {
		repo.Logger.Errorln(err)
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
		return
	}

	sessionService.Login(authid.Value, remember)
	sessionService.ToFlash(flashMsg)
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}
