package handler

import (
	"html/template"
	"net/http"
	"strconv"
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

func parsePage(r *http.Request) int {
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		return 1
	}
	return page
}

const postsPerPage = 10

func (h *SSRHandler) Home(w http.ResponseWriter, r *http.Request) {
	page := parsePage(r)

	paged, err := h.svc.GetPagedPosts(r.Context(), page, postsPerPage)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"Title":      "My Blog",
		"Posts":      paged.Posts,
		"Pagination": paged,
		"Year":       time.Now().Year(),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.ExecuteTemplate(w, "home.html", data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *SSRHandler) Post(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	post, err := h.svc.GetPostBySlug(r.Context(), slug)
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
		"Title":    post.Title + " - My Blog",
		"Post":     post,
		"PostDate": post.Date,
		"HTML":     template.HTML(htmlContent),
		"Year":     time.Now().Year(),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.ExecuteTemplate(w, "post.html", data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *SSRHandler) Category(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	category, err := h.svc.GetCategoryBySlug(r.Context(), slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	page := parsePage(r)
	paged, err := h.svc.GetPagedPostsByCategorySlug(r.Context(), slug, page, postsPerPage)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"Title":      category.Name + " - My Blog",
		"Posts":      paged.Posts,
		"PageHeader": "Category: " + category.Name,
		"Pagination": paged,
		"Year":       time.Now().Year(),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.ExecuteTemplate(w, "home.html", data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *SSRHandler) Tag(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	tag, err := h.svc.GetTagBySlug(r.Context(), slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	page := parsePage(r)
	paged, err := h.svc.GetPagedPostsByTagSlug(r.Context(), slug, page, postsPerPage)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"Title":      tag.Name + " - My Blog",
		"Posts":      paged.Posts,
		"PageHeader": "Tag: " + tag.Name,
		"Pagination": paged,
		"Year":       time.Now().Year(),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.ExecuteTemplate(w, "home.html", data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
