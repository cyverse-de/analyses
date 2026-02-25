package clients

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDoJSONGet(t *testing.T) {
	t.Run("200 with valid JSON", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"id": "123"})
		}))
		defer ts.Close()

		result, err := doJSONGet(ts.URL)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result["id"] != "123" {
			t.Errorf("expected id=123, got %v", result["id"])
		}
	})

	t.Run("non-200 status", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("server error"))
		}))
		defer ts.Close()

		_, err := doJSONGet(ts.URL)
		if err == nil {
			t.Fatal("expected error for non-200 status")
		}
		if !strings.Contains(err.Error(), "500") {
			t.Errorf("expected error to mention status 500, got: %v", err)
		}
	})

	t.Run("invalid JSON on 200", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("not json"))
		}))
		defer ts.Close()

		_, err := doJSONGet(ts.URL)
		if err == nil {
			t.Fatal("expected error for invalid JSON")
		}
		if !strings.Contains(err.Error(), "decode") {
			t.Errorf("expected decode error, got: %v", err)
		}
	})
}

func TestAppsClient_GetApp(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify URL structure
			if !strings.Contains(r.URL.Path, "/apps/de/app-123") {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			if r.URL.Query().Get("user") != "testuser" {
				t.Errorf("unexpected user param: %s", r.URL.Query().Get("user"))
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"id": "app-123", "name": "My App"})
		}))
		defer ts.Close()

		client := NewAppsClient(ts.URL)
		result, err := client.GetApp("testuser@example.com", "de", "app-123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result["id"] != "app-123" {
			t.Errorf("expected id=app-123, got %v", result["id"])
		}
	})

	t.Run("error response", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("not found"))
		}))
		defer ts.Close()

		client := NewAppsClient(ts.URL)
		_, err := client.GetApp("user", "de", "bad-id")
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestAppsClient_GetAppVersion(t *testing.T) {
	t.Run("success with version path", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.URL.Path, "/apps/de/app-123/versions/v1") {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"id": "app-123", "version_id": "v1"})
		}))
		defer ts.Close()

		client := NewAppsClient(ts.URL)
		result, err := client.GetAppVersion("user", "de", "app-123", "v1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result["version_id"] != "v1" {
			t.Errorf("expected version_id=v1, got %v", result["version_id"])
		}
	})
}

func TestDataInfoClient_PathsAccessibleBy(t *testing.T) {
	t.Run("empty paths returns true", func(t *testing.T) {
		client := NewDataInfoClient("http://unused")
		ok, err := client.PathsAccessibleBy([]string{}, "user")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !ok {
			t.Error("expected true for empty paths")
		}
	})

	t.Run("all accessible", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"paths": map[string]any{
					"/a": map[string]any{},
					"/b": map[string]any{},
				},
			})
		}))
		defer ts.Close()

		client := NewDataInfoClient(ts.URL)
		ok, err := client.PathsAccessibleBy([]string{"/a", "/b"}, "user")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !ok {
			t.Error("expected true when all paths are accessible")
		}
	})

	t.Run("some paths missing from response", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"paths": map[string]any{
					"/a": map[string]any{},
				},
			})
		}))
		defer ts.Close()

		client := NewDataInfoClient(ts.URL)
		ok, err := client.PathsAccessibleBy([]string{"/a", "/b"}, "user")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ok {
			t.Error("expected false when some paths are missing")
		}
	})

	t.Run("500 response returns false without error", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer ts.Close()

		client := NewDataInfoClient(ts.URL)
		ok, err := client.PathsAccessibleBy([]string{"/a"}, "user")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ok {
			t.Error("expected false for 500 response")
		}
	})

	t.Run("non-200 non-500 returns error", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("forbidden"))
		}))
		defer ts.Close()

		client := NewDataInfoClient(ts.URL)
		_, err := client.PathsAccessibleBy([]string{"/a"}, "user")
		if err == nil {
			t.Fatal("expected error for 403 response")
		}
		if !strings.Contains(err.Error(), "403") {
			t.Errorf("expected error to mention 403, got: %v", err)
		}
	})
}
