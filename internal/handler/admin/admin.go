package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"

	"blog/internal/store"
)

const adminPerPage = 10

type AdminHandler struct {
	st       *store.Store
	validate *validator.Validate
}

func NewAdminHandler(st *store.Store) *AdminHandler {
	v := validator.New()
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "" {
			return fld.Name
		}
		return name
	})
	return &AdminHandler{st: st, validate: v}
}

type cursorPage[T any] struct {
	Data     []T    `json:"data"`
	HasNext  bool   `json:"has_next"`
	NextSlug string `json:"next_slug,omitempty"`
}

func respondJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

type validationErrors map[string][]string

func writeValidationErrors(w http.ResponseWriter, ve validationErrors) {
	respondJSON(w, http.StatusBadRequest, map[string]validationErrors{"errors": ve})
}

func translateValidationErrors(verr validator.ValidationErrors) validationErrors {
	ve := validationErrors{}
	for _, fe := range verr {
		field := fe.Field()
		tag := fe.Tag()
		param := fe.Param()

		switch tag {
		case "required":
			ve[field] = append(ve[field], field+" is required")
		case "datetime":
			ve[field] = append(ve[field], fmt.Sprintf("%s must be in %s format", field, param))
		case "gt":
			ve[field] = append(ve[field], field+" is required")
		default:
			ve[field] = append(ve[field], fmt.Sprintf("%s failed on %s", field, tag))
		}
	}
	return ve
}

func (h *AdminHandler) validateStruct(s any) validationErrors {
	err := h.validate.Struct(s)
	if err == nil {
		return nil
	}
	if verr, ok := err.(validator.ValidationErrors); ok {
		return translateValidationErrors(verr)
	}
	return nil
}
