// Package handler はHTTPの関心事だけを扱う。
//
// 設計方針:
//   - リクエストのパース・バリデーション・レスポンス変換だけを行う
//   - ビジネスロジックは持たない（それはserviceパッケージの責務）
//   - Ginにのみ依存し、database/sqlを直接importしない
package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"example.com/taskapi/internal/dto"
	"example.com/taskapi/internal/repository"
)

// TaskService はHandlerが必要とするservice層の操作。
// handlerパッケージ側でinterfaceを定義することで、
// 「Handlerが何を必要としているか」がこのファイルだけを読めばわかるようにする
// （Chapter1「呼び出し側がinterfaceを定義する」を参照）。
// 引数・戻り値の型としてrepository.Taskを使うのは、これがDBのテーブル構造ではなく
// 「アプリケーション全体で共有されるドメインモデル」だからで、Go実務でもよくある形。
type TaskService interface {
	ListTasks(ctx context.Context) ([]repository.Task, error)
	GetTask(ctx context.Context, id int) (repository.Task, error)
	CreateTask(ctx context.Context, title, description string) (repository.Task, error)
	UpdateTask(ctx context.Context, id int, title, description string, done bool) (repository.Task, error)
	DeleteTask(ctx context.Context, id int) error
}

type TaskHandler struct {
	service TaskService
}

func NewTaskHandler(service TaskService) *TaskHandler {
	return &TaskHandler{service: service}
}

func (h *TaskHandler) List(c *gin.Context) {
	tasks, err := h.service.ListTasks(c.Request.Context())
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, toTaskResponseList(tasks))
}

func (h *TaskHandler) Get(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}

	task, err := h.service.GetTask(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, toTaskResponse(task))
}

func (h *TaskHandler) Create(c *gin.Context) {
	var req dto.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	task, err := h.service.CreateTask(c.Request.Context(), req.Title, req.Description)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toTaskResponse(task))
}

func (h *TaskHandler) Update(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}

	var req dto.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	task, err := h.service.UpdateTask(c.Request.Context(), id, req.Title, req.Description, req.Done)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, toTaskResponse(task))
}

func (h *TaskHandler) Delete(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}

	if err := h.service.DeleteTask(c.Request.Context(), id); err != nil {
		handleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func parseIDParam(c *gin.Context) (int, bool) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid id"})
		return 0, false
	}
	return id, true
}
