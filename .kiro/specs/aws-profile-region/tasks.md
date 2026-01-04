# Implementation Plan: AWS Profile & Region Options

## Overview

S3 Tables CLI に `--profile` と `--region` グローバルフラグを追加する実装計画。Cobra の PersistentFlags を使用し、AWS SDK の config options を動的に構築する。

## Tasks

- [x] 1. グローバルフラグの追加
  - [x] 1.1 `cmd/root.go` に `--profile` と `--region` フラグを追加
    - `awsProfile` と `awsRegion` 変数を定義
    - `init()` で `PersistentFlags().StringVar()` を使用してフラグを登録
    - フラグの説明文を追加
    - _Requirements: 1.4, 2.4_

- [x] 2. AWS クライアント初期化の拡張
  - [x] 2.1 `buildConfigOptions` 関数を作成
    - profile と region を受け取り、`[]func(*config.LoadOptions) error` を返す
    - profile が空でない場合は `config.WithSharedConfigProfile` を追加
    - region が空でない場合は `config.WithRegion` を追加
    - _Requirements: 1.1, 1.2, 2.1, 2.2, 3.1, 3.2, 3.3_
  - [x] 2.2 `buildConfigOptions` のプロパティテストを作成
    - **Property 1: Config Options Building Correctness**
    - **Validates: Requirements 1.1, 1.2, 2.1, 2.2, 3.1, 3.2, 3.3**
  - [x] 2.3 `initAWSClient` を更新して `buildConfigOptions` を使用
    - フラグ値を `buildConfigOptions` に渡す
    - 返された options を `config.LoadDefaultConfig` に渡す
    - _Requirements: 1.1, 2.1, 3.1_

- [x] 3. エラーハンドリングの改善
  - [x] 3.1 `handleConfigError` 関数を作成
    - profile 指定時は profile 固有のエラーメッセージを返す
    - profile 未指定時は既存のエラーメッセージを返す
    - _Requirements: 1.3_
  - [x] 3.2 `initAWSClient` でエラーハンドリングを更新
    - `handleConfigError` を使用してエラーをラップ
    - _Requirements: 1.3_

- [x] 4. ヘルプメッセージの更新
  - [x] 4.1 `rootCmd.Long` を更新して使用例を追加
    - `--profile` と `--region` の使用例を追加
    - _Requirements: 4.1, 4.2_

- [x] 5. Checkpoint - 動作確認
  - すべてのテストが通ることを確認
  - `go build` でビルドが成功することを確認
  - `s3t --help` でフラグが表示されることを確認
  - 質問があればユーザーに確認

## Notes

- 各タスクは前のタスクに依存するため、順番に実行する
- `buildConfigOptions` を抽出することで、AWS SDK への依存なしにテスト可能
