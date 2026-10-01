package handlers

import (
	"app/internal/constants"
	board_models "app/internal/models/board"
	email_ban_models "app/internal/models/email_ban"
	image_models "app/internal/models/image"
	image_utils "app/internal/utilities/image"
	randoms_utils "app/internal/utilities/randoms"
	sessions_utils "app/internal/utilities/sessions"
	strings_utils "app/internal/utilities/strings"
	template_utils "app/internal/utilities/templates"
	"database/sql"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "golang.org/x/image/webp"
)

func (repo *HandlerRepo) GetBoardEditBoard(w http.ResponseWriter, r *http.Request) {
	td := r.Context().Value(constants.TemplateDataContextKey).(*template_utils.TemplateData)

	sessionService := sessions_utils.NewSessionService(r, repo.Session)
	currentUser, err := sessionService.GetUserFromSession(repo.DB)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, sessions_utils.ErrNoUserInSession) {
			http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
			return
		}
		repo.Logger.Errorln(err)
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
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

	boardDbService := board_models.NewBoardDbService(repo.DB)
	slug := r.PathValue("slug")
	board, err := boardDbService.ReadBySlug(slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			template_utils.Abort(w, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
			return
		}
		repo.Logger.Errorln(err)
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
		return
	}

	if board.OwnerID != currentUser.ID {
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
		return
	}

	td.Data["BoardSlug"] = board.Slug

	website := ""
	if board.Website.Valid {
		website = board.Website.String
	}

	description := ""
	if board.Description.Valid {
		description = board.Description.String
	}

	td.Data["Title"] = board.Title
	td.Data["Website"] = website
	td.Data["Description"] = description
	template_utils.RenderTemplate(w, repo.Templates["board_edit.html"], td)
}

