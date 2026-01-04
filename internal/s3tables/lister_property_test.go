package s3tables

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3tables"
	"github.com/aws/aws-sdk-go-v2/service/s3tables/types"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// PaginatedMockS3TablesAPI simulates paginated API responses
type PaginatedMockS3TablesAPI struct {
	// TableBuckets to return, split across pages
	TableBuckets []types.TableBucketSummary
	PageSize     int

	// Namespaces to return, split across pages
	Namespaces []types.NamespaceSummary

	// Tables to return, split across pages
	Tables []types.TableSummary

	// Callbacks for tracking API calls
	OnListTableBuckets func()
	OnListNamespaces   func()
	OnListTables       func()
}

func (m *PaginatedMockS3TablesAPI) ListTableBuckets(ctx context.Context, params *s3tables.ListTableBucketsInput, optFns ...func(*s3tables.Options)) (*s3tables.ListTableBucketsOutput, error) {
	if m.OnListTableBuckets != nil {
		m.OnListTableBuckets()
	}

	startIndex := 0
	if params.ContinuationToken != nil && *params.ContinuationToken != "" {
		fmt.Sscanf(*params.ContinuationToken, "%d", &startIndex)
	}

	endIndex := startIndex + m.PageSize
	if endIndex > len(m.TableBuckets) {
		endIndex = len(m.TableBuckets)
	}

	var nextToken *string
	if endIndex < len(m.TableBuckets) {
		token := fmt.Sprintf("%d", endIndex)
		nextToken = &token
	}

	return &s3tables.ListTableBucketsOutput{
		TableBuckets:      m.TableBuckets[startIndex:endIndex],
		ContinuationToken: nextToken,
	}, nil
}

func (m *PaginatedMockS3TablesAPI) GetTableBucket(ctx context.Context, params *s3tables.GetTableBucketInput, optFns ...func(*s3tables.Options)) (*s3tables.GetTableBucketOutput, error) {
	return nil, &types.NotFoundException{Message: aws.String("not found")}
}

