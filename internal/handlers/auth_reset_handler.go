package handlers

import (
	"app/internal/constants"
	authid_models "app/internal/models/authid"
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
	"net/http"
	"os"
	"strings"
	"time"
)

func (repo *HandlerRepo) GetResetPassword(w http.ResponseWriter, r *http.Request) {
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

	emailBanDbService := email_ban_models.NewEmailBanDbService(repo.DB)
	if currentUser != nil {

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
			http.Redirect(w, r, r.URL.RequestURI(), http.StatusSeeOther)
			return

		}
	}

	token := r.URL.Query().Get("token")
	hashedToken := strings_utils.HashToken(token)
	resetTokenDbService := reset_models.NewResetTokenDbService(repo.DB)
	resetToken, err := resetTokenDbService.ReadByHashed(hashedToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			template_utils.Abort(w, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
			return
		}

		repo.Logger.Errorln(err)
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
		return
	}

	if resetToken.IsUsed {
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
		return
	}

	userDbService := user_models.NewUserDbService(repo.DB)
	userToken, err := userDbService.ReadById(resetToken.UserId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			template_utils.Abort(w, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
			return
		}
		repo.Logger.Errorln(err)
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
		return
	}

	// this prevents logged in user that isnt the owner of the reset link
	// to replace the password of the link owner
	if currentUser != nil {
		if currentUser.ID != userToken.ID {
			sessionService.ToFlash("Link does not belong to the current user")
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}
	}

	// check ban for user who owns token
	ban, err := emailBanDbService.ReadByEmail(userToken.Email)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			repo.Logger.Errorln(err)
			template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
			return
		}
	}
	if ban != nil {
		sessionService.ToFlash("User associated with this token is currently banned")
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusForbidden, constants.Error403Title, constants.Error403Message)
		return
	}

	if resetToken.ExpiresAt.Before(time.Now().UTC()) {
		sessionService.ToFlash("Link expired. Request a new one.")
		http.Redirect(w, r, "/auth/forgot_password", http.StatusSeeOther)
		return
	}

	template_utils.RenderTemplate(w, repo.Templates["reset_password.html"], td)
}

func (repo *HandlerRepo) PostResetPassword(w http.ResponseWriter, r *http.Request) {
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

	emailBanDbService := email_ban_models.NewEmailBanDbService(repo.DB)
	if currentUser != nil {

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
			http.Redirect(w, r, r.URL.RequestURI(), http.StatusSeeOther)
			return

		}
	}

	token := r.URL.Query().Get("token")
	hashedToken := strings_utils.HashToken(token)
	resetTokenDbService := reset_models.NewResetTokenDbService(repo.DB)
	resetToken, err := resetTokenDbService.ReadByHashed(hashedToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			template_utils.Abort(w, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
			return
		}

		repo.Logger.Errorln(err)
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
		return
	}

	if resetToken.IsUsed {
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
		return
	}

	userDbService := user_models.NewUserDbService(repo.DB)
	userToken, err := userDbService.ReadById(resetToken.UserId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			template_utils.Abort(w, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
			return
		}
		repo.Logger.Errorln(err)
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
		return
	}

	if os.Getenv("ENVIRONMENT") == "production" {
		captchaToken := r.FormValue(constants.TurnstileCaptchaFormName)
		if err := captcha_utils.CheckCaptcha(captchaToken); err != nil {
			repo.Logger.Errorln(err)
			td.Data["ErrorCaptcha"] = "Verification failed. Please refresh the page and try again."
			template_utils.RenderTemplate(w, repo.Templates["reset_password.html"], td)
			return
		}
	}

	password := r.FormValue("password")
	repeat := r.FormValue("repeat")

	if len(strings.TrimSpace(password)) < 1 {
		td.Data["ErrorPassword"] = "Do not use ALL spaces. This is not allowed"
		template_utils.RenderTemplate(w, repo.Templates["reset_password.html"], td)
		return
	}

	if len(password) < 8 {
		td.Data["ErrorPassword"] = "Must be 8 characters of longer"
		template_utils.RenderTemplate(w, repo.Templates["reset_password.html"], td)
		return
	}

	if password != repeat {
		td.Data["ErrorRepeat"] = "Did not match new password"
		template_utils.RenderTemplate(w, repo.Templates["reset_password.html"], td)
		return
	}

	// anonymous function for transaction
	authidDbService := authid_models.NewAuthidDbService(repo.DB)
	errTx := func() error {
		tx, err := repo.DB.Beginx()
		if err != nil {
			return err
		}
		defer tx.Rollback()

		// mark the reset token model as used
		if err := resetTokenDbService.UpdateIsUsedWithTx(tx, resetToken, true); err != nil {
			return err
		}

		// update the user. must create new authid
		// it would be the same as changing password via account
		// flow: get current authid, assign ttl, create new authid, assign to user

		authid, err := authidDbService.ReadByIdWithTx(tx, userToken.AuthidId)
		if err != nil {
			return err
		}

		// assigned ttl to the current authid
		now := time.Now().UTC()
		ttl := now.AddDate(1, 0, 0)
		if err = authidDbService.UpdateTTLWithTx(tx, authid, &ttl); err != nil {
			return err
		}

		// create new authid
		var newAuthid *authid_models.Authid
		for range 5 {

			val, err := randoms_utils.GenerateAuthidValue()
			if err != nil {
				continue
			}

			authid, err := authidDbService.CreateWithTx(tx, val)
			if err != nil {
				continue
			} else {
				newAuthid = authid
				break
			}
		}

		if newAuthid == nil {
			err := errors.New("Unable to generate new authid")
			return err
		}

		// update authid - forces a global logout
		if err := userDbService.UpdateAuthidIdWithTx(tx, userToken, newAuthid); err != nil {
			return err
		}

		// update the password
		if err := userDbService.UpdatePasswordWithTx(tx, userToken, password); err != nil {
			return err
		}

		// commit
		if err := tx.Commit(); err != nil {
			return err
		}
		return nil
	}()

	if errTx != nil {
		repo.Logger.Errorln(err)
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
		return
	}

	if currentUser != nil {
		sessionService.DestroySession()
	}

	if os.Getenv("ENVIRONMENT") == "production" {
		go func() {
			if err := email_utils.SendPasswordChangedEmail(userToken.Email, userToken.Firstname); err != nil {
				repo.Logger.Errorln(err)
			}
		}()
	}

	sessionService.ToFlash("Please login with your new password")
	http.Redirect(w, r, "/auth/login", http.StatusSeeOther)

}
