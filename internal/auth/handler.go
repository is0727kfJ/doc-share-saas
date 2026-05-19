package auth

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var secretKey = []byte("super-secret-key")

// Handler : 認証関連のAPIをまとめる構造体
type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

// ログイン時のリクエストデータ
type LoginRequest struct {
	UserID string `json:"user_id"`
}

// ログイン成功時に返すデータ（トークン）
type LoginResponse struct {
	Token string `json:"token"`
}

// Login : ユーザーIDを受け取り、JWTを発行する
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "無効なデータ形式です", http.StatusBadRequest)
		return
	}

	claims := jwt.MapClaims{
		"user_id": req.UserID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}

	// 2. トークンオブジェクトの生成（アルゴリズムはHS256を指定）
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 3. 秘密鍵を使って「署名（ハンコ）」を押し、文字列（ey...）にする
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		http.Error(w, "トークンの生成に失敗しました", http.StatusInternalServerError)
		return
	}

	// 4. 完成したトークンをJSONで返す
	response := LoginResponse{
		Token: tokenString,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
