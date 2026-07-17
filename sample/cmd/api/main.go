// cmd/api/main.go は起動処理のみを担う。
//
// このファイルの役割は以下の3つだけ:
//   1. 設定・DB接続などの「外部リソース」を用意する
//   2. 依存関係を組み立てる（repository -> service -> handler の順に生成し、注入する）
//   3. ルーティングを登録してサーバーを起動する
//
// ビジネスロジックやHTTPの詳細はここには書かない（それぞれ internal/ 配下の責務）。
// C#の Program.cs で builder.Services.AddScoped<...>() を並べる代わりに、
// Goでは New〇〇(...) を手続き的に呼んで依存を組み立てる（Chapter4参照）。
package main

import (
	"database/sql"
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib" // driver登録のための副作用import

	"example.com/taskapi/internal/config"
	"example.com/taskapi/internal/handler"
	"example.com/taskapi/internal/middleware"
	"example.com/taskapi/internal/repository"
	"example.com/taskapi/internal/service"
)

func main() {
	cfg := config.Load()

	db, err := connectDB(cfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// 依存関係の組み立て（DI）。ここが「配線図」そのもの。
	taskRepo := repository.NewPostgresTaskRepository(db)
	taskService := service.NewTaskService(taskRepo)
	taskHandler := handler.NewTaskHandler(taskService)

	r := gin.Default()
	r.Use(middleware.Logging(), gin.Recovery())

	registerRoutes(r, taskHandler)

	log.Printf("starting server on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}

func registerRoutes(r *gin.Engine, taskHandler *handler.TaskHandler) {
	tasks := r.Group("/tasks")
	{
		tasks.GET("", taskHandler.List)
		tasks.GET("/:id", taskHandler.Get)
		tasks.POST("", taskHandler.Create)
		tasks.PUT("/:id", taskHandler.Update)
		tasks.DELETE("/:id", taskHandler.Delete)
	}

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
}

func connectDB(cfg config.Config) (*sql.DB, error) {
	db, err := sql.Open("pgx", cfg.DSN())
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
