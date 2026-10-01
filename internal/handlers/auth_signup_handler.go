package handlers

import (
	"app/internal/constants"
	authid_models "app/internal/models/authid"
	confirm_email_models "app/internal/models/confirm_email"
	email_ban_models "app/internal/models/email_ban"
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

func (repo *HandlerRepo) GetAuthSignup(w http.ResponseWriter, r *http.Request) {
	td := r.Context().Value(constants.TemplateDataContextKey).(*template_utils.TemplateData)
	td.Data["FirstnameMinLen"] = constants.FirstnameMinLength
	td.Data["FirstnameMaxLen"] = constants.FirstnameMaxLength
	td.Data["LastnameMinLen"] = constants.LastnameMinLength
	td.Data["LastnameMaxLen"] = constants.LastnameMaxLength
	td.Data["EmailMinLen"] = constants.EmailMinLength
	td.Data["EmailMaxLen"] = constants.EmailMaxLength
	td.Data["PasswordMinLen"] = constants.PasswordMinLength
	td.Data["PasswordMaxLen"] = constants.PasswordMaxLength

	sessionService := sessions_utils.NewSessionService(r, repo.Session)
	currentUser, err := sessionService.GetUserFromSession(repo.DB)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) && !errors.Is(err, sessions_utils.ErrNoUserInSession) {
			template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
			return
		}
	}
	if currentUser != nil {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}
	template_utils.RenderTemplate(w, repo.Templates["signup.html"], td)
}

