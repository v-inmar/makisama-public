package handlers

import (
	"app/internal/constants"
	board_models "app/internal/models/board"
	email_ban_models "app/internal/models/email_ban"
	feedback_models "app/internal/models/feedback"
	user_models "app/internal/models/user"
	captcha_utils "app/internal/utilities/captcha"
	sessions_utils "app/internal/utilities/sessions"
	strings_utils "app/internal/utilities/strings"
	template_utils "app/internal/utilities/templates"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// i.e. /feedbacks/new?b=1234 (b is the board slug)

func (repo *HandlerRepo) GetFeedbackCreate(w http.ResponseWriter, r *http.Request) {
	td := r.Context().Value(constants.TemplateDataContextKey).(*template_utils.TemplateData)

	sessionService := sessions_utils.NewSessionService(r, repo.Session)

	td.IsAuthenticated = false
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
		td.IsAuthenticated = true
	}

	boardDbService := board_models.NewBoardDbService(repo.DB)
	boardSlug := r.URL.Query().Get("b")
	board, err := boardDbService.ReadBySlug(boardSlug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			template_utils.Abort(w, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
			return
		}
		repo.Logger.Errorln(err)
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
		return
	}

	userDbService := user_models.NewUserDbService(repo.DB)
	// check that board owner has not been deactivated
	boardOwner, err := userDbService.ReadById(board.OwnerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			template_utils.Abort(w, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
			return
		}
		repo.Logger.Errorln(err)
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
		return
	}

	// check board owner band
	ban, err := emailBanDbService.ReadByEmail(boardOwner.Email)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			repo.Logger.Errorln(err)
			template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
			return
		}
	}
	if ban != nil {
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
		return
	}

	if boardOwner.DeactivatedAt.Valid {
		sessionService.ToFlash("Unable to create a feedback for a deactivated board")
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
		return
	}

	td.Data["BoardSlug"] = board.Slug
	td.Data["TitleMinLength"] = constants.FeedbackTitleMinLength
	td.Data["TitleMaxLength"] = constants.FeedbackTitleMaxLength
	td.Data["EmailMaxLength"] = constants.EmailMaxLength
	td.Data["DescriptionMaxLength"] = constants.FeedbackDescriptionMaxLength
	template_utils.RenderTemplate(w, repo.Templates["feedback_new.html"], td)
}

