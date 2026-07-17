// Package repository はデータ永続化の関心事だけを扱う。
//
// 設計方針:
//   - ドメインモデル(Task)はこのpackageで定義する（DBのテーブル構造に近い形）
//   - TaskRepository interface は「利用する側」が必要とする操作だけを定義する
//     （Goでは提供側ではなく利用側がinterfaceを定義する、という考え方。Chapter1参照）
//   - このpackageはGinや net/http を一切importしない。
//     HTTPの都合(ステータスコードなど)を知る必要がないため。
package repository

import (
	"context"
	"errors"
	"time"
)

// Task はドメインのモデル。DBのtasksテーブルに対応する。
type Task struct {
	ID          int
	Title       string
	Description string
	Done        bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ErrTaskNotFound はレコードが見つからなかったことを表すドメインエラー。
// sql.ErrNoRows のようなDB固有のエラーをそのまま上位層に漏らさないための変換先。
var ErrTaskNotFound = errors.New("task not found")

// TaskRepository はタスクの永続化操作を表すinterface。
// service層・handler層はこのinterfaceだけに依存し、具体的な実装(Postgres)を知らない。
type TaskRepository interface {
	FindAll(ctx context.Context) ([]Task, error)
	FindByID(ctx context.Context, id int) (Task, error)
	Create(ctx context.Context, task Task) (Task, error)
	Update(ctx context.Context, task Task) (Task, error)
	Delete(ctx context.Context, id int) error
}
