# Chapter 3: API開発 ― Validation / DTO / PostgreSQL / CRUD / Error Response

対象読者: Chapter2を完了した人

この章では、Chapter2で作ったインメモリのTask APIを、**実際にPostgreSQLへ永続化するCRUD API**に発展させます。
あわせて、実務で必須になる「バリデーション」「DTOと内部モデルの分離」「統一されたエラーレスポンス」を扱います。

---

## 1. 目的

- リクエストの検証をGinの`binding`タグで行う（ASP.NET Coreの`DataAnnotations`、TypeScriptのzod/class-validatorとの対比）
- 「APIの入出力の形（DTO）」と「ドメインのデータ構造」を分離する理由を理解する
- `database/sql` を使ったPostgreSQLへの素朴なアクセス方法を理解する
- エラーレスポンスのフォーマットを統一し、HTTPステータスコードと対応付ける

---

## 2. 解説

### 2.1 DTO ― なぜstructを2つ用意するのか

Chapter2では `Task` structをそのままAPIのレスポンスとして使っていました。実務ではこれはアンチパターンです。

```go
// ドメインのモデル（DBのテーブル構造に近い）
type Task struct {
    ID        int
    Title       string
    Description string
    Done        bool
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// リクエスト用DTO（クライアントから受け取る値だけを持つ）
type CreateTaskRequest struct {
    Title       string `json:"title" binding:"required,min=1,max=100"`
    Description string `json:"description" binding:"max=1000"`
}

// レスポンス用DTO（クライアントに返す値だけを持つ）
type TaskResponse struct {
    ID          int       `json:"id"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    Done        bool      `json:"done"`
    CreatedAt   time.Time `json:"created_at"`
}
```

C#実務での `Entity` と `DTO`(または`ViewModel`)を分ける設計、TypeScriptでの `Domain Model` と `API Schema` を
分ける設計と同じ理由です。

- DBのカラムがそのままAPIに漏れるのを防ぐ（例: 内部管理用のフラグを誤って返さない）
- リクエストごとに必要なバリデーションルールが異なる（作成時は`title`必須、更新時は任意、など）
- DBのテーブル構造を変えてもAPIの互換性を保てる

Goでは自動マッピングライブラリ（AutoMapperのようなもの）を使う文化が薄く、**変換関数を自分で書く**のが一般的です。

```go
func toTaskResponse(t Task) TaskResponse {
    return TaskResponse{
        ID:          t.ID,
        Title:       t.Title,
        Description: t.Description,
        Done:        t.Done,
        CreatedAt:   t.CreatedAt,
    }
}
```

冗長に見えますが、「何が変換されているか」がコード上に明示され、**リフレクションを使った暗黙の変換に依存しない**というGoらしさの現れです。

### 2.2 Validation ― struct tagによる宣言的な検証

Ginは内部で `go-playground/validator` を使っており、struct tagの `binding` で検証ルールを宣言します。

```go
type CreateTaskRequest struct {
    Title string `json:"title" binding:"required,min=1,max=100"`
}

func (h *TaskHandler) Create(c *gin.Context) {
    var req CreateTaskRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    // この時点でreqは検証済み
}
```

C#の `DataAnnotations`(`[Required]`, `[StringLength]`)、TypeScriptの`class-validator`や`zod`のスキーマ定義と
発想は同じです。C#/TypeScriptとの違いは、**Goには「検証結果を自動でHTTP 400に変換する仕組み」が標準搭載されていない**点です。
`ShouldBindJSON`が返す`err`を、自分でハンドリングして適切なレスポンスに変換する必要があります（2.4節で扱います）。

### 2.3 PostgreSQL ― database/sqlの素朴な使い方

Goの標準ライブラリには `database/sql` というDBアクセスの共通インターフェースがあります。
C#の`ADO.NET`（`DbConnection`/`DbCommand`）に近い立ち位置で、ORMではなく**薄いラッパー**です。

```go
import (
    "database/sql"
    _ "github.com/jackc/pgx/v5/stdlib" // driverを登録するためのimport（副作用import）
)

