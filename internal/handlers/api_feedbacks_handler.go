package handlers

import (
	"app/internal/constants"
	dataaccess "app/internal/data_access"
	board_models "app/internal/models/board"
	email_ban_models "app/internal/models/email_ban"
	feedback_models "app/internal/models/feedback"
	user_models "app/internal/models/user"
	api_utils "app/internal/utilities/api"
	date_utils "app/internal/utilities/dates"
	sessions_utils "app/internal/utilities/sessions"
	strings_utils "app/internal/utilities/strings"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
)

func (repo *HandlerRepo) GetApiFeedback(w http.ResponseWriter, r *http.Request) {

	sessionService := sessions_utils.NewSessionService(r, repo.Session)
	apiService := api_utils.NewApiResponseService(w)
	currentUser, err := sessionService.GetUserFromSession(repo.DB)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) && !errors.Is(err, sessions_utils.ErrNoUserInSession) {
			repo.Logger.Errorln(err)
			apiService.Abort(http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
			return
		}
	}

	emailBanDbService := email_ban_models.NewEmailBanDbService(repo.DB)
	if currentUser != nil {
		// check current user ban
		ban, err := emailBanDbService.ReadByEmail(currentUser.Email)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				repo.Logger.Errorln(err)
				apiService.Abort(http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
				return
			}
		}

		if ban != nil {
			http.Redirect(w, r, r.URL.RequestURI(), http.StatusSeeOther)
			return
		}
	}

	feedbackDbService := feedback_models.NewFeedbackDbService(repo.DB)
	slug := r.PathValue("slug")
	feedback, err := feedbackDbService.ReadBySlug(slug)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			repo.Logger.Errorln(err)
			apiService.Abort(http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
			return
		}
		apiService.Abort(http.StatusNotFound, constants.Error404Title, constants.Error404Message)
		return
	}

	if feedback.DeletedAt.Valid {
		apiService.Abort(http.StatusNotFound, constants.Error404Title, constants.Error404Message)
		return
	}

	// get board
	boardDbService := board_models.NewBoardDbService(repo.DB)
	board, err := boardDbService.ReadById(feedback.BoardID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			repo.Logger.Errorln(err)
			apiService.Abort(http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
			return
		}
		apiService.Abort(http.StatusNotFound, constants.Error404Title, constants.Error404Message)
		return
	}

	// get board owner user
	userDbService := user_models.NewUserDbService(repo.DB)
	boardOwnerUser, err := userDbService.ReadById(board.OwnerID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			apiService.Abort(http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
			return
		}
		apiService.Abort(http.StatusNotFound, constants.Error404Title, constants.Error404Message)
		return
	}

	// check if board owner is banned
	ban, err := emailBanDbService.ReadByEmail(boardOwnerUser.Email)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			apiService.Abort(http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
			return
		}
	}
	if ban != nil {
		apiService.Abort(http.StatusNotFound, constants.Error404Title, constants.Error404Message)
		return
	}

	if boardOwnerUser.DeactivatedAt.Valid {
		apiService.Abort(http.StatusNotFound, constants.Error404Title, constants.Error404Message)
		return
	}

	description := ""
	if feedback.Description.Valid {
		description = feedback.Description.String
	}

	resolution := ""
	if feedback.Resolution.Valid {
		resolution = strings_utils.Capitalize(feedback.Resolution.String)
	}

	scope := ""
	if feedback.Scope.Valid {
		scope = strings_utils.Capitalize(feedback.Scope.String)
	}

	expiry := ""
	if feedback.ExpiresAt.Valid {
		expiry = date_utils.FormatTime(feedback.ExpiresAt.Time)
	}

	isBoardOwner := false
	if currentUser != nil {
		isBoardOwner = currentUser.ID == boardOwnerUser.ID
	}

	payload := dataaccess.ReadFeedbackResponseDao{
		Title:        feedback.Title,
		Description:  description,
		Slug:         feedback.Slug,
		Resolution:   resolution,
		Scope:        scope,
		ExpiresAt:    expiry,
		IsBoardOwner: isBoardOwner,
		CreatedAt:    feedback.CreatedAt,
		BoardSlug:    board.Slug,
	}

	statsuCode := http.StatusOK

	d := &dataaccess.ApiResponseDao{
		Code:    statsuCode,
		Status:  "OK",
		Payload: payload,
	}

	apiService.Response(
		statsuCode,
		d,
	)
}

