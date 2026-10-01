package handlers

import (
	"app/internal/constants"
	email_ban_models "app/internal/models/email_ban"
	image_models "app/internal/models/image"
	sessions_utils "app/internal/utilities/sessions"
	template_utils "app/internal/utilities/templates"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
)

func (repo *HandlerRepo) GetFilesRead(w http.ResponseWriter, r *http.Request) {
	// td := r.Context().Value(constants.TemplateDataContextKey).(*template_utils.TemplateData)

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
	}

	filename := r.PathValue("filename")
	secFilename := filepath.Base(filename)

	imgDbService := image_models.NewImageDbService(repo.DB)
	img, err := imgDbService.ReadByFilename(secFilename)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			template_utils.Abort(w, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
			return
		}
		repo.Logger.Errorln(err)
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
		return
	}

	if img.TTL.Valid {
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
		return
	}

	finalDiskPath := filepath.Join(constants.BoardUploadImageDirectory, fmt.Sprintf("%v.png", secFilename))
	http.ServeFile(w, r, finalDiskPath)
}

// func (repo *FilesHandlerRepo) Files(w http.ResponseWriter, r *http.Request) {
// 	imgDbService := image_models.NewImageDbService(repo.DB)
// 	rndr := template_utils.NewTemplateRenderer(repo.Session)
// 	sessionService := sessions_utils.NewSessionService(r, repo.Session)
// 	emailBanDbService := email_ban_models.NewEmailBanDbService(repo.DB)

// 	currentUser, err := sessionService.GetUserFromSession(repo.DB)
// 	if err != nil {
// 		if !errors.Is(err, sql.ErrNoRows) && !errors.Is(err, sessions_utils.ErrNoUserInSession) {
// 			repo.Logger.Errorln(err)
// 			rndr.Abort(w, r, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
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
// 	}

// 	if r.Method != http.MethodGet {
// 		rndr.Abort(w, r, repo.Templates["error.html"], http.StatusMethodNotAllowed, constants.Error405Title, constants.Error405Message)
// 		return
// 	}

// 	filename := r.PathValue("filename")
// 	secFilename := filepath.Base(filename)

// 	img, err := imgDbService.ReadByFilename(secFilename)
// 	if err != nil {
// 		if errors.Is(err, sql.ErrNoRows) {
// 			rndr.Abort(w, r, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
// 			return
// 		}
// 		repo.Logger.Errorln(err)
// 		rndr.Abort(w, r, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
// 		return
// 	}

// 	if img.TTL.Valid {
// 		rndr.Abort(w, r, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
// 		return
// 	}

// 	finalDiskPath := filepath.Join(constants.BoardUploadImageDirectory, fmt.Sprintf("%v.png", secFilename))
// 	http.ServeFile(w, r, finalDiskPath)

// }
