package handlers

import (
	"app/internal/constants"
	strings_utils "app/internal/utilities/strings"
	template_utils "app/internal/utilities/templates"
	"net/http"
)

func (repo *HandlerRepo) GetCompanyAbout(w http.ResponseWriter, r *http.Request) {
	td := r.Context().Value(constants.TemplateDataContextKey).(*template_utils.TemplateData)
	scopeList := make([]string, 0)
	for k := range constants.FeedbackScopes {
		scopeList = append(scopeList, strings_utils.Capitalize(k))
	}

	resolutionList := make([]string, 0)
	for k := range constants.FeedbackResolutions {
		resolutionList = append(resolutionList, strings_utils.Capitalize(k))
	}
	td.Data["Resolutions"] = resolutionList
	td.Data["Scopes"] = scopeList
	template_utils.RenderTemplate(w, repo.Templates["about.html"], td)
}
