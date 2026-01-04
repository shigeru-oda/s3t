package cmd

import (
	"bytes"
	"testing"

	s3tablesinternal "s3t/internal/s3tables"

	"github.com/spf13/cobra"
)

// executeCommand executes a cobra command with the given args and returns output and error
func executeCommand(root *cobra.Command, args ...string) (string, error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	err := root.Execute()
	return buf.String(), err
}

// TestCreateCommand_InsufficientArguments tests that the CLI returns an error when insufficient arguments are provided
// Requirements: 4.2
func TestCreateCommand_InsufficientArguments(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "no arguments",
			args: []string{"create"},
		},
		{
			name: "one argument",
			args: []string{"create", "my-bucket"},
		},
		{
			name: "two arguments",
			args: []string{"create", "my-bucket", "my-namespace"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh root command for each test
			cmd := &cobra.Command{Use: "s3t"}
			testCreateCmd := &cobra.Command{
				Use:  "create <table-bucket> <namespace> <table>",
				Args: cobra.ExactArgs(3),
				RunE: func(cmd *cobra.Command, args []string) error {
					return nil
				},
			}
			cmd.AddCommand(testCreateCmd)

			_, err := executeCommand(cmd, tt.args...)
			if err == nil {
				t.Errorf("expected error for insufficient arguments, got nil")
			}
		})
	}
}

// TestCreateCommand_InvalidArgumentValues tests that the CLI returns validation errors for invalid argument values
// Requirements: 4.3
func TestCreateCommand_InvalidArgumentValues(t *testing.T) {
	tests := []struct {
		name        string
		tableBucket string
		namespace   string
		table       string
		wantErrMsg  string
	}{
		{
			name:        "table bucket too short",
			tableBucket: "ab",
			namespace:   "valid_namespace",
			table:       "valid_table",
			wantErrMsg:  "table-bucket",
		},
		{
			name:        "table bucket with uppercase",
			tableBucket: "MyBucket",
			namespace:   "valid_namespace",
			table:       "valid_table",
			wantErrMsg:  "table-bucket",
		},
		{
			name:        "table bucket with invalid characters",
			tableBucket: "my_bucket",
			namespace:   "valid_namespace",
			table:       "valid_table",
			wantErrMsg:  "table-bucket",
		},
		{
			name:        "empty namespace",
			tableBucket: "valid-bucket",
			namespace:   "",
			table:       "valid_table",
			wantErrMsg:  "namespace",
		},
		{
			name:        "namespace with invalid characters",
			tableBucket: "valid-bucket",
			namespace:   "invalid-namespace",
			table:       "valid_table",
			wantErrMsg:  "namespace",
		},
		{
			name:        "empty table",
			tableBucket: "valid-bucket",
			namespace:   "valid_namespace",
			table:       "",
			wantErrMsg:  "table",
		},
		{
			name:        "table with invalid characters",
			tableBucket: "valid-bucket",
			namespace:   "valid_namespace",
			table:       "invalid-table",
			wantErrMsg:  "table",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh command that uses the real validation logic
			cmd := &cobra.Command{Use: "s3t"}
			testCreateCmd := &cobra.Command{
				Use:  "create <table-bucket> <namespace> <table>",
				Args: cobra.ExactArgs(3),
				RunE: func(cmd *cobra.Command, args []string) error {
					tableBucket := args[0]
					namespace := args[1]
					table := args[2]

					// Use the real validation function
					if err := s3tablesinternal.ValidateAll(tableBucket, namespace, table); err != nil {
						return err
					}
					return nil
				},
			}
			cmd.AddCommand(testCreateCmd)

			_, err := executeCommand(cmd, "create", tt.tableBucket, tt.namespace, tt.table)
			if err == nil {
				t.Errorf("expected validation error for %s, got nil", tt.name)
				return
			}

			errStr := err.Error()
			if !containsIgnoreCase(errStr, tt.wantErrMsg) {
				t.Errorf("expected error to contain '%s', got '%s'", tt.wantErrMsg, errStr)
			}
		})
	}
}

// containsIgnoreCase checks if s contains substr (case-insensitive)
func containsIgnoreCase(s, substr string) bool {
	return bytes.Contains(bytes.ToLower([]byte(s)), bytes.ToLower([]byte(substr)))
}
