package handlers

import (
	"app/internal/constants"
	board_models "app/internal/models/board"
	email_ban_models "app/internal/models/email_ban"
	feedback_models "app/internal/models/feedback"
	report_models "app/internal/models/report"
	user_models "app/internal/models/user"
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

func (repo *HandlerRepo) GetReportFeedback(w http.ResponseWriter, r *http.Request) {
	td := r.Context().Value(constants.TemplateDataContextKey).(*template_utils.TemplateData)
	td.Data["ReasonList"] = constants.ReportReasons
	td.Data["EmailMaxLength"] = constants.EmailMaxLength
	td.Data["CommentMaxLength"] = constants.ReportCommentMaxLength

	sessionService := sessions_utils.NewSessionService(r, repo.Session)

	td.IsAuthenticated = false
	currentUser, err := sessionService.GetUserFromSession(repo.DB)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) && !errors.Is(err, sessions_utils.ErrNoUserInSession) {
			repo.Logger.Errorf("%v", err)
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

	feedbackDbService := feedback_models.NewFeedbackDbService(repo.DB)
	slug := r.URL.Query().Get("f")
	feedback, err := feedbackDbService.ReadBySlug(slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			template_utils.Abort(w, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
			return
		}
		repo.Logger.Errorf("%v", err)
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
		return
	}

	if feedback.DeletedAt.Valid {
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
		return
	}

	td.Data["FeedbackSlug"] = feedback.Slug
	td.Data["FeedbackTitle"] = feedback.Title

	// check if board that the feedback belongs to still exist
	boardDbService := board_models.NewBoardDbService(repo.DB)
	board, err := boardDbService.ReadById(feedback.BoardID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			template_utils.Abort(w, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
			return
		}
		repo.Logger.Errorf("%v", err)
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
		return
	}

	// check that board owner still exist (not deactivated)
	userDbService := user_models.NewUserDbService(repo.DB)
	boardOwnerUser, err := userDbService.ReadById(board.OwnerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			template_utils.Abort(w, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
			return
		}
		repo.Logger.Errorf("%v", err)
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
		return
	}

	ban, err := emailBanDbService.ReadByEmail(boardOwnerUser.Email)
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

	if boardOwnerUser.DeactivatedAt.Valid {
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
		return
	}

	template_utils.RenderTemplate(w, repo.Templates["report_feedback.html"], td)
}

func (repo *HandlerRepo) PostReportFeedback(w http.ResponseWriter, r *http.Request) {
	td := r.Context().Value(constants.TemplateDataContextKey).(*template_utils.TemplateData)
	td.Data["ReasonList"] = constants.ReportReasons
	td.Data["EmailMaxLength"] = constants.EmailMaxLength
	td.Data["CommentMaxLength"] = constants.ReportCommentMaxLength

	sessionService := sessions_utils.NewSessionService(r, repo.Session)

	td.IsAuthenticated = false
	currentUser, err := sessionService.GetUserFromSession(repo.DB)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) && !errors.Is(err, sessions_utils.ErrNoUserInSession) {
			repo.Logger.Errorf("%v", err)
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

	feedbackDbService := feedback_models.NewFeedbackDbService(repo.DB)
	slug := r.URL.Query().Get("f")
	feedback, err := feedbackDbService.ReadBySlug(slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			template_utils.Abort(w, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
			return
		}
		repo.Logger.Errorf("%v", err)
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
		return
	}

	if feedback.DeletedAt.Valid {
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
		return
	}

	td.Data["FeedbackSlug"] = feedback.Slug
	td.Data["FeedbackTitle"] = feedback.Title

	// check if board that the feedback belongs to still exist
	boardDbService := board_models.NewBoardDbService(repo.DB)
	board, err := boardDbService.ReadById(feedback.BoardID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			template_utils.Abort(w, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
			return
		}
		repo.Logger.Errorf("%v", err)
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
		return
	}

	// check that board owner still exist (not deactivated)
	userDbService := user_models.NewUserDbService(repo.DB)
	boardOwnerUser, err := userDbService.ReadById(board.OwnerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			template_utils.Abort(w, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
			return
		}
		repo.Logger.Errorf("%v", err)
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
		return
	}

	ban, err := emailBanDbService.ReadByEmail(boardOwnerUser.Email)
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

	if boardOwnerUser.DeactivatedAt.Valid {
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
		return
	}

	if err := r.ParseForm(); err != nil {
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusBadRequest, constants.Error400Title, constants.Error400Message)
		return
	}

	reason := r.Form.Get("reason")
	comment := r.Form.Get("comment")
	email := r.Form.Get("email")

	td.Data["Reason"] = reason
	td.Data["Comment"] = comment
	td.Data["Email"] = email

	if currentUser != nil {
		email = currentUser.Email
	}

	if currentUser == nil {
		token := r.FormValue("cf-turnstile-response")
		if err := captcha_utils.CheckCaptcha(token); err != nil {
			td.Data["ErrorCaptcha"] = "Verification failed. Please refresh the page and try again."
			template_utils.RenderTemplate(w, repo.Templates["report_feedback.html"], td)
			return
		}
	}

	if _, exist := constants.ReportReasons[reason]; !exist {
		td.Data["ErrorReason"] = "Please choose one from the list"
		template_utils.RenderTemplate(w, repo.Templates["report_feedback.html"], td)
		return
	}

	if currentUser == nil {
		email = strings.TrimSpace(email)
		if len(email) > 0 {
			if len(email) > constants.EmailMaxLength {
				td.Data["ErrorEmail"] = fmt.Sprintf("Email address exceeded max length of %d", constants.EmailMaxLength)
				template_utils.RenderTemplate(w, repo.Templates["report_feedback.html"], td)
				return
			}

			if !strings_utils.IsValidEmail(email) {
				td.Data["ErrorEmail"] = "Invalid format"
				template_utils.RenderTemplate(w, repo.Templates["report_feedback.html"], td)
				return
			}
		}
	}

	if len(comment) > constants.ReportCommentMaxLength {
		comment = strings_utils.Substring(comment, constants.ReportCommentMaxLength)
	}

	reportDbSerivce := report_models.NewReportFeedbackDbService(repo.DB)
	reportFeedback, err := reportDbSerivce.Create(reason, comment, email, feedback.ID)
	if err != nil {
		repo.Logger.Errorf("%v", err)
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
		return
	}

	if os.Getenv("ENVIRONMENT") == "production" {
		go func() {
			if err := email_utils.SendFeedbackReportEmailToAdmin(reason, feedback.ID, board.ID, reportFeedback.ID); err != nil {
				repo.Logger.Errorln(err)
			}
		}()
	}

	flashMsg := "Feedback reported! Please allow up to 14 days for the investigation to be completed and any necessary action to be taken"
	if email != "" {
		flashMsg = "Feedback reported. Please allow up to 14 days for the investigation to be completed and any necessary action to be taken. We will notify you by email of any actions taken"
	}
	sessionService.ToFlash(flashMsg)
	url := fmt.Sprintf("/report/feedbacks?f=%s", feedback.Slug)
	http.Redirect(w, r, url, http.StatusSeeOther)

}

