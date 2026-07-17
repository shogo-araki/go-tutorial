# Chapter 1: Go基礎（最低限）

対象読者: C#(.NET)実務3年、TypeScript実務1年

この章は「Go入門」ではありません。**Ginを使い始めるために必要な最低限の文法と、Go特有の設計思想**だけを扱います。
`for`文の書き方のような基礎文法の説明は省略します（他言語経験者なら見ればわかるため）。

---

## 1. 目的

- C#/TypeScriptの知識をベースに、Goの型システムと設計思想の"違い"を理解する
- struct / method / interface / error / package という、Chapter2以降で毎回使う5つの概念を体に入れる
- 「Goはなぜこの書き方をするのか」を理解し、Chapter2でGinのコードを読んだときに迷わないようにする

---

## 2. 解説

### 2.1 Go Modules ― パッケージ管理の単位

C#のプロジェクトファイル(`.csproj`)、Node.jsの`package.json`に相当するのが `go.mod` です。

```bash
go mod init example.com/chapter1
```

これで `go.mod` が生成されます。重要なのは、**Goには「プロジェクト」という単位がなく「モジュール」という単位しかない**という点です。
1つの `go.mod` ＝ 1つの依存関係の管理単位、という考え方はNode.jsの`package.json`に近いですが、
Goでは**モジュールパス自体がimportパスとして使われる**という違いがあります。

```go
// go.mod の module名が example.com/chapter1 なら
import "example.com/chapter1/internal/task"
// のように、自分のプロジェクト内のパッケージも「フルパスで」importする
```

C#の名前空間(`namespace`)やTypeScriptの相対import(`../models/task`)と違い、
**Goのimportは常にモジュールルートからの絶対パス**です。慣れないうちは冗長に感じますが、
「このファイルがどこにあるパッケージのコードを使っているか」が一目でわかるという利点があります。

### 2.2 package ― Goにおける「名前空間」の正体

C#では `namespace` はファイルの中の宣言に過ぎず、フォルダ構成と一致させなくても動きます。
TypeScriptではファイル＝モジュールで、名前空間という概念自体をほぼ使いません。

Goはその中間で、**「1つのフォルダ＝1つのpackage」という強いルール**があります。

```text
task/
├── task.go       // package task
└── repository.go // package task  ← 同じフォルダは同じpackage名でなければならない
```

同じフォルダ内の複数ファイルは自動的に同じpackage名前空間を共有します（`import`不要で相互に参照できる）。
これはC#の「同じ`namespace`内は`using`不要」に近い感覚ですが、Goでは**フォルダ構成がそのままpackage構成になる**という点が異なります。
「どこにファイルを置くか」が「どの名前空間に属するか」を直接決めるため、実務では**フォルダ設計＝package設計**として扱われます（Chapter4で詳しく扱います）。

### 2.3 struct ― クラスではなく「データの型」

Goには `class` がありません。データの形を定義するのは `struct` です。

```go
type Task struct {
    ID          int
    Title       string
    Description string
    Done        bool
}
```

C#の `class` や `record`、TypeScriptの `interface`/`type` に近いですが、決定的に違うのは
**structには継承がない**という点です。Goには `class A : B` のような継承構文が存在しません。

代わりにGoでは「**Composition（合成）**」を使います。

```go
type Base struct {
    ID        int
    CreatedAt time.Time
}

type Task struct {
    Base        // 埋め込み(embedding)。継承ではなく「フィールドとして持つ」
    Title string
}

t := Task{}
t.ID // Base.IDに直接アクセスできる（継承しているように見えるが、実体は合成）
```

これは「継承より合成を優先せよ」というオブジェクト指向設計の原則を、**言語仕様として強制している**と理解すると腹落ちします。
C#で `abstract class` を使って共通処理をまとめていた設計は、Goでは素直に移植できません。後述のinterfaceと組み合わせて考え直す必要があります。

### 2.4 method ― structに"ぶら下げる"関数

Goのmethodは、C#の「クラスのメンバーメソッド」と似ていますが、構文的には**通常の関数にレシーバを付けただけ**です。

```go
func (t Task) IsOverdue(now time.Time) bool {
    return !t.Done && now.After(t.Deadline)
}

func (t *Task) MarkDone() {
    t.Done = true
}
```

ここで最重要なのが **値レシーバ `(t Task)` とポインタレシーバ `(t *Task)` の違い**です。

- 値レシーバ: `t` はコピー。メソッド内で `t` を書き換えても呼び出し元には反映されない
- ポインタレシーバ: `t` は元データへの参照。書き換えが呼び出し元に反映される

C#でいうと、`struct`(値型)のメソッドと`class`(参照型)のメソッドの違いに近いです。
Goの `struct` はデフォルトが値型なので、**状態を変更するメソッドは基本的にポインタレシーバで書く**のが実務での定石です。
「読み取り専用ならどちらでもいいが、変更するなら`*T`」というルールを覚えておけば十分です。

### 2.5 pointer ― 参照ではなく「アドレス」

TypeScriptにはポインタの概念がなく、C#では`ref`/`out`くらいでしか意識しません。Goでは日常的に使います。

