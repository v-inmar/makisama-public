package handlers

import (
	"app/internal/constants"
	board_models "app/internal/models/board"
	email_ban_models "app/internal/models/email_ban"
	feedback_models "app/internal/models/feedback"
	user_models "app/internal/models/user"
	date_utils "app/internal/utilities/dates"
	sessions_utils "app/internal/utilities/sessions"
	template_utils "app/internal/utilities/templates"
	"database/sql"
	"errors"
	"net/http"
)

func (repo *HandlerRepo) GetFeedbackRead(w http.ResponseWriter, r *http.Request) {
	td := r.Context().Value(constants.TemplateDataContextKey).(*template_utils.TemplateData)
	td.IsAuthenticated = false
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
		td.IsAuthenticated = true
	}

	feedbackDbService := feedback_models.NewFeedbackDbService(repo.DB)
	slug := r.PathValue("slug")
	f, err := feedbackDbService.ReadBySlug(slug)
	td.Data["FeedbackSlug"] = f.Slug

	// check if feedback has been deleted
	if f.DeletedAt.Valid {
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
		return
	}

	// check if it has expiry date - means its not publicly viewable
	if f.ExpiresAt.Valid {
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
		return
	}

	// no scope or no resolution is not viewable by public
	if !f.Scope.Valid || !f.Resolution.Valid {
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
		return
	}

	description := ""
	if f.Description.Valid {
		description = f.Description.String
	}

	td.Data["Title"] = f.Title
	td.Data["Description"] = description
	td.Data["Scope"] = f.Scope.String
	td.Data["Resolution"] = f.Resolution.String
	td.Data["CreatedAt"] = date_utils.FormatTime(f.CreatedAt)

	// check board still exist
	boardDbService := board_models.NewBoardDbService(repo.DB)
	board, err := boardDbService.ReadById(f.BoardID)
	if err != nil {

		if errors.Is(err, sql.ErrNoRows) {
			template_utils.Abort(w, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
			return
		}
		repo.Logger.Errorln(err)
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
		return
	}
	td.Data["BoardSlug"] = board.Slug

	// check the board owner is not deactivated or deleted
	userDbService := user_models.NewUserDbService(repo.DB)
	userBoardOwner, err := userDbService.ReadById(board.OwnerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			template_utils.Abort(w, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
			return
		}
		repo.Logger.Errorln(err)
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
		return
	}
	ban, err := emailBanDbService.ReadByEmail(userBoardOwner.Email)
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

	if userBoardOwner.DeactivatedAt.Valid {
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
		return
	}

	td.Data["IsBoardOwner"] = false
	// check if current user is board owner
	if currentUser != nil {
		if userBoardOwner.ID == currentUser.ID {
			td.Data["IsBoardOwner"] = true
		}
	}

	template_utils.RenderTemplate(w, repo.Templates["feedback_read.html"], td)

}

// // auth and public users

// // /feedbacks/{slug}
// func (repo *FeedbackHandlerRepo) Feedback(w http.ResponseWriter, r *http.Request) {
// 	rndr := template_utils.NewTemplateRenderer(repo.Session)
// 	feedbackDbService := feedback_models.NewFeedbackDbService(repo.DB)
// 	boardDbService := board_models.NewBoardDbService(repo.DB)
// 	userDbService := user_models.NewUserDbService(repo.DB)
// 	sessionService := sessions_utils.NewSessionService(r, repo.Session)
// 	emailBanDbService := email_ban_models.NewEmailBanDbService(repo.DB)

// 	if r.Method != http.MethodGet {
// 		rndr.Abort(w, r, repo.Templates["error.html"], http.StatusMethodNotAllowed, constants.Error405Title, constants.Error405Message)
// 		return
// 	}

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

// 	slug := r.PathValue("slug")
// 	f, err := feedbackDbService.ReadBySlug(slug)
// 	if err != nil {
// 		if errors.Is(err, sql.ErrNoRows) {
// 			rndr.Abort(w, r, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
// 			return
// 		}
// 		repo.Logger.Errorln(err)
// 		rndr.Abort(w, r, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
// 		return
// 	}
// 	td.Data["FeedbackSlug"] = f.Slug

// 	// check if feedback has been deleted
// 	if f.DeletedAt.Valid {
// 		rndr.Abort(w, r, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
// 		return
// 	}

// 	// check if it has expiry date - means its not publicly viewable
// 	if f.ExpiresAt.Valid {
// 		rndr.Abort(w, r, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
// 		return
// 	}

// 	// no scope or no resolution is not viewable by public
// 	if !f.Scope.Valid || !f.Resolution.Valid {
// 		rndr.Abort(w, r, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
// 		return
// 	}

// 	description := ""
// 	if f.Description.Valid {
// 		description = f.Description.String
// 	}

// 	td.Data["Title"] = f.Title
// 	td.Data["Description"] = description
// 	td.Data["Scope"] = f.Scope.String
// 	td.Data["Resolution"] = f.Resolution.String
// 	td.Data["CreatedAt"] = date_utils.FormatTime(f.CreatedAt)

// 	// check board still exist
// 	board, err := boardDbService.ReadById(f.BoardID)
// 	if err != nil {
// 		if errors.Is(err, sql.ErrNoRows) {
// 			rndr.Abort(w, r, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
// 			return
// 		}
// 		repo.Logger.Errorln(err)
// 		rndr.Abort(w, r, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
// 		return
// 	}
// 	td.Data["BoardSlug"] = board.Slug

// 	// check the board owner is not deactivated or deleted
// 	userBoardOwner, err := userDbService.ReadById(board.OwnerID)
// 	if err != nil {
// 		if errors.Is(err, sql.ErrNoRows) {
// 			rndr.Abort(w, r, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
// 			return
// 		}
// 		repo.Logger.Errorln(err)
// 		rndr.Abort(w, r, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
// 		return
// 	}
// 	ban, err := emailBanDbService.ReadByEmail(userBoardOwner.Email)
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

// 	if userBoardOwner.DeactivatedAt.Valid {
// 		rndr.Abort(w, r, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
// 		return
// 	}

// 	td.Data["IsBoardOwner"] = false
// 	// check if current user is board owner
// 	if currentUser != nil {
// 		if userBoardOwner.ID == currentUser.ID {
// 			td.Data["IsBoardOwner"] = true
// 		}
// 	}

// 	rndr.RenderTemplate(w, r, repo.Templates["feedback_read.html"], td)
// }