// // both auth and public user
// // Note: feedback owners can report their own feedback since they cannot delete it
// // /report/feedbacks?f=123abc (f is feedback slug)
// func (repo *ReportHandlerRepo) ReportFeedback(w http.ResponseWriter, r *http.Request) {
// 	sessionService := sessions_utils.NewSessionService(r, repo.Session)
// 	boardDbService := board_models.NewBoardDbService(repo.DB)
// 	feedbackDbService := feedback_models.NewFeedbackDbService(repo.DB)
// 	userDbService := user_models.NewUserDbService(repo.DB)
// 	reportDbSerivce := report_models.NewReportFeedbackDbService(repo.DB)
// 	rndr := template_utils.NewTemplateRenderer(repo.Session)
// 	emailBanDbService := email_ban_models.NewEmailBanDbService(repo.DB)

// 	td := r.Context().Value(constants.TemplateDataContextKey).(*template_utils.TemplateData)
// 	td.Data["ReasonList"] = constants.ReportReasons
// 	td.Data["EmailMaxLength"] = constants.EmailMaxLength
// 	td.Data["CommentMaxLength"] = constants.ReportCommentMaxLength
// 	td.IsAuthenticated = false

// 	currentUser, err := sessionService.GetUserFromSession(repo.DB)
// 	if err != nil {
// 		if !errors.Is(err, sql.ErrNoRows) && !errors.Is(err, sessions_utils.ErrNoUserInSession) {
// 			repo.Logger.Errorf("%v", err)
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

// 	if currentUser != nil {
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

// 	slug := r.URL.Query().Get("f")
// 	feedback, err := feedbackDbService.ReadBySlug(slug)
// 	if err != nil {
// 		if errors.Is(err, sql.ErrNoRows) {
// 			rndr.Abort(w, r, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
// 			return
// 		}
// 		repo.Logger.Errorf("%v", err)
// 		rndr.Abort(w, r, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
// 		return
// 	}

// 	if feedback.DeletedAt.Valid {
// 		rndr.Abort(w, r, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
// 		return
// 	}

