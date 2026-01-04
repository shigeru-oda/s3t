package s3tables

import (
	"regexp"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Property 1: 入力バリデーションの一貫性
// For any input string, the validation function should return true if and only if
// the string matches the AWS API constraints (length and pattern).
// **Validates: Requirements 4.3**

// Reference patterns for property testing
var (
	tableBucketPatternRef = regexp.MustCompile(`^[0-9a-z-]+$`)
	namespacePatternRef   = regexp.MustCompile(`^[0-9a-z_]+$`)
	tablePatternRef       = regexp.MustCompile(`^[0-9a-z_]+$`)
)

// isValidTableBucket is the reference implementation for property testing
func isValidTableBucket(s string) bool {
	return len(s) >= 3 && len(s) <= 63 && tableBucketPatternRef.MatchString(s)
}

// isValidNamespace is the reference implementation for property testing
func isValidNamespace(s string) bool {
	return len(s) >= 1 && len(s) <= 255 && namespacePatternRef.MatchString(s)
}

// isValidTable is the reference implementation for property testing
func isValidTable(s string) bool {
	return len(s) >= 1 && len(s) <= 255 && tablePatternRef.MatchString(s)
}

// TestPropertyTableBucketValidation tests that ValidateTableBucket returns nil
// if and only if the input matches AWS API constraints
// Feature: s3tables-cli, Property 1: 入力バリデーションの一貫性
func TestPropertyTableBucketValidation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Generator for valid table bucket names (3-63 chars, [0-9a-z-])
	validTableBucketGen := gen.RegexMatch(`[0-9a-z-]{3,63}`)

	// Generator for arbitrary strings to test invalid inputs
	arbitraryStringGen := gen.AnyString()

	// Property: Valid inputs should pass validation
	properties.Property("valid table bucket names pass validation", prop.ForAll(
		func(name string) bool {
			err := ValidateTableBucket(name)
			expected := isValidTableBucket(name)
			return (err == nil) == expected
		},
		validTableBucketGen,
	))

	// Property: Validation result matches reference implementation for any string
	properties.Property("validation consistency with reference for arbitrary strings", prop.ForAll(
		func(name string) bool {
			err := ValidateTableBucket(name)
			expected := isValidTableBucket(name)
			return (err == nil) == expected
		},
		arbitraryStringGen,
	))

	properties.TestingRun(t)
}

// TestPropertyNamespaceValidation tests that ValidateNamespace returns nil
// if and only if the input matches AWS API constraints
// Feature: s3tables-cli, Property 1: 入力バリデーションの一貫性
func TestPropertyNamespaceValidation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Generator for valid namespace names (1-255 chars, [0-9a-z_])
	validNamespaceGen := gen.RegexMatch(`[0-9a-z_]{1,255}`)

	// Generator for arbitrary strings
	arbitraryStringGen := gen.AnyString()

	// Property: Valid inputs should pass validation
	properties.Property("valid namespace names pass validation", prop.ForAll(
		func(name string) bool {
			err := ValidateNamespace(name)
			expected := isValidNamespace(name)
			return (err == nil) == expected
		},
		validNamespaceGen,
	))

	// Property: Validation result matches reference implementation for any string
	properties.Property("validation consistency with reference for arbitrary strings", prop.ForAll(
		func(name string) bool {
			err := ValidateNamespace(name)
			expected := isValidNamespace(name)
			return (err == nil) == expected
		},
		arbitraryStringGen,
	))

	properties.TestingRun(t)
}

// TestPropertyTableValidation tests that ValidateTable returns nil
// if and only if the input matches AWS API constraints
// Feature: s3tables-cli, Property 1: 入力バリデーションの一貫性
func TestPropertyTableValidation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Generator for valid table names (1-255 chars, [0-9a-z_])
	validTableGen := gen.RegexMatch(`[0-9a-z_]{1,255}`)

	// Generator for arbitrary strings
	arbitraryStringGen := gen.AnyString()

	// Property: Valid inputs should pass validation
	properties.Property("valid table names pass validation", prop.ForAll(
		func(name string) bool {
			err := ValidateTable(name)
			expected := isValidTable(name)
			return (err == nil) == expected
		},
		validTableGen,
	))

	// Property: Validation result matches reference implementation for any string
	properties.Property("validation consistency with reference for arbitrary strings", prop.ForAll(
		func(name string) bool {
			err := ValidateTable(name)
			expected := isValidTable(name)
			return (err == nil) == expected
		},
		arbitraryStringGen,
	))

	properties.TestingRun(t)
}
