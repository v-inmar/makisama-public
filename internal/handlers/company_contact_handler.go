package handlers

import (
	"app/internal/constants"
	company_models "app/internal/models/company"
	email_ban_models "app/internal/models/email_ban"
	captcha_utils "app/internal/utilities/captcha"
	email_utils "app/internal/utilities/email"
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

func (repo *HandlerRepo) GetCompanyContact(w http.ResponseWriter, r *http.Request) {
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

	if currentUser != nil {
		td.Data["Firstname"] = currentUser.Firstname
		td.Data["Lastname"] = currentUser.Lastname
		td.Data["Email"] = strings_utils.ObfuscateEmail(currentUser.Email)
	}
	template_utils.RenderTemplate(w, repo.Templates["contact.html"], td)
}

func (repo *HandlerRepo) PostCompanyContact(w http.ResponseWriter, r *http.Request) {
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

	if err := r.ParseForm(); err != nil {
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusBadRequest, constants.Error400Title, constants.Error400Message)
		return
	}

	var firstname, lastname, email string

	message := r.Form.Get("message")
	if currentUser != nil {
		firstname = currentUser.Firstname
		lastname = currentUser.Lastname
		email = currentUser.Email
	} else {
		firstname = r.FormValue("firstname")
		lastname = r.FormValue("lastname")
		email = r.FormValue("email")
		td.Data["Firstname"] = firstname
		td.Data["Lastname"] = lastname
		td.Data["Email"] = email

		if err := captcha_utils.CheckCaptcha(r.FormValue("cf-turnstile-response")); err != nil {
			td.Data["ErrorCaptcha"] = "Verification failed. Please refresh the page and try again."
			template_utils.RenderTemplate(w, repo.Templates["contact.html"], td)
			return
		}

		email = strings.ToLower(strings.TrimSpace(email))
		if !strings_utils.IsValidEmail(email) {
			td.Data["ErrorEmail"] = "Invalid format."
			template_utils.RenderTemplate(w, repo.Templates["contact.html"], td)
			return
		}

		firstname = strings.TrimSpace(firstname)
		if !strings_utils.IsValidLength(firstname, constants.FirstnameMinLength, constants.FirstnameMaxLength) {
			td.Data["ErrorFirstname"] = fmt.Sprintf("Must be between %d and %d characters long.", constants.FirstnameMinLength, constants.FirstnameMaxLength)
			template_utils.RenderTemplate(w, repo.Templates["contact.html"], td)
			return
		}

		lastname = strings.TrimSpace(lastname)
		if !strings_utils.IsValidLength(lastname, constants.LastnameMinLength, constants.LastnameMaxLength) {
			td.Data["ErrorLastname"] = fmt.Sprintf("Must be between %d and %d characters long.", constants.LastnameMinLength, constants.LastnameMaxLength)
			template_utils.RenderTemplate(w, repo.Templates["contact.html"], td)
			return
		}
	}

	if len(message) < 1 {
		td.Data["ErrorMessage"] = "Please write a message"
		template_utils.RenderTemplate(w, repo.Templates["contact.html"], td)
		return
	}

	if len(message) > 5000 {
		message = strings_utils.Substring(message, 5000)
	}

	contactDbService := company_models.NewContactDbService(repo.DB)
	contact, errCreate := contactDbService.Create(firstname, lastname, email, message)
	if errCreate != nil {
		repo.Logger.Errorln(errCreate)
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
		return
	}

	if os.Getenv("ENVIRONMENT") == "production" {
		go func() {
			if err := email_utils.SendContactEmailToAdmin(contact.ID, message); err != nil {
				repo.Logger.Errorln(err)
			}
		}()
	}

	sessionService.ToFlash("Message submitted. Please allow 14 days for a response.")
	http.Redirect(w, r, "/company/contact", http.StatusSeeOther)
}

// func (repo *CompanyHandlerRepo) Contact(w http.ResponseWriter, r *http.Request) {
// 	rndr := template_utils.NewTemplateRenderer(repo.Session)
// 	sessionService := sessions_utils.NewSessionService(r, repo.Session)
// 	contactDbService := company_models.NewContactDbService(repo.DB)
// 	emailBanDbService := email_ban_models.NewEmailBanDbService(repo.DB)

// 	td := r.Context().Value(constants.TemplateDataContextKey).(*template_utils.TemplateData)

