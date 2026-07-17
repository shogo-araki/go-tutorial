// Package dto はAPIの入出力の「形」だけを定義する。
//
// なぜrepositoryパッケージのTaskをそのままJSONで返さないのか（Chapter3参照）:
//   - DBのテーブル構造の変更がそのままAPI仕様の変更にならないようにするため
//   - リクエストごとに異なるバリデーションルールを表現するため
package dto

import "time"

// CreateTaskRequest は POST /tasks のリクエストボディ。
type CreateTaskRequest struct {
	Title       string `json:"title" binding:"required,min=1,max=100"`
	Description string `json:"description" binding:"max=1000"`
}

// UpdateTaskRequest は PUT /tasks/:id のリクエストボディ。
type UpdateTaskRequest struct {
	Title       string `json:"title" binding:"required,min=1,max=100"`
	Description string `json:"description" binding:"max=1000"`
	Done        bool   `json:"done"`
}

// TaskResponse はクライアントに返すタスクの形。
type TaskResponse struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Done        bool      `json:"done"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ErrorResponse は統一されたエラーレスポンスの形。
type ErrorResponse struct {
	Error string `json:"error"`
}
