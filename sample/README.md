# sample ― 完成版コード（Chapter4 Stage4相当）

Chapter0〜Chapter4で学んだ内容をすべて反映した、完成版のTask管理APIです。
`cmd/` + `internal/` レイアウト、Repository / Service / Handler / Middleware の分離、
DTOによる入出力の分離、統一されたエラーレスポンスを実装しています。

## ディレクトリ構成

```text
sample/
├── cmd/
│   └── api/
│       └── main.go          起動処理・DI配線・ルーティング登録
├── internal/
│   ├── config/               環境変数の読み込み
│   ├── dto/                  リクエスト/レスポンスの形（バリデーションタグ含む）
│   ├── repository/           ドメインモデル + 永続化(PostgreSQL)
│   ├── service/               ビジネスロジック
│   ├── handler/               HTTPの関心事（Ginに依存する層）
│   └── middleware/           ロギングなどの横断的関心事
├── migrations/
│   └── 001_create_tasks.sql
└── .air.toml
```

## 起動方法

このプロジェクトルート（`project/`）で `docker compose up -d` を実行し、devcontainerに入った状態で以下を実行してください。

```bash
cd sample
go mod tidy               # go.sum を生成し、依存関係を確定させる
psql -h db -U postgres -d training_db -f migrations/001_create_tasks.sql

# ホットリロードで起動する場合
air

# あるいは直接起動する場合
go run ./cmd/api
```

## 動作確認

```bash
curl -X POST localhost:8080/tasks \
  -H "Content-Type: application/json" \
  -d '{"title":"Ginを学ぶ","description":"教材を最後まで終える"}'

curl localhost:8080/tasks
curl localhost:8080/tasks/1
curl -X PUT localhost:8080/tasks/1 \
  -H "Content-Type: application/json" \
  -d '{"title":"Ginを学ぶ","description":"教材を最後まで終える","done":true}'
curl -X DELETE localhost:8080/tasks/1
curl localhost:8080/healthz
```

## この完成版を読む際のポイント

- `internal/repository/task.go`: ドメインモデルと、利用側が定義したinterface
- `internal/handler/task_handler.go`: Handler自身が必要とするservice interfaceを定義している点（Chapter1「呼び出し側がinterfaceを定義する」の実例）
- `cmd/api/main.go`: 依存関係の組み立て（DI）がここに集約されている点
