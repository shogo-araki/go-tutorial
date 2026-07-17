// Package middleware はリクエスト処理の前後に挟む共通処理を提供する。
// ASP.NET CoreのMiddleware、Express.jsの app.use(...) に相当する（Chapter4参照）。
package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// Logging は全リクエストの処理時間・ステータスコードをログ出力するmiddleware。
func Logging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next() // 後続のHandler（および他のmiddleware）を実行する

		duration := time.Since(start)
		status := c.Writer.Status()

		log.Printf("[%s] %s %s -> %d (%v)",
			c.Request.Method,
			c.Request.URL.Path,
			c.ClientIP(),
			status,
			duration,
		)

		// handler.handleError で c.Error(err) された内部エラーがあればここでログに残す。
		// クライアントには詳細を返さず、サーバー側のログにだけ詳細を残すのが実務での定石。
		for _, e := range c.Errors {
			log.Printf("  internal error: %v", e.Err)
		}
	}
}
