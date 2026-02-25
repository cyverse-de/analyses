package clients

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// httpClient is a shared HTTP client with a reasonable timeout for
// inter-service calls to apps and data-info.
var httpClient = &http.Client{Timeout: 30 * time.Second}

var domainSuffix = regexp.MustCompile(`@.*$`)

// stripDomain removes the @domain suffix from a username.
func stripDomain(username string) string {
	return domainSuffix.ReplaceAllString(username, "")
}

// buildURL constructs a full URL from a parsed base URL, path components,
// a username, and optional query parameters.
func buildURL(baseURL *url.URL, components []string, username string, query map[string]string) string {
	escaped := make([]string, len(components))
	for i, c := range components {
		escaped[i] = url.PathEscape(c)
	}
	u := baseURL.JoinPath(escaped...)
	q := u.Query()
	q.Set("user", stripDomain(username))
	for k, v := range query {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	return u.String()
}

// AppsClient interacts with the apps service.
type AppsClient struct {
	baseURL *url.URL
}

// NewAppsClient creates a new AppsClient. Returns an error if the base URL
// is invalid, catching misconfiguration early before any requests are sent.
func NewAppsClient(rawURL string) (*AppsClient, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid apps base URL %q: %w", rawURL, err)
	}
	return &AppsClient{baseURL: u}, nil
}

// GetApp retrieves an app definition from the apps service.
func (c *AppsClient) GetApp(user, systemID, appID string) (map[string]any, error) {
	reqURL := buildURL(c.baseURL, []string{"apps", systemID, appID}, user, nil)
	return doJSONGet(reqURL)
}

// GetAppVersion retrieves a specific app version from the apps service.
func (c *AppsClient) GetAppVersion(user, systemID, appID, versionID string) (map[string]any, error) {
	reqURL := buildURL(c.baseURL, []string{"apps", systemID, appID, "versions", versionID}, user, nil)
	return doJSONGet(reqURL)
}

// DataInfoClient interacts with the data-info service.
type DataInfoClient struct {
	baseURL *url.URL
}

// NewDataInfoClient creates a new DataInfoClient. Returns an error if the base
// URL is invalid, catching misconfiguration early.
func NewDataInfoClient(rawURL string) (*DataInfoClient, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid data-info base URL %q: %w", rawURL, err)
	}
	return &DataInfoClient{baseURL: u}, nil
}

// PathsAccessibleBy checks if paths are accessible by the given user.
// Uses the ignore-missing query parameter so data-info returns 200 even for
// missing paths. We verify that every requested path appears in the response.
func (c *DataInfoClient) PathsAccessibleBy(paths []string, user string) (bool, error) {
	if len(paths) == 0 {
		return true, nil
	}

	reqURL := buildURL(c.baseURL, []string{"path-info"}, user, map[string]string{"ignore-missing": "true"})
	body := map[string]any{"paths": paths}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return false, err
	}

	resp, err := httpClient.Post(reqURL, "application/json", strings.NewReader(string(bodyBytes)))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return false, fmt.Errorf("unexpected status %d from path-info and failed to read body: %w", resp.StatusCode, err)
		}
		return false, fmt.Errorf("unexpected status %d from path-info: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Paths map[string]any `json:"paths"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("failed to decode path-info response: %w", err)
	}

	for _, p := range paths {
		if _, ok := result.Paths[p]; !ok {
			return false, nil
		}
	}
	return true, nil
}

// PathChecker verifies path accessibility for a user.
type PathChecker interface {
	PathsAccessibleBy(paths []string, user string) (bool, error)
}

const publicUser = "public"

// AppParam represents a single parameter within an app parameter group.
type AppParam struct {
	ID           string `json:"id"`
	Type         string `json:"type"`
	Value        any    `json:"value,omitempty"`
	DefaultValue any    `json:"defaultValue,omitempty"`
}

// AppGroup represents a group of parameters in an app definition.
type AppGroup struct {
	Parameters []AppParam `json:"parameters"`
}

