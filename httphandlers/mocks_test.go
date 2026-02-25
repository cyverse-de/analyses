package httphandlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/cyverse-de/analyses/db"
	"github.com/labstack/echo/v4"
)

// mockDB implements DatabaseStore with function fields.
type mockDB struct {
	GetQuickLaunchFn        func(id, user string) (*db.QuickLaunch, error)
	GetAllQuickLaunchesFn   func(user string) ([]db.QuickLaunch, error)
	GetQuickLaunchesByAppFn func(appID, user string) ([]db.QuickLaunch, error)
	AddQuickLaunchFn        func(user string, nql *db.NewQuickLaunch) (*db.QuickLaunch, error)
	UpdateQuickLaunchFn     func(id, user string, uql *db.UpdateQuickLaunchRequest) (*db.QuickLaunch, error)
	DeleteQuickLaunchFn     func(id, user string) error
	MergeSubmissionFn       func(qlID, user string, newSubmission json.RawMessage) (json.RawMessage, error)

	GetAllFavoritesFn func(user string) ([]db.QuickLaunchFavorite, error)
	GetFavoriteFn     func(user, favID string) (*db.QuickLaunchFavorite, error)
	AddFavoriteFn     func(user, quickLaunchID string) (*db.QuickLaunchFavorite, error)
	DeleteFavoriteFn  func(user, favID string) error

	GetUserDefaultFn     func(user, id string) (*db.QuickLaunchUserDefault, error)
	GetAllUserDefaultsFn func(user string) ([]db.QuickLaunchUserDefault, error)
	AddUserDefaultFn     func(user string, nud *db.NewQuickLaunchUserDefault) (*db.QuickLaunchUserDefault, error)
	UpdateUserDefaultFn  func(id, user string, update *db.UpdateQuickLaunchUserDefaultRequest) (*db.QuickLaunchUserDefault, error)
	DeleteUserDefaultFn  func(user, id string) error

	GetGlobalDefaultFn     func(user, id string) (*db.QuickLaunchGlobalDefault, error)
	GetAllGlobalDefaultsFn func(user string) ([]db.QuickLaunchGlobalDefault, error)
	AddGlobalDefaultFn     func(user string, ngd *db.NewQuickLaunchGlobalDefault) (*db.QuickLaunchGlobalDefault, error)
	UpdateGlobalDefaultFn  func(id, user string, update *db.UpdateQuickLaunchGlobalDefaultRequest) (*db.QuickLaunchGlobalDefault, error)
	DeleteGlobalDefaultFn  func(user, id string) error

	ListConcurrentJobLimitsFn  func() ([]db.ConcurrentJobLimit, error)
	GetConcurrentJobLimitFn    func(username string) (*db.ConcurrentJobLimit, error)
	SetConcurrentJobLimitFn    func(username string, limit int) (*db.ConcurrentJobLimit, error)
	RemoveConcurrentJobLimitFn func(username string) (*db.ConcurrentJobLimit, error)
}