func (repo *HandlerRepo) PostFeedbackCreate(w http.ResponseWriter, r *http.Request) {
	td := r.Context().Value(constants.TemplateDataContextKey).(*template_utils.TemplateData)

	sessionService := sessions_utils.NewSessionService(r, repo.Session)
	td.IsAuthenticated = false
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
		td.IsAuthenticated = true
	}

	boardDbService := board_models.NewBoardDbService(repo.DB)
	boardSlug := r.URL.Query().Get("b")
	board, err := boardDbService.ReadBySlug(boardSlug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			template_utils.Abort(w, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
			return
		}
		repo.Logger.Errorln(err)
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
		return
	}

	userDbService := user_models.NewUserDbService(repo.DB)
	// check that board owner has not been deactivated
	boardOwner, err := userDbService.ReadById(board.OwnerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			template_utils.Abort(w, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
			return
		}
		repo.Logger.Errorln(err)
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
		return
	}

	// check board owner band
	ban, err := emailBanDbService.ReadByEmail(boardOwner.Email)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			repo.Logger.Errorln(err)
			template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
			return
		}
	}
	if ban != nil {
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
		return
	}

	if boardOwner.DeactivatedAt.Valid {
		sessionService.ToFlash("Unable to create a feedback for a deactivated board")
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
		return
	}

	if err := r.ParseForm(); err != nil {
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusBadRequest, constants.Error400Title, constants.Error400Message)
		return
	}

	title := r.FormValue("title")
	description := r.FormValue("description")
	email := r.FormValue("email")

	td.Data["Title"] = title
	td.Data["Description"] = description
	td.Data["Email"] = email

	if currentUser == nil {
		token := r.FormValue("cf-turnstile-response")
		if err := captcha_utils.CheckCaptcha(token); err != nil {
			td.Data["ErrorCaptcha"] = "Verification failed. Please refresh the page and try again."
			template_utils.RenderTemplate(w, repo.Templates["feedback_new.html"], td)
			return
		}
	}

	title = strings.TrimSpace(title)
	if !strings_utils.IsValidLength(title, constants.FeedbackTitleMinLength, constants.FeedbackTitleMaxLength) {
		td.Data["ErrorTitle"] = fmt.Sprintf("Must be between %d and %d characters long", constants.FeedbackTitleMinLength, constants.FeedbackTitleMaxLength)
		template_utils.RenderTemplate(w, repo.Templates["feedback_new.html"], td)
		return
	}

	email = strings.TrimSpace(strings.ToLower(email))
	if len(email) > 0 {
		if !strings_utils.IsValidEmail(email) {
			td.Data["ErrorEmail"] = "Invalid format"
			template_utils.RenderTemplate(w, repo.Templates["feedback_new.html"], td)
			return
		}
	}

	description = strings.TrimSpace(description)
	if len(description) > constants.FeedbackDescriptionMaxLength {
		description = strings_utils.Substring(description, constants.FeedbackDescriptionMaxLength)
	}

	var userId int64
	if currentUser != nil {
		userId = currentUser.ID
	}

	feedbackDbService := feedback_models.NewFeedbackDbService(repo.DB)
	if _, err := feedbackDbService.Create(title, description, email, board.ID, userId); err != nil {
		repo.Logger.Errorln(err)
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
		return
	}

	// send email to board owner
	// NOTE: NOT NOW, CAN'T SEND TOO MANY EMAILS, MIGHT RAN OUT OF EMAIL
	// SINCE WE ARE USING 3RD PARTY

	flashMsg := "Feedback submitted and waiting for board owner's review"
	sessionService.ToFlash(flashMsg)
	url := fmt.Sprintf("/%s?page=1", board.Slug)
	http.Redirect(w, r, url, http.StatusSeeOther)

}

// // For non authenticated users (public)
// // but authenticated users can also use this if they want

// // i.e. /feedbacks/new?b=1234 (b is the board slug)
// func (repo *FeedbackHandlerRepo) CreateFeedback(w http.ResponseWriter, r *http.Request) {
// 	rndr := template_utils.NewTemplateRenderer(repo.Session)
// 	boardDbService := board_models.NewBoardDbService(repo.DB)
// 	sessionService := sessions_utils.NewSessionService(r, repo.Session)
// 	userDbService := user_models.NewUserDbService(repo.DB)
// 	feedbackDbService := feedback_models.NewFeedbackDbService(repo.DB)
// 	emailBanDbService := email_ban_models.NewEmailBanDbService(repo.DB)

// 	td := r.Context().Value(constants.TemplateDataContextKey).(*template_utils.TemplateData)

// 	currentUser, err := sessionService.GetUserFromSession(repo.DB)
// 	if err != nil {
// 		if !errors.Is(err, sql.ErrNoRows) && !errors.Is(err, sessions_utils.ErrNoUserInSession) {
// 			repo.Logger.Errorln(err)
// 			rndr.Abort(w, r, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
// 			return
// 		}
// 		td.IsAuthenticated = false
// 	} else {

// 		ban, err := emailBanDbService.ReadByEmail(currentUser.Email)
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
// 		td.IsAuthenticated = true
// 	}

// 	boardSlug := r.URL.Query().Get("b")
// 	board, err := boardDbService.ReadBySlug(boardSlug)
// 	if err != nil {
// 		if errors.Is(err, sql.ErrNoRows) {
// 			rndr.Abort(w, r, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
// 			return
// 		}
// 		repo.Logger.Errorln(err)
// 		rndr.Abort(w, r, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
// 		return
// 	}

// 	// check that board owner has not been deactivated
// 	boardOwner, err := userDbService.ReadById(board.OwnerID)
// 	if err != nil {
// 		if errors.Is(err, sql.ErrNoRows) {
// 			rndr.Abort(w, r, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
// 			return
// 		}
// 		repo.Logger.Errorln(err)
// 		rndr.Abort(w, r, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
// 		return
// 	}

