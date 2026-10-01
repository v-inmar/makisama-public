package handlers

import (
	"app/internal/constants"
	dataaccess "app/internal/data_access"
	board_models "app/internal/models/board"
	email_ban_models "app/internal/models/email_ban"
	feedback_models "app/internal/models/feedback"
	image_models "app/internal/models/image"
	user_models "app/internal/models/user"
	number_utils "app/internal/utilities/number"
	sessions_utils "app/internal/utilities/sessions"
	template_utils "app/internal/utilities/templates"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

func (repo *HandlerRepo) GetBoardRead(w http.ResponseWriter, r *http.Request) {
	td := r.Context().Value(constants.TemplateDataContextKey).(*template_utils.TemplateData)
	td.Data["ScopeList"] = constants.FeedbackScopes
	td.Data["ResolutionList"] = constants.FeedbackResolutions

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

	boardDbService := board_models.NewBoardDbService(repo.DB)

	slugParam := r.PathValue("slug")
	board, err := boardDbService.ReadBySlug(slugParam)
	if err != nil {
		if err == sql.ErrNoRows {
			template_utils.Abort(w, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
			return
		}
		repo.Logger.Errorln(err)
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
		return
	}

	td.Data["BoardTitle"] = board.Title
	td.Data["BoardWebsite"] = board.Website.String
	td.Data["BoardSlug"] = board.Slug
	td.Data["BoardDescription"] = board.Description.String

	boardImageDbService := image_models.NewBoardImageDbService(repo.DB)
	td.Data["BoardImage"] = ""
	boardImageModel, err := boardImageDbService.ReadByBoardId(board.ID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			repo.Logger.Errorln(err)
			template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
			return
		}
	}

	imgDbService := image_models.NewImageDbService(repo.DB)
	if boardImageModel != nil {
		imgModel, err := imgDbService.ReadById(boardImageModel.ImageId)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				repo.Logger.Errorln(err)
				template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
				return
			}
		}
		if imgModel != nil {
			td.Data["BoardImage"] = imgModel.Filename
		}
	}

	userDbService := user_models.NewUserDbService(repo.DB)
	boardOwnerUser, errUser := userDbService.ReadById(board.OwnerID)
	if errUser != nil {
		if errUser == sql.ErrNoRows { // user has been deleted
			if errUser == sql.ErrNoRows {
				template_utils.Abort(w, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
				return
			}
			repo.Logger.Errorln(errUser)
			template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
			return
		}
	}

	ban, err := emailBanDbService.ReadByEmail(boardOwnerUser.Email)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			repo.Logger.Errorln(errUser)
			template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
			return
		}
	}
	if ban != nil {
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
		return
	}

	// board owner is not active - board not accessible
	if boardOwnerUser.DeactivatedAt.Valid {
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
		return
	}

	td.Data["IsOwner"] = false
	if currentUser != nil {
		if currentUser.ID == boardOwnerUser.ID {
			td.Data["IsOwner"] = true
		}
	}

	// ---- for the stats ----- //
	feedbackDbService := feedback_models.NewFeedbackDbService(repo.DB)

	totalActiveFeedbackCount, err := feedbackDbService.ReadTotalActiveFeedbackCountByBoardId(board.ID)
	if err != nil {
		repo.Logger.Errorf("%v", err)
		totalActiveFeedbackCount = 0

	}

	td.Data["TotalNumberOfFeedback"] = totalActiveFeedbackCount

	type ScopeStat struct {
		Scope string
		Count int
	}

	scopeStats := make([]ScopeStat, 0)

	for k := range constants.FeedbackScopes {
		c, err := feedbackDbService.ReadTotalScopeCountByBoardIdAndScopeValue(board.ID, k)
		if err != nil {
			continue
		}

		scopeStats = append(scopeStats, ScopeStat{
			Scope: k,
			Count: c,
		})

	}
	td.Data["ScopeStats"] = scopeStats

	type ResolutionStat struct {
		Resolution string
		Count      int
	}

	resStats := make([]ResolutionStat, 0)

	for k := range constants.FeedbackResolutions {
		c, err := feedbackDbService.ReadTotalResolutionCountByBoardIdAndResValue(board.ID, k)
		if err != nil {
			continue
		}

		resStats = append(resStats, ResolutionStat{
			Resolution: k,
			Count:      c,
		})

	}

	td.Data["ResolutionStats"] = resStats

	// ------- check if there are pending feedbacks ------- //
	isTherePending, err := feedbackDbService.IsTherePending(board.ID)
	if err != nil {
		repo.Logger.Errorf("%v", err)
		isTherePending = false
	}
	td.Data["IsTherePending"] = isTherePending

	// ------------ get all query params ----------- //
	scopeParam := r.URL.Query().Get("scope")
	if _, exist := constants.FeedbackScopes[scopeParam]; !exist {
		scopeParam = "all"
	}
	// if !slices.Contains(constants.FeedbackScopes, scopeParam) {
	// 	scopeParam = "all"
	// }
	td.Data["QueryScope"] = scopeParam

	resolutionParam := r.URL.Query().Get("resolution")
	if _, exist := constants.FeedbackResolutions[resolutionParam]; !exist {
		resolutionParam = "all"
	}

	td.Data["QueryResolution"] = resolutionParam

	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		page = 0
	}
	td.Data["QueryPage"] = page

	// ------- safe guard floor and ceiling of pagination --------- //
	totalFeedbackCountWithFilters, err := feedbackDbService.ReadActiveTotalCountFilteredFeedbacksByBoardId(
		board.ID,
		scopeParam,
		resolutionParam,
	)
	if err != nil {
		repo.Logger.Errorf("%v", err)
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
		return
	}
	td.Data["TotalFeedbackCount"] = totalFeedbackCountWithFilters

	// calcuate total number of pages
	totalCountOfPages := int((totalFeedbackCountWithFilters + constants.PerPageConstant - 1) / constants.PerPageConstant)
	// protect pagination
	redirect := false
	switch {
	case totalCountOfPages == 0:
		// doesn't matter the page number or whether there is page number
		// no feedbacks anyways
		// this should help with losing of flash message on redirect
		redirect = false
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
		// any flash messages has been lost because of this redirect
		// such as after creating new feedback, it redirects back to board
		// and if it hits, this, the flash wont be triggered because session already changed
		url := fmt.Sprintf("/%s?scope=%s&resolution=%s&page=%d", board.Slug, scopeParam, resolutionParam, page)
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

	// --------- db read with filters ----------------- //
	feedbacks, err := feedbackDbService.ReadActivePaginatedFilteredFeedbacksByBoardId(
		board.ID,
		page,
		constants.PerPageConstant,
		scopeParam,
		resolutionParam,
	)
	// check err
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			// must abort for any type of db error that is not sql.ErrNoRows
			repo.Logger.Errorf("%v", err)
			template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
			return
		}
		// any error returns nil, so we make it have 0 length
		// avoids unnecessary behaviour
		feedbacks = make([]feedback_models.Feedback, 0)
	}

	// strip to stop unnecessary exposure of data
	fs := make([]dataaccess.BoardFeedbackStrippedDao, 0)
	if len(feedbacks) > 0 {
		for _, f := range feedbacks {
			desc := ""
			if f.Description.Valid {
				desc = f.Description.String
			}
			scope := ""
			if f.Scope.Valid {
				scope = f.Scope.String
			}
			reso := ""
			if f.Resolution.Valid {
				reso = f.Resolution.String
			}

			d := dataaccess.BoardFeedbackStrippedDao{
				Title:       f.Title,
				Description: desc,
				Slug:        f.Slug,
				Scope:       scope,
				Resolution:  reso,
				CreatedAt:   f.CreatedAt,
			}
			fs = append(fs, d)
		}
	}

	td.Data["Feedbacks"] = fs

	template_utils.RenderTemplate(w, repo.Templates["board.html"], td)
}
