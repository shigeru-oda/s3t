package s3tables

import (
	"fmt"
	"regexp"
)

var (
	// TableBucket: 3-63 characters, lowercase letters, numbers, and hyphens
	tableBucketPattern = regexp.MustCompile(`^[0-9a-z-]+$`)
	// Namespace: 1-255 characters, lowercase letters, numbers, and underscores
	namespacePattern = regexp.MustCompile(`^[0-9a-z_]+$`)
	// Table: 1-255 characters, lowercase letters, numbers, and underscores
	tablePattern = regexp.MustCompile(`^[0-9a-z_]+$`)
)

// ValidationError represents a validation error with field name and reason
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("invalid %s: %s", e.Field, e.Message)
}

// ValidateTableBucket validates a Table Bucket name according to AWS API constraints
// - Length: 3-63 characters
// - Pattern: lowercase letters, numbers, and hyphens only
func ValidateTableBucket(name string) error {
	if len(name) < 3 {
		return &ValidationError{
			Field:   "table-bucket",
			Message: "must be at least 3 characters",
		}
	}
	if len(name) > 63 {
		return &ValidationError{
			Field:   "table-bucket",
			Message: "must be at most 63 characters",
		}
	}
	if !tableBucketPattern.MatchString(name) {
		return &ValidationError{
			Field:   "table-bucket",
			Message: "must contain only lowercase letters, numbers, and hyphens",
		}
	}
	return nil
}

// ValidateNamespace validates a Namespace name according to AWS API constraints
// - Length: 1-255 characters
// - Pattern: lowercase letters, numbers, and underscores only
func ValidateNamespace(name string) error {
	if len(name) < 1 {
		return &ValidationError{
			Field:   "namespace",
			Message: "must be at least 1 character",
		}
	}
	if len(name) > 255 {
		return &ValidationError{
			Field:   "namespace",
			Message: "must be at most 255 characters",
		}
	}
	if !namespacePattern.MatchString(name) {
		return &ValidationError{
			Field:   "namespace",
			Message: "must contain only lowercase letters, numbers, and underscores",
		}
	}
	return nil
}

// ValidateTable validates a Table name according to AWS API constraints
// - Length: 1-255 characters
// - Pattern: lowercase letters, numbers, and underscores only
func ValidateTable(name string) error {
	if len(name) < 1 {
		return &ValidationError{
			Field:   "table",
			Message: "must be at least 1 character",
		}
	}
	if len(name) > 255 {
		return &ValidationError{
			Field:   "table",
			Message: "must be at most 255 characters",
		}
	}
	if !tablePattern.MatchString(name) {
		return &ValidationError{
			Field:   "table",
			Message: "must contain only lowercase letters, numbers, and underscores",
		}
	}
	return nil
}

// ValidateAll validates all input parameters for the create command
func ValidateAll(tableBucket, namespace, table string) error {
	if err := ValidateTableBucket(tableBucket); err != nil {
		return err
	}
	if err := ValidateNamespace(namespace); err != nil {
		return err
	}
	if err := ValidateTable(table); err != nil {
		return err
	}
	return nil
}
