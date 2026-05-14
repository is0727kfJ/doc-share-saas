package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"

	"github.com/is0727kfJ/doc-share-saas/internal/database"
)

func main() {
	// 1. .envファイルから設定を読み込む
	err := godotenv.Load()
	if err != nil {
		log.Println("環境変数ファイル(.env)が見つかりません。")
	}

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
	defer conn.Close(ctx) // プログラム終了時に自動で接続を閉じる
	fmt.Println("✅ データベース接続成功！")

	// 3. sqlcで自動生成された関数群（クライアント）を呼び出す準備
	queries := database.New(conn)

	// 4. テスト：新しいユーザーを作成してみる
	// ※CognitoのIDはダミーのランダム文字列を作成して使います
	dummySub := "cognito-sub-" + uuid.New().String()[:8]
	newUser, err := queries.CreateUser(ctx, database.CreateUserParams{
		CognitoSub: dummySub,
		Name:       "テストエンジニア",
		Email:      dummySub + "@example.com",
	})
	if err != nil {
		log.Fatalf("ユーザー作成失敗: %v\n", err)
	}

	// 成功したら結果を表示
	fmt.Printf("🎉 ユーザー作成成功！\nID: %s\n名前: %s\nEmail: %s\n", newUser.ID, newUser.Name, newUser.Email)
}
