// Chapter3 演習スケルトン
//
// 事前準備:
//   go mod init example.com/chapter3
//   go get github.com/gin-gonic/gin
//   go get github.com/jackc/pgx/v5
//
// 事前にPostgreSQLへテーブルを作成しておいてください:
//   psql -h db -U postgres -d training_db -f migrations/001_create_tasks.sql
//
// docs/chapter3-api-development.md の「3. 実装」「7. 演習」を参照してください。
//
// 実行方法:
//   air
package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib" // driver登録のための副作用import
)

// ---- ドメインモデル ----

type Task struct {
	ID          int
	Title       string
	Description string
	Done        bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ---- ドメインエラー ----
// TODO(1): ErrTaskNotFound を errors.New で定義してください。

// ---- DTO ----
// TODO(2): CreateTaskRequest, UpdateTaskRequest, TaskResponse を
// docs/chapter3-api-development.md の「2.1 DTO」を参考に定義してください。
// binding タグ (required, min, max) も忘れずに。

// ---- Repository ----

type TaskRepository interface {
	FindAll() ([]Task, error)
	FindByID(id int) (Task, error)
	Create(task Task) (Task, error)
	Update(task Task) (Task, error)
	Delete(id int) error
}

type PostgresTaskRepository struct {
	db *sql.DB
}

func NewPostgresTaskRepository(db *sql.DB) *PostgresTaskRepository {
	return &PostgresTaskRepository{db: db}
}

// TODO(3): PostgresTaskRepositoryにTaskRepositoryの全メソッドを実装してください。
// SQL文の例:
//   SELECT id, title, description, done, created_at, updated_at FROM tasks WHERE id = $1
//   INSERT INTO tasks (title, description) VALUES ($1, $2) RETURNING id, ...
//   UPDATE tasks SET title=$1, description=$2, done=$3, updated_at=now() WHERE id=$4
//   DELETE FROM tasks WHERE id = $1
//
// sql.ErrNoRows が返ってきたらErrTaskNotFoundに変換することを忘れずに。

// ---- Handler ----

type TaskHandler struct {
	repo TaskRepository
}

func NewTaskHandler(repo TaskRepository) *TaskHandler {
	return &TaskHandler{repo: repo}
}

// TODO(4): List / Get / Create / Update / Delete のHandlerメソッドを実装してください。
// エラー発生時は docs/chapter3-api-development.md の「2.4 Error Response」を参考に
// handleError関数を用意し、ErrTaskNotFoundなら404、それ以外は500を返すようにしてください。

func connectDB() (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME"),
	)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func main() {
	db, err := connectDB()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	repo := NewPostgresTaskRepository(db)
	handler := NewTaskHandler(repo)
	_ = handler // TODO(5): ルーティングを登録したらこの行は削除してください

	r := gin.Default()

	// TODO(5): GET /tasks, GET /tasks/:id, POST /tasks, PUT /tasks/:id, DELETE /tasks/:id を登録してください。

	r.Run(":8080")
}

var _ = errors.New // TODO実装時に使うため一時的に残しています。使わなくなったら削除してください。
