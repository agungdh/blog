package admin

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"blog/internal/model"
)

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

	ve := validationErrors{}

	if p.Name == "" {
		ve["name"] = append(ve["name"], "name is required")
	}
	if p.Slug == "" {
		ve["slug"] = append(ve["slug"], "slug is required")
	}

	if p.Slug != "" {
		existing, _ := h.st.GetTagBySlug(r.Context(), p.Slug)
		if existing != nil {
			ve["slug"] = append(ve["slug"], "slug already exists")
		}
	}

	if len(ve) > 0 {
		writeValidationErrors(w, ve)
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

	ve := validationErrors{}

	if p.Name == "" {
		ve["name"] = append(ve["name"], "name is required")
	}
	if p.Slug == "" {
		ve["slug"] = append(ve["slug"], "slug is required")
	}

	if p.Slug != "" {
		dup, _ := h.st.GetTagBySlug(r.Context(), p.Slug)
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
