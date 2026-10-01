package handlers

import (
	"app/internal/constants"
	board_models "app/internal/models/board"
	email_ban_models "app/internal/models/email_ban"
	feedback_models "app/internal/models/feedback"
	date_utils "app/internal/utilities/dates"
	sessions_utils "app/internal/utilities/sessions"
	template_utils "app/internal/utilities/templates"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

func (repo *HandlerRepo) GetFeedbackReview(w http.ResponseWriter, r *http.Request) {
	td := r.Context().Value(constants.TemplateDataContextKey).(*template_utils.TemplateData)

	scopeList := make([]string, 0)
	for k := range constants.FeedbackScopes {
		scopeList = append(scopeList, k)
	}

	resolutionList := make([]string, 0)
	for k := range constants.FeedbackResolutions {
		resolutionList = append(resolutionList, k)
	}
	td.Data["ResolutionList"] = resolutionList
	td.Data["ScopeList"] = scopeList

	sessionService := sessions_utils.NewSessionService(r, repo.Session)
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

	// get the feedback
	feedbackDbService := feedback_models.NewFeedbackDbService(repo.DB)
	feedbackSlug := r.URL.Query().Get("f")
	f, err := feedbackDbService.ReadBySlug(feedbackSlug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			template_utils.Abort(
				w,
				repo.Templates["error.html"],
				http.StatusNotFound,
				constants.Error404Title,
				constants.Error404Message,
			)
			return
		}
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

	if f.DeletedAt.Valid {
		template_utils.Abort(
			w,
			repo.Templates["error.html"],
			http.StatusNotFound,
			constants.Error404Title,
			constants.Error404Message,
		)
		return
	}

	// get the board
	boardDbService := board_models.NewBoardDbService(repo.DB)
	board, err := boardDbService.ReadById(f.BoardID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			template_utils.Abort(
				w,
				repo.Templates["error.html"],
				http.StatusNotFound,
				constants.Error404Title,
				constants.Error404Message,
			)
			return
		}
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

	// no need to check if the board owner user has been deactivated or not
	// since we are checking equal of current user and board owner id.
	// if current user is not active, this page is inaccessible
	if board.OwnerID != currentUser.ID {
		template_utils.Abort(
			w,
			repo.Templates["error.html"],
			http.StatusNotFound,
			constants.Error404Title,
			constants.Error404Message,
		)
		return
	}

	scope := ""
	if f.Scope.Valid {
		scope = f.Scope.String
	}

	resolution := ""
	if f.Resolution.Valid {
		resolution = f.Resolution.String
	}

	description := ""
	if f.Description.Valid {
		description = f.Description.String
	}

	expiresAt := ""
	if f.ExpiresAt.Valid {
		expiresAt = date_utils.FormatNullTime(f.ExpiresAt)
	}

	td.Data["BoardSlug"] = board.Slug
	td.Data["FeedbackSlug"] = f.Slug
	td.Data["Title"] = f.Title
	td.Data["Description"] = description
	td.Data["ExpiresAt"] = expiresAt
	td.Data["CreatedAt"] = date_utils.FormatTime(f.CreatedAt)
	td.Data["Resolution"] = resolution
	td.Data["Scope"] = scope

	template_utils.RenderTemplate(
		w,
		repo.Templates["feedback_review.html"],
		td,
	)
}

func (repo *HandlerRepo) PostFeedbackReview(w http.ResponseWriter, r *http.Request) {
	td := r.Context().Value(constants.TemplateDataContextKey).(*template_utils.TemplateData)

	scopeList := make([]string, 0)
	for k := range constants.FeedbackScopes {
		scopeList = append(scopeList, k)
	}

	resolutionList := make([]string, 0)
	for k := range constants.FeedbackResolutions {
		resolutionList = append(resolutionList, k)
	}
	td.Data["ResolutionList"] = resolutionList
	td.Data["ScopeList"] = scopeList

	sessionService := sessions_utils.NewSessionService(r, repo.Session)
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

	// get the feedback
	feedbackDbService := feedback_models.NewFeedbackDbService(repo.DB)
	feedbackSlug := r.URL.Query().Get("f")
	f, err := feedbackDbService.ReadBySlug(feedbackSlug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			template_utils.Abort(
				w,
				repo.Templates["error.html"],
				http.StatusNotFound,
				constants.Error404Title,
				constants.Error404Message,
			)
			return
		}
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

	if f.DeletedAt.Valid {
		template_utils.Abort(
			w,
			repo.Templates["error.html"],
			http.StatusNotFound,
			constants.Error404Title,
			constants.Error404Message,
		)
		return
	}

	// get the board
	boardDbService := board_models.NewBoardDbService(repo.DB)
	board, err := boardDbService.ReadById(f.BoardID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			template_utils.Abort(
				w,
				repo.Templates["error.html"],
				http.StatusNotFound,
				constants.Error404Title,
				constants.Error404Message,
			)
			return
		}
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

	// no need to check if the board owner user has been deactivated or not
	// since we are checking equal of current user and board owner id.
	// if current user is not active, this page is inaccessible
	if board.OwnerID != currentUser.ID {
		template_utils.Abort(
			w,
			repo.Templates["error.html"],
			http.StatusNotFound,
			constants.Error404Title,
			constants.Error404Message,
		)
		return
	}

	td.Data["BoardSlug"] = board.Slug
	td.Data["FeedbackSlug"] = f.Slug
	td.Data["Title"] = f.Title

	if err := r.ParseForm(); err != nil {
		template_utils.Abort(
			w,
			repo.Templates["error.html"],
			http.StatusBadRequest,
			constants.Error400Title,
			constants.Error400Message,
		)
		return
	}

	scopeForm := strings.ToLower(r.Form.Get("scope"))
	resolutionForm := strings.ToLower(r.Form.Get("resolution"))
	td.Data["Scope"] = scopeForm
	td.Data["Resolution"] = resolutionForm
	if _, exist := constants.FeedbackScopes[scopeForm]; !exist {
		td.Data["ErrorScope"] = "Choose one from the list"
		template_utils.RenderTemplate(w, repo.Templates["feedback_review.html"], td)
		return
	}

	if _, exist := constants.FeedbackResolutions[resolutionForm]; !exist {
		td.Data["ErrorResolution"] = "Choose one from the list"
		template_utils.RenderTemplate(w, repo.Templates["feedback_review.html"], td)
		return
	}

	// if !slices.Contains(constants.FeedbackResolutions, resolutionForm) {
	// 	td.Data["ErrorResolution"] = "Choose one from the list"
	// 	rndr.RenderTemplate(w,r, repo.Templates["feedback_review.html"], td)
	// 	return
	// }

	if err := func() error {
		tx, err := repo.DB.Beginx()
		if err != nil {
			return err
		}
		defer tx.Rollback()

		if err := feedbackDbService.UpdateScopeWihTx(tx, f, scopeForm); err != nil {
			return err
		}

		if err := feedbackDbService.UpdateResolutionWihTx(tx, f, resolutionForm); err != nil {
			return err
		}

		expiry := sql.NullTime{
			Valid: false,
		}
		if err := feedbackDbService.UpdateExpiryWihTx(tx, f, expiry); err != nil {
			return err
		}

		if err := tx.Commit(); err != nil {
			return err
		}

		return nil
	}(); err != nil {
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

	// TODO: send email to feedback owner if email is present

	sessionService.ToFlash("Feedback updated")
	http.Redirect(w, r, fmt.Sprintf("/feedbacks/%s", f.Slug), http.StatusSeeOther)

}

// // onlt authed user and user must be board owner

// // feedbacks/review?f="slug-here"
// func (repo *FeedbackHandlerRepo) ReviewFeedback(w http.ResponseWriter, r *http.Request) {
// 	sessionService := sessions_utils.NewSessionService(r, repo.Session)
// 	boardDbService := board_models.NewBoardDbService(repo.DB)
// 	feedbackDbService := feedback_models.NewFeedbackDbService(repo.DB)
// 	rndr := template_utils.NewTemplateRenderer(repo.Session)
// 	emailBanDbService := email_ban_models.NewEmailBanDbService(repo.DB)

// 	currentUser, err := sessionService.GetUserFromSession(repo.DB)
// 	if err != nil {
// 		if !errors.Is(err, sql.ErrNoRows) && !errors.Is(err, sessions_utils.ErrNoUserInSession) {
// 			repo.Logger.Errorln(err)
// 			rndr.Abort(w, r, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
// 			return
// 		}
// 		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
// 		return
// 	}
// 	ban, err := emailBanDbService.ReadByEmail(currentUser.Email)
// 	if err != nil {
// 		if !errors.Is(err, sql.ErrNoRows) {
// 			repo.Logger.Errorln(err)
// 			rndr.Abort(w, r, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
// 			return
// 		}
// 	}
// 	if ban != nil {
// 		sessionService.DestroySession()
// 		rndr.Abort(w, r, repo.Templates["error.html"], http.StatusForbidden, constants.Error403Title, constants.Error403Message)
// 		return
// 	}

// 	td := r.Context().Value(constants.TemplateDataContextKey).(*template_utils.TemplateData)
// 	td.Data["ResolutionList"] = constants.FeedbackResolutions
// 	td.Data["ScopeList"] = constants.FeedbackScopes

// 	// get the feedback
// 	feedbackSlug := r.URL.Query().Get("f")
// 	f, err := feedbackDbService.ReadBySlug(feedbackSlug)
// 	if err != nil {
// 		if errors.Is(err, sql.ErrNoRows) {
// 			rndr.Abort(
// 				w, r,
// 				repo.Templates["error.html"],
// 				http.StatusNotFound,
// 				constants.Error404Title,
// 				constants.Error404Message,
// 			)
// 			return
// 		}
// 		repo.Logger.Errorf("%v", err)
// 		rndr.Abort(
// 			w, r,
// 			repo.Templates["error.html"],
// 			http.StatusInternalServerError,
// 			constants.Error500Title,
// 			constants.Error500Message,
// 		)
// 		return
// 	}

// 	if f.DeletedAt.Valid {
// 		rndr.Abort(
// 			w, r,
// 			repo.Templates["error.html"],
// 			http.StatusNotFound,
// 			constants.Error404Title,
// 			constants.Error404Message,
// 		)
// 		return
// 	}

// 	// get the board
// 	board, err := boardDbService.ReadById(f.BoardID)
// 	if err != nil {
// 		if errors.Is(err, sql.ErrNoRows) {
// 			rndr.Abort(
// 				w, r,
// 				repo.Templates["error.html"],
// 				http.StatusNotFound,
// 				constants.Error404Title,
// 				constants.Error404Message,
// 			)
// 			return
// 		}
// 		repo.Logger.Errorf("%v", err)
// 		rndr.Abort(
// 			w, r,
// 			repo.Templates["error.html"],
// 			http.StatusInternalServerError,
// 			constants.Error500Title,
// 			constants.Error500Message,
// 		)
// 		return
// 	}

// 	// no need to check if the board owner user has been deactivated or not
// 	// since we are checking equal of current user and board owner id.
// 	// if current user is not active, this page is inaccessible
// 	if board.OwnerID != currentUser.ID {
// 		rndr.Abort(
// 			w, r,
// 			repo.Templates["error.html"],
// 			http.StatusNotFound,
// 			constants.Error404Title,
// 			constants.Error404Message,
// 		)
// 		return
// 	}

// 	scope := ""
// 	if f.Scope.Valid {
// 		scope = f.Scope.String
// 	}

// 	resolution := ""
// 	if f.Resolution.Valid {
// 		resolution = f.Resolution.String
// 	}

// 	description := ""
// 	if f.Description.Valid {
// 		description = f.Description.String
// 	}

// 	expiresAt := ""
// 	if f.ExpiresAt.Valid {
// 		expiresAt = date_utils.FormatNullTime(f.ExpiresAt)
// 	}

// 	td.Data["BoardSlug"] = board.Slug
// 	td.Data["FeedbackSlug"] = f.Slug
// 	td.Data["Title"] = f.Title

// 	// ============= GET =============== //
// 	if r.Method == http.MethodGet {
// 		td.Data["Description"] = description
// 		td.Data["ExpiresAt"] = expiresAt
// 		td.Data["CreatedAt"] = date_utils.FormatTime(f.CreatedAt)
// 		td.Data["Resolution"] = resolution
// 		td.Data["Scope"] = scope

// 		rndr.RenderTemplate(
// 			w, r,
// 			repo.Templates["feedback_review.html"],
// 			td,
// 		)
// 		return

// 	}

// 	// ============= POST ============== //
// 	if r.Method == http.MethodPost {
// 		scopeForm := strings.ToLower(r.FormValue("scope"))
// 		resolutionForm := strings.ToLower(r.FormValue("resolution"))
// 		td.Data["Scope"] = scopeForm
// 		td.Data["Resolution"] = resolutionForm
// 		if _, exist := constants.FeedbackScopes[scopeForm]; !exist {
// 			td.Data["ErrorScope"] = "Choose one from the list"
// 			rndr.RenderTemplate(w, r, repo.Templates["feedback_review.html"], td)
// 			return
// 		}
// 		// if !slices.Contains(constants.FeedbackScopes, scopeForm) {
// 		// 	td.Data["ErrorScope"] = "Choose one from the list"
// 		// 	rndr.RenderTemplate(w,r, repo.Templates["feedback_review.html"], td)
// 		// 	return
// 		// }

// 		if _, exist := constants.FeedbackResolutions[resolutionForm]; !exist {
// 			td.Data["ErrorResolution"] = "Choose one from the list"
// 			rndr.RenderTemplate(w, r, repo.Templates["feedback_review.html"], td)
// 			return
// 		}

// 		// if !slices.Contains(constants.FeedbackResolutions, resolutionForm) {
// 		// 	td.Data["ErrorResolution"] = "Choose one from the list"
// 		// 	rndr.RenderTemplate(w,r, repo.Templates["feedback_review.html"], td)
// 		// 	return
// 		// }

// 		err := func() error {
// 			tx, err := repo.DB.Beginx()
// 			if err != nil {
// 				return err
// 			}
// 			defer tx.Rollback()

// 			if err := feedbackDbService.UpdateScopeWihTx(tx, f, scopeForm); err != nil {
// 				return err
// 			}

// 			if err := feedbackDbService.UpdateResolutionWihTx(tx, f, resolutionForm); err != nil {
// 				return err
// 			}

// 			expiry := sql.NullTime{
// 				Valid: false,
// 			}
// 			if err := feedbackDbService.UpdateExpiryWihTx(tx, f, expiry); err != nil {
// 				return err
// 			}

// 			if err := tx.Commit(); err != nil {
// 				return err
// 			}

// 			return nil
// 		}()

// 		if err != nil {
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

// 		// send email to feedback owner if email is present

// 		sessionService.ToFlash("Feedback updated")
// 		http.Redirect(w, r, fmt.Sprintf("/feedbacks/%s", f.Slug), http.StatusSeeOther)
// 		return

// 	}

// 	rndr.Abort(w, r, repo.Templates["error.html"], http.StatusMethodNotAllowed, constants.Error405Title, constants.Error405Message)

// }
