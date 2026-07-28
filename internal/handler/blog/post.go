package blog

import (
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

func (h *SSRHandler) Post(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	post, err := h.svc.GetPostBySlug(r.Context(), slug)
	if err != nil {
		h.NotFound(w, r)
		return
	}

	htmlContent, err := h.svc.RenderMarkdown(post.Markdown)
	if err != nil {
		h.serverError(w, r, err)
		return
	}

	data := map[string]any{
		"Title":    post.Title + " - My Blog",
		"Post":     post,
		"PostDate": post.Date,
		"HTML":     template.HTML(htmlContent),
		"Year":     time.Now().Year(),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.ExecuteTemplate(w, "post.html", data); err != nil {
		log.Printf("template error (post): %v", err)
	}
}

func (h *SSRHandler) Category(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	http.Redirect(w, r, "/?category="+template.URLQueryEscaper(slug), http.StatusMovedPermanently)
}

func (h *SSRHandler) Tag(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	http.Redirect(w, r, "/?tags="+template.URLQueryEscaper(slug), http.StatusMovedPermanently)
}
