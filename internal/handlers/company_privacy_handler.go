package handlers

import (
	"app/internal/constants"
	template_utils "app/internal/utilities/templates"
	"net/http"
)

func (repo *HandlerRepo) GetCompanyPrivacyPolicy(w http.ResponseWriter, r *http.Request) {
	td := r.Context().Value(constants.TemplateDataContextKey).(*template_utils.TemplateData)
	template_utils.RenderTemplate(w, repo.Templates["privacy_policy.html"], td)
}
