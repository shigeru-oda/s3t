package cmd

import (
	"context"
	"testing"
	"time"

	"s3t/internal/s3tables"

	awss3tables "github.com/aws/aws-sdk-go-v2/service/s3tables"
	"github.com/aws/aws-sdk-go-v2/service/s3tables/types"
	"github.com/spf13/cobra"
)

// mockS3TablesAPI implements s3tables.S3TablesAPI for testing
type mockS3TablesAPI struct {
	listTableBucketsFunc func(ctx context.Context, params *awss3tables.ListTableBucketsInput, optFns ...func(*awss3tables.Options)) (*awss3tables.ListTableBucketsOutput, error)
	listNamespacesFunc   func(ctx context.Context, params *awss3tables.ListNamespacesInput, optFns ...func(*awss3tables.Options)) (*awss3tables.ListNamespacesOutput, error)
	listTablesFunc       func(ctx context.Context, params *awss3tables.ListTablesInput, optFns ...func(*awss3tables.Options)) (*awss3tables.ListTablesOutput, error)
	getTableFunc         func(ctx context.Context, params *awss3tables.GetTableInput, optFns ...func(*awss3tables.Options)) (*awss3tables.GetTableOutput, error)
}

func (m *mockS3TablesAPI) ListTableBuckets(ctx context.Context, params *awss3tables.ListTableBucketsInput, optFns ...func(*awss3tables.Options)) (*awss3tables.ListTableBucketsOutput, error) {
	if m.listTableBucketsFunc != nil {
		return m.listTableBucketsFunc(ctx, params, optFns...)
	}
	return &awss3tables.ListTableBucketsOutput{}, nil
}

func (m *mockS3TablesAPI) ListNamespaces(ctx context.Context, params *awss3tables.ListNamespacesInput, optFns ...func(*awss3tables.Options)) (*awss3tables.ListNamespacesOutput, error) {
	if m.listNamespacesFunc != nil {
		return m.listNamespacesFunc(ctx, params, optFns...)
	}
	return &awss3tables.ListNamespacesOutput{}, nil
}

func (m *mockS3TablesAPI) ListTables(ctx context.Context, params *awss3tables.ListTablesInput, optFns ...func(*awss3tables.Options)) (*awss3tables.ListTablesOutput, error) {
	if m.listTablesFunc != nil {
		return m.listTablesFunc(ctx, params, optFns...)
	}
	return &awss3tables.ListTablesOutput{}, nil
}

func (m *mockS3TablesAPI) GetTable(ctx context.Context, params *awss3tables.GetTableInput, optFns ...func(*awss3tables.Options)) (*awss3tables.GetTableOutput, error) {
	if m.getTableFunc != nil {
		return m.getTableFunc(ctx, params, optFns...)
	}
	return &awss3tables.GetTableOutput{}, nil
}

func (m *mockS3TablesAPI) GetTableBucket(ctx context.Context, params *awss3tables.GetTableBucketInput, optFns ...func(*awss3tables.Options)) (*awss3tables.GetTableBucketOutput, error) {
	return &awss3tables.GetTableBucketOutput{}, nil
}

func (m *mockS3TablesAPI) CreateTableBucket(ctx context.Context, params *awss3tables.CreateTableBucketInput, optFns ...func(*awss3tables.Options)) (*awss3tables.CreateTableBucketOutput, error) {
	return &awss3tables.CreateTableBucketOutput{}, nil
}

func (m *mockS3TablesAPI) GetNamespace(ctx context.Context, params *awss3tables.GetNamespaceInput, optFns ...func(*awss3tables.Options)) (*awss3tables.GetNamespaceOutput, error) {
	return &awss3tables.GetNamespaceOutput{}, nil
}

func (m *mockS3TablesAPI) CreateNamespace(ctx context.Context, params *awss3tables.CreateNamespaceInput, optFns ...func(*awss3tables.Options)) (*awss3tables.CreateNamespaceOutput, error) {
	return &awss3tables.CreateNamespaceOutput{}, nil
}

func (m *mockS3TablesAPI) CreateTable(ctx context.Context, params *awss3tables.CreateTableInput, optFns ...func(*awss3tables.Options)) (*awss3tables.CreateTableOutput, error) {
	return &awss3tables.CreateTableOutput{}, nil
}

// mockInteractiveSelector implements s3tables.InteractiveSelector for testing
type mockInteractiveSelector struct {
	selectWithFilterFunc func(label string, items []string, showBack bool) (*s3tables.SelectionResult, error)
}

func (m *mockInteractiveSelector) SelectWithFilter(label string, items []string, showBack bool) (*s3tables.SelectionResult, error) {
	if m.selectWithFilterFunc != nil {
		return m.selectWithFilterFunc(label, items, showBack)
	}
	if len(items) > 0 {
		return &s3tables.SelectionResult{Selected: items[0], Action: s3tables.ActionSelect}, nil
	}
	return &s3tables.SelectionResult{Action: s3tables.ActionBack}, nil
}

