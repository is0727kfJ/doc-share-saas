package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"

	"github.com/is0727kfJ/doc-share-saas/internal/database"
	"github.com/is0727kfJ/doc-share-saas/internal/document"
	"github.com/is0727kfJ/doc-share-saas/internal/team"
	"github.com/is0727kfJ/doc-share-saas/internal/user"
)

func main() {
	_ = godotenv.Load() // .envファイルがなくても続行するため、エラーは無視
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URLが設定されていません")
	}

	// 2. データベースへ接続する
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		log.Fatalf("データベース接続エラー: %v\n", err)
	}
	defer conn.Close(ctx)

	// 3. sqlcで自動生成された関数群呼び出す準備
	queries := database.New(conn)

	r := chi.NewRouter()

	r.Use(middleware.Logger)    // ロギングミドルウェアを追加
	r.Use(middleware.Recoverer) // パニックからの回復ミドルウェアを追加
	r.Use(cors.Handler(cors.Options{
		// Next.jsが動くポート（3000）からのアクセスだけを許可する
		AllowedOrigins: []string{"http://localhost:3000"},
		// 許可するHTTPメソッド（今回作ったCRUDをすべて許可）
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		// フロントエンドから送られてくるヘッダーを許可
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		// ブラウザにキャッシュさせる時間（秒）
		MaxAge: 300,
	}))

	// ルートハンドラーを定義
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	userHandler := user.NewHandler(queries)
	docHandler := document.NewHandler(queries)
	teamHandler := team.NewHandler(queries)

	r.Post("/api/users", userHandler.CreateUser)
	r.Post("/api/documents", docHandler.CreateDocument)
	r.Get("/api/documents", docHandler.ListDocuments)
	r.Post("/api/teams", teamHandler.CreateTeam)

	r.Route("/api/teams/{team_id}", func(r chi.Router) {
		// GET /api/teams/{team_id}/documents
		r.Get("/documents", docHandler.ListDocumentsByTeam)
		// POST /api/teams/{team_id}/members
		r.Post("/members", teamHandler.AddMember)
		// GET /api/teams/{team_id}/members
		r.Get("/members", teamHandler.ListTeamMembers)
	})

	r.Route("/api/documents/{document_id}", func(r chi.Router) {
		r.Put("/", docHandler.UpdateDocument)
		r.Delete("/", docHandler.DeleteDocument)
	})

	// 4. HTTPサーバーを起動する
	fmt.Println("APIサーバーが http://localhost:8080 で起動しました")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("サーバーエラー: %v\n", err)
	}

}
