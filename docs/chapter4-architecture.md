# Chapter 4: 実務設計 ― ディレクトリ構成を段階的にリファクタリングする

対象読者: Chapter3を完了した人

この章がこの教材の山場です。Chapter3までで書いた「`main.go`に全部乗せ」のコードを、
**実務のGoプロジェクトで一般的な構成へ段階的にリファクタリング**します。
一気に完成形を見せるのではなく、「なぜ分割するのか」を1ステップずつ体験してください。

過剰なClean Architecture（Use Case層、Entity層、Interface Adapter層…といった多層構造）は採用しません。
Goのコミュニティで好まれる**シンプルな層分けのみ**を扱います。

---

## 1. 目的

- `main.go`肥大化の問題点を体感し、責務ごとにファイルを分ける判断基準を身につける
- Repository Pattern / Service Layer の役割分担をC#実務の経験と対比して理解する
- Goの依存注入（DIコンテナなしの手動配線）の考え方を身につける
- `cmd/` `internal/` というGoコミュニティ標準のレイアウトを理解する

---

## 2. 解説 ― 4つの段階

### Stage 1: `main.go` に全部書く（Chapter3の状態）

```text
main.go   ← struct定義、DBアクセス、Handler、ルーティング、起動処理が全部ここ
```

これは決して「悪いコード」ではありません。**小さいアプリケーションであればこれで十分**というのがGoの価値観です。
C#で最初から `Controllers/` `Services/` `Repositories/` とレイヤーを切るのが当然とされる文化とは対照的に、
Goでは「必要になるまで分割しない（YAGNI）」という考え方が強く支持されています。

ただし、エンドポイントが増え、テストを書き始めると、以下の問題が顕在化します。

- HTTPの関心事（Handler）とDBの関心事（永続化）が同じファイルに混在し、見通しが悪くなる
- Handlerのテストをしたいだけなのに、DB接続まで用意しないとテストできない
- 複数人で開発する際、同じファイルへの変更が衝突しやすくなる

これが「分割する動機」です。**動機がない段階で分割しない**ことも、同じくらい重要な設計判断です。

### Stage 2: `handler/` と `repository/` に分離する

```text
main.go
handler/
  task_handler.go
repository/
  task_repository.go
```

最初に分離するのはこの2つです。理由は「関心事の性質が明確に違う」からです。

- `handler`: HTTPリクエスト/レスポンスの変換だけを担当する（Ginに依存する層）
- `repository`: DBとのやり取りだけを担当する（`database/sql`に依存する層）

```go
// repository/task_repository.go
package repository

type TaskRepository interface {
    FindAll(ctx context.Context) ([]Task, error)
    FindByID(ctx context.Context, id int) (Task, error)
    // ...
}
```

```go
// handler/task_handler.go
package handler

import "example.com/chapter4/repository"

type TaskHandler struct {
    repo repository.TaskRepository
}
```

ここで**Goのpackage設計の基本原則**が登場します。「フォルダ＝package」というルール（Chapter1参照）により、
`handler`パッケージと`repository`パッケージは物理的に独立したコード単位になります。
C#で`namespace`を分けるのとの決定的な違いは、**Goではpackageをまたぐと公開(exported)/非公開(unexported)の境界が明確になる**点です。

```go
type taskHandler struct { ... }  // 小文字始まり = 同じpackage内でしか使えない（非公開）
type TaskHandler struct { ... }  // 大文字始まり = 他packageからimportして使える（公開）
```

これはC#の `public`/`internal`/`private` に相当しますが、**アクセス修飾子キーワードが存在せず、
命名（大文字/小文字）だけでアクセス制御を表現する**のがGoの特徴です。実務では「本当に外部に公開すべきものだけ大文字にする」
ことが、良いpackage設計の第一歩です。

### Stage 3: `service/` と `middleware/` を追加する

```text
main.go
handler/
  task_handler.go
service/
  task_service.go
repository/
  task_repository.go
middleware/
  logging.go
  recovery.go
```

Handlerが太ってきたら（バリデーション以外のビジネスロジックが増えてきたら）、`service`層を導入します。

```go
// service/task_service.go
package service

type TaskService struct {
    repo repository.TaskRepository
}

func (s *TaskService) CompleteTask(ctx context.Context, id int) (Task, error) {
    task, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return Task{}, err
    }
    task.Done = true
    task.UpdatedAt = time.Now()
    return s.repo.Update(ctx, task)
}
```

