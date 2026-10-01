package handlers

import (
	"app/internal/constants"
	confirm_email_models "app/internal/models/confirm_email"
	email_ban_models "app/internal/models/email_ban"
	email_utils "app/internal/utilities/email"
	randoms_utils "app/internal/utilities/randoms"
	sessions_utils "app/internal/utilities/sessions"
	strings_utils "app/internal/utilities/strings"
	template_utils "app/internal/utilities/templates"
	"database/sql"
	"errors"
	"net/http"
	"os"
)

func (repo *HandlerRepo) GetAuthRequestConfirm(w http.ResponseWriter, r *http.Request) {
	td := r.Context().Value(constants.TemplateDataContextKey).(*template_utils.TemplateData)

	sessionService := sessions_utils.NewSessionService(r, repo.Session)
	emailBanDbService := email_ban_models.NewEmailBanDbService(repo.DB)

	currentUser, err := sessionService.GetUserFromSession(repo.DB)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) && !errors.Is(err, sessions_utils.ErrNoUserInSession) {
			repo.Logger.Errorln(err)
			template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
			return
		}
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	// check ban
	ban, err := emailBanDbService.ReadByEmail(currentUser.Email)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			repo.Logger.Errorln(err)
			template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
			return
		}
	}

	if ban != nil {
		// redirect back here with session destroyed
		sessionService.DestroySession()
		http.Redirect(w, r, r.URL.RequestURI(), http.StatusSeeOther)
		return
	}
	td.Data["Email"] = strings_utils.ObfuscateEmail(currentUser.Email)

	template_utils.RenderTemplate(w, repo.Templates["request_confirm.html"], td)
}

func (repo *HandlerRepo) PostAuthRequestConfirm(w http.ResponseWriter, r *http.Request) {
	td := r.Context().Value(constants.TemplateDataContextKey).(*template_utils.TemplateData)

	sessionService := sessions_utils.NewSessionService(r, repo.Session)
	emailBanDbService := email_ban_models.NewEmailBanDbService(repo.DB)

	currentUser, err := sessionService.GetUserFromSession(repo.DB)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) && !errors.Is(err, sessions_utils.ErrNoUserInSession) {
			repo.Logger.Errorln(err)
			template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
			return
		}
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	// check ban
	ban, err := emailBanDbService.ReadByEmail(currentUser.Email)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			repo.Logger.Errorln(err)
			template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
			return
		}
	}

	if ban != nil {
		// redirect back here with session destroyed
		sessionService.DestroySession()
		http.Redirect(w, r, r.URL.RequestURI(), http.StatusSeeOther)
		return
	}

	userConfirmDbService := confirm_email_models.NewUserConfirmDbService(repo.DB)
	userConfirm, err := userConfirmDbService.ReadByUserID(currentUser.ID)
	if err != nil {
		repo.Logger.Errorln(err)
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
		return
	}

	td.Data["Email"] = strings_utils.ObfuscateEmail(currentUser.Email)

	// get the user confirm model
	confirmDbService := confirm_email_models.NewConfirmTokenDbSerivce(repo.DB)
	var token string
	if err := func() error {
		tx, err := repo.DB.Beginx()
		if err != nil {
			return err
		}
		defer tx.Rollback()

		var confirmToken *confirm_email_models.ConfirmToken
		for range 5 {
			tok, err := randoms_utils.GenerateConfirmToken()
			if err != nil {
				continue
			}

			hashed := strings_utils.HashToken(tok)

			c, err := confirmDbService.CreateWithTx(tx, hashed)
			if err != nil {
				continue
			}

			confirmToken = c
			token = tok
			break

		}

		if confirmToken == nil {
			return errors.New("unable to generate new confirm token model")
		}

		if err := userConfirmDbService.UpdateConfirmIDWithTx(tx, userConfirm, confirmToken.ID); err != nil {
			return err
		}

		if err := tx.Commit(); err != nil {
			return err
		}

		return nil

	}(); err != nil {
		repo.Logger.Errorln(err)
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
		return
	}

	if token == "" {
		repo.Logger.Error(errors.New("token is empty"))
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
		return
	}

	if os.Getenv("ENVIRONMENT") == "production" {
		go func() {
			if err := email_utils.SendNewConfirmEmail(currentUser.Email, currentUser.Firstname, token); err != nil {
				repo.Logger.Errorln(err)
			}
		}()
	}

	sessionService.ToFlash("New confirmation email has been sent. Check your inbox or spam folder")
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}
