package common

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestErrorResponse_Error(t *testing.T) {
	e := ErrorResponse{Message: "something went wrong", ErrorCode: "ERR_TEST"}
	got := e.Error()

	var parsed map[string]any
	if err := json.Unmarshal([]byte(got), &parsed); err != nil {
		t.Fatalf("Error() did not return valid JSON: %v", err)
	}
	if parsed["message"] != "something went wrong" {
		t.Errorf("expected message='something went wrong', got %v", parsed["message"])
	}
	if parsed["error_code"] != "ERR_TEST" {
		t.Errorf("expected error_code='ERR_TEST', got %v", parsed["error_code"])
	}
}

func TestErrorResponse_ErrorBytes(t *testing.T) {
	e := ErrorResponse{Message: "test error"}
	b := e.ErrorBytes()

	var parsed map[string]any
	if err := json.Unmarshal(b, &parsed); err != nil {
		t.Fatalf("ErrorBytes() did not return valid JSON: %v", err)
	}
	if parsed["message"] != "test error" {
		t.Errorf("expected message='test error', got %v", parsed["message"])
	}
}

func TestNewErrorResponse(t *testing.T) {
	t.Run("with ErrorResponse input", func(t *testing.T) {
		original := ErrorResponse{Message: "original", ErrorCode: "E1"}
		got := NewErrorResponse(original)
		if got.Message != "original" || got.ErrorCode != "E1" {
			t.Errorf("expected original ErrorResponse back, got %+v", got)
		}
	})

	t.Run("with generic error", func(t *testing.T) {
		got := NewErrorResponse(errors.New("generic error"))
		if got.Message != "generic error" {
			t.Errorf("expected message='generic error', got %q", got.Message)
		}
		if got.ErrorCode != "" {
			t.Errorf("expected empty error_code, got %q", got.ErrorCode)
		}
	})
}
