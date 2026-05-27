package document

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/is0727kfJ/doc-share-saas/internal/database"
	"github.com/is0727kfJ/doc-share-saas/internal/middleware"
)

// Handler 構造体
type Handler struct {
	queries *database.Queries
}

func NewHandler(q *database.Queries) *Handler {
	return &Handler{queries: q}
}

type CreateDocumentRequest struct {
	TeamID  string `json:"team_id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type UpdateDocumentRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (h *Handler) CreateDocument(w http.ResponseWriter, r *http.Request) {
	var req CreateDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "無効なデータ形式です", http.StatusBadRequest)
		return
	}

	// リクエストの裏側（Context）に隠されている、ミドルウェアがセットした user_id を取り出す
	userIDValue := r.Context().Value(middleware.UserIDKey)
	if userIDValue == nil {
		http.Error(w, "認証情報が見つかりません", http.StatusUnauthorized)
		return
	}

	userIDStr, ok := userIDValue.(string)
	if !ok {
		http.Error(w, "ユーザーIDの形式が不正です", http.StatusInternalServerError)
		return
	}

	authorID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "無効なユーザーIDです", http.StatusInternalServerError)
		return
	}

	teamID, err := uuid.Parse(req.TeamID)
	if err != nil {
		http.Error(w, "無効なチームIDです", http.StatusBadRequest)
		return
	}

	// データベースに登録（AuthorID は、Contextから抜き出した確実な本人データを使う！）
	doc, err := h.queries.CreateDocument(r.Context(), database.CreateDocumentParams{
		Title:    req.Title,
		Content:  req.Content,
		AuthorID: authorID, //  偽造不可能なユーザーID
		TeamID:   teamID,
	})
	if err != nil {
		http.Error(w, "ドキュメントの作成に失敗しました", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(doc)
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

	limit := int32(10)
	offset := int32(0)

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.ParseInt(l, 10, 32); err == nil && parsed > 0 {
			limit = int32(parsed)
		}
	}

	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.ParseInt(o, 10, 32); err == nil && parsed >= 0 {
			offset = int32(parsed)
		}
	}

	docs, err := h.queries.ListDocumentsByTeam(r.Context(), database.ListDocumentsByTeamParams{
		TeamID: teamID,
		Limit:  limit,
		Offset: offset,
	})

	if err != nil {
		http.Error(w, "ドキュメントの取得に失敗しました", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(docs)
}

func (h *Handler) UpdateDocument(w http.ResponseWriter, r *http.Request) {
	// 1. URLから更新したいドキュメントのIDを取得
	docIDStr := chi.URLParam(r, "document_id")
	docID, err := uuid.Parse(docIDStr)
	if err != nil {
		http.Error(w, "無効なドキュメントIDです", http.StatusBadRequest)
		return
	}

	// 2. Bodyから新しいタイトルとコンテンツを受け取る
	var req UpdateDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "無効なデータ形式です", http.StatusBadRequest)
		return
	}

	// 3. 「操作している本人」のIDを取得
	userIDValue := r.Context().Value(middleware.UserIDKey)
	userIDStr, _ := userIDValue.(string)
	authorID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "無効なユーザーIDです", http.StatusInternalServerError)
		return
	}

	// 4. データベースの更新を実行
	doc, err := h.queries.UpdateDocument(r.Context(), database.UpdateDocumentParams{
		ID:       docID,
		Title:    req.Title,
		Content:  req.Content,
		AuthorID: authorID, // 追加：他人が上書きしようとしても弾かれる
	})
	if err != nil {
		http.Error(w, "ドキュメントの更新に失敗したか、権限がありません", http.StatusInternalServerError)
		return
	}

	// 成功時は更新された最新のドキュメント情報を返す
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(doc)
}
func (h *Handler) DeleteDocument(w http.ResponseWriter, r *http.Request) {
	// 1. URLから削除したいドキュメントのIDを取得
	docIDStr := chi.URLParam(r, "document_id")
	docID, err := uuid.Parse(docIDStr)
	if err != nil {
		http.Error(w, "無効なドキュメントIDです", http.StatusBadRequest)
		return
	}

	// 2. JWT（Context）から「操作している本人」のIDを取得
	userIDValue := r.Context().Value(middleware.UserIDKey)
	userIDStr, _ := userIDValue.(string)
	authorID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "無効なユーザーIDです", http.StatusInternalServerError)
		return
	}

	// 3. データベースの削除を実行
	err = h.queries.DeleteDocument(r.Context(), database.DeleteDocumentParams{
		ID:       docID,
		AuthorID: authorID, // ← 追加：他人のものはここで弾かれる
	})
	if err != nil {
		http.Error(w, "ドキュメントの削除に失敗したか、権限がありません", http.StatusInternalServerError)
		return
	}

	// 成功時は 204 No Content を返す
	w.WriteHeader(http.StatusNoContent)
}
