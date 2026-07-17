// Chapter1 演習スケルトン
//
// docs/chapter1-go-basics.md の「3. 実装」「7. 演習」を参照しながら
// 以下のTODOを埋めてください。写経ではなく、自分で考えて実装することが目的です。
//
// 実行方法:
//   go run .
package main

import (
	"fmt"
)

// TODO(1): Task struct を定義してください。
// フィールド: ID int, Title string, Done bool
// type Task struct {
// }

// TODO(2): Task に MarkDone() メソッドをポインタレシーバで実装してください。
// (状態を変更するメソッドはポインタレシーバにする、という設計思想を思い出してください)
//
// func (t *Task) MarkDone() {
// }

// TODO(3): TaskRepository interface を定義してください。
// 必要なメソッド: FindByID(id int) (Task, error) / FindAll() []Task / Save(task Task) error
//
// type TaskRepository interface {
// }

// TODO(4): InMemoryTaskRepository struct を定義し、上のinterfaceを満たす実装を書いてください。
// フィールドとして map[int]Task を持たせるとよいでしょう。
//
// type InMemoryTaskRepository struct {
// }
//
// func NewInMemoryTaskRepository() *InMemoryTaskRepository {
// 	return &InMemoryTaskRepository{}
// }

func main() {
	fmt.Println("Chapter1: TODOを実装してこのメッセージを書き換えてください")

	// TODO(5): 実装したTaskRepositoryを使って、
	//   - Taskを2件保存
	//   - 1件取得してMarkDone()
	//   - 保存し直して結果を表示
	//   - 存在しないIDで検索してerrorが返ることを確認
	// という一連の流れを実装してください。
}