func (repo *HandlerRepo) PostApiFeedback(w http.ResponseWriter, r *http.Request) {
	sessionService := sessions_utils.NewSessionService(r, repo.Session)
	apiService := api_utils.NewApiResponseService(w)
	// get user is logged in
	currentUser, err := sessionService.GetUserFromSession(repo.DB)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) && !errors.Is(err, sessions_utils.ErrNoUserInSession) {
			repo.Logger.Errorf("%v", err)
			apiService.Abort(
				http.StatusInternalServerError,
				constants.Error500Title,
				constants.Error500Message,
			)
			return

		}

		apiService.Abort(
			http.StatusNotFound,
			constants.Error404Title,
			constants.Error404Message,
		)
		return

		// Only authenticated user are allowed to use this endpoint

	}

	emailBanDbService := email_ban_models.NewEmailBanDbService(repo.DB)
	ban, err := emailBanDbService.ReadByEmail(currentUser.Email)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			repo.Logger.Errorf("%v", err)
			apiService.Abort(
				http.StatusInternalServerError,
				constants.Error500Title,
				constants.Error500Message,
			)
			return
		}
	}
	if ban != nil {
		apiService.Abort(http.StatusForbidden, constants.Error403Title, constants.Error403Message)
		return
	}

	var body dataaccess.ApiCreateFeedbackRequestDao
	errDecodeBody := json.NewDecoder(r.Body).Decode(&body)
	if errDecodeBody != nil {
		repo.Logger.Errorln(errDecodeBody)
		apiService.Abort(http.StatusBadRequest, constants.Error400Title, errDecodeBody.Error())
		return
	}

	// check board exist

	boardDbService := board_models.NewBoardDbService(repo.DB)
	board, err := boardDbService.ReadBySlug(body.BoardSlug)
	if err != nil {
		if err == sql.ErrNoRows {
			apiService.Abort(http.StatusNotFound, constants.Error404Title, "board does not exist")
			return
		}
		repo.Logger.Errorln(err)
		apiService.Abort(http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
		return

	}

	// check board owner (user) is not deactivated
	userDbService := user_models.NewUserDbService(repo.DB)
	boardOwnerUser, err := userDbService.ReadById(board.OwnerID)
	if err != nil {
		if err == sql.ErrNoRows {
			apiService.Abort(http.StatusNotFound, constants.Error404Title, "board owner does not exist")
			return
		}
		repo.Logger.Errorln(err)
		apiService.Abort(http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
		return

	}

	// check board owner ban
	// ban, err = emailBanDbService.ReadByEmail(boardOwnerUser.Email)
	// if err != nil {
	// 	if !errors.Is(err, sql.ErrNoRows) {
	// 		repo.Logger.Errorln(err)
	// 		apiService.Abort(http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
	// 		return
	// 	}
	// }
	// if ban != nil {
	// 	apiService.Abort(http.StatusNotFound, constants.Error404Title, constants.Error404Message)
	// 	return
	// }

	if boardOwnerUser.DeactivatedAt.Valid {
		apiService.Abort(http.StatusNotFound, constants.Error404Title, "board owner does not exist")
		return
	}

	feedbackDbService := feedback_models.NewFeedbackDbService(repo.DB)
	_, errCreate := feedbackDbService.Create(body.Title, body.Description, "", board.ID, currentUser.ID)
	if errCreate != nil {
		repo.Logger.Errorf("%v", errCreate)
		apiService.Abort(http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
		return
	}

	// send email to board owner
	// NOTE: NOT NOW, CAN'T SEND TOO MANY EMAILS, MIGHT RAN OUT OF EMAIL
	// SINCE WE ARE USING 3RD PARTY

	statusCode := http.StatusCreated

	msg := "Feedback submitted and waiting for board owner's review"

	d := &dataaccess.ApiResponseDao{
		Code:   statusCode,
		Status: "Created",
		Payload: dataaccess.CreateFeedbackResponseDao{
			Message: msg,
		},
	}

	apiService.Response(
		statusCode,
		d,
	)
}

// func (repo *ApiFeedbackHandlerRepo) ApiPostCreate(w http.ResponseWriter, r *http.Request) {
// 	boardDbService := board_models.NewBoardDbService(repo.DB)
// 	feedbackDbService := feedback_models.NewFeedbackDbService(repo.DB)
// 	sessionService := sessions_utils.NewSessionService(r, repo.Session)
// 	userDbService := user_models.NewUserDbService(repo.DB)
// 	apiService := api_utils.NewApiResponseService(w)
// 	emailBanDbService := email_ban_models.NewEmailBanDbService(repo.DB)

// 	if r.Method != http.MethodPost {
// 		repo.Logger.Errorf("method used is not allowed: %s", r.Method)
// 		apiService.Abort(http.StatusMethodNotAllowed, constants.Error405Title, constants.Error405Message)
// 		return
// 	}

// 	// get user is logged in
// 	currentUser, err := sessionService.GetUserFromSession(repo.DB)
// 	if err != nil {
// 		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, sessions_utils.ErrNoUserInSession) {
// 			// just respond with 404
// 			apiService.Abort(
// 				http.StatusNotFound,
// 				constants.Error404Title,
// 				constants.Error404Message,
// 			)
// 			return
// 		}
// 		repo.Logger.Errorf("%v", err)
// 		apiService.Abort(
// 			http.StatusInternalServerError,
// 			constants.Error500Title,
// 			constants.Error500Message,
// 		)
// 		return

// 		// Only authenticated user are allowed to use this endpoint

// 	}

// 	ban, err := emailBanDbService.ReadByEmail(currentUser.Email)
// 	if err != nil {
// 		if !errors.Is(err, sql.ErrNoRows) {
// 			repo.Logger.Errorf("%v", err)
// 			apiService.Abort(
// 				http.StatusInternalServerError,
// 				constants.Error500Title,
// 				constants.Error500Message,
// 			)
// 			return
// 		}
// 	}
// 	if ban != nil {
// 		apiService.Abort(http.StatusForbidden, constants.Error403Title, constants.Error403Message)
// 		return
// 	}

// 	var body dataaccess.ApiCreateFeedbackRequestDao
// 	errDecodeBody := json.NewDecoder(r.Body).Decode(&body)
// 	if errDecodeBody != nil {
// 		repo.Logger.Errorln(errDecodeBody)
// 		apiService.Abort(http.StatusBadRequest, constants.Error400Title, errDecodeBody.Error())
// 		return
// 	}

// 	// check board exist
// 	board, err := boardDbService.ReadBySlug(body.BoardSlug)
// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			apiService.Abort(http.StatusNotFound, constants.Error404Title, "board does not exist")
// 			return
// 		}
// 		repo.Logger.Errorln(err)
// 		apiService.Abort(http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
// 		return

// 	}

// 	// check board owner (user) is not deactivated
// 	boardOwnerUser, err := userDbService.ReadById(board.OwnerID)
// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			apiService.Abort(http.StatusNotFound, constants.Error404Title, "board owner does not exist")
// 			return
// 		}
// 		repo.Logger.Errorln(err)
// 		apiService.Abort(http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
// 		return

// 	}

// 	// check board owner ban
// 	// ban, err = emailBanDbService.ReadByEmail(boardOwnerUser.Email)
// 	// if err != nil {
// 	// 	if !errors.Is(err, sql.ErrNoRows) {
// 	// 		repo.Logger.Errorln(err)
// 	// 		apiService.Abort(http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
// 	// 		return
// 	// 	}
// 	// }
// 	// if ban != nil {
// 	// 	apiService.Abort(http.StatusNotFound, constants.Error404Title, constants.Error404Message)
// 	// 	return
// 	// }

// 	if boardOwnerUser.DeactivatedAt.Valid {
// 		apiService.Abort(http.StatusNotFound, constants.Error404Title, "board owner does not exist")
// 		return
// 	}

// 	_, errCreate := feedbackDbService.Create(body.Title, body.Description, "", board.ID, currentUser.ID)
// 	if errCreate != nil {
// 		repo.Logger.Errorf("%v", errCreate)
// 		apiService.Abort(http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
// 		return
// 	}

// 	// send email to board owner
// 	// NOTE: NOT NOW, CAN'T SEND TOO MANY EMAILS, MIGHT RAN OUT OF EMAIL
// 	// SINCE WE ARE USING 3RD PARTY

// 	statusCode := http.StatusCreated

// 	msg := "Feedback submitted and waiting for board owner's review"

// 	d := &dataaccess.ApiResponseDao{
// 		Code:   statusCode,
// 		Status: "Created",
// 		Payload: dataaccess.CreateFeedbackResponseDao{
// 			Message: msg,
// 		},
// 	}

// 	apiService.Response(
// 		statusCode,
// 		d,
// 	)

// }
