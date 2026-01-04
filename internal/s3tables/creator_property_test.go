package s3tables

import (
	"context"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3tables"
	"github.com/aws/aws-sdk-go-v2/service/s3tables/types"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Property 2: 既存リソース検出の正確性
// For any resource state (existing or not), the Creator should correctly identify
// whether each resource exists and only attempt creation for non-existing resources.
// **Validates: Requirements 1.2, 2.2, 3.2**

// MockS3TablesAPI is a mock implementation of S3TablesAPI for property testing
type MockS3TablesAPI struct {
	// State tracking
	TableBucketExists bool
	NamespaceExists   bool
	TableExists       bool

	// Call tracking
	ListTableBucketsCalled  bool
	GetTableBucketCalled    bool
	CreateTableBucketCalled bool
	GetNamespaceCalled      bool
	CreateNamespaceCalled   bool
	GetTableCalled          bool
	CreateTableCalled       bool
}

func (m *MockS3TablesAPI) ListTableBuckets(ctx context.Context, params *s3tables.ListTableBucketsInput, optFns ...func(*s3tables.Options)) (*s3tables.ListTableBucketsOutput, error) {
	m.ListTableBucketsCalled = true
	if m.TableBucketExists {
		return &s3tables.ListTableBucketsOutput{
			TableBuckets: []types.TableBucketSummary{
				{
					Name: aws.String("test-bucket"),
					Arn:  aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test-bucket"),
				},
			},
		}, nil
	}
	return &s3tables.ListTableBucketsOutput{
		TableBuckets: []types.TableBucketSummary{},
	}, nil
}

func (m *MockS3TablesAPI) GetTableBucket(ctx context.Context, params *s3tables.GetTableBucketInput, optFns ...func(*s3tables.Options)) (*s3tables.GetTableBucketOutput, error) {
	m.GetTableBucketCalled = true
	if m.TableBucketExists {
		return &s3tables.GetTableBucketOutput{
			Arn: aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test-bucket"),
		}, nil
	}
	return nil, &types.NotFoundException{Message: aws.String("Table bucket not found")}
}

func (m *MockS3TablesAPI) CreateTableBucket(ctx context.Context, params *s3tables.CreateTableBucketInput, optFns ...func(*s3tables.Options)) (*s3tables.CreateTableBucketOutput, error) {
	m.CreateTableBucketCalled = true
	return &s3tables.CreateTableBucketOutput{
		Arn: aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test-bucket"),
	}, nil
}

func (m *MockS3TablesAPI) GetNamespace(ctx context.Context, params *s3tables.GetNamespaceInput, optFns ...func(*s3tables.Options)) (*s3tables.GetNamespaceOutput, error) {
	m.GetNamespaceCalled = true
	if m.NamespaceExists {
		return &s3tables.GetNamespaceOutput{
			Namespace: []string{aws.ToString(params.Namespace)},
		}, nil
	}
	return nil, &types.NotFoundException{Message: aws.String("Namespace not found")}
}

func (m *MockS3TablesAPI) CreateNamespace(ctx context.Context, params *s3tables.CreateNamespaceInput, optFns ...func(*s3tables.Options)) (*s3tables.CreateNamespaceOutput, error) {
	m.CreateNamespaceCalled = true
	return &s3tables.CreateNamespaceOutput{
		Namespace: params.Namespace,
	}, nil
}

func (m *MockS3TablesAPI) GetTable(ctx context.Context, params *s3tables.GetTableInput, optFns ...func(*s3tables.Options)) (*s3tables.GetTableOutput, error) {
	m.GetTableCalled = true
	if m.TableExists {
		return &s3tables.GetTableOutput{
			TableARN: aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test-bucket/table/test-table"),
		}, nil
	}
	return nil, &types.NotFoundException{Message: aws.String("Table not found")}
}

func (m *MockS3TablesAPI) CreateTable(ctx context.Context, params *s3tables.CreateTableInput, optFns ...func(*s3tables.Options)) (*s3tables.CreateTableOutput, error) {
	m.CreateTableCalled = true
	return &s3tables.CreateTableOutput{
		TableARN: aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test-bucket/table/test-table"),
	}, nil
}

// ResourceState represents the existence state of all three resources
type ResourceState struct {
	TableBucketExists bool
	NamespaceExists   bool
	TableExists       bool
}

// TestPropertyExistingResourceDetection tests that the Creator correctly detects
// existing resources and only creates non-existing ones
// Feature: s3tables-cli, Property 2: 既存リソース検出の正確性
func TestPropertyExistingResourceDetection(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Generator for resource state combinations
	resourceStateGen := gen.Struct(reflect.TypeOf(ResourceState{}), map[string]gopter.Gen{
		"TableBucketExists": gen.Bool(),
		"NamespaceExists":   gen.Bool(),
		"TableExists":       gen.Bool(),
	})

	// Property: Creator should only call Create* for non-existing resources
	properties.Property("only creates non-existing resources", prop.ForAll(
		func(state ResourceState) bool {
			mock := &MockS3TablesAPI{
				TableBucketExists: state.TableBucketExists,
				NamespaceExists:   state.NamespaceExists,
				TableExists:       state.TableExists,
			}
			creator := NewS3TablesCreator(mock)

			result, err := creator.Create(context.Background(), "test-bucket", "test_namespace", "test_table")
			if err != nil {
				return false
			}

			// Verify: ListTableBuckets should always be called to check existence
			if !mock.ListTableBucketsCalled {
				return false
			}
			if !mock.GetNamespaceCalled {
				return false
			}
			if !mock.GetTableCalled {
				return false
			}

			// Verify: Create* should only be called if resource doesn't exist
			// TableBucket: CreateTableBucket called iff TableBucket doesn't exist
			if mock.CreateTableBucketCalled != !state.TableBucketExists {
				return false
			}
			// Namespace: CreateNamespace called iff Namespace doesn't exist
			if mock.CreateNamespaceCalled != !state.NamespaceExists {
				return false
			}
			// Table: CreateTable called iff Table doesn't exist
			if mock.CreateTableCalled != !state.TableExists {
				return false
			}

			// Verify: Result flags match creation actions
			if result.TableBucketCreated != !state.TableBucketExists {
				return false
			}
			if result.NamespaceCreated != !state.NamespaceExists {
				return false
			}
			if result.TableCreated != !state.TableExists {
				return false
			}

			return true
		},
		resourceStateGen,
	))

	properties.TestingRun(t)
}

// Property 3: 階層的作成の順序保証
// For any valid input, the Creator should attempt to create resources in the correct order:
// Table Bucket → Namespace → Table, and should not attempt to create child resources
// if parent creation fails.
// **Validates: Requirements 1.1, 2.1, 3.1**

// OrderTrackingMockS3TablesAPI tracks the order of API calls
type OrderTrackingMockS3TablesAPI struct {
	// Call order tracking
	CallOrder []string

	// Failure configuration
	FailTableBucketCreation bool
	FailNamespaceCreation   bool
	FailTableCreation       bool
}

func (m *OrderTrackingMockS3TablesAPI) ListTableBuckets(ctx context.Context, params *s3tables.ListTableBucketsInput, optFns ...func(*s3tables.Options)) (*s3tables.ListTableBucketsOutput, error) {
	m.CallOrder = append(m.CallOrder, "ListTableBuckets")
	// Always return empty to trigger creation
	return &s3tables.ListTableBucketsOutput{
		TableBuckets: []types.TableBucketSummary{},
	}, nil
}

func (m *OrderTrackingMockS3TablesAPI) GetTableBucket(ctx context.Context, params *s3tables.GetTableBucketInput, optFns ...func(*s3tables.Options)) (*s3tables.GetTableBucketOutput, error) {
	m.CallOrder = append(m.CallOrder, "GetTableBucket")
	// Always return not found to trigger creation
	return nil, &types.NotFoundException{Message: aws.String("Table bucket not found")}
}

func (m *OrderTrackingMockS3TablesAPI) CreateTableBucket(ctx context.Context, params *s3tables.CreateTableBucketInput, optFns ...func(*s3tables.Options)) (*s3tables.CreateTableBucketOutput, error) {
	m.CallOrder = append(m.CallOrder, "CreateTableBucket")
	if m.FailTableBucketCreation {
		return nil, &types.InternalServerErrorException{Message: aws.String("Internal server error")}
	}
	return &s3tables.CreateTableBucketOutput{
		Arn: aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test-bucket"),
	}, nil
}

func (m *OrderTrackingMockS3TablesAPI) GetNamespace(ctx context.Context, params *s3tables.GetNamespaceInput, optFns ...func(*s3tables.Options)) (*s3tables.GetNamespaceOutput, error) {
	m.CallOrder = append(m.CallOrder, "GetNamespace")
	// Always return not found to trigger creation
	return nil, &types.NotFoundException{Message: aws.String("Namespace not found")}
}

func (m *OrderTrackingMockS3TablesAPI) CreateNamespace(ctx context.Context, params *s3tables.CreateNamespaceInput, optFns ...func(*s3tables.Options)) (*s3tables.CreateNamespaceOutput, error) {
	m.CallOrder = append(m.CallOrder, "CreateNamespace")
	if m.FailNamespaceCreation {
		return nil, &types.InternalServerErrorException{Message: aws.String("Internal server error")}
	}
	return &s3tables.CreateNamespaceOutput{
		Namespace: params.Namespace,
	}, nil
}

func (m *OrderTrackingMockS3TablesAPI) GetTable(ctx context.Context, params *s3tables.GetTableInput, optFns ...func(*s3tables.Options)) (*s3tables.GetTableOutput, error) {
	m.CallOrder = append(m.CallOrder, "GetTable")
	// Always return not found to trigger creation
	return nil, &types.NotFoundException{Message: aws.String("Table not found")}
}

func (m *OrderTrackingMockS3TablesAPI) CreateTable(ctx context.Context, params *s3tables.CreateTableInput, optFns ...func(*s3tables.Options)) (*s3tables.CreateTableOutput, error) {
	m.CallOrder = append(m.CallOrder, "CreateTable")
	if m.FailTableCreation {
		return nil, &types.InternalServerErrorException{Message: aws.String("Internal server error")}
	}
	return &s3tables.CreateTableOutput{
		TableARN: aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test-bucket/table/test-table"),
	}, nil
}

// FailureScenario represents which resource creation should fail
type FailureScenario struct {
	FailTableBucket bool
	FailNamespace   bool
	FailTable       bool
}

// TestPropertyHierarchicalCreationOrder tests that resources are created in the correct order
// and child resources are not created if parent creation fails
// Feature: s3tables-cli, Property 3: 階層的作成の順序保証
func TestPropertyHierarchicalCreationOrder(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Generator for failure scenarios
	failureScenarioGen := gen.Struct(reflect.TypeOf(FailureScenario{}), map[string]gopter.Gen{
		"FailTableBucket": gen.Bool(),
		"FailNamespace":   gen.Bool(),
		"FailTable":       gen.Bool(),
	})

	// Property: Resources are created in hierarchical order and failures stop child creation
	properties.Property("hierarchical creation order is maintained", prop.ForAll(
		func(scenario FailureScenario) bool {
			mock := &OrderTrackingMockS3TablesAPI{
				CallOrder:               make([]string, 0),
				FailTableBucketCreation: scenario.FailTableBucket,
				FailNamespaceCreation:   scenario.FailNamespace,
				FailTableCreation:       scenario.FailTable,
			}
			creator := NewS3TablesCreator(mock)

			_, err := creator.Create(context.Background(), "test-bucket", "test_namespace", "test_table")

			// Verify call order based on failure scenario
			callOrder := mock.CallOrder

			// ListTableBuckets should always be called first
			if len(callOrder) < 1 || callOrder[0] != "ListTableBuckets" {
				return false
			}

			// CreateTableBucket should be called second (since ListTableBuckets returns empty)
			if len(callOrder) < 2 || callOrder[1] != "CreateTableBucket" {
				return false
			}

			// If TableBucket creation fails, no further calls should be made
			if scenario.FailTableBucket {
				if err == nil {
					return false // Should have returned an error
				}
				// Should stop after CreateTableBucket
				if len(callOrder) != 2 {
					return false
				}
				return true
			}

			// GetNamespace should be called third
			if len(callOrder) < 3 || callOrder[2] != "GetNamespace" {
				return false
			}

			// CreateNamespace should be called fourth
			if len(callOrder) < 4 || callOrder[3] != "CreateNamespace" {
				return false
			}

			// If Namespace creation fails, no further calls should be made
			if scenario.FailNamespace {
				if err == nil {
					return false // Should have returned an error
				}
				// Should stop after CreateNamespace
				if len(callOrder) != 4 {
					return false
				}
				return true
			}

			// GetTable should be called fifth
			if len(callOrder) < 5 || callOrder[4] != "GetTable" {
				return false
			}

			// CreateTable should be called sixth
			if len(callOrder) < 6 || callOrder[5] != "CreateTable" {
				return false
			}

			// If Table creation fails, should return error
			if scenario.FailTable {
				if err == nil {
					return false // Should have returned an error
				}
				// Should have exactly 6 calls
				if len(callOrder) != 6 {
					return false
				}
				return true
			}

			// If no failures, should complete successfully with exactly 6 calls
			if err != nil {
				return false
			}
			if len(callOrder) != 6 {
				return false
			}

			// Verify the complete order: ListTableBuckets/Create pairs in hierarchical order
			expectedOrder := []string{
				"ListTableBuckets", "CreateTableBucket",
				"GetNamespace", "CreateNamespace",
				"GetTable", "CreateTable",
			}
			for i, expected := range expectedOrder {
				if callOrder[i] != expected {
					return false
				}
			}

			return true
		},
		failureScenarioGen,
	))

	properties.TestingRun(t)
}
