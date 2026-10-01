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

	_ "golang.org/x/image/webp"
)

func (repo *HandlerRepo) GetBoardCreateBoard(w http.ResponseWriter, r *http.Request) {
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

	template_utils.RenderTemplate(w, repo.Templates["board_new.html"], td)
}

func (repo *HandlerRepo) PostBoardCreateBoard(w http.ResponseWriter, r *http.Request) {
	// Hard-limit the incoming request body size at the network reader level.
	r.Body = http.MaxBytesReader(w, r.Body, int64(constants.MaxImageUploadSize)+1024)

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

	if err := r.ParseMultipartForm(int64(constants.MaxImageUploadSize) + 1024); err != nil {
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusBadRequest, constants.Error400Title, "Exceeded maximum limit of body size")
		return
	}

	title := r.Form.Get("title")
	description := r.Form.Get("description")
	website := r.Form.Get("website")
	acknowledge := r.Form.Get("acknowledge") != ""
	td.Data["Title"] = title
	td.Data["Description"] = description
	td.Data["Website"] = website

	if !strings_utils.IsValidLength(title, constants.BoardTitleMinLength, constants.BoardTitleMaxLength) {
		td.Data["ErrorTitle"] = fmt.Sprintf("Must be between %v and %v characters long", constants.BoardTitleMinLength, constants.BoardTitleMaxLength)
		template_utils.RenderTemplate(w, repo.Templates["board_new.html"], td)
		return
	}

	if len(website) > constants.BoardWebsiteMaxLength {
		td.Data["ErrorWebsite"] = fmt.Sprintf("Exceeded maximum length of %v characters", constants.BoardWebsiteMaxLength)
		template_utils.RenderTemplate(w, repo.Templates["board_new.html"], td)
		return
	}

	if len(website) > 0 && len(website) <= constants.BoardWebsiteMaxLength {
		websiteNormalized, err := strings_utils.NormalizeWebsiteURL(website)
		if err != nil {
			repo.Logger.Errorln(err)
			td.Data["ErrorWebsite"] = "Format invalid"
			template_utils.RenderTemplate(w, repo.Templates["board_new.html"], td)
			return
		}
		website = websiteNormalized
	}

	if len(description) > 0 {
		description = strings_utils.Substring(description, constants.BoardDescriptionMaxLength)
	}

	if !acknowledge {
		td.Data["ErrorAcknowledge"] = "You must acknowledge to create a board"
		template_utils.RenderTemplate(w, repo.Templates["board_new.html"], td)
		return
	}

	var destinationPath string
	isFileUpload := true

	file, header, err := r.FormFile("image")
	if err != nil {
		if !errors.Is(err, http.ErrMissingFile) {
			td.Data["ErrorImage"] = "Could not process image upload"
			template_utils.RenderTemplate(w, repo.Templates["board_new.html"], td)
			return
		}
		isFileUpload = false
	}

	boardDbService := board_models.NewBoardDbService(repo.DB)
	var boardNew *board_models.Board
	errCreate := func() error {
		tx, err := repo.DB.Beginx()
		if err != nil {
			return fmt.Errorf("500:%v", err)
		}
		defer tx.Rollback()

		board, err := boardDbService.CreateWithTx(tx, title, website, description, currentUser.ID)
		if err != nil {
			return fmt.Errorf("500:%v", err)
		}
		if board != nil {
			boardNew = board
		}

		var imageModel *image_models.Image
		if isFileUpload {
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

			if imageModel == nil {
				return errors.New("500:unable to generate filename value")
			}

			defer file.Close()

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

			// Eliminate Path Traversal Attacks
			// Replace user-provided file names entirely with a 32-character secure hex token.
			// secFilename, err := randoms_utils.GenerateImageFilename()
			// if err != nil {
			// 	return errors.New(fmt.Sprintf("500 file error: %v", err))
			// }
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

			// if _, err := io.Copy(dst, file); err != nil {
			// 	return fmt.Errorf("500:%v", err)
			// }

			// Re-compress and save the image to disk specifically as a true PNG
			if err := png.Encode(dst, img); err != nil {
				return fmt.Errorf("500:%v", err)
			}
		}

		if imageModel != nil {
			boardImageDbService := image_models.NewBoardImageDbService(repo.DB)
			_, err := boardImageDbService.CreateWithTx(tx, board.ID, imageModel.ID)
			if err != nil {
				return fmt.Errorf("500:%v", err)
			}
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("500:%v", err)
		}

		return nil

	}()

	if errCreate != nil {
		splitErrorString := strings.Split(errCreate.Error(), ":")
		if splitErrorString[0] == "400" {
			td.Data["ErrorImage"] = splitErrorString[1]
			template_utils.RenderTemplate(w, repo.Templates["board_new.html"], td)
			return
		}
		repo.Logger.Errorln(errCreate)
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
		return
	}

	sessionService.ToFlash("Board created!")
	http.Redirect(w, r, fmt.Sprintf("/%v", boardNew.Slug), http.StatusSeeOther)

}
