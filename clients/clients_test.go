package clients

import (
	"net/url"
	"slices"
	"strings"
	"testing"
)

func TestStripDomain(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"with domain", "user@example.com", "user"},
		{"without domain", "user", "user"},
		{"empty string", "", ""},
		{"at-only", "@", ""},
		{"multiple @", "user@domain@extra", "user"},
		{"domain only", "@domain.com", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripDomain(tt.input)
			if got != tt.want {
				t.Errorf("stripDomain(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func mustParseURL(t *testing.T, rawURL string) *url.URL {
	t.Helper()
	u, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("failed to parse URL %q: %v", rawURL, err)
	}
	return u
}

func TestBuildURL(t *testing.T) {
	tests := []struct {
		name         string
		baseURL      string
		components   []string
		username     string
		query        map[string]string
		wantContains string // substring the result should contain
	}{
		{
			name:         "basic path",
			baseURL:      "http://localhost:8080",
			components:   []string{"apps", "de", "123"},
			username:     "testuser",
			wantContains: "/apps/de/123",
		},
		{
			name:         "with query params",
			baseURL:      "http://localhost:8080",
			components:   []string{"apps"},
			username:     "testuser",
			query:        map[string]string{"foo": "bar"},
			wantContains: "foo=bar",
		},
		{
			name:         "strips domain from user",
			baseURL:      "http://localhost:8080",
			components:   []string{"apps"},
			username:     "testuser@example.com",
			wantContains: "user=testuser",
		},
		{
			name:         "special chars in components",
			baseURL:      "http://localhost:8080",
			components:   []string{"apps", "hello world"},
			username:     "user",
			wantContains: "/apps/hello%20world",
		},
		{
			name:         "empty components",
			baseURL:      "http://localhost:8080",
			components:   []string{},
			username:     "user",
			wantContains: "user=user",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := mustParseURL(t, tt.baseURL)
			got := buildURL(base, tt.components, tt.username, tt.query)
			if tt.wantContains != "" {
				if !strings.Contains(got, tt.wantContains) {
					t.Errorf("buildURL() = %q, want it to contain %q", got, tt.wantContains)
				}
			}
		})
	}
}

