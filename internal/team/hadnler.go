package team

import (
	"encoding/json"
	"fmt"
	"net/http"

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
