package handlers

import (
	"app/internal/constants"
	confirm_email_models "app/internal/models/confirm_email"
	email_ban_models "app/internal/models/email_ban"
	user_models "app/internal/models/user"
	sessions_utils "app/internal/utilities/sessions"
	strings_utils "app/internal/utilities/strings"
	template_utils "app/internal/utilities/templates"
	"database/sql"
	"errors"
	"net/http"
	"time"
)

func (repo *HandlerRepo) GetAuthConfirmEmail(w http.ResponseWriter, r *http.Request) {
	// td := r.Context().Value(constants.TemplateDataContextKey).(*template_utils.TemplateData)

	sessionService := sessions_utils.NewSessionService(r, repo.Session)
	currentUser, err := sessionService.GetUserFromSession(repo.DB)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) && !errors.Is(err, sessions_utils.ErrNoUserInSession) {
			repo.Logger.Errorln(err)
			template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
			return
		}
		sessionService.ToFlash("Login in and re-click the link")
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	token := r.URL.Query().Get("token")
	hashed := strings_utils.HashToken(token)
	tokenDbService := confirm_email_models.NewConfirmTokenDbSerivce(repo.DB)
	c, err := tokenDbService.ReadByHashed(hashed)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			template_utils.Abort(w, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
			return
		}
		repo.Logger.Errorln(err)
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
		return
	}

	userConfirmDbService := confirm_email_models.NewUserConfirmDbService(repo.DB)
	uc, err := userConfirmDbService.ReadByConfirmID(c.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			template_utils.Abort(w, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
			return
		}
		repo.Logger.Errorln(err)
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
		return
	}

	// check the current user/email is banned
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

	// check user owns the token
	if currentUser.ID != uc.UserId {
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
		return
	}

	// check if user is already verified
	if currentUser.VerifiedAt.Valid {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	// check if token had expired
	if c.ExpiresAt.Before(time.Now().UTC()) {
		// redirect to request new verification
		sessionService.ToFlash("Link expired. Please request a new one.")
		http.Redirect(w, r, "/auth/request_confirm", http.StatusSeeOther)
		return
	}

	verifiedTime := sql.NullTime{
		Time:  time.Now().UTC(),
		Valid: true,
	}

	userDbService := user_models.NewUserDbService(repo.DB)
	if err := userDbService.UpdateVerified(currentUser, &verifiedTime); err != nil {
		repo.Logger.Errorln(err)
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
		return
	}

	sessionService.ToFlash("Email verified!")
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}
