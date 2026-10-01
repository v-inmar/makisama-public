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

func (repo *HandlerRepo) GetBoardReads(w http.ResponseWriter, r *http.Request) {
	td := r.Context().Value(constants.TemplateDataContextKey).(*template_utils.TemplateData)

	sessionService := sessions_utils.NewSessionService(r, repo.Session)
	currentUser, err := sessionService.GetUserFromSession(repo.DB)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) && !errors.Is(err, sessions_utils.ErrNoUserInSession) {
			repo.Logger.Errorln(err)
			template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
			return
		}

		template_utils.Abort(w, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
		return

	}

	emailBanDbService := email_ban_models.NewEmailBanDbService(repo.DB)
	ban, err := emailBanDbService.ReadByEmail(currentUser.Email)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			repo.Logger.Errorf("%v", err)
			template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
			return
		}
	}
	if ban != nil {
		sessionService.DestroySession()
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusForbidden, constants.Error403Title, constants.Error403Message)
		return
	}

	boardDbService := board_models.NewBoardDbService(repo.DB)

	totalBoardCount, err := boardDbService.ReadCountActiveBoardsByUserId(currentUser.ID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			repo.Logger.Errorf("%v", err)
			template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
			return
		}
		totalBoardCount = 0
	}
	td.Data["TotalBoardCount"] = totalBoardCount

	// this is just for getting the query param
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		page = 1 // value if query paran has invalid value
	}
	td.Data["QueryPage"] = page

	// calcuate total number of pages
	totalCountOfPages := int((totalBoardCount + constants.PerPageConstant - 1) / constants.PerPageConstant)

	// protect pagination
	redirect := false
	switch {
	case totalCountOfPages == 0 && page != 0:
		redirect = true
		page = 0
	case page > totalCountOfPages:
		redirect = true
		page = totalCountOfPages
	case totalCountOfPages > 0 && page < 1:
		redirect = true
		page = 1
	default:
		redirect = false
	}

	if redirect {
		url := fmt.Sprintf("/boards?page=%d", page)
		http.Redirect(w, r, url, http.StatusSeeOther)
		return
	}

	td.Data["CurrentPage"] = page

	// ----- to help producing the pagination navigator ---- //
	pageSeq := make([]int, 0)
	if totalCountOfPages > 0 {
		pageSeq = number_utils.MakeSliceFromInt(totalCountOfPages)
	}
	td.Data["PageSequence"] = pageSeq

	boards, err := boardDbService.ReadPaginatedActiveBoardsByUserId(currentUser.ID, page, constants.PerPageConstant)

	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			repo.Logger.Errorf("%v", err)
			template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
			return
		}

		boards = make([]board_models.Board, 0)
	}

	feedbackDbService := feedback_models.NewFeedbackDbService(repo.DB)

	boardDaos := make([]dataaccess.BoardListDataDao, 0)
	if len(boards) > 0 {
		for _, b := range boards {
			totalFeedbackCount, err := feedbackDbService.ReadTotalActiveFeedbackCountByBoardId(b.ID)
			if err != nil {
				repo.Logger.Errorf("%v", err)
				continue
			}

			desc := ""
			if b.Description.Valid {
				desc = b.Description.String
			}

			bd := dataaccess.BoardListDataDao{
				Title:         b.Title,
				Slug:          b.Slug,
				Description:   desc,
				CreatedAt:     b.CreatedAt,
				FeedbackCount: totalFeedbackCount,
			}
			boardDaos = append(boardDaos, bd)
		}
	}

	td.Data["Boards"] = boardDaos

	template_utils.RenderTemplate(
		w,
		repo.Templates["boards.html"],
		td,
	)
}

// func (repo *BoardHandlerRepo) Boards(w http.ResponseWriter, r *http.Request) {
// 	rndr := template_utils.NewTemplateRenderer(repo.Session)
// 	sessionService := sessions_utils.NewSessionService(r, repo.Session)
// 	boardDbService := board_models.NewBoardDbService(repo.DB)
// 	feedbackDbService := feedback_models.NewFeedbackDbService(repo.DB)
// 	emailBanDbService := email_ban_models.NewEmailBanDbService(repo.DB)

// 	if r.Method != http.MethodGet {
// 		rndr.Abort(w, r, repo.Templates["error.html"], http.StatusMethodNotAllowed, constants.Error405Title, constants.Error405Message)
// 		return
// 	}

