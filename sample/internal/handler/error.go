package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"example.com/taskapi/internal/dto"
	"example.com/taskapi/internal/repository"
)

// handleError はドメインエラーをHTTPステータスコードへ変換する唯一の場所。
// Handlerごとにエラー変換ロジックを重複させないための共通関数（Chapter3参照）。
func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, repository.ErrTaskNotFound):
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
	default:
		// 想定外のエラーは詳細をクライアントに漏らさない。
		// ログには詳細を出すが、レスポンスは一般的なメッセージにとどめる。
		c.Error(err) // gin.Context.Errors に記録。ロギングmiddlewareで出力される。
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal server error"})
	}
}
