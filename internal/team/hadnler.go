package team

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/is0727kfJ/doc-share-saas/internal/database"
)

type Handler struct {
	queries *database.Queries
}

func NewHandler(q *database.Queries) *Handler {
	return &Handler{queries: q}
}

type CreateTeamRequest struct {
	Name string `json:"name"`
}

type AddMemberRequest struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"` // "admin" や "member" など
}

// CreateTeam : 新しいチームを作成する
func (h *Handler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	var req CreateTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "無効なリクエストボディ", http.StatusBadRequest)
		return
	}

	newTeam, err := h.queries.CreateTeam(r.Context(), req.Name)
	if err != nil {
		http.Error(w, fmt.Sprintf("チーム作成エラー: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newTeam)
}

// AddMember : チームにメンバーを追加する
func (h *Handler) AddMember(w http.ResponseWriter, r *http.Request) {
	teamIDStr := chi.URLParam(r, "team_id") // URLからteamIDを取得
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		http.Error(w, "無効なチームID", http.StatusBadRequest)
		return
	}

	var req AddMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "無効なリクエストボディ", http.StatusBadRequest)
		return
	}

	userID, err := uuid.Parse(req.UserID) // userIDをUUIDに変換
	if err != nil {
		http.Error(w, "無効なユーザーIDです", http.StatusBadRequest)
		return
	}

	teamMember, err := h.queries.AddTeamMember(r.Context(), database.AddTeamMemberParams{
		UserID: userID,
		TeamID: teamID,
		Role:   req.Role,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("メンバー追加エラー: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(teamMember)
}

// ListTeamMembers : チームのメンバー一覧を取得する
func (h *Handler) ListTeamMembers(w http.ResponseWriter, r *http.Request) {
	// URLからチームIDを取得
	teamIDStr := chi.URLParam(r, "team_id")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		http.Error(w, "無効なチームIDです", http.StatusBadRequest)
		return
	}

	// データベースからメンバー一覧を取得
	members, err := h.queries.ListTeamMembers(r.Context(), teamID)
	if err != nil {
		http.Error(w, "メンバー一覧の取得に失敗しました", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(members)
}
