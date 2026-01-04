# Requirements Document

## Introduction

S3 Tables CLI (s3t) に AWS Profile と Region をオプションで指定できる機能を追加する。これにより、ユーザーは複数の AWS アカウントや異なるリージョンのリソースを柔軟に管理できるようになる。

## Glossary

- **Profile**: AWS CLI で設定された認証情報のセット（~/.aws/credentials に保存）
- **Region**: AWS リソースが配置される地理的なリージョン（例: ap-northeast-1, us-east-1）

## Requirements

### Requirement 1: Profile オプションの追加

**User Story:** As a developer, I want to specify an AWS profile when running s3t commands, so that I can work with multiple AWS accounts without changing my default configuration.

#### Acceptance Criteria

1. WHEN a user specifies `--profile` flag with a profile name, THE CLI SHALL use the specified profile for AWS authentication
2. WHEN a user does not specify `--profile` flag, THE CLI SHALL use the default AWS credential chain
3. WHEN a user specifies a non-existent profile name, THE CLI SHALL return a clear error message indicating the profile was not found
4. THE CLI SHALL accept `--profile` as a global flag available to all subcommands

### Requirement 2: Region オプションの追加

**User Story:** As a developer, I want to specify an AWS region when running s3t commands, so that I can manage S3 Tables resources in different regions.

#### Acceptance Criteria

1. WHEN a user specifies `--region` flag with a region name, THE CLI SHALL use the specified region for AWS API calls
2. WHEN a user does not specify `--region` flag, THE CLI SHALL use the region from the profile or default configuration
3. WHEN a user specifies an invalid region name, THE CLI SHALL return a clear error message from the AWS API
4. THE CLI SHALL accept `--region` as a global flag available to all subcommands

### Requirement 3: オプションの組み合わせ

**User Story:** As a developer, I want to combine profile and region options, so that I can flexibly configure my AWS connection.

#### Acceptance Criteria

1. WHEN a user specifies both `--profile` and `--region` flags, THE CLI SHALL use the specified profile with the overridden region
2. WHEN a user specifies only `--region` flag, THE CLI SHALL use the default profile with the specified region
3. WHEN a user specifies only `--profile` flag, THE CLI SHALL use the specified profile's configured region

### Requirement 4: ヘルプメッセージの更新

**User Story:** As a developer, I want to see clear documentation for the new options, so that I can understand how to use them.

#### Acceptance Criteria

1. WHEN a user runs `s3t --help`, THE CLI SHALL display `--profile` and `--region` options with descriptions
2. THE CLI SHALL display usage examples showing how to use profile and region options