```go
func UpdateTitle(t *Task, title string) {
    t.Title = title // ポインタ経由でTaskの中身を書き換える
}

task := Task{Title: "old"}
UpdateTitle(&task, "new")
// task.Title は "new" になっている
```

Goの関数は**デフォルトで値渡し**です。構造体をそのまま渡すとコピーが発生するため、
「大きなstructを引数に渡す」「関数内で状態を書き換えたい」場合は `*Task` のようにポインタを渡します。
C#の `class` がデフォルトで参照型なのとは逆の発想である点に注意してください。Goでは**参照型か値型かを自分で選ぶ**のです。

### 2.6 slice / map ― コレクションの基本

| 用途 | C# | TypeScript | Go |
|---|---|---|---|
| 可変長配列 | `List<T>` | `Array` | `[]T` (slice) |
| キー付きコレクション | `Dictionary<K,V>` | `Map`/`Record` | `map[K]V` |

```go
tasks := []Task{}                 // List<Task> 相当
tasks = append(tasks, Task{ID: 1})

byID := map[int]Task{}            // Dictionary<int, Task> 相当
byID[1] = Task{ID: 1}
value, ok := byID[1]              // TryGetValueに相当する「ok pattern」
```

`value, ok := byID[1]` の書き方はGo特有です。C#の `TryGetValue` 、TypeScriptの `map.has()` に相当する処理を、
**戻り値2つ受け取るだけ**で表現できます。この「複数戻り値」パターンはこの後の`error`処理でも中心的な役割を果たします。

### 2.7 interface ― 「実装」ではなく「形」による契約

ここがC#/TypeScript経験者が最も驚くポイントです。GoのinterfaceはC#の `interface` のように
**明示的な実装宣言(`class Foo : IBar`)が不要**です。

```go
type TaskRepository interface {
    FindByID(id int) (Task, error)
    Save(task Task) error
}

type InMemoryTaskRepository struct {
    data map[int]Task
}

func (r *InMemoryTaskRepository) FindByID(id int) (Task, error) { /* ... */ }
func (r *InMemoryTaskRepository) Save(task Task) error          { /* ... */ }

// InMemoryTaskRepository は「implements TaskRepository」と一言も書いていないが、
// 同じシグネチャのメソッドを持っているだけで自動的にTaskRepositoryとして扱える
var repo TaskRepository = &InMemoryTaskRepository{}
```

これは **構造的型付け（Structural Typing）** と呼ばれる考え方で、TypeScriptの `interface` の考え方に近いです
（TypeScriptもダックタイピング的にinterfaceを満たせば良い）。C#の「明示的にinterfaceを実装宣言する」文化とは根本的に異なります。

実務での使いどころは以下の通りです。

- **呼び出し側（利用する側）がinterfaceを定義する**のがGoの流儀です（C#では提供側が定義しがち）
- 「小さいinterfaceを、必要な場所ごとに定義する」のがGoらしい設計（後述の`package設計`で詳しく扱う）
- テスト時にモック実装を差し込みやすくするために使う（DIコンテナなしで実現できる）

### 2.8 error ― 例外ではなく「戻り値」

C#/TypeScriptの `try/catch` に相当するものはGoには基本的にありません（`panic/recover`はあるが例外機構としては使わない）。
Goでは**エラーは戻り値として明示的に返す**のが基本です。

```go
func FindTask(id int) (Task, error) {
    task, ok := store[id]
    if !ok {
        return Task{}, fmt.Errorf("task not found: id=%d", id)
    }
    return task, nil
}

task, err := FindTask(1)
if err != nil {
    // ここでエラーハンドリング。呼び出し元は絶対にerrをチェックする必要がある
    log.Println(err)
    return
}
```

最初は冗長に感じますが、これには明確な設計思想があります。

- **どの関数がエラーを返しうるか、シグネチャを見ただけでわかる**（C#の例外は宣言に現れないため、ドキュメントを読まないとわからない）
- **エラーハンドリングを"握りつぶす"ことが構文上しにくい**（`if err != nil` を書き忘れると`err`が未使用変数としてコンパイルエラーになりやすい）
- 呼び出し側で「エラーを上位に伝播するか」「ここで対処するか」を毎回明示的に選択させられる

C#の `try/catch` に慣れていると「なぜ毎回`if err != nil`を書くのか」と感じますが、
これは「例外的なケースかどうかの判断」をコンパイラではなく人間の目に強制する、Goの設計思想そのものです。

---

## 3. 実装（演習用スケルトン）

`workspace/chapter1/` にスケルトンコードを用意しています。`TODO` コメントの箇所を自分で実装してください。
題材は「タスク管理」で、Chapter2以降でもそのままGin APIの題材として発展させます。

```bash
cd workspace/chapter1
go run .
```

実装すべき内容（`main.go` 内の `TODO` を参照）:

1. `Task` structを定義する（ID, Title, Done）
2. `TaskRepository` interfaceを定義する（`FindByID`, `Save`, `FindAll`）
3. `InMemoryTaskRepository` を実装し、上記interfaceを満たす
4. `MarkDone` メソッドをポインタレシーバで実装する
5. 存在しないIDを検索したときに `error` を返す

