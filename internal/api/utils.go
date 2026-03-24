package api

import (
	"encoding/json"
	"net/http"
	"strings"
)

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(data); err != nil {
		http.Error(w, `{"error":{"message":"internal server error"}}`, http.StatusInternalServerError)
	}
}

func cleanSlug(path, prefix string) string {
	trimmed := strings.TrimPrefix(path, prefix)
	trimmed = strings.Trim(trimmed, "/")

	if trimmed == "" {
		return ""
	}

	parts := strings.Split(trimmed, "/")
	return parts[0]
}

func respondError(w http.ResponseWriter, code int, message, field string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	resp := map[string]interface{}{
		"error": map[string]string{"message": message},
	}
	if field != "" {
		resp["error"].(map[string]string)["field"] = field
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func respondValidationError(w http.ResponseWriter, field, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	ve := struct {
		Error struct {
			Field   string `json:"field"`
			Message string `json:"message"`
		} `json:"error"`
	}{}
	ve.Error.Field = field
	ve.Error.Message = msg
	_ = json.NewEncoder(w).Encode(ve)
}