// ValidationRequest holds the typed inputs for ValidateSubmission.
type ValidationRequest struct {
	App      map[string]any
	Config   map[string]any
	IsPublic bool
	User     string
}

// extractAppGroups converts the untyped "groups" slice from an app definition
// into typed AppGroup values.
func extractAppGroups(app map[string]any) []AppGroup {
	rawGroups, _ := app["groups"].([]any)
	groups := make([]AppGroup, 0, len(rawGroups))
	for _, g := range rawGroups {
		gm, ok := g.(map[string]any)
		if !ok {
			continue
		}
		rawParams, _ := gm["parameters"].([]any)
		params := make([]AppParam, 0, len(rawParams))
		for _, p := range rawParams {
			pm, ok := p.(map[string]any)
			if !ok {
				continue
			}
			id, _ := pm["id"].(string)
			typ, _ := pm["type"].(string)
			params = append(params, AppParam{ID: id, Type: typ, Value: pm["value"], DefaultValue: pm["defaultValue"]})
		}
		groups = append(groups, AppGroup{Parameters: params})
	}
	return groups
}

// ValidateSubmission validates the submission config params against the app definition.
func ValidateSubmission(dataInfoClient PathChecker, req *ValidationRequest) error {
	if req.App == nil || req.Config == nil {
		return nil
	}

	groups := extractAppGroups(req.App)

	for _, group := range groups {
		for _, param := range group.Parameters {
			val, hasValue := req.Config[param.ID]
			if !hasValue {
				continue
			}

			if isInputType(param.Type) {
				u := req.User
				if req.IsPublic {
					u = publicUser
				}

				paths := extractPaths(val)
				if len(paths) == 0 {
					continue
				}

				accessible, err := dataInfoClient.PathsAccessibleBy(paths, u)
				if err != nil {
					return fmt.Errorf("error checking path accessibility: %w", err)
				}
				if !accessible {
					return fmt.Errorf("paths not accessible by user %s: %v", u, paths)
				}
			}
		}
	}

	return nil
}

func isInputType(t string) bool {
	return t == "FileInput" || t == "FolderInput" || t == "MultiFileSelector"
}

// extractPaths extracts path strings from a config value, which may be a single
// string (FileInput, FolderInput) or an array of strings (MultiFileSelector).
func extractPaths(val any) []string {
	switch v := val.(type) {
	case string:
		if v == "" {
			return nil
		}
		return []string{v}
	case []any:
		paths := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok && s != "" {
				paths = append(paths, s)
			}
		}
		return paths
	default:
		return nil
	}
}

// QuickLaunchAppInfo populates app params from a submission's config.
func QuickLaunchAppInfo(submission, app map[string]any, sysID string) map[string]any {
	config, _ := submission["config"].(map[string]any)
	debug, _ := submission["debug"].(bool)
	app["debug"] = debug

	rawGroups, ok := app["groups"].([]any)
	if !ok || config == nil {
		return app
	}

	groups := extractAppGroups(app)
	for i, group := range groups {
		gm := rawGroups[i].(map[string]any)
		rawParams := gm["parameters"].([]any)
		for j, param := range group.Parameters {
			val, hasVal := config[param.ID]
			if !hasVal {
				continue
			}
			pm := rawParams[j].(map[string]any)
			if isInputType(param.Type) {
				pathVal := map[string]any{"path": val}
				pm["value"] = pathVal
				pm["defaultValue"] = pathVal
			} else {
				pm["value"] = val
				pm["defaultValue"] = val
			}
			rawParams[j] = pm
		}
		gm["parameters"] = rawParams
		rawGroups[i] = gm
	}

	app["groups"] = rawGroups
	return app
}

func doJSONGet(reqURL string) (map[string]any, error) {
	resp, err := httpClient.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("unexpected status %d and failed to read body: %w", resp.StatusCode, err)
		}
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return result, nil
}
