package db

import (
	"errors"
	"fmt"
	"testing"
)

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"NotFoundError pointer", NewNotFoundError("quick launch", "abc-123"), true},
		{"wrapped NotFoundError", fmt.Errorf("wrapped: %w", NewNotFoundError("user", "bob")), true},
		{"unrelated error", errors.New("connection refused"), false},
		{"plain string error", errors.New("not found"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsNotFound(tt.err)
			if got != tt.want {
				t.Errorf("IsNotFound(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestNotFoundError_Error(t *testing.T) {
	err := NewNotFoundError("quick launch", "id-42")
	msg := err.Error()
	if msg != "quick launch not found: id-42" {
		t.Errorf("Error() = %q, want %q", msg, "quick launch not found: id-42")
	}
}
