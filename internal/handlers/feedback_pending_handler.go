package handlers

import (
	"app/internal/constants"
	dataaccess "app/internal/data_access"
	board_models "app/internal/models/board"
	email_ban_models "app/internal/models/email_ban"
	feedback_models "app/internal/models/feedback"
	number_utils "app/internal/utilities/number"
	sessions_utils "app/internal/utilities/sessions"
	template_utils "app/internal/utilities/templates"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

func (repo *HandlerRepo) GetFeedbackPending(w http.ResponseWriter, r *http.Request) {
	td := r.Context().Value(constants.TemplateDataContextKey).(*template_utils.TemplateData)

	sessionService := sessions_utils.NewSessionService(r, repo.Session)
	currentUser, err := sessionService.GetUserFromSession(repo.DB)
	if err != nil {
		if !errors.Is(err, sessions_utils.ErrNoUserInSession) && !errors.Is(err, sql.ErrNoRows) {
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
		template_utils.Abort(
			w,
			repo.Templates["error.html"],
			http.StatusForbidden,
			constants.Error403Title,
			constants.Error403Message,
		)
		return
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
	td.Data["BoardSlug"] = board.Slug

	// check if current user is board owner
	if currentUser.ID != board.OwnerID {
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
		return
	}

	feedbackDbService := feedback_models.NewFeedbackDbService(repo.DB)
	totalPendingCount, err := feedbackDbService.ReadPendingTotalCountFeedbacksByBoardId(board.ID)
	if err != nil {
		repo.Logger.Errorf("%q", err)
		template_utils.Abort(
			w,
			repo.Templates["error.html"],
			http.StatusInternalServerError,
			constants.Error500Title,
			constants.Error500Message,
		)
		return
	}
	td.Data["TotalPendingCount"] = totalPendingCount

	totalCountOfPages := int((totalPendingCount + constants.PerPageConstant - 1) / constants.PerPageConstant)

	pageSeq := make([]int, 0)
	if totalCountOfPages > 0 {
		pageSeq = number_utils.MakeSliceFromInt(totalCountOfPages)
	}
	td.Data["PageSequence"] = pageSeq

	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		page = 1
	}

	redirect := false
	switch {
	case totalCountOfPages == 0:
		// doesn't matter the page number or whether there is page number
		// no feedbacks anyways
		// this should help with losing of flash message on redirect
		redirect = false
	case totalCountOfPages != 0 && page < 1:
		redirect = true
		page = 1
	case totalCountOfPages != 0 && page > totalCountOfPages:
		redirect = true
		page = totalCountOfPages
	case totalCountOfPages > 0 && page <= totalCountOfPages:
		redirect = false
	default:
		redirect = false
	}

	if redirect {
		url := fmt.Sprintf("/feedbacks/pending?b=%s&page=%d", board.Slug, page)
		http.Redirect(w, r, url, http.StatusSeeOther)
		return
	}
	td.Data["CurrentPage"] = page

	feedbacks, err := feedbackDbService.ReadPaginatedPendingFeedbacksByBoardId(board.ID, page, constants.PerPageConstant)
	if err != nil {
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

	fs := make([]dataaccess.PendingFeedbackStrippedDao, 0)
	if len(feedbacks) > 0 {
		for _, f := range feedbacks {
			desc := ""
			if f.Description.Valid {
				desc = f.Description.String
			}
			d := dataaccess.PendingFeedbackStrippedDao{
				Title:       f.Title,
				Slug:        f.Slug,
				Description: desc,
				ExpiresAt:   f.ExpiresAt.Time,
				CreatedAt:   f.CreatedAt,
			}

			fs = append(fs, d)
		}
	}
	td.Data["Feedbacks"] = fs
	template_utils.RenderTemplate(
		w,
		repo.Templates["feedback_pending.html"],
		td,
	)
}

// // Only authenticated user
// // board owner must be the same as authenticated user

// // i.e. /feedbacks/pending?b=1234&page=1 (b is board slug)
// func (repo *FeedbackHandlerRepo) PendingFeedbacks(w http.ResponseWriter, r *http.Request) {
// 	rndr := template_utils.NewTemplateRenderer(repo.Session)
// 	sessionService := sessions_utils.NewSessionService(r, repo.Session)
// 	boardDbService := board_models.NewBoardDbService(repo.DB)
// 	feedbackDbService := feedback_models.NewFeedbackDbService(repo.DB)
// 	emailBanDbService := email_ban_models.NewEmailBanDbService(repo.DB)

// 	td := r.Context().Value(constants.TemplateDataContextKey).(*template_utils.TemplateData)

// 	if r.Method != http.MethodGet {
// 		rndr.Abort(w, r, repo.Templates["error.html"], http.StatusMethodNotAllowed, constants.Error405Title, constants.Error405Message)
// 		return
// 	}

// 	td.Data["BoardSlug"] = ""
// 	td.Data["TotalPendingCount"] = 0
// 	td.Data["PageSequence"] = ""
// 	td.Data["CurrentPage"] = 0
// 	td.Data["Feedbacks"] = make([]dataaccess.PendingFeedbackStrippedDao, 0)

// 	user, err := sessionService.GetUserFromSession(repo.DB)
// 	if err != nil {
// 		if !errors.Is(err, sessions_utils.ErrNoUserInSession) && !errors.Is(err, sql.ErrNoRows) {
// 			repo.Logger.Errorln(err)
// 			rndr.Abort(w, r, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
// 			return
// 		}
// 		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
// 		return
// 	}
// 	ban, err := emailBanDbService.ReadByEmail(user.Email)
// 	if err != nil {
// 		if !errors.Is(err, sql.ErrNoRows) {
// 			repo.Logger.Errorln(err)
// 			rndr.Abort(w, r, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
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
// 	td.Data["BoardSlug"] = board.Slug

// 	// check if current user is board owner
// 	if user.ID != board.OwnerID {
// 		rndr.Abort(w, r, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
// 		return
// 	}

// 	totalPendingCount, err := feedbackDbService.ReadPendingTotalCountFeedbacksByBoardId(board.ID)
// 	if err != nil {
// 		repo.Logger.Errorf("%q", err)
// 		rndr.Abort(
// 			w, r,
// 			repo.Templates["error.html"],
// 			http.StatusInternalServerError,
// 			constants.Error500Title,
// 			constants.Error500Message,
// 		)
// 		return
// 	}
// 	td.Data["TotalPendingCount"] = totalPendingCount

// 	totalCountOfPages := int((totalPendingCount + constants.PerPageConstant - 1) / constants.PerPageConstant)

// 	pageSeq := make([]int, 0)
// 	if totalCountOfPages > 0 {
// 		pageSeq = number_utils.MakeSliceFromInt(totalCountOfPages)
// 	}
// 	td.Data["PageSequence"] = pageSeq

// 	page, err := strconv.Atoi(r.URL.Query().Get("page"))
// 	if err != nil {
// 		page = 1
// 	}

// 	redirect := false
// 	switch {
// 	case totalCountOfPages == 0:
// 		// doesn't matter the page number or whether there is page number
// 		// no feedbacks anyways
// 		// this should help with losing of flash message on redirect
// 		redirect = false
// 	case totalCountOfPages != 0 && page < 1:
// 		redirect = true
// 		page = 1
// 	case totalCountOfPages != 0 && page > totalCountOfPages:
// 		redirect = true
// 		page = totalCountOfPages
// 	case totalCountOfPages > 0 && page <= totalCountOfPages:
// 		redirect = false
// 	default:
// 		redirect = false
// 	}

// 	if redirect {
// 		url := fmt.Sprintf("/feedbacks/pending?b=%s&page=%d", board.Slug, page)
// 		http.Redirect(w, r, url, http.StatusSeeOther)
// 		return
// 	}
// 	td.Data["CurrentPage"] = page

// 	feedbacks, err := feedbackDbService.ReadPaginatedPendingFeedbacksByBoardId(board.ID, page, constants.PerPageConstant)
// 	if err != nil {
// 		repo.Logger.Errorf("%q", err, debug.Stack())
// 		rndr.Abort(
// 			w, r,
// 			repo.Templates["error.html"],
// 			http.StatusInternalServerError,
// 			constants.Error500Title,
// 			constants.Error500Message,
// 		)
// 		return
// 	}

// 	fs := make([]dataaccess.PendingFeedbackStrippedDao, 0)
// 	if len(feedbacks) > 0 {
// 		for _, f := range feedbacks {
// 			desc := ""
// 			if f.Description.Valid {
// 				desc = f.Description.String
// 			}
// 			d := dataaccess.PendingFeedbackStrippedDao{
// 				Title:       f.Title,
// 				Slug:        f.Slug,
// 				Description: desc,
// 				ExpiresAt:   f.ExpiresAt.Time,
// 				CreatedAt:   f.CreatedAt,
// 			}

// 			fs = append(fs, d)
// 		}
// 	}
// 	td.Data["Feedbacks"] = fs
// 	rndr.RenderTemplate(
// 		w, r,
// 		repo.Templates["feedback_pending.html"],
// 		td,
// 	)

// }
