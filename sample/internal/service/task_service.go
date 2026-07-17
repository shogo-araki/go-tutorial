// Package service はビジネスロジックを扱う。
//
// 設計方針（Chapter4参照）:
//   - 単純なCRUDの受け渡しだけならHandlerがRepositoryを直接呼んでもよい
//   - 複数の手順を組み合わせる処理（例: CompleteTask）はServiceに集約する
//   - Serviceはrepository.TaskRepository interfaceにのみ依存し、
//     Postgresなどの具体的な実装を知らない
package service

import (
	"context"
	"fmt"
	"time"

	"example.com/taskapi/internal/repository"
)

type TaskService struct {
	repo repository.TaskRepository
}

func NewTaskService(repo repository.TaskRepository) *TaskService {
	return &TaskService{repo: repo}
}

func (s *TaskService) ListTasks(ctx context.Context) ([]repository.Task, error) {
	return s.repo.FindAll(ctx)
}

func (s *TaskService) GetTask(ctx context.Context, id int) (repository.Task, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *TaskService) CreateTask(ctx context.Context, title, description string) (repository.Task, error) {
	return s.repo.Create(ctx, repository.Task{Title: title, Description: description})
}

func (s *TaskService) UpdateTask(ctx context.Context, id int, title, description string, done bool) (repository.Task, error) {
	return s.repo.Update(ctx, repository.Task{
		ID:          id,
		Title:       title,
		Description: description,
		Done:        done,
	})
}

func (s *TaskService) DeleteTask(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

// CompleteTask はタスクを「完了」にするドメイン操作。
// 単純なUpdateとは異なり、「既存のタスクを取得してから、完了フラグだけを変更して保存する」
// という複数手順を1つの操作としてまとめている。これがServiceに処理を集約する典型例。
func (s *TaskService) CompleteTask(ctx context.Context, id int) (repository.Task, error) {
	task, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return repository.Task{}, err
	}

	if task.Done {
		return repository.Task{}, fmt.Errorf("task %d is already done", id)
	}

	task.Done = true
	task.UpdatedAt = time.Now()

	return s.repo.Update(ctx, task)
}
