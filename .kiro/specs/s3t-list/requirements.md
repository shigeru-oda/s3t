# Requirements Document

## Introduction

S3 Tables CLI に `list` サブコマンドを追加し、Table Bucket、Namespace、Table の階層的なリソース一覧表示機能を提供する。引数の指定レベルに応じて適切な階層から一覧を開始し、インタラクティブな選択とLike検索による絞り込みをサポートする。

## Glossary

- **List_Command**: `s3t list` サブコマンド。S3 Tables リソースの一覧表示を行う
- **Table_Bucket**: S3 Tables の最上位リソース。Namespace を格納するコンテナ
- **Namespace**: Table Bucket 配下のリソース。Table を格納するコンテナ
- **Table**: Namespace 配下の最下位リソース。実際のデータを格納
- **Interactive_Selector**: ユーザーがリソースを選択するためのインタラクティブなUI。リアルタイムフィルタリングとESCキーによる戻る機能をサポート
- **Like_Filter**: 部分一致検索によるリソース絞り込み機能
- **Pagination**: 大量のリソースをページ単位で取得する機能

## Requirements

### Requirement 1: Table Bucket 一覧表示

**User Story:** As a user, I want to list all table buckets in my AWS account, so that I can see what resources are available.

#### Acceptance Criteria

1. WHEN a user executes `s3t list` without arguments, THE List_Command SHALL retrieve and display all Table Buckets in the configured region
2. WHEN Table Buckets exist, THE List_Command SHALL display them in an Interactive_Selector for user selection
3. WHEN a user selects a Table Bucket, THE List_Command SHALL proceed to display Namespaces within that bucket
4. WHEN no Table Buckets exist, THE List_Command SHALL display a message indicating no resources found
5. WHEN the `--profile` flag is provided, THE List_Command SHALL use the specified AWS profile for authentication
6. WHEN the `--region` flag is provided, THE List_Command SHALL list resources in the specified region

### Requirement 2: Namespace 一覧表示

**User Story:** As a user, I want to list namespaces within a table bucket, so that I can navigate the resource hierarchy.

#### Acceptance Criteria

1. WHEN a user executes `s3t list <table-bucket>`, THE List_Command SHALL retrieve and display all Namespaces within the specified Table Bucket
2. WHEN Namespaces exist, THE List_Command SHALL display them in an Interactive_Selector for user selection
3. WHEN a user selects a Namespace, THE List_Command SHALL proceed to display Tables within that namespace
4. WHEN no Namespaces exist in the bucket, THE List_Command SHALL display a message indicating no resources found
5. WHEN the specified Table Bucket does not exist, THE List_Command SHALL display an appropriate error message

### Requirement 3: Table 一覧表示

**User Story:** As a user, I want to list tables within a namespace, so that I can see the available data tables.

#### Acceptance Criteria

1. WHEN a user executes `s3t list <table-bucket> <namespace>`, THE List_Command SHALL retrieve and display all Tables within the specified Namespace
2. WHEN Tables exist, THE List_Command SHALL display them in a list format
3. WHEN a user executes `s3t list <table-bucket> <namespace> <table>`, THE List_Command SHALL display details of the specified Table
4. WHEN no Tables exist in the namespace, THE List_Command SHALL display a message indicating no resources found
5. WHEN the specified Namespace does not exist, THE List_Command SHALL display an appropriate error message

### Requirement 4: Interactive_Selector 内でのフィルタリング

**User Story:** As a user, I want to filter resources within the interactive selector, so that I can quickly find specific resources without leaving the selection interface.

#### Acceptance Criteria

1. WHEN a user types characters in the Interactive_Selector, THE Like_Filter SHALL filter the displayed resources using case-insensitive partial matching in real-time
2. THE Like_Filter SHALL automatically treat the input as a substring search (equivalent to `*input*` pattern) without requiring explicit wildcards
3. WHEN the filter pattern matches resources, THE Interactive_Selector SHALL display only the matching resources
4. WHEN the filter pattern matches no resources, THE Interactive_Selector SHALL display a message indicating no matching resources found
5. WHEN a user clears the filter input, THE Interactive_Selector SHALL display all resources again

### Requirement 7: ESC キーによるナビゲーション

**User Story:** As a user, I want to navigate back to the previous screen using the ESC key, so that I can easily move through the resource hierarchy.

#### Acceptance Criteria

1. WHEN a user presses ESC in the Table selection screen, THE List_Command SHALL return to the Namespace selection screen
2. WHEN a user presses ESC in the Namespace selection screen, THE List_Command SHALL return to the Table Bucket selection screen
3. WHEN a user presses ESC in the Table Bucket selection screen, THE List_Command SHALL exit the application
4. WHEN navigating back, THE Interactive_Selector SHALL preserve the previously loaded resources without re-fetching from the API

### Requirement 5: ページネーション対応

**User Story:** As a user, I want the list command to handle large numbers of resources efficiently, so that I can work with accounts containing many resources.

#### Acceptance Criteria

1. WHEN retrieving resources from AWS API, THE List_Command SHALL use Pagination to fetch all available resources
2. WHEN the API returns a continuation token, THE List_Command SHALL automatically fetch subsequent pages
3. THE List_Command SHALL aggregate all pages before displaying results to the user

### Requirement 6: エラーハンドリング

**User Story:** As a user, I want clear error messages when something goes wrong, so that I can understand and resolve issues.

#### Acceptance Criteria

1. IF an AWS API error occurs, THEN THE List_Command SHALL display a user-friendly error message with the underlying cause
2. IF authentication fails, THEN THE List_Command SHALL suggest checking AWS credentials configuration
3. IF a specified resource does not exist, THEN THE List_Command SHALL clearly indicate which resource was not found
4. IF network connectivity issues occur, THEN THE List_Command SHALL display an appropriate error message
