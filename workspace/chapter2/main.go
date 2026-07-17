// Chapter2 演習スケルトン
//
// 事前準備 (このファイルと同じディレクトリで実行してください):
//   go mod init example.com/chapter2
//   go get github.com/gin-gonic/gin
//
// docs/chapter2-gin.md の「3. 実装」「7. 演習」を参照しながら
// 以下のTODOを埋めてください。
//
// 実行方法:
//   air        (ホットリロード。.air.toml を先に用意すること)
//   go run .   (ホットリロードなしで直接実行)
package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// Chapter1のTaskをベースに、JSON用のstruct tagを追加してください。
// TODO(1): Task struct に `json:"id"` などのタグを追加する
type Task struct {
	ID    int
	Title string
	Done  bool
}

// TODO(2): TaskRepository interface と InMemoryTaskRepository を
// Chapter1の実装を参考に用意してください（Save時にIDを自動採番するようにする）。

// TODO(3): TaskHandler struct を定義し、repo をフィールドに持たせてください。
// Handlerは以下の3つのメソッドを持ちます。
//   - List(c *gin.Context)   GET /tasks
//   - Get(c *gin.Context)    GET /tasks/:id
//   - Create(c *gin.Context) POST /tasks
//
// ヒント:
//   c.Param("id")        パスパラメータ取得
//   c.ShouldBindJSON(&v) リクエストボディのJSONをvにバインド
//   c.JSON(status, body) レスポンス送信

func main() {
	r := gin.Default()

	// TODO(4): repoとhandlerを生成し、ルーティングを登録してください。
	// r.GET("/tasks", handler.List)
	// r.GET("/tasks/:id", handler.Get)
	// r.POST("/tasks", handler.Create)

	fmt.Println("TODOを実装してください")

	r.Run(":8080")
}
