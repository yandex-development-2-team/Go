package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/yandex-development-2-team/Go/internal/auth"
	"github.com/yandex-development-2-team/Go/internal/database/repository"
	"github.com/yandex-development-2-team/Go/internal/models"
	"go.uber.org/zap"
)

type AuthConfig struct {
	JWTSecret             string
	AccessTokenTTLMinutes int
	RefreshTokenTTLHours  int
}

type RegisterRequest struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	Role        string `json:"role"` // admin/manager
	InviteToken string `json:"invite_token"`
}

type RegisterResponse struct {
	User         *models.AuthUser `json:"user"`
	Token        string           `json:"token"`
	RefreshToken string           `json:"refresh_token"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}
type LogoutResponse struct {
	Message string `json:"message"`
}

type ErrorEnvelope struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type AuthHandlers struct {
	db            *sqlx.DB
	users         *repository.AuthUserRepository
	refreshTokens *repository.RefreshTokenRepository
	cfg           AuthConfig
	logger        *zap.Logger
}

func NewAuthHandlers(db *sqlx.DB, cfg AuthConfig, logger *zap.Logger) *AuthHandlers {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &AuthHandlers{
		db:            db,
		users:         repository.NewAuthUserRepository(db),
		refreshTokens: repository.NewRefreshTokenRepository(db),
		cfg:           cfg,
		logger:        logger,
	}
}

func (h *AuthHandlers) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Name = strings.TrimSpace(req.Name)
	req.Role = strings.TrimSpace(strings.ToLower(req.Role))

	if req.Email == "" || !strings.Contains(req.Email, "@") {
		http.Error(w, "invalid email", http.StatusBadRequest)
		return
	}
	if len(req.Password) < 8 {
		http.Error(w, "weak password", http.StatusBadRequest)
		return
	}
	if req.Role != "admin" && req.Role != "manager" {
		http.Error(w, "invalid role", http.StatusBadRequest)
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		h.logger.Error("hash_password_failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	u, err := h.users.Create(r.Context(), req.Name, req.Email, hash, req.Role)
	if err != nil {
		if errors.Is(err, repository.ErrEmailExists) {
			var resp ErrorEnvelope
			resp.Error.Code = "EMAIL_ALREADY_EXISTS"
			resp.Error.Message = "email already exists"
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		h.logger.Error("create_user_failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	accessTTL := time.Duration(h.cfg.AccessTokenTTLMinutes) * time.Minute
	token, err := auth.GenerateAccessToken([]byte(h.cfg.JWTSecret), u.ID, u.Role, accessTTL)
	if err != nil {
		h.logger.Error("generate_access_failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	refresh, err := auth.GenerateRefreshToken()
	if err != nil {
		h.logger.Error("generate_refresh_failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	refreshExp := time.Now().Add(time.Duration(h.cfg.RefreshTokenTTLHours) * time.Hour)
	if err := h.refreshTokens.Insert(r.Context(), refresh, u.ID, refreshExp); err != nil {
		h.logger.Error("insert_refresh_failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	resp := RegisterResponse{
		User:         u,
		Token:        token,
		RefreshToken: refresh,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *AuthHandlers) Refresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	req.RefreshToken = strings.TrimSpace(req.RefreshToken)
	if req.RefreshToken == "" {
		http.Error(w, "refresh_token required", http.StatusBadRequest)
		return
	}

	newRefresh, err := auth.GenerateRefreshToken()
	if err != nil {
		h.logger.Error("generate_refresh_failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	newExp := time.Now().Add(time.Duration(h.cfg.RefreshTokenTTLHours) * time.Hour)

	userID, err := h.refreshTokens.Rotate(r.Context(), req.RefreshToken, newRefresh, newExp)
	if err != nil {
		h.logger.Error("refresh_failed", zap.Error(err))
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var role string
	err = h.db.GetContext(r.Context(), &role, "SELECT role FROM auth_users WHERE id = $1", userID)
	if err != nil {
		h.logger.Error("get_role_failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	accessTTL := time.Duration(h.cfg.AccessTokenTTLMinutes) * time.Minute
	token, err := auth.GenerateAccessToken([]byte(h.cfg.JWTSecret), userID, role, accessTTL)
	if err != nil {
		h.logger.Error("generate_access_failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	resp := RefreshResponse{
		Token:        token,
		RefreshToken: newRefresh,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *AuthHandlers) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	req.RefreshToken = strings.TrimSpace(req.RefreshToken)
	if req.RefreshToken == "" {
		http.Error(w, "refresh_token required", http.StatusBadRequest)
		return
	}

	err := h.refreshTokens.Revoke(r.Context(), req.RefreshToken)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	resp := LogoutResponse{Message: "Logged out successfully"}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