// 	td := r.Context().Value(constants.TemplateDataContextKey).(*template_utils.TemplateData)

// 	user, err := sessionService.GetUserFromSession(repo.DB)
// 	if err != nil {
// 		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, sessions_utils.ErrNoUserInSession) {
// 			rndr.Abort(w, r, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
// 			return
// 		}

// 		repo.Logger.Errorf("%v", err)
// 		rndr.Abort(w, r, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
// 		return
// 	}
// 	ban, err := emailBanDbService.ReadByEmail(user.Email)
// 	if err != nil {
// 		if !errors.Is(err, sql.ErrNoRows) {
// 			repo.Logger.Errorf("%v", err)
// 			rndr.Abort(w, r, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
// 			return
// 		}
// 	}
// 	if ban != nil {
// 		sessionService.DestroySession()
// 		rndr.Abort(w, r, repo.Templates["error.html"], http.StatusForbidden, constants.Error403Title, constants.Error403Message)
// 		return
// 	}

// 	totalBoardCount, err := boardDbService.ReadCountActiveBoardsByUserId(user.ID)
// 	if err != nil {
// 		if !errors.Is(err, sql.ErrNoRows) {
// 			repo.Logger.Errorf("%v", err)
// 			rndr.Abort(w, r, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
// 			return
// 		}
// 		totalBoardCount = 0
// 	}
// 	td.Data["TotalBoardCount"] = totalBoardCount

// 	// this is just for getting the query param
// 	page, err := strconv.Atoi(r.URL.Query().Get("page"))
// 	if err != nil {
// 		page = 1 // value if query paran has invalid value
// 	}
// 	td.Data["QueryPage"] = page

// 	// calcuate total number of pages
// 	totalCountOfPages := int((totalBoardCount + constants.PerPageConstant - 1) / constants.PerPageConstant)

// 	// protect pagination
// 	redirect := false
// 	switch {
// 	case totalCountOfPages == 0 && page != 0:
// 		redirect = true
// 		page = 0
// 	case page > totalCountOfPages:
// 		redirect = true
// 		page = totalCountOfPages
// 	case totalCountOfPages > 0 && page < 1:
// 		redirect = true
// 		page = 1
// 	default:
// 		redirect = false
// 	}

// 	if redirect {
// 		url := fmt.Sprintf("/boards?page=%d", page)
// 		http.Redirect(w, r, url, http.StatusSeeOther)
// 		return
// 	}

// 	td.Data["CurrentPage"] = page

// 	// ----- to help producing the pagination navigator ---- //
// 	pageSeq := make([]int, 0)
// 	if totalCountOfPages > 0 {
// 		pageSeq = number_utils.MakeSliceFromInt(totalCountOfPages)
// 	}
// 	td.Data["PageSequence"] = pageSeq

// 	boards, err := boardDbService.ReadPaginatedActiveBoardsByUserId(user.ID, page, constants.PerPageConstant)

// 	if err != nil {
// 		if !errors.Is(err, sql.ErrNoRows) {
// 			repo.Logger.Errorf("%v", err)
// 			rndr.Abort(w, r, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
// 			return
// 		}

// 		boards = make([]board_models.Board, 0)
// 	}

// 	boardDaos := make([]dataaccess.BoardListDataDao, 0)

// 	if len(boards) > 0 {
// 		for _, b := range boards {
// 			totalFeedbackCount, err := feedbackDbService.ReadTotalActiveFeedbackCountByBoardId(b.ID)
// 			if err != nil {
// 				repo.Logger.Errorf("%v", err)
// 				continue
// 			}

// 			desc := ""
// 			if b.Description.Valid {
// 				desc = b.Description.String
// 			}

// 			bd := dataaccess.BoardListDataDao{
// 				Title:         b.Title,
// 				Slug:          b.Slug,
// 				Description:   desc,
// 				CreatedAt:     b.CreatedAt,
// 				FeedbackCount: totalFeedbackCount,
// 			}
// 			boardDaos = append(boardDaos, bd)
// 		}
// 	}

// 	td.Data["Boards"] = boardDaos

// 	rndr.RenderTemplate(
// 		w, r,
// 		repo.Templates["boards.html"],
// 		td,
// 	)
// }
