package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"go.uber.org/zap"

	"github.com/yandex-development-2-team/Go/internal/database/repository"
	"github.com/yandex-development-2-team/Go/internal/service/storage"
)

type BoxesHandler struct {
	boxRepo      *repository.BoxRepository
	imageStorage storage.ImageStorage
	logger       *zap.Logger
}

func NewBoxesHandler(boxRepo *repository.BoxRepository, imageStorage storage.ImageStorage, logger *zap.Logger) *BoxesHandler {
	return &BoxesHandler{
		boxRepo:      boxRepo,
		imageStorage: imageStorage,
		logger:       logger,
	}
}

func (h *BoxesHandler) GetBoxByID(w http.ResponseWriter, r *http.Request) {
	if err := r.Context().Err(); err != nil {
		w.WriteHeader(http.StatusRequestTimeout)
		h.logger.Error("context cancelled")
		return
	}
	id := r.URL.Query().Get("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		h.logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	box, err := h.boxRepo.SelectByID(r.Context(), idInt)
	if err != nil {
		if err == sql.ErrNoRows {
			h.logger.Error("box not found")
			w.WriteHeader(http.StatusNotFound)
			return
		}
		h.logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(box)

}

func (h *BoxesHandler) UpdateBox(w http.ResponseWriter, r *http.Request) {
	if err := r.Context().Err(); err != nil {
		h.logger.Error("context cancelled")
		return
	}
}

func (h *BoxesHandler) DeleteBox(w http.ResponseWriter, r *http.Request) {
	if err := r.Context().Err(); err != nil {
		h.logger.Error("context cancelled")
		return
	}
}

func (h *BoxesHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	if err := r.Context().Err(); err != nil {
		h.logger.Error("context cancelled")
		return
	}
}

func (h *BoxesHandler) UploadImage(w http.ResponseWriter, r *http.Request) {
	if err := r.Context().Err(); err != nil {
		h.logger.Error("context cancelled")
		return
	}
}

func (h *BoxesHandler) ExportBoxes(w http.ResponseWriter, r *http.Request) {
	if err := r.Context().Err(); err != nil {
		h.logger.Error("context cancelled")
		return
	}
}
