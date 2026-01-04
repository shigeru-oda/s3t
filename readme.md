# S3 Tables CLI (s3t)

Amazon S3 Tables リソースを簡単に作成・管理するためのコマンドラインツール。

## 特徴

- Table Bucket、Namespace、Table の階層的な作成
- 既存リソースの自動検出とスキップ（冪等性）
- ユーザーフレンドリーなエラーメッセージと解決策の提示
- AWS デフォルト認証チェーンのサポート

## インストール

### ソースからビルド

```bash
git clone https://github.com/your-username/s3t.git
cd s3t
go build -o s3t .
```

### バイナリを PATH に追加

```bash
mv s3t /usr/local/bin/
```

## 前提条件

- Go 1.25.5 以上（ビルドする場合）
- AWS 認証情報の設定

### AWS 認証情報の設定方法

以下のいずれかの方法で認証情報を設定してください：

```bash
# AWS CLI を使用
aws configure

# 環境変数を使用
export AWS_ACCESS_KEY_ID=your-access-key
export AWS_SECRET_ACCESS_KEY=your-secret-key
export AWS_REGION=ap-northeast-1

# IAM ロール（EC2/ECS/Lambda 環境）
# 自動的に認証情報が取得されます
```

## 使い方

### 基本コマンド

```bash
# S3 Tables リソースを作成
s3t create <table-bucket> <namespace> <table>
```

### 例

```bash
# my-bucket に analytics 名前空間と sales テーブルを作成
s3t create my-bucket analytics sales
```

### 出力例

```
=== S3 Tables Resource Creation Summary ===

  • Table Bucket 'my-bucket' created successfully
  • Namespace 'analytics' created successfully
  • Table 'sales' created successfully

Created: 3 resource(s)

Table Bucket ARN: arn:aws:s3tables:ap-northeast-1:123456789012:bucket/my-bucket
Table ARN: arn:aws:s3tables:ap-northeast-1:123456789012:bucket/my-bucket/table/analytics/sales
```

既存リソースがある場合は自動的にスキップされます：

```
=== S3 Tables Resource Creation Summary ===

  • Table Bucket 'my-bucket' already exists
  • Namespace 'analytics' already exists
  • Table 'sales' created successfully

Created: 1 resource(s)
Already existed: 2 resource(s)
```

## リソース命名規則

### Table Bucket
- 長さ: 3-63 文字
- 使用可能文字: 小文字、数字、ハイフン（`-`）

### Namespace
- 長さ: 1-255 文字
- 使用可能文字: 小文字、数字、アンダースコア（`_`）

### Table
- 長さ: 1-255 文字
- 使用可能文字: 小文字、数字、アンダースコア（`_`）

## ヘルプ

```bash
# 全体のヘルプ
s3t --help

# create コマンドのヘルプ
s3t create --help

# バージョン確認
s3t --version
```

## 必要な IAM 権限

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3tables:CreateTableBucket",
        "s3tables:GetTableBucket",
        "s3tables:CreateNamespace",
        "s3tables:GetNamespace",
        "s3tables:CreateTable",
        "s3tables:GetTable"
      ],
      "Resource": "*"
    }
  ]
}
```

## ライセンス

本プロジェクトはMITライセンスの下で公開されています。詳細は[LICENSE](LICENSE)ファイルをご覧ください。