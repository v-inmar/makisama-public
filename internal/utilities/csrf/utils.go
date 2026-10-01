package csrf_utils

import (
	api_utils "app/internal/utilities/api"
	template_utils "app/internal/utilities/templates"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/justinas/nosurf"
)

// func NoSurf(next http.Handler) http.Handler {
// 	csrf := nosurf.NewPure(next)

// 	// csrf.SetBaseCookie(http.Cookie{
// 	// 	HttpOnly: true,
// 	// 	SameSite: http.SameSiteLaxMode,
// 	// 	Path:     "/",
// 	// 	MaxAge:   60 * 60 * 12, // 12 hours
// 	// 	Secure:   false,        // TODO: Keep false for localhost development without HTTPS, make true in prod
// 	// })

// 	// csrf.SetFailureHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 	// 	fmt.Println(r.Cookies())

// 	// 	statuseCode := http.StatusBadRequest
// 	// 	title := "Security Token Invalid"
// 	// 	msg := "Your security token could not be verified. This usually happens if you stay on a page for too long or modify the form improperly. Please refresh and try again."

// 	// 	isApi := false
// 	// 	pathList := strings.Split(r.URL.Path, "/")
// 	// 	if len(pathList) > 1 {
// 	// 		if strings.ToLower(pathList[1]) == "api" {
// 	// 			isApi = true
// 	// 		}
// 	// 	}

// 	// 	if isApi {
// 	// 		api_utils.ApiAbort(w, statuseCode, title, msg)
// 	// 	} else {
// 	// 		td := &template_utils.TemplateData{
// 	// 			IsAuthenticated: false,
// 	// 			Data: map[string]any{
// 	// 				"StatusCode": statuseCode,
// 	// 				"Title":      title,
// 	// 				"Message":    msg,
// 	// 			},
// 	// 		}
// 	// 		tmpl := template.Must(template.ParseFiles(
// 	// 			"templates/base.html",
// 	// 			"templates/error.html",
// 	// 		))

// 	// 		err := tmpl.ExecuteTemplate(w, "base", td)
// 	// 		if err != nil {
// 	// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 	// 		}
// 	// 	}

// 	// }))

// 	return csrf
// }

// TODO: put this is middleware package
// func NoSurf(next http.Handler) http.Handler {
// 	csrf := nosurf.New(next)

// 	csrf.SetBaseCookie(http.Cookie{
// 		HttpOnly: true,
// 		SameSite: http.SameSiteLaxMode,
// 		Path:     "/",
// 		MaxAge:   60 * 60 * 12, // 12 hours
// 		Secure:   false,        // TODO: change to true on prod
// 	})

// 	csrf.SetFailureHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

// 		println("=== CSRF FAIL ===")
// 		println("Method:", r.Method)
// 		println("URL:", r.URL.String())
// 		println("Cookie:", r.Header.Get("Cookie"))
// 		println("Form token:", r.FormValue("csrf_token"))

// 		http.Error(w, "CSRF FAILED", http.StatusBadRequest)

// 		apiService := api_utils.NewApiResponseService(w)
// 		statuseCode := http.StatusBadRequest
// 		title := "Security Token Invalid"
// 		msg := "Your security token could not be verified. Please refresh and try again."

// 		isApi := apiService.IsRequestApi(r)

// 		if isApi {
// 			apiService.Abort(statuseCode, title, msg)
// 		} else {
// 			td := &template_utils.TemplateData{
// 				IsAuthenticated: false,
// 				Data: map[string]any{
// 					"StatusCode": statuseCode,
// 					"Title":      title,
// 					"Message":    msg,
// 				},
// 			}
// 			tmpl := template.Must(template.ParseFiles("templates/base.html", "templates/error.html"))
// 			_ = tmpl.ExecuteTemplate(w, "base", td)
// 		}
// 	}))

// 	return csrf
// }

// func NoSurf(next http.Handler) http.Handler {
// 	csrf := nosurf.New(next)

// 	csrf.SetFailureHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

// 		println("=== CSRF FAIL ===")
// 		println("Cookie:", r.Header.Get("Cookie"))
// 		println("Form token:", r.FormValue("csrf_token"))
// 		println("Method:", r.Method)
// 		println("URL:", r.URL.String())
// 		println("Origin:", r.Header.Get("Origin"))
// 		println("Referer:", r.Header.Get("Referer"))

// 		http.Error(w, "CSRF FAILED", http.StatusBadRequest)
// 	}))

// 	return csrf
// }

// func DebugRequest(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

// 		println("=== REQUEST ===")
// 		println("METHOD:", r.Method)
// 		println("URL:", r.URL.String())
// 		println("HOST:", r.Host)
// 		println("ORIGIN:", r.Header.Get("Origin"))
// 		println("REFERER:", r.Header.Get("Referer"))
// 		println("COOKIE:", r.Header.Get("Cookie"))

// 		next.ServeHTTP(w, r)
// 	})
// }

func NoSurf(next http.Handler) http.Handler {
	csrf := nosurf.New(next)

	secure := false
	if os.Getenv("ENVIRONMENT") == "production" {
		secure = true
	}

	csrf.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   60 * 60 * 12,
		Secure:   secure, // true in production (HTTPS)
	})

	// TODO: remove me
	// if os.Getenv("ENVIRONMENT") != "production" {
	// 	org, err := nosurf.StaticOrigins("http://18.175.240.151:5000")
	// 	if err != nil {
	// 		log.Println(err)
	// 	}
	// 	csrf.SetIsAllowedOriginFunc(org)
	// }

	// Create the custom failure handler
	csrf.SetFailureHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(nosurf.Reason(r))
		apiService := api_utils.NewApiResponseService(w)
		statuseCode := http.StatusBadRequest
		title := "Security Token Invalid"
		msg := "Your security token could not be verified. Please refresh and try again."

		isApi := apiService.IsRequestApi(r)

		if isApi {
			apiService.Abort(statuseCode, title, msg)
		} else {
			td := &template_utils.TemplateData{
				Data: map[string]any{
					"StatusCode": statuseCode,
					"Title":      title,
					"Message":    msg,
				},
			}
			tmpl := template.Must(template.ParseFiles("templates/base.html", "templates/error.html"))
			_ = tmpl.ExecuteTemplate(w, "base", td)
		}
	}))

	return csrf
}
