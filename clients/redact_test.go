package clients

import (
	"encoding/json"
	"testing"
)

func sensitiveApp() map[string]any {
	return map[string]any{
		"groups": []any{
			map[string]any{
				"parameters": []any{
					map[string]any{"id": "step_secret", "type": "EnvironmentVariable", "is_sensitive": true},
					map[string]any{"id": "step_plain", "type": "EnvironmentVariable"},
				},
			},
		},
	}
}

func decodeConfig(t *testing.T, sub json.RawMessage, key string) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(sub, &m); err != nil {
		t.Fatalf("redacted submission isn't valid JSON: %v", err)
	}
	config, _ := m[key].(map[string]any)
	return config
}

func TestRedactSensitiveParams(t *testing.T) {
	t.Run("sensitive values are removed from config and job_config", func(t *testing.T) {
		sub := json.RawMessage(`{"name":"ql","config":{"step_secret":"hunter2","step_plain":"visible"},` +
			`"job_config":{"step_secret":"hunter2"}}`)
		got, err := RedactSensitiveParams(sub, sensitiveApp())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		config := decodeConfig(t, got, "config")
		if _, has := config["step_secret"]; has {
			t.Error("expected step_secret to be removed from config")
		}
		if config["step_plain"] != "visible" {
			t.Errorf("expected step_plain to be kept, got %v", config["step_plain"])
		}
		if _, has := decodeConfig(t, got, "job_config")["step_secret"]; has {
			t.Error("expected step_secret to be removed from job_config")
		}
	})

	t.Run("submission is unchanged when the app has no sensitive params", func(t *testing.T) {
		sub := json.RawMessage(`{"config":{"step_secret":"hunter2"}}`)
		app := map[string]any{"groups": []any{map[string]any{"parameters": []any{
			map[string]any{"id": "step_secret", "type": "EnvironmentVariable"},
		}}}}
		got, err := RedactSensitiveParams(sub, app)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(got) != string(sub) {
			t.Errorf("expected the submission unchanged, got %s", got)
		}
	})

	t.Run("submission without a config is unchanged", func(t *testing.T) {
		sub := json.RawMessage(`{"name":"ql"}`)
		got, err := RedactSensitiveParams(sub, sensitiveApp())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(got) != string(sub) {
			t.Errorf("expected the submission unchanged, got %s", got)
		}
	})

	t.Run("invalid JSON is an error", func(t *testing.T) {
		if _, err := RedactSensitiveParams(json.RawMessage(`{bad`), sensitiveApp()); err == nil {
			t.Error("expected an error for invalid JSON")
		}
	})
}

func TestQuickLaunchAppInfoSkipsSensitiveParams(t *testing.T) {
	submission := map[string]any{"config": map[string]any{"step_secret": "hunter2", "step_plain": "visible"}}
	result := QuickLaunchAppInfo(submission, sensitiveApp(), "de")
	params := result["groups"].([]any)[0].(map[string]any)["parameters"].([]any)
	if _, has := params[0].(map[string]any)["value"]; has {
		t.Error("expected the sensitive param not to be populated from a stored value")
	}
	if params[1].(map[string]any)["value"] != "visible" {
		t.Errorf("expected the plain param to be populated, got %v", params[1].(map[string]any)["value"])
	}
}
