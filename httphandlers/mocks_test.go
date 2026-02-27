package httphandlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/cyverse-de/analyses/db"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

// stubResult is a minimal sql.Result used by mockTx.ExecContext.
type stubResult struct{}

func (stubResult) LastInsertId() (int64, error) { return 0, nil }
func (stubResult) RowsAffected() (int64, error) { return 1, nil }

// mockTx implements db.Tx for testing.
type mockTx struct {
	CommitFn   func() error
	RollbackFn func() error
}

func (m *mockTx) ExecContext(_ context.Context, query string, args ...any) (sql.Result, error) {
	return stubResult{}, nil
}

func (m *mockTx) QueryRowContext(_ context.Context, query string, args ...any) *sql.Row {
	return nil
}

func (m *mockTx) QueryRowxContext(_ context.Context, query string, args ...any) *sqlx.Row {
	return nil
}

func (m *mockTx) SelectContext(_ context.Context, dest any, query string, args ...any) error {
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
	BeginTxFn func(ctx context.Context) (db.Tx, error)

	GetQuickLaunchFn        func(ctx context.Context, tx db.Tx, id, user string) (*db.QuickLaunch, error)
	GetAllQuickLaunchesFn   func(ctx context.Context, tx db.Tx, user string) ([]db.QuickLaunch, error)
	GetQuickLaunchesByAppFn func(ctx context.Context, tx db.Tx, appID, user string) ([]db.QuickLaunch, error)
	AddQuickLaunchFn        func(ctx context.Context, tx db.Tx, user string, nql *db.NewQuickLaunch) (*db.QuickLaunch, error)
	UpdateQuickLaunchFn     func(ctx context.Context, tx db.Tx, id, user string, uql *db.UpdateQuickLaunchRequest) (*db.QuickLaunch, error)
	DeleteQuickLaunchFn     func(ctx context.Context, tx db.Tx, id, user string) error
	MergeSubmissionFn       func(ctx context.Context, tx db.Tx, qlID, user string, newSubmission json.RawMessage) (json.RawMessage, error)

	GetAllFavoritesFn func(ctx context.Context, tx db.Tx, user string) ([]db.QuickLaunchFavorite, error)
	GetFavoriteFn     func(ctx context.Context, tx db.Tx, user, favID string) (*db.QuickLaunchFavorite, error)
	AddFavoriteFn     func(ctx context.Context, tx db.Tx, user, quickLaunchID string) (*db.QuickLaunchFavorite, error)
	DeleteFavoriteFn  func(ctx context.Context, tx db.Tx, user, favID string) error

	GetUserDefaultFn     func(ctx context.Context, tx db.Tx, user, id string) (*db.QuickLaunchUserDefault, error)
	GetAllUserDefaultsFn func(ctx context.Context, tx db.Tx, user string) ([]db.QuickLaunchUserDefault, error)
	AddUserDefaultFn     func(ctx context.Context, tx db.Tx, user string, nud *db.NewQuickLaunchUserDefault) (*db.QuickLaunchUserDefault, error)
	UpdateUserDefaultFn  func(ctx context.Context, tx db.Tx, id, user string, update *db.UpdateQuickLaunchUserDefaultRequest) (*db.QuickLaunchUserDefault, error)
	DeleteUserDefaultFn  func(ctx context.Context, tx db.Tx, user, id string) error

	GetGlobalDefaultFn     func(ctx context.Context, tx db.Tx, user, id string) (*db.QuickLaunchGlobalDefault, error)
	GetAllGlobalDefaultsFn func(ctx context.Context, tx db.Tx, user string) ([]db.QuickLaunchGlobalDefault, error)
	AddGlobalDefaultFn     func(ctx context.Context, tx db.Tx, user string, ngd *db.NewQuickLaunchGlobalDefault) (*db.QuickLaunchGlobalDefault, error)
	UpdateGlobalDefaultFn  func(ctx context.Context, tx db.Tx, id, user string, update *db.UpdateQuickLaunchGlobalDefaultRequest) (*db.QuickLaunchGlobalDefault, error)
	DeleteGlobalDefaultFn  func(ctx context.Context, tx db.Tx, user, id string) error

	ListConcurrentJobLimitsFn  func(ctx context.Context, tx db.Tx) ([]db.ConcurrentJobLimit, error)
	GetConcurrentJobLimitFn    func(ctx context.Context, tx db.Tx, username string) (*db.ConcurrentJobLimit, error)
	SetConcurrentJobLimitFn    func(ctx context.Context, tx db.Tx, username string, limit int) (*db.ConcurrentJobLimit, error)
	RemoveConcurrentJobLimitFn func(ctx context.Context, tx db.Tx, username string) (*db.ConcurrentJobLimit, error)
}

func (m *mockDB) BeginTx(ctx context.Context) (db.Tx, error) {
	return m.BeginTxFn(ctx)
}