---

## 4. コード解説

`workspace/chapter1/main.go` のスケルトンでは、意図的に以下の順番でコメントを配置しています。

```go
// 1. まずデータの形を決める → struct
// 2. 次にそのデータに対する操作を決める → method
// 3. 「操作の集合」を抽象化する → interface
// 4. 失敗しうる操作は必ずerrorを返す → error
```

これはGoにおける典型的な設計の順序であり、C#のように「まずinterfaceを設計してからDIコンテナに登録する」という
トップダウンの設計とは順序が逆になりがちです。Goでは**具体的な実装を先に書き、必要になった時点でinterfaceを抽出する**
（"Accept interfaces, return structs" という格言があります）というボトムアップな流儀が好まれます。

---

## 5. C# / TypeScript比較

| 概念 | C# | TypeScript | Go |
|---|---|---|---|
| データ定義 | `class` / `record` | `interface` / `type` | `struct` |
| 継承 | `class A : B` | `extends` | なし（`Composition`のみ） |
| interfaceの実装 | 明示的宣言必須 | 構造的型付け（Goに近い） | 構造的型付け（宣言不要） |
| null安全 | `Nullable<T>` / `?` | `strictNullChecks` | ゼロ値（`nil`はポインタ/interface/slice/mapのみ） |
| エラー処理 | `try/catch`（例外） | `try/catch`（例外） | 戻り値としての`error` |
| パッケージ管理 | NuGet + `.csproj` | npm + `package.json` | Go Modules + `go.mod` |

especially重要なのは **null安全の考え方**です。GoにはC#の `Nullable<T>` のような仕組みはなく、
**すべての型に「ゼロ値」が存在する**という設計です（`int`のゼロ値は`0`、`string`は`""`、`bool`は`false`）。
`nil`になりうるのは pointer・interface・slice・map・channel・func の6種類のみで、
「値が未設定である」ことと「値が存在しない」ことをC#ほど明確には区別しません。この違いはstructのゼロ値運用（Chapter3のDTO設計）で重要になります。

---

## 6. 実務利用例

実務のGoプロジェクトでは、この章の内容がそのまま以下の場面で使われます。

- **struct**: DBのテーブル1行、APIのリクエスト/レスポンスの形、すべてstructで表現する
- **interface**: `Repository`層と`Service`層を疎結合にするための境界として使う（Chapter4で本格的に扱う）
- **error**: `errors.Is` / `errors.As` を使ったエラー種別判定、HTTPステータスコードへのマッピング（Chapter3で扱う）

---

## 7. 演習

1. `workspace/chapter1/main.go` の `TODO` をすべて実装し、`go run .` でエラーなく動くようにしてください。
2. `Task` に `Deadline time.Time` フィールドを追加し、`IsOverdue(now time.Time) bool` メソッド（値レシーバ）を実装してください。
3. `FindByID` で存在しないIDを渡した際、`fmt.Errorf` で意味のあるエラーメッセージを返すようにしてください。
4. （発展）`TaskRepository` interfaceを満たす2つ目の実装（例: `LoggingTaskRepository`。呼び出しをログ出力してから内部の実装に委譲する）を作り、
   「interfaceさえ満たせば実装を差し替えられる」ことを体験してください。

---

## 8. 完成例

<details>
<summary>クリックして完成例コードを表示</summary>

```go
package main

import (
	"errors"
	"fmt"
)

// 1. データの形
type Task struct {
	ID    int
	Title string
	Done  bool
}

// 2. データに対する操作（ポインタレシーバ = 状態を変更する）
func (t *Task) MarkDone() {
	t.Done = true
}

// 3. 操作の集合を抽象化
type TaskRepository interface {
	FindByID(id int) (Task, error)
	FindAll() []Task
	Save(task Task) error
}

// 4. interfaceを満たす具体的な実装
type InMemoryTaskRepository struct {
	data map[int]Task
}

func NewInMemoryTaskRepository() *InMemoryTaskRepository {
	return &InMemoryTaskRepository{data: map[int]Task{}}
}

func (r *InMemoryTaskRepository) FindByID(id int) (Task, error) {
	task, ok := r.data[id]
	if !ok {
		return Task{}, fmt.Errorf("task not found: id=%d", id)
	}
	return task, nil
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
		return errors.New("task ID must not be zero")
	}
	r.data[task.ID] = task
	return nil
}

func main() {
	var repo TaskRepository = NewInMemoryTaskRepository()

	_ = repo.Save(Task{ID: 1, Title: "Goの基礎を学ぶ"})
	_ = repo.Save(Task{ID: 2, Title: "Ginを導入する"})

	task, err := repo.FindByID(1)
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	task.MarkDone()
	_ = repo.Save(task)

	updated, _ := repo.FindByID(1)
	fmt.Printf("Task#%d %s done=%v\n", updated.ID, updated.Title, updated.Done)

	_, err = repo.FindByID(999)
	fmt.Println("expected error:", err)
}
```

</details>

次章 `docs/chapter2-gin.md` に進んでください。ここからGinを導入します。
