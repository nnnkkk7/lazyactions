package github

import (
	"errors"
	"net/http"
	"testing"
	"time"

	ghlib "github.com/google/go-github/v68/github"
)

func TestErrorType_Constants(t *testing.T) {
	// Verify error type constants are distinct
	types := []ErrorType{
		ErrTypeNetwork,
		ErrTypeAuth,
		ErrTypeRateLimit,
		ErrTypeNotFound,
		ErrTypeServer,
		ErrTypeUnknown,
	}

	seen := make(map[ErrorType]bool)
	for _, et := range types {
		if seen[et] {
			t.Errorf("Duplicate ErrorType value: %v", et)
		}
		seen[et] = true
	}
}

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name    string
		appErr  *AppError
		wantMsg string
	}{
		{
			name: "simple message",
			appErr: &AppError{
				Type:    ErrTypeAuth,
				Message: "Authentication failed",
			},
			wantMsg: "Authentication failed",
		},
		{
			name: "message with cause",
			appErr: &AppError{
				Type:    ErrTypeNetwork,
				Message: "Network error",
				Cause:   errors.New("connection refused"),
			},
			wantMsg: "Network error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.appErr.Error()
			if got != tt.wantMsg {
				t.Errorf("AppError.Error() = %v, want %v", got, tt.wantMsg)
			}
		})
	}
}

func TestAppError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")
	appErr := &AppError{
		Type:    ErrTypeUnknown,
		Message: "Wrapped error",
		Cause:   originalErr,
	}

	unwrapped := appErr.Unwrap()
	if unwrapped != originalErr {
		t.Errorf("AppError.Unwrap() = %v, want %v", unwrapped, originalErr)
	}

	// Test errors.Is works through Unwrap
	if !errors.Is(appErr, originalErr) {
		t.Error("errors.Is failed to find original error through Unwrap")
	}
}

func TestAppError_Unwrap_Nil(t *testing.T) {
	appErr := &AppError{
		Type:    ErrTypeUnknown,
		Message: "No cause",
		Cause:   nil,
	}

	unwrapped := appErr.Unwrap()
	if unwrapped != nil {
		t.Errorf("AppError.Unwrap() = %v, want nil", unwrapped)
	}
}

func TestAppError_Fields(t *testing.T) {
	appErr := &AppError{
		Type:       ErrTypeRateLimit,
		Message:    "Rate limit exceeded",
		Cause:      errors.New("429 too many requests"),
		Retryable:  true,
		RetryAfter: 60 * time.Second,
	}

	if appErr.Type != ErrTypeRateLimit {
		t.Errorf("AppError.Type = %v, want %v", appErr.Type, ErrTypeRateLimit)
	}
	if appErr.Message != "Rate limit exceeded" {
		t.Errorf("AppError.Message = %v, want 'Rate limit exceeded'", appErr.Message)
	}
	if !appErr.Retryable {
		t.Errorf("AppError.Retryable = false, want true")
	}
	if appErr.RetryAfter != 60*time.Second {
		t.Errorf("AppError.RetryAfter = %v, want 60s", appErr.RetryAfter)
	}
}

func TestWrapAPIError_Nil(t *testing.T) {
	result := WrapAPIError(nil)
	if result != nil {
		t.Errorf("WrapAPIError(nil) = %v, want nil", result)
	}
}

func TestWrapAPIError_401(t *testing.T) {
	ghErr := &ghlib.ErrorResponse{
		Response: &http.Response{
			StatusCode: 401,
		},
		Message: "Bad credentials",
	}

	appErr := WrapAPIError(ghErr)
	if appErr == nil {
		t.Fatal("WrapAPIError returned nil for 401 error")
	}
	if appErr.Type != ErrTypeAuth {
		t.Errorf("AppError.Type = %v, want %v", appErr.Type, ErrTypeAuth)
	}
	if appErr.Retryable {
		t.Error("401 error should not be retryable")
	}
	if appErr.Cause != ghErr {
		t.Error("AppError.Cause should be the original error")
	}
}

func TestWrapAPIError_403_RateLimit(t *testing.T) {
	header := make(http.Header)
	header.Set("X-RateLimit-Remaining", "0")

	ghErr := &ghlib.ErrorResponse{
		Response: &http.Response{
			StatusCode: 403,
			Header:     header,
		},
		Message: "API rate limit exceeded",
	}

	appErr := WrapAPIError(ghErr)
	if appErr == nil {
		t.Fatal("WrapAPIError returned nil for 403 rate limit error")
	}
	if appErr.Type != ErrTypeRateLimit {
		t.Errorf("AppError.Type = %v, want %v", appErr.Type, ErrTypeRateLimit)
	}
	if !appErr.Retryable {
		t.Error("Rate limit error should be retryable")
	}
}

func TestWrapAPIError_403_Forbidden(t *testing.T) {
	header := make(http.Header)
	header.Set("X-RateLimit-Remaining", "5000")

	ghErr := &ghlib.ErrorResponse{
		Response: &http.Response{
			StatusCode: 403,
			Header:     header,
		},
		Message: "Forbidden",
	}

	appErr := WrapAPIError(ghErr)
	if appErr == nil {
		t.Fatal("WrapAPIError returned nil for 403 forbidden error")
	}
	if appErr.Type != ErrTypeAuth {
		t.Errorf("AppError.Type = %v, want %v", appErr.Type, ErrTypeAuth)
	}
	if appErr.Retryable {
		t.Error("403 forbidden error should not be retryable")
	}
}