**C#実務での`Service`層とほぼ同じ役割**です。「Handlerは薄く、Serviceにビジネスロジックを集める」という方針もASP.NET Coreと同じです。
ただし、Goでは**シンプルなCRUDだけのエンドポイントに無理にServiceを作らない**ことが推奨されます。
「HandlerがRepositoryを直接呼ぶだけで済むなら、Serviceを挟む必要はない」という判断も、Goらしいシンプルさの一部です。

Middlewareは、ASP.NET Coreの `Middleware`、Express.jsの `app.use(...)` と同じ役割です。

```go
// middleware/logging.go
package middleware

func Logging() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        c.Next() // 次のHandler/Middlewareへ処理を渡す
        log.Printf("%s %s %d %v", c.Request.Method, c.Request.URL.Path, c.Writer.Status(), time.Since(start))
    }
}

// main.go
r.Use(middleware.Logging(), gin.Recovery())
```

`c.Next()` はASP.NET Coreの `await next(context)`、Express.jsの `next()` に相当します。
「前処理 → `c.Next()` → 後処理」という構造で、リクエストのライフサイクルに割り込めます。

### Stage 4: `cmd/` と `internal/` ― Goコミュニティ標準のレイアウト

最終形として、以下のレイアウトへ移行します。

```text
cmd/
  api/
    main.go              ← 起動処理のみ。数十行程度に薄くする
internal/
  handler/
    task_handler.go
  service/
    task_service.go
  repository/
    task_repository.go
  middleware/
    logging.go
    recovery.go
  config/
    config.go            ← 環境変数の読み込み
```

2つのポイントがあります。

**1. `cmd/` は「エントリーポイントの置き場所」**

C#でいう`Program.cs`、Node.jsでいう`index.ts`にあたるものですが、Goでは`cmd/<バイナリ名>/main.go`という形で
**複数の実行可能ファイルを1つのモジュールから作れる**構成が一般的です。例えば将来的に
`cmd/api/main.go`（APIサーバー）と `cmd/migrate/main.go`（マイグレーション実行コマンド）を同じリポジトリで管理する、
といったことがよくあります。

**2. `internal/` は「このモジュールの外からimportできない」という強制力を持つ特別なフォルダ名**

これはGoツールチェインが特別扱いする唯一のディレクトリ名です。`internal/`配下のpackageは、
**`internal`の親ディレクトリ配下のコードからしかimportできません**。他のモジュール（別リポジトリ）から
`import "example.com/chapter4/internal/repository"` としても、コンパイルエラーになります。

これはC#の`internal`アクセス修飾子（アセンブリ内限定）と考え方は似ていますが、Goでは**ディレクトリ構造そのものが
アクセス制御を表現する**という点が異なります。「このモジュールの外に公開したいコードだけを`internal/`の外に置く」
という設計判断を、フォルダ構成を見ただけで読み取れるようにする効果があります。

---

## 3. 実装

`workspace/chapter4/` にStage1〜Stage4の演習用ディレクトリを用意しています。

```text
workspace/chapter4/
├── README.md              ← 各ステージの進め方
├── step1-flat/            ← Stage1: Chapter3相当の完成コード（ここから始める）
├── step2-handler-repository/ ← Stage2: 自分でhandler/repositoryに分割する
├── step3-service-middleware/ ← Stage3: 自分でservice/middlewareを追加する
└── step4-cmd-internal/       ← Stage4: 自分でcmd/internal構成に組み替える
```

`step1-flat/` から始めて、各ステージのREADMEの指示に従って**自分の手でファイルを移動・分割**してください。
最終的に `step4-cmd-internal/` が `sample/` と同じ構成になれば完成です。

---

## 4. コード解説

段階を経るごとに `main.go` がどう変化するかに注目してください。

- Stage1: `main.go` が200行前後（すべてがここにある）
- Stage4: `cmd/api/main.go` が50行未満（依存関係を組み立てて起動するだけ）

