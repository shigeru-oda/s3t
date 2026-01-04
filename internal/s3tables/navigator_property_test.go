package s3tables

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3tables/types"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// TestNavigationLevelString tests NavigationLevel.String() method
func TestNavigationLevelString(t *testing.T) {
	tests := []struct {
		level    NavigationLevel
		expected string
	}{
		{LevelTableBucket, "TableBucket"},
		{LevelNamespace, "Namespace"},
		{LevelTable, "Table"},
		{NavigationLevel(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.level.String(); got != tt.expected {
				t.Errorf("NavigationLevel.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestNavigationActionString tests NavigationAction.String() method
func TestNavigationActionString(t *testing.T) {
	tests := []struct {
		action   NavigationAction
		expected string
	}{
		{ActionSelect, "Select"},
		{ActionBack, "Back"},
		{ActionExit, "Exit"},
		{NavigationAction(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.action.String(); got != tt.expected {
				t.Errorf("NavigationAction.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestGetState tests GetState method
func TestGetState(t *testing.T) {
	lister := NewS3TablesLister(&PaginatedMockS3TablesAPI{})
	selector := &MockInteractiveSelector{}
	controller := NewNavigationController(lister, selector)

	state := controller.GetState()
	if state == nil {
		t.Error("GetState() returned nil")
	}
	if state != controller.state {
		t.Error("GetState() returned different state object")
	}
}

// MockInteractiveSelector for testing navigation
type MockInteractiveSelector struct {
	SelectWithFilterFunc func(label string, items []string, showBack bool) (*SelectionResult, error)
	CallCount            int
	CallHistory          []MockSelectorCall
}

// MockSelectorCall records a call to SelectWithFilter
type MockSelectorCall struct {
	Label    string
	Items    []string
	ShowBack bool
}

func (m *MockInteractiveSelector) SelectWithFilter(label string, items []string, showBack bool) (*SelectionResult, error) {
	m.CallCount++
	m.CallHistory = append(m.CallHistory, MockSelectorCall{Label: label, Items: items, ShowBack: showBack})
	if m.SelectWithFilterFunc != nil {
		return m.SelectWithFilterFunc(label, items, showBack)
	}
	// Default: select first item
	if len(items) > 0 {
		return &SelectionResult{Selected: items[0], Action: ActionSelect}, nil
	}
	return &SelectionResult{Action: ActionBack}, nil
}

// NavigationTestInput represents the input for navigation property tests
type NavigationTestInput struct {
	BucketCount    int
	NamespaceCount int
	TableCount     int
}

// TestPropertyNavigationLevelDeterminesCorrectResourceListing tests Property 1
// Feature: s3t-list, Property 1: Navigation level determines correct resource listing
// **Validates: Requirements 1.1, 1.2, 2.1, 2.2, 3.1, 3.2**
func TestPropertyNavigationLevelDeterminesCorrectResourceListing(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Generator for navigation test input
	navigationInputGen := gen.Struct(reflect.TypeOf(NavigationTestInput{}), map[string]gopter.Gen{
		"BucketCount":    gen.IntRange(1, 10),
		"NamespaceCount": gen.IntRange(1, 10),
		"TableCount":     gen.IntRange(1, 10),
	})

	// Property: Starting at LevelTableBucket calls ListTableBuckets and returns TableBucketInfo
	properties.Property("LevelTableBucket fetches table buckets", prop.ForAll(
		func(input NavigationTestInput) bool {
			// Generate test data
			buckets := make([]types.TableBucketSummary, input.BucketCount)
			now := time.Now()
			for i := 0; i < input.BucketCount; i++ {
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
				PageSize:     100,
			}
			lister := NewS3TablesLister(mock)

			callCount := 0
			selector := &MockInteractiveSelector{
				SelectWithFilterFunc: func(label string, items []string, showBack bool) (*SelectionResult, error) {
					callCount++
					// Verify correct label
					if label != "Select Table Bucket" {
						return nil, fmt.Errorf("unexpected label: %s", label)
					}
					// Verify correct number of items
					if len(items) != input.BucketCount {
						return nil, fmt.Errorf("expected %d items, got %d", input.BucketCount, len(items))
					}
					// Return back to exit navigation
					return &SelectionResult{Action: ActionBack}, nil
				},
			}

			controller := NewNavigationController(lister, selector)
			err := controller.Navigate(context.Background(), LevelTableBucket)
			if err != nil {
				return false
			}

			// Verify selector was called
			return callCount == 1
		},
		navigationInputGen,
	))

	// Property: Starting at LevelNamespace calls ListNamespaces and returns NamespaceInfo
	properties.Property("LevelNamespace fetches namespaces", prop.ForAll(
		func(input NavigationTestInput) bool {
			// Generate test data
			namespaces := make([]types.NamespaceSummary, input.NamespaceCount)
			now := time.Now()
			for i := 0; i < input.NamespaceCount; i++ {
				name := fmt.Sprintf("namespace_%d", i)
				namespaces[i] = types.NamespaceSummary{
					Namespace: []string{name},
					CreatedAt: aws.Time(now),
				}
			}

			mock := &PaginatedMockS3TablesAPI{
				Namespaces: namespaces,
				PageSize:   100,
			}
			lister := NewS3TablesLister(mock)

			callCount := 0
			selector := &MockInteractiveSelector{
				SelectWithFilterFunc: func(label string, items []string, showBack bool) (*SelectionResult, error) {
					callCount++
					// Verify correct label
					if label != "Select Namespace" {
						return nil, fmt.Errorf("unexpected label: %s", label)
					}
					// Verify correct number of items
					if len(items) != input.NamespaceCount {
						return nil, fmt.Errorf("expected %d items, got %d", input.NamespaceCount, len(items))
					}
					// Return back to exit navigation
					return &SelectionResult{Action: ActionBack}, nil
				},
			}

			controller := NewNavigationController(lister, selector)
			// Set initial state for namespace navigation
			controller.SetInitialState("test-bucket", "arn:aws:s3tables:us-east-1:123456789012:bucket/test-bucket", "")
			controller.state.Level = LevelNamespace

			// Navigate from namespace level
			action, err := controller.navigateNamespaces(context.Background())
			if err != nil {
				return false
			}

			// Verify selector was called and action is back
			return callCount == 1 && action == ActionBack
		},
		navigationInputGen,
	))

	// Property: Starting at LevelTable calls ListTables and returns TableInfo
	properties.Property("LevelTable fetches tables", prop.ForAll(
		func(input NavigationTestInput) bool {
			// Generate test data
			tables := make([]types.TableSummary, input.TableCount)
			now := time.Now()
			for i := 0; i < input.TableCount; i++ {
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
				PageSize: 100,
			}
			lister := NewS3TablesLister(mock)

			callCount := 0
			selector := &MockInteractiveSelector{
				SelectWithFilterFunc: func(label string, items []string, showBack bool) (*SelectionResult, error) {
					callCount++
					// Verify correct label
					if label != "Select Table" {
						return nil, fmt.Errorf("unexpected label: %s", label)
					}
					// Verify correct number of items
					if len(items) != input.TableCount {
						return nil, fmt.Errorf("expected %d items, got %d", input.TableCount, len(items))
					}
					// Return back to exit navigation
					return &SelectionResult{Action: ActionBack}, nil
				},
			}

			controller := NewNavigationController(lister, selector)
			// Set initial state for table navigation
			controller.SetInitialState("test-bucket", "arn:aws:s3tables:us-east-1:123456789012:bucket/test-bucket", "test_ns")
			controller.state.Level = LevelTable

			// Navigate from table level
			action, err := controller.navigateTables(context.Background())
			if err != nil {
				return false
			}

			// Verify selector was called and action is back
			return callCount == 1 && action == ActionBack
		},
		navigationInputGen,
	))

	properties.TestingRun(t)
}

// TestPropertySelectionAdvancesNavigationToNextLevel tests Property 2
// Feature: s3t-list, Property 2: Selection advances navigation to next level
// **Validates: Requirements 1.3, 2.3**
func TestPropertySelectionAdvancesNavigationToNextLevel(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Generator for navigation test input
	navigationInputGen := gen.Struct(reflect.TypeOf(NavigationTestInput{}), map[string]gopter.Gen{
		"BucketCount":    gen.IntRange(1, 10),
		"NamespaceCount": gen.IntRange(1, 10),
		"TableCount":     gen.IntRange(1, 10),
	})

	// Property: Selection at LevelTableBucket advances to LevelNamespace
	// We test this by calling navigateTableBuckets directly and checking the returned action
	properties.Property("Selection at TableBucket advances to Namespace", prop.ForAll(
		func(input NavigationTestInput) bool {
			// Generate test data
			buckets := make([]types.TableBucketSummary, input.BucketCount)
			now := time.Now()

			for i := 0; i < input.BucketCount; i++ {
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
				PageSize:     100,
			}
			lister := NewS3TablesLister(mock)

			selector := &MockInteractiveSelector{
				SelectWithFilterFunc: func(label string, items []string, showBack bool) (*SelectionResult, error) {
					// Select first item
					return &SelectionResult{Selected: items[0], Action: ActionSelect}, nil
				},
			}

			controller := NewNavigationController(lister, selector)
			controller.state.Level = LevelTableBucket

			// Call navigateTableBuckets directly
			action, err := controller.navigateTableBuckets(context.Background())
			if err != nil {
				return false
			}

			// Verify action is Select (which triggers level advancement in Navigate loop)
			// and that the selected bucket info is stored
			return action == ActionSelect &&
				controller.state.SelectedBucket != "" &&
				controller.state.SelectedBucketARN != ""
		},
		navigationInputGen,
	))

	// Property: Selection at LevelNamespace advances to LevelTable
	// We test this by calling navigateNamespaces directly and checking the returned action
	properties.Property("Selection at Namespace advances to Table", prop.ForAll(
		func(input NavigationTestInput) bool {
			// Generate test data
			namespaces := make([]types.NamespaceSummary, input.NamespaceCount)
			now := time.Now()

			for i := 0; i < input.NamespaceCount; i++ {
				name := fmt.Sprintf("namespace_%d", i)
				namespaces[i] = types.NamespaceSummary{
					Namespace: []string{name},
					CreatedAt: aws.Time(now),
				}
			}

			mock := &PaginatedMockS3TablesAPI{
				Namespaces: namespaces,
				PageSize:   100,
			}
			lister := NewS3TablesLister(mock)

			selector := &MockInteractiveSelector{
				SelectWithFilterFunc: func(label string, items []string, showBack bool) (*SelectionResult, error) {
					// Select first item
					return &SelectionResult{Selected: items[0], Action: ActionSelect}, nil
				},
			}

			controller := NewNavigationController(lister, selector)
			// Set initial state for namespace navigation
			controller.SetInitialState("test-bucket", "arn:aws:s3tables:us-east-1:123456789012:bucket/test-bucket", "")
			controller.state.Level = LevelNamespace

			// Call navigateNamespaces directly
			action, err := controller.navigateNamespaces(context.Background())
			if err != nil {
				return false
			}

			// Verify action is Select (which triggers level advancement in Navigate loop)
			// and that the selected namespace is stored
			return action == ActionSelect && controller.state.SelectedNamespace != ""
		},
		navigationInputGen,
	))

	// Property: Full navigation flow advances through all levels
	properties.Property("Full navigation advances through all levels", prop.ForAll(
		func(input NavigationTestInput) bool {
			// Generate test data
			buckets := make([]types.TableBucketSummary, input.BucketCount)
			namespaces := make([]types.NamespaceSummary, input.NamespaceCount)
			tables := make([]types.TableSummary, input.TableCount)
			now := time.Now()

			for i := 0; i < input.BucketCount; i++ {
				name := fmt.Sprintf("bucket-%d", i)
				arn := fmt.Sprintf("arn:aws:s3tables:us-east-1:123456789012:bucket/%s", name)
				buckets[i] = types.TableBucketSummary{
					Name:      aws.String(name),
					Arn:       aws.String(arn),
					CreatedAt: aws.Time(now),
				}
			}

			for i := 0; i < input.NamespaceCount; i++ {
				name := fmt.Sprintf("namespace_%d", i)
				namespaces[i] = types.NamespaceSummary{
					Namespace: []string{name},
					CreatedAt: aws.Time(now),
				}
			}

			for i := 0; i < input.TableCount; i++ {
				name := fmt.Sprintf("table_%d", i)
				arn := fmt.Sprintf("arn:aws:s3tables:us-east-1:123456789012:bucket/test/table/%s", name)
				tables[i] = types.TableSummary{
					Name:      aws.String(name),
					TableARN:  aws.String(arn),
					Namespace: []string{"namespace_0"},
					CreatedAt: aws.Time(now),
					Type:      types.TableTypeCustomer,
				}
			}

			mock := &PaginatedMockS3TablesAPI{
				TableBuckets: buckets,
				Namespaces:   namespaces,
				Tables:       tables,
				PageSize:     100,
			}
			lister := NewS3TablesLister(mock)

			labelsVisited := []string{}
			selector := &MockInteractiveSelector{
				SelectWithFilterFunc: func(label string, items []string, showBack bool) (*SelectionResult, error) {
					labelsVisited = append(labelsVisited, label)
					// Always select first item
					return &SelectionResult{Selected: items[0], Action: ActionSelect}, nil
				},
			}

			controller := NewNavigationController(lister, selector)
			err := controller.Navigate(context.Background(), LevelTableBucket)
			if err != nil {
				return false
			}

			// Verify we visited all three levels in order
			if len(labelsVisited) != 3 {
				return false
			}
			return labelsVisited[0] == "Select Table Bucket" &&
				labelsVisited[1] == "Select Namespace" &&
				labelsVisited[2] == "Select Table"
		},
		navigationInputGen,
	))

	properties.TestingRun(t)
}

// TestPropertyESCNavigatesBackOneLevel tests Property 5
// Feature: s3t-list, Property 5: ESC navigates back one level
// **Validates: Requirements 7.1, 7.2, 7.3**
func TestPropertyESCNavigatesBackOneLevel(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Generator for navigation test input
	navigationInputGen := gen.Struct(reflect.TypeOf(NavigationTestInput{}), map[string]gopter.Gen{
		"BucketCount":    gen.IntRange(1, 10),
		"NamespaceCount": gen.IntRange(1, 10),
		"TableCount":     gen.IntRange(1, 10),
	})

	// Property: ESC at LevelTable returns to LevelNamespace
	properties.Property("ESC at Table level returns to Namespace level", prop.ForAll(
		func(input NavigationTestInput) bool {
			// Generate test data
			buckets := make([]types.TableBucketSummary, input.BucketCount)
			namespaces := make([]types.NamespaceSummary, input.NamespaceCount)
			tables := make([]types.TableSummary, input.TableCount)
			now := time.Now()

			for i := 0; i < input.BucketCount; i++ {
				name := fmt.Sprintf("bucket-%d", i)
				arn := fmt.Sprintf("arn:aws:s3tables:us-east-1:123456789012:bucket/%s", name)
				buckets[i] = types.TableBucketSummary{
					Name:      aws.String(name),
					Arn:       aws.String(arn),
					CreatedAt: aws.Time(now),
				}
			}

			for i := 0; i < input.NamespaceCount; i++ {
				name := fmt.Sprintf("namespace_%d", i)
				namespaces[i] = types.NamespaceSummary{
					Namespace: []string{name},
					CreatedAt: aws.Time(now),
				}
			}

			for i := 0; i < input.TableCount; i++ {
				name := fmt.Sprintf("table_%d", i)
				arn := fmt.Sprintf("arn:aws:s3tables:us-east-1:123456789012:bucket/test/table/%s", name)
				tables[i] = types.TableSummary{
					Name:      aws.String(name),
					TableARN:  aws.String(arn),
					Namespace: []string{"namespace_0"},
					CreatedAt: aws.Time(now),
					Type:      types.TableTypeCustomer,
				}
			}

			mock := &PaginatedMockS3TablesAPI{
				TableBuckets: buckets,
				Namespaces:   namespaces,
				Tables:       tables,
				PageSize:     100,
			}
			lister := NewS3TablesLister(mock)

			// Track navigation levels visited
			levelsVisited := []string{}
			callCount := 0

			selector := &MockInteractiveSelector{
				SelectWithFilterFunc: func(label string, items []string, showBack bool) (*SelectionResult, error) {
					callCount++
					levelsVisited = append(levelsVisited, label)

					// Navigate: Bucket -> Namespace -> Table, then ESC back to Namespace, then ESC to Bucket, then ESC to exit
					switch callCount {
					case 1: // At Bucket level - select first bucket
						return &SelectionResult{Selected: items[0], Action: ActionSelect}, nil
					case 2: // At Namespace level - select first namespace
						return &SelectionResult{Selected: items[0], Action: ActionSelect}, nil
					case 3: // At Table level - press ESC to go back
						return &SelectionResult{Action: ActionBack}, nil
					case 4: // Back at Namespace level - press ESC to go back
						return &SelectionResult{Action: ActionBack}, nil
					case 5: // Back at Bucket level - press ESC to exit
						return &SelectionResult{Action: ActionBack}, nil
					default:
						return &SelectionResult{Action: ActionBack}, nil
					}
				},
			}

			controller := NewNavigationController(lister, selector)
			err := controller.Navigate(context.Background(), LevelTableBucket)
			if err != nil {
				return false
			}

			// Verify navigation sequence: Bucket -> Namespace -> Table -> Namespace -> Bucket
			expectedLabels := []string{
				"Select Table Bucket",
				"Select Namespace",
				"Select Table",
				"Select Namespace",
				"Select Table Bucket",
			}

			if len(levelsVisited) != len(expectedLabels) {
				return false
			}

			for i, expected := range expectedLabels {
				if levelsVisited[i] != expected {
					return false
				}
			}

			return true
		},
		navigationInputGen,
	))

	// Property: ESC at LevelNamespace returns to LevelTableBucket
	properties.Property("ESC at Namespace level returns to TableBucket level", prop.ForAll(
		func(input NavigationTestInput) bool {
			// Generate test data
			buckets := make([]types.TableBucketSummary, input.BucketCount)
			namespaces := make([]types.NamespaceSummary, input.NamespaceCount)
			now := time.Now()

			for i := 0; i < input.BucketCount; i++ {
				name := fmt.Sprintf("bucket-%d", i)
				arn := fmt.Sprintf("arn:aws:s3tables:us-east-1:123456789012:bucket/%s", name)
				buckets[i] = types.TableBucketSummary{
					Name:      aws.String(name),
					Arn:       aws.String(arn),
					CreatedAt: aws.Time(now),
				}
			}

			for i := 0; i < input.NamespaceCount; i++ {
				name := fmt.Sprintf("namespace_%d", i)
				namespaces[i] = types.NamespaceSummary{
					Namespace: []string{name},
					CreatedAt: aws.Time(now),
				}
			}

			mock := &PaginatedMockS3TablesAPI{
				TableBuckets: buckets,
				Namespaces:   namespaces,
				PageSize:     100,
			}
			lister := NewS3TablesLister(mock)

			// Track navigation levels visited
			levelsVisited := []string{}
			callCount := 0

			selector := &MockInteractiveSelector{
				SelectWithFilterFunc: func(label string, items []string, showBack bool) (*SelectionResult, error) {
					callCount++
					levelsVisited = append(levelsVisited, label)

					// Navigate: Bucket -> Namespace, then ESC back to Bucket, then ESC to exit
					switch callCount {
					case 1: // At Bucket level - select first bucket
						return &SelectionResult{Selected: items[0], Action: ActionSelect}, nil
					case 2: // At Namespace level - press ESC to go back
						return &SelectionResult{Action: ActionBack}, nil
					case 3: // Back at Bucket level - press ESC to exit
						return &SelectionResult{Action: ActionBack}, nil
					default:
						return &SelectionResult{Action: ActionBack}, nil
					}
				},
			}

			controller := NewNavigationController(lister, selector)
			err := controller.Navigate(context.Background(), LevelTableBucket)
			if err != nil {
				return false
			}

			// Verify navigation sequence: Bucket -> Namespace -> Bucket
			expectedLabels := []string{
				"Select Table Bucket",
				"Select Namespace",
				"Select Table Bucket",
			}

			if len(levelsVisited) != len(expectedLabels) {
				return false
			}

			for i, expected := range expectedLabels {
				if levelsVisited[i] != expected {
					return false
				}
			}

			return true
		},
		navigationInputGen,
	))

	// Property: ESC at LevelTableBucket exits application
	properties.Property("ESC at TableBucket level exits application", prop.ForAll(
		func(input NavigationTestInput) bool {
			// Generate test data
			buckets := make([]types.TableBucketSummary, input.BucketCount)
			now := time.Now()

			for i := 0; i < input.BucketCount; i++ {
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
				PageSize:     100,
			}
			lister := NewS3TablesLister(mock)

			callCount := 0
			selector := &MockInteractiveSelector{
				SelectWithFilterFunc: func(label string, items []string, showBack bool) (*SelectionResult, error) {
					callCount++
					// Immediately press ESC at bucket level
					return &SelectionResult{Action: ActionBack}, nil
				},
			}

			controller := NewNavigationController(lister, selector)
			err := controller.Navigate(context.Background(), LevelTableBucket)

			// Verify: no error, only one call (at bucket level), and navigation exited
			return err == nil && callCount == 1
		},
		navigationInputGen,
	))

	properties.TestingRun(t)
}

// TestPropertyBackNavigationUsesCachedData tests Property 6
// Feature: s3t-list, Property 6: Back navigation uses cached data
// **Validates: Requirements 7.4**
func TestPropertyBackNavigationUsesCachedData(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Generator for navigation test input
	navigationInputGen := gen.Struct(reflect.TypeOf(NavigationTestInput{}), map[string]gopter.Gen{
		"BucketCount":    gen.IntRange(1, 10),
		"NamespaceCount": gen.IntRange(1, 10),
		"TableCount":     gen.IntRange(1, 10),
	})

	// Property: After navigating Bucket -> Namespace -> Table, pressing ESC twice should not trigger any API calls
	properties.Property("Back navigation does not trigger API calls", prop.ForAll(
		func(input NavigationTestInput) bool {
			// Generate test data
			buckets := make([]types.TableBucketSummary, input.BucketCount)
			namespaces := make([]types.NamespaceSummary, input.NamespaceCount)
			tables := make([]types.TableSummary, input.TableCount)
			now := time.Now()

			for i := 0; i < input.BucketCount; i++ {
				name := fmt.Sprintf("bucket-%d", i)
				arn := fmt.Sprintf("arn:aws:s3tables:us-east-1:123456789012:bucket/%s", name)
				buckets[i] = types.TableBucketSummary{
					Name:      aws.String(name),
					Arn:       aws.String(arn),
					CreatedAt: aws.Time(now),
				}
			}

			for i := 0; i < input.NamespaceCount; i++ {
				name := fmt.Sprintf("namespace_%d", i)
				namespaces[i] = types.NamespaceSummary{
					Namespace: []string{name},
					CreatedAt: aws.Time(now),
				}
			}

			for i := 0; i < input.TableCount; i++ {
				name := fmt.Sprintf("table_%d", i)
				arn := fmt.Sprintf("arn:aws:s3tables:us-east-1:123456789012:bucket/test/table/%s", name)
				tables[i] = types.TableSummary{
					Name:      aws.String(name),
					TableARN:  aws.String(arn),
					Namespace: []string{"namespace_0"},
					CreatedAt: aws.Time(now),
					Type:      types.TableTypeCustomer,
				}
			}

			// Track API calls
			apiCallCount := 0
			mock := &PaginatedMockS3TablesAPI{
				TableBuckets: buckets,
				Namespaces:   namespaces,
				Tables:       tables,
				PageSize:     100,
				OnListTableBuckets: func() {
					apiCallCount++
				},
				OnListNamespaces: func() {
					apiCallCount++
				},
				OnListTables: func() {
					apiCallCount++
				},
			}
			lister := NewS3TablesLister(mock)

			callCount := 0
			selector := &MockInteractiveSelector{
				SelectWithFilterFunc: func(label string, items []string, showBack bool) (*SelectionResult, error) {
					callCount++

					// Navigate: Bucket -> Namespace -> Table, then ESC back twice, then ESC to exit
					switch callCount {
					case 1: // At Bucket level - select first bucket
						return &SelectionResult{Selected: items[0], Action: ActionSelect}, nil
					case 2: // At Namespace level - select first namespace
						return &SelectionResult{Selected: items[0], Action: ActionSelect}, nil
					case 3: // At Table level - press ESC to go back
						return &SelectionResult{Action: ActionBack}, nil
					case 4: // Back at Namespace level - press ESC to go back
						return &SelectionResult{Action: ActionBack}, nil
					case 5: // Back at Bucket level - press ESC to exit
						return &SelectionResult{Action: ActionBack}, nil
					default:
						return &SelectionResult{Action: ActionBack}, nil
					}
				},
			}

			controller := NewNavigationController(lister, selector)

			// Record API call count before navigation
			initialAPICallCount := apiCallCount

			err := controller.Navigate(context.Background(), LevelTableBucket)
			if err != nil {
				return false
			}

			// API calls should only happen during forward navigation (3 calls: buckets, namespaces, tables)
			// Back navigation should NOT trigger any additional API calls
			// Total API calls should be exactly 3
			totalAPICalls := apiCallCount - initialAPICallCount
			return totalAPICalls == 3
		},
		navigationInputGen,
	))

	// Property: Cached data should be identical to originally fetched data
	properties.Property("Cached data is identical to originally fetched data", prop.ForAll(
		func(input NavigationTestInput) bool {
			// Generate test data
			buckets := make([]types.TableBucketSummary, input.BucketCount)
			namespaces := make([]types.NamespaceSummary, input.NamespaceCount)
			tables := make([]types.TableSummary, input.TableCount)
			now := time.Now()

			for i := 0; i < input.BucketCount; i++ {
				name := fmt.Sprintf("bucket-%d", i)
				arn := fmt.Sprintf("arn:aws:s3tables:us-east-1:123456789012:bucket/%s", name)
				buckets[i] = types.TableBucketSummary{
					Name:      aws.String(name),
					Arn:       aws.String(arn),
					CreatedAt: aws.Time(now),
				}
			}

			for i := 0; i < input.NamespaceCount; i++ {
				name := fmt.Sprintf("namespace_%d", i)
				namespaces[i] = types.NamespaceSummary{
					Namespace: []string{name},
					CreatedAt: aws.Time(now),
				}
			}

			for i := 0; i < input.TableCount; i++ {
				name := fmt.Sprintf("table_%d", i)
				arn := fmt.Sprintf("arn:aws:s3tables:us-east-1:123456789012:bucket/test/table/%s", name)
				tables[i] = types.TableSummary{
					Name:      aws.String(name),
					TableARN:  aws.String(arn),
					Namespace: []string{"namespace_0"},
					CreatedAt: aws.Time(now),
					Type:      types.TableTypeCustomer,
				}
			}

			mock := &PaginatedMockS3TablesAPI{
				TableBuckets: buckets,
				Namespaces:   namespaces,
				Tables:       tables,
				PageSize:     100,
			}
			lister := NewS3TablesLister(mock)

			// Track items shown at each level
			bucketItemsFirstVisit := []string{}
			namespaceItemsFirstVisit := []string{}
			bucketItemsSecondVisit := []string{}
			namespaceItemsSecondVisit := []string{}

			callCount := 0
			selector := &MockInteractiveSelector{
				SelectWithFilterFunc: func(label string, items []string, showBack bool) (*SelectionResult, error) {
					callCount++

					switch callCount {
					case 1: // First visit to Bucket level
						bucketItemsFirstVisit = append([]string{}, items...)
						return &SelectionResult{Selected: items[0], Action: ActionSelect}, nil
					case 2: // First visit to Namespace level
						namespaceItemsFirstVisit = append([]string{}, items...)
						return &SelectionResult{Selected: items[0], Action: ActionSelect}, nil
					case 3: // At Table level - press ESC to go back
						return &SelectionResult{Action: ActionBack}, nil
					case 4: // Second visit to Namespace level (from cache)
						namespaceItemsSecondVisit = append([]string{}, items...)
						return &SelectionResult{Action: ActionBack}, nil
					case 5: // Second visit to Bucket level (from cache)
						bucketItemsSecondVisit = append([]string{}, items...)
						return &SelectionResult{Action: ActionBack}, nil
					default:
						return &SelectionResult{Action: ActionBack}, nil
					}
				},
			}

			controller := NewNavigationController(lister, selector)
			err := controller.Navigate(context.Background(), LevelTableBucket)
			if err != nil {
				return false
			}

			// Verify cached data is identical to originally fetched data
			if len(bucketItemsFirstVisit) != len(bucketItemsSecondVisit) {
				return false
			}
			for i := range bucketItemsFirstVisit {
				if bucketItemsFirstVisit[i] != bucketItemsSecondVisit[i] {
					return false
				}
			}

			if len(namespaceItemsFirstVisit) != len(namespaceItemsSecondVisit) {
				return false
			}
			for i := range namespaceItemsFirstVisit {
				if namespaceItemsFirstVisit[i] != namespaceItemsSecondVisit[i] {
					return false
				}
			}

			return true
		},
		navigationInputGen,
	))

	properties.TestingRun(t)
}

// TestNavigateTableBucketsError tests navigateTableBuckets error handling
func TestNavigateTableBucketsError(t *testing.T) {
	mock := &PaginatedMockS3TablesAPI{
		ListTableBucketsError: fmt.Errorf("api error"),
		PageSize:              10,
	}
	lister := NewS3TablesLister(mock)
	selector := &MockInteractiveSelector{}
	controller := NewNavigationController(lister, selector)

	action, err := controller.navigateTableBuckets(context.Background())
	if err == nil {
		t.Error("navigateTableBuckets() should return error")
	}
	if action != ActionExit {
		t.Errorf("navigateTableBuckets() action = %v, want ActionExit", action)
	}
}

// TestNavigateTableBucketsSelectorError tests navigateTableBuckets selector error
func TestNavigateTableBucketsSelectorError(t *testing.T) {
	now := time.Now()
	mock := &PaginatedMockS3TablesAPI{
		TableBuckets: []types.TableBucketSummary{
			{Name: aws.String("bucket-1"), Arn: aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/bucket-1"), CreatedAt: aws.Time(now)},
		},
		PageSize: 10,
	}
	lister := NewS3TablesLister(mock)
	selector := &MockInteractiveSelector{
		SelectWithFilterFunc: func(label string, items []string, showBack bool) (*SelectionResult, error) {
			return nil, fmt.Errorf("selector error")
		},
	}
	controller := NewNavigationController(lister, selector)

	action, err := controller.navigateTableBuckets(context.Background())
	if err == nil {
		t.Error("navigateTableBuckets() should return error")
	}
	if action != ActionExit {
		t.Errorf("navigateTableBuckets() action = %v, want ActionExit", action)
	}
}

// TestNavigateTableBucketsExit tests navigateTableBuckets exit action
func TestNavigateTableBucketsExit(t *testing.T) {
	now := time.Now()
	mock := &PaginatedMockS3TablesAPI{
		TableBuckets: []types.TableBucketSummary{
			{Name: aws.String("bucket-1"), Arn: aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/bucket-1"), CreatedAt: aws.Time(now)},
		},
		PageSize: 10,
	}
	lister := NewS3TablesLister(mock)
	selector := &MockInteractiveSelector{
		SelectWithFilterFunc: func(label string, items []string, showBack bool) (*SelectionResult, error) {
			return &SelectionResult{Action: ActionExit}, nil
		},
	}
	controller := NewNavigationController(lister, selector)

	action, err := controller.navigateTableBuckets(context.Background())
	if err != nil {
		t.Errorf("navigateTableBuckets() error = %v", err)
	}
	if action != ActionExit {
		t.Errorf("navigateTableBuckets() action = %v, want ActionExit", action)
	}
}

// TestNavigateNamespacesError tests navigateNamespaces error handling
func TestNavigateNamespacesError(t *testing.T) {
	mock := &PaginatedMockS3TablesAPI{
		ListNamespacesError: fmt.Errorf("api error"),
		PageSize:            10,
	}
	lister := NewS3TablesLister(mock)
	selector := &MockInteractiveSelector{}
	controller := NewNavigationController(lister, selector)
	controller.SetInitialState("test-bucket", "arn:aws:s3tables:us-east-1:123456789012:bucket/test-bucket", "")

	action, err := controller.navigateNamespaces(context.Background())
	if err == nil {
		t.Error("navigateNamespaces() should return error")
	}
	if action != ActionExit {
		t.Errorf("navigateNamespaces() action = %v, want ActionExit", action)
	}
}

// TestNavigateNamespacesSelectorError tests navigateNamespaces selector error
func TestNavigateNamespacesSelectorError(t *testing.T) {
	now := time.Now()
	mock := &PaginatedMockS3TablesAPI{
		Namespaces: []types.NamespaceSummary{
			{Namespace: []string{"ns-1"}, CreatedAt: aws.Time(now)},
		},
		PageSize: 10,
	}
	lister := NewS3TablesLister(mock)
	selector := &MockInteractiveSelector{
		SelectWithFilterFunc: func(label string, items []string, showBack bool) (*SelectionResult, error) {
			return nil, fmt.Errorf("selector error")
		},
	}
	controller := NewNavigationController(lister, selector)
	controller.SetInitialState("test-bucket", "arn:aws:s3tables:us-east-1:123456789012:bucket/test-bucket", "")

	action, err := controller.navigateNamespaces(context.Background())
	if err == nil {
		t.Error("navigateNamespaces() should return error")
	}
	if action != ActionExit {
		t.Errorf("navigateNamespaces() action = %v, want ActionExit", action)
	}
}

// TestNavigateNamespacesExit tests navigateNamespaces exit action
func TestNavigateNamespacesExit(t *testing.T) {
	now := time.Now()
	mock := &PaginatedMockS3TablesAPI{
		Namespaces: []types.NamespaceSummary{
			{Namespace: []string{"ns-1"}, CreatedAt: aws.Time(now)},
		},
		PageSize: 10,
	}
	lister := NewS3TablesLister(mock)
	selector := &MockInteractiveSelector{
		SelectWithFilterFunc: func(label string, items []string, showBack bool) (*SelectionResult, error) {
			return &SelectionResult{Action: ActionExit}, nil
		},
	}
	controller := NewNavigationController(lister, selector)
	controller.SetInitialState("test-bucket", "arn:aws:s3tables:us-east-1:123456789012:bucket/test-bucket", "")

	action, err := controller.navigateNamespaces(context.Background())
	if err != nil {
		t.Errorf("navigateNamespaces() error = %v", err)
	}
	if action != ActionExit {
		t.Errorf("navigateNamespaces() action = %v, want ActionExit", action)
	}
}

// TestNavigateTablesError tests navigateTables error handling
func TestNavigateTablesError(t *testing.T) {
	mock := &PaginatedMockS3TablesAPI{
		ListTablesError: fmt.Errorf("api error"),
		PageSize:        10,
	}
	lister := NewS3TablesLister(mock)
	selector := &MockInteractiveSelector{}
	controller := NewNavigationController(lister, selector)
	controller.SetInitialState("test-bucket", "arn:aws:s3tables:us-east-1:123456789012:bucket/test-bucket", "test_ns")

	action, err := controller.navigateTables(context.Background())
	if err == nil {
		t.Error("navigateTables() should return error")
	}
	if action != ActionExit {
		t.Errorf("navigateTables() action = %v, want ActionExit", action)
	}
}

// TestNavigateTablesSelectorError tests navigateTables selector error
func TestNavigateTablesSelectorError(t *testing.T) {
	now := time.Now()
	mock := &PaginatedMockS3TablesAPI{
		Tables: []types.TableSummary{
			{Name: aws.String("table-1"), TableARN: aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test/table/table-1"), Namespace: []string{"test_ns"}, CreatedAt: aws.Time(now), Type: types.TableTypeCustomer},
		},
		PageSize: 10,
	}
	lister := NewS3TablesLister(mock)
	selector := &MockInteractiveSelector{
		SelectWithFilterFunc: func(label string, items []string, showBack bool) (*SelectionResult, error) {
			return nil, fmt.Errorf("selector error")
		},
	}
	controller := NewNavigationController(lister, selector)
	controller.SetInitialState("test-bucket", "arn:aws:s3tables:us-east-1:123456789012:bucket/test-bucket", "test_ns")

	action, err := controller.navigateTables(context.Background())
	if err == nil {
		t.Error("navigateTables() should return error")
	}
	if action != ActionExit {
		t.Errorf("navigateTables() action = %v, want ActionExit", action)
	}
}

// TestNavigateTablesExit tests navigateTables exit action
func TestNavigateTablesExit(t *testing.T) {
	now := time.Now()
	mock := &PaginatedMockS3TablesAPI{
		Tables: []types.TableSummary{
			{Name: aws.String("table-1"), TableARN: aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test/table/table-1"), Namespace: []string{"test_ns"}, CreatedAt: aws.Time(now), Type: types.TableTypeCustomer},
		},
		PageSize: 10,
	}
	lister := NewS3TablesLister(mock)
	selector := &MockInteractiveSelector{
		SelectWithFilterFunc: func(label string, items []string, showBack bool) (*SelectionResult, error) {
			return &SelectionResult{Action: ActionExit}, nil
		},
	}
	controller := NewNavigationController(lister, selector)
	controller.SetInitialState("test-bucket", "arn:aws:s3tables:us-east-1:123456789012:bucket/test-bucket", "test_ns")

	action, err := controller.navigateTables(context.Background())
	if err != nil {
		t.Errorf("navigateTables() error = %v", err)
	}
	if action != ActionExit {
		t.Errorf("navigateTables() action = %v, want ActionExit", action)
	}
}

// TestNavigateFromNamespaceLevel tests Navigate starting from namespace level
func TestNavigateFromNamespaceLevel(t *testing.T) {
	now := time.Now()
	mock := &PaginatedMockS3TablesAPI{
		Namespaces: []types.NamespaceSummary{
			{Namespace: []string{"ns-1"}, CreatedAt: aws.Time(now)},
		},
		Tables: []types.TableSummary{
			{Name: aws.String("table-1"), TableARN: aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test/table/table-1"), Namespace: []string{"ns-1"}, CreatedAt: aws.Time(now), Type: types.TableTypeCustomer},
		},
		PageSize: 10,
	}
	lister := NewS3TablesLister(mock)

	callCount := 0
	selector := &MockInteractiveSelector{
		SelectWithFilterFunc: func(label string, items []string, showBack bool) (*SelectionResult, error) {
			callCount++
			if callCount == 1 {
				// Select namespace
				return &SelectionResult{Selected: items[0], Action: ActionSelect}, nil
			}
			// Select table
			return &SelectionResult{Selected: items[0], Action: ActionSelect}, nil
		},
	}
	controller := NewNavigationController(lister, selector)
	controller.SetInitialState("test-bucket", "arn:aws:s3tables:us-east-1:123456789012:bucket/test-bucket", "")

	err := controller.Navigate(context.Background(), LevelNamespace)
	if err != nil {
		t.Errorf("Navigate() error = %v", err)
	}
	if callCount != 2 {
		t.Errorf("Navigate() callCount = %v, want 2", callCount)
	}
}

// TestNavigateFromTableLevel tests Navigate starting from table level
func TestNavigateFromTableLevel(t *testing.T) {
	now := time.Now()
	mock := &PaginatedMockS3TablesAPI{
		Tables: []types.TableSummary{
			{Name: aws.String("table-1"), TableARN: aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test/table/table-1"), Namespace: []string{"test_ns"}, CreatedAt: aws.Time(now), Type: types.TableTypeCustomer},
		},
		PageSize: 10,
	}
	lister := NewS3TablesLister(mock)

	selector := &MockInteractiveSelector{
		SelectWithFilterFunc: func(label string, items []string, showBack bool) (*SelectionResult, error) {
			return &SelectionResult{Selected: items[0], Action: ActionSelect}, nil
		},
	}
	controller := NewNavigationController(lister, selector)
	controller.SetInitialState("test-bucket", "arn:aws:s3tables:us-east-1:123456789012:bucket/test-bucket", "test_ns")

	err := controller.Navigate(context.Background(), LevelTable)
	if err != nil {
		t.Errorf("Navigate() error = %v", err)
	}
}

// TestNavigateNamespaceExitFromNavigate tests Navigate exit from namespace level
func TestNavigateNamespaceExitFromNavigate(t *testing.T) {
	now := time.Now()
	mock := &PaginatedMockS3TablesAPI{
		TableBuckets: []types.TableBucketSummary{
			{Name: aws.String("bucket-1"), Arn: aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/bucket-1"), CreatedAt: aws.Time(now)},
		},
		Namespaces: []types.NamespaceSummary{
			{Namespace: []string{"ns-1"}, CreatedAt: aws.Time(now)},
		},
		PageSize: 10,
	}
	lister := NewS3TablesLister(mock)

	callCount := 0
	selector := &MockInteractiveSelector{
		SelectWithFilterFunc: func(label string, items []string, showBack bool) (*SelectionResult, error) {
			callCount++
			if callCount == 1 {
				// Select bucket
				return &SelectionResult{Selected: items[0], Action: ActionSelect}, nil
			}
			// Exit from namespace
			return &SelectionResult{Action: ActionExit}, nil
		},
	}
	controller := NewNavigationController(lister, selector)

	err := controller.Navigate(context.Background(), LevelTableBucket)
	if err != nil {
		t.Errorf("Navigate() error = %v", err)
	}
}

// TestNavigateTableExitFromNavigate tests Navigate exit from table level
func TestNavigateTableExitFromNavigate(t *testing.T) {
	now := time.Now()
	mock := &PaginatedMockS3TablesAPI{
		TableBuckets: []types.TableBucketSummary{
			{Name: aws.String("bucket-1"), Arn: aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/bucket-1"), CreatedAt: aws.Time(now)},
		},
		Namespaces: []types.NamespaceSummary{
			{Namespace: []string{"ns-1"}, CreatedAt: aws.Time(now)},
		},
		Tables: []types.TableSummary{
			{Name: aws.String("table-1"), TableARN: aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test/table/table-1"), Namespace: []string{"ns-1"}, CreatedAt: aws.Time(now), Type: types.TableTypeCustomer},
		},
		PageSize: 10,
	}
	lister := NewS3TablesLister(mock)

	callCount := 0
	selector := &MockInteractiveSelector{
		SelectWithFilterFunc: func(label string, items []string, showBack bool) (*SelectionResult, error) {
			callCount++
			if callCount == 1 {
				// Select bucket
				return &SelectionResult{Selected: items[0], Action: ActionSelect}, nil
			}
			if callCount == 2 {
				// Select namespace
				return &SelectionResult{Selected: items[0], Action: ActionSelect}, nil
			}
			// Exit from table
			return &SelectionResult{Action: ActionExit}, nil
		},
	}
	controller := NewNavigationController(lister, selector)

	err := controller.Navigate(context.Background(), LevelTableBucket)
	if err != nil {
		t.Errorf("Navigate() error = %v", err)
	}
}

// TestNavigateTableBackFromNavigate tests Navigate back from table level
func TestNavigateTableBackFromNavigate(t *testing.T) {
	now := time.Now()
	mock := &PaginatedMockS3TablesAPI{
		TableBuckets: []types.TableBucketSummary{
			{Name: aws.String("bucket-1"), Arn: aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/bucket-1"), CreatedAt: aws.Time(now)},
		},
		Namespaces: []types.NamespaceSummary{
			{Namespace: []string{"ns-1"}, CreatedAt: aws.Time(now)},
		},
		Tables: []types.TableSummary{
			{Name: aws.String("table-1"), TableARN: aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test/table/table-1"), Namespace: []string{"ns-1"}, CreatedAt: aws.Time(now), Type: types.TableTypeCustomer},
		},
		PageSize: 10,
	}
	lister := NewS3TablesLister(mock)

	callCount := 0
	selector := &MockInteractiveSelector{
		SelectWithFilterFunc: func(label string, items []string, showBack bool) (*SelectionResult, error) {
			callCount++
			switch callCount {
			case 1:
				// Select bucket
				return &SelectionResult{Selected: items[0], Action: ActionSelect}, nil
			case 2:
				// Select namespace
				return &SelectionResult{Selected: items[0], Action: ActionSelect}, nil
			case 3:
				// Back from table
				return &SelectionResult{Action: ActionBack}, nil
			case 4:
				// Back from namespace
				return &SelectionResult{Action: ActionBack}, nil
			default:
				// Exit from bucket
				return &SelectionResult{Action: ActionBack}, nil
			}
		},
	}
	controller := NewNavigationController(lister, selector)

	err := controller.Navigate(context.Background(), LevelTableBucket)
	if err != nil {
		t.Errorf("Navigate() error = %v", err)
	}
}

// TestNavigateTableBucketError tests Navigate with table bucket error
func TestNavigateTableBucketError(t *testing.T) {
	mock := &PaginatedMockS3TablesAPI{
		ListTableBucketsError: fmt.Errorf("api error"),
		PageSize:              10,
	}
	lister := NewS3TablesLister(mock)
	selector := &MockInteractiveSelector{}
	controller := NewNavigationController(lister, selector)

	err := controller.Navigate(context.Background(), LevelTableBucket)
	if err == nil {
		t.Error("Navigate() should return error")
	}
}

// TestNavigateNamespaceError tests Navigate with namespace error
func TestNavigateNamespaceError(t *testing.T) {
	now := time.Now()
	mock := &PaginatedMockS3TablesAPI{
		TableBuckets: []types.TableBucketSummary{
			{Name: aws.String("bucket-1"), Arn: aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/bucket-1"), CreatedAt: aws.Time(now)},
		},
		ListNamespacesError: fmt.Errorf("api error"),
		PageSize:            10,
	}
	lister := NewS3TablesLister(mock)

	callCount := 0
	selector := &MockInteractiveSelector{
		SelectWithFilterFunc: func(label string, items []string, showBack bool) (*SelectionResult, error) {
			callCount++
			return &SelectionResult{Selected: items[0], Action: ActionSelect}, nil
		},
	}
	controller := NewNavigationController(lister, selector)

	err := controller.Navigate(context.Background(), LevelTableBucket)
	if err == nil {
		t.Error("Navigate() should return error")
	}
}

// TestNavigateTableError tests Navigate with table error
func TestNavigateTableError(t *testing.T) {
	now := time.Now()
	mock := &PaginatedMockS3TablesAPI{
		TableBuckets: []types.TableBucketSummary{
			{Name: aws.String("bucket-1"), Arn: aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/bucket-1"), CreatedAt: aws.Time(now)},
		},
		Namespaces: []types.NamespaceSummary{
			{Namespace: []string{"ns-1"}, CreatedAt: aws.Time(now)},
		},
		ListTablesError: fmt.Errorf("api error"),
		PageSize:        10,
	}
	lister := NewS3TablesLister(mock)

	callCount := 0
	selector := &MockInteractiveSelector{
		SelectWithFilterFunc: func(label string, items []string, showBack bool) (*SelectionResult, error) {
			callCount++
			return &SelectionResult{Selected: items[0], Action: ActionSelect}, nil
		},
	}
	controller := NewNavigationController(lister, selector)

	err := controller.Navigate(context.Background(), LevelTableBucket)
	if err == nil {
		t.Error("Navigate() should return error")
	}
}

// TestNavigateEmptyTableBuckets tests Navigate with empty table buckets
func TestNavigateEmptyTableBuckets(t *testing.T) {
	mock := &PaginatedMockS3TablesAPI{
		TableBuckets: []types.TableBucketSummary{},
		PageSize:     10,
	}
	lister := NewS3TablesLister(mock)
	selector := &MockInteractiveSelector{}
	controller := NewNavigationController(lister, selector)

	err := controller.Navigate(context.Background(), LevelTableBucket)
	if err != nil {
		t.Errorf("Navigate() error = %v", err)
	}
}

// TestNavigateEmptyNamespaces tests Navigate with empty namespaces
func TestNavigateEmptyNamespaces(t *testing.T) {
	now := time.Now()
	mock := &PaginatedMockS3TablesAPI{
		TableBuckets: []types.TableBucketSummary{
			{Name: aws.String("bucket-1"), Arn: aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/bucket-1"), CreatedAt: aws.Time(now)},
		},
		Namespaces: []types.NamespaceSummary{},
		PageSize:   10,
	}
	lister := NewS3TablesLister(mock)

	callCount := 0
	selector := &MockInteractiveSelector{
		SelectWithFilterFunc: func(label string, items []string, showBack bool) (*SelectionResult, error) {
			callCount++
			if callCount == 1 {
				return &SelectionResult{Selected: items[0], Action: ActionSelect}, nil
			}
			return &SelectionResult{Action: ActionBack}, nil
		},
	}
	controller := NewNavigationController(lister, selector)

	err := controller.Navigate(context.Background(), LevelTableBucket)
	if err != nil {
		t.Errorf("Navigate() error = %v", err)
	}
}

// TestNavigateEmptyTables tests Navigate with empty tables
func TestNavigateEmptyTables(t *testing.T) {
	now := time.Now()
	mock := &PaginatedMockS3TablesAPI{
		TableBuckets: []types.TableBucketSummary{
			{Name: aws.String("bucket-1"), Arn: aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/bucket-1"), CreatedAt: aws.Time(now)},
		},
		Namespaces: []types.NamespaceSummary{
			{Namespace: []string{"ns-1"}, CreatedAt: aws.Time(now)},
		},
		Tables:   []types.TableSummary{},
		PageSize: 10,
	}
	lister := NewS3TablesLister(mock)

	callCount := 0
	selector := &MockInteractiveSelector{
		SelectWithFilterFunc: func(label string, items []string, showBack bool) (*SelectionResult, error) {
			callCount++
			if callCount <= 2 {
				return &SelectionResult{Selected: items[0], Action: ActionSelect}, nil
			}
			return &SelectionResult{Action: ActionBack}, nil
		},
	}
	controller := NewNavigationController(lister, selector)

	err := controller.Navigate(context.Background(), LevelTableBucket)
	if err != nil {
		t.Errorf("Navigate() error = %v", err)
	}
}