func (m *PaginatedMockS3TablesAPI) CreateTableBucket(ctx context.Context, params *s3tables.CreateTableBucketInput, optFns ...func(*s3tables.Options)) (*s3tables.CreateTableBucketOutput, error) {
	return &s3tables.CreateTableBucketOutput{Arn: aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test")}, nil
}

func (m *PaginatedMockS3TablesAPI) GetNamespace(ctx context.Context, params *s3tables.GetNamespaceInput, optFns ...func(*s3tables.Options)) (*s3tables.GetNamespaceOutput, error) {
	return nil, &types.NotFoundException{Message: aws.String("not found")}
}

func (m *PaginatedMockS3TablesAPI) CreateNamespace(ctx context.Context, params *s3tables.CreateNamespaceInput, optFns ...func(*s3tables.Options)) (*s3tables.CreateNamespaceOutput, error) {
	return &s3tables.CreateNamespaceOutput{Namespace: params.Namespace}, nil
}

func (m *PaginatedMockS3TablesAPI) GetTable(ctx context.Context, params *s3tables.GetTableInput, optFns ...func(*s3tables.Options)) (*s3tables.GetTableOutput, error) {
	return nil, &types.NotFoundException{Message: aws.String("not found")}
}

func (m *PaginatedMockS3TablesAPI) CreateTable(ctx context.Context, params *s3tables.CreateTableInput, optFns ...func(*s3tables.Options)) (*s3tables.CreateTableOutput, error) {
	return &s3tables.CreateTableOutput{TableARN: aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test/table/t")}, nil
}

func (m *PaginatedMockS3TablesAPI) ListNamespaces(ctx context.Context, params *s3tables.ListNamespacesInput, optFns ...func(*s3tables.Options)) (*s3tables.ListNamespacesOutput, error) {
	if m.OnListNamespaces != nil {
		m.OnListNamespaces()
	}

	startIndex := 0
	if params.ContinuationToken != nil && *params.ContinuationToken != "" {
		fmt.Sscanf(*params.ContinuationToken, "%d", &startIndex)
	}

	endIndex := startIndex + m.PageSize
	if endIndex > len(m.Namespaces) {
		endIndex = len(m.Namespaces)
	}

	var nextToken *string
	if endIndex < len(m.Namespaces) {
		token := fmt.Sprintf("%d", endIndex)
		nextToken = &token
	}

	return &s3tables.ListNamespacesOutput{
		Namespaces:        m.Namespaces[startIndex:endIndex],
		ContinuationToken: nextToken,
	}, nil
}

func (m *PaginatedMockS3TablesAPI) ListTables(ctx context.Context, params *s3tables.ListTablesInput, optFns ...func(*s3tables.Options)) (*s3tables.ListTablesOutput, error) {
	if m.OnListTables != nil {
		m.OnListTables()
	}

	startIndex := 0
	if params.ContinuationToken != nil && *params.ContinuationToken != "" {
		fmt.Sscanf(*params.ContinuationToken, "%d", &startIndex)
	}

	endIndex := startIndex + m.PageSize
	if endIndex > len(m.Tables) {
		endIndex = len(m.Tables)
	}

	var nextToken *string
	if endIndex < len(m.Tables) {
		token := fmt.Sprintf("%d", endIndex)
		nextToken = &token
	}

	return &s3tables.ListTablesOutput{
		Tables:            m.Tables[startIndex:endIndex],
		ContinuationToken: nextToken,
	}, nil
}

// PaginationTestInput represents the input for pagination property tests
type PaginationTestInput struct {
	ItemCount int
	PageSize  int
}

// TestPropertyPaginationAggregatesAllResources tests that pagination correctly aggregates all resources
// Feature: s3t-list, Property 1: Pagination aggregates all resources
// **Validates: Requirements 5.1, 5.2, 5.3**
func TestPropertyPaginationAggregatesAllResources(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Generator for pagination test input
	paginationInputGen := gen.Struct(reflect.TypeOf(PaginationTestInput{}), map[string]gopter.Gen{
		"ItemCount": gen.IntRange(0, 50),
		"PageSize":  gen.IntRange(1, 10),
	})

	// Property: ListTableBucketsAll aggregates all buckets across pages
	properties.Property("ListTableBucketsAll aggregates all buckets", prop.ForAll(
		func(input PaginationTestInput) bool {
			// Generate test data
			buckets := make([]types.TableBucketSummary, input.ItemCount)
			now := time.Now()
			for i := 0; i < input.ItemCount; i++ {
				name := fmt.Sprintf("bucket-%d", i)
				arn := fmt.Sprintf("arn:aws:s3tables:us-east-1:123456789012:bucket/%s", name)
				buckets[i] = types.TableBucketSummary{
					Name:      aws.String(name),
					Arn:       aws.String(arn),
					CreatedAt: aws.Time(now),
				}
			}

			mock := &PaginatedMockS3TablesAPI{
				TableBuckets: buckets,
				PageSize:     input.PageSize,
			}
			lister := NewS3TablesLister(mock)

			result, err := lister.ListTableBucketsAll(context.Background(), "")
			if err != nil {
				return false
			}

			// Verify count matches
			if len(result) != input.ItemCount {
				return false
			}

			// Verify all items are present and in order
			for i, bucket := range result {
				expectedName := fmt.Sprintf("bucket-%d", i)
				if bucket.Name != expectedName {
					return false
				}
			}

			return true
		},
		paginationInputGen,
	))

	// Property: ListNamespacesAll aggregates all namespaces across pages
	properties.Property("ListNamespacesAll aggregates all namespaces", prop.ForAll(
		func(input PaginationTestInput) bool {
			// Generate test data
			namespaces := make([]types.NamespaceSummary, input.ItemCount)
			now := time.Now()
			for i := 0; i < input.ItemCount; i++ {
				name := fmt.Sprintf("namespace_%d", i)
				namespaces[i] = types.NamespaceSummary{
					Namespace: []string{name},
					CreatedAt: aws.Time(now),
				}
			}

			mock := &PaginatedMockS3TablesAPI{
				Namespaces: namespaces,
				PageSize:   input.PageSize,
			}
			lister := NewS3TablesLister(mock)

			result, err := lister.ListNamespacesAll(context.Background(), "arn:aws:s3tables:us-east-1:123456789012:bucket/test", "")
			if err != nil {
				return false
			}

			// Verify count matches
			if len(result) != input.ItemCount {
				return false
			}

			// Verify all items are present and in order
			for i, ns := range result {
				expectedName := fmt.Sprintf("namespace_%d", i)
				if ns.Name != expectedName {
					return false
				}
			}

			return true
		},
		paginationInputGen,
	))

	// Property: ListTablesAll aggregates all tables across pages
	properties.Property("ListTablesAll aggregates all tables", prop.ForAll(
		func(input PaginationTestInput) bool {
			// Generate test data
			tables := make([]types.TableSummary, input.ItemCount)
			now := time.Now()
			for i := 0; i < input.ItemCount; i++ {
				name := fmt.Sprintf("table_%d", i)
				arn := fmt.Sprintf("arn:aws:s3tables:us-east-1:123456789012:bucket/test/table/%s", name)
				tables[i] = types.TableSummary{
					Name:      aws.String(name),
					TableARN:  aws.String(arn),
					Namespace: []string{"test_ns"},
					CreatedAt: aws.Time(now),
					Type:      types.TableTypeCustomer,
				}
			}

			mock := &PaginatedMockS3TablesAPI{
				Tables:   tables,
				PageSize: input.PageSize,
			}
			lister := NewS3TablesLister(mock)

			result, err := lister.ListTablesAll(context.Background(), "arn:aws:s3tables:us-east-1:123456789012:bucket/test", "test_ns", "")
			if err != nil {
				return false
			}

			// Verify count matches
			if len(result) != input.ItemCount {
				return false
			}

			// Verify all items are present and in order
			for i, tbl := range result {
				expectedName := fmt.Sprintf("table_%d", i)
				if tbl.Name != expectedName {
					return false
				}
			}

			return true
		},
		paginationInputGen,
	))

	properties.TestingRun(t)
}
