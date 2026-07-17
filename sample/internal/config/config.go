// Package config は環境変数の読み込みを1箇所にまとめる。
//
// なぜ集約するのか:
//   os.Getenv をコードのあちこちで直接呼ぶと、「どんな環境変数が必要か」が
//   コード全体に散らばってしまう。config パッケージに集約することで、
//   「このアプリが必要とする設定値の一覧」がこのファイル1つでわかるようにする。
//   C#でいう appsettings.json + IOptions<T> の役割に近い。
package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port       string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
}

// Load は環境変数からConfigを組み立てる。
// 値が未設定の場合はデフォルト値を使う（開発時の利便性のため）。
func Load() Config {
	return Config{
		Port:       getEnv("PORT", "8080"),
		DBHost:     getEnv("DB_HOST", "db"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "training_db"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// DSN はPostgreSQLへの接続文字列を組み立てる。
func (c Config) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName,
	)
}
