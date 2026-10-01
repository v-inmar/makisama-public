package handlers

import (
	"app/internal/constants"
	email_ban_models "app/internal/models/email_ban"
	user_models "app/internal/models/user"
	sessions_utils "app/internal/utilities/sessions"
	strings_utils "app/internal/utilities/strings"
	template_utils "app/internal/utilities/templates"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

func (repo *HandlerRepo) GetUserAccountEditName(w http.ResponseWriter, r *http.Request) {
	td := r.Context().Value(constants.TemplateDataContextKey).(*template_utils.TemplateData)
	td.Data["FirstnameMinLength"] = constants.FirstnameMinLength
	td.Data["FirstnameMaxLength"] = constants.FirstnameMaxLength
	td.Data["LastnameMinLength"] = constants.LastnameMinLength
	td.Data["LastnameMaxLength"] = constants.LastnameMaxLength

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

	td.Data["Firstname"] = currentUser.Firstname
	td.Data["Lastname"] = currentUser.Lastname

	template_utils.RenderTemplate(
		w,
		repo.Templates["edit_name.html"],
		td,
	)

}

func (repo *HandlerRepo) PostUserAccountEditName(w http.ResponseWriter, r *http.Request) {
	td := r.Context().Value(constants.TemplateDataContextKey).(*template_utils.TemplateData)
	td.Data["FirstnameMinLength"] = constants.FirstnameMinLength
	td.Data["FirstnameMaxLength"] = constants.FirstnameMaxLength
	td.Data["LastnameMinLength"] = constants.LastnameMinLength
	td.Data["LastnameMaxLength"] = constants.LastnameMaxLength

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

	if err := r.ParseForm(); err != nil {
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusBadRequest, constants.Error400Title, constants.Error400Message)
		return
	}

	firstname := r.Form.Get("firstname")
	lastname := r.Form.Get("lastname")

	td.Data["Firstname"] = firstname
	td.Data["Lastname"] = lastname

	firstname = strings.TrimSpace(firstname)
	if !strings_utils.IsValidLength(firstname, constants.FirstnameMinLength, constants.FirstnameMaxLength) {
		td.Data["ErrorFirstname"] = fmt.Sprintf("Must be between %d and %d characters long.", constants.FirstnameMinLength, constants.FirstnameMaxLength)
		template_utils.RenderTemplate(w, repo.Templates["edit_name.html"], td)
		return
	}

	lastname = strings.TrimSpace(lastname)
	if !strings_utils.IsValidLength(lastname, constants.LastnameMinLength, constants.LastnameMaxLength) {
		td.Data["ErrorLastname"] = fmt.Sprintf("Must be between %d and %d characters long.", constants.LastnameMinLength, constants.LastnameMaxLength)
		template_utils.RenderTemplate(w, repo.Templates["edit_name.html"], td)
		return
	}

	userDbService := user_models.NewUserDbService(repo.DB)
	if err := func() error {
		tx, err := repo.DB.Beginx()
		if err != nil {
			return err
		}
		defer tx.Rollback()

		if err := userDbService.UpdateFirstnameWithTx(tx, currentUser, firstname); err != nil {
			return err
		}

		if err := userDbService.UpdateLastnameWithTx(tx, currentUser, lastname); err != nil {
			return err
		}

		if err := tx.Commit(); err != nil {
			return err
		}
		return nil
	}(); err != nil {
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

	sessionService.ToFlash("Name updated!")
	http.Redirect(w, r, "/accounts", http.StatusSeeOther)

}

// func (repo *UserHandlerRepo) AccountEditName(w http.ResponseWriter, r *http.Request) {
// 	sessionService := sessions_utils.NewSessionService(r, repo.Session)
// 	userDbService := user_models.NewUserDbService(repo.DB)
// 	rndr := template_utils.NewTemplateRenderer(repo.Session)
// 	emailBanDbService := email_ban_models.NewEmailBanDbService(repo.DB)

// 	td := r.Context().Value(constants.TemplateDataContextKey).(*template_utils.TemplateData)
// 	td.Data["FirstnameMinLength"] = constants.FirstnameMinLength
// 	td.Data["FirstnameMaxLength"] = constants.FirstnameMaxLength
// 	td.Data["LastnameMinLength"] = constants.LastnameMinLength
// 	td.Data["LastnameMaxLength"] = constants.LastnameMaxLength

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

// 	// ============= GET =========== //
// 	if r.Method == http.MethodGet {
// 		td.Data["Firstname"] = currentUser.Firstname
// 		td.Data["Lastname"] = currentUser.Lastname

// 		rndr.RenderTemplate(
// 			w, r,
// 			repo.Templates["edit_name.html"],
// 			td,
// 		)
// 		return
// 	}

// 	// ============== POST ========== //
// 	if r.Method == http.MethodPost {
// 		firstname := r.FormValue("firstname")
// 		lastname := r.FormValue("lastname")

// 		td.Data["Firstname"] = firstname
// 		td.Data["Lastname"] = lastname

// 		firstname = strings.TrimSpace(firstname)
// 		if !strings_utils.IsValidLength(firstname, constants.FirstnameMinLength, constants.FirstnameMaxLength) {
// 			td.Data["ErrorFirstname"] = fmt.Sprintf("Must be between %d and %d characters long.", constants.FirstnameMinLength, constants.FirstnameMaxLength)
// 			rndr.RenderTemplate(w, r, repo.Templates["edit_name.html"], td)
// 			return
// 		}

// 		lastname = strings.TrimSpace(lastname)
// 		if !strings_utils.IsValidLength(lastname, constants.LastnameMinLength, constants.LastnameMaxLength) {
// 			td.Data["ErrorLastname"] = fmt.Sprintf("Must be between %d and %d characters long.", constants.LastnameMinLength, constants.LastnameMaxLength)
// 			rndr.RenderTemplate(w, r, repo.Templates["edit_name.html"], td)
// 			return
// 		}

// 		if err := func() error {
// 			tx, err := repo.DB.Beginx()
// 			if err != nil {
// 				return err
// 			}
// 			defer tx.Rollback()

// 			if err := userDbService.UpdateFirstnameWithTx(tx, currentUser, firstname); err != nil {
// 				return err
// 			}

// 			if err := userDbService.UpdateLastnameWithTx(tx, currentUser, lastname); err != nil {
// 				return err
// 			}

// 			if err := tx.Commit(); err != nil {
// 				return err
// 			}
// 			return nil
// 		}(); err != nil {
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

// 		sessionService.ToFlash("Name updated!")
// 		http.Redirect(w, r, "/accounts", http.StatusSeeOther)
// 		return
// 	}

// 	rndr.Abort(w, r, repo.Templates["error.html"], http.StatusMethodNotAllowed, constants.Error405Title, constants.Error405Message)

// }
