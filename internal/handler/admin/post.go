package admin

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"blog/internal/model"
)

type postPayload struct {
	Title      string  `json:"title"`
	Slug       string  `json:"slug"`
	Markdown   string  `json:"markdown"`
	Date       string  `json:"date"`
	CategoryID int64   `json:"category_id"`
	TagIDs     []int64 `json:"tag_ids"`
}

type postResponse struct {
	ID        int64             `json:"id"`
	Slug      string            `json:"slug"`
	Title     string            `json:"title"`
	Date      string            `json:"date"`
	Markdown  string            `json:"markdown"`
	Category  *categoryResponse `json:"category,omitempty"`
	Tags      []tagResponse     `json:"tags"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

func postToResp(p model.Post) postResponse {
	resp := postResponse{
		ID:        p.ID,
		Slug:      p.Slug,
		Title:     p.Title,
		Date:      p.Date,
		Markdown:  p.Markdown,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
	if p.Category != nil {
		c := catToResp(*p.Category)
		resp.Category = &c
	}
	tags := make([]tagResponse, len(p.Tags))
	for i, t := range p.Tags {
		tags[i] = tagToResp(t)
	}
	resp.Tags = tags
	return resp
}

func (h *AdminHandler) ListPosts(w http.ResponseWriter, r *http.Request) {
	after := r.URL.Query().Get("after")

	var (
		posts []model.Post
		err   error
	)

	if after != "" {
		posts, err = h.st.GetPostsAfter(r.Context(), after, adminPerPage+1)
	} else {
		posts, err = h.st.GetLatestPosts(r.Context(), adminPerPage+1)
	}
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch posts"})
		return
	}

	hasNext := len(posts) > adminPerPage
	if hasNext {
		posts = posts[:adminPerPage]
	}

	var nextSlug string
	if hasNext && len(posts) > 0 {
		nextSlug = posts[len(posts)-1].Slug
	}

	result := make([]postResponse, len(posts))
	for i, p := range posts {
		result[i] = postToResp(p)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cursorPage[postResponse]{Data: result, HasNext: hasNext, NextSlug: nextSlug})
}

func (h *AdminHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	p, err := h.st.GetPostByID(r.Context(), id)
	if err != nil {
		respondJSON(w, http.StatusNotFound, map[string]string{"error": "post not found"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(postToResp(*p))
}

func (h *AdminHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	var p postPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	ve := validationErrors{}

	if p.Title == "" {
		ve["title"] = append(ve["title"], "title is required")
	}
	if p.Slug == "" {
		ve["slug"] = append(ve["slug"], "slug is required")
	}
	if p.Markdown == "" {
		ve["markdown"] = append(ve["markdown"], "markdown is required")
	}
	if p.Date == "" {
		ve["date"] = append(ve["date"], "date is required")
	} else if _, err := time.Parse("2006-01-02", p.Date); err != nil {
		ve["date"] = append(ve["date"], "date must be in YYYY-MM-DD format")
	}
	if p.CategoryID <= 0 {
		ve["category_id"] = append(ve["category_id"], "category_id is required")
	}

	if p.Slug != "" {
		existing, _ := h.st.GetPostBySlug(r.Context(), p.Slug)
		if existing != nil {
			ve["slug"] = append(ve["slug"], "slug already exists")
		}
	}

	if p.CategoryID > 0 {
		_, err := h.st.GetCategoryByID(r.Context(), p.CategoryID)
		if err != nil {
			ve["category_id"] = append(ve["category_id"], "category not found")
		}
	}

	for _, tid := range p.TagIDs {
		_, err := h.st.GetTagByID(r.Context(), tid)
		if err != nil {
			ve["tag_ids"] = append(ve["tag_ids"], "tag not found: id "+strconv.FormatInt(tid, 10))
		}
	}

	if len(ve) > 0 {
		writeValidationErrors(w, ve)
		return
	}

	cat, _ := h.st.GetCategoryByID(r.Context(), p.CategoryID)

	var tags []model.Tag
	for _, tid := range p.TagIDs {
		t, _ := h.st.GetTagByID(r.Context(), tid)
		tags = append(tags, *t)
	}

	post := &model.Post{
		Slug:       p.Slug,
		Title:      p.Title,
		Date:       p.Date,
		Markdown:   p.Markdown,
		CategoryID: cat.ID,
		Tags:       tags,
	}
	if err := h.st.CreatePost(r.Context(), post); err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create post"})
		return
	}

	created, _ := h.st.GetPostByID(r.Context(), post.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if created != nil {
		json.NewEncoder(w).Encode(postToResp(*created))
	} else {
		json.NewEncoder(w).Encode(postToResp(*post))
	}
}

func (h *AdminHandler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	existing, err := h.st.GetPostByID(r.Context(), id)
	if err != nil {
		respondJSON(w, http.StatusNotFound, map[string]string{"error": "post not found"})
		return
	}

	var p postPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	ve := validationErrors{}

	if p.Title == "" {
		ve["title"] = append(ve["title"], "title is required")
	}
	if p.Slug == "" {
		ve["slug"] = append(ve["slug"], "slug is required")
	}
	if p.Markdown == "" {
		ve["markdown"] = append(ve["markdown"], "markdown is required")
	}
	if p.Date == "" {
		ve["date"] = append(ve["date"], "date is required")
	} else if _, err := time.Parse("2006-01-02", p.Date); err != nil {
		ve["date"] = append(ve["date"], "date must be in YYYY-MM-DD format")
	}
	if p.CategoryID <= 0 {
		ve["category_id"] = append(ve["category_id"], "category_id is required")
	}

	if p.Slug != "" {
		dup, _ := h.st.GetPostBySlug(r.Context(), p.Slug)
		if dup != nil && dup.ID != id {
			ve["slug"] = append(ve["slug"], "slug already exists")
		}
	}

	if p.CategoryID > 0 {
		_, err := h.st.GetCategoryByID(r.Context(), p.CategoryID)
		if err != nil {
			ve["category_id"] = append(ve["category_id"], "category not found")
		}
	}

	for _, tid := range p.TagIDs {
		_, err := h.st.GetTagByID(r.Context(), tid)
		if err != nil {
			ve["tag_ids"] = append(ve["tag_ids"], "tag not found: id "+strconv.FormatInt(tid, 10))
		}
	}

	if len(ve) > 0 {
		writeValidationErrors(w, ve)
		return
	}

	cat, _ := h.st.GetCategoryByID(r.Context(), p.CategoryID)

	var tags []model.Tag
	for _, tid := range p.TagIDs {
		t, _ := h.st.GetTagByID(r.Context(), tid)
		tags = append(tags, *t)
	}

	existing.Slug = p.Slug
	existing.Title = p.Title
	existing.Date = p.Date
	existing.Markdown = p.Markdown
	existing.CategoryID = cat.ID
	existing.Category = cat
	existing.Tags = tags

	if err := h.st.UpdatePostByID(r.Context(), existing); err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update post"})
		return
	}

	updated, _ := h.st.GetPostByID(r.Context(), id)

	w.Header().Set("Content-Type", "application/json")
	if updated != nil {
		json.NewEncoder(w).Encode(postToResp(*updated))
	} else {
		json.NewEncoder(w).Encode(postToResp(*existing))
	}
}

func (h *AdminHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	_, err = h.st.GetPostByID(r.Context(), id)
	if err != nil {
		respondJSON(w, http.StatusNotFound, map[string]string{"error": "post not found"})
		return
	}

	if err := h.st.DeletePostByID(r.Context(), id); err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete post"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
