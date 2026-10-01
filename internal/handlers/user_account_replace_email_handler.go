package handlers

import (
	"app/internal/constants"
	authid_models "app/internal/models/authid"
	confirm_email_models "app/internal/models/confirm_email"
	email_ban_models "app/internal/models/email_ban"
	user_models "app/internal/models/user"
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

	"golang.org/x/crypto/bcrypt"
)

func (repo *HandlerRepo) GetUserAccountReplaceEmail(w http.ResponseWriter, r *http.Request) {
	td := r.Context().Value(constants.TemplateDataContextKey).(*template_utils.TemplateData)
	sessionService := sessions_utils.NewSessionService(r, repo.Session)
	currentUser, err := sessionService.GetUserFromSession(repo.DB)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) && !errors.Is(err, sessions_utils.ErrNoUserInSession) {
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

	template_utils.RenderTemplate(w, repo.Templates["replace_email.html"], td)
}

func (repo *HandlerRepo) PostUserAccountReplaceEmail(w http.ResponseWriter, r *http.Request) {
	td := r.Context().Value(constants.TemplateDataContextKey).(*template_utils.TemplateData)
	sessionService := sessions_utils.NewSessionService(r, repo.Session)
	currentUser, err := sessionService.GetUserFromSession(repo.DB)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) && !errors.Is(err, sessions_utils.ErrNoUserInSession) {
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
			http.StatusInternalServerError,
			constants.Error500Title,
			constants.Error500Message,
		)
		return
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	if err := bcrypt.CompareHashAndPassword([]byte(currentUser.Password), []byte(password)); err != nil {
		td.Data["ErrorPassword"] = "Password not recognized"
		template_utils.RenderTemplate(w, repo.Templates["replace_email.html"], td)
		return
	}

	email = strings.TrimSpace(email)
	if !strings_utils.IsValidLength(email, constants.EmailMinLength, constants.EmailMaxLength) {
		td.Data["ErrorEmail"] = "Invalid format"
		template_utils.RenderTemplate(w, repo.Templates["replace_email.html"], td)
		return
	}

	email = strings.ToLower(email)
	if currentUser.Email == email {
		http.Redirect(w, r, "/accounts", http.StatusSeeOther)
		return
	}

	// check if email is banned
	ban, errBan := emailBanDbService.ReadByEmail(email)
	if errBan != nil {
		if !errors.Is(errBan, sql.ErrNoRows) {
			repo.Logger.Errorln(errBan)
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
		td.Data["ErrorEmail"] = "This email is currently banned"
		template_utils.RenderTemplate(w, repo.Templates["replace_email.html"], td)
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
	}

	if user != nil {
		td.Data["ErrorEmail"] = "Email address already in use"
		template_utils.RenderTemplate(w, repo.Templates["replace_email.html"], td)
		return
	}

	authidDbService := authid_models.NewAuthidDbService(repo.DB)
	userConfirmDbService := confirm_email_models.NewUserConfirmDbService(repo.DB)
	confirmTokenDbService := confirm_email_models.NewConfirmTokenDbSerivce(repo.DB)
	var token string
	if errReplace := func() error {
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

		// add ttl to current authid
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

		// replace old authid with new one
		if err := userDbService.UpdateAuthidIdWithTx(tx, currentUser, newAuthid); err != nil {
			return err
		}

		// replace email
		if err := userDbService.UpdateEmailWithTx(tx, currentUser, email); err != nil {
			return err
		}

		// replace verified at
		verified_at := sql.NullTime{
			Valid: false,
		}
		if err := userDbService.UpdateVerifiedWithTx(tx, currentUser, &verified_at); err != nil {
			return err
		}

		// get user confirm model
		userConfirm, err := userConfirmDbService.ReadByUserIDWithTx(tx, currentUser.ID)
		if err != nil {
			return err
		}

		// create new confirm token
		for range 5 {
			tok, err := randoms_utils.GenerateConfirmToken()
			if err != nil {
				continue
			}

			hashed := strings_utils.HashToken(tok)

			confirmToken, err := confirmTokenDbService.CreateWithTx(tx, hashed)
			if err != nil {
				continue
			}

			// replace the user confirm token id
			if err := userConfirmDbService.UpdateConfirmIDWithTx(tx, userConfirm, confirmToken.ID); err != nil {
				return err
			}

			token = tok
			break
		}

		if token == "" {
			return errors.New("unable to generate confirm email token")
		}

		// commit
		if err := tx.Commit(); err != nil {
			return err
		}

		return nil
	}(); errReplace != nil {
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

	if os.Getenv("ENVIRONMENT") == "production" {
		go func() {
			if err := email_utils.SendReplaceEmail(currentUser.Email, currentUser.Firstname, token); err != nil {
				repo.Logger.Errorln(err)
			}
		}()
	}
	sessionService.DestroySession()
	sessionService.ToFlash("Email address updated. Check your inbox to confirm your email")
	http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
}

// func (repo *UserHandlerRepo) AccountReplaceEmail(w http.ResponseWriter, r *http.Request) {
// 	sessionService := sessions_utils.NewSessionService(r, repo.Session)
// 	rndr := template_utils.NewTemplateRenderer(repo.Session)
// 	userDbService := user_models.NewUserDbService(repo.DB)
// 	authidDbService := authid_models.NewAuthidDbService(repo.DB)
// 	confirmTokenDbService := confirm_email_models.NewConfirmTokenDbSerivce(repo.DB)
// 	userConfirmDbService := confirm_email_models.NewUserConfirmDbService(repo.DB)
// 	emailBanDbService := email_ban_models.NewEmailBanDbService(repo.DB)

// 	td := r.Context().Value(constants.TemplateDataContextKey).(*template_utils.TemplateData)
// 	// td.Data["ErrorEmail"] = ""
// 	// td.Data["ErrorPassword"] = ""

// 	currentUser, err := sessionService.GetUserFromSession(repo.DB)
// 	if err != nil {
// 		if !errors.Is(err, sql.ErrNoRows) && !errors.Is(err, sessions_utils.ErrNoUserInSession) {
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

// 	// =========== GET ========== //
// 	if r.Method == http.MethodGet {
// 		rndr.RenderTemplate(w, r, repo.Templates["replace_email.html"], td)
// 		return
// 	}

// 	// =========== POST ========== //
// 	if r.Method == http.MethodPost {
// 		email := r.FormValue("email")
// 		password := r.FormValue("password")

// 		if err := bcrypt.CompareHashAndPassword([]byte(currentUser.Password), []byte(password)); err != nil {
// 			td.Data["ErrorPassword"] = "Password not recognized"
// 			rndr.RenderTemplate(w, r, repo.Templates["replace_email.html"], td)
// 			return
// 		}

// 		email = strings.TrimSpace(email)
// 		if !strings_utils.IsValidLength(email, constants.EmailMinLength, constants.EmailMaxLength) {
// 			td.Data["ErrorEmail"] = "Invalid format"
// 			rndr.RenderTemplate(w, r, repo.Templates["replace_email.html"], td)
// 			return
// 		}

// 		email = strings.ToLower(email)
// 		if currentUser.Email == email {
// 			http.Redirect(w, r, "/accounts", http.StatusSeeOther)
// 			return
// 		}

// 		// check if email is banned
// 		ban, err := emailBanDbService.ReadByEmail(email)
// 		if err != nil {
// 			if !errors.Is(err, sql.ErrNoRows) {
// 				repo.Logger.Errorln(err)
// 				rndr.Abort(
// 					w, r,
// 					repo.Templates["error.html"],
// 					http.StatusInternalServerError,
// 					constants.Error500Title,
// 					constants.Error500Message,
// 				)
// 				return
// 			}
// 		}
// 		if ban != nil {
// 			td.Data["ErrorEmail"] = "This email is currently banned"
// 			rndr.RenderTemplate(w, r, repo.Templates["replace_email.html"], td)
// 			return
// 		}

// 		user, err := userDbService.ReadByEmail(email)
// 		if err != nil {
// 			if !errors.Is(err, sql.ErrNoRows) {
// 				repo.Logger.Errorln(err)
// 				rndr.Abort(w, r, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
// 				return
// 			}
// 		}

// 		if user != nil {
// 			td.Data["ErrorEmail"] = "Email address already in use"
// 			rndr.RenderTemplate(w, r, repo.Templates["replace_email.html"], td)
// 			return
// 		}

// 		var token string
// 		if errReplace := func() error {
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

// 			// add ttl to current authid
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

// 			// replace old authid with new one
// 			if err := userDbService.UpdateAuthidIdWithTx(tx, currentUser, newAuthid); err != nil {
// 				return err
// 			}

// 			// replace email
// 			if err := userDbService.UpdateEmailWithTx(tx, currentUser, email); err != nil {
// 				return err
// 			}

// 			// replace verified at
// 			verified_at := sql.NullTime{
// 				Valid: false,
// 			}
// 			if err := userDbService.UpdateVerifiedWithTx(tx, currentUser, &verified_at); err != nil {
// 				return err
// 			}

// 			// get user confirm model
// 			userConfirm, err := userConfirmDbService.ReadByUserIDWithTx(tx, currentUser.ID)
// 			if err != nil {
// 				return err
// 			}

// 			// create new confirm token
// 			for range 5 {
// 				tok, err := randoms_utils.GenerateConfirmToken()
// 				if err != nil {
// 					continue
// 				}

// 				hashed := strings_utils.HashToken(tok)

// 				confirmToken, err := confirmTokenDbService.CreateWithTx(tx, hashed)
// 				if err != nil {
// 					continue
// 				}

// 				// replace the user confirm token id
// 				if err := userConfirmDbService.UpdateConfirmIDWithTx(tx, userConfirm, confirmToken.ID); err != nil {
// 					return err
// 				}

// 				token = tok
// 				break
// 			}

// 			if token == "" {
// 				return errors.New("unable to generate confirm email token")
// 			}

// 			// commit
// 			if err := tx.Commit(); err != nil {
// 				return err
// 			}

// 			return nil
// 		}(); errReplace != nil {
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

// 		if os.Getenv("ENVIRONMENT") == "production" {
// 			go func() {
// 				if err := email_utils.SendReplaceEmail(currentUser.Email, currentUser.Firstname, token); err != nil {
// 					repo.Logger.Errorln(err)
// 				}
// 			}()
// 		}
// 		sessionService.DestroySession()
// 		sessionService.ToFlash("Email address updated. Check your inbox to confirm your email")
// 		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
// 		return

// 	}

// 	rndr.Abort(
// 		w, r,
// 		repo.Templates["error.html"],
// 		http.StatusMethodNotAllowed,
// 		constants.Error405Title,
// 		constants.Error405Message,
// 	)

// }
