package s3tables

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3tables/types"
	"github.com/aws/smithy-go"
)

// mockAPIError implements smithy.APIError for testing
type mockAPIError struct {
	code    string
	message string
}

func (e *mockAPIError) Error() string                 { return e.message }
func (e *mockAPIError) ErrorCode() string             { return e.code }
func (e *mockAPIError) ErrorMessage() string          { return e.message }
func (e *mockAPIError) ErrorFault() smithy.ErrorFault { return smithy.FaultUnknown }

func TestWrapError_NotFoundException(t *testing.T) {
	originalErr := &types.NotFoundException{Message: ptrString("resource not found")}
	err := WrapError("GetTableBucket", originalErr)

	s3tErr, ok := err.(*S3TablesError)
	if !ok {
		t.Fatalf("expected *S3TablesError, got %T", err)
	}

	if s3tErr.Type != ErrorTypeNotFound {
		t.Errorf("expected ErrorTypeNotFound, got %v", s3tErr.Type)
	}
	if s3tErr.Operation != "GetTableBucket" {
		t.Errorf("expected operation 'GetTableBucket', got '%s'", s3tErr.Operation)
	}
	if s3tErr.Suggestion == "" {
		t.Error("expected non-empty suggestion")
	}
}

func TestWrapError_ConflictException(t *testing.T) {
	originalErr := &types.ConflictException{Message: ptrString("resource already exists")}
	err := WrapError("CreateTableBucket", originalErr)

	s3tErr, ok := err.(*S3TablesError)
	if !ok {
		t.Fatalf("expected *S3TablesError, got %T", err)
	}

	if s3tErr.Type != ErrorTypeConflict {
		t.Errorf("expected ErrorTypeConflict, got %v", s3tErr.Type)
	}
	if s3tErr.Operation != "CreateTableBucket" {
		t.Errorf("expected operation 'CreateTableBucket', got '%s'", s3tErr.Operation)
	}
}

func TestWrapError_ForbiddenException(t *testing.T) {
	originalErr := &types.ForbiddenException{Message: ptrString("access denied")}
	err := WrapError("CreateNamespace", originalErr)

	s3tErr, ok := err.(*S3TablesError)
	if !ok {
		t.Fatalf("expected *S3TablesError, got %T", err)
	}

	if s3tErr.Type != ErrorTypeForbidden {
		t.Errorf("expected ErrorTypeForbidden, got %v", s3tErr.Type)
	}
	if s3tErr.Suggestion == "" {
		t.Error("expected non-empty suggestion for credential check")
	}
}

func TestWrapError_BadRequestException(t *testing.T) {
	originalErr := &types.BadRequestException{Message: ptrString("invalid parameter")}
	err := WrapError("CreateTable", originalErr)

	s3tErr, ok := err.(*S3TablesError)
	if !ok {
		t.Fatalf("expected *S3TablesError, got %T", err)
	}

	if s3tErr.Type != ErrorTypeBadRequest {
		t.Errorf("expected ErrorTypeBadRequest, got %v", s3tErr.Type)
	}
}

func TestWrapError_InternalServerErrorException(t *testing.T) {
	originalErr := &types.InternalServerErrorException{Message: ptrString("internal error")}
	err := WrapError("GetTable", originalErr)

	s3tErr, ok := err.(*S3TablesError)
	if !ok {
		t.Fatalf("expected *S3TablesError, got %T", err)
	}

	if s3tErr.Type != ErrorTypeInternalServer {
		t.Errorf("expected ErrorTypeInternalServer, got %v", s3tErr.Type)
	}
	if s3tErr.Suggestion == "" {
		t.Error("expected retry suggestion")
	}
}

func TestWrapError_SmithyAPIError(t *testing.T) {
	tests := []struct {
		name         string
		code         string
		message      string
		expectedType ErrorType
	}{
		{"AccessDenied", "AccessDenied", "test error", ErrorTypeForbidden},
		{"AccessDeniedException", "AccessDeniedException", "test error", ErrorTypeForbidden},
		{"ForbiddenException", "ForbiddenException", "test error", ErrorTypeForbidden},
		{"ValidationException", "ValidationException", "test error", ErrorTypeBadRequest},
		{"BadRequestException", "BadRequestException", "test error", ErrorTypeBadRequest},
		{"ServiceException", "ServiceException", "test error", ErrorTypeInternalServer},
		{"InternalServerError", "InternalServerError", "test error", ErrorTypeInternalServer},
		{"InternalServerErrorException", "InternalServerErrorException", "test error", ErrorTypeInternalServer},
		{"UnrecognizedClientException", "UnrecognizedClientException", "test error", ErrorTypeCredentials},
		{"InvalidSignatureException", "InvalidSignatureException", "test error", ErrorTypeCredentials},
		{"NotFoundException", "NotFoundException", "test error", ErrorTypeNotFound},
		{"ConflictException", "ConflictException", "test error", ErrorTypeConflict},
		{"UnknownCode", "UnknownCode", "test error", ErrorTypeUnknown},
		{"UnknownCodeEmptyMessage", "UnknownCode", "", ErrorTypeUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiErr := &mockAPIError{code: tt.code, message: tt.message}
			err := WrapError("TestOperation", apiErr)

			s3tErr, ok := err.(*S3TablesError)
			if !ok {
				t.Fatalf("expected *S3TablesError, got %T", err)
			}

			if s3tErr.Type != tt.expectedType {
				t.Errorf("expected %v, got %v", tt.expectedType, s3tErr.Type)
			}
		})
	}
}

