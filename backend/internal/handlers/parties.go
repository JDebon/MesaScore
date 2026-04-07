package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"mesascore/internal/auth"
	"mesascore/internal/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PartyHandler struct {
	pool *pgxpool.Pool
}

func NewPartyHandler(pool *pgxpool.Pool) *PartyHandler {
	return &PartyHandler{pool: pool}
}

func (h *PartyHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	var req struct {
		Name        string  `json:"name"`
		Description *string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error":  "validation failed",
			"fields": map[string]string{"name": "required"},
		})
		return
	}

	inviteCode := auth.GenerateInviteCode()
	var partyID string

	err := h.pool.QueryRow(r.Context(),
		`INSERT INTO parties (name, description, admin_user_id, invite_code) VALUES ($1, $2, $3, $4) RETURNING id`,
		req.Name, req.Description, userID, inviteCode).Scan(&partyID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create party"})
		return
	}

	// Add creator as first member
	_, err = h.pool.Exec(r.Context(),
		"INSERT INTO party_members (party_id, user_id) VALUES ($1, $2)", partyID, userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to add creator as member"})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"id": partyID, "invite_code": inviteCode})
}

func (h *PartyHandler) Get(w http.ResponseWriter, r *http.Request) {
	partyID := middleware.PartyIDFromContext(r.Context())

	var id, name, adminID, inviteCode string
	var description *string
	var createdAt time.Time
	var adminDisplayName string
	var adminAvatarURL *string

	err := h.pool.QueryRow(r.Context(),
		`SELECT p.id, p.name, p.description, p.admin_user_id, p.invite_code, p.created_at,
		        u.display_name, u.avatar_url
		 FROM parties p JOIN users u ON u.id = p.admin_user_id
		 WHERE p.id = $1`, partyID).
		Scan(&id, &name, &description, &adminID, &inviteCode, &createdAt, &adminDisplayName, &adminAvatarURL)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "party not found"})
		return
	}

	var memberCount int
	h.pool.QueryRow(r.Context(), "SELECT COUNT(*) FROM party_members WHERE party_id = $1", partyID).Scan(&memberCount)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":          id,
		"name":        name,
		"description": description,
		"admin": map[string]interface{}{
			"id":           adminID,
			"display_name": adminDisplayName,
			"avatar_url":   adminAvatarURL,
		},
		"invite_code":  inviteCode,
		"member_count": memberCount,
		"created_at":   createdAt,
	})
}

func (h *PartyHandler) Update(w http.ResponseWriter, r *http.Request) {
	partyID := middleware.PartyIDFromContext(r.Context())

	var req struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Name != nil {
		h.pool.Exec(r.Context(), "UPDATE parties SET name = $1 WHERE id = $2", *req.Name, partyID)
	}
	if req.Description != nil {
		h.pool.Exec(r.Context(), "UPDATE parties SET description = $1 WHERE id = $2", *req.Description, partyID)
	}

	// Return updated party - reuse Get logic
	h.getPartyDetail(w, r, partyID)
}

func (h *PartyHandler) getPartyDetail(w http.ResponseWriter, r *http.Request, partyID string) {
	var id, name, adminID, inviteCode string
	var description *string
	var createdAt time.Time
	var adminDisplayName string
	var adminAvatarURL *string

	err := h.pool.QueryRow(r.Context(),
		`SELECT p.id, p.name, p.description, p.admin_user_id, p.invite_code, p.created_at,
		        u.display_name, u.avatar_url
		 FROM parties p JOIN users u ON u.id = p.admin_user_id
		 WHERE p.id = $1`, partyID).
		Scan(&id, &name, &description, &adminID, &inviteCode, &createdAt, &adminDisplayName, &adminAvatarURL)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "party not found"})
		return
	}

	var memberCount int
	h.pool.QueryRow(r.Context(), "SELECT COUNT(*) FROM party_members WHERE party_id = $1", partyID).Scan(&memberCount)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":          id,
		"name":        name,
		"description": description,
		"admin": map[string]interface{}{
			"id":           adminID,
			"display_name": adminDisplayName,
			"avatar_url":   adminAvatarURL,
		},
		"invite_code":  inviteCode,
		"member_count": memberCount,
		"created_at":   createdAt,
	})
}

func (h *PartyHandler) JoinPreview(w http.ResponseWriter, r *http.Request) {
	inviteCode := chi.URLParam(r, "inviteCode")

	var partyID, name string
	err := h.pool.QueryRow(r.Context(),
		"SELECT id, name FROM parties WHERE invite_code = $1", inviteCode).
		Scan(&partyID, &name)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "invite link not found"})
		return
	}

	var memberCount int
	h.pool.QueryRow(r.Context(), "SELECT COUNT(*) FROM party_members WHERE party_id = $1", partyID).Scan(&memberCount)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"party": map[string]interface{}{
			"id":           partyID,
			"name":         name,
			"member_count": memberCount,
		},
	})
}