```go
// cmd/api/main.go (Stage4のイメージ)
func main() {
    cfg := config.Load()
    db := mustConnectDB(cfg)
    defer db.Close()

    taskRepo := repository.NewPostgresTaskRepository(db)
    taskService := service.NewTaskService(taskRepo)
    taskHandler := handler.NewTaskHandler(taskService)

    r := gin.Default()
    r.Use(middleware.Logging(), gin.Recovery())
    r.GET("/tasks", taskHandler.List)
    // ...
    r.Run(":" + cfg.Port)
}
```

**依存関係の組み立て（Dependency Injection）が`main.go`に集約されている**点が重要です。
C#の`Program.cs`で`builder.Services.AddScoped<ITaskRepository, TaskRepository>()`のように
DIコンテナへ登録する代わりに、Goでは`main.go`で**手続き的に**`New〇〇(...)`を呼んで組み立てます。
コンテナのような「魔法」がない分、コードを読むだけで依存関係の全体像を追えるのがGoの利点です。

---

## 5. C# / TypeScript比較

| 概念 | ASP.NET Core | NestJS(TypeScript) | Go(このプロジェクト) |
|---|---|---|---|
| エントリーポイント | `Program.cs` | `main.ts` | `cmd/api/main.go` |
| DI | `IServiceCollection`（コンテナ） | `@Injectable` + モジュールシステム | `main.go`で手動配線 |
| レイヤー分割 | Controller/Service/Repository | Controller/Service/Repository | handler/service/repository |
| アクセス制御 | `public`/`internal`/`private` | `public`/`private`（TSの型のみ、実行時は無効） | 命名（大文字/小文字）+ `internal/`ディレクトリ |
| 複数バイナリ | 複数`.csproj` | 通常はモノレポ+ツールで対応 | `cmd/`配下に複数の`main.go` |

DIコンテナがない点を「不便」と感じるかもしれませんが、実務のGoプロジェクトの多くは
**数百〜数千行規模でもDIコンテナなしで十分に見通しよく保てます**。依存が複雑化しすぎた場合のみ、
`google/wire`のようなコード生成ベースのDIツールを検討する、という順序が実務的です。

---

## 6. 実務利用例

- `internal/`配下のパッケージ構成は、チームやプロジェクトの規模に応じて`handler`をさらに`task_handler.go`
  `user_handler.go`のようにドメイン単位で分けるのが一般的です
- `config/`パッケージでは`os.Getenv`を直接使わず、`github.com/caarlos0/env`のようなライブラリで
  構造体に環境変数をマッピングする現場も多くあります
- 本番用の`Dockerfile`はマルチステージビルドにし、`cmd/api`だけをビルドした軽量イメージを作るのが定石です

```dockerfile
# 本番用Dockerfileのイメージ（マルチステージビルド）
FROM golang:1.23 AS builder
WORKDIR /src
COPY . .
RUN go build -o /app ./cmd/api

FROM gcr.io/distroless/base-debian12
COPY --from=builder /app /app
ENTRYPOINT ["/app"]
```

---

## 7. 演習

1. `step1-flat/` のコードを読み、どの部分が「HTTPの関心事」で、どの部分が「DBの関心事」かをコメントで分類してください。
2. `step2-handler-repository/` で、`handler`パッケージと`repository`パッケージに分割してください。`TaskRepository`をinterfaceとして`repository`パッケージに定義し、`handler`パッケージからはinterface経由でのみ利用してください。
3. `step3-service-middleware/` で、`CompleteTask`（タスクを完了状態にする）のロジックを`service`層に切り出し、リクエストロギング用の`middleware`を追加してください。
4. `step4-cmd-internal/` で、`cmd/api/main.go`と`internal/`配下のpackage構成に組み替え、`go build ./cmd/api`が通ることを確認してください。
5. （発展）`internal/`配下のpackageを、モジュール外の別ディレクトリから`import`しようとしてコンパイルエラーになることを実際に確認してください。

---

## 8. 完成例

Stage4の完成形は `sample/` ディレクトリに完全なコードとして用意しています。

```bash
cd sample
go run ./cmd/api
```

`sample/README.md`（このプロジェクトのルートREADME.mdからもリンクしています）に起動手順を記載しています。
`sample/internal/` 配下の各パッケージを実際に読み、この章で説明した設計がどう反映されているかを確認してください。

これで教材は完了です。お疲れさまでした。次のステップとして、Testing（`_test.go`によるユニットテスト）や
OpenAPIによるAPI仕様書作成など、この教材で扱わなかったテーマにもぜひ挑戦してみてください。
