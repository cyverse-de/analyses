package db

import (
	"errors"
	"testing"
)

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"suffix 'not found'", errors.New("quick launch not found"), true},
		{"contains 'not found:'", errors.New("not found: some detail"), true},
		{"unrelated error", errors.New("connection refused"), false},
		{"exact match", errors.New("not found"), true},
		{"case sensitive - uppercase", errors.New("Not Found"), false},
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
