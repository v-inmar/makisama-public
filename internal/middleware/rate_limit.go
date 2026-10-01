package middlewares

import (
	"app/internal/constants"
	api_utils "app/internal/utilities/api"
	template_utils "app/internal/utilities/templates"
	"html/template"
	"net/http"
	"time"

	"github.com/go-chi/httprate"
)

type RateLimit struct {
	NumOfReq int
}

func NewRateLimit(numOfReq int) *RateLimit {
	return &RateLimit{
		NumOfReq: numOfReq,
	}
}

func (rl *RateLimit) ExecuteLimit(tmpl *template.Template) func(next http.Handler) http.Handler {
	return httprate.Limit(
		rl.NumOfReq, // requests
		1*time.Minute,
		httprate.WithKeyFuncs(httprate.KeyByRealIP),
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			td := r.Context().Value(constants.TemplateDataContextKey).(*template_utils.TemplateData)
			apiService := api_utils.NewApiResponseService(w)
			statuseCode := http.StatusTooManyRequests

			if apiService.IsRequestApi(r) {
				apiService.Abort(statuseCode, constants.Error429Title, constants.Error429Message)
			} else {

				td.Data["StatusCode"] = statuseCode
				td.Data["Title"] = constants.Error429Title
				td.Data["Message"] = constants.Error429Message

				template_utils.RenderTemplate(w, tmpl, td)
			}
		}),
	)
}
