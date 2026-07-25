package handler

import (
	"html/template"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"blog/internal/service"
)

type SSRHandler struct {
	svc  *service.PostService
	tmpl *template.Template
}

func NewSSR(svc *service.PostService, tmpl *template.Template) *SSRHandler {
	return &SSRHandler{svc: svc, tmpl: tmpl}
}

func (h *SSRHandler) Home(w http.ResponseWriter, r *http.Request) {
	posts, err := h.svc.GetAllPosts()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"Title": "My Blog",
		"Posts": posts,
		"Year":  time.Now().Year(),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.ExecuteTemplate(w, "home.html", data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *SSRHandler) Post(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	post, err := h.svc.GetPostBySlug(slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	htmlContent, err := h.svc.RenderMarkdown(post.Markdown)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"Title": post.Title + " - My Blog",
		"Date":  post.Date,
		"HTML":  template.HTML(htmlContent),
		"Year":  time.Now().Year(),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.ExecuteTemplate(w, "post.html", data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
