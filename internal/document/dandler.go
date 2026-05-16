package document

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/google/uuid"
	"github.com/is0727kfJ/doc-share-saas/internal/database"
)

// Handler 構造体
type Handler struct {
	queries *database.Queries
}

func NewHandler(q *database.Queries) *Handler {
	return &Handler{queries: q}
}

type CreateDocumentRequest struct {
	TeamID   uuid.UUID `json:"team_id"`
	AuthorID uuid.UUID `json:"author_id"`
	Title    string    `json:"title"`
	Content  string    `json:"content"`
}

func (h *Handler) CreateDocument(w http.ResponseWriter, r *http.Request) {
	var req CreateDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "無効なリクエストボディ", http.StatusBadRequest)
		return
	}

	newDoc, err := h.queries.CreateDocument(r.Context(), database.CreateDocumentParams{
		TeamID:   req.TeamID,
		AuthorID: req.AuthorID,
		Title:    req.Title,
		Content:  req.Content,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("ドキュメント作成エラー: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newDoc)
}

// ListDocuments : ドキュメントの一覧を取得して返す処理
func (h *Handler) ListDocuments(w http.ResponseWriter, r *http.Request) {
	// データベースから新しい順にドキュメント一覧を取得
	docs, err := h.queries.ListDocuments(r.Context())
	if err != nil {
		http.Error(w, "ドキュメントの取得に失敗しました", http.StatusInternalServerError)
		return
	}

	// 取得した一覧データ(docs)をJSONにしてフロントエンドに返す
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(docs)
}

func (h *Handler) ListDocumentsByTeam(w http.ResponseWriter, r *http.Request) {
	teamIDStr := chi.URLParam(r, "team_id")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		http.Error(w, "無効なチームID", http.StatusBadRequest)
		return
	}

	docs, err := h.queries.ListDocumentsByTeam(r.Context(), teamID)
	if err != nil {
		http.Error(w, "ドキュメントの取得に失敗しました", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(docs)
}
