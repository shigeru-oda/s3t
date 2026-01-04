package s3tables

import (
	"context"
	"errors"
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

	// Error simulation
	ListTableBucketsError error
	ListNamespacesError   error
	ListTablesError       error
	GetTableError         error

	// GetTable response
	GetTableResponse *s3tables.GetTableOutput
}

func (m *PaginatedMockS3TablesAPI) ListTableBuckets(ctx context.Context, params *s3tables.ListTableBucketsInput, optFns ...func(*s3tables.Options)) (*s3tables.ListTableBucketsOutput, error) {
	if m.OnListTableBuckets != nil {
		m.OnListTableBuckets()
	}

	if m.ListTableBucketsError != nil {
		return nil, m.ListTableBucketsError
	}

	startIndex := 0
	if params.ContinuationToken != nil && *params.ContinuationToken != "" {
		_, _ = fmt.Sscanf(*params.ContinuationToken, "%d", &startIndex)
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
	if m.GetTableError != nil {
		return nil, m.GetTableError
	}
	if m.GetTableResponse != nil {
		return m.GetTableResponse, nil
	}
	return nil, &types.NotFoundException{Message: aws.String("not found")}
}

func (m *PaginatedMockS3TablesAPI) CreateTable(ctx context.Context, params *s3tables.CreateTableInput, optFns ...func(*s3tables.Options)) (*s3tables.CreateTableOutput, error) {
	return &s3tables.CreateTableOutput{TableARN: aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test/table/t")}, nil
}

func (m *PaginatedMockS3TablesAPI) ListNamespaces(ctx context.Context, params *s3tables.ListNamespacesInput, optFns ...func(*s3tables.Options)) (*s3tables.ListNamespacesOutput, error) {
	if m.OnListNamespaces != nil {
		m.OnListNamespaces()
	}

	if m.ListNamespacesError != nil {
		return nil, m.ListNamespacesError
	}

	startIndex := 0
	if params.ContinuationToken != nil && *params.ContinuationToken != "" {
		_, _ = fmt.Sscanf(*params.ContinuationToken, "%d", &startIndex)
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

	if m.ListTablesError != nil {
		return nil, m.ListTablesError
	}

	startIndex := 0
	if params.ContinuationToken != nil && *params.ContinuationToken != "" {
		_, _ = fmt.Sscanf(*params.ContinuationToken, "%d", &startIndex)
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

// TestListTableBucketsAllError tests ListTableBucketsAll error handling
func TestListTableBucketsAllError(t *testing.T) {
	mock := &PaginatedMockS3TablesAPI{
		ListTableBucketsError: errors.New("api error"),
		PageSize:              10,
	}
	lister := NewS3TablesLister(mock)

	_, err := lister.ListTableBucketsAll(context.Background(), "")
	if err == nil {
		t.Error("ListTableBucketsAll() should return error")
	}
}

// TestListNamespacesAllError tests ListNamespacesAll error handling
func TestListNamespacesAllError(t *testing.T) {
	mock := &PaginatedMockS3TablesAPI{
		ListNamespacesError: errors.New("api error"),
		PageSize:            10,
	}
	lister := NewS3TablesLister(mock)

	_, err := lister.ListNamespacesAll(context.Background(), "arn:aws:s3tables:us-east-1:123456789012:bucket/test", "")
	if err == nil {
		t.Error("ListNamespacesAll() should return error")
	}
}

// TestListTablesAllError tests ListTablesAll error handling
func TestListTablesAllError(t *testing.T) {
	mock := &PaginatedMockS3TablesAPI{
		ListTablesError: errors.New("api error"),
		PageSize:        10,
	}
	lister := NewS3TablesLister(mock)

	_, err := lister.ListTablesAll(context.Background(), "arn:aws:s3tables:us-east-1:123456789012:bucket/test", "test_ns", "")
	if err == nil {
		t.Error("ListTablesAll() should return error")
	}
}

// TestGetTableDetailsSuccess tests GetTableDetails success case
func TestGetTableDetailsSuccess(t *testing.T) {
	now := time.Now()
	mock := &PaginatedMockS3TablesAPI{
		GetTableResponse: &s3tables.GetTableOutput{
			Name:      aws.String("test-table"),
			TableARN:  aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test/table/test-table"),
			CreatedAt: aws.Time(now),
			Type:      types.TableTypeCustomer,
		},
		PageSize: 10,
	}
	lister := NewS3TablesLister(mock)

	result, err := lister.GetTableDetails(context.Background(), "arn:aws:s3tables:us-east-1:123456789012:bucket/test", "test_ns", "test-table")
	if err != nil {
		t.Errorf("GetTableDetails() error = %v", err)
	}
	if result.Name != "test-table" {
		t.Errorf("GetTableDetails() Name = %v, want test-table", result.Name)
	}
}

// TestGetTableDetailsError tests GetTableDetails error handling
func TestGetTableDetailsError(t *testing.T) {
	mock := &PaginatedMockS3TablesAPI{
		GetTableError: errors.New("api error"),
		PageSize:      10,
	}
	lister := NewS3TablesLister(mock)

	_, err := lister.GetTableDetails(context.Background(), "arn:aws:s3tables:us-east-1:123456789012:bucket/test", "test_ns", "test-table")
	if err == nil {
		t.Error("GetTableDetails() should return error")
	}
}

// TestGetTableBucketARNSuccess tests GetTableBucketARN success case
func TestGetTableBucketARNSuccess(t *testing.T) {
	now := time.Now()
	mock := &PaginatedMockS3TablesAPI{
		TableBuckets: []types.TableBucketSummary{
			{
				Name:      aws.String("test-bucket"),
				Arn:       aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test-bucket"),
				CreatedAt: aws.Time(now),
			},
		},
		PageSize: 10,
	}
	lister := NewS3TablesLister(mock)

	arn, err := lister.GetTableBucketARN(context.Background(), "test-bucket")
	if err != nil {
		t.Errorf("GetTableBucketARN() error = %v", err)
	}
	if arn != "arn:aws:s3tables:us-east-1:123456789012:bucket/test-bucket" {
		t.Errorf("GetTableBucketARN() = %v, want arn:aws:s3tables:us-east-1:123456789012:bucket/test-bucket", arn)
	}
}

// TestGetTableBucketARNNotFound tests GetTableBucketARN not found case
func TestGetTableBucketARNNotFound(t *testing.T) {
	now := time.Now()
	mock := &PaginatedMockS3TablesAPI{
		TableBuckets: []types.TableBucketSummary{
			{
				Name:      aws.String("other-bucket"),
				Arn:       aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/other-bucket"),
				CreatedAt: aws.Time(now),
			},
		},
		PageSize: 10,
	}
	lister := NewS3TablesLister(mock)

	_, err := lister.GetTableBucketARN(context.Background(), "test-bucket")
	if err == nil {
		t.Error("GetTableBucketARN() should return error for not found bucket")
	}
	if !IsNotFoundError(err) {
		t.Errorf("GetTableBucketARN() error should be NotFoundError, got %v", err)
	}
}

// TestGetTableBucketARNListError tests GetTableBucketARN list error case
func TestGetTableBucketARNListError(t *testing.T) {
	mock := &PaginatedMockS3TablesAPI{
		ListTableBucketsError: errors.New("api error"),
		PageSize:              10,
	}
	lister := NewS3TablesLister(mock)

	_, err := lister.GetTableBucketARN(context.Background(), "test-bucket")
	if err == nil {
		t.Error("GetTableBucketARN() should return error")
	}
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

// TestListTableBucketsAllWithPrefix tests ListTableBucketsAll with prefix
func TestListTableBucketsAllWithPrefix(t *testing.T) {
	now := time.Now()
	mock := &PaginatedMockS3TablesAPI{
		TableBuckets: []types.TableBucketSummary{
			{Name: aws.String("test-bucket"), Arn: aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test-bucket"), CreatedAt: aws.Time(now)},
		},
		PageSize: 10,
	}
	lister := NewS3TablesLister(mock)

	result, err := lister.ListTableBucketsAll(context.Background(), "test")
	if err != nil {
		t.Errorf("ListTableBucketsAll() error = %v", err)
	}
	if len(result) != 1 {
		t.Errorf("ListTableBucketsAll() len = %v, want 1", len(result))
	}
}

// TestListNamespacesAllWithPrefix tests ListNamespacesAll with prefix
func TestListNamespacesAllWithPrefix(t *testing.T) {
	now := time.Now()
	mock := &PaginatedMockS3TablesAPI{
		Namespaces: []types.NamespaceSummary{
			{Namespace: []string{"test_ns"}, CreatedAt: aws.Time(now)},
		},
		PageSize: 10,
	}
	lister := NewS3TablesLister(mock)

	result, err := lister.ListNamespacesAll(context.Background(), "arn:aws:s3tables:us-east-1:123456789012:bucket/test", "test")
	if err != nil {
		t.Errorf("ListNamespacesAll() error = %v", err)
	}
	if len(result) != 1 {
		t.Errorf("ListNamespacesAll() len = %v, want 1", len(result))
	}
}

// TestListTablesAllWithPrefix tests ListTablesAll with prefix
func TestListTablesAllWithPrefix(t *testing.T) {
	now := time.Now()
	mock := &PaginatedMockS3TablesAPI{
		Tables: []types.TableSummary{
			{Name: aws.String("test_table"), TableARN: aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test/table/test_table"), Namespace: []string{"test_ns"}, CreatedAt: aws.Time(now), Type: types.TableTypeCustomer},
		},
		PageSize: 10,
	}
	lister := NewS3TablesLister(mock)

	result, err := lister.ListTablesAll(context.Background(), "arn:aws:s3tables:us-east-1:123456789012:bucket/test", "test_ns", "test")
	if err != nil {
		t.Errorf("ListTablesAll() error = %v", err)
	}
	if len(result) != 1 {
		t.Errorf("ListTablesAll() len = %v, want 1", len(result))
	}
}

// TestListNamespacesAllEmptyNamespace tests ListNamespacesAll with empty namespace array
func TestListNamespacesAllEmptyNamespace(t *testing.T) {
	now := time.Now()
	mock := &PaginatedMockS3TablesAPI{
		Namespaces: []types.NamespaceSummary{
			{Namespace: []string{}, CreatedAt: aws.Time(now)},
		},
		PageSize: 10,
	}
	lister := NewS3TablesLister(mock)

	result, err := lister.ListNamespacesAll(context.Background(), "arn:aws:s3tables:us-east-1:123456789012:bucket/test", "")
	if err != nil {
		t.Errorf("ListNamespacesAll() error = %v", err)
	}
	if len(result) != 1 {
		t.Errorf("ListNamespacesAll() len = %v, want 1", len(result))
	}
	if result[0].Name != "" {
		t.Errorf("ListNamespacesAll() Name = %v, want empty string", result[0].Name)
	}
}

// TestListTablesAllEmptyNamespace tests ListTablesAll with empty namespace array
func TestListTablesAllEmptyNamespace(t *testing.T) {
	now := time.Now()
	mock := &PaginatedMockS3TablesAPI{
		Tables: []types.TableSummary{
			{Name: aws.String("test_table"), TableARN: aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test/table/test_table"), Namespace: []string{}, CreatedAt: aws.Time(now), Type: types.TableTypeCustomer},
		},
		PageSize: 10,
	}
	lister := NewS3TablesLister(mock)

	result, err := lister.ListTablesAll(context.Background(), "arn:aws:s3tables:us-east-1:123456789012:bucket/test", "test_ns", "")
	if err != nil {
		t.Errorf("ListTablesAll() error = %v", err)
	}
	if len(result) != 1 {
		t.Errorf("ListTablesAll() len = %v, want 1", len(result))
	}
	if result[0].Namespace != "" {
		t.Errorf("ListTablesAll() Namespace = %v, want empty string", result[0].Namespace)
	}
}
