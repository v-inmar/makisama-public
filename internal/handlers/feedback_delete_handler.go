package handlers

import (
	"app/internal/constants"
	board_models "app/internal/models/board"
	email_ban_models "app/internal/models/email_ban"
	feedback_models "app/internal/models/feedback"
	sessions_utils "app/internal/utilities/sessions"
	template_utils "app/internal/utilities/templates"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"
)

func (repo *HandlerRepo) PostFeedbackDelete(w http.ResponseWriter, r *http.Request) {
	// td := r.Context().Value(constants.TemplateDataContextKey).(*template_utils.TemplateData)

	sessionService := sessions_utils.NewSessionService(r, repo.Session)
	// check there is an actual user
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

	feedbackDbService := feedback_models.NewFeedbackDbService(repo.DB)
	slug := r.URL.Query().Get("f")
	f, err := feedbackDbService.ReadBySlug(slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			template_utils.Abort(w, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
			return
		}
		repo.Logger.Errorf("%v", err)
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
		return
	}

	if f.DeletedAt.Valid {
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
		return
	}

	// check if the feedback had already expired
	// cannot delete expired feedback
	if f.ExpiresAt.Valid {
		if f.ExpiresAt.Time.Before(time.Now().UTC()) {
			template_utils.Abort(w, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
			return
		}
	}

	boardDbService := board_models.NewBoardDbService(repo.DB)
	// get the board and check if the board is still active/not deleted
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

	// check if current user is board owner
	// only board owners can delete feedback
	// no need to check if the board owner is not deleted or deactivated
	// it was done when we grab the user from seesion above
	if currentUser.ID != board.OwnerID {
		template_utils.Abort(
			w,
			repo.Templates["error.html"],
			http.StatusNotFound,
			constants.Error404Title,
			constants.Error404Message,
		)
		return
	}

	if err := feedbackDbService.DeleteFeedback(f); err != nil {
		template_utils.Abort(
			w,
			repo.Templates["error.html"],
			http.StatusInternalServerError,
			constants.Error500Title,
			constants.Error500Message,
		)
		return
	}

	sessionService.ToFlash("Feedback deleted!")
	http.Redirect(
		w,
		r,
		fmt.Sprintf("/%s", board.Slug),
		http.StatusSeeOther,
	)
}

// // Only board owners can delete feedbacks
// // think of it as every feedback is submitted by public/anonymous users
// // even if authenticated users submitted feedback

// // i.e. /feedbacks/delete?f=1234 (f is feedback slug)
// func (repo *FeedbackHandlerRepo) DeleteFeedback(w http.ResponseWriter, r *http.Request) {
// 	sessionService := sessions_utils.NewSessionService(r, repo.Session)
// 	feedbackDbService := feedback_models.NewFeedbackDbService(repo.DB)
// 	boardDbService := board_models.NewBoardDbService(repo.DB)
// 	rndr := template_utils.NewTemplateRenderer(repo.Session)
// 	emailBanDbService := email_ban_models.NewEmailBanDbService(repo.DB)

// 	// check there is an actual user
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

// 	// get the feedback using its slug
// 	slug := r.URL.Query().Get("f")
// 	f, err := feedbackDbService.ReadBySlug(slug)
// 	if err != nil {
// 		if errors.Is(err, sql.ErrNoRows) {
// 			rndr.Abort(w, r, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
// 			return
// 		}
// 		repo.Logger.Errorf("%v", err)
// 		rndr.Abort(w, r, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
// 		return
// 	}

// 	if f.DeletedAt.Valid {
// 		rndr.Abort(w, r, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
// 		return
// 	}

// 	// check if the feedback had already expired
// 	// cannot delete expired feedback
// 	if f.ExpiresAt.Valid {
// 		if f.ExpiresAt.Time.Before(time.Now().UTC()) {
// 			rndr.Abort(w, r, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
// 			return
// 		}
// 	}

// 	// get the board and check if the board is still active/not deleted
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

// 	// check if current user is board owner
// 	// only board owners can delete feedback
// 	// no need to check if the board owner is not deleted or deactivated
// 	// it was done when we grab the user from seesion above
// 	if currentUser.ID != board.OwnerID {
// 		rndr.Abort(
// 			w, r,
// 			repo.Templates["error.html"],
// 			http.StatusNotFound,
// 			constants.Error404Title,
// 			constants.Error404Message,
// 		)
// 		return
// 	}

// 	if err := feedbackDbService.DeleteFeedback(f); err != nil {
// 		rndr.Abort(
// 			w, r,
// 			repo.Templates["error.html"],
// 			http.StatusInternalServerError,
// 			constants.Error500Title,
// 			constants.Error500Message,
// 		)
// 		return
// 	}

// 	sessionService.ToFlash("Feedback deleted!")
// 	http.Redirect(
// 		w,
// 		r,
// 		fmt.Sprintf("/%s", board.Slug),
// 		http.StatusSeeOther,
// 	)
// }
