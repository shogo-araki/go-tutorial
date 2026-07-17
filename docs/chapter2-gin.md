# Chapter 2: Gin入門 ― Routing / Handler / Context / JSON

対象読者: Chapter1を完了した人

この章の目的は「Ginを覚えること」ではなく、**Ginという薄いフレームワークを通してGoのHTTPサーバーの素朴な作り方を理解すること**です。
ASP.NET Core(Controller/Middleware)やExpress.js(Router/Middleware)との対比で説明します。

---

## 1. 目的

- `net/http` の上にGinがどう乗っているかを理解する
- Handler / Context / Routing の役割をC#・Node.jsのフレームワークと対応付けて理解する
- Chapter1で作った `Task` / `TaskRepository` をGin経由でHTTP公開する

---

## 2. 解説

### 2.1 なぜGinなのか（そしてなぜ最初から入れないのか）

Goの標準ライブラリ `net/http` だけでもWebサーバーは作れます。実際、Ginは`net/http`の薄いラッパーに過ぎません。
最初からGinを入れずChapter1を`net/http`なしで進めたのは、**「フレームワークが何を肩代わりしてくれているか」を後から実感するため**です。

ASP.NET Coreの `Controller` や `IActionResult`、Express.jsの `app.get(path, handler)` に相当する機能を、
Ginは以下の3点で提供します。

- **Router**: URLパスとHandlerの対応付け
- **Context**: リクエスト情報の取得とレスポンスの書き込みをまとめたオブジェクト
- **Middleware**: リクエスト処理の前後に共通処理を挟む仕組み

### 2.2 Ginの導入

```bash
cd workspace/chapter2
go mod init example.com/chapter2
go get github.com/gin-gonic/gin
```

`go get` は npmでいう `npm install`、NuGetでいう `dotnet add package` に相当します。
実行すると `go.mod` に依存関係が追記され、`go.sum` というチェックサムファイルが生成されます。
`go.sum` は `package-lock.json` や `.csproj` の `<PackageReference>` バージョン固定に相当し、
**依存パッケージの改ざん検知**の役割も持ちます。commitに含めるのが基本です。

### 2.3 Routing ― URLとHandlerの対応

```go
r := gin.Default()

r.GET("/tasks", listTasksHandler)
r.GET("/tasks/:id", getTaskHandler)
r.POST("/tasks", createTaskHandler)

r.Run(":8080")
```

ASP.NET Coreの `[HttpGet("tasks/{id}")]` 属性ルーティング、Express.jsの `router.get('/tasks/:id', handler)` と
ほぼ同じ感覚です。`:id` はパスパラメータで、Express.jsの記法とそのまま一致します。

### 2.4 Handler ― 「関数」であってControllerクラスではない

GinのHandlerは以下のシグネチャを持つ**ただの関数**です。

```go
func getTaskHandler(c *gin.Context) {
    // ...
}
```

ASP.NET Coreのように「Controllerクラスにactionメソッドを生やす」設計ではありません。
Goには継承もクラスもないため、Handlerを分類・整理する単位は「クラス」ではなく **struct + method** になります。

```go
type TaskHandler struct {
    repo TaskRepository // Chapter1で作ったinterface
}

func NewTaskHandler(repo TaskRepository) *TaskHandler {
    return &TaskHandler{repo: repo}
}

func (h *TaskHandler) Get(c *gin.Context) {
    // h.repo を使って処理する
}

// ルーティング登録時
r.GET("/tasks/:id", taskHandler.Get)
```

これは「Controllerクラスにフィールドとしてservice/repositoryを注入する」ASP.NET Coreの設計とほぼ同じ発想です。
違いはDIコンテナが自動でやってくれるか、`NewTaskHandler(repo)` のように**自分でコンストラクタ関数を書いて手動配線するか**という点です
（Goには標準のDIコンテナがありません。これはChapter4の「Dependency管理」で詳しく扱います）。

### 2.5 Context ― Request/Responseをまとめたオブジェクト

`*gin.Context` は、ASP.NET Coreの `HttpContext`、Express.jsの `(req, res)` に相当します。

```go
func (h *TaskHandler) Get(c *gin.Context) {
    id := c.Param("id")               // ルートパラメータ取得 (req.params.id 相当)
    query := c.Query("verbose")       // クエリパラメータ取得 (req.query.verbose 相当)

    task, err := h.repo.FindByID(...)
    if err != nil {
        c.JSON(404, gin.H{"error": err.Error()}) // res.status(404).json(...) 相当
        return
    }
    c.JSON(200, task)
}
```

Express.jsを触った経験があれば `(req, res) => {}` とほぼ1対1で対応付けて理解できるはずです。
違いは、Ginでは `req` と `res` が1つの `Context` オブジェクトに統合されている点だけです。

### 2.6 JSON ― structとJSONの対応付け

Goでは構造体のJSON変換ルールを **struct tag** で指定します。

```go
type Task struct {
    ID    int    `json:"id"`
    Title string `json:"title"`
    Done  bool   `json:"done"`
}
```

C#の `[JsonPropertyName("title")]` 属性、TypeScriptで型とAPIレスポンスをそのまま対応付ける感覚に近いです。
`c.JSON(200, task)` を呼ぶと、structが自動的にこのタグに従ってJSONへ変換されます。
struct tagを省略した場合、フィールド名がそのまま(大文字始まりの)JSONキーになる点に注意してください
（実務では基本的に全フィールドに`json`タグを明示するのが定石です）。

---

## 3. 実装

`workspace/chapter2/` で以下を行ってください。

