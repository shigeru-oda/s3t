# Go スタイルガイド

このドキュメントは以下の公式ガイドラインに基づいています：
- [Effective Go](https://go.dev/doc/effective_go)
- [Go Wiki: Code Review Comments](https://go.dev/wiki/CodeReviewComments)
- [Google Go Style Guide](https://google.github.io/styleguide/go/)

## 基本原則

### 優先順位（重要度順）
1. **明確性 (Clarity)**: コードの目的と理由が読み手に明確
2. **簡潔性 (Simplicity)**: 最もシンプルな方法で目標を達成
3. **簡明性 (Concision)**: 高いシグナル対ノイズ比
4. **保守性 (Maintainability)**: 容易にメンテナンス可能
5. **一貫性 (Consistency)**: コードベース全体で統一

## フォーマット

### gofmt
- すべての Go コードは `gofmt` の出力に従う
- インデントにはタブを使用
- 行の長さに固定制限はないが、長すぎる場合はリファクタリングを検討

### MixedCaps
- 複数語の名前には `MixedCaps` または `mixedCaps` を使用
- アンダースコア（snake_case）は使用しない
- 例: `MaxLength`（エクスポート）、`maxLength`（非エクスポート）

## 命名規則

### パッケージ名
- 小文字、単一語、短く簡潔に
- `util`, `common`, `misc`, `api`, `types` などの無意味な名前は避ける
- パッケージ名と型名の重複を避ける（`chubby.ChubbyFile` → `chubby.File`）

### 変数名
- 短い名前を好む（特にスコープが限定されている場合）
- ループインデックス: `i`, `j`, `k`
- リーダー: `r`
- ライター: `w`
- 宣言から遠い場所で使用される変数ほど説明的な名前が必要

### レシーバー名
- 型の1-2文字の省略形を使用（例: `Client` → `c` または `cl`）
- `me`, `this`, `self` は使用しない
- 同じ型のすべてのメソッドで一貫した名前を使用

### 頭字語
- 一貫した大文字/小文字を維持: `URL` または `url`（`Url` は不可）
- `ID`（識別子）も同様: `appID`（`appId` は不可）
- 例: `ServeHTTP`, `XMLHTTPRequest`

### Getter/Setter
- Getter に `Get` プレフィックスは不要: `Owner()`（`GetOwner()` は不可）
- Setter には `Set` プレフィックスを使用: `SetOwner()`

### インターフェース名
- 単一メソッドのインターフェースは `-er` サフィックス: `Reader`, `Writer`, `Formatter`

## エラーハンドリング

### 基本ルール
- エラーは必ずチェックする（`_` で無視しない）
- エラーは呼び出し元に返すか、適切に処理する
- 通常のエラー処理に `panic` は使用しない

### エラー文字列
- 小文字で開始（固有名詞・頭字語を除く）
- 句読点で終わらない
- 例: `fmt.Errorf("something bad")`（`"Something bad."` は不可）

### エラーフロー
- 正常パスを最小インデントに保つ
- エラー処理を先に行い、早期リターン

```go
// Good
if err != nil {
    return err
}
// normal code

// Bad
if err != nil {
    // error handling
} else {
    // normal code
}
```

### エラーラッピング
- `fmt.Errorf` と `%w` でエラーをラップ
- 冗長な情報は追加しない（基底エラーが既に提供している情報）
- `%w` は文字列の末尾に配置: `fmt.Errorf("failed to process: %w", err)`

## コメント

### Doc コメント
- エクスポートされた名前には必ず Doc コメントを付ける
- 説明対象の名前で開始し、ピリオドで終了
- 完全な文で記述

```go
// Request represents a request to run a command.
type Request struct { ... }

// Encode writes the JSON encoding of req to w.
func Encode(w io.Writer, req *Request) { ... }
```

### パッケージコメント
- `package` 句の直前に配置（空行なし）
- `// Package xxx ...` の形式

## 変数宣言

### ゼロ値
- ゼロ値で初期化する場合は `var` を使用
```go
var s []string  // nil スライス（推奨）
s := []string{} // 非 nil だが長さゼロ（JSON エンコード時など特定の場合のみ）
```

### 非ゼロ値
- 非ゼロ値で初期化する場合は `:=` を使用
```go
i := 42
s := "hello"
```

## Context

- 関数の最初のパラメータとして受け取る: `func F(ctx context.Context, ...)`
- 構造体のメンバーにしない
- カスタム Context 型を作成しない

## インターフェース

- 使用側のパッケージで定義する（実装側ではない）
- 使用前にインターフェースを定義しない
- モック用に実装側でインターフェースを定義しない

## テスト

### テスト関数
- テストロジックは `Test` 関数内に保持
- アサーションヘルパーより、テーブル駆動テストを優先

### エラーメッセージ
- 何が間違っていたか、入力は何か、実際の結果と期待値を含める
```go
if got != want {
    t.Errorf("Foo(%q) = %d; want %d", input, got, want)
}
```

### t.Error vs t.Fatal
- セットアップ失敗など、続行不可能な場合のみ `t.Fatal` を使用
- 別の goroutine から `t.Fatal` を呼び出さない

## 並行処理

### Goroutine
- goroutine がいつ終了するか明確にする
- goroutine のリークを避ける

### Channel
- 同期関数を優先（非同期より）
- 必要に応じて呼び出し側で並行性を追加

## その他のベストプラクティス

### 暗号
- 鍵生成には `crypto/rand` を使用（`math/rand` は不可）

### ポインタ vs 値
- 小さな不変の構造体や基本型には値レシーバー
- 大きな構造体やミューテーションが必要な場合はポインタレシーバー
- レシーバー型を混在させない

### defer
- リソースのクリーンアップに `defer` を使用
- `defer` は関数の先頭近くに配置

## 参考リンク

- https://go.dev/doc/effective_go
- https://go.dev/wiki/CodeReviewComments
- https://google.github.io/styleguide/go/guide
- https://google.github.io/styleguide/go/best-practices
- https://google.github.io/styleguide/go/decisions
