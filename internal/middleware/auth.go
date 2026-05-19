// internal/middleware/auth.go
package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// JWTの署名検証に使う秘密鍵
var secretKey = []byte("super-secret-key")

// ContextKey : Goのコンテキストにデータを詰め込むための専用の型
type ContextKey string

const UserIDKey ContextKey = "user_id"

// Auth : JWTを検証して不正なアクセスを弾くミドルウェア
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. ヘッダーから通行証（Bearer トークン）を取得
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "認証が必要です", http.StatusUnauthorized)
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// 2. JWTの署名を検証
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// アルゴリズムがHMACであることを確認
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("予期しない署名アルゴリズム")
			}
			return secretKey, nil
		})

		// 3. トークンが不正、または期限切れの場合
		if err != nil || !token.Valid {
			http.Error(w, "無効なトークンです", http.StatusUnauthorized)
			return
		}

		// 4. トークンの中身（ペイロード）から「この通信は誰なのか」を取り出す
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "権限情報の取得に失敗しました", http.StatusUnauthorized)
			return
		}

		// 5. ユーザーIDをContext（リクエストの裏側）に忍ばせて、本来の処理へバトンタッチ
		ctx := context.WithValue(r.Context(), UserIDKey, claims["user_id"])
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
