# Project Structure

## ディレクトリ構成

```
s3t/
├── main.go                 # エントリーポイント
├── go.mod                  # Go モジュール定義
├── go.sum                  # 依存関係ロックファイル
├── cmd/                    # CLI コマンド
│   ├── root.go             # ルートコマンド、AWS クライアント初期化、グローバルフラグ
│   ├── root_property_test.go # root コマンドのプロパティテスト
│   ├── create.go           # create サブコマンド
│   ├── create_test.go      # create コマンドのテスト
│   ├── list.go             # list サブコマンド（階層的リソース一覧表示）
│   └── list_test.go        # list コマンドのテスト
└── internal/               # 内部パッケージ
    └── s3tables/           # S3 Tables ビジネスロジック
        ├── creator.go      # リソース作成ロジック
        ├── creator_property_test.go
        ├── errors.go       # エラーハンドリング
        ├── errors_test.go
        ├── lister.go       # リソース一覧取得（ページネーション対応）
        ├── lister_property_test.go
        ├── navigator.go    # 階層的ナビゲーション制御
        ├── navigator_property_test.go
        ├── selector.go     # インタラクティブ選択UI（リアルタイムフィルタリング対応）
        ├── selector_property_test.go
        ├── validation.go   # 入力バリデーション
        ├── validation_property_test.go
        └── validation_test.go
```

## ファイル命名規則
- Go ソースファイル: snake_case.go
- テストファイル: *_test.go（対象ファイルと同じディレクトリ）
- プロパティテスト: *_property_test.go

## パッケージ構成

### cmd/
CLI コマンドの定義。Cobra を使用してサブコマンドを構成。

- `root.go` - ルートコマンド、AWS クライアント初期化、`--profile`/`--region` グローバルフラグ
- `root_property_test.go` - `buildConfigOptions` 関数のプロパティテスト
- `create.go` - create サブコマンド
- `create_test.go` - create コマンドのテスト
- `list.go` - list サブコマンド（階層的リソース一覧表示とインタラクティブナビゲーション）
- `list_test.go` - list コマンドのテスト

### internal/s3tables/
S3 Tables 操作のビジネスロジック。外部パッケージからはインポート不可。

- `creator.go` - リソース作成の主要ロジック
- `errors.go` - エラー型とラッピング
- `lister.go` - Table Bucket/Namespace/Table の一覧取得（ページネーション対応）
- `navigator.go` - 階層的ナビゲーション制御（状態管理、キャッシュ、戻る機能）
- `selector.go` - インタラクティブ選択UI（promptui使用、リアルタイムフィルタリング、部分一致検索）
- `validation.go` - 入力値のバリデーション

## AWS S3 Tables リソース制約

### Table Bucket
- 長さ: 3-63文字
- パターン: 小文字、数字、ハイフンのみ

### Namespace
- 長さ: 1-255文字
- パターン: 小文字、数字、アンダースコアのみ

### Table
- 長さ: 1-255文字
- パターン: 小文字、数字、アンダースコアのみ

## ビルド・テスト

```bash
# ビルド
go build -o s3t .

# 全テスト実行
go test ./...

# 特定パッケージのテスト
go test ./internal/s3tables/...
```
