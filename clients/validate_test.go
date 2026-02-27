package clients

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newDataInfoServer(t *testing.T, handler http.HandlerFunc) *DataInfoClient {
	t.Helper()
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	c, err := NewDataInfoClient(ts.URL, nil)
	if err != nil {
		t.Fatalf("NewDataInfoClient: %v", err)
	}
	return c
}

func accessibleHandler(paths map[string]any) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"paths": paths}) //nolint:errcheck // test server write
	}
}

func TestValidateSubmission(t *testing.T) {
	t.Run("nil app returns nil", func(t *testing.T) {
		err := ValidateSubmission(nil, &ValidationRequest{App: nil, Config: map[string]any{}})
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("nil config returns nil", func(t *testing.T) {
		err := ValidateSubmission(nil, &ValidationRequest{App: map[string]any{}, Config: nil})
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("no input params skips validation", func(t *testing.T) {
		app := map[string]any{
			"groups": []any{
				map[string]any{
					"parameters": []any{
						map[string]any{"id": "p1", "type": "TextInput"},
					},
				},
			},
		}
		config := map[string]any{"p1": "some text"}
		err := ValidateSubmission(nil, &ValidationRequest{App: app, Config: config, User: "user"})
		if err != nil {
			t.Fatalf("expected nil for non-input params, got %v", err)
		}
	})

	t.Run("accessible paths pass", func(t *testing.T) {
		diClient := newDataInfoServer(t, accessibleHandler(map[string]any{
			"/iplant/home/user/file.txt": map[string]any{},
		}))

		app := map[string]any{
			"groups": []any{
				map[string]any{
					"parameters": []any{
						map[string]any{"id": "p1", "type": "FileInput"},
					},
				},
			},
		}
		config := map[string]any{"p1": "/iplant/home/user/file.txt"}

		err := ValidateSubmission(diClient, &ValidationRequest{
			App: app, Config: config, User: "user",
		})
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("inaccessible paths fail", func(t *testing.T) {
		diClient := newDataInfoServer(t, accessibleHandler(map[string]any{}))

		app := map[string]any{
			"groups": []any{
				map[string]any{
					"parameters": []any{
						map[string]any{"id": "p1", "type": "FileInput"},
					},
				},
			},
		}
		config := map[string]any{"p1": "/iplant/home/user/file.txt"}

		err := ValidateSubmission(diClient, &ValidationRequest{
			App: app, Config: config, User: "user",
		})
		if err == nil {
			t.Fatal("expected error for inaccessible paths")
		}
		if !strings.Contains(err.Error(), "not accessible") {
			t.Errorf("expected 'not accessible' in error, got: %v", err)
		}
	})

	t.Run("data-info 500 returns error", func(t *testing.T) {
		diClient := newDataInfoServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})

		app := map[string]any{
			"groups": []any{
				map[string]any{
					"parameters": []any{
						map[string]any{"id": "p1", "type": "FileInput"},
					},
				},
			},
		}
		config := map[string]any{"p1": "/iplant/home/user/file.txt"}

		err := ValidateSubmission(diClient, &ValidationRequest{
			App: app, Config: config, User: "user",
		})
		if err == nil {
			t.Fatal("expected error for inaccessible paths (500 response)")
		}
	})

	t.Run("IsPublic uses public user", func(t *testing.T) {
		var capturedUser string
		diClient := newDataInfoServer(t, func(w http.ResponseWriter, r *http.Request) {
			capturedUser = r.URL.Query().Get("user")
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck // test server write
				"paths": map[string]any{"/a": map[string]any{}},
			})
		})

		app := map[string]any{
			"groups": []any{
				map[string]any{
					"parameters": []any{
						map[string]any{"id": "p1", "type": "FileInput"},
					},
				},
			},
		}
		config := map[string]any{"p1": "/a"}

		err := ValidateSubmission(diClient, &ValidationRequest{
			App: app, Config: config, User: "realuser", IsPublic: true,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if capturedUser != "public" {
			t.Errorf("expected user=public, got %q", capturedUser)
		}
	})

	t.Run("empty path value skips validation", func(t *testing.T) {
		app := map[string]any{
			"groups": []any{
				map[string]any{
					"parameters": []any{
						map[string]any{"id": "p1", "type": "FileInput"},
					},
				},
			},
		}
		config := map[string]any{"p1": ""}

		err := ValidateSubmission(nil, &ValidationRequest{
			App: app, Config: config, User: "user",
		})
		if err != nil {
			t.Fatalf("expected nil for empty path value, got %v", err)
		}
	})
}