func (h *PartyHandler) Join(w http.ResponseWriter, r *http.Request) {
	inviteCode := chi.URLParam(r, "inviteCode")
	userID := middleware.UserIDFromContext(r.Context())

	var partyID string
	err := h.pool.QueryRow(r.Context(),
		"SELECT id FROM parties WHERE invite_code = $1", inviteCode).Scan(&partyID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "invite link not found"})
		return
	}

	// Check if already a member
	var exists bool
	h.pool.QueryRow(r.Context(),
		"SELECT EXISTS(SELECT 1 FROM party_members WHERE party_id = $1 AND user_id = $2)",
		partyID, userID).Scan(&exists)
	if exists {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "already a member"})
		return
	}

	_, err = h.pool.Exec(r.Context(),
		"INSERT INTO party_members (party_id, user_id) VALUES ($1, $2)", partyID, userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to join party"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"party_id": partyID})
}

func (h *PartyHandler) Members(w http.ResponseWriter, r *http.Request) {
	partyID := middleware.PartyIDFromContext(r.Context())
	userID := middleware.UserIDFromContext(r.Context())

	// Get admin ID
	var adminID string
	h.pool.QueryRow(r.Context(), "SELECT admin_user_id FROM parties WHERE id = $1", partyID).Scan(&adminID)

	// Members
	rows, err := h.pool.Query(r.Context(),
		`SELECT u.id, u.username, u.display_name, u.avatar_url, pm.joined_at
		 FROM party_members pm JOIN users u ON u.id = pm.user_id
		 WHERE pm.party_id = $1
		 ORDER BY pm.joined_at`, partyID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	defer rows.Close()

	members := []map[string]interface{}{}
	for rows.Next() {
		var id, username, displayName string
		var avatarURL *string
		var joinedAt time.Time
		rows.Scan(&id, &username, &displayName, &avatarURL, &joinedAt)
		members = append(members, map[string]interface{}{
			"id":           id,
			"username":     username,
			"display_name": displayName,
			"avatar_url":   avatarURL,
			"is_admin":     id == adminID,
			"joined_at":    joinedAt,
		})
	}

	// Invites (only for admin)
	invites := []map[string]interface{}{}
	if userID == adminID {
		invRows, err := h.pool.Query(r.Context(),
			`SELECT pi.id, u.id, u.username, u.display_name, pi.status, pi.created_at
			 FROM party_invites pi JOIN users u ON u.id = pi.invited_user_id
			 WHERE pi.party_id = $1
			 ORDER BY pi.created_at DESC`, partyID)
		if err == nil {
			defer invRows.Close()
			for invRows.Next() {
				var invID, uID, uUsername, uDisplayName, status string
				var createdAt time.Time
				invRows.Scan(&invID, &uID, &uUsername, &uDisplayName, &status, &createdAt)
				invites = append(invites, map[string]interface{}{
					"id": invID,
					"invited_user": map[string]interface{}{
						"id":           uID,
						"username":     uUsername,
						"display_name": uDisplayName,
					},
					"status":     status,
					"created_at": createdAt,
				})
			}
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"members": members,
		"invites": invites,
	})
}

func (h *PartyHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	partyID := middleware.PartyIDFromContext(r.Context())
	targetUserID := chi.URLParam(r, "userId")

	// Cannot remove admin
	var adminID string
	h.pool.QueryRow(r.Context(), "SELECT admin_user_id FROM parties WHERE id = $1", partyID).Scan(&adminID)
	if targetUserID == adminID {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "cannot remove admin"})
		return
	}

	tag, err := h.pool.Exec(r.Context(),
		"DELETE FROM party_members WHERE party_id = $1 AND user_id = $2", partyID, targetUserID)
	if err != nil || tag.RowsAffected() == 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not a member"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *PartyHandler) Leave(w http.ResponseWriter, r *http.Request) {
	partyID := middleware.PartyIDFromContext(r.Context())
	userID := middleware.UserIDFromContext(r.Context())

	// Check not admin
	var adminID string
	h.pool.QueryRow(r.Context(), "SELECT admin_user_id FROM parties WHERE id = $1", partyID).Scan(&adminID)
	if userID == adminID {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "admin must transfer ownership before leaving"})
		return
	}

	h.pool.Exec(r.Context(),
		"DELETE FROM party_members WHERE party_id = $1 AND user_id = $2", partyID, userID)

	w.WriteHeader(http.StatusNoContent)
}

