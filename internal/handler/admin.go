package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"blog/internal/model"
	"blog/internal/store"
)

const adminPerPage = 10

type AdminHandler struct {
	st *store.Store
}

func NewAdminHandler(st *store.Store) *AdminHandler {
	return &AdminHandler{st: st}
}

type cursorPage[T any] struct {
	Data     []T    `json:"data"`
	HasNext  bool   `json:"has_next"`
	NextSlug string `json:"next_slug,omitempty"`
}

type categoryPayload struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type categoryResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func catToResp(c model.Category) categoryResponse {
	return categoryResponse{ID: c.ID, Name: c.Name, Slug: c.Slug}
}

func (h *AdminHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	after := r.URL.Query().Get("after")

	var (
		categories []model.Category
		err        error
	)

	if after != "" {
		categories, err = h.st.GetCategoriesAfter(r.Context(), after, adminPerPage+1)
	} else {
		categories, err = h.st.GetLatestCategories(r.Context(), adminPerPage+1)
	}
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch categories"})
		return
	}

	hasNext := len(categories) > adminPerPage
	if hasNext {
		categories = categories[:adminPerPage]
	}

	var nextSlug string
	if hasNext && len(categories) > 0 {
		nextSlug = categories[len(categories)-1].Slug
	}

	result := make([]categoryResponse, len(categories))
	for i, c := range categories {
		result[i] = catToResp(c)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cursorPage[categoryResponse]{Data: result, HasNext: hasNext, NextSlug: nextSlug})
}

func (h *AdminHandler) GetCategory(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	c, err := h.st.GetCategoryByID(r.Context(), id)
	if err != nil {
		respondJSON(w, http.StatusNotFound, map[string]string{"error": "category not found"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(catToResp(*c))
}

func (h *AdminHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var p categoryPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if p.Name == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "name is required"})
		return
	}
	if p.Slug == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "slug is required"})
		return
	}

	existing, _ := h.st.GetCategoryBySlug(r.Context(), p.Slug)
	if existing != nil {
		respondJSON(w, http.StatusConflict, map[string]string{"error": "slug already exists"})
		return
	}

	c := &model.Category{Name: p.Name, Slug: p.Slug}
	if err := h.st.CreateCategory(r.Context(), c); err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create category"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(catToResp(*c))
}

func (h *AdminHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	existing, err := h.st.GetCategoryByID(r.Context(), id)
	if err != nil {
		respondJSON(w, http.StatusNotFound, map[string]string{"error": "category not found"})
		return
	}

	var p categoryPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if p.Name == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "name is required"})
		return
	}
	if p.Slug == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "slug is required"})
		return
	}

	dup, _ := h.st.GetCategoryBySlug(r.Context(), p.Slug)
	if dup != nil && dup.ID != id {
		respondJSON(w, http.StatusConflict, map[string]string{"error": "slug already exists"})
		return
	}

	existing.Name = p.Name
	existing.Slug = p.Slug
	if err := h.st.UpdateCategory(r.Context(), existing); err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update category"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(catToResp(*existing))
}

func (h *AdminHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	_, err = h.st.GetCategoryByID(r.Context(), id)
	if err != nil {
		respondJSON(w, http.StatusNotFound, map[string]string{"error": "category not found"})
		return
	}

	if err := h.st.DeleteCategory(r.Context(), id); err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete category"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type tagPayload struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type tagResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func tagToResp(t model.Tag) tagResponse {
	return tagResponse{ID: t.ID, Name: t.Name, Slug: t.Slug}
}

func (h *AdminHandler) ListTags(w http.ResponseWriter, r *http.Request) {
	after := r.URL.Query().Get("after")

	var (
		tags []model.Tag
		err  error
	)

	if after != "" {
		tags, err = h.st.GetTagsAfter(r.Context(), after, adminPerPage+1)
	} else {
		tags, err = h.st.GetLatestTags(r.Context(), adminPerPage+1)
	}
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch tags"})
		return
	}

	hasNext := len(tags) > adminPerPage
	if hasNext {
		tags = tags[:adminPerPage]
	}

	var nextSlug string
	if hasNext && len(tags) > 0 {
		nextSlug = tags[len(tags)-1].Slug
	}

	result := make([]tagResponse, len(tags))
	for i, t := range tags {
		result[i] = tagToResp(t)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cursorPage[tagResponse]{Data: result, HasNext: hasNext, NextSlug: nextSlug})
}

