package handlers

import (
	"app/internal/constants"
	email_ban_models "app/internal/models/email_ban"
	search_models "app/internal/models/search"
	number_utils "app/internal/utilities/number"
	sessions_utils "app/internal/utilities/sessions"
	template_utils "app/internal/utilities/templates"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func (repo *HandlerRepo) GetSearch(w http.ResponseWriter, r *http.Request) {
	td := r.Context().Value(constants.TemplateDataContextKey).(*template_utils.TemplateData)
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

	searchValue := r.URL.Query().Get("q") // auto escapes i.e. one+two+three  ->  one two three
	td.Data["SearchInput"] = searchValue
	searchValue = strings.TrimSpace(searchValue)
	if len(searchValue) < 1 {
		td.Data["Results"] = make([]string, 0)
		template_utils.RenderTemplate(w, repo.Templates["search.html"], td)
		return
	}

	searchDbService := search_models.NewSearchDbService(repo.DB)
	totalCount, err := searchDbService.ReadCount(searchValue)
	if err != nil {
		repo.Logger.Errorln(err)
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
		return
	}

	if totalCount < 1 {
		td.Data["Results"] = make([]string, 0)
		template_utils.RenderTemplate(w, repo.Templates["search.html"], td)
		return
	}

	td.Data["TotalCount"] = totalCount

	//---- beyond this, there is data ---- //

	// calcuate total number of pages
	totalCountOfPages := int((totalCount + constants.PerPageConstant - 1) / constants.PerPageConstant)
	td.Data["Pages"] = number_utils.MakeSliceFromInt(totalCountOfPages)

	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		page = 1
	}

	// protect the pagination
	if totalCountOfPages > 0 && page < 1 {
		page = 1
	} else if totalCountOfPages > 0 && page > totalCountOfPages {
		page = totalCountOfPages
	}

	td.Data["CurrentPage"] = page

	results, err := searchDbService.ReadPaginated(searchValue, page, constants.PerPageConstant)
	if err != nil {
		repo.Logger.Errorln(err)
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
		return
	}

	td.Data["Results"] = results

	template_utils.RenderTemplate(w, repo.Templates["search.html"], td)
}

func (repo *HandlerRepo) PostSearch(w http.ResponseWriter, r *http.Request) {
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

	if err := r.ParseForm(); err != nil {
		template_utils.Abort(w, repo.Templates["error.html"], http.StatusBadRequest, constants.Error400Title, constants.Error400Message)
		return
	}
	search := r.FormValue("search")
	search = strings.TrimSpace(search)
	if len(search) < 1 {
		http.Redirect(w, r, "/search", http.StatusSeeOther)
		return
	}
	search = url.QueryEscape(search)
	http.Redirect(w, r, fmt.Sprintf("/search?q=%s", search), http.StatusSeeOther)

}

// func (repo *SearchHandlerRepo) Search(w http.ResponseWriter, r *http.Request) {
// 	rndr := template_utils.NewTemplateRenderer(repo.Session)
// 	searchDbService := search_models.NewSearchDbService(repo.DB)
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

// 	td := r.Context().Value(constants.TemplateDataContextKey).(*template_utils.TemplateData)
// 	// ========== GET ========== //
// 	if r.Method == http.MethodGet {
// 		searchValue := r.URL.Query().Get("q") // auto escapes i.e. one+two+three  ->  one two three
// 		td.Data["SearchInput"] = searchValue
// 		searchValue = strings.TrimSpace(searchValue)
// 		if len(searchValue) < 1 {
// 			td.Data["Results"] = make([]string, 0)
// 			rndr.RenderTemplate(w, r, repo.Templates["search.html"], td)
// 			return
// 		}

// 		totalCount, err := searchDbService.ReadCount(searchValue)
// 		if err != nil {
// 			repo.Logger.Errorln(err)
// 			rndr.Abort(w, r, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
// 			return
// 		}

// 		if totalCount < 1 {
// 			td.Data["Results"] = make([]string, 0)
// 			rndr.RenderTemplate(w, r, repo.Templates["search.html"], td)
// 			return
// 		}

// 		td.Data["TotalCount"] = totalCount

// 		//---- beyond this, there is data ---- //

// 		// calcuate total number of pages
// 		totalCountOfPages := int((totalCount + constants.PerPageConstant - 1) / constants.PerPageConstant)
// 		td.Data["Pages"] = number_utils.MakeSliceFromInt(totalCountOfPages)

// 		page, err := strconv.Atoi(r.URL.Query().Get("page"))
// 		if err != nil {
// 			page = 1
// 		}

// 		// protect the pagination
// 		if totalCountOfPages > 0 && page < 1 {
// 			page = 1
// 		} else if totalCountOfPages > 0 && page > totalCountOfPages {
// 			page = totalCountOfPages
// 		}

// 		td.Data["CurrentPage"] = page

// 		results, err := searchDbService.ReadPaginated(searchValue, page, constants.PerPageConstant)
// 		if err != nil {
// 			repo.Logger.Errorln(err)
// 			rndr.Abort(w, r, repo.Templates["error.html"], http.StatusInternalServerError, constants.Error500Title, constants.Error500Message)
// 			return
// 		}

// 		td.Data["Results"] = results

// 		rndr.RenderTemplate(w, r, repo.Templates["search.html"], td)
// 		return
// 	}

// 	// ========== POST ========= //
// 	if r.Method == http.MethodPost {
// 		search := r.FormValue("search")
// 		search = strings.TrimSpace(search)
// 		if len(search) < 1 {
// 			http.Redirect(w, r, "/search", http.StatusSeeOther)
// 			return
// 		} else {
// 			search = url.QueryEscape(search)
// 			http.Redirect(w, r, fmt.Sprintf("/search?q=%s", search), http.StatusSeeOther)
// 			return
// 		}
// 	}

// 	rndr.Abort(w, r, repo.Templates["error.html"], http.StatusMethodNotAllowed, constants.Error405Title, constants.Error405Message)

// }
