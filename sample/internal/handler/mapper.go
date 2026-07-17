package handler

import (
	"example.com/taskapi/internal/dto"
	"example.com/taskapi/internal/repository"
)

// toTaskResponse はドメインモデル(repository.Task)をAPIレスポンス用のDTOに変換する。
// 自動マッピングライブラリを使わず手動で書くのがGoの流儀（Chapter3参照）。
func toTaskResponse(t repository.Task) dto.TaskResponse {
	return dto.TaskResponse{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		Done:        t.Done,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

func toTaskResponseList(tasks []repository.Task) []dto.TaskResponse {
	result := make([]dto.TaskResponse, 0, len(tasks))
	for _, t := range tasks {
		result = append(result, toTaskResponse(t))
	}
	return result
}