func (m *mockDB) GetQuickLaunch(ctx context.Context, tx db.Tx, id, user string) (*db.QuickLaunch, error) {
	return m.GetQuickLaunchFn(ctx, tx, id, user)
}
func (m *mockDB) GetAllQuickLaunches(ctx context.Context, tx db.Tx, user string) ([]db.QuickLaunch, error) {
	return m.GetAllQuickLaunchesFn(ctx, tx, user)
}
func (m *mockDB) GetQuickLaunchesByApp(ctx context.Context, tx db.Tx, appID, user string) ([]db.QuickLaunch, error) {
	return m.GetQuickLaunchesByAppFn(ctx, tx, appID, user)
}
func (m *mockDB) AddQuickLaunch(ctx context.Context, tx db.Tx, user string, nql *db.NewQuickLaunch) (*db.QuickLaunch, error) {
	return m.AddQuickLaunchFn(ctx, tx, user, nql)
}
func (m *mockDB) UpdateQuickLaunch(ctx context.Context, tx db.Tx, id, user string, uql *db.UpdateQuickLaunchRequest) (*db.QuickLaunch, error) {
	return m.UpdateQuickLaunchFn(ctx, tx, id, user, uql)
}
func (m *mockDB) DeleteQuickLaunch(ctx context.Context, tx db.Tx, id, user string) error {
	return m.DeleteQuickLaunchFn(ctx, tx, id, user)
}
func (m *mockDB) MergeSubmission(ctx context.Context, tx db.Tx, qlID, user string, newSubmission json.RawMessage) (json.RawMessage, error) {
	return m.MergeSubmissionFn(ctx, tx, qlID, user, newSubmission)
}
func (m *mockDB) GetAllFavorites(ctx context.Context, tx db.Tx, user string) ([]db.QuickLaunchFavorite, error) {
	return m.GetAllFavoritesFn(ctx, tx, user)
}
func (m *mockDB) GetFavorite(ctx context.Context, tx db.Tx, user, favID string) (*db.QuickLaunchFavorite, error) {
	return m.GetFavoriteFn(ctx, tx, user, favID)
}
func (m *mockDB) AddFavorite(ctx context.Context, tx db.Tx, user, quickLaunchID string) (*db.QuickLaunchFavorite, error) {
	return m.AddFavoriteFn(ctx, tx, user, quickLaunchID)
}
func (m *mockDB) DeleteFavorite(ctx context.Context, tx db.Tx, user, favID string) error {
	return m.DeleteFavoriteFn(ctx, tx, user, favID)
}
func (m *mockDB) GetUserDefault(ctx context.Context, tx db.Tx, user, id string) (*db.QuickLaunchUserDefault, error) {
	return m.GetUserDefaultFn(ctx, tx, user, id)
}
func (m *mockDB) GetAllUserDefaults(ctx context.Context, tx db.Tx, user string) ([]db.QuickLaunchUserDefault, error) {
	return m.GetAllUserDefaultsFn(ctx, tx, user)
}
func (m *mockDB) AddUserDefault(ctx context.Context, tx db.Tx, user string, nud *db.NewQuickLaunchUserDefault) (*db.QuickLaunchUserDefault, error) {
	return m.AddUserDefaultFn(ctx, tx, user, nud)
}
func (m *mockDB) UpdateUserDefault(ctx context.Context, tx db.Tx, id, user string, update *db.UpdateQuickLaunchUserDefaultRequest) (*db.QuickLaunchUserDefault, error) {
	return m.UpdateUserDefaultFn(ctx, tx, id, user, update)
}
func (m *mockDB) DeleteUserDefault(ctx context.Context, tx db.Tx, user, id string) error {
	return m.DeleteUserDefaultFn(ctx, tx, user, id)
}
func (m *mockDB) GetGlobalDefault(ctx context.Context, tx db.Tx, user, id string) (*db.QuickLaunchGlobalDefault, error) {
	return m.GetGlobalDefaultFn(ctx, tx, user, id)
}
func (m *mockDB) GetAllGlobalDefaults(ctx context.Context, tx db.Tx, user string) ([]db.QuickLaunchGlobalDefault, error) {
	return m.GetAllGlobalDefaultsFn(ctx, tx, user)
}
func (m *mockDB) AddGlobalDefault(ctx context.Context, tx db.Tx, user string, ngd *db.NewQuickLaunchGlobalDefault) (*db.QuickLaunchGlobalDefault, error) {
	return m.AddGlobalDefaultFn(ctx, tx, user, ngd)
}
func (m *mockDB) UpdateGlobalDefault(ctx context.Context, tx db.Tx, id, user string, update *db.UpdateQuickLaunchGlobalDefaultRequest) (*db.QuickLaunchGlobalDefault, error) {
	return m.UpdateGlobalDefaultFn(ctx, tx, id, user, update)
}
func (m *mockDB) DeleteGlobalDefault(ctx context.Context, tx db.Tx, user, id string) error {
	return m.DeleteGlobalDefaultFn(ctx, tx, user, id)
}
func (m *mockDB) ListConcurrentJobLimits(ctx context.Context, tx db.Tx) ([]db.ConcurrentJobLimit, error) {
	return m.ListConcurrentJobLimitsFn(ctx, tx)
}
func (m *mockDB) GetConcurrentJobLimit(ctx context.Context, tx db.Tx, username string) (*db.ConcurrentJobLimit, error) {
	return m.GetConcurrentJobLimitFn(ctx, tx, username)
}
func (m *mockDB) SetConcurrentJobLimit(ctx context.Context, tx db.Tx, username string, limit int) (*db.ConcurrentJobLimit, error) {
	return m.SetConcurrentJobLimitFn(ctx, tx, username, limit)
}
func (m *mockDB) RemoveConcurrentJobLimit(ctx context.Context, tx db.Tx, username string) (*db.ConcurrentJobLimit, error) {
	return m.RemoveConcurrentJobLimitFn(ctx, tx, username)
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
