package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"mesascore/internal/auth"
	"mesascore/internal/email"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthHandler struct {
	pool      *pgxpool.Pool
	jwtSecret string
	mailer    *email.Sender
}

func NewAuthHandler(pool *pgxpool.Pool, jwtSecret string, mailer *email.Sender) *AuthHandler {
	return &AuthHandler{pool: pool, jwtSecret: jwtSecret, mailer: mailer}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username    string `json:"username"`
		DisplayName string `json:"display_name"`
		Email       string `json:"email"`
		Password    string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	// Validate fields
	fields := map[string]string{}
	if req.Username == "" {
		fields["username"] = "required"
	}
	if req.DisplayName == "" {
		fields["display_name"] = "required"
	}
	if req.Email == "" {
		fields["email"] = "required"
	}
	if len(req.Password) < 8 {
		fields["password"] = "must be at least 8 characters"
	}
	if len(fields) > 0 {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{"error": "validation failed", "fields": fields})
		return
	}

	// Check uniqueness
	var exists bool
	h.pool.QueryRow(r.Context(), "SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)", req.Username).Scan(&exists)
	if exists {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "username already taken"})
		return
	}
	h.pool.QueryRow(r.Context(), "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", req.Email).Scan(&exists)
	if exists {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "email already registered"})
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	token := auth.GenerateVerificationToken()
	now := time.Now()

	_, err = h.pool.Exec(r.Context(),
		`INSERT INTO users (username, display_name, email, password_hash, email_verified, verification_token, verification_sent_at)
		 VALUES ($1, $2, $3, $4, FALSE, $5, $6)`,
		req.Username, req.DisplayName, req.Email, hash, token, now)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create user"})
		return
	}

	// Send verification email (non-blocking failure — user is created regardless)
	go h.mailer.SendVerificationEmail(req.Email, token)

	writeJSON(w, http.StatusCreated, map[string]string{"message": "Check your email to verify your account"})
}

func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "token required"})
		return
	}

	var userID string
	var sentAt *time.Time
	err := h.pool.QueryRow(r.Context(),
		"SELECT id, verification_sent_at FROM users WHERE verification_token = $1 AND email_verified = FALSE",
		token).Scan(&userID, &sentAt)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid or expired token"})
		return
	}

	if sentAt == nil || time.Since(*sentAt) > 24*time.Hour {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "token expired"})
		return
	}

	_, err = h.pool.Exec(r.Context(),
		"UPDATE users SET email_verified = TRUE, verification_token = NULL, verification_sent_at = NULL WHERE id = $1",
		userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Email verified successfully"})
}

func (h *AuthHandler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusOK, map[string]string{"message": "If that email is registered, a verification email has been sent"})
		return
	}

	var userID string
	var verified bool
	err := h.pool.QueryRow(r.Context(),
		"SELECT id, email_verified FROM users WHERE email = $1", req.Email).Scan(&userID, &verified)
	if err != nil || verified {
		// Don't reveal whether email exists
		writeJSON(w, http.StatusOK, map[string]string{"message": "If that email is registered, a verification email has been sent"})
		return
	}

	token := auth.GenerateVerificationToken()
	now := time.Now()

	h.pool.Exec(r.Context(),
		"UPDATE users SET verification_token = $1, verification_sent_at = $2 WHERE id = $3",
		token, now, userID)

	go h.mailer.SendVerificationEmail(req.Email, token)

	writeJSON(w, http.StatusOK, map[string]string{"message": "If that email is registered, a verification email has been sent"})
}

func (h *AuthHandler) CheckUsername(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		writeJSON(w, http.StatusOK, map[string]bool{"available": false})
		return
	}

	var exists bool
	h.pool.QueryRow(r.Context(), "SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)", username).Scan(&exists)
	writeJSON(w, http.StatusOK, map[string]bool{"available": !exists})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	var userID, passwordHash, username, displayName string
	var emailVerified bool
	var avatarURL *string

	err := h.pool.QueryRow(r.Context(),
		"SELECT id, password_hash, email_verified, username, display_name, avatar_url FROM users WHERE email = $1",
		req.Email).Scan(&userID, &passwordHash, &emailVerified, &username, &displayName, &avatarURL)
	if err != nil {
		if err == pgx.ErrNoRows {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	if !emailVerified {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "email_not_verified"})
		return
	}

	if !auth.CheckPassword(passwordHash, req.Password) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}

	token, err := auth.IssueJWT(userID, h.jwtSecret, 7*24*time.Hour)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"token": token,
		"user": map[string]interface{}{
			"id":           userID,
			"username":     username,
			"display_name": displayName,
			"avatar_url":   avatarURL,
		},
	})
}
