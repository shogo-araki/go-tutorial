# Go × Gin Web API 教材

3~5年目のエンジニアが、**Go未経験の状態からGinを使った実務レベルのWeb APIを作れるようになる**ための教材です。

この教材はGinの使い方を覚えることを目的としていません。**Ginを使いながらGoの実務的な書き方・設計思想を学ぶこと**が目的です。
struct / method / interface / error handling / package設計 / dependency管理 / Goらしいシンプルな設計 を重点的に扱います。

---

## ディレクトリ構成

```text
project/
├── .devcontainer/     開発環境定義 (devcontainer.json, Dockerfile, post-create.sh)
├── docker-compose.yml Go + PostgreSQL の開発環境定義
├── README.md          このファイル
├── docs/              教材本体（Chapter0〜4）
├── workspace/         学習用コード（TODOを自分で実装する演習スケルトン）
└── sample/            完成版コード（Chapter4 Stage4相当の最終形）
```

---

## 起動方法

### 1. devcontainerで開く

VS Codeでこのフォルダを開き、コマンドパレットから `Dev Containers: Reopen in Container` を実行してください。
`.devcontainer/Dockerfile` のビルドが行われ、Go・Air・PostgreSQLクライアントなどが揃った開発コンテナが起動します。

devcontainerを使わない場合は、直接 `docker compose up -d` でも環境は起動できますが、
VS CodeのGo拡張機能や補完を使うには devcontainer 経由での起動を推奨します。

### 2. 起動確認

```bash
go version
air -v
psql -h db -U postgres -d training_db -c "SELECT 1;"
```

すべて実行できれば準備完了です。

---

## 学習順序

`docs/` 配下のドキュメントを、以下の順番で読み進めてください。各章は「目的 → 解説 → 実装 → コード解説 → C#/TypeScript比較 → 実務利用例 → 演習 → 完成例」の構成になっています。

| 章 | ドキュメント | 内容 | 対応する演習コード |
|---|---|---|---|
| Chapter0 | [docs/chapter0-environment.md](docs/chapter0-environment.md) | devcontainer / Docker / PostgreSQL / Air | - |
| Chapter1 | [docs/chapter1-go-basics.md](docs/chapter1-go-basics.md) | Go Modules / struct / method / interface / error（最低限） | `workspace/chapter1/` |
| Chapter2 | [docs/chapter2-gin.md](docs/chapter2-gin.md) | Ginの導入、Routing / Handler / Context / JSON | `workspace/chapter2/` |
| Chapter3 | [docs/chapter3-api-development.md](docs/chapter3-api-development.md) | Validation / DTO / PostgreSQL / CRUD / Error Response | `workspace/chapter3/` |
| Chapter4 | [docs/chapter4-architecture.md](docs/chapter4-architecture.md) | 実務的なディレクトリ構成への段階的リファクタリング | `workspace/chapter4/` → `sample/` |

**進め方の原則**: 写経ではなく、各章の演習を自分の手で実装してください。行き詰まったときだけ、各ドキュメント末尾の「8. 完成例」や `sample/` を確認してください。

---

## workspace/ の使い方

```text
workspace/
├── chapter1/    Go基礎の演習（Ginなし）
├── chapter2/    Gin導入の演習
├── chapter3/    PostgreSQL連携・CRUD APIの演習
└── chapter4/    ディレクトリ構成リファクタリングの演習（Stage1〜4）
```

各章のディレクトリには、`TODO` コメント付きのスケルトンコードが入っています。
`docs/` の対応する章を読みながら実装してください。

## sample/ の使い方

`sample/` はChapter4のStage4（`cmd/` + `internal/` レイアウト）に到達した完成版のコードです。
自分の実装がうまくいかない場合の参考、またはChapter4完了後の「模範解答との比較」に使ってください。

```bash
cd sample
go mod tidy
psql -h db -U postgres -d training_db -f migrations/001_create_tasks.sql
air
```

詳細は [sample/README.md](sample/README.md) を参照してください。

---

## この教材で扱っていないこと（発展学習として推奨）

意図的にスコープ外としている内容です。この教材を終えた後の次のステップとして取り組んでください。

- `_test.go` によるユニットテスト・テーブル駆動テスト
- OpenAPI(Swagger)によるAPI仕様書の自動生成
- 認証・認可（JWTなど）
- `golang-migrate` などによるマイグレーションのバージョン管理
- 本番用マルチステージビルドDockerfileとCI/CD

過剰なClean Architecture（Use Case層、Entity層などの多層構造）はGoらしいシンプルさを優先するため、意図的に採用していません。