func (h *PartyHandler) TransferAdmin(w http.ResponseWriter, r *http.Request) {
	partyID := middleware.PartyIDFromContext(r.Context())

	var req struct {
		NewAdminUserID string `json:"new_admin_user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	// Verify target is a member
	var exists bool
	h.pool.QueryRow(r.Context(),
		"SELECT EXISTS(SELECT 1 FROM party_members WHERE party_id = $1 AND user_id = $2)",
		partyID, req.NewAdminUserID).Scan(&exists)
	if !exists {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "target user is not a party member"})
		return
	}

	_, err := h.pool.Exec(r.Context(),
		"UPDATE parties SET admin_user_id = $1 WHERE id = $2", req.NewAdminUserID, partyID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to transfer ownership"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Ownership transferred"})
}

func (h *PartyHandler) RegenerateInvite(w http.ResponseWriter, r *http.Request) {
	partyID := middleware.PartyIDFromContext(r.Context())
	newCode := auth.GenerateInviteCode()

	_, err := h.pool.Exec(r.Context(),
		"UPDATE parties SET invite_code = $1 WHERE id = $2", newCode, partyID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to regenerate invite"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"invite_code": newCode})
}

func (h *PartyHandler) SendInvite(w http.ResponseWriter, r *http.Request) {
	partyID := middleware.PartyIDFromContext(r.Context())
	inviterID := middleware.UserIDFromContext(r.Context())

	var req struct {
		UserID string `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	// Check user exists
	var exists bool
	h.pool.QueryRow(r.Context(), "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", req.UserID).Scan(&exists)
	if !exists {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}

	// Check not already a member
	h.pool.QueryRow(r.Context(),
		"SELECT EXISTS(SELECT 1 FROM party_members WHERE party_id = $1 AND user_id = $2)",
		partyID, req.UserID).Scan(&exists)
	if exists {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "user is already a member"})
		return
	}

	// Check no pending invite
	h.pool.QueryRow(r.Context(),
		"SELECT EXISTS(SELECT 1 FROM party_invites WHERE party_id = $1 AND invited_user_id = $2 AND status = 'pending')",
		partyID, req.UserID).Scan(&exists)
	if exists {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "pending invite already exists"})
		return
	}

	_, err := h.pool.Exec(r.Context(),
		`INSERT INTO party_invites (party_id, invited_user_id, invited_by_user_id, status)
		 VALUES ($1, $2, $3, 'pending')`,
		partyID, req.UserID, inviterID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to send invite"})
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *PartyHandler) AcceptInvite(w http.ResponseWriter, r *http.Request) {
	partyID := chi.URLParam(r, "partyId")
	if partyID == "" {
		partyID = middleware.PartyIDFromContext(r.Context())
	}
	inviteID := chi.URLParam(r, "inviteId")
	userID := middleware.UserIDFromContext(r.Context())

	// Verify invite belongs to this user
	var invitedUserID string
	err := h.pool.QueryRow(r.Context(),
		"SELECT invited_user_id FROM party_invites WHERE id = $1 AND party_id = $2 AND status = 'pending'",
		inviteID, partyID).Scan(&invitedUserID)
	if err != nil {
		if err == pgx.ErrNoRows {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "invite not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	if invitedUserID != userID {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "not your invite"})
		return
	}

	// Update invite status
	h.pool.Exec(r.Context(),
		"UPDATE party_invites SET status = 'accepted' WHERE id = $1", inviteID)

	// Add as member
	h.pool.Exec(r.Context(),
		"INSERT INTO party_members (party_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING",
		partyID, userID)

	writeJSON(w, http.StatusOK, map[string]string{"party_id": partyID})
}

func (h *PartyHandler) DeclineInvite(w http.ResponseWriter, r *http.Request) {
	inviteID := chi.URLParam(r, "inviteId")
	partyID := chi.URLParam(r, "partyId")
	if partyID == "" {
		partyID = middleware.PartyIDFromContext(r.Context())
	}
	userID := middleware.UserIDFromContext(r.Context())

	var invitedUserID string
	err := h.pool.QueryRow(r.Context(),
		"SELECT invited_user_id FROM party_invites WHERE id = $1 AND party_id = $2 AND status = 'pending'",
		inviteID, partyID).Scan(&invitedUserID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "invite not found"})
		return
	}

	if invitedUserID != userID {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "not your invite"})
		return
	}

	h.pool.Exec(r.Context(),
		"UPDATE party_invites SET status = 'declined' WHERE id = $1", inviteID)

	w.WriteHeader(http.StatusNoContent)
}

func (h *PartyHandler) PendingInvites(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	rows, err := h.pool.Query(r.Context(),
		`SELECT pi.id, p.id, p.name, u.id, u.display_name, pi.created_at
		 FROM party_invites pi
		 JOIN parties p ON p.id = pi.party_id
		 JOIN users u ON u.id = pi.invited_by_user_id
		 WHERE pi.invited_user_id = $1 AND pi.status = 'pending'
		 ORDER BY pi.created_at DESC`, userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	defer rows.Close()

	invites := []map[string]interface{}{}
	for rows.Next() {
		var invID, pID, pName, uID, uDisplayName string
		var createdAt time.Time
		rows.Scan(&invID, &pID, &pName, &uID, &uDisplayName, &createdAt)
		invites = append(invites, map[string]interface{}{
			"id":    invID,
			"party": map[string]interface{}{"id": pID, "name": pName},
			"invited_by": map[string]interface{}{
				"id":           uID,
				"display_name": uDisplayName,
			},
			"created_at": createdAt,
		})
	}

	writeJSON(w, http.StatusOK, invites)
}
