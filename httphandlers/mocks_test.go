package httphandlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/cyverse-de/analyses/db"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

// stubResult is a minimal sql.Result used by mockTx.Exec.
type stubResult struct{}

func (stubResult) LastInsertId() (int64, error) { return 0, nil }
func (stubResult) RowsAffected() (int64, error) { return 1, nil }

// mockTx implements db.Tx for testing.
type mockTx struct {
	CommitFn   func() error
	RollbackFn func() error
}

func (m *mockTx) Exec(query string, args ...any) (sql.Result, error) {
	return stubResult{}, nil
}

func (m *mockTx) QueryRow(query string, args ...any) *sql.Row {
	return nil
}

func (m *mockTx) QueryRowx(query string, args ...any) *sqlx.Row {
	return nil
}

func (m *mockTx) Select(dest any, query string, args ...any) error {
	return nil
}

func (m *mockTx) Commit() error {
	if m.CommitFn != nil {
		return m.CommitFn()
	}
	return nil
}

func (m *mockTx) Rollback() error {
	if m.RollbackFn != nil {
		return m.RollbackFn()
	}
	return nil
}

// mockDB implements DatabaseStore with function fields.
type mockDB struct {
	BeginTxFn func() (db.Tx, error)

	GetQuickLaunchFn        func(tx db.Tx, id, user string) (*db.QuickLaunch, error)
	GetAllQuickLaunchesFn   func(tx db.Tx, user string) ([]db.QuickLaunch, error)
	GetQuickLaunchesByAppFn func(tx db.Tx, appID, user string) ([]db.QuickLaunch, error)
	AddQuickLaunchFn        func(tx db.Tx, user string, nql *db.NewQuickLaunch) (*db.QuickLaunch, error)
	UpdateQuickLaunchFn     func(tx db.Tx, id, user string, uql *db.UpdateQuickLaunchRequest) (*db.QuickLaunch, error)
	DeleteQuickLaunchFn     func(tx db.Tx, id, user string) error
	MergeSubmissionFn       func(tx db.Tx, qlID, user string, newSubmission json.RawMessage) (json.RawMessage, error)

	GetAllFavoritesFn func(tx db.Tx, user string) ([]db.QuickLaunchFavorite, error)
	GetFavoriteFn     func(tx db.Tx, user, favID string) (*db.QuickLaunchFavorite, error)
	AddFavoriteFn     func(tx db.Tx, user, quickLaunchID string) (*db.QuickLaunchFavorite, error)
	DeleteFavoriteFn  func(tx db.Tx, user, favID string) error

	GetUserDefaultFn     func(tx db.Tx, user, id string) (*db.QuickLaunchUserDefault, error)
	GetAllUserDefaultsFn func(tx db.Tx, user string) ([]db.QuickLaunchUserDefault, error)
	AddUserDefaultFn     func(tx db.Tx, user string, nud *db.NewQuickLaunchUserDefault) (*db.QuickLaunchUserDefault, error)
	UpdateUserDefaultFn  func(tx db.Tx, id, user string, update *db.UpdateQuickLaunchUserDefaultRequest) (*db.QuickLaunchUserDefault, error)
	DeleteUserDefaultFn  func(tx db.Tx, user, id string) error

	GetGlobalDefaultFn     func(tx db.Tx, user, id string) (*db.QuickLaunchGlobalDefault, error)
	GetAllGlobalDefaultsFn func(tx db.Tx, user string) ([]db.QuickLaunchGlobalDefault, error)
	AddGlobalDefaultFn     func(tx db.Tx, user string, ngd *db.NewQuickLaunchGlobalDefault) (*db.QuickLaunchGlobalDefault, error)
	UpdateGlobalDefaultFn  func(tx db.Tx, id, user string, update *db.UpdateQuickLaunchGlobalDefaultRequest) (*db.QuickLaunchGlobalDefault, error)
	DeleteGlobalDefaultFn  func(tx db.Tx, user, id string) error

	ListConcurrentJobLimitsFn  func(tx db.Tx) ([]db.ConcurrentJobLimit, error)
	GetConcurrentJobLimitFn    func(tx db.Tx, username string) (*db.ConcurrentJobLimit, error)
	SetConcurrentJobLimitFn    func(tx db.Tx, username string, limit int) (*db.ConcurrentJobLimit, error)
	RemoveConcurrentJobLimitFn func(tx db.Tx, username string) (*db.ConcurrentJobLimit, error)
}

func (m *mockDB) BeginTx() (db.Tx, error) {
	return m.BeginTxFn()
}