// 	td.Data["FeedbackSlug"] = feedback.Slug
// 	td.Data["FeedbackTitle"] = feedback.Title

// 	// check if board that the feedback belongs to still exist
// 	board, err := boardDbService.ReadById(feedback.BoardID)
// 	if err != nil {
// 		if errors.Is(err, sql.ErrNoRows) {
// 			rndr.Abort(w, r, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
// 			return
// 		}
// 		repo.Logger.Errorf("%v", err)
// 		rndr.Abort(w, r, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
// 		return
// 	}

// 	// check that board owner still exist (not deactivated)
// 	boardOwnerUser, err := userDbService.ReadById(board.OwnerID)
// 	if err != nil {
// 		if errors.Is(err, sql.ErrNoRows) {
// 			rndr.Abort(w, r, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
// 			return
// 		}
// 		repo.Logger.Errorf("%v", err)
// 		rndr.Abort(w, r, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
// 		return
// 	}

// 	ban, err := emailBanDbService.ReadByEmail(boardOwnerUser.Email)
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

// 	if boardOwnerUser.DeactivatedAt.Valid {
// 		rndr.Abort(w, r, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
// 		return
// 	}

// 	// =============== GET =============== //
// 	if r.Method == http.MethodGet {
// 		rndr.RenderTemplate(w, r, repo.Templates["report_feedback.html"], td)
// 		return
// 	}

// 	// ============== POST =============== //
// 	if r.Method == http.MethodPost {
// 		reason := r.FormValue("reason")
// 		comment := r.FormValue("comment")
// 		email := r.FormValue("email")

// 		td.Data["Reason"] = reason
// 		td.Data["Comment"] = comment
// 		td.Data["Email"] = email

// 		if currentUser != nil {
// 			email = currentUser.Email
// 		}

// 		if currentUser == nil {
// 			token := r.FormValue("cf-turnstile-response")
// 			if err := captcha_utils.CheckCaptcha(token); err != nil {
// 				td.Data["ErrorCaptcha"] = "Verification failed. Please refresh the page and try again."
// 				rndr.RenderTemplate(w, r, repo.Templates["report_feedback.html"], td)
// 				return
// 			}
// 		}

// 		if _, exist := constants.ReportReasons[reason]; !exist {
// 			td.Data["ErrorReason"] = "Please choose one from the list"
// 			rndr.RenderTemplate(w, r, repo.Templates["report_feedback.html"], td)
// 			return
// 		}

// 		if currentUser == nil {
// 			email = strings.TrimSpace(email)
// 			if len(email) > 0 {
// 				if len(email) > constants.EmailMaxLength {
// 					td.Data["ErrorEmail"] = fmt.Sprintf("Email address exceeded max length of %d", constants.EmailMaxLength)
// 					rndr.RenderTemplate(w, r, repo.Templates["report_feedback.html"], td)
// 					return
// 				}

// 				if !strings_utils.IsValidEmail(email) {
// 					td.Data["ErrorEmail"] = "Invalid format"
// 					rndr.RenderTemplate(w, r, repo.Templates["report_feedback.html"], td)
// 					return
// 				}
// 			}
// 		}

// 		if len(comment) > constants.ReportCommentMaxLength {
// 			comment = strings_utils.Substring(comment, constants.ReportCommentMaxLength)
// 		}

// 		reportFeedback, err := reportDbSerivce.Create(reason, comment, email, feedback.ID)
// 		if err != nil {
// 			repo.Logger.Errorf("%v", err)
// 			rndr.Abort(w, r, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
// 			return
// 		}

// 		if os.Getenv("ENVIRONMENT") == "production" {
// 			go func() {
// 				if err := email_utils.SendFeedbackReportEmailToAdmin(reason, feedback.ID, board.ID, reportFeedback.ID); err != nil {
// 					repo.Logger.Errorln(err)
// 				}
// 			}()
// 		}

// 		flashMsg := "Feedback reported! Please allow up to 14 days for the investigation to be completed and any necessary action to be taken"
// 		if email != "" {
// 			flashMsg = "Feedback reported. Please allow up to 14 days for the investigation to be completed and any necessary action to be taken. We will notify you by email of any actions taken"
// 		}
// 		sessionService.ToFlash(flashMsg)
// 		url := fmt.Sprintf("/report/feedbacks?f=%s", feedback.Slug)
// 		http.Redirect(w, r, url, http.StatusSeeOther)
// 		return
// 	}

// 	rndr.Abort(w, r, repo.Templates["error.html"], http.StatusMethodNotAllowed, constants.Error405Title, constants.Error405Message)
// }
