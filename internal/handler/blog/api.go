package blog

import (
	"encoding/json"
	"net/http"

	"blog/internal/service"
)

type searchResult struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
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
		h.serverError(w, r, err)
		return
	}
	writeJSON(w, paged)
}

func (h *SSRHandler) APISearchCategories(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	categories, err := h.svc.SearchCategories(r.Context(), q, 10)
	if err != nil {
		h.serverError(w, r, err)
		return
	}
	results := make([]searchResult, len(categories))
	for i, c := range categories {
		results[i] = searchResult{Name: c.Name, Slug: c.Slug}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func (h *SSRHandler) APISearchTags(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	tags, err := h.svc.SearchTags(r.Context(), q, 10)
	if err != nil {
		h.serverError(w, r, err)
		return
	}
	results := make([]searchResult, len(tags))
	for i, t := range tags {
		results[i] = searchResult{Name: t.Name, Slug: t.Slug}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
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