func (m *mockDB) GetQuickLaunch(tx db.Tx, id, user string) (*db.QuickLaunch, error) {
	return m.GetQuickLaunchFn(tx, id, user)
}
func (m *mockDB) GetAllQuickLaunches(tx db.Tx, user string) ([]db.QuickLaunch, error) {
	return m.GetAllQuickLaunchesFn(tx, user)
}
func (m *mockDB) GetQuickLaunchesByApp(tx db.Tx, appID, user string) ([]db.QuickLaunch, error) {
	return m.GetQuickLaunchesByAppFn(tx, appID, user)
}
func (m *mockDB) AddQuickLaunch(tx db.Tx, user string, nql *db.NewQuickLaunch) (*db.QuickLaunch, error) {
	return m.AddQuickLaunchFn(tx, user, nql)
}
func (m *mockDB) UpdateQuickLaunch(tx db.Tx, id, user string, uql *db.UpdateQuickLaunchRequest) (*db.QuickLaunch, error) {
	return m.UpdateQuickLaunchFn(tx, id, user, uql)
}
func (m *mockDB) DeleteQuickLaunch(tx db.Tx, id, user string) error {
	return m.DeleteQuickLaunchFn(tx, id, user)
}
func (m *mockDB) MergeSubmission(tx db.Tx, qlID, user string, newSubmission json.RawMessage) (json.RawMessage, error) {
	return m.MergeSubmissionFn(tx, qlID, user, newSubmission)
}
func (m *mockDB) GetAllFavorites(tx db.Tx, user string) ([]db.QuickLaunchFavorite, error) {
	return m.GetAllFavoritesFn(tx, user)
}
func (m *mockDB) GetFavorite(tx db.Tx, user, favID string) (*db.QuickLaunchFavorite, error) {
	return m.GetFavoriteFn(tx, user, favID)
}
func (m *mockDB) AddFavorite(tx db.Tx, user, quickLaunchID string) (*db.QuickLaunchFavorite, error) {
	return m.AddFavoriteFn(tx, user, quickLaunchID)
}
func (m *mockDB) DeleteFavorite(tx db.Tx, user, favID string) error {
	return m.DeleteFavoriteFn(tx, user, favID)
}
func (m *mockDB) GetUserDefault(tx db.Tx, user, id string) (*db.QuickLaunchUserDefault, error) {
	return m.GetUserDefaultFn(tx, user, id)
}
func (m *mockDB) GetAllUserDefaults(tx db.Tx, user string) ([]db.QuickLaunchUserDefault, error) {
	return m.GetAllUserDefaultsFn(tx, user)
}
func (m *mockDB) AddUserDefault(tx db.Tx, user string, nud *db.NewQuickLaunchUserDefault) (*db.QuickLaunchUserDefault, error) {
	return m.AddUserDefaultFn(tx, user, nud)
}
func (m *mockDB) UpdateUserDefault(tx db.Tx, id, user string, update *db.UpdateQuickLaunchUserDefaultRequest) (*db.QuickLaunchUserDefault, error) {
	return m.UpdateUserDefaultFn(tx, id, user, update)
}
func (m *mockDB) DeleteUserDefault(tx db.Tx, user, id string) error {
	return m.DeleteUserDefaultFn(tx, user, id)
}
func (m *mockDB) GetGlobalDefault(tx db.Tx, user, id string) (*db.QuickLaunchGlobalDefault, error) {
	return m.GetGlobalDefaultFn(tx, user, id)
}
func (m *mockDB) GetAllGlobalDefaults(tx db.Tx, user string) ([]db.QuickLaunchGlobalDefault, error) {
	return m.GetAllGlobalDefaultsFn(tx, user)
}
func (m *mockDB) AddGlobalDefault(tx db.Tx, user string, ngd *db.NewQuickLaunchGlobalDefault) (*db.QuickLaunchGlobalDefault, error) {
	return m.AddGlobalDefaultFn(tx, user, ngd)
}
func (m *mockDB) UpdateGlobalDefault(tx db.Tx, id, user string, update *db.UpdateQuickLaunchGlobalDefaultRequest) (*db.QuickLaunchGlobalDefault, error) {
	return m.UpdateGlobalDefaultFn(tx, id, user, update)
}
func (m *mockDB) DeleteGlobalDefault(tx db.Tx, user, id string) error {
	return m.DeleteGlobalDefaultFn(tx, user, id)
}
func (m *mockDB) ListConcurrentJobLimits(tx db.Tx) ([]db.ConcurrentJobLimit, error) {
	return m.ListConcurrentJobLimitsFn(tx)
}
func (m *mockDB) GetConcurrentJobLimit(tx db.Tx, username string) (*db.ConcurrentJobLimit, error) {
	return m.GetConcurrentJobLimitFn(tx, username)
}
func (m *mockDB) SetConcurrentJobLimit(tx db.Tx, username string, limit int) (*db.ConcurrentJobLimit, error) {
	return m.SetConcurrentJobLimitFn(tx, username, limit)
}
func (m *mockDB) RemoveConcurrentJobLimit(tx db.Tx, username string) (*db.ConcurrentJobLimit, error) {
	return m.RemoveConcurrentJobLimitFn(tx, username)
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
