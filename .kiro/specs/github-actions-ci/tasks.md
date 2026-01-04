# Implementation Plan: GitHub Actions CI

## Overview

GitHub Actions CI ワークフローを実装する。ワークフローファイル（ci.yml）と lint 設定ファイル（.golangci.yml）を作成し、テスト、ビルド、静的解析、フォーマットチェック、依存関係検証、セキュリティスキャン、カバレッジレポートを自動実行する。

## Tasks

- [x] 1. GitHub Actions ワークフローディレクトリの作成
  - `.github/workflows/` ディレクトリを作成
  - _Requirements: 1.1, 1.2_

- [x] 2. CI ワークフローファイルの作成
  - [x] 2.1 ワークフローの基本構造とトリガー設定
    - `ci.yml` ファイルを作成
    - `name: CI` を設定
    - `on.push.branches: [main]` を設定
    - `on.pull_request.branches: ['*']` を設定
    - _Requirements: 1.1, 1.2_

  - [x] 2.2 ジョブとセットアップステップの追加
    - `jobs.ci` ジョブを定義
    - `runs-on: ubuntu-latest` を設定
    - `actions/checkout@v4` ステップを追加
    - `actions/setup-go@v5` ステップを追加（go-version-file: 'go.mod', cache: true）
    - _Requirements: 1.3_

  - [x] 2.3 テストステップの追加
    - `go test -v -race -coverprofile=coverage.out ./...` コマンドを追加
    - ステップ名を「Run tests」に設定
    - _Requirements: 2.1, 2.2, 2.3, 2.4_

  - [x] 2.4 ビルドステップの追加
    - `go build -v ./...` コマンドを追加
    - ステップ名を「Build」に設定
    - _Requirements: 3.1, 3.2_

  - [x] 2.5 Lint ステップの追加
    - `golangci/golangci-lint-action@v6` アクションを使用
    - ステップ名を「Lint」に設定
    - _Requirements: 4.1, 4.2_

  - [x] 2.6 フォーマットチェックステップの追加
    - gofmt チェック: `test -z "$(gofmt -l .)"` を追加
    - goimports チェック: `go run golang.org/x/tools/cmd/goimports@latest -l .` を追加
    - ステップ名を「Check formatting」に設定
    - _Requirements: 5.1, 5.2, 5.3_

  - [x] 2.7 依存関係チェックステップの追加
    - `go mod verify` コマンドを追加
    - `go mod tidy && git diff --exit-code go.mod go.sum` コマンドを追加
    - ステップ名を「Check dependencies」に設定
    - _Requirements: 6.1, 6.2, 6.3_

  - [x] 2.8 セキュリティスキャンステップの追加
    - `go run golang.org/x/vuln/cmd/govulncheck@latest ./...` コマンドを追加
    - ステップ名を「Security scan」に設定
    - _Requirements: 7.1, 7.2_

  - [x] 2.9 カバレッジレポートステップの追加
    - `go tool cover -func=coverage.out` コマンドを追加
    - ステップ名を「Coverage report」に設定
    - _Requirements: 8.1, 8.2_

- [x] 3. golangci-lint 設定ファイルの作成
  - `.golangci.yml` ファイルをプロジェクトルートに作成
  - 基本的な linter を有効化（errcheck, gosimple, govet, ineffassign, staticcheck, unused）
  - _Requirements: 4.2_

- [x] 4. Checkpoint - ワークフローの検証
  - YAML 構文が正しいことを確認
  - すべての要件がカバーされていることを確認
  - ユーザーに質問があれば確認

## Notes

- この機能は設定ファイルの作成のみで、Go コードの変更は不要
- プロパティベーステストは適用しない（ランタイムコードではないため）
- 実際の CI 動作確認は PR 作成時に行う