func (m *mockDB) GetQuickLaunch(id, user string) (*db.QuickLaunch, error) {
	return m.GetQuickLaunchFn(id, user)
}
func (m *mockDB) GetAllQuickLaunches(user string) ([]db.QuickLaunch, error) {
	return m.GetAllQuickLaunchesFn(user)
}
func (m *mockDB) GetQuickLaunchesByApp(appID, user string) ([]db.QuickLaunch, error) {
	return m.GetQuickLaunchesByAppFn(appID, user)
}
func (m *mockDB) AddQuickLaunch(user string, nql *db.NewQuickLaunch) (*db.QuickLaunch, error) {
	return m.AddQuickLaunchFn(user, nql)
}
func (m *mockDB) UpdateQuickLaunch(id, user string, uql *db.UpdateQuickLaunchRequest) (*db.QuickLaunch, error) {
	return m.UpdateQuickLaunchFn(id, user, uql)
}
func (m *mockDB) DeleteQuickLaunch(id, user string) error {
	return m.DeleteQuickLaunchFn(id, user)
}
func (m *mockDB) MergeSubmission(qlID, user string, newSubmission json.RawMessage) (json.RawMessage, error) {
	return m.MergeSubmissionFn(qlID, user, newSubmission)
}
func (m *mockDB) GetAllFavorites(user string) ([]db.QuickLaunchFavorite, error) {
	return m.GetAllFavoritesFn(user)
}
func (m *mockDB) GetFavorite(user, favID string) (*db.QuickLaunchFavorite, error) {
	return m.GetFavoriteFn(user, favID)
}
func (m *mockDB) AddFavorite(user, quickLaunchID string) (*db.QuickLaunchFavorite, error) {
	return m.AddFavoriteFn(user, quickLaunchID)
}
func (m *mockDB) DeleteFavorite(user, favID string) error {
	return m.DeleteFavoriteFn(user, favID)
}
func (m *mockDB) GetUserDefault(user, id string) (*db.QuickLaunchUserDefault, error) {
	return m.GetUserDefaultFn(user, id)
}
func (m *mockDB) GetAllUserDefaults(user string) ([]db.QuickLaunchUserDefault, error) {
	return m.GetAllUserDefaultsFn(user)
}
func (m *mockDB) AddUserDefault(user string, nud *db.NewQuickLaunchUserDefault) (*db.QuickLaunchUserDefault, error) {
	return m.AddUserDefaultFn(user, nud)
}
func (m *mockDB) UpdateUserDefault(id, user string, update *db.UpdateQuickLaunchUserDefaultRequest) (*db.QuickLaunchUserDefault, error) {
	return m.UpdateUserDefaultFn(id, user, update)
}
func (m *mockDB) DeleteUserDefault(user, id string) error {
	return m.DeleteUserDefaultFn(user, id)
}
func (m *mockDB) GetGlobalDefault(user, id string) (*db.QuickLaunchGlobalDefault, error) {
	return m.GetGlobalDefaultFn(user, id)
}
func (m *mockDB) GetAllGlobalDefaults(user string) ([]db.QuickLaunchGlobalDefault, error) {
	return m.GetAllGlobalDefaultsFn(user)
}
func (m *mockDB) AddGlobalDefault(user string, ngd *db.NewQuickLaunchGlobalDefault) (*db.QuickLaunchGlobalDefault, error) {
	return m.AddGlobalDefaultFn(user, ngd)
}
func (m *mockDB) UpdateGlobalDefault(id, user string, update *db.UpdateQuickLaunchGlobalDefaultRequest) (*db.QuickLaunchGlobalDefault, error) {
	return m.UpdateGlobalDefaultFn(id, user, update)
}
func (m *mockDB) DeleteGlobalDefault(user, id string) error {
	return m.DeleteGlobalDefaultFn(user, id)
}
func (m *mockDB) ListConcurrentJobLimits() ([]db.ConcurrentJobLimit, error) {
	return m.ListConcurrentJobLimitsFn()
}
func (m *mockDB) GetConcurrentJobLimit(username string) (*db.ConcurrentJobLimit, error) {
	return m.GetConcurrentJobLimitFn(username)
}
func (m *mockDB) SetConcurrentJobLimit(username string, limit int) (*db.ConcurrentJobLimit, error) {
	return m.SetConcurrentJobLimitFn(username, limit)
}
func (m *mockDB) RemoveConcurrentJobLimit(username string) (*db.ConcurrentJobLimit, error) {
	return m.RemoveConcurrentJobLimitFn(username)
}

// mockAppFetcher implements AppFetcher.
type mockAppFetcher struct {
	GetAppFn        func(user, systemID, appID string) (map[string]any, error)
	GetAppVersionFn func(user, systemID, appID, versionID string) (map[string]any, error)
}

func (m *mockAppFetcher) GetApp(user, systemID, appID string) (map[string]any, error) {
	return m.GetAppFn(user, systemID, appID)
}
func (m *mockAppFetcher) GetAppVersion(user, systemID, appID, versionID string) (map[string]any, error) {
	return m.GetAppVersionFn(user, systemID, appID, versionID)
}

// mockPathChecker implements PathChecker.
type mockPathChecker struct {
	PathsAccessibleByFn func(paths []string, user string) (bool, error)
}

func (m *mockPathChecker) PathsAccessibleBy(paths []string, user string) (bool, error) {
	return m.PathsAccessibleByFn(paths, user)
}

// newTestContext creates an Echo context for testing.
func newTestContext(method, path string, body string, params map[string]string, query map[string]string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	} else {
		req = httptest.NewRequest(method, path, nil)
	}

	// Set query params
	q := req.URL.Query()
	for k, v := range query {
		q.Set(k, v)
	}
	req.URL.RawQuery = q.Encode()

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Set path params
	names := make([]string, 0, len(params))
	values := make([]string, 0, len(params))
	for k, v := range params {
		names = append(names, k)
		values = append(values, v)
	}
	c.SetParamNames(names...)
	c.SetParamValues(values...)

	return c, rec
}
