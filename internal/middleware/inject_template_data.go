package middlewares

import (
	"app/internal/constants"
	template_utils "app/internal/utilities/templates"
	"app/internal/version"
	"context"
	"net/http"
	"os"

	"github.com/alexedwards/scs/v2"
	"github.com/justinas/nosurf"
)

type InjectTemplateDataMiddleware struct {
	Session *scs.SessionManager
}

func NewInjectTemplateDataMiddleware(session *scs.SessionManager) *InjectTemplateDataMiddleware {
	return &InjectTemplateDataMiddleware{
		Session: session,
	}
}

func (m *InjectTemplateDataMiddleware) InjectTemplateData(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		td := &template_utils.TemplateData{
			IsAuthenticated: m.Session.Exists(r.Context(), constants.SessionManagerSessionKey),
			CSRFToken:       nosurf.Token(r),
			Version:         version.Version,
			CaptchaSiteKey:  os.Getenv("TURNSTILE_SITEKEY"),
			Data:            make(map[string]any),
		}

		if m.Session.Exists(r.Context(), constants.FlashSessionKey) {
			td.Data["Flash"] = m.Session.Pop(r.Context(), constants.FlashSessionKey)
		}

		ctx := context.WithValue(r.Context(), constants.TemplateDataContextKey, td)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// InjectTemplateData is our assembly line middleware step.
// It creates a blank TemplateData structure on every request, fills in security tokens,
// and passes it down so our handlers can read it cleanly.
// func (m *InjectTemplateDataMiddleware) InjectTemplateData(next http.Handler) http.Handler {
// 	// Immediately drop into the runtime request execution block
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		td := &template_utils.TemplateData{
// 			IsAuthenticated: m.Session.Exists(r.Context(), constants.SessionManagerSessionKey),
// 			CSRFToken:       nosurf.Token(r),
// 			CaptchaSiteKey:  os.Getenv("TURNSTILE_SITEKEY"),
// 			Data:            make(map[string]any),
// 		}

// 		if m.Session.Exists(r.Context(), constants.FlashSessionKey) {

// 			td.Data["Flash"] = m.Session.Pop(r.Context(), constants.FlashSessionKey)

// 		}

// 		// Securely attach the model context to the request background life cycle
// 		ctx := context.WithValue(r.Context(), "templateData", td)
// 		next.ServeHTTP(w, r.WithContext(ctx))

// 	})
// }

// func InjectTemplateData(session *scs.SessionManager) func(http.Handler) http.Handler {
// 	return func(next http.Handler) http.Handler {
// 		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

// 			td := &template_utils.TemplateData{
// 				IsAuthenticated: session.Exists(r.Context(), "user"),
// 				CaptchaSiteKey:  os.Getenv("TURNSTILE_SITEKEY"),
// 				Data:            map[string]any{},
// 			}

// 			// IMPORTANT: safe here because request is already wrapped by nosurf
// 			td.CSRFToken = nosurf.Token(r)

// 			ctx := context.WithValue(r.Context(), constants.TemplateDataContextKey, td)

// 			next.ServeHTTP(w, r.WithContext(ctx))
// 		})
// 	}
// }

/*
func (m *InjectTemplateDataMiddleware) InjectTemplateData(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		td := &models.TemplateData{
			// ... data setup
		}

		//  CORRECT: Pass r.Context() as parent to preserve everything Chi is doing
		ctx := context.WithValue(r.Context(), "templateData", td)

		//  CORRECT: Send the updated context along the wire
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

*/

// func (m *InjectTemplateDataMiddleware) InjectTemplateData() func(http.Handler) http.Handler {
// 	return func(next http.Handler) http.Handler {
// 		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			td := &models.TemplateData{
// 				IsAuthenticated: m.Session.Exists(r.Context(), constants.SessionManagerSessionKey),
// 				CSRFToken:       nosurf.Token(r),
// 				Data:            make(map[string]any),
// 			}

// 			// if m.Session.Exists(r.Context(), constants.SessionManagerSessionKey) {
// 			// 	td.IsAuthenticated = true
// 			// }

// 			// if m.Session.Exists(r.Context(), "flash") {
// 			// 	td.Data["Flash"] = m.Session.Pop(r.Context(), "flash")
// 			// }

// 			// Securely attach the model context to the request background life cycle
// 			ctx := context.WithValue(r.Context(), "templateData", td)
// 			next.ServeHTTP(w, r.WithContext(ctx))
// 		})
// 	}
// }
