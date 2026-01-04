# Implementation Plan: S3 Tables List Command Enhancement

## Overview

既存の `list` コマンドを拡張し、Interactive_Selector 内でのリアルタイムフィルタリングと ESC キーによる階層ナビゲーションを実装する。

## Tasks

- [x] 1. NavigationController の基盤実装
  - [x] 1.1 NavigationLevel, NavigationAction, NavigationState 型を定義
    - `internal/s3tables/navigator.go` に新規作成
    - _Requirements: 1.1, 2.1, 3.1, 7.1, 7.2, 7.3_

  - [x] 1.2 NavigationController 構造体と基本メソッドを実装
    - Navigate, navigateTableBuckets, navigateNamespaces, navigateTables メソッド
    - _Requirements: 1.1, 1.2, 1.3, 2.1, 2.2, 2.3, 3.1, 3.2_

  - [x] 1.3 Property 1 のプロパティテストを作成
    - **Property 1: Navigation level determines correct resource listing**
    - **Validates: Requirements 1.1, 1.2, 2.1, 2.2, 3.1, 3.2**

  - [x] 1.4 Property 2 のプロパティテストを作成
    - **Property 2: Selection advances navigation to next level**
    - **Validates: Requirements 1.3, 2.3**

- [x] 2. InteractiveSelector の拡張
  - [x] 2.1 SelectionResult 型と InteractiveSelector インターフェースを定義
    - `internal/s3tables/selector.go` を更新
    - _Requirements: 4.1, 7.1, 7.2, 7.3_

  - [x] 2.2 FilterablePromptSelector を実装
    - promptui の Searcher 機能を使用したリアルタイムフィルタリング
    - ESC キー (promptui.ErrInterrupt) を ActionBack として処理
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 7.1, 7.2, 7.3_

  - [x] 2.3 Property 3 のプロパティテストを作成
    - **Property 3: Filter correctly filters items by pattern**
    - **Validates: Requirements 4.1, 4.2, 4.5**

  - [x] 2.4 Property 4 のプロパティテストを作成
    - **Property 4: Substring matching is automatic**
    - **Validates: Requirements 4.2**

- [x] 3. Checkpoint - NavigationController と Selector のテスト確認
  - Ensure all tests pass, ask the user if questions arise.

- [x] 4. ESC キーによるナビゲーションの実装
  - [x] 4.1 NavigationController に戻るロジックを実装
    - ActionBack 時に前のレベルに戻る
    - LevelTableBucket で ActionBack 時はアプリケーション終了
    - _Requirements: 7.1, 7.2, 7.3_

  - [x] 4.2 キャッシュを使用した戻るナビゲーションを実装
    - 戻る際は API を再呼び出しせず、キャッシュされたデータを使用
    - _Requirements: 7.4_

  - [x] 4.3 Property 5 のプロパティテストを作成
    - **Property 5: ESC navigates back one level**
    - **Validates: Requirements 7.1, 7.2, 7.3**

  - [x] 4.4 Property 6 のプロパティテストを作成
    - **Property 6: Back navigation uses cached data**
    - **Validates: Requirements 7.4**

- [x] 5. cmd/list.go の更新
  - [x] 5.1 --filter オプションを削除
    - filterPattern 変数と関連フラグを削除
    - _Requirements: 4.1_

  - [x] 5.2 NavigationController を使用するように runList を更新
    - 引数に応じて適切な NavigationLevel から開始
    - _Requirements: 1.1, 2.1, 3.1_

  - [x] 5.3 ヘルプテキストを更新
    - Interactive Features の説明を追加
    - _Requirements: 4.1, 7.1, 7.2, 7.3_

- [x] 6. エッジケースとエラーハンドリング
  - [x] 6.1 空のリソースリストの処理を実装
    - リソースが見つからない場合のメッセージ表示
    - `navigator.go` の `navigateTableBuckets`, `navigateNamespaces`, `navigateTables` で実装済み
    - _Requirements: 1.4, 2.4, 3.4_

  - [x] 6.2 存在しないリソースへのアクセス時のエラー処理
    - `errors.go` の `WrapError` と `IsNotFoundError` で実装済み
    - `lister.go` の `GetTableBucketARN` で存在しないバケットのエラー処理実装済み
    - _Requirements: 2.5, 3.5, 6.1, 6.2, 6.3_

  - [x] 6.3 エッジケースのユニットテストを作成
    - `cmd/list_test.go` の `TestNavigationController_TableBucketNotFound` で空リストテスト実装済み
    - `internal/s3tables/errors_test.go` で各種エラーハンドリングテスト実装済み
    - _Requirements: 1.4, 2.4, 2.5, 3.4, 3.5_

- [x] 7. Final checkpoint - 全テストの確認
  - 全テストがパスすることを確認済み (`go test ./...`)

## Notes

- All tasks including tests are required for comprehensive implementation
- Each task references specific requirements for traceability
- Checkpoints ensure incremental validation
- Property tests validate universal correctness properties
- Unit tests validate specific examples and edge cases