func TestWrapAPIError_404(t *testing.T) {
	ghErr := &ghlib.ErrorResponse{
		Response: &http.Response{
			StatusCode: 404,
		},
		Message: "Not Found",
	}

	appErr := WrapAPIError(ghErr)
	if appErr == nil {
		t.Fatal("WrapAPIError returned nil for 404 error")
	}
	if appErr.Type != ErrTypeNotFound {
		t.Errorf("AppError.Type = %v, want %v", appErr.Type, ErrTypeNotFound)
	}
	if appErr.Retryable {
		t.Error("404 error should not be retryable")
	}
}

func TestWrapAPIError_429(t *testing.T) {
	ghErr := &ghlib.ErrorResponse{
		Response: &http.Response{
			StatusCode: 429,
		},
		Message: "Too Many Requests",
	}

	appErr := WrapAPIError(ghErr)
	if appErr == nil {
		t.Fatal("WrapAPIError returned nil for 429 error")
	}
	if appErr.Type != ErrTypeRateLimit {
		t.Errorf("AppError.Type = %v, want %v", appErr.Type, ErrTypeRateLimit)
	}
	if !appErr.Retryable {
		t.Error("429 error should be retryable")
	}
}

func TestWrapAPIError_500(t *testing.T) {
	ghErr := &ghlib.ErrorResponse{
		Response: &http.Response{
			StatusCode: 500,
		},
		Message: "Internal Server Error",
	}

	appErr := WrapAPIError(ghErr)
	if appErr == nil {
		t.Fatal("WrapAPIError returned nil for 500 error")
	}
	if appErr.Type != ErrTypeServer {
		t.Errorf("AppError.Type = %v, want %v", appErr.Type, ErrTypeServer)
	}
	if !appErr.Retryable {
		t.Error("Server error should be retryable")
	}
}

func TestWrapAPIError_502(t *testing.T) {
	ghErr := &ghlib.ErrorResponse{
		Response: &http.Response{
			StatusCode: 502,
		},
		Message: "Bad Gateway",
	}

	appErr := WrapAPIError(ghErr)
	if appErr == nil {
		t.Fatal("WrapAPIError returned nil for 502 error")
	}
	if appErr.Type != ErrTypeServer {
		t.Errorf("AppError.Type = %v, want %v", appErr.Type, ErrTypeServer)
	}
	if !appErr.Retryable {
		t.Error("502 error should be retryable")
	}
}

func TestWrapAPIError_503(t *testing.T) {
	ghErr := &ghlib.ErrorResponse{
		Response: &http.Response{
			StatusCode: 503,
		},
		Message: "Service Unavailable",
	}

	appErr := WrapAPIError(ghErr)
	if appErr == nil {
		t.Fatal("WrapAPIError returned nil for 503 error")
	}
	if appErr.Type != ErrTypeServer {
		t.Errorf("AppError.Type = %v, want %v", appErr.Type, ErrTypeServer)
	}
	if !appErr.Retryable {
		t.Error("503 error should be retryable")
	}
}

func TestWrapAPIError_UnknownStatusCode(t *testing.T) {
	ghErr := &ghlib.ErrorResponse{
		Response: &http.Response{
			StatusCode: 418, // I'm a teapot
		},
		Message: "I'm a teapot",
	}

	appErr := WrapAPIError(ghErr)
	if appErr == nil {
		t.Fatal("WrapAPIError returned nil for unknown error")
	}
	if appErr.Type != ErrTypeUnknown {
		t.Errorf("AppError.Type = %v, want %v", appErr.Type, ErrTypeUnknown)
	}
	if appErr.Retryable {
		t.Error("Unknown error should not be retryable")
	}
}

func TestWrapAPIError_NonGitHubError(t *testing.T) {
	plainErr := errors.New("some network error")

	appErr := WrapAPIError(plainErr)
	if appErr == nil {
		t.Fatal("WrapAPIError returned nil for non-GitHub error")
	}
	if appErr.Type != ErrTypeUnknown {
		t.Errorf("AppError.Type = %v, want %v", appErr.Type, ErrTypeUnknown)
	}
	if appErr.Retryable {
		t.Error("Unknown error should not be retryable")
	}
	if appErr.Cause != plainErr {
		t.Error("AppError.Cause should be the original error")
	}
}

func TestIsRetryable_True(t *testing.T) {
	appErr := &AppError{
		Type:      ErrTypeRateLimit,
		Message:   "Rate limit exceeded",
		Retryable: true,
	}

	if !IsRetryable(appErr) {
		t.Error("IsRetryable should return true for retryable AppError")
	}
}

func TestIsRetryable_False(t *testing.T) {
	appErr := &AppError{
		Type:      ErrTypeAuth,
		Message:   "Authentication failed",
		Retryable: false,
	}

	if IsRetryable(appErr) {
		t.Error("IsRetryable should return false for non-retryable AppError")
	}
}

func TestIsRetryable_NonAppError(t *testing.T) {
	plainErr := errors.New("some error")

	if IsRetryable(plainErr) {
		t.Error("IsRetryable should return false for non-AppError")
	}
}

func TestIsRetryable_WrappedAppError(t *testing.T) {
	appErr := &AppError{
		Type:      ErrTypeServer,
		Message:   "Server error",
		Retryable: true,
	}
	wrappedErr := errors.Join(errors.New("context"), appErr)

	if !IsRetryable(wrappedErr) {
		t.Error("IsRetryable should return true for wrapped retryable AppError")
	}
}

func TestIsRetryable_Nil(t *testing.T) {
	if IsRetryable(nil) {
		t.Error("IsRetryable should return false for nil error")
	}
}
