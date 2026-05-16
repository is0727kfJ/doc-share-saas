package user

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/is0727kfJ/doc-share-saas/internal/database"
)

type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Handler struct {
	queries *database.Queries
}

func NewHandler(q *database.Queries) *Handler {
	return &Handler{queries: q}
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "無効なリクエストボディ", http.StatusBadRequest)
		return
	}

	dummySub := "cognito-sub-" + uuid.New().String()[:8]
	newUser, err := h.queries.CreateUser(r.Context(), database.CreateUserParams{
		CognitoSub: dummySub,
		Name:       req.Name,
		Email:      req.Email,
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("ユーザー作成エラー: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newUser)
}