func (h *AdminHandler) GetTag(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	t, err := h.st.GetTagByID(r.Context(), id)
	if err != nil {
		respondJSON(w, http.StatusNotFound, map[string]string{"error": "tag not found"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tagToResp(*t))
}

func (h *AdminHandler) CreateTag(w http.ResponseWriter, r *http.Request) {
	var p tagPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if p.Name == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "name is required"})
		return
	}
	if p.Slug == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "slug is required"})
		return
	}

	existing, _ := h.st.GetTagBySlug(r.Context(), p.Slug)
	if existing != nil {
		respondJSON(w, http.StatusConflict, map[string]string{"error": "slug already exists"})
		return
	}

	t := &model.Tag{Name: p.Name, Slug: p.Slug}
	if err := h.st.CreateTag(r.Context(), t); err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create tag"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tagToResp(*t))
}

func (h *AdminHandler) UpdateTag(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	existing, err := h.st.GetTagByID(r.Context(), id)
	if err != nil {
		respondJSON(w, http.StatusNotFound, map[string]string{"error": "tag not found"})
		return
	}

	var p tagPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if p.Name == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "name is required"})
		return
	}
	if p.Slug == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "slug is required"})
		return
	}

	dup, _ := h.st.GetTagBySlug(r.Context(), p.Slug)
	if dup != nil && dup.ID != id {
		respondJSON(w, http.StatusConflict, map[string]string{"error": "slug already exists"})
		return
	}

	existing.Name = p.Name
	existing.Slug = p.Slug
	if err := h.st.UpdateTag(r.Context(), existing); err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update tag"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tagToResp(*existing))
}

func (h *AdminHandler) DeleteTag(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	_, err = h.st.GetTagByID(r.Context(), id)
	if err != nil {
		respondJSON(w, http.StatusNotFound, map[string]string{"error": "tag not found"})
		return
	}

	if err := h.st.DeleteTag(r.Context(), id); err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete tag"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type postPayload struct {
	Title      string `json:"title"`
	Slug       string `json:"slug"`
	Markdown   string `json:"markdown"`
	Date       string `json:"date"`
	CategoryID int64  `json:"category_id"`
	TagIDs     []int64 `json:"tag_ids"`
}

type postResponse struct {
	ID         int64             `json:"id"`
	Slug       string            `json:"slug"`
	Title      string            `json:"title"`
	Date       string            `json:"date"`
	Markdown   string            `json:"markdown"`
	Category   *categoryResponse `json:"category,omitempty"`
	Tags       []tagResponse     `json:"tags"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
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

	if p.Title == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "title is required"})
		return
	}
	if p.Slug == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "slug is required"})
		return
	}
	if p.Markdown == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "markdown is required"})
		return
	}
	if p.Date == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "date is required"})
		return
	}
	if _, err := time.Parse("2006-01-02", p.Date); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "date must be in YYYY-MM-DD format"})
		return
	}
	if p.CategoryID <= 0 {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "category_id is required"})
		return
	}

	existing, _ := h.st.GetPostBySlug(r.Context(), p.Slug)
	if existing != nil {
		respondJSON(w, http.StatusConflict, map[string]string{"error": "slug already exists"})
		return
	}

	cat, err := h.st.GetCategoryByID(r.Context(), p.CategoryID)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "category not found"})
		return
	}

	var tags []model.Tag
	for _, tid := range p.TagIDs {
		t, err := h.st.GetTagByID(r.Context(), tid)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, map[string]string{"error": "tag not found: id " + strconv.FormatInt(tid, 10)})
			return
		}
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

	if p.Title == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "title is required"})
		return
	}
	if p.Slug == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "slug is required"})
		return
	}
	if p.Markdown == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "markdown is required"})
		return
	}
	if p.Date == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "date is required"})
		return
	}
	if _, err := time.Parse("2006-01-02", p.Date); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "date must be in YYYY-MM-DD format"})
		return
	}
	if p.CategoryID <= 0 {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "category_id is required"})
		return
	}

	dup, _ := h.st.GetPostBySlug(r.Context(), p.Slug)
	if dup != nil && dup.ID != id {
		respondJSON(w, http.StatusConflict, map[string]string{"error": "slug already exists"})
		return
	}

	cat, err := h.st.GetCategoryByID(r.Context(), p.CategoryID)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "category not found"})
		return
	}

	var tags []model.Tag
	for _, tid := range p.TagIDs {
		t, err := h.st.GetTagByID(r.Context(), tid)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, map[string]string{"error": "tag not found: id " + strconv.FormatInt(tid, 10)})
			return
		}
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