func TestWrapError_CredentialError(t *testing.T) {
	credentialErrors := []string{
		"no credentials found",
		"NoCredentialProviders: no valid providers",
		"SharedConfigProfileNotExist: profile not found",
	}

	for _, errMsg := range credentialErrors {
		t.Run(errMsg, func(t *testing.T) {
			err := WrapError("TestOperation", errors.New(errMsg))

			s3tErr, ok := err.(*S3TablesError)
			if !ok {
				t.Fatalf("expected *S3TablesError, got %T", err)
			}

			if s3tErr.Type != ErrorTypeCredentials {
				t.Errorf("expected ErrorTypeCredentials for '%s', got %v", errMsg, s3tErr.Type)
			}
		})
	}
}

func TestWrapError_NilError(t *testing.T) {
	err := WrapError("TestOperation", nil)
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestWrapError_UnknownError(t *testing.T) {
	// Test with a regular error that is not a credential error
	err := WrapError("TestOperation", errors.New("some unknown error"))

	s3tErr, ok := err.(*S3TablesError)
	if !ok {
		t.Fatalf("expected *S3TablesError, got %T", err)
	}

	if s3tErr.Type != ErrorTypeUnknown {
		t.Errorf("expected ErrorTypeUnknown, got %v", s3tErr.Type)
	}
	if s3tErr.Message != "some unknown error" {
		t.Errorf("expected message 'some unknown error', got '%s'", s3tErr.Message)
	}
}

func TestS3TablesError_Error(t *testing.T) {
	tests := []struct {
		name       string
		err        *S3TablesError
		wantSubstr string
	}{
		{
			name: "with suggestion",
			err: &S3TablesError{
				Operation:  "CreateTableBucket",
				Message:    "access denied",
				Suggestion: "check your credentials",
			},
			wantSubstr: "check your credentials",
		},
		{
			name: "without suggestion",
			err: &S3TablesError{
				Operation: "GetTable",
				Message:   "unknown error",
			},
			wantSubstr: "Error: GetTable: unknown error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errStr := tt.err.Error()
			if !containsIgnoreCase(errStr, tt.wantSubstr) {
				t.Errorf("error string '%s' does not contain '%s'", errStr, tt.wantSubstr)
			}
		})
	}
}

func TestIsNotFoundError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "S3TablesError NotFound",
			err:      &S3TablesError{Type: ErrorTypeNotFound},
			expected: true,
		},
		{
			name:     "types.NotFoundException",
			err:      &types.NotFoundException{},
			expected: true,
		},
		{
			name:     "smithy API error NotFound",
			err:      &mockAPIError{code: "NotFoundException"},
			expected: true,
		},
		{
			name:     "other error",
			err:      errors.New("some error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNotFoundError(tt.err)
			if result != tt.expected {
				t.Errorf("IsNotFoundError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsConflictError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "S3TablesError Conflict",
			err:      &S3TablesError{Type: ErrorTypeConflict},
			expected: true,
		},
		{
			name:     "types.ConflictException",
			err:      &types.ConflictException{},
			expected: true,
		},
		{
			name:     "other error",
			err:      errors.New("some error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsConflictError(tt.err)
			if result != tt.expected {
				t.Errorf("IsConflictError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetErrorType(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected ErrorType
	}{
		{
			name:     "NotFound",
			err:      &S3TablesError{Type: ErrorTypeNotFound},
			expected: ErrorTypeNotFound,
		},
		{
			name:     "Conflict",
			err:      &S3TablesError{Type: ErrorTypeConflict},
			expected: ErrorTypeConflict,
		},
		{
			name:     "Unknown for non-S3TablesError",
			err:      errors.New("generic error"),
			expected: ErrorTypeUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetErrorType(tt.err)
			if result != tt.expected {
				t.Errorf("GetErrorType() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsCredentialError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "S3TablesError Credentials",
			err:      &S3TablesError{Type: ErrorTypeCredentials},
			expected: true,
		},
		{
			name:     "no credentials error",
			err:      errors.New("no credentials found"),
			expected: true,
		},
		{
			name:     "NoCredentialProviders error",
			err:      errors.New("NoCredentialProviders: no valid providers"),
			expected: true,
		},
		{
			name:     "SharedConfigProfileNotExist error",
			err:      errors.New("SharedConfigProfileNotExist: profile not found"),
			expected: true,
		},
		{
			name:     "failed to refresh cached credentials",
			err:      errors.New("failed to refresh cached credentials"),
			expected: true,
		},
		{
			name:     "other error",
			err:      errors.New("some other error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsCredentialError(tt.err)
			if result != tt.expected {
				t.Errorf("IsCredentialError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestS3TablesError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")
	s3tErr := &S3TablesError{
		Operation:   "TestOp",
		OriginalErr: originalErr,
	}

	unwrapped := s3tErr.Unwrap()
	if unwrapped != originalErr {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, originalErr)
	}
}

// Helper function
func ptrString(s string) *string {
	return &s
}
