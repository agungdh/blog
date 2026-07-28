package admin

import (
	"encoding/json"
	"net/http"

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

func respondJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

type validationErrors map[string][]string

func writeValidationErrors(w http.ResponseWriter, ve validationErrors) {
	respondJSON(w, http.StatusBadRequest, map[string]validationErrors{"errors": ve})
}
