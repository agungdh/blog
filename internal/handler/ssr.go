package handler

import (
	"encoding/json"
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

const postsPerPage = 10

func (h *SSRHandler) Home(w http.ResponseWriter, r *http.Request) {
	after := r.URL.Query().Get("after")

	paged, err := h.svc.GetCursorPosts(r.Context(), after, postsPerPage)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"Title":      "My Blog",
		"Posts":      paged.Posts,
		"Pagination": paged,
		"ApiPath":    "/api/posts",
		"Year":       time.Now().Year(),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.ExecuteTemplate(w, "home.html", data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *SSRHandler) APIPosts(w http.ResponseWriter, r *http.Request) {
	after := r.URL.Query().Get("after")
	paged, err := h.svc.GetCursorPosts(r.Context(), after, postsPerPage)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, paged)
}

func (h *SSRHandler) APICategoryPosts(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	after := r.URL.Query().Get("after")
	paged, err := h.svc.GetCursorPostsByCategorySlug(r.Context(), slug, after, postsPerPage)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, paged)
}

func (h *SSRHandler) APITagPosts(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	after := r.URL.Query().Get("after")
	paged, err := h.svc.GetCursorPostsByTagSlug(r.Context(), slug, after, postsPerPage)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, paged)
}

type jsonPost struct {
	Slug     string    `json:"slug"`
	Title    string    `json:"title"`
	Date     string    `json:"date"`
	Summary  string    `json:"summary"`
	Category *jsonCat  `json:"category,omitempty"`
	Tags     []jsonTag `json:"tags,omitempty"`
}

type jsonCat struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type jsonTag struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func writeJSON(w http.ResponseWriter, paged *service.CursorPage) {
	posts := make([]jsonPost, len(paged.Posts))
	for i, p := range paged.Posts {
		jp := jsonPost{
			Slug:    p.Slug,
			Title:   p.Title,
			Date:    p.Date,
			Summary: p.Summary,
		}
		if p.Category != nil {
			jp.Category = &jsonCat{Name: p.Category.Name, Slug: p.Category.Slug}
		}
		tags := make([]jsonTag, len(p.Tags))
		for j, t := range p.Tags {
			tags[j] = jsonTag{Name: t.Name, Slug: t.Slug}
		}
		jp.Tags = tags
		posts[i] = jp
	}

	resp := map[string]any{
		"posts":     posts,
		"has_next":  paged.HasNext,
		"next_slug": paged.NextSlug,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
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

	after := r.URL.Query().Get("after")

	paged, err := h.svc.GetCursorPostsByCategorySlug(r.Context(), slug, after, postsPerPage)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"Title":      category.Name + " - My Blog",
		"Posts":      paged.Posts,
		"PageHeader": "Category: " + category.Name,
		"Pagination": paged,
		"ApiPath":    "/api/categories/" + slug + "/posts",
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

	after := r.URL.Query().Get("after")

	paged, err := h.svc.GetCursorPostsByTagSlug(r.Context(), slug, after, postsPerPage)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"Title":      tag.Name + " - My Blog",
		"Posts":      paged.Posts,
		"PageHeader": "Tag: " + tag.Name,
		"Pagination": paged,
		"ApiPath":    "/api/tags/" + slug + "/posts",
		"Year":       time.Now().Year(),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.ExecuteTemplate(w, "home.html", data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
