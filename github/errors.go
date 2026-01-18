package github

import (
	"errors"
	"time"

	ghlib "github.com/google/go-github/v68/github"
)

// ErrorType represents the type of error.
type ErrorType int

const (
	ErrTypeNetwork ErrorType = iota
	ErrTypeAuth
	ErrTypeRateLimit
	ErrTypeNotFound
	ErrTypeServer
	ErrTypeUnknown
)

// AppError represents an application-level error with additional context.
type AppError struct {
	Type       ErrorType
	Message    string        // User-facing message
	Cause      error         // Underlying error
	Retryable  bool          // Whether the operation can be retried
	RetryAfter time.Duration // How long to wait before retrying
}

// Error implements the error interface.
func (e *AppError) Error() string {
	return e.Message
}

// Unwrap returns the underlying error for errors.Is/As support.
func (e *AppError) Unwrap() error {
	return e.Cause
}

// WrapAPIError wraps a GitHub API error into an AppError.
func WrapAPIError(err error) *AppError {
	if err == nil {
		return nil
	}

	var ghErr *ghlib.ErrorResponse
	if errors.As(err, &ghErr) {
		switch ghErr.Response.StatusCode {
		case 401:
			return &AppError{
				Type:      ErrTypeAuth,
				Message:   "Authentication failed",
				Cause:     err,
				Retryable: false,
			}
		case 403:
			if ghErr.Response.Header.Get("X-RateLimit-Remaining") == "0" {
				return &AppError{
					Type:      ErrTypeRateLimit,
					Message:   "Rate limit exceeded",
					Cause:     err,
					Retryable: true,
				}
			}
			return &AppError{
				Type:      ErrTypeAuth,
				Message:   "Access denied",
				Cause:     err,
				Retryable: false,
			}
		case 404:
			return &AppError{
				Type:      ErrTypeNotFound,
				Message:   "Resource not found",
				Cause:     err,
				Retryable: false,
			}
		case 429:
			return &AppError{
				Type:      ErrTypeRateLimit,
				Message:   "Too many requests",
				Cause:     err,
				Retryable: true,
			}
		default:
			if ghErr.Response.StatusCode >= 500 {
				return &AppError{
					Type:      ErrTypeServer,
					Message:   "GitHub server error",
					Cause:     err,
					Retryable: true,
				}
			}
		}
	}

	return &AppError{
		Type:      ErrTypeUnknown,
		Message:   "Unexpected error",
		Cause:     err,
		Retryable: false,
	}
}

// IsRetryable returns true if the error can be retried.
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Retryable
	}
	return false
}
