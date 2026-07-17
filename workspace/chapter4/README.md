# Chapter4 演習の進め方

`docs/chapter4-architecture.md` を読みながら、以下の順番で進めてください。
**写経ではなく、自分の手でファイルを作成・移動しながら進める**のがこの章の一番の学習ポイントです。

## Stage 1: step1-flat/

Chapter3相当の「全部`main.go`に書かれた」完成コードです。まずはこれを読んで、
「どこがHTTPの関心事で、どこがDBの関心事か」をコメントで分類してください（演習1）。

```bash
cd step1-flat
go mod init example.com/chapter4
go get github.com/gin-gonic/gin
go get github.com/jackc/pgx/v5
go run .
```

## Stage 2: step2-handler-repository/

`step1-flat/main.go` の内容を、自分の手で以下のように分割してください。

```text
step2-handler-repository/
├── main.go              ← 起動処理とルーティング登録のみ残す
├── handler/
│   └── task_handler.go
└── repository/
    └── task_repository.go
```

`repository`パッケージに`TaskRepository` interfaceと`PostgresTaskRepository`実装を、
`handler`パッケージに`TaskHandler`を置いてください。`handler`パッケージは`repository`パッケージを
importしてよいですが、逆（`repository`が`handler`をimport）は発生しないようにしてください
（依存の方向は「Handler → Repository」の一方通行にする）。

## Stage 3: step3-service-middleware/

Stage2の構成に、`service`パッケージと`middleware`パッケージを追加してください。

```text
step3-service-middleware/
├── main.go
├── handler/
├── service/
│   └── task_service.go   ← CompleteTaskなどのビジネスロジックをここに
├── repository/
└── middleware/
    └── logging.go         ← リクエストロギング
```

`handler`は`service`を、`service`は`repository`を呼び出す一方通行の依存にしてください。

## Stage 4: step4-cmd-internal/

Stage3の構成を、`cmd/api/main.go` + `internal/`配下の構成に組み替えてください。

```text
step4-cmd-internal/
├── cmd/
│   └── api/
│       └── main.go
└── internal/
    ├── handler/
    ├── service/
    ├── repository/
    ├── middleware/
    └── config/
        └── config.go      ← 環境変数をまとめて読み込むstruct
```

完成したら以下が通ることを確認してください。

```bash
cd step4-cmd-internal
go build ./cmd/api
```

行き詰まったら `sample/` ディレクトリの完成版コードを参照してください（ただし、まずは自力で試すことを推奨します）。