// 	// check board owner band
// 	ban, err := emailBanDbService.ReadByEmail(boardOwner.Email)
// 	if err != nil {
// 		if !errors.Is(err, sql.ErrNoRows) {
// 			repo.Logger.Errorln(err)
// 			rndr.Abort(w, r, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
// 			return
// 		}
// 	}
// 	if ban != nil {
// 		rndr.Abort(w, r, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
// 		return
// 	}

// 	if boardOwner.DeactivatedAt.Valid {
// 		sessionService.ToFlash("Unable to create a feedback for a deactivated board")
// 		rndr.Abort(w, r, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
// 		return
// 	}

// 	td.Data["BoardSlug"] = board.Slug
// 	td.Data["TitleMinLength"] = constants.FeedbackTitleMinLength
// 	td.Data["TitleMaxLength"] = constants.FeedbackTitleMaxLength
// 	td.Data["EmailMaxLength"] = constants.EmailMaxLength
// 	td.Data["DescriptionMaxLength"] = constants.FeedbackDescriptionMaxLength

// 	// =========== GET ========== //
// 	if r.Method == http.MethodGet {
// 		rndr.RenderTemplate(w, r, repo.Templates["feedback_new.html"], td)
// 		return

// 	}

// 	// ============ POST ============ //
// 	if r.Method == http.MethodPost {
// 		title := r.FormValue("title")
// 		description := r.FormValue("description")
// 		email := r.FormValue("email")

// 		td.Data["Title"] = title
// 		td.Data["Description"] = description
// 		td.Data["Email"] = email

// 		if currentUser == nil {
// 			token := r.FormValue("cf-turnstile-response")
// 			if err := captcha_utils.CheckCaptcha(token); err != nil {
// 				td.Data["ErrorCaptcha"] = "Verification failed. Please refresh the page and try again."
// 				rndr.RenderTemplate(w, r, repo.Templates["feedback_new.html"], td)
// 				return
// 			}
// 		}

// 		title = strings.TrimSpace(title)
// 		if !strings_utils.IsValidLength(title, constants.FeedbackTitleMinLength, constants.FeedbackTitleMaxLength) {
// 			td.Data["ErrorTitle"] = fmt.Sprintf("Must be between %d and %d characters long", constants.FeedbackTitleMinLength, constants.FeedbackTitleMaxLength)
// 			rndr.RenderTemplate(w, r, repo.Templates["feedback_new.html"], td)
// 			return
// 		}

// 		email = strings.TrimSpace(strings.ToLower(email))
// 		if len(email) > 0 {
// 			if !strings_utils.IsValidEmail(email) {
// 				td.Data["ErrorEmail"] = "Invalid format"
// 				rndr.RenderTemplate(w, r, repo.Templates["feedback_new.html"], td)
// 				return
// 			}
// 		}

// 		description = strings.TrimSpace(description)
// 		if len(description) > constants.FeedbackDescriptionMaxLength {
// 			description = strings_utils.Substring(description, constants.FeedbackDescriptionMaxLength)
// 		}

// 		var userId int64
// 		if currentUser != nil {
// 			userId = currentUser.ID
// 		}

// 		if _, err := feedbackDbService.Create(title, description, email, board.ID, userId); err != nil {
// 			repo.Logger.Errorln(err)
// 			rndr.Abort(w, r, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
// 			return
// 		}

// 		// send email to board owner
// 		// NOTE: NOT NOW, CAN'T SEND TOO MANY EMAILS, MIGHT RAN OUT OF EMAIL
// 		// SINCE WE ARE USING 3RD PARTY

// 		flashMsg := "Feedback submitted and waiting for board owner's review"
// 		sessionService.ToFlash(flashMsg)
// 		url := fmt.Sprintf("/%s?page=1", board.Slug)
// 		http.Redirect(w, r, url, http.StatusSeeOther)
// 		return
// 	}

// 	rndr.Abort(w, r, repo.Templates["error.html"], http.StatusMethodNotAllowed, constants.Error405Title, constants.Error405Message)

// }
