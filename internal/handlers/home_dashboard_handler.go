package handlers

import (
	"app/internal/constants"
	board_models "app/internal/models/board"
	email_ban_models "app/internal/models/email_ban"
	feedback_models "app/internal/models/feedback"
	sessions_utils "app/internal/utilities/sessions"
	strings_utils "app/internal/utilities/strings"
	template_utils "app/internal/utilities/templates"
	"database/sql"
	"errors"
	"net/http"
)

func (repo *HandlerRepo) GetHomeDashboard(w http.ResponseWriter, r *http.Request) {
	td := r.Context().Value(constants.TemplateDataContextKey).(*template_utils.TemplateData)

	sessionService := sessions_utils.NewSessionService(r, repo.Session)
	boardDbService := board_models.NewBoardDbService(repo.DB)
	feedbackDbService := feedback_models.NewFeedbackDbService(repo.DB)
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

	td.Data["Name"] = strings_utils.Capitalize(currentUser.Firstname)

	boardCount, err := boardDbService.ReadCountActiveBoardsByUserId(currentUser.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			boardCount = 0
		} else {
			repo.Logger.Errorf("%v", err)
			template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
			return
		}
	}
	td.Data["BoardCount"] = boardCount

	feedbackCount, err := feedbackDbService.ReadCountFeedbacksByUserId(currentUser.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			feedbackCount = 0
		} else {
			repo.Logger.Errorf("%v", err)
			template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
			return
		}
	}
	td.Data["FeedbackCount"] = feedbackCount

	// Get boards
	template_utils.RenderTemplate(
		w,
		repo.Templates["dashboard.html"],
		td,
	)
}
