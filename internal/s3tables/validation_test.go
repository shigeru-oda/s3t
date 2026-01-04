package s3tables

import "testing"

func TestValidationError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *ValidationError
		expected string
	}{
		{
			name: "table-bucket error",
			err: &ValidationError{
				Field:   "table-bucket",
				Message: "must be at least 3 characters",
			},
			expected: "invalid table-bucket: must be at least 3 characters",
		},
		{
			name: "namespace error",
			err: &ValidationError{
				Field:   "namespace",
				Message: "must contain only lowercase letters",
			},
			expected: "invalid namespace: must contain only lowercase letters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.expected {
				t.Errorf("Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestValidateAll(t *testing.T) {
	tests := []struct {
		name        string
		tableBucket string
		namespace   string
		table       string
		wantErr     bool
		errField    string
	}{
		{
			name:        "all valid",
			tableBucket: "my-bucket",
			namespace:   "my_namespace",
			table:       "my_table",
			wantErr:     false,
		},
		{
			name:        "invalid table bucket",
			tableBucket: "ab",
			namespace:   "my_namespace",
			table:       "my_table",
			wantErr:     true,
			errField:    "table-bucket",
		},
		{
			name:        "invalid namespace",
			tableBucket: "my-bucket",
			namespace:   "invalid-ns",
			table:       "my_table",
			wantErr:     true,
			errField:    "namespace",
		},
		{
			name:        "invalid table",
			tableBucket: "my-bucket",
			namespace:   "my_namespace",
			table:       "invalid-table",
			wantErr:     true,
			errField:    "table",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAll(tt.tableBucket, tt.namespace, tt.table)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				valErr, ok := err.(*ValidationError)
				if !ok {
					t.Errorf("expected *ValidationError, got %T", err)
					return
				}
				if valErr.Field != tt.errField {
					t.Errorf("error field = %q, want %q", valErr.Field, tt.errField)
				}
			}
		})
	}
}
