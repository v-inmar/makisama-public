package server

import (
	"html/template"
	"net/http"
	"os"

	"app/internal/constants"
	"app/internal/handlers"
	middlewares "app/internal/middleware"
	api_utils "app/internal/utilities/api"
	csrf_utils "app/internal/utilities/csrf"
	template_utils "app/internal/utilities/templates"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type Routes struct {
	Router chi.Router
	Config *Config
}

func NewRoutes(r chi.Router, c *Config) *Routes {
	return &Routes{
		Router: r,
		Config: c,
	}
}

// Config acts as a clean configuration package to boot up our network shell
type Config struct {
	DB        *sqlx.DB
	Session   *scs.SessionManager
	Templates map[string]*template.Template
	Logger    *zap.SugaredLogger
}

// ConfigureRoutes sets up the routing middleware chain and maps paths to handlers
func ConfigureRoutes(cfg *Config) http.Handler {
	router := chi.NewRouter()

	router.Use(middleware.RealIP)

	// default middlewares
	router.Use(middleware.Recoverer)

	env := os.Getenv("ENVIRONMENT")
	if env != "production" {
		router.Use(middleware.Logger)
	}
	router.Use(middleware.RedirectSlashes)

	// router.Use(csrf_utils.DebugRequest)
	// session middleware
	router.Use(cfg.Session.LoadAndSave)

	// csrf middleware
	// router.Use(nosurf.NewPure)
	// router.Use(nosurf.NewPure)
	router.Use(csrf_utils.NoSurf)

	// // create the inject template data middleware
	injectTemplateDataMiddleware := middlewares.NewInjectTemplateDataMiddleware(cfg.Session)
	router.Use(injectTemplateDataMiddleware.InjectTemplateData)
	// router.Use(middlewares.InjectTemplateData(cfg.Session))

	// For css and js in static directory
	fileServer := http.FileServer(http.Dir("./static"))
	router.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	h := handlers.NewRepo(cfg.Templates, cfg.Session, cfg.DB, cfg.Logger)

	abuseProneLimit := middlewares.NewRateLimit(constants.RateLimit / 6)
	semiRelaxedLimit := middlewares.NewRateLimit(constants.RateLimit / 2)
	relaxedLimit := middlewares.NewRateLimit(constants.RateLimit)
	helperLimit := middlewares.NewRateLimit(constants.RateLimit + 10)
	router.Group(func(r chi.Router) {
		r.Use(abuseProneLimit.ExecuteLimit(cfg.Templates["error.html"]))
		r.Get("/auth/signup", h.GetAuthSignup)
		r.Post("/auth/signup", h.PostAuthSignup)
		r.Get("/auth/login", h.GetAuthLogin)
		r.Post("/auth/login", h.PostAuthLogin)
		r.Get("/auth/forgot_password", h.GetAuthForgot)
		r.Post("/auth/forgot_password", h.PostAuthForgot)
		r.Get("/auth/reset_password", h.GetResetPassword)
		r.Post("/auth/reset_password", h.PostResetPassword)
		r.Post("/auth/logout", h.PostAuthLogout)
		r.Get("/auth/confirm_email", h.GetAuthConfirmEmail)
		r.Get("/auth/request_confirm", h.GetAuthRequestConfirm)
		r.Post("/auth/request_confirm", h.PostAuthRequestConfirm)
		r.Get("/company/contact", h.GetCompanyContact)
		r.Post("/company/contact", h.PostCompanyContact)
		r.Get("/feedbacks/new", h.GetFeedbackCreate)
		r.Post("/feedbacks/new", h.PostFeedbackCreate)
		r.Get("/report/boards", h.GetReportBoard)
		r.Post("/report/boards", h.PostReportBoard)
		r.Get("/report/feedbacks", h.GetReportFeedback)
		r.Post("/report/feedbacks", h.PostReportFeedback)
	})

	router.Group(func(r chi.Router) {
		r.Use(semiRelaxedLimit.ExecuteLimit(cfg.Templates["error.html"]))
		r.Get("/", h.GetHomeHome)
		r.Get("/{slug}", h.GetBoardRead)
		r.Get("/company/about", h.GetCompanyAbout)
		r.Get("/company/privacy", h.GetCompanyPrivacyPolicy)
		r.Get("/company/terms", h.GetCompanyTermsOfService)
		r.Get("/feedbacks/{slug}", h.GetFeedbackRead)
		r.Get("/files/images/{filename}", h.GetFilesRead)
		r.Get("/search", h.GetSearch)
		r.Post("/search", h.PostSearch)
		r.Get("/api/feedbacks/{slug}", h.GetApiFeedback)
	})

	router.Group(func(r chi.Router) {
		r.Use(relaxedLimit.ExecuteLimit(cfg.Templates["error.html"]))
		r.Get("/dashboard", h.GetHomeDashboard)
		r.Get("/boards/new", h.GetBoardCreateBoard)
		r.Post("/boards/new", h.PostBoardCreateBoard)
		r.Post("/{slug}/delete", h.PostBoardDeleteBoard)
		r.Get("/{slug}/edit", h.GetBoardEditBoard)
		r.Post("/{slug}/edit", h.PostBoardEditBoard)
		r.Get("/boards", h.GetBoardReads)
		r.Post("/feedbacks/delete", h.PostFeedbackDelete)
		r.Get("/feedbacks/pending", h.GetFeedbackPending)
		r.Get("/feedbacks/review", h.GetFeedbackReview)
		r.Post("/feedbacks/review", h.PostFeedbackReview)
		r.Post("/api/feedbacks", h.PostApiFeedback)
		r.Get("/accounts", h.GetUserAccount)
		r.Get("/accounts/deactivate", h.GetUserAccountDeactivate)
		r.Post("/accounts/deactivate", h.PostUserAccountDeactivate)
		r.Get("/accounts/delete", h.GetUserAccountDelete)
		r.Post("/accounts/delete", h.PostUserAccountDelete)
		r.Get("/accounts/edit_name", h.GetUserAccountEditName)
		r.Post("/accounts/edit_name", h.PostUserAccountEditName)
		r.Get("/accounts/logout_all", h.GetUserAccountLogoutAll)
		r.Post("/accounts/logout_all", h.PostUserAccountLogoutAll)
		r.Get("/accounts/replace_email", h.GetUserAccountReplaceEmail)
		r.Post("/accounts/replace_email", h.PostUserAccountReplaceEmail)
		r.Get("/accounts/update_password", h.GetUserAccountUpdatePassword)
		r.Post("/accounts/update_password", h.PostUserAccountUpdatePassword)

	})

	router.Group(func(r chi.Router) {
		r.Use(helperLimit.ExecuteLimit(cfg.Templates["error.html"]))
		r.Get("/api/version", h.GetApiVersion)
	})

	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		statusCode := http.StatusNotFound
		apiService := api_utils.NewApiResponseService(w)
		if apiService.IsRequestApi(r) {
			apiService.Abort(statusCode, constants.Error404Title, constants.Error404Message)
		} else {
			template_utils.Abort(w, cfg.Templates["error.html"], statusCode, constants.Error404Title, constants.Error404Message)
		}
	})
	router.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		statusCode := http.StatusMethodNotAllowed
		apiService := api_utils.NewApiResponseService(w)
		if apiService.IsRequestApi(r) {
			apiService.Abort(statusCode, constants.Error405Title, constants.Error405Message)
		} else {
			template_utils.Abort(w, cfg.Templates["error.html"], statusCode, constants.Error405Title, constants.Error405Message)
		}
	})

	return router
}
