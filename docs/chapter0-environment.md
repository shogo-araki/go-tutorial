# Chapter 0: 環境構築

対象読者: C#(.NET)実務3年、TypeScript実務1年、Docker経験ありのエンジニア

---

## 1. 目的

この章のゴールは「Goのコードを1行も書く前に、実務と同じ土台を整える」ことです。

C#やTypeScriptの現場では、当たり前のように以下が揃っていたはずです。

- IDEの補完・フォーマッタ
- Dockerによる環境の再現性
- ホットリロード（`dotnet watch`、`nodemon`/`ts-node-dev` 相当）
- DBがすぐ使える状態

Goでも同じことをやります。Go特有の部分は「ビルドが速い」「バイナリ1つで完結する」という特徴くらいで、
環境構築の考え方自体はC#/TypeScriptと変わりません。

---

## 2. 解説

### 2.1 なぜdevcontainerなのか

`devcontainer.json` は VS Code の Dev Containers 拡張機能が読む設定ファイルです。
これは「このプロジェクトを開くと自動的にDockerコンテナの中で開発が始まる」という仕組みで、

- .NETでいう `.devcontainer` を使った経験があればほぼ同じ
- Node.jsプロジェクトでdevcontainerを使ったことがあれば同じ

**Go特有の注意点は1つだけ**です。GoはNode.jsの`node_modules`のような「プロジェクトローカルの依存関係フォルダ」を作らず、
`$GOPATH/pkg/mod` という共有キャッシュにモジュールを保存します。そのため `docker-compose.yml` では
`go-mod-cache` という名前付きボリュームを用意し、コンテナを再ビルドしても依存関係の再ダウンロードが走らないようにしています。

### 2.2 なぜdocker-composeでDBまで用意するのか

C#実務であれば「ローカルにSQL Server入れるのが面倒だからDockerで」という経験があるはずです。考え方は同じです。
このプロジェクトでは`app`(Goの開発コンテナ)と`db`(PostgreSQL)の2サービス構成にしています。

`depends_on` に `condition: service_healthy` を指定しているのもポイントです。
PostgreSQLは「コンテナが起動した」瞬間と「接続を受け付けられる」瞬間にタイムラグがあるため、
ヘルスチェックが通るまで `app` の起動を待たせています。これは実務のdocker-composeでも頻出のパターンです。

### 2.3 なぜAirを使うのか

Goはインタプリタ言語ではなくコンパイル言語です。ソースコードを保存しただけでは何も起きません。
`dotnet watch run` や `nodemon` のように、ファイル変更を検知して自動的に再ビルド・再起動してくれるツールが必要で、
Goの世界でそれを担うのが **Air** です。

Airは `.air.toml` という設定ファイルを見て、監視対象の拡張子やビルドコマンドを判断します。
これはChapter1でGoプロジェクトを作った後に設定します（今の時点ではまだ設定ファイルはありません）。

---

### 2.4 ポート競合について

`docker-compose.yml` では、ホストマシンからPostgreSQLに直接繋ぐ用に `5433:5432` でポートを公開しています
（コンテナ内部の`5432`番を、ホスト側の`5433`番にマッピング）。

あえて標準の5432番をホストに公開していないのは、**ホスト側で既に別のPostgreSQLが5432番を使っていて
`Bind for :::5432 failed: port is already allocated` エラーになるケースが非常に多い**ためです。
`app`コンテナから`db`コンテナへは、docker-composeが作る内部ネットワーク経由で常に`db:5432`としてアクセスするため、
ホスト側の公開ポート番号（5433）はアプリの動作には一切影響しません。

もし`up`時に別のポートで衝突エラーが出た場合は、`docker ps -a`で残っている古いコンテナがないか確認し、
`docker compose down`してから再度起動してください。

---

## 3. 実装（今の時点でやること）

1. このプロジェクトフォルダをVS Codeで開く
2. 右下に出る通知、または コマンドパレット `Dev Containers: Reopen in Container` を実行
3. `.devcontainer/Dockerfile` のビルドが走り、`app` コンテナと `db` コンテナが起動する
4. ターミナルを開き、以下を確認する

```bash
go version
air -v
psql -h db -U postgres -d training_db -c "SELECT 1;"
```

すべて実行できればOKです。

---

## 4. コード解説

`.devcontainer/Dockerfile` では、開発に必要なツール（Air, delve, golangci-lint）を
コンテナビルド時に `go install` でまとめて入れています。

`.devcontainer/post-create.sh` はコンテナ作成後に一度だけ実行されるスクリプトです。
「毎回手動でやるセットアップ作業」をコード化しておくことで、チームメンバーが増えても
同じ手順を再現できます。これは `Dockerfile` に全部書かず `post-create.sh` に分けている点がポイントで、
「イメージのビルド（不変な部分）」と「コンテナ作成後の初期化（可変な部分）」を分離する実務的な設計です。

---

## 5. C# / TypeScript比較

| 概念 | C#(.NET) | TypeScript(Node.js) | Go |
|---|---|---|---|
| ホットリロード | `dotnet watch run` | `nodemon` / `ts-node-dev` | `air` |
| 依存関係キャッシュ | NuGetキャッシュ | `node_modules`(プロジェクト単位) | `$GOPATH/pkg/mod`(グローバル共有) |
| 実行形態 | JIT/AOTコンパイル、`.dll`/ネイティブ実行ファイル | インタプリタ(V8) | 事前コンパイルされた単一バイナリ |
| ローカルDB | Docker上のSQL Server/PostgreSQL | 同左 | 同左（考え方は同じ） |

Goで最も意識が変わるのは「依存関係キャッシュがプロジェクトローカルではなくグローバル」という点です。
`node_modules` のような「プロジェクトごとに肥大化するフォルダ」はGoにはありません。

---

## 6. 実務利用例

実務のGoプロジェクトでも、この章で作った構成（devcontainer + docker-compose + Air）は
ほぼそのまま使われます。違いがあるとすれば以下の点です。

- 本番用の`Dockerfile`はマルチステージビルドで最終的に数MB〜数十MBの軽量イメージにする（Chapter4で扱います）
- CI環境では Air は使わず `go build` を直接実行する
- DBのマイグレーション管理ツール（`golang-migrate`など）を追加することが多い

---

## 7. 演習

1. `docker compose up -d` でコンテナを起動し、`docker compose ps` で `app` と `db` が起動していることを確認してください。
2. コンテナ内で `psql -h db -U postgres -d training_db` を実行し、対話的にPostgreSQLへ接続できることを確認してください。
3. `docker compose down -v` で一度環境を破棄し、再度 `docker compose up -d` してもDBデータ用ボリューム(`db-data`)以外は再構築されることを確認してください（学習用に一度壊してみる経験は重要です）。

---

## 8. 完成例

この章にコードの完成例はありません。`.devcontainer/` と `docker-compose.yml` がこのプロジェクトの完成形そのものです。
次章 `docs/chapter1-go-basics.md` に進んでください。