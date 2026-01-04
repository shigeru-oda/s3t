# Tech Stack

## 言語
- Go 1.25.5

## 主要ライブラリ

### CLI フレームワーク
- github.com/spf13/cobra - コマンドライン引数の解析とサブコマンド管理

### AWS SDK
- github.com/aws/aws-sdk-go-v2 - AWS API クライアント
- github.com/aws/aws-sdk-go-v2/config - 認証設定の読み込み
- github.com/aws/aws-sdk-go-v2/service/s3tables - S3 Tables API
- github.com/aws/smithy-go - AWS エラーハンドリング

### インタラクティブUI
- github.com/manifoldco/promptui - インタラクティブ選択UI（リアルタイムフィルタリング対応）

### テスト
- github.com/leanovate/gopter - プロパティベーステスト

## 開発ツール
- gofmt - コードフォーマット
- go test - テスト実行
- go build - ビルド

## コーディング規約

### Go スタイル
- 標準の Go フォーマット (`gofmt`) に従う
- エラーは必ずラップして返す (`fmt.Errorf` または `WrapError`)
- インターフェースは使用側で定義する

### 命名規則
- パッケージ名: 小文字、単数形
- エクスポートされる関数/型: PascalCase
- 内部関数/変数: camelCase

### エラーハンドリング
- AWS API エラーは `WrapError` でラップする
- ユーザーフレンドリーなメッセージと解決策を提供する
- エラータイプ（`ErrorType`）で分類する

### テスト
- プロパティベーステストを活用する
- モックには interface を使用する
