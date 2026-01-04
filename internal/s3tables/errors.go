package s3tables

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3tables/types"
	"github.com/aws/smithy-go"
)

// ErrorType represents the type of error encountered
type ErrorType int

const (
	// ErrorTypeUnknown represents an unknown error type
	ErrorTypeUnknown ErrorType = iota
	// ErrorTypeNotFound represents a resource not found error (404)
	ErrorTypeNotFound
	// ErrorTypeConflict represents a resource conflict error (409)
	ErrorTypeConflict
	// ErrorTypeForbidden represents an authentication/authorization error (403)
	ErrorTypeForbidden
	// ErrorTypeBadRequest represents an invalid input error (400)
	ErrorTypeBadRequest
	// ErrorTypeInternalServer represents a server-side error (500)
	ErrorTypeInternalServer
	// ErrorTypeCredentials represents missing or invalid AWS credentials
	ErrorTypeCredentials
)

// S3TablesError represents a user-friendly error from S3 Tables operations
type S3TablesError struct {
	OriginalErr error
	Operation   string
	Message     string
	Suggestion  string
	Type        ErrorType
}

func (e *S3TablesError) Error() string {
	if e.Suggestion != "" {
		return fmt.Sprintf("Error: %s: %s - %s", e.Operation, e.Message, e.Suggestion)
	}
	return fmt.Sprintf("Error: %s: %s", e.Operation, e.Message)
}

func (e *S3TablesError) Unwrap() error {
	return e.OriginalErr
}

// WrapError converts an AWS API error to a user-friendly S3TablesError
func WrapError(operation string, err error) error {
	if err == nil {
		return nil
	}

	s3tErr := &S3TablesError{
		Operation:   operation,
		OriginalErr: err,
	}

	// Check for specific S3 Tables exception types
	var notFoundErr *types.NotFoundException
	var conflictErr *types.ConflictException
	var forbiddenErr *types.ForbiddenException
	var badRequestErr *types.BadRequestException
	var internalErr *types.InternalServerErrorException

	switch {
	case errors.As(err, &notFoundErr):
		s3tErr.Type = ErrorTypeNotFound
		s3tErr.Message = "resource not found"
		s3tErr.Suggestion = "verify the resource name and try again"

	case errors.As(err, &conflictErr):
		s3tErr.Type = ErrorTypeConflict
		s3tErr.Message = "resource already exists"
		s3tErr.Suggestion = "use a different name or check existing resources"

	case errors.As(err, &forbiddenErr):
		s3tErr.Type = ErrorTypeForbidden
		s3tErr.Message = "access denied"
		s3tErr.Suggestion = "check your AWS credentials and permissions"

	case errors.As(err, &badRequestErr):
		s3tErr.Type = ErrorTypeBadRequest
		s3tErr.Message = "invalid request"
		s3tErr.Suggestion = "check your input parameters"

	case errors.As(err, &internalErr):
		s3tErr.Type = ErrorTypeInternalServer
		s3tErr.Message = "AWS service error"
		s3tErr.Suggestion = "please retry the operation"

	default:
		// Check for smithy API errors
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			s3tErr = handleAPIError(operation, apiErr, err)
		} else {
			// Check for credential-related errors
			if isCredentialError(err) {
				s3tErr.Type = ErrorTypeCredentials
				s3tErr.Message = "AWS credentials not configured"
				s3tErr.Suggestion = "configure AWS credentials using 'aws configure' or environment variables"
			} else {
				s3tErr.Type = ErrorTypeUnknown
				s3tErr.Message = err.Error()
			}
		}
	}

	return s3tErr
}