func (repo *HandlerRepo) PostBoardEditBoard(w http.ResponseWriter, r *http.Request) {

	td := r.Context().Value(constants.TemplateDataContextKey).(*template_utils.TemplateData)
	sessionService := sessions_utils.NewSessionService(r, repo.Session)
	currentUser, err := sessionService.GetUserFromSession(repo.DB)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, sessions_utils.ErrNoUserInSession) {
			http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
			return
		}
		repo.Logger.Errorln(err)
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
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

	boardDbService := board_models.NewBoardDbService(repo.DB)
	slug := r.PathValue("slug")
	board, err := boardDbService.ReadBySlug(slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			template_utils.Abort(w, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
			return
		}
		repo.Logger.Errorln(err)
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
		return
	}

	if board.OwnerID != currentUser.ID {
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusNotFound, constants.Error404Title, constants.Error404Message)
		return
	}

	if err := r.ParseMultipartForm(int64(constants.MaxImageUploadSize) + 1024); err != nil {
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusBadRequest, constants.Error400Title, "Exceeded maximum limit of body size")
		return
	}

	newTitle := strings.TrimSpace(r.Form.Get("title"))
	newWebsite := strings.TrimSpace(r.Form.Get("website"))
	newDescription := strings.TrimSpace(r.Form.Get("description"))
	td.Data["Title"] = newTitle
	td.Data["Website"] = newWebsite
	td.Data["Description"] = newDescription

	if !strings_utils.IsValidLength(newTitle, constants.BoardTitleMinLength, constants.BoardTitleMaxLength) {
		td.Data["ErrorTitle"] = fmt.Sprintf("Must be between %v and %v characters long", constants.BoardTitleMinLength, constants.BoardTitleMaxLength)
		template_utils.RenderTemplate(w, repo.Templates["board_edit.html"], td)
		return
	}

	if len(newWebsite) > constants.BoardWebsiteMaxLength {
		td.Data["ErrorWebsite"] = fmt.Sprintf("Exceeded maximum length of %v characters", constants.BoardWebsiteMaxLength)
		template_utils.RenderTemplate(w, repo.Templates["board_edit.html"], td)
		return
	}

	if len(newWebsite) > 0 && len(newWebsite) <= constants.BoardWebsiteMaxLength {
		websiteNormalized, err := strings_utils.NormalizeWebsiteURL(newWebsite)
		if err != nil {
			repo.Logger.Errorln(err)
			td.Data["ErrorWebsite"] = "Format invalid"
			template_utils.RenderTemplate(w, repo.Templates["board_edit.html"], td)
			return
		}
		newWebsite = websiteNormalized
	}

	if len(newDescription) > 0 {
		newDescription = strings_utils.Substring(newDescription, constants.BoardDescriptionMaxLength)
	}

	var destinationPath string
	isFileUpload := true

	file, header, err := r.FormFile("image")
	if err != nil {
		if !errors.Is(err, http.ErrMissingFile) {
			td.Data["ErrorImage"] = "Could not process image upload"
			template_utils.RenderTemplate(w, repo.Templates["board_edit.html"], td)
			return
		}
		isFileUpload = false
	}
	var imgToDelete string
	errEdit := func() error {
		tx, err := repo.DB.Beginx()
		if err != nil {
			return fmt.Errorf("500:%v", err)
		}
		defer tx.Rollback()

		if err := boardDbService.UpdateWithTx(tx, board, newTitle, newWebsite, newDescription); err != nil {
			return fmt.Errorf("500:%v", err)
		}

		var imageModel *image_models.Image
		if isFileUpload {
			defer file.Close()

			imgDbService := image_models.NewImageDbService(repo.DB)
			for range 5 {
				val, err := randoms_utils.GenerateImageFilename()
				if err != nil {
					continue
				}

				i, err := imgDbService.CreateWithTx(tx, val, "png")
				if err != nil {
					continue
				}

				imageModel = i
				break
			}

			if header.Size > int64(constants.MaxImageUploadSize) {
				return fmt.Errorf("400:The uploaded file is too large. Max allowed size is 2MiB")
			}

			// Binary Content Sniffing (MIME Type Verification)
			// Read the first 512 bytes to discover the actual underlying file payload type.
			buffer := make([]byte, 512)
			n, err := file.Read(buffer)
			if err != nil && err != io.EOF {
				return fmt.Errorf("500:%v", err)
			}

			// Rewind the file stream read pointer back to index 0
			// so io.Copy can read the complete file out from the start later.
			if _, err := file.Seek(0, io.SeekStart); err != nil {
				return fmt.Errorf("500:%v", err)
			}

			contentType := http.DetectContentType(buffer[:n])
			if contentType == "application/octet-stream" {
				return fmt.Errorf("400:Unable to determine file type. Edit the image and save it as .png and try again.")
			}
			targetExtension := image_utils.GetExtFromContentType(contentType)
			if targetExtension == "" {
				return fmt.Errorf("400:Only valid JPEG, PNG, and WEBP image structures are allowed")
			}

			img, _, err := image.Decode(file)
			if err != nil {
				return fmt.Errorf("500:%v", err)
			}

			safeFilename := imageModel.Filename + ".png"

			// create directory if it doesnt exist
			if err := os.MkdirAll(constants.BoardUploadImageDirectory, os.ModePerm); err != nil {
				return fmt.Errorf("500:%v", err)
			}

			// Bind directory and clean name together safely
			destinationPath = filepath.Join(constants.BoardUploadImageDirectory, safeFilename)

			// create the file using the relative path with the filename
			dst, err := os.Create(destinationPath)
			if err != nil {
				return fmt.Errorf("500:%v", err)
			}
			defer dst.Close()

			if _, err := io.Copy(dst, file); err != nil {
				return fmt.Errorf("500:%v", err)
			}

			// Re-compress and save the image to disk specifically as a true PNG
			if err := png.Encode(dst, img); err != nil {
				return fmt.Errorf("500:%v", err)
			}

			// get bord image model, if it exist
			boardImageDbService := image_models.NewBoardImageDbService(repo.DB)
			bim, err := boardImageDbService.ReadByBoardIdWithTx(tx, board.ID)
			if err != nil {
				if !errors.Is(err, sql.ErrNoRows) {
					return fmt.Errorf("500:%v", err)
				}
			}
			if bim == nil {
				// create new board image
				_, err := boardImageDbService.CreateWithTx(tx, board.ID, imageModel.ID)
				if err != nil {
					return fmt.Errorf("500:%v", err)
				}
			} else {
				// edit the current board image

				// get the current image and set ttl
				currImg, err := imgDbService.ReadByIdWithTx(tx, bim.ImageId)
				if err != nil {
					return fmt.Errorf("500:%v", err)
				}

				now := time.Now().UTC()
				ttl := now.AddDate(0, 0, 365)
				if err := imgDbService.UpdateTTLWithTx(tx, currImg, &ttl); err != nil {
					return fmt.Errorf("500:%v", err)
				}

				// get the file to delete
				imgToDelete = fmt.Sprintf("%s.%s", currImg.Filename, currImg.Extension)

				// replace image id from board image with new image
				if err := boardImageDbService.UpdateImageWithTx(tx, bim, imageModel.ID); err != nil {
					return fmt.Errorf("500:%v", err)
				}
			}

		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("500:%v", err)
		}

		return nil

	}()

	if errEdit != nil {
		splitErrorString := strings.Split(errEdit.Error(), ":")
		if splitErrorString[0] == "400" {
			td.Data["ErrorImage"] = splitErrorString[1]
			template_utils.RenderTemplate(w, repo.Templates["board_edit.html"], td)
			return
		}
		repo.Logger.Errorln(errEdit)
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
		return
	}

	if imgToDelete != "" {
		go func() {
			image_utils.DeleteImageFile(filepath.Join(constants.BoardUploadImageDirectory, imgToDelete))
		}()
	}

	sessionService.ToFlash("Board updated")
	http.Redirect(w, r, fmt.Sprintf("/%v?page=1", board.Slug), http.StatusSeeOther)

}
