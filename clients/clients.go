package clients

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

var domainSuffix = regexp.MustCompile(`@.*$`)

// StripDomain removes the @domain suffix from a username.
func StripDomain(username string) string {
	return domainSuffix.ReplaceAllString(username, "")
}

// AppsClient interacts with the apps service.
type AppsClient struct {
	BaseURL string
}

// NewAppsClient creates a new AppsClient.
func NewAppsClient(baseURL string) *AppsClient {
	return &AppsClient{BaseURL: strings.TrimRight(baseURL, "/")}
}

func (c *AppsClient) buildURL(components []string, username string, query map[string]string) string {
	u, _ := url.Parse(c.BaseURL)
	u.Path = strings.Join(append([]string{u.Path}, components...), "/")
	q := u.Query()
	q.Set("user", StripDomain(username))
	for k, v := range query {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	return u.String()
}

// GetApp retrieves an app definition from the apps service.
func (c *AppsClient) GetApp(user, systemID, appID string) (map[string]interface{}, error) {
	reqURL := c.buildURL([]string{"apps", systemID, appID}, user, nil)
	return doJSONGet(reqURL)
}

// GetAppVersion retrieves a specific app version from the apps service.
func (c *AppsClient) GetAppVersion(user, systemID, appID, versionID string) (map[string]interface{}, error) {
	reqURL := c.buildURL([]string{"apps", systemID, appID, "versions", versionID}, user, nil)
	return doJSONGet(reqURL)
}

// DataInfoClient interacts with the data-info service.
type DataInfoClient struct {
	BaseURL string
}

// NewDataInfoClient creates a new DataInfoClient.
func NewDataInfoClient(baseURL string) *DataInfoClient {
	return &DataInfoClient{BaseURL: strings.TrimRight(baseURL, "/")}
}

func (c *DataInfoClient) buildURL(components []string, username string, query map[string]string) string {
	u, _ := url.Parse(c.BaseURL)
	u.Path = strings.Join(append([]string{u.Path}, components...), "/")
	q := u.Query()
	q.Set("user", StripDomain(username))
	for k, v := range query {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	return u.String()
}

// PathsAccessibleBy checks if paths are accessible by the given user.
func (c *DataInfoClient) PathsAccessibleBy(paths []string, user string) (bool, error) {
	if len(paths) == 0 {
		return true, nil
	}

	reqURL := c.buildURL([]string{"path-info"}, user, nil)
	body := map[string]interface{}{"paths": paths}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return false, err
	}

	resp, err := http.Post(reqURL, "application/json", strings.NewReader(string(bodyBytes)))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode == http.StatusInternalServerError {
		return false, nil
	}
	return resp.StatusCode == http.StatusOK, nil
}

const publicUser = "public"

// ValidateSubmission validates the submission config params against the app definition.
func ValidateSubmission(appsClient *AppsClient, dataInfoClient *DataInfoClient, qlInfo map[string]interface{}) error {
	quicklaunch, _ := qlInfo["quicklaunch"].(map[string]interface{})
	app, _ := qlInfo["app"].(map[string]interface{})
	user, _ := qlInfo["user"].(string)

	if app == nil || quicklaunch == nil {
		return nil
	}

	submission, _ := quicklaunch["submission"].(map[string]interface{})
	if submission == nil {
		return nil
	}
	config, _ := submission["config"].(map[string]interface{})
	if config == nil {
		return nil
	}

	isPublic, _ := quicklaunch["is_public"].(bool)
	groups, _ := app["groups"].([]interface{})

	for _, g := range groups {
		group, ok := g.(map[string]interface{})
		if !ok {
			continue
		}
		params, _ := group["parameters"].([]interface{})
		for _, p := range params {
			param, ok := p.(map[string]interface{})
			if !ok {
				continue
			}
			paramID, _ := param["id"].(string)
			paramType, _ := param["type"].(string)

			if _, hasValue := config[paramID]; !hasValue {
				continue
			}

			if isInputType(paramType) {
				value, _ := config[paramID].(string)
				u := user
				if isPublic {
					u = publicUser
				}
				accessible, err := dataInfoClient.PathsAccessibleBy([]string{value}, u)
				if err != nil {
					return fmt.Errorf("error checking path accessibility: %w", err)
				}
				if !accessible {
					return fmt.Errorf("path not accessible by user %s: %s", u, value)
				}
			}
		}
	}

	return nil
}

func isInputType(t string) bool {
	return t == "FileInput" || t == "FolderInput" || t == "MultiFileSelector"
}

// QuickLaunchAppInfo populates app params from a submission's config.
func QuickLaunchAppInfo(submission, app map[string]interface{}, sysID string) map[string]interface{} {
	config, _ := submission["config"].(map[string]interface{})
	debug, _ := submission["debug"].(bool)
	app["debug"] = debug

	groups, ok := app["groups"].([]interface{})
	if !ok || config == nil {
		return app
	}

	for i, g := range groups {
		group, ok := g.(map[string]interface{})
		if !ok {
			continue
		}
		params, _ := group["parameters"].([]interface{})
		for j, p := range params {
			param, ok := p.(map[string]interface{})
			if !ok {
				continue
			}
			paramID, _ := param["id"].(string)
			paramType, _ := param["type"].(string)

			if val, hasVal := config[paramID]; hasVal {
				if isInputType(paramType) {
					pathVal := map[string]interface{}{"path": val}
					param["value"] = pathVal
					param["defaultValue"] = pathVal
				} else {
					param["value"] = val
					param["defaultValue"] = val
				}
				params[j] = param
			}
		}
		group["parameters"] = params
		groups[i] = group
	}

	app["groups"] = groups
	return app
}

func doJSONGet(reqURL string) (map[string]interface{}, error) {
	resp, err := http.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return result, nil
}
