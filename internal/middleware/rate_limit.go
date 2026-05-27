package middleware

import (
	"net"
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

// IPごとのリミッターを管理する構造体
type IPRateLimiter struct {
	ips   map[string]*rate.Limiter
	mu    sync.RWMutex
	rate  rate.Limit
	burst int
}

// 新しいIPRateLimiterを生成するファクトリ関数
func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	return &IPRateLimiter{
		ips:   make(map[string]*rate.Limiter),
		rate:  r,
		burst: b,
	}
}

// IPアドレスに対するリミッターを取得（または新規作成）する
func (i *IPRateLimiter) getLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter, exists := i.ips[ip]
	if !exists {
		// 存在しない場合は新しくバケツ（リミッター）を作成
		limiter = rate.NewLimiter(i.rate, i.burst)
		i.ips[ip] = limiter
	}
	return limiter
}

// レートリミットを適用するミドルウェア
func (i *IPRateLimiter) RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// リクエスト元からIPアドレスを抽出
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// そのIPのバケツからトークンを取り出せるかチェック
		limiter := i.getLimiter(ip)
		if !limiter.Allow() {
			// トークンが空（制限超過）の場合は429エラーを返し、後続の処理を遮断
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		// トークンがある場合は、次の処理（ハンドラー）へ進む
		next.ServeHTTP(w, r)
	})
}
