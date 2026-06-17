package handlers

import (
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/luqmanshaban/gomont/internals/api/web"
)

// PageHandler serves the server-rendered HTML client (login, signup,
// dashboard, settings) plus its static CSS/JS assets, all embedded into
// the binary via the web package.
type PageHandler struct {
	templates map[string]*template.Template
	static    http.Handler
}

// pageData is passed into every template execution. PageTitle fills the
// <title> tag in base.html; Active marks which nav link is highlighted
// (matched against the values used in partials/navbar.html, e.g. "dashboard").
type pageData struct {
	PageTitle string
	Active    string
}

// NewPageHandler parses every page template together with base.html and
// the shared navbar partial, so each page can be rendered independently
// at request time without re-parsing the filesystem on every request.
func NewPageHandler() *PageHandler {
	pages := []struct {
		name  string
		title string
	}{
		{"login", "Log in"},
		{"signup", "Sign up"},
		{"dashboard", "Dashboard"},
		{"settings", "Settings"},
	}

	templates := make(map[string]*template.Template)
	for _, p := range pages {
		tmpl, err := template.ParseFS(
			web.TemplatesFS,
			"templates/base.html",
			"templates/partials/navbar.html",
			"templates/"+p.name+".html",
		)
		if err != nil {
			// Templates are embedded at compile time, so a parse failure here
			// means a real bug shipped in the binary, not a runtime/config
			// issue. Fail loudly and immediately rather than serving broken
			// pages.
			slog.Error("failed to parse template", "page", p.name, "error", err)
			panic(err)
		}
		templates[p.name] = tmpl
	}

	staticSub, err := fs.Sub(web.StaticFS, "static")
	if err != nil {
		panic(err)
	}

	return &PageHandler{
		templates: templates,
		static:    http.FileServer(http.FS(staticSub)),
	}
}

// render executes the named page's template set against base.html.
func (h *PageHandler) render(w http.ResponseWriter, name, title, active string) {
	tmpl, ok := h.templates[name]
	if !ok {
		http.Error(w, "page not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "base", pageData{PageTitle: title, Active: active}); err != nil {
		slog.Error("failed to render template", "page", name, "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

func (h *PageHandler) Login(w http.ResponseWriter, r *http.Request) {
	h.render(w, "login", "Log in", "")
}

func (h *PageHandler) Signup(w http.ResponseWriter, r *http.Request) {
	h.render(w, "signup", "Sign up", "")
}

func (h *PageHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	h.render(w, "dashboard", "Dashboard", "dashboard")
}

func (h *PageHandler) Settings(w http.ResponseWriter, r *http.Request) {
	h.render(w, "settings", "Settings", "settings")
}

// Static serves embedded CSS/JS under the /static/ prefix. Register with
// http.StripPrefix("/static/", pageHandler.Static) on the mux.
func (h *PageHandler) Static() http.Handler {
	return h.static
}