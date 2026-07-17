// Stage1: Chapter3相当の「全部乗せ」main.go
//
// このファイルはこのまま動作します。まずは読んで、演習1の指示に従い
// 「HTTPの関心事」「DBの関心事」をコメントで分類してください。
//
// 事前準備:
//   go mod init example.com/chapter4
//   go get github.com/gin-gonic/gin
//   go get github.com/jackc/pgx/v5
//   psql -h db -U postgres -d training_db -f ../../chapter3/migrations/001_create_tasks.sql
package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Task struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Done        bool      `json:"done"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

var ErrTaskNotFound = errors.New("task not found")

type CreateTaskRequest struct {
	Title       string `json:"title" binding:"required,min=1,max=100"`
	Description string `json:"description" binding:"max=1000"`
}

type UpdateTaskRequest struct {
	Title       string `json:"title" binding:"required,min=1,max=100"`
	Description string `json:"description" binding:"max=1000"`
	Done        bool   `json:"done"`
}

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

func (r *PostgresTaskRepository) FindAll() ([]Task, error) {
	rows, err := r.db.Query(`SELECT id, title, description, done, created_at, updated_at FROM tasks ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("find all tasks: %w", err)
	}
	defer rows.Close()

	tasks := []Task{}
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Done, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (r *PostgresTaskRepository) FindByID(id int) (Task, error) {
	var t Task
	query := `SELECT id, title, description, done, created_at, updated_at FROM tasks WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(&t.ID, &t.Title, &t.Description, &t.Done, &t.CreatedAt, &t.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Task{}, ErrTaskNotFound
	}
	if err != nil {
		return Task{}, fmt.Errorf("find task by id: %w", err)
	}
	return t, nil
}

func (r *PostgresTaskRepository) Create(task Task) (Task, error) {
	query := `INSERT INTO tasks (title, description) VALUES ($1, $2)
	          RETURNING id, title, description, done, created_at, updated_at`
	var t Task
	err := r.db.QueryRow(query, task.Title, task.Description).
		Scan(&t.ID, &t.Title, &t.Description, &t.Done, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return Task{}, fmt.Errorf("create task: %w", err)
	}
	return t, nil
}

func (r *PostgresTaskRepository) Update(task Task) (Task, error) {
	query := `UPDATE tasks SET title=$1, description=$2, done=$3, updated_at=now()
	          WHERE id=$4
	          RETURNING id, title, description, done, created_at, updated_at`
	var t Task
	err := r.db.QueryRow(query, task.Title, task.Description, task.Done, task.ID).
		Scan(&t.ID, &t.Title, &t.Description, &t.Done, &t.CreatedAt, &t.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Task{}, ErrTaskNotFound
	}
	if err != nil {
		return Task{}, fmt.Errorf("update task: %w", err)
	}
	return t, nil
}

func (r *PostgresTaskRepository) Delete(id int) error {
	result, err := r.db.Exec(`DELETE FROM tasks WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete task rows affected: %w", err)
	}
	if affected == 0 {
		return ErrTaskNotFound
	}
	return nil
}

type TaskHandler struct {
	repo TaskRepository
}

func NewTaskHandler(repo TaskRepository) *TaskHandler {
	return &TaskHandler{repo: repo}
}

func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrTaskNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}

func (h *TaskHandler) List(c *gin.Context) {
	tasks, err := h.repo.FindAll()
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, tasks)
}

func parseIDParam(c *gin.Context) (int, bool) {
	var id int
	if _, err := fmt.Sscanf(c.Param("id"), "%d", &id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return 0, false
	}
	return id, true
}

func (h *TaskHandler) Get(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	task, err := h.repo.FindByID(id)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, task)
}

func (h *TaskHandler) Create(c *gin.Context) {
	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	task, err := h.repo.Create(Task{Title: req.Title, Description: req.Description})
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, task)
}

func (h *TaskHandler) Update(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	var req UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	task, err := h.repo.Update(Task{ID: id, Title: req.Title, Description: req.Description, Done: req.Done})
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, task)
}

func (h *TaskHandler) Delete(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	if err := h.repo.Delete(id); err != nil {
		handleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

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

	r := gin.Default()
	r.GET("/tasks", handler.List)
	r.GET("/tasks/:id", handler.Get)
	r.POST("/tasks", handler.Create)
	r.PUT("/tasks/:id", handler.Update)
	r.DELETE("/tasks/:id", handler.Delete)

	r.Run(":8080")
}