// TestListCommand_TooManyArguments tests that the CLI returns an error when too many arguments are provided
// Requirements: 1.1, 2.1, 3.1
func TestListCommand_TooManyArguments(t *testing.T) {
	cmd := &cobra.Command{Use: "s3t"}
	testListCmd := &cobra.Command{
		Use:  "list [table-bucket] [namespace] [table]",
		Args: cobra.MaximumNArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	cmd.AddCommand(testListCmd)

	_, err := executeCommand(cmd, "list", "bucket", "namespace", "table", "extra")
	if err == nil {
		t.Errorf("expected error for too many arguments, got nil")
	}
}

// TestNavigationController_FromTableBucket tests navigation starting from table bucket level
// Requirements: 1.1, 1.2, 1.3
func TestNavigationController_FromTableBucket(t *testing.T) {
	now := time.Now()
	bucketName := "test-bucket"
	bucketARN := "arn:aws:s3tables:us-east-1:123456789012:bucket/test-bucket"
	namespace := "test_namespace"
	tableName := "test_table"
	tableARN := "arn:aws:s3tables:us-east-1:123456789012:bucket/test-bucket/table/test_table"

	mock := &mockS3TablesAPI{
		listTableBucketsFunc: func(ctx context.Context, params *awss3tables.ListTableBucketsInput, optFns ...func(*awss3tables.Options)) (*awss3tables.ListTableBucketsOutput, error) {
			return &awss3tables.ListTableBucketsOutput{
				TableBuckets: []types.TableBucketSummary{
					{
						Name:      &bucketName,
						Arn:       &bucketARN,
						CreatedAt: &now,
					},
				},
			}, nil
		},
		listNamespacesFunc: func(ctx context.Context, params *awss3tables.ListNamespacesInput, optFns ...func(*awss3tables.Options)) (*awss3tables.ListNamespacesOutput, error) {
			return &awss3tables.ListNamespacesOutput{
				Namespaces: []types.NamespaceSummary{
					{
						Namespace: []string{namespace},
						CreatedAt: &now,
					},
				},
			}, nil
		},
		listTablesFunc: func(ctx context.Context, params *awss3tables.ListTablesInput, optFns ...func(*awss3tables.Options)) (*awss3tables.ListTablesOutput, error) {
			return &awss3tables.ListTablesOutput{
				Tables: []types.TableSummary{
					{
						Name:      &tableName,
						TableARN:  &tableARN,
						Namespace: []string{namespace},
						CreatedAt: &now,
						Type:      types.TableTypeCustomer,
					},
				},
			}, nil
		},
	}

	lister := s3tables.NewS3TablesLister(mock)

	// Mock selector that selects first item each time
	callCount := 0
	selector := &mockInteractiveSelector{
		selectWithFilterFunc: func(label string, items []string, showBack bool) (*s3tables.SelectionResult, error) {
			callCount++
			if len(items) > 0 {
				return &s3tables.SelectionResult{Selected: items[0], Action: s3tables.ActionSelect}, nil
			}
			return &s3tables.SelectionResult{Action: s3tables.ActionBack}, nil
		},
	}

	controller := s3tables.NewNavigationController(lister, selector)

	ctx := context.Background()
	err := controller.Navigate(ctx, s3tables.LevelTableBucket)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Should have called selector 3 times (bucket, namespace, table)
	if callCount != 3 {
		t.Errorf("expected 3 selector calls, got %d", callCount)
	}
}

// TestShowTableDetails tests the showTableDetails function
// Requirements: 3.3
func TestShowTableDetails(t *testing.T) {
	now := time.Now()
	bucketName := "test-bucket"
	bucketARN := "arn:aws:s3tables:us-east-1:123456789012:bucket/test-bucket"
	namespace := "test_namespace"
	tableName := "test_table"
	tableARN := "arn:aws:s3tables:us-east-1:123456789012:bucket/test-bucket/table/test_table"

	mock := &mockS3TablesAPI{
		listTableBucketsFunc: func(ctx context.Context, params *awss3tables.ListTableBucketsInput, optFns ...func(*awss3tables.Options)) (*awss3tables.ListTableBucketsOutput, error) {
			return &awss3tables.ListTableBucketsOutput{
				TableBuckets: []types.TableBucketSummary{
					{
						Name:      &bucketName,
						Arn:       &bucketARN,
						CreatedAt: &now,
					},
				},
			}, nil
		},
		getTableFunc: func(ctx context.Context, params *awss3tables.GetTableInput, optFns ...func(*awss3tables.Options)) (*awss3tables.GetTableOutput, error) {
			return &awss3tables.GetTableOutput{
				Name:      &tableName,
				TableARN:  &tableARN,
				CreatedAt: &now,
				Type:      types.TableTypeCustomer,
			}, nil
		},
	}

	lister := s3tables.NewS3TablesLister(mock)

	ctx := context.Background()
	err := showTableDetails(ctx, lister, bucketName, namespace, tableName)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestNavigationController_TableBucketNotFound tests error handling when table bucket is not found
// Requirements: 2.5, 6.3
func TestNavigationController_TableBucketNotFound(t *testing.T) {
	mock := &mockS3TablesAPI{
		listTableBucketsFunc: func(ctx context.Context, params *awss3tables.ListTableBucketsInput, optFns ...func(*awss3tables.Options)) (*awss3tables.ListTableBucketsOutput, error) {
			return &awss3tables.ListTableBucketsOutput{
				TableBuckets: []types.TableBucketSummary{},
			}, nil
		},
	}

	lister := s3tables.NewS3TablesLister(mock)

	// Mock selector that should not be called
	selector := &mockInteractiveSelector{
		selectWithFilterFunc: func(label string, items []string, showBack bool) (*s3tables.SelectionResult, error) {
			t.Error("selector should not be called when no buckets exist")
			return &s3tables.SelectionResult{Action: s3tables.ActionBack}, nil
		},
	}

	controller := s3tables.NewNavigationController(lister, selector)

	ctx := context.Background()
	err := controller.Navigate(ctx, s3tables.LevelTableBucket)
	// When no buckets found, it should exit gracefully without error
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