func (repo *HandlerRepo) PostAuthSignup(w http.ResponseWriter, r *http.Request) {
	td := r.Context().Value(constants.TemplateDataContextKey).(*template_utils.TemplateData)
	td.Data["FirstnameMinLen"] = constants.FirstnameMinLength
	td.Data["FirstnameMaxLen"] = constants.FirstnameMaxLength
	td.Data["LastnameMinLen"] = constants.LastnameMinLength
	td.Data["LastnameMaxLen"] = constants.LastnameMaxLength
	td.Data["EmailMinLen"] = constants.EmailMinLength
	td.Data["EmailMaxLen"] = constants.EmailMaxLength
	td.Data["PasswordMinLen"] = constants.PasswordMinLength
	td.Data["PasswordMaxLen"] = constants.PasswordMaxLength

	sessionService := sessions_utils.NewSessionService(r, repo.Session)
	currentUser, err := sessionService.GetUserFromSession(repo.DB)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) && !errors.Is(err, sessions_utils.ErrNoUserInSession) {
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

	firstname := r.Form.Get("firstname")
	lastname := r.Form.Get("lastname")
	email := r.Form.Get("email")
	password := r.Form.Get("password")
	repeat := r.Form.Get("repeat")

	td.Data["Firstname"] = firstname
	td.Data["Lastname"] = lastname
	td.Data["Email"] = email

	if os.Getenv("ENVIRONMENT") == "production" {
		captchaToken := r.Form.Get(constants.TurnstileCaptchaFormName)
		if err := captcha_utils.CheckCaptcha(captchaToken); err != nil {

			td.Data["ErrorCaptcha"] = "Verification failed. Please refresh the page and try again."
			template_utils.RenderTemplate(w, repo.Templates["signup.html"], td)
			return
		}
	}

	email = strings.ToLower(strings.TrimSpace(email))
	if !strings_utils.IsValidEmail(email) {

		td.Data["ErrorEmail"] = "Invalid format."
		template_utils.RenderTemplate(w, repo.Templates["signup.html"], td)
		return
	}
	if !strings_utils.IsValidLength(email, constants.EmailMinLength, constants.EmailMaxLength) {

		td.Data["ErrorEmail"] = fmt.Sprintf("Must be between %d and %d characters long.", constants.EmailMinLength, constants.EmailMaxLength)
		template_utils.RenderTemplate(w, repo.Templates["signup.html"], td)
		return
	}

	// check if email is ban
	banDbService := email_ban_models.NewEmailBanDbService(repo.DB)
	ban, err := banDbService.ReadByEmail(email)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			repo.Logger.Errorln(err)
			template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
			return
		}
	}
	if ban != nil {
		td.Data["ErrorEmail"] = "This email address is not accepted. Use another one"
		template_utils.RenderTemplate(w, repo.Templates["signup.html"], td)
		return
	}

	userDbService := user_models.NewUserDbService(repo.DB)
	emailUser, err := userDbService.ReadByEmail(email)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			repo.Logger.Errorln(err)
			template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
			return
		}

	}
	if emailUser != nil {
		td.Data["ErrorEmail"] = "Email address is already in use"
		template_utils.RenderTemplate(w, repo.Templates["signup.html"], td)
		return
	}

	// check firstname
	firstname = strings.TrimSpace(firstname)
	if !strings_utils.IsValidLength(firstname, constants.FirstnameMinLength, constants.FirstnameMaxLength) {

		td.Data["ErrorFirstname"] = fmt.Sprintf("Must be between %d and %d characters long.", constants.FirstnameMinLength, constants.FirstnameMaxLength)
		template_utils.RenderTemplate(w, repo.Templates["signup.html"], td)
		return
	}

	// check lastname
	lastname = strings.TrimSpace(lastname)
	if !strings_utils.IsValidLength(lastname, constants.LastnameMinLength, constants.LastnameMaxLength) {

		td.Data["ErrorLastname"] = fmt.Sprintf("Must be between %d and %d characters long.", constants.LastnameMinLength, constants.LastnameMaxLength)
		template_utils.RenderTemplate(w, repo.Templates["signup.html"], td)
		return
	}

	// prevents users from using all spaces as password
	if len(strings.TrimSpace(password)) < 1 {
		td.Data["ErrorPassword"] = "All spaces is not allowed"
		template_utils.RenderTemplate(w, repo.Templates["signup.html"], td)
		return
	}

	if !strings_utils.IsValidLength(password, constants.PasswordMinLength, constants.PasswordMaxLength) {

		td.Data["ErrorPassword"] = fmt.Sprintf("Must be between %d and %d characters long.", constants.PasswordMinLength, constants.PasswordMaxLength)
		template_utils.RenderTemplate(w, repo.Templates["signup.html"], td)
		return
	}

	if password != repeat {
		td.Data["ErrorRepeat"] = "Must match Password"
		template_utils.RenderTemplate(w, repo.Templates["signup.html"], td)
		return
	}

	// anonymous function for transaction
	// returns the newly created authid
	// so we can grab the value and place it in session
	var newAuthidValue string
	var newConfirmTokenValue string
	authidDbService := authid_models.NewAuthidDbService(repo.DB)
	errCreate := func() error {

		// start transaction
		tx, err := repo.DB.Beginx()
		if err != nil {
			return err
		}
		defer tx.Rollback()

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

		u, err := userDbService.CreateWithTx(tx, firstname, lastname, email, password, newAuthid.ID)
		if err != nil {
			return err
		}

		// create confirm token
		confirmTokenDbService := confirm_email_models.NewConfirmTokenDbSerivce(repo.DB)
		var confirmToken *confirm_email_models.ConfirmToken
		var newToken string
		for range 5 {
			token, err := randoms_utils.GenerateConfirmToken()
			if err != nil {
				continue
			}

			hashed := strings_utils.HashToken(token)
			ct, err := confirmTokenDbService.CreateWithTx(tx, hashed)
			if err != nil {
				continue
			}

			confirmToken = ct
			newToken = token
			break
		}

		if confirmToken == nil {
			err := errors.New("Unable to generate new confirm token")
			return err
		}

		userConfirmDbService := confirm_email_models.NewUserConfirmDbService(repo.DB)
		if _, err := userConfirmDbService.CreateWithTx(tx, u.ID, confirmToken.ID); err != nil {
			return err
		}

		if err := tx.Commit(); err != nil {
			return err
		}

		// this is at the end so they wont get populated until everything is committed
		// otherwise they are emmpty
		newAuthidValue = newAuthid.Value
		newConfirmTokenValue = newToken

		return nil
	}()

	if errCreate != nil {
		repo.Logger.Errorln(errCreate)
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
		return
	}

	if os.Getenv("ENVIRONMENT") == "production" {
		if newConfirmTokenValue != "" {
			// send email
			// async call

			go func() {
				if err := email_utils.SendConfirmEmail(email, firstname, newConfirmTokenValue); err != nil {
					repo.Logger.Errorln(err)
				}
			}()
		} else {
			repo.Logger.Errorln("No new confirm token")
		}
	}

	if newAuthidValue == "" {
		flashMessage := "Account created. Please login to continue. Check your email inbox to confirm your email."
		sessionService.ToFlash(flashMessage)
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	flashMessage := fmt.Sprintf("Hello %s! Welcome to %s. Check your email inbox to confirm your email.", strings_utils.Capitalize(firstname), strings_utils.Capitalize(constants.AppName))
	sessionService.Login(newAuthidValue, true)
	sessionService.ToFlash(flashMessage)
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)

}