// 	user, err := sessionService.GetUserFromSession(repo.DB)
// 	if err != nil {
// 		if !errors.Is(err, sql.ErrNoRows) && !errors.Is(err, sessions_utils.ErrNoUserInSession) {
// 			repo.Logger.Errorln(err)
// 			rndr.Abort(w, r, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
// 			return
// 		}
// 	}
// 	if user != nil {
// 		ban, err := emailBanDbService.ReadByEmail(user.Email)
// 		if err != nil {
// 			if !errors.Is(err, sql.ErrNoRows) {
// 				repo.Logger.Errorln(err)
// 				rndr.Abort(w, r, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
// 				return
// 			}
// 		}
// 		if ban != nil {
// 			sessionService.DestroySession()
// 			http.Redirect(w, r, r.URL.RequestURI(), http.StatusSeeOther)
// 			return
// 		}
// 	}

// 	// ========= GET ========== //
// 	if r.Method == http.MethodGet {
// 		if user != nil {
// 			td.Data["Firstname"] = user.Firstname
// 			td.Data["Lastname"] = user.Lastname
// 			td.Data["Email"] = strings_utils.ObfuscateEmail(user.Email)
// 		}
// 		rndr.RenderTemplate(w, r, repo.Templates["contact.html"], td)
// 		return
// 	}

// 	// ======== POST ========== //
// 	if r.Method == http.MethodPost {

// 		var firstname, lastname, email string

// 		message := r.FormValue("message")

// 		// check captcha for unauthed users
// 		if user != nil {
// 			firstname = user.Firstname
// 			lastname = user.Lastname
// 			email = user.Email
// 		} else {
// 			firstname = r.FormValue("firstname")
// 			lastname = r.FormValue("lastname")
// 			email = r.FormValue("email")
// 			td.Data["Firstname"] = firstname
// 			td.Data["Lastname"] = lastname
// 			td.Data["Email"] = email

// 			if err := captcha_utils.CheckCaptcha(r.FormValue("cf-turnstile-response")); err != nil {
// 				td.Data["ErrorCaptcha"] = "Verification failed. Please refresh the page and try again."
// 				rndr.RenderTemplate(w, r, repo.Templates["contact.html"], td)
// 				return
// 			}

// 			email = strings.ToLower(strings.TrimSpace(email))
// 			if !strings_utils.IsValidEmail(email) {
// 				td.Data["ErrorEmail"] = "Invalid format."
// 				rndr.RenderTemplate(w, r, repo.Templates["contact.html"], td)
// 				return
// 			}

// 			firstname = strings.TrimSpace(firstname)
// 			if !strings_utils.IsValidLength(firstname, constants.FirstnameMinLength, constants.FirstnameMaxLength) {
// 				td.Data["ErrorFirstname"] = fmt.Sprintf("Must be between %d and %d characters long.", constants.FirstnameMinLength, constants.FirstnameMaxLength)
// 				rndr.RenderTemplate(w, r, repo.Templates["contact.html"], td)
// 				return
// 			}

// 			lastname = strings.TrimSpace(lastname)
// 			if !strings_utils.IsValidLength(lastname, constants.LastnameMinLength, constants.LastnameMaxLength) {
// 				td.Data["ErrorLastname"] = fmt.Sprintf("Must be between %d and %d characters long.", constants.LastnameMinLength, constants.LastnameMaxLength)
// 				rndr.RenderTemplate(w, r, repo.Templates["contact.html"], td)
// 				return
// 			}
// 		}

// 		if len(message) < 1 {
// 			td.Data["ErrorMessage"] = "Please write a message"
// 			rndr.RenderTemplate(w, r, repo.Templates["contact.html"], td)
// 			return
// 		}

// 		if len(message) > 5000 {
// 			message = strings_utils.Substring(message, 5000)
// 		}

// 		contact, errCreate := contactDbService.Create(firstname, lastname, email, message)
// 		if errCreate != nil {
// 			repo.Logger.Errorln(errCreate)
// 			rndr.Abort(w, r, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
// 			return
// 		}

// 		if os.Getenv("ENVIRONMENT") == "production" {
// 			go func() {
// 				if err := email_utils.SendContactEmailToAdmin(contact.ID, message); err != nil {
// 					repo.Logger.Errorln(err)
// 				}
// 			}()
// 		}

// 		sessionService.ToFlash("Message submitted. Please allow 14 days for a response.")
// 		http.Redirect(w, r, "/company/contact", http.StatusSeeOther)
// 		return
// 	}

// 	rndr.Abort(w, r, repo.Templates["error.html"], http.StatusMethodNotAllowed, constants.Error405Title, constants.Error405Message)

// }
