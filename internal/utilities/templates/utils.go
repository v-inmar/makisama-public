package template_utils

import (
	"app/internal/constants"
	date_utils "app/internal/utilities/dates"
	number_utils "app/internal/utilities/number"
	strings_utils "app/internal/utilities/strings"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"

	"github.com/alexedwards/scs/v2"
	"github.com/justinas/nosurf"
)

type TemplateData struct {
	IsAuthenticated bool
	CSRFToken       string
	CaptchaSiteKey  string
	Version         string
	Data            map[string]any
}

// func NewTemplateData(r *http.Request) *TemplateData {
// 	td, ok := r.Context().Value(constants.TemplateDataContextKey).(*TemplateData)
// 	if !ok {
// 		// Fallback safe initialization if the middleware didn't run for some reason
// 		return &TemplateData{}
// 	}

// 	// Initialize the inner page map so handlers don't panic when trying to assign values to it
// 	if td.Data == nil {
// 		td.Data = make(map[string]any)
// 	}

// 	return td
// }

func ParseTemplates() map[string]*template.Template {
	tmplts := make(map[string]*template.Template)

	formatNullTime := template.FuncMap{
		"formatNullTime": date_utils.FormatNullTime,
	}

	formatNullString := template.FuncMap{
		"formatNullString": strings_utils.FormatNullString,
	}

	addFuncMap := template.FuncMap{
		"plusplus": number_utils.Addition,
	}

	subtractFuncMap := template.FuncMap{
		"minusminus": number_utils.Subtraction,
	}

	formatTime := template.FuncMap{
		"formatTime": date_utils.FormatTime,
	}

	capitalize := template.FuncMap{
		"capitalize": strings_utils.Capitalize,
	}

	var files []string

	filepath.WalkDir("templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			panic(err)
		}

		if !d.IsDir() {
			files = append(files, path)
		}
		return nil
	})

	for _, file := range files {
		name := filepath.Base(file)

		if name == "base.html" {
			continue
		}

		ts, err := template.New(name).
			Funcs(formatNullTime).
			Funcs(formatTime).
			Funcs(addFuncMap).
			Funcs(subtractFuncMap).
			Funcs(capitalize).
			Funcs(formatNullString).
			ParseFiles("templates/base.html", file)

		if err != nil {
			panic(err)
		}

		tmplts[name] = ts
	}

	return tmplts
}

type TemplateRenderer struct {
	Session *scs.SessionManager
}

func NewTemplateRenderer(s *scs.SessionManager) *TemplateRenderer {
	return &TemplateRenderer{
		Session: s,
	}
}

// func (rnd *TemplateRenderer) NewTemplateData() *TemplateData {
// 	td, ok := rnd.Request.Context().Value(constants.TemplateDataContextKey).(*TemplateData)
// 	if !ok {
// 		// Fallback safe initialization if the middleware didn't run for some reason
// 		return &TemplateData{}
// 	}

// 	// Initialize the inner page map so handlers don't panic when trying to assign values to it
// 	if td.Data == nil {
// 		td.Data = make(map[string]any)
// 	}

// 	return td
// }

func (rnd *TemplateRenderer) Abort(w http.ResponseWriter, r *http.Request, tmpl *template.Template, statusCode int, title, message string) {

	w.WriteHeader(statusCode) // Sets the proper HTTP network status code
	td := &TemplateData{
		Data: map[string]any{
			"StatusCode": statusCode,
			"Title":      title,
			"Message":    message,
		},
	}

	_, ok := rnd.Session.Get(r.Context(), constants.SessionManagerSessionKey).(string)
	if !ok {
		td.IsAuthenticated = false
	} else {
		td.IsAuthenticated = true
	}

	// pass a standalone TemplateData struct manually here since requests
	// targeting global 404/405 states might bypass the regular middleware group chain.
	// td := &TemplateData{
	// 	IsAuthenticated: false,
	// 	Data: map[string]any{
	// 		"StatusCode": statusCode,
	// 		"Title":      title,
	// 		"Message":    message,
	// 	},
	// }

	// err := tmpl.ExecuteTemplate(w, "base", td)
	rnd.RenderTemplate(w, r, tmpl, td)
	// err := util.Template.ExecuteTemplate(util.ResponseWriter, "base", td)
	// if err != nil {
	// 	// will produce a text-based html page error
	// 	http.Error(util.ResponseWriter, "Server error. Try again later", http.StatusInternalServerError)
	// }
}

func (rnd *TemplateRenderer) RenderTemplate(w http.ResponseWriter, r *http.Request, tmpl *template.Template, templateData *TemplateData) {
	// templateData.CSRFToken = nosurf.Token(r)
	err := tmpl.ExecuteTemplate(w, "base", templateData)
	if err != nil {
		// will produce a text-based html page error
		http.Error(w, "Server error. Try again later", http.StatusInternalServerError)
	}
}

func RenderTemplate(w http.ResponseWriter, tmpl *template.Template, templateData *TemplateData) {
	// templateData.CSRFToken = nosurf.Token(r)
	err := tmpl.ExecuteTemplate(w, "base", templateData)
	if err != nil {
		// will produce a text-based html page error
		http.Error(w, "Server error. Try again later", http.StatusInternalServerError)
	}
}

func Abort(w http.ResponseWriter, tmpl *template.Template, statusCode int, title, message string) {

	w.WriteHeader(statusCode) // Sets the proper HTTP network status code
	td := &TemplateData{
		Data: map[string]any{
			"StatusCode": statusCode,
			"Title":      title,
			"Message":    message,
		},
	}
	RenderTemplate(w, tmpl, td)

}

func AddDefaultData(td *TemplateData, r *http.Request, session *scs.SessionManager) *TemplateData {
	if td == nil {
		td = &TemplateData{}
	}
	td.CSRFToken = nosurf.Token(r)
	td.IsAuthenticated = session.Exists(r.Context(), constants.SessionManagerSessionKey)
	return td
}