db, err := sql.Open("pgx", dsn)
```

`_ "github.com/jackc/pgx/v5/stdlib"` という書き方に注目してください。**パッケージ名を使わずimportだけする**構文で、
「このパッケージの`init()`関数だけ実行してほしい（driverの登録処理だけ動かしたい）」という意図を表します。
C#/TypeScriptにはない書き方ですが、「副作用のためだけにモジュールを読み込む」という点ではNode.jsの
`import "some-polyfill"` に近い感覚です。

このプロジェクトではORM（GORMなど）をあえて使わず、SQLを直接書く方針にしています。理由は「実務で使う設計を学ぶ」という
教材の目的上、**SQLとGoの型の対応関係を最初にきちんと理解してほしい**ためです。ORMは後から自分で選定・導入できます。

```go
func (r *PostgresTaskRepository) FindByID(ctx context.Context, id int) (Task, error) {
    var t Task
    query := `SELECT id, title, description, done, created_at, updated_at FROM tasks WHERE id = $1`
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
```

`context.Context` を第一引数に取っているのもGoらしいポイントです。C#の`CancellationToken`に相当し、
「リクエストがキャンセルされたらDBアクセスも中断する」という制御を、**関数シグネチャ上で明示的に伝搬させる**のがGoの流儀です
（`ctx`を第一引数に取るのはGoコミュニティの強い慣習です）。

`fmt.Errorf("find task by id: %w", err)` の `%w` にも注目してください。これは**エラーのラップ**です。
元のエラー(`err`)を保持したまま、文脈情報（「どの処理で失敗したか」）を付加できます。
呼び出し元は `errors.Is` / `errors.As` を使って「元のエラーが何だったか」を判定できます。
C#の `throw new ApplicationException("...", innerException)` に近い考え方です。

### 2.4 Error Response ― エラーを一貫した形でクライアントに返す

Handlerごとに `gin.H{"error": err.Error()}` をバラバラに書くと、実務ではすぐに破綻します。
「ドメインのエラー種別」と「HTTPステータスコード」を対応付ける仕組みを用意します。

```go
var (
    ErrTaskNotFound   = errors.New("task not found")
    ErrValidation     = errors.New("validation error")
)

func handleError(c *gin.Context, err error) {
    switch {
    case errors.Is(err, ErrTaskNotFound):
        c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
    case errors.Is(err, ErrValidation):
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    default:
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
    }
}
```

これはC#のASP.NET Coreにおける`ExceptionFilter`やグローバルエラーハンドラーに相当する役割を、
**例外機構なしで（戻り値のerrorだけで）実現している**点がGoらしさです。
Chapter4ではこれをmiddleware化し、Handlerごとの重複をさらに減らします。

---

## 3. 実装

`workspace/chapter3/` で以下を行ってください。

1. `docker compose up -d` でPostgreSQLを起動する（Chapter0で構築済みの環境を使う）
2. `workspace/chapter3/migrations/001_create_tasks.sql` を参考にテーブルを作成する
3. `go get github.com/jackc/pgx/v5` を実行する
4. スケルトンの `TODO` を埋め、以下のCRUD APIを実装する

| Method | Path | 内容 |
|---|---|---|
| GET | `/tasks` | 一覧取得 |
| GET | `/tasks/:id` | 1件取得 |
| POST | `/tasks` | 作成（バリデーションあり） |
| PUT | `/tasks/:id` | 更新 |
| DELETE | `/tasks/:id` | 削除 |

---

## 4. コード解説

スケルトンでは `TaskRepository` interfaceをChapter1/2からそのまま引き継ぎ、実装だけを
`InMemoryTaskRepository` から `PostgresTaskRepository` に差し替える構成にしています。
**interfaceを介して実装を差し替えられる**というChapter1で学んだ設計が、ここで初めて実務的な価値を持ちます
（テスト時はInMemory実装、本番はPostgres実装、という使い分けが可能になります）。

---

## 5. C# / TypeScript比較

| 概念 | C#(.NET) | TypeScript(Node.js) | Go |
|---|---|---|---|
| DBアクセス | Entity Framework Core / Dapper | Prisma / TypeORM | `database/sql` + driver（素のSQL） |
| バリデーション | DataAnnotations / FluentValidation | zod / class-validator | struct tag (`binding:"..."`) |
| DTO変換 | AutoMapper | 手動 or class-transformer | 手動の変換関数 |
| エラー→HTTPステータス対応 | ExceptionFilter | Express error middleware | `errors.Is`による分岐 + 自前関数 |

C#実務者にとって一番の違和感は「ORMを使わずSQLを直接書く」ことだと思います。
Goのコミュニティでは「シンプルさを優先し、必要な分だけ抽象化する」文化が強く、
GORMのようなORMも存在しますが、**素のSQLで十分に生産性が出る**という判断をするプロジェクトが多くあります。
本教材でも、まずは素のSQLで「何が起きているか」を理解することを優先しています。

---

## 6. 実務利用例

- `sqlc` や `sqlboiler` のような「SQLからGoコードを生成するツール」を導入し、素のSQLの型安全性を高める現場もあります
- マイグレーションは `golang-migrate/migrate` のような専用ツールで管理するのが一般的です
- バリデーションエラーのレスポンス形式は、フロントエンドと事前にAPI仕様（OpenAPIなど）で合意しておくことが重要です

---

## 7. 演習

1. `tasks` テーブルをマイグレーションSQLで作成し、CRUD APIをすべて実装してください。
2. `POST /tasks` で `title` が空文字のリクエストを送り、`400 Bad Request` が返ることを確認してください。
3. `GET /tasks/:id` で存在しないIDを指定し、`404 Not Found` が統一フォーマットで返ることを確認してください。
4. （発展）`Description` のバリデーションに `binding:"max=1000"` を追加し、1001文字のリクエストで400が返ることを確認してください。

---

## 8. 完成例

完成例は分量が多いため、`sample/` ディレクトリに完全なコードを用意しています（Chapter4の最終構成に合わせた形）。
`sample/internal/repository/task_postgres.go` と `sample/internal/handler/task_handler.go` を参照してください。
この章の時点では、`main.go` に全部書いてもかまいません（Chapter4で分割する体験こそが学習の山場です）。

次章 `docs/chapter4-architecture.md` に進んでください。ここから実務的なディレクトリ構成へ段階的にリファクタリングします。