func TestIsInputType(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"FileInput", "FileInput", true},
		{"FolderInput", "FolderInput", true},
		{"MultiFileSelector", "MultiFileSelector", true},
		{"TextInput", "TextInput", false},
		{"empty", "", false},
		{"wrong case", "fileinput", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isInputType(tt.input); got != tt.want {
				t.Errorf("isInputType(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestExtractPaths(t *testing.T) {
	tests := []struct {
		name string
		val  any
		want []string
	}{
		{"string value", "/path/to/file", []string{"/path/to/file"}},
		{"empty string", "", nil},
		{"nil", nil, nil},
		{"slice of strings", []any{"/a", "/b"}, []string{"/a", "/b"}},
		{"slice mixed types", []any{"/a", 42, "/b"}, []string{"/a", "/b"}},
		{"empty slice", []any{}, nil},
		{"non-string non-slice", 42, nil},
		{"slice with empty strings", []any{"/a", "", "/b"}, []string{"/a", "/b"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractPaths(tt.val)
			if !slices.Equal(got, tt.want) {
				t.Errorf("extractPaths(%v) = %v, want %v", tt.val, got, tt.want)
			}
		})
	}
}

func TestExtractAppGroups(t *testing.T) {
	tests := []struct {
		name string
		app  map[string]any
		want int // expected number of groups
	}{
		{"nil groups", map[string]any{}, 0},
		{"non-slice groups", map[string]any{"groups": "bad"}, 0},
		{
			"valid groups with params",
			map[string]any{
				"groups": []any{
					map[string]any{
						"parameters": []any{
							map[string]any{"id": "p1", "type": "FileInput"},
							map[string]any{"id": "p2", "type": "TextInput"},
						},
					},
				},
			},
			1,
		},
		{
			"non-map group is skipped",
			map[string]any{"groups": []any{"not a map"}},
			0,
		},
		{
			"non-map param is skipped",
			map[string]any{
				"groups": []any{
					map[string]any{
						"parameters": []any{"not a map"},
					},
				},
			},
			1, // group created but with 0 params
		},
		{
			"missing id and type fields",
			map[string]any{
				"groups": []any{
					map[string]any{
						"parameters": []any{
							map[string]any{"other": "field"},
						},
					},
				},
			},
			1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractAppGroups(tt.app)
			if len(got) != tt.want {
				t.Errorf("extractAppGroups() returned %d groups, want %d", len(got), tt.want)
			}
		})
	}

	// Additional check: verify param fields are populated
	t.Run("param fields populated", func(t *testing.T) {
		app := map[string]any{
			"groups": []any{
				map[string]any{
					"parameters": []any{
						map[string]any{"id": "p1", "type": "FileInput", "value": "/path", "defaultValue": "/default"},
					},
				},
			},
		}
		groups := extractAppGroups(app)
		if len(groups) != 1 || len(groups[0].Parameters) != 1 {
			t.Fatal("expected 1 group with 1 param")
		}
		p := groups[0].Parameters[0]
		if p.ID != "p1" || p.Type != "FileInput" {
			t.Errorf("got id=%q type=%q, want p1/FileInput", p.ID, p.Type)
		}
	})
}

func TestQuickLaunchAppInfo(t *testing.T) {
	t.Run("no config in submission", func(t *testing.T) {
		submission := map[string]any{}
		app := map[string]any{"groups": []any{}}
		result := QuickLaunchAppInfo(submission, app, "de")
		if _, ok := result["debug"]; !ok {
			t.Error("expected debug field to be set")
		}
	})

	t.Run("no groups in app", func(t *testing.T) {
		submission := map[string]any{"config": map[string]any{"p1": "val"}}
		app := map[string]any{}
		result := QuickLaunchAppInfo(submission, app, "de")
		if result["debug"] != false {
			t.Error("expected debug to be false")
		}
	})

	t.Run("FileInput param gets path wrapped", func(t *testing.T) {
		submission := map[string]any{
			"config": map[string]any{"p1": "/iplant/home/user/file.txt"},
			"debug":  true,
		}
		app := map[string]any{
			"groups": []any{
				map[string]any{
					"parameters": []any{
						map[string]any{"id": "p1", "type": "FileInput"},
					},
				},
			},
		}
		result := QuickLaunchAppInfo(submission, app, "de")
		if result["debug"] != true {
			t.Error("expected debug to be true")
		}
		groups := result["groups"].([]any)
		params := groups[0].(map[string]any)["parameters"].([]any)
		pm := params[0].(map[string]any)
		valMap, ok := pm["value"].(map[string]any)
		if !ok {
			t.Fatal("expected value to be a map")
		}
		if valMap["path"] != "/iplant/home/user/file.txt" {
			t.Errorf("expected path=/iplant/home/user/file.txt, got %v", valMap["path"])
		}
	})

	t.Run("non-input param gets raw value", func(t *testing.T) {
		submission := map[string]any{
			"config": map[string]any{"p1": "some-text-value"},
		}
		app := map[string]any{
			"groups": []any{
				map[string]any{
					"parameters": []any{
						map[string]any{"id": "p1", "type": "TextInput"},
					},
				},
			},
		}
		result := QuickLaunchAppInfo(submission, app, "de")
		groups := result["groups"].([]any)
		params := groups[0].(map[string]any)["parameters"].([]any)
		pm := params[0].(map[string]any)
		if pm["value"] != "some-text-value" {
			t.Errorf("expected raw value, got %v", pm["value"])
		}
	})

	t.Run("unmatched config key is ignored", func(t *testing.T) {
		submission := map[string]any{
			"config": map[string]any{"no-match": "val"},
		}
		app := map[string]any{
			"groups": []any{
				map[string]any{
					"parameters": []any{
						map[string]any{"id": "p1", "type": "TextInput"},
					},
				},
			},
		}
		result := QuickLaunchAppInfo(submission, app, "de")
		groups := result["groups"].([]any)
		params := groups[0].(map[string]any)["parameters"].([]any)
		pm := params[0].(map[string]any)
		if _, hasValue := pm["value"]; hasValue {
			t.Error("expected value to not be set for unmatched config key")
		}
	})

	t.Run("debug flag propagated", func(t *testing.T) {
		submission := map[string]any{"debug": true}
		app := map[string]any{}
		result := QuickLaunchAppInfo(submission, app, "de")
		if result["debug"] != true {
			t.Error("expected debug=true")
		}
	})
}
