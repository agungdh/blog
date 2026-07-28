package blog

import (
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"blog/internal/service"
	"blog/internal/store"
)

const postsPerPage = 10

type AssetHashes struct {
	CSS             string
	InfiniteScroll  string
	SearchableFilter string
}

type SSRHandler struct {
	svc         *service.PostService
	tmpl        *template.Template
	AssetHashes AssetHashes
}

func NewSSR(svc *service.PostService, tmpl *template.Template, hashes AssetHashes) *SSRHandler {
	return &SSRHandler{svc: svc, tmpl: tmpl, AssetHashes: hashes}
}

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

func (h *SSRHandler) NotFound(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	data := map[string]any{
		"Title":       "404 - My Blog",
		"Year":        time.Now().Year(),
		"AssetHashes": h.AssetHashes,
	}
	if err := h.tmpl.ExecuteTemplate(w, "404.html", data); err != nil {
		log.Printf("template error (404): %v", err)
	}
}

func (h *SSRHandler) serverError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("server error: %s %s: %v", r.Method, r.URL.Path, err)
	w.WriteHeader(http.StatusInternalServerError)
	data := map[string]any{
		"Title":       "500 - My Blog",
		"Year":        time.Now().Year(),
		"AssetHashes": h.AssetHashes,
	}
	if tmplErr := h.tmpl.ExecuteTemplate(w, "500.html", data); tmplErr != nil {
		log.Printf("template error (500): %v", tmplErr)
	}
}