1. `go mod init example.com/chapter2`
2. `go get github.com/gin-gonic/gin`
3. スケルトンの `TODO` を埋めて、以下のエンドポイントを実装する

| Method | Path | 内容 |
|---|---|---|
| GET | `/tasks` | 全件取得 |
| GET | `/tasks/:id` | 1件取得（存在しなければ404） |
| POST | `/tasks` | 新規作成（Bodyの`title`を使う） |

4. Air用の設定ファイル `.air.toml` を用意し、ホットリロードで動作確認する

```bash
cd workspace/chapter2
air
```

保存するたびに自動でビルド・再起動されることを確認してください。

---

## 4. コード解説

スケルトンでは `main.go` の中に直接Handlerを書いていますが、これはあくまでChapter2時点での**仮の姿**です。
実務では `main.go` に全部書くことはなく、Chapter4で段階的に `handler/` パッケージへ分離していきます。
「まず動くものを作り、後から整理する」というボトムアップの進め方は、Goのコミュニティでもよく採られるアプローチです。

---

## 5. C# / TypeScript比較

| 概念 | ASP.NET Core | Express.js | Gin |
|---|---|---|---|
| ルーティング定義 | 属性ルーティング / `MapGet` | `router.get(path, handler)` | `r.GET(path, handler)` |
| Handlerの単位 | Controllerクラスのaction | 関数 | struct method（または関数） |
| Request/Response | `HttpContext`（分離アクセス可） | `(req, res)` | `*gin.Context`（統合） |
| JSONシリアライズ | System.Text.Json + 属性 | 標準で暗黙変換 | struct tag (`json:"..."`) |
| DI | 標準搭載（`IServiceCollection`） | 標準搭載なし（自前 or ライブラリ） | 標準搭載なし（自前で配線） |

一番の違いは **DIコンテナの有無**です。ASP.NET Coreの `builder.Services.AddScoped<...>()` のような仕組みはGinにはありません。
Goでは「コンストラクタ関数(`NewXxx`)を自分で呼んで依存関係を手渡しする」のが標準的なやり方です。
これは面倒に見えますが、**「何が何に依存しているか」がコード上に明示的に現れる**という利点があり、Goらしいシンプルさの象徴でもあります。

---

## 6. 実務利用例

実務では以下のような構成でGinを使うことが一般的です。

- Handlerはstructにして依存(`service`や`repository`)をフィールドに持たせる
- Middlewareでログ・認証・panic recoveryを共通化する（Chapter4で扱う）
- ルーティング定義は `router.go` のような専用ファイルに集約し、`main.go` を薄く保つ

---

## 7. 演習

1. `GET /tasks`, `GET /tasks/:id`, `POST /tasks` を実装し、`curl` で動作確認してください。
2. 存在しないIDで `GET /tasks/:id` を呼んだ際、ステータスコード404と `{"error": "..."}` 形式のJSONが返るようにしてください。
3. `.air.toml` を設定し、コード変更時に自動リロードされることを確認してください。
4. （発展）`PUT /tasks/:id`（更新）と `DELETE /tasks/:id`（削除）を自分で実装してください（Chapter1の`TaskRepository`にメソッド追加が必要な場合は追加してください）。

---

## 8. 完成例

<details>
<summary>クリックして完成例コードを表示</summary>

```go
package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Task struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

type TaskRepository interface {
	FindByID(id int) (Task, error)
	FindAll() []Task
	Save(task Task) error
}

type InMemoryTaskRepository struct {
	data   map[int]Task
	nextID int
}

func NewInMemoryTaskRepository() *InMemoryTaskRepository {
	return &InMemoryTaskRepository{data: map[int]Task{}, nextID: 1}
}

func (r *InMemoryTaskRepository) FindByID(id int) (Task, error) {
	t, ok := r.data[id]
	if !ok {
		return Task{}, fmt.Errorf("task not found: id=%d", id)
	}
	return t, nil
}

func (r *InMemoryTaskRepository) FindAll() []Task {
	result := make([]Task, 0, len(r.data))
	for _, t := range r.data {
		result = append(result, t)
	}
	return result
}

func (r *InMemoryTaskRepository) Save(task Task) error {
	if task.ID == 0 {
		task.ID = r.nextID
		r.nextID++
	}
	r.data[task.ID] = task
	return nil
}

// TaskHandler はGinのHandlerをまとめるstruct（Controllerに相当）
type TaskHandler struct {
	repo TaskRepository
}

func NewTaskHandler(repo TaskRepository) *TaskHandler {
	return &TaskHandler{repo: repo}
}

func (h *TaskHandler) List(c *gin.Context) {
	c.JSON(http.StatusOK, h.repo.FindAll())
}

func (h *TaskHandler) Get(c *gin.Context) {
	var id int
	if _, err := fmt.Sscanf(c.Param("id"), "%d", &id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	task, err := h.repo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, task)
}

type createTaskRequest struct {
	Title string `json:"title"`
}

func (h *TaskHandler) Create(c *gin.Context) {
	var req createTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.New("title is required").Error()})
		return
	}

	task := Task{Title: req.Title}
	if err := h.repo.Save(task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, task)
}

func main() {
	repo := NewInMemoryTaskRepository()
	handler := NewTaskHandler(repo)

	r := gin.Default()
	r.GET("/tasks", handler.List)
	r.GET("/tasks/:id", handler.Get)
	r.POST("/tasks", handler.Create)

	r.Run(":8080")
}
```

</details>

次章 `docs/chapter3-api-development.md` に進んでください。ここからPostgreSQLとバリデーションを導入します。
