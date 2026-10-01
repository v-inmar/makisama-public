package handlers

import (
	"app/internal/constants"
	email_ban_models "app/internal/models/email_ban"
	reset_models "app/internal/models/reset"
	user_models "app/internal/models/user"
	captcha_utils "app/internal/utilities/captcha"
	email_utils "app/internal/utilities/email"
	randoms_utils "app/internal/utilities/randoms"
	sessions_utils "app/internal/utilities/sessions"
	strings_utils "app/internal/utilities/strings"
	template_utils "app/internal/utilities/templates"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
)

func (repo *HandlerRepo) GetAuthForgot(w http.ResponseWriter, r *http.Request) {
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

	template_utils.RenderTemplate(w, repo.Templates["forgot_password.html"], td)
}

func (repo *HandlerRepo) PostAuthForgot(w http.ResponseWriter, r *http.Request) {
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

	if err := r.ParseForm(); err != nil {
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusBadRequest, constants.Error400Title, constants.Error400Message)
		return
	}

	if os.Getenv("ENVIRONMENT") == "production" {
		captchaToken := r.Form.Get(constants.TurnstileCaptchaFormName)
		if err := captcha_utils.CheckCaptcha(captchaToken); err != nil {
			repo.Logger.Errorln(err)
			td.Data["ErrorCaptcha"] = "Verification failed. Please refresh the page and try again."
			template_utils.RenderTemplate(w, repo.Templates["reset_password.html"], td)
			return
		}
	}

	email := r.Form.Get("email")
	td.Data["Email"] = email
	email = strings.TrimSpace(strings.ToLower(email))
	if !strings_utils.IsValidLength(email, constants.EmailMinLength, constants.EmailMaxLength) {
		td.Data["ErrorEmail"] = fmt.Sprintf("Must be between %d and %d characters long", constants.EmailMinLength, constants.EmailMaxLength)
		template_utils.RenderTemplate(w, repo.Templates["forgot_password.html"], td)
		return
	}

	if !strings_utils.IsValidEmail(email) {
		td.Data["ErrorEmail"] = "Invalid format"
		template_utils.RenderTemplate(w, repo.Templates["forgot_password.html"], td)
		return
	}

	flashMsg := "If an account exists for this email, we’ve sent password reset instructions. Please check your inbox or spam folder."

	// check the email is not banned
	emailBanDbService := email_ban_models.NewEmailBanDbService(repo.DB)
	ban, err := emailBanDbService.ReadByEmail(email)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			repo.Logger.Errorln(err)
			template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
			return
		}
	}
	if ban != nil {
		// dont let user know if the email banned or not
		// to avoid email guessing
		// just redirect
		sessionService.ToFlash(flashMsg)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// get the user attached to the email
	userDbService := user_models.NewUserDbService(repo.DB)
	user, err := userDbService.ReadByEmail(email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			sessionService.ToFlash(flashMsg)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return

		}
		repo.Logger.Errorln(err)
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
		return
	}

	// if user.DeactivatedAt.Valid {
	// 	sessionService.ToFlash(flashMsg)
	// 	// so it preserves the flash message and not get lost in multiple reidrects
	// 	if currentUser != nil {
	// 		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	// 		return
	// 	}
	// 	http.Redirect(w, r, "/", http.StatusSeeOther)
	// 	return
	// }

	// try 5 times just incase collision happens
	resetTokenDbService := reset_models.NewResetTokenDbService(repo.DB)
	var token string
	for range 5 {
		// token to be sent to the user as part of the url (query param token=)
		tok, err := randoms_utils.GenerateResetToken()
		if err != nil {
			repo.Logger.Errorln(err)
			continue
		}

		// store this in db
		hashedToken := strings_utils.HashToken(tok)

		_, errCreate := resetTokenDbService.Create(user.ID, hashedToken)
		if errCreate != nil {
			repo.Logger.Errorln(errCreate)
			continue
		}

		token = tok
		break
	}

	if token == "" {
		repo.Logger.Errorln(errors.New("Unable to generate token"))
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
		return
	}

	if os.Getenv("ENVIRONMENT") == "production" {
		go func() {
			if err := email_utils.SendForgotPasswordEmail(user.Email, user.Firstname, token); err != nil {
				repo.Logger.Errorln(err)
			}
		}()
	}

	// fmt.Printf("Token: %s\n", token)
	// fmt.Printf("Hashed Token: %s\n", hashedToken)
	// fmt.Printf("User ID: %d\n", user.ID)
	// fmt.Printf("The url will be like this: localhost:5000/auth/reset_password?token=%s\n", token)

	sessionService.ToFlash(flashMsg)
	http.Redirect(w, r, "/", http.StatusSeeOther)

}
