# Implementation Plan: S3 Tables CLI

## Overview

GoベースのCLIツールを段階的に実装する。プロジェクト構造のセットアップから始め、コア機能、CLI統合、テストの順で進める。

## Tasks

- [x] 1. プロジェクト構造とGoモジュールのセットアップ
  - Go moduleの初期化 (`go mod init`)
  - 必要な依存関係の追加 (Cobra, AWS SDK for Go v2)
  - ディレクトリ構造の作成 (cmd/, internal/s3tables/)
  - _Requirements: 4.1_

- [x] 2. 入力バリデーションの実装
  - [x] 2.1 バリデーション関数の実装
    - `internal/s3tables/validation.go`を作成
    - TableBucket, Namespace, Tableの各バリデーション関数を実装
    - AWS APIの制約に基づくパターンマッチング
    - _Requirements: 4.3_
  - [x] 2.2 バリデーションのプロパティテスト
    - **Property 1: 入力バリデーションの一貫性**
    - **Validates: Requirements 4.3**

- [x] 3. S3TablesCreatorの実装
  - [x] 3.1 インターフェースとCreator構造体の定義
    - `internal/s3tables/creator.go`を作成
    - S3TablesAPIインターフェースの定義
    - S3TablesCreator構造体とコンストラクタの実装
    - _Requirements: 1.1, 2.1, 3.1_
  - [x] 3.2 リソース存在確認ロジックの実装
    - GetTableBucket, GetNamespace, GetTableの呼び出し
    - NotFoundExceptionの判定ロジック
    - _Requirements: 1.2, 2.2, 3.2_
  - [x] 3.3 階層的作成ロジックの実装
    - Create関数の実装
    - Table Bucket → Namespace → Tableの順序で作成
    - 既存リソースの通知と続行ロジック
    - _Requirements: 1.1, 1.2, 2.1, 2.2, 3.1, 3.2_
  - [x] 3.4 既存リソース検出のプロパティテスト
    - **Property 2: 既存リソース検出の正確性**
    - **Validates: Requirements 1.2, 2.2, 3.2**
  - [x] 3.5 階層的作成順序のプロパティテスト
    - **Property 3: 階層的作成の順序保証**
    - **Validates: Requirements 1.1, 2.1, 3.1**

- [x] 4. エラーハンドリングの実装
  - [x] 4.1 エラー種別の定義と変換
    - AWS APIエラーからユーザーフレンドリーなメッセージへの変換
    - _Requirements: 1.3, 2.3, 3.3, 5.2_
  - [x] 4.2 エラーハンドリングの単体テスト
    - 各エラー種別のテストケース
    - _Requirements: 1.3, 2.3, 3.3_

- [x] 5. CLIコマンドの実装
  - [x] 5.1 Cobraを使用したcreateコマンドの実装
    - `cmd/create.go`を作成
    - 引数パース、バリデーション、Creator呼び出し
    - 結果の出力フォーマット
    - _Requirements: 4.1, 4.2_
  - [x] 5.2 main.goの実装
    - ルートコマンドの設定
    - AWS SDKクライアントの初期化
    - _Requirements: 5.1_
  - [x] 5.3 CLIの単体テスト
    - 引数不足時のエラー
    - 無効な引数値のエラー
    - _Requirements: 4.2, 4.3_

- [x] 6. チェックポイント - 全テスト実行
  - 全テストが通ることを確認
  - 質問があればユーザーに確認

- [x] 7. Table Bucket ARN取得ロジックの修正
  - [x] 7.1 S3TablesAPIインターフェースにListTableBucketsを追加
    - ListTableBuckets APIを使用してTable Bucket名からARNを取得
    - _Requirements: 2.1, 3.1_
  - [x] 7.2 checkTableBucketExists関数の修正
    - ListTableBucketsでprefixフィルタを使用して名前からARNを取得
    - GetTableBucketはARNが必要なため、ListTableBucketsを先に呼び出す
    - _Requirements: 1.2, 2.1, 3.1_
  - [x] 7.3 プロパティテストの更新
    - モックにListTableBucketsを追加
    - **Property 2: 既存リソース検出の正確性**の更新
    - **Validates: Requirements 1.2, 2.2, 3.2**

- [x] 8. チェックポイント - 全テスト実行
  - 全テストが通ることを確認
  - 質問があればユーザーに確認

## Notes

- `*`マークのタスクはオプション（MVPでは省略可能）
- 各タスクは特定の要件にトレースバック可能
- プロパティテストはgopterライブラリを使用
