package poller_test

import (
	"errors"
	"testing"

	"github.com/ole-larsen/binance-subscriber/internal/poller"
)

func TestError_Error(t *testing.T) {
	// Create a standard error
	stdErr := errors.New("something went wrong")
	// Create a custom Error instance
	customErr := poller.NewError(stdErr)

	// Use errors.As to perform the type assertion
	var pollerErr *poller.Error
	if !errors.As(customErr, &pollerErr) {
		// Type assertion failed
		t.Fatalf("expected *Error, got %T", customErr)
	}

	// Test Error method
	expectedMsg := "[poller]: something went wrong"
	if pollerErr.Error() != expectedMsg {
		t.Errorf("expected error message %q, got %q", expectedMsg, pollerErr.Error())
	}
}

func TestError_Unwrap(t *testing.T) {
	// Create a standard error
	stdErr := errors.New("something went wrong")
	// Create a custom Error instance
	customErr := poller.NewError(stdErr)

	// Use errors.As to perform the type assertion
	var pollerErr *poller.Error
	if !errors.As(customErr, &pollerErr) {
		// Type assertion failed
		t.Fatalf("expected *Error, got %T", customErr)
	}

	// Test Unwrap method
	if !errors.Is(pollerErr.Unwrap(), stdErr) {
		t.Errorf("expected %v, got %v", stdErr, pollerErr.Unwrap())
	}
}

func TestNewError(t *testing.T) {
	// Test with a non-nil error
	stdErr := errors.New("something went wrong")

	err := poller.NewError(stdErr)
	if err == nil {
		t.Fatal("expected non-nil error")
	}

	// Use errors.As to perform the type assertion
	var pollerErr *poller.Error
	if !errors.As(err, &pollerErr) {
		// Type assertion failed
		t.Fatalf("expected *Error, got %T", err)
	}

	// Ensure the underlying error is the same
	if !errors.Is(pollerErr.Unwrap(), stdErr) {
		t.Errorf("expected %v, got %v", stdErr, pollerErr.Unwrap())
	}

	// Test with a nil error
	err = poller.NewError(nil)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}
