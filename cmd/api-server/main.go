package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"golang.org/x/time/rate"

	"github.com/is0727kfJ/doc-share-saas/internal/auth"
	"github.com/is0727kfJ/doc-share-saas/internal/database"
	"github.com/is0727kfJ/doc-share-saas/internal/document"
	mymiddleware "github.com/is0727kfJ/doc-share-saas/internal/middleware"
	"github.com/is0727kfJ/doc-share-saas/internal/team"
	"github.com/is0727kfJ/doc-share-saas/internal/user"
)

func main() {
	// ==========================================
	// 1. 環境変数の読み込みとデータベース接続
	// ==========================================
	err := godotenv.Load()
	if err != nil {
		log.Println(".envファイルが見つかりません。OSの環境変数を使用します。")
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URLが設定されていません")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("データベースの接続に失敗:", err)
	}
	defer db.Close()

	// sqlcが自動生成したクエリの初期化
	queries := database.New(db)

	// ==========================================
	// 2. ハンドラー（機能）の初期化
	// ==========================================
	userHandler := user.NewHandler(queries)
	authHandler := auth.NewHandler()
	teamHandler := team.NewHandler(queries)
	docHandler := document.NewHandler(queries)

	// ==========================================
	// 3. ルーターとミドルウェア（門番）の設定
	// ==========================================
	r := chi.NewRouter()

	// 基本のログ出力
	r.Use(middleware.Logger)

	// CORS設定（Next.jsなどのフロントエンドからの通信を許可）
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"http://localhost:3000"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		MaxAge:         300,
	}))

	limiter := mymiddleware.NewIPRateLimiter(rate.Limit(1), 3)
	r.Use(limiter.RateLimitMiddleware)

	// ==========================================
	// 4. APIエンドポイント（URL）の定義
	// ==========================================

	// 【公開API】（防弾ガラスの外側：誰でもアクセス可能）
	r.Post("/api/users", userHandler.CreateUser) // 新規登録
	r.Post("/api/auth/login", authHandler.Login) // ログイン（JWT発行）

	// 【認証API】（防弾ガラスの内側：JWTトークンが必須）
	// チーム関連の操作
	r.Route("/api/teams/{team_id}", func(r chi.Router) {
		r.Use(mymiddleware.Auth) // ここを通るにはトークンが必要！

		r.Get("/documents", docHandler.ListDocumentsByTeam) // チームのドキュメント一覧
		r.Post("/members", teamHandler.AddMember)           // メンバー招待
		r.Get("/members", teamHandler.ListTeamMembers)      // メンバー一覧
	})

	// ドキュメント単体の操作
	r.Route("/api/documents", func(r chi.Router) {
		r.Use(mymiddleware.Auth) // ここを通るにはトークンが必要！

		r.Post("/", docHandler.CreateDocument) // ドキュメント作成（JWTから作成者を自動判定）

		// 特定のドキュメント（ID指定）の操作
		r.Route("/{document_id}", func(r chi.Router) {
			r.Put("/", docHandler.UpdateDocument)    // 更新（本人のみ）
			r.Delete("/", docHandler.DeleteDocument) // 削除（本人のみ）
		})
	})

	// ==========================================
	// 5. サーバー起動
	// ==========================================
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("サーバーをポート %s で起動しました...\n", port)
	err = http.ListenAndServe(":"+port, r)
	if err != nil {
		log.Fatal("サーバーの起動に失敗しました:", err)
	}
}
