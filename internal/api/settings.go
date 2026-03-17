package api

import (
	"encoding/json"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/yandex-development-2-team/Go/internal/models"
	"go.uber.org/zap"
)

func NewSettingsHandler(db *sqlx.DB, logger *zap.Logger) http.HandlerFunc {
	if logger == nil {
		logger = zap.NewNop()
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
		var raw struct {
			Notifications []byte `json:"notifications"`
			Booking       []byte `json:"booking"`
			General       []byte `json:"general"`
		}

		if err := db.GetContext(r.Context(), &raw, `
SELECT notifications, booking, general
FROM settings
WHERE id = 1
`); err != nil {
			logger.Error("failed_to_get_settings", zap.Error(err))
			http.Error(w, "failed to get settings", http.StatusInternalServerError)
			return
		}

		var resp models.Settings
		if err := json.Unmarshal(raw.Notifications, &resp.Notifications); err != nil {
			http.Error(w, "invalid notifications json", http.StatusInternalServerError)
			return
		}
		if err := json.Unmarshal(raw.Booking, &resp.Booking); err != nil {
			http.Error(w, "invalid booking json", http.StatusInternalServerError)
			return
		}
		if err := json.Unmarshal(raw.General, &resp.General); err != nil {
			http.Error(w, "invalid general json", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}
}
