package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// PostgresTaskRepository は TaskRepository interface のPostgreSQL実装。
// テスト時はこの実装の代わりにインメモリ実装を差し込むことができる
// （Chapter1で学んだ「interfaceによる差し替え可能性」の実務での使いどころ）。
type PostgresTaskRepository struct {
	db *sql.DB
}

func NewPostgresTaskRepository(db *sql.DB) *PostgresTaskRepository {
	return &PostgresTaskRepository{db: db}
}

// コンパイル時にinterfaceを満たしているかを検査するための慣用句。
// 「_」に代入することで、実行時には何もしないが、
// PostgresTaskRepositoryがTaskRepositoryを満たしていなければここでコンパイルエラーになる。
var _ TaskRepository = (*PostgresTaskRepository)(nil)

func (r *PostgresTaskRepository) FindAll(ctx context.Context) ([]Task, error) {
	const query = `
		SELECT id, title, description, done, created_at, updated_at
		FROM tasks
		ORDER BY id`

	rows, err := r.db.QueryContext(ctx, query)
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
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tasks: %w", err)
	}
	return tasks, nil
}

func (r *PostgresTaskRepository) FindByID(ctx context.Context, id int) (Task, error) {
	const query = `
		SELECT id, title, description, done, created_at, updated_at
		FROM tasks
		WHERE id = $1`

	var t Task
	err := r.db.QueryRowContext(ctx, query, id).
		Scan(&t.ID, &t.Title, &t.Description, &t.Done, &t.CreatedAt, &t.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Task{}, ErrTaskNotFound
	}
	if err != nil {
		return Task{}, fmt.Errorf("find task by id: %w", err)
	}
	return t, nil
}

func (r *PostgresTaskRepository) Create(ctx context.Context, task Task) (Task, error) {
	const query = `
		INSERT INTO tasks (title, description)
		VALUES ($1, $2)
		RETURNING id, title, description, done, created_at, updated_at`

	var t Task
	err := r.db.QueryRowContext(ctx, query, task.Title, task.Description).
		Scan(&t.ID, &t.Title, &t.Description, &t.Done, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return Task{}, fmt.Errorf("create task: %w", err)
	}
	return t, nil
}

func (r *PostgresTaskRepository) Update(ctx context.Context, task Task) (Task, error) {
	const query = `
		UPDATE tasks
		SET title = $1, description = $2, done = $3, updated_at = now()
		WHERE id = $4
		RETURNING id, title, description, done, created_at, updated_at`

	var t Task
	err := r.db.QueryRowContext(ctx, query, task.Title, task.Description, task.Done, task.ID).
		Scan(&t.ID, &t.Title, &t.Description, &t.Done, &t.CreatedAt, &t.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Task{}, ErrTaskNotFound
	}
	if err != nil {
		return Task{}, fmt.Errorf("update task: %w", err)
	}
	return t, nil
}

func (r *PostgresTaskRepository) Delete(ctx context.Context, id int) error {
	const query = `DELETE FROM tasks WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
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
