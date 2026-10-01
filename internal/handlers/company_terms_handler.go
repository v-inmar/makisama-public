package handlers

import (
	"app/internal/constants"
	template_utils "app/internal/utilities/templates"
	"net/http"
)

func (repo *HandlerRepo) GetCompanyTermsOfService(w http.ResponseWriter, r *http.Request) {
	td := r.Context().Value(constants.TemplateDataContextKey).(*template_utils.TemplateData)
	template_utils.RenderTemplate(w, repo.Templates["terms_of_service.html"], td)
}
