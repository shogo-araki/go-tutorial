# =========================================================
# 学習用の開発コンテナ (devcontainer) 用 Dockerfile
#
# ポイント:
# - 本番用イメージではなく「開発環境」を作るためのDockerfile
# - Air（ホットリロードツール）をここでインストールしておく
# - PostgreSQLクライアント(psql)も入れておくと動作確認に便利
# =========================================================
FROM golang:1.23-bookworm

ARG USERNAME=vscode
ARG USER_UID=1000
ARG USER_GID=$USER_UID

# 開発用ユーザーを作成（rootで作業しないため）
RUN groupadd --gid $USER_GID $USERNAME \
    && useradd --uid $USER_UID --gid $USER_GID -m $USERNAME \
    && apt-get update \
    && apt-get install -y --no-install-recommends \
        git \
        curl \
        make \
        locales \
        postgresql-client \
        ca-certificates \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* \
    # ロケールを生成する。未設定のままだとpsqlなどが内部で使うperlスクリプトが
    # 「perl: warning: Setting locale failed」という警告を出すため。
    && sed -i '/en_US.UTF-8/s/^# //g' /etc/locale.gen \
    && locale-gen

ENV LANG=en_US.UTF-8
ENV LANGUAGE=en_US:en
ENV LC_ALL=en_US.UTF-8

# Goの開発ツール類をインストール
# - air: ホットリロード
# - dlv: デバッガ
# - golangci-lint: 静的解析（実務でよく使われる）
#
# ポイント:
#   1つのRUNに&&で繋げず分割している。理由は「どのインストールが失敗したか」を
#   ビルドログから即座に特定できるようにするため（実務のDockerfileでも同様の配慮をする）
#   バージョンは@latestではなく明示的にピン留めし、
#   「ある日突然ビルドが壊れる」事態を防いでいる。
RUN go install github.com/air-verse/air@v1.61.5

RUN go install github.com/go-delve/delve/cmd/dlv@v1.23.1

# golangci-lintはインストールスクリプト(install.sh)ではなく go install を使う。
# install.shはGitHub側の仕様変更やネットワーク環境の影響で失敗しやすいため。
RUN go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6

ENV GOPATH=/go
ENV PATH=$GOPATH/bin:$PATH

USER $USERNAME

WORKDIR /workspace