// handleAPIError handles smithy API errors
func handleAPIError(operation string, apiErr smithy.APIError, originalErr error) *S3TablesError {
	s3tErr := &S3TablesError{
		Operation:   operation,
		OriginalErr: originalErr,
	}

	code := apiErr.ErrorCode()
	switch code {
	case "NotFoundException":
		s3tErr.Type = ErrorTypeNotFound
		s3tErr.Message = "resource not found"
		s3tErr.Suggestion = "verify the resource name and try again"

	case "ConflictException":
		s3tErr.Type = ErrorTypeConflict
		s3tErr.Message = "resource already exists"
		s3tErr.Suggestion = "use a different name or check existing resources"

	case "ForbiddenException", "AccessDeniedException", "AccessDenied":
		s3tErr.Type = ErrorTypeForbidden
		s3tErr.Message = "access denied"
		s3tErr.Suggestion = "check your AWS credentials and permissions"

	case "BadRequestException", "ValidationException":
		s3tErr.Type = ErrorTypeBadRequest
		s3tErr.Message = "invalid request"
		if msg := apiErr.ErrorMessage(); msg != "" {
			s3tErr.Message = msg
		}
		s3tErr.Suggestion = "check your input parameters"

	case "InternalServerErrorException", "InternalServerError", "ServiceException":
		s3tErr.Type = ErrorTypeInternalServer
		s3tErr.Message = "AWS service error"
		s3tErr.Suggestion = "please retry the operation"

	case "UnrecognizedClientException", "InvalidSignatureException":
		s3tErr.Type = ErrorTypeCredentials
		s3tErr.Message = "invalid AWS credentials"
		s3tErr.Suggestion = "check your AWS credentials configuration"

	default:
		s3tErr.Type = ErrorTypeUnknown
		s3tErr.Message = apiErr.ErrorMessage()
		if s3tErr.Message == "" {
			s3tErr.Message = code
		}
	}

	return s3tErr
}

// isCredentialError checks if the error is related to AWS credentials
func isCredentialError(err error) bool {
	errStr := err.Error()
	credentialKeywords := []string{
		"no credentials",
		"credential",
		"NoCredentialProviders",
		"SharedConfigProfileNotExist",
		"failed to refresh cached credentials",
	}
	for _, keyword := range credentialKeywords {
		if containsIgnoreCase(errStr, keyword) {
			return true
		}
	}
	return false
}

// containsIgnoreCase checks if s contains substr (case-insensitive)
func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && containsIgnoreCaseHelper(s, substr)))
}

func containsIgnoreCaseHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if equalFoldAt(s, i, substr) {
			return true
		}
	}
	return false
}

func equalFoldAt(s string, start int, substr string) bool {
	for j := 0; j < len(substr); j++ {
		c1 := s[start+j]
		c2 := substr[j]
		if c1 != c2 && toLower(c1) != toLower(c2) {
			return false
		}
	}
	return true
}

func toLower(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		return c + 32
	}
	return c
}

// IsNotFoundError checks if the error is a not found error
func IsNotFoundError(err error) bool {
	var s3tErr *S3TablesError
	if errors.As(err, &s3tErr) {
		return s3tErr.Type == ErrorTypeNotFound
	}
	// Also check original AWS error types
	var nfe *types.NotFoundException
	if errors.As(err, &nfe) {
		return true
	}
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		return apiErr.ErrorCode() == "NotFoundException"
	}
	return false
}

// IsConflictError checks if the error is a conflict error
func IsConflictError(err error) bool {
	var s3tErr *S3TablesError
	if errors.As(err, &s3tErr) {
		return s3tErr.Type == ErrorTypeConflict
	}
	var ce *types.ConflictException
	return errors.As(err, &ce)
}

// IsCredentialError checks if the error is a credential error
func IsCredentialError(err error) bool {
	var s3tErr *S3TablesError
	if errors.As(err, &s3tErr) {
		return s3tErr.Type == ErrorTypeCredentials
	}
	return isCredentialError(err)
}

// GetErrorType returns the error type from an error
func GetErrorType(err error) ErrorType {
	var s3tErr *S3TablesError
	if errors.As(err, &s3tErr) {
		return s3tErr.Type
	}
	return ErrorTypeUnknown
}
