package admin

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"blog/internal/model"
)

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
		cats []model.Category
		err  error
	)

	if after != "" {
		cats, err = h.st.GetCategoriesAfter(r.Context(), after, adminPerPage+1)
	} else {
		cats, err = h.st.GetLatestCategories(r.Context(), adminPerPage+1)
	}
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch categories"})
		return
	}

	hasNext := len(cats) > adminPerPage
	if hasNext {
		cats = cats[:adminPerPage]
	}

	var nextSlug string
	if hasNext && len(cats) > 0 {
		nextSlug = cats[len(cats)-1].Slug
	}

	result := make([]categoryResponse, len(cats))
	for i, c := range cats {
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

	ve := validationErrors{}

	if p.Name == "" {
		ve["name"] = append(ve["name"], "name is required")
	}
	if p.Slug == "" {
		ve["slug"] = append(ve["slug"], "slug is required")
	}

	if p.Slug != "" {
		existing, _ := h.st.GetCategoryBySlug(r.Context(), p.Slug)
		if existing != nil {
			ve["slug"] = append(ve["slug"], "slug already exists")
		}
	}

	if len(ve) > 0 {
		writeValidationErrors(w, ve)
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

	ve := validationErrors{}

	if p.Name == "" {
		ve["name"] = append(ve["name"], "name is required")
	}
	if p.Slug == "" {
		ve["slug"] = append(ve["slug"], "slug is required")
	}

	if p.Slug != "" {
		dup, _ := h.st.GetCategoryBySlug(r.Context(), p.Slug)
		if dup != nil && dup.ID != id {
			ve["slug"] = append(ve["slug"], "slug already exists")
		}
	}

	if len(ve) > 0 {
		writeValidationErrors(w, ve)
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
