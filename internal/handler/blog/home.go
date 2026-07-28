package blog

import (
	"log"
	"net/http"
	"time"

	"blog/internal/model"
	"blog/internal/service"
)

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
		h.serverError(w, r, err)
		return
	}

	var selectedCategory *model.Category
	if filter.Category != "" {
		cats, _ := h.svc.GetCategoriesBySlugs(r.Context(), []string{filter.Category})
		if len(cats) > 0 {
			selectedCategory = &cats[0]
		}
	}

	var selectedTags []model.Tag
	if len(filter.Tags) > 0 {
		selectedTags, _ = h.svc.GetTagsBySlugs(r.Context(), filter.Tags)
	}

	data := map[string]any{
		"Title":            "My Blog",
		"Posts":            paged.Posts,
		"Pagination":       paged,
		"ApiPath":          "/api/posts",
		"Filter":           filter,
		"FilterParams":     filterQueryParams(filter),
		"SelectedCategory": selectedCategory,
		"SelectedTags":     selectedTags,
		"Year":             time.Now().Year(),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.ExecuteTemplate(w, "home.html", data); err != nil {
		log.Printf("template error (home): %v", err)
	}
}
