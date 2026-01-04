# Requirements Document

## Introduction

Amazon S3 Tablesリソース（Table Bucket、Namespace、Table）を作成するためのGoベースのCLIツール。階層的なリソース作成をサポートし、既存リソースの検出と適切な通知を行う。

## Glossary

- **S3Tables_CLI**: S3 Tablesリソースを管理するコマンドラインツール
- **Table_Bucket**: S3 Tablesの最上位コンテナリソース（CLI引数ではTable Bucket名を指定）
- **Table_Bucket_ARN**: Table Bucketを一意に識別するAmazon Resource Name（例: `arn:aws:s3tables:region:account:bucket/bucket-name`）。GetTableBucket APIから取得される
- **Namespace**: Table_Bucket配下のリソースグループ化単位（AWS APIではTable_Bucket_ARNを使用してアクセス）
- **Table**: Namespace配下の実際のテーブルリソース（AWS APIではTable_Bucket_ARNを使用してアクセス）
- **AWS_SDK**: AWSサービスとの通信に使用するSDK

## Requirements

### Requirement 1: Table Bucket作成

**User Story:** As a developer, I want to create a Table Bucket, so that I can organize my S3 Tables resources.

#### Acceptance Criteria

1. WHEN a user executes `s3t create $TableBucket $NameSpace $Table` with a non-existent Table Bucket, THE S3Tables_CLI SHALL create the Table Bucket
2. WHEN a Table Bucket already exists, THE S3Tables_CLI SHALL notify the user that the Table Bucket is already created and continue with Namespace creation
3. IF Table Bucket creation fails due to AWS API error, THEN THE S3Tables_CLI SHALL return a descriptive error message

### Requirement 2: Namespace作成

**User Story:** As a developer, I want to create a Namespace under a Table Bucket, so that I can logically group my tables.

#### Acceptance Criteria

1. WHEN a Table Bucket exists and Namespace does not exist, THE S3Tables_CLI SHALL retrieve the Table_Bucket_ARN from GetTableBucket API and create the Namespace using that ARN
2. WHEN a Namespace already exists under the Table Bucket, THE S3Tables_CLI SHALL notify the user that the Namespace is already created and continue with Table creation
3. IF Namespace creation fails due to AWS API error, THEN THE S3Tables_CLI SHALL return a descriptive error message

### Requirement 3: Table作成

**User Story:** As a developer, I want to create a Table under a Namespace, so that I can store and manage tabular data.

#### Acceptance Criteria

1. WHEN a Table Bucket and Namespace exist and Table does not exist, THE S3Tables_CLI SHALL create the Table using the internally retrieved Table_Bucket_ARN and Namespace name
2. WHEN a Table already exists under the Namespace, THE S3Tables_CLI SHALL notify the user that all resources (Table Bucket, Namespace, Table) are already created
3. IF Table creation fails due to AWS API error, THEN THE S3Tables_CLI SHALL return a descriptive error message

### Requirement 4: コマンドライン引数処理

**User Story:** As a developer, I want to use a simple CLI interface, so that I can easily create S3 Tables resources.

#### Acceptance Criteria

1. THE S3Tables_CLI SHALL accept the command format `s3t create <table-bucket> <namespace> <table>`
2. WHEN insufficient arguments are provided, THE S3Tables_CLI SHALL display usage information and exit with an error code
3. WHEN invalid argument values are provided, THE S3Tables_CLI SHALL return a descriptive validation error

### Requirement 5: AWS認証

**User Story:** As a developer, I want the CLI to use my AWS credentials, so that I can authenticate with AWS services.

#### Acceptance Criteria

1. THE S3Tables_CLI SHALL use the default AWS credential chain for authentication
2. IF AWS credentials are not configured, THEN THE S3Tables_CLI SHALL return a descriptive authentication error
