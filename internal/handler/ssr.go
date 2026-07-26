package handler

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"blog/internal/service"
	"blog/internal/store"
)

type SSRHandler struct {
	svc  *service.PostService
	tmpl *template.Template
}

func NewSSR(svc *service.PostService, tmpl *template.Template) *SSRHandler {
	return &SSRHandler{svc: svc, tmpl: tmpl}
}

const postsPerPage = 10

func parseFilterParams(r *http.Request) store.FilterParams {
	var tags []string
	for _, t := range r.URL.Query()["tags"] {
		t = strings.TrimSpace(t)
		if t != "" {
			tags = append(tags, t)
		}
	}
	return store.FilterParams{
		Search:   r.URL.Query().Get("q"),
		Category: r.URL.Query().Get("category"),
		Tags:     tags,
		DateFrom: r.URL.Query().Get("from"),
		DateTo:   r.URL.Query().Get("to"),
	}
}

func filterQueryParams(f store.FilterParams) string {
	var parts []string
	if f.Search != "" {
		parts = append(parts, "q="+template.URLQueryEscaper(f.Search))
	}
	if f.Category != "" {
		parts = append(parts, "category="+template.URLQueryEscaper(f.Category))
	}
	if len(f.Tags) > 0 {
		for _, t := range f.Tags {
			parts = append(parts, "tags="+template.URLQueryEscaper(t))
		}
	}
	if f.DateFrom != "" {
		parts = append(parts, "from="+template.URLQueryEscaper(f.DateFrom))
	}
	if f.DateTo != "" {
		parts = append(parts, "to="+template.URLQueryEscaper(f.DateTo))
	}
	if len(parts) == 0 {
		return ""
	}
	return "&" + strings.Join(parts, "&")
}

func (h *SSRHandler) Home(w http.ResponseWriter, r *http.Request) {
	filter := parseFilterParams(r)
	after := r.URL.Query().Get("after")

	var paged *service.CursorPage
	var err error
	if filter.HasFilter() {
		paged, err = h.svc.GetCursorPostsFiltered(r.Context(), after, postsPerPage, filter)
	} else {
		paged, err = h.svc.GetCursorPosts(r.Context(), after, postsPerPage)
	}
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	categories, _ := h.svc.GetAllCategories(r.Context())
	tags, _ := h.svc.GetAllTags(r.Context())

	data := map[string]any{
		"Title":          "My Blog",
		"Posts":          paged.Posts,
		"Pagination":     paged,
		"ApiPath":        "/api/posts",
		"Filter":         filter,
		"FilterParams":   filterQueryParams(filter),
		"AllCategories":  categories,
		"AllTags":        tags,
		"Year":           time.Now().Year(),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.ExecuteTemplate(w, "home.html", data); err != nil {
		log.Printf("template error (home): %v", err)
	}
}

func (h *SSRHandler) APIPosts(w http.ResponseWriter, r *http.Request) {
	filter := parseFilterParams(r)
	after := r.URL.Query().Get("after")

	var paged *service.CursorPage
	var err error
	if filter.HasFilter() {
		paged, err = h.svc.GetCursorPostsFiltered(r.Context(), after, postsPerPage, filter)
	} else {
		paged, err = h.svc.GetCursorPosts(r.Context(), after, postsPerPage)
	}
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, paged)
}

func (h *SSRHandler) APICategoryPosts(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	filter := parseFilterParams(r)
	filter.Category = slug
	after := r.URL.Query().Get("after")

	paged, err := h.svc.GetCursorPostsFiltered(r.Context(), after, postsPerPage, filter)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, paged)
}

func (h *SSRHandler) APITagPosts(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	filter := parseFilterParams(r)
	if !contains(filter.Tags, slug) {
		filter.Tags = append(filter.Tags, slug)
	}
	after := r.URL.Query().Get("after")

	paged, err := h.svc.GetCursorPostsFiltered(r.Context(), after, postsPerPage, filter)
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
		log.Printf("template error (post): %v", err)
	}
}

func (h *SSRHandler) Category(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	category, err := h.svc.GetCategoryBySlug(r.Context(), slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	filter := parseFilterParams(r)
	filter.Category = slug
	after := r.URL.Query().Get("after")

	paged, err := h.svc.GetCursorPostsFiltered(r.Context(), after, postsPerPage, filter)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	tags, _ := h.svc.GetAllTags(r.Context())

	data := map[string]any{
		"Title":        category.Name + " - My Blog",
		"Posts":        paged.Posts,
		"PageHeader":   "Category: " + category.Name,
		"Pagination":   paged,
		"ApiPath":      "/api/categories/" + slug + "/posts",
		"Filter":       filter,
		"FilterParams": filterQueryParams(filter),
		"AllTags":      tags,
		"Year":         time.Now().Year(),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.ExecuteTemplate(w, "home.html", data); err != nil {
		log.Printf("template error (category): %v", err)
	}
}

func (h *SSRHandler) Tag(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	tag, err := h.svc.GetTagBySlug(r.Context(), slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	filter := parseFilterParams(r)
	if !contains(filter.Tags, slug) {
		filter.Tags = append(filter.Tags, slug)
	}
	after := r.URL.Query().Get("after")

	paged, err := h.svc.GetCursorPostsFiltered(r.Context(), after, postsPerPage, filter)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	tags, _ := h.svc.GetAllTags(r.Context())

	data := map[string]any{
		"Title":        tag.Name + " - My Blog",
		"Posts":        paged.Posts,
		"PageHeader":   "Tag: " + tag.Name,
		"Pagination":   paged,
		"ApiPath":      "/api/tags/" + slug + "/posts",
		"Filter":       filter,
		"FilterParams": filterQueryParams(filter),
		"AllTags":      tags,
		"Year":         time.Now().Year(),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.ExecuteTemplate(w, "home.html", data); err != nil {
		log.Printf("template error (tag): %v", err)
	}
}

func (h *SSRHandler) Categories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.svc.GetAllCategories(r.Context())
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"Title":      "Categories - My Blog",
		"Categories": categories,
		"Year":       time.Now().Year(),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.ExecuteTemplate(w, "categories.html", data); err != nil {
		log.Printf("template error (categories): %v", err)
	}
}

func (h *SSRHandler) Tags(w http.ResponseWriter, r *http.Request) {
	tags, err := h.svc.GetAllTags(r.Context())
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"Title": "Tags - My Blog",
		"Tags":  tags,
		"Year":  time.Now().Year(),
	}

	if err := h.tmpl.ExecuteTemplate(w, "tags.html", data); err != nil {
		log.Printf("template error (tags): %v", err)
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
