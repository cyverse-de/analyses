package httphandlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/cyverse-de/analyses/db"
	"github.com/labstack/echo/v4"
)

// assertHTTPError checks that err is an *echo.HTTPError with the given status.
func assertHTTPError(t *testing.T, err error, wantCode int) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error with status %d, got nil", wantCode)
	}
	he, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected *echo.HTTPError, got %T: %v", err, err)
	}
	if he.Code != wantCode {
		t.Errorf("expected status %d, got %d (message: %v)", wantCode, he.Code, he.Message)
	}
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// defaultBeginTx is a helper that returns a BeginTxFn producing a plain mockTx.
func defaultBeginTx() func(context.Context) (db.Tx, error) {
	return func(context.Context) (db.Tx, error) { return &mockTx{}, nil }
}

// ──────────────────────────────────────────────────────────────
// Quick Launch Handlers
// ──────────────────────────────────────────────────────────────

func TestGetAllQuickLaunchesHandler(t *testing.T) {
	t.Run("missing user", func(t *testing.T) {
		h := &Handlers{}
		c, _ := newTestContext(http.MethodGet, "/quicklaunches", "", nil, nil)
		err := h.GetAllQuickLaunchesHandler(c)
		assertHTTPError(t, err, http.StatusBadRequest)
	})

	t.Run("begin tx error", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: func(context.Context) (db.Tx, error) { return nil, fmt.Errorf("tx error") },
		}}
		c, _ := newTestContext(http.MethodGet, "/quicklaunches", "", nil, map[string]string{"user": "testuser"})
		err := h.GetAllQuickLaunchesHandler(c)
		assertHTTPError(t, err, http.StatusInternalServerError)
	})

	t.Run("db error", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: defaultBeginTx(),
			GetAllQuickLaunchesFn: func(_ context.Context, _ db.Tx, user string) ([]db.QuickLaunch, error) {
				return nil, fmt.Errorf("db error")
			},
		}}
		c, _ := newTestContext(http.MethodGet, "/quicklaunches", "", nil, map[string]string{"user": "testuser"})
		err := h.GetAllQuickLaunchesHandler(c)
		assertHTTPError(t, err, http.StatusInternalServerError)
	})

	t.Run("success", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: defaultBeginTx(),
			GetAllQuickLaunchesFn: func(_ context.Context, _ db.Tx, user string) ([]db.QuickLaunch, error) {
				return []db.QuickLaunch{{ID: "ql-1"}}, nil
			},
		}}
		c, rec := newTestContext(http.MethodGet, "/quicklaunches", "", nil, map[string]string{"user": "testuser"})
		err := h.GetAllQuickLaunchesHandler(c)
		assertNoError(t, err)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})
}

func TestGetQuickLaunchesByAppHandler(t *testing.T) {
	t.Run("missing user", func(t *testing.T) {
		h := &Handlers{}
		c, _ := newTestContext(http.MethodGet, "/quicklaunches/apps/app-1", "", map[string]string{"id": "app-1"}, nil)
		err := h.GetQuickLaunchesByAppHandler(c)
		assertHTTPError(t, err, http.StatusBadRequest)
	})

	t.Run("begin tx error", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: func(context.Context) (db.Tx, error) { return nil, fmt.Errorf("tx error") },
		}}
		c, _ := newTestContext(http.MethodGet, "/quicklaunches/apps/app-1", "", map[string]string{"id": "app-1"}, map[string]string{"user": "u"})
		err := h.GetQuickLaunchesByAppHandler(c)
		assertHTTPError(t, err, http.StatusInternalServerError)
	})

	t.Run("db error", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: defaultBeginTx(),
			GetQuickLaunchesByAppFn: func(_ context.Context, _ db.Tx, appID, user string) ([]db.QuickLaunch, error) {
				return nil, fmt.Errorf("db error")
			},
		}}
		c, _ := newTestContext(http.MethodGet, "/quicklaunches/apps/app-1", "", map[string]string{"id": "app-1"}, map[string]string{"user": "u"})
		err := h.GetQuickLaunchesByAppHandler(c)
		assertHTTPError(t, err, http.StatusInternalServerError)
	})

	t.Run("success", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: defaultBeginTx(),
			GetQuickLaunchesByAppFn: func(_ context.Context, _ db.Tx, appID, user string) ([]db.QuickLaunch, error) {
				return []db.QuickLaunch{}, nil
			},
		}}
		c, rec := newTestContext(http.MethodGet, "/quicklaunches/apps/app-1", "", map[string]string{"id": "app-1"}, map[string]string{"user": "u"})
		err := h.GetQuickLaunchesByAppHandler(c)
		assertNoError(t, err)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})
}

func TestGetQuickLaunchHandler(t *testing.T) {
	t.Run("missing user", func(t *testing.T) {
		h := &Handlers{}
		c, _ := newTestContext(http.MethodGet, "/quicklaunches/ql-1", "", map[string]string{"id": "ql-1"}, nil)
		err := h.GetQuickLaunchHandler(c)
		assertHTTPError(t, err, http.StatusBadRequest)
	})

	t.Run("begin tx error", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: func(context.Context) (db.Tx, error) { return nil, fmt.Errorf("tx error") },
		}}
		c, _ := newTestContext(http.MethodGet, "/quicklaunches/ql-1", "", map[string]string{"id": "ql-1"}, map[string]string{"user": "u"})
		err := h.GetQuickLaunchHandler(c)
		assertHTTPError(t, err, http.StatusInternalServerError)
	})

	t.Run("not found", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: defaultBeginTx(),
			GetQuickLaunchFn: func(_ context.Context, _ db.Tx, id, user string) (*db.QuickLaunch, error) {
				return nil, db.NewNotFoundError("quick launch", id)
			},
		}}
		c, _ := newTestContext(http.MethodGet, "/quicklaunches/ql-1", "", map[string]string{"id": "ql-1"}, map[string]string{"user": "u"})
		err := h.GetQuickLaunchHandler(c)
		assertHTTPError(t, err, http.StatusNotFound)
	})

	t.Run("db error", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: defaultBeginTx(),
			GetQuickLaunchFn: func(_ context.Context, _ db.Tx, id, user string) (*db.QuickLaunch, error) {
				return nil, fmt.Errorf("connection refused")
			},
		}}
		c, _ := newTestContext(http.MethodGet, "/quicklaunches/ql-1", "", map[string]string{"id": "ql-1"}, map[string]string{"user": "u"})
		err := h.GetQuickLaunchHandler(c)
		assertHTTPError(t, err, http.StatusInternalServerError)
	})

	t.Run("success", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: defaultBeginTx(),
			GetQuickLaunchFn: func(_ context.Context, _ db.Tx, id, user string) (*db.QuickLaunch, error) {
				return &db.QuickLaunch{ID: id}, nil
			},
		}}
		c, rec := newTestContext(http.MethodGet, "/quicklaunches/ql-1", "", map[string]string{"id": "ql-1"}, map[string]string{"user": "u"})
		err := h.GetQuickLaunchHandler(c)
		assertNoError(t, err)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})
}

func TestAddQuickLaunchHandler(t *testing.T) {
	t.Run("missing user", func(t *testing.T) {
		h := &Handlers{}
		c, _ := newTestContext(http.MethodPost, "/quicklaunches", `{}`, nil, nil)
		err := h.AddQuickLaunchHandler(c)
		assertHTTPError(t, err, http.StatusBadRequest)
	})

	t.Run("invalid body", func(t *testing.T) {
		h := &Handlers{}
		c, _ := newTestContext(http.MethodPost, "/quicklaunches", `{invalid`, nil, map[string]string{"user": "u"})
		err := h.AddQuickLaunchHandler(c)
		assertHTTPError(t, err, http.StatusBadRequest)
	})

	t.Run("app fetch failure", func(t *testing.T) {
		h := &Handlers{
			AppsClient: &mockAppFetcher{
				GetAppFn: func(user, systemID, appID string) (map[string]any, error) {
					return nil, fmt.Errorf("app not found")
				},
			},
		}
		body := `{"app_id":"app-1","submission":{"config":{}}}`
		c, _ := newTestContext(http.MethodPost, "/quicklaunches", body, nil, map[string]string{"user": "u"})
		err := h.AddQuickLaunchHandler(c)
		assertHTTPError(t, err, http.StatusBadRequest)
	})

	t.Run("validation failure", func(t *testing.T) {
		h := &Handlers{
			AppsClient: &mockAppFetcher{
				GetAppFn: func(user, systemID, appID string) (map[string]any, error) {
					return map[string]any{
						"version_id": "v1",
						"groups": []any{
							map[string]any{
								"parameters": []any{
									map[string]any{"id": "p1", "type": "FileInput"},
								},
							},
						},
					}, nil
				},
			},
			DataInfoClient: &mockPathChecker{
				PathsAccessibleByFn: func(paths []string, user string) (bool, error) {
					return false, nil
				},
			},
		}
		// submission needs to be valid JSON that marshals to a map
		sub := map[string]any{"config": map[string]any{"p1": "/path"}}
		subBytes, _ := json.Marshal(sub)
		nql := map[string]any{"app_id": "app-1", "submission": json.RawMessage(subBytes)}
		bodyBytes, _ := json.Marshal(nql)
		c, _ := newTestContext(http.MethodPost, "/quicklaunches", string(bodyBytes), nil, map[string]string{"user": "u"})
		err := h.AddQuickLaunchHandler(c)
		assertHTTPError(t, err, http.StatusBadRequest)
	})

	t.Run("begin tx error", func(t *testing.T) {
		h := &Handlers{
			AppsClient: &mockAppFetcher{
				GetAppFn: func(user, systemID, appID string) (map[string]any, error) {
					return map[string]any{"version_id": "v1"}, nil
				},
			},
			DataInfoClient: &mockPathChecker{
				PathsAccessibleByFn: func(paths []string, user string) (bool, error) {
					return true, nil
				},
			},
			DB: &mockDB{
				BeginTxFn: func(context.Context) (db.Tx, error) { return nil, fmt.Errorf("tx error") },
			},
		}
		sub := map[string]any{"config": map[string]any{}}
		subBytes, _ := json.Marshal(sub)
		nql := map[string]any{"app_id": "app-1", "submission": json.RawMessage(subBytes)}
		bodyBytes, _ := json.Marshal(nql)
		c, _ := newTestContext(http.MethodPost, "/quicklaunches", string(bodyBytes), nil, map[string]string{"user": "u"})
		err := h.AddQuickLaunchHandler(c)
		assertHTTPError(t, err, http.StatusInternalServerError)
	})

	t.Run("db error on add", func(t *testing.T) {
		h := &Handlers{
			AppsClient: &mockAppFetcher{
				GetAppFn: func(user, systemID, appID string) (map[string]any, error) {
					return map[string]any{"version_id": "v1"}, nil
				},
			},
			DataInfoClient: &mockPathChecker{
				PathsAccessibleByFn: func(paths []string, user string) (bool, error) {
					return true, nil
				},
			},
			DB: &mockDB{
				BeginTxFn: defaultBeginTx(),
				AddQuickLaunchFn: func(_ context.Context, _ db.Tx, user string, nql *db.NewQuickLaunch) (*db.QuickLaunch, error) {
					return nil, fmt.Errorf("db error")
				},
			},
		}
		sub := map[string]any{"config": map[string]any{}}
		subBytes, _ := json.Marshal(sub)
		nql := map[string]any{"app_id": "app-1", "submission": json.RawMessage(subBytes)}
		bodyBytes, _ := json.Marshal(nql)
		c, _ := newTestContext(http.MethodPost, "/quicklaunches", string(bodyBytes), nil, map[string]string{"user": "u"})
		err := h.AddQuickLaunchHandler(c)
		assertHTTPError(t, err, http.StatusInternalServerError)
	})

	t.Run("success", func(t *testing.T) {
		h := &Handlers{
			AppsClient: &mockAppFetcher{
				GetAppFn: func(user, systemID, appID string) (map[string]any, error) {
					return map[string]any{"version_id": "v1"}, nil
				},
			},
			DataInfoClient: &mockPathChecker{
				PathsAccessibleByFn: func(paths []string, user string) (bool, error) {
					return true, nil
				},
			},
			DB: &mockDB{
				BeginTxFn: defaultBeginTx(),
				AddQuickLaunchFn: func(_ context.Context, _ db.Tx, user string, nql *db.NewQuickLaunch) (*db.QuickLaunch, error) {
					return &db.QuickLaunch{ID: "new-ql"}, nil
				},
			},
		}
		sub := map[string]any{"config": map[string]any{}}
		subBytes, _ := json.Marshal(sub)
		nql := map[string]any{"app_id": "app-1", "submission": json.RawMessage(subBytes)}
		bodyBytes, _ := json.Marshal(nql)
		c, rec := newTestContext(http.MethodPost, "/quicklaunches", string(bodyBytes), nil, map[string]string{"user": "u"})
		err := h.AddQuickLaunchHandler(c)
		assertNoError(t, err)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("with version_id uses GetAppVersion", func(t *testing.T) {
		var called bool
		h := &Handlers{
			AppsClient: &mockAppFetcher{
				GetAppFn: func(user, systemID, appID string) (map[string]any, error) {
					t.Error("GetApp should not be called when version_id is set")
					return nil, nil
				},
				GetAppVersionFn: func(user, systemID, appID, versionID string) (map[string]any, error) {
					called = true
					return map[string]any{}, nil
				},
			},
			DataInfoClient: &mockPathChecker{
				PathsAccessibleByFn: func(paths []string, user string) (bool, error) {
					return true, nil
				},
			},
			DB: &mockDB{
				BeginTxFn: defaultBeginTx(),
				AddQuickLaunchFn: func(_ context.Context, _ db.Tx, user string, nql *db.NewQuickLaunch) (*db.QuickLaunch, error) {
					return &db.QuickLaunch{ID: "new-ql"}, nil
				},
			},
		}
		sub := map[string]any{"config": map[string]any{}}
		subBytes, _ := json.Marshal(sub)
		nql := map[string]any{"app_id": "app-1", "app_version_id": "v1", "submission": json.RawMessage(subBytes)}
		bodyBytes, _ := json.Marshal(nql)
		c, _ := newTestContext(http.MethodPost, "/quicklaunches", string(bodyBytes), nil, map[string]string{"user": "u"})
		err := h.AddQuickLaunchHandler(c)
		assertNoError(t, err)
		if !called {
			t.Error("expected GetAppVersion to be called")
		}
	})
}

func TestUpdateQuickLaunchHandler(t *testing.T) {
	t.Run("missing user", func(t *testing.T) {
		h := &Handlers{}
		c, _ := newTestContext(http.MethodPatch, "/quicklaunches/ql-1", `{}`, map[string]string{"id": "ql-1"}, nil)
		err := h.UpdateQuickLaunchHandler(c)
		assertHTTPError(t, err, http.StatusBadRequest)
	})

	t.Run("invalid body", func(t *testing.T) {
		h := &Handlers{}
		c, _ := newTestContext(http.MethodPatch, "/quicklaunches/ql-1", `{bad`, map[string]string{"id": "ql-1"}, map[string]string{"user": "u"})
		err := h.UpdateQuickLaunchHandler(c)
		assertHTTPError(t, err, http.StatusBadRequest)
	})

	t.Run("begin tx error", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: func(context.Context) (db.Tx, error) { return nil, fmt.Errorf("tx error") },
		}}
		c, _ := newTestContext(http.MethodPatch, "/quicklaunches/ql-1", `{}`, map[string]string{"id": "ql-1"}, map[string]string{"user": "u"})
		err := h.UpdateQuickLaunchHandler(c)
		assertHTTPError(t, err, http.StatusInternalServerError)
	})

	t.Run("ql not found", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: defaultBeginTx(),
			GetQuickLaunchFn: func(_ context.Context, _ db.Tx, id, user string) (*db.QuickLaunch, error) {
				return nil, db.NewNotFoundError("quick launch", id)
			},
		}}
		c, _ := newTestContext(http.MethodPatch, "/quicklaunches/ql-1", `{}`, map[string]string{"id": "ql-1"}, map[string]string{"user": "u"})
		err := h.UpdateQuickLaunchHandler(c)
		assertHTTPError(t, err, http.StatusNotFound)
	})

	t.Run("app fetch failure", func(t *testing.T) {
		h := &Handlers{
			DB: &mockDB{
				BeginTxFn: defaultBeginTx(),
				GetQuickLaunchFn: func(_ context.Context, _ db.Tx, id, user string) (*db.QuickLaunch, error) {
					return &db.QuickLaunch{
						ID: id, AppID: "app-1", AppVersionID: "v1",
						Submission: json.RawMessage(`{"config":{}}`),
					}, nil
				},
			},
			AppsClient: &mockAppFetcher{
				GetAppVersionFn: func(user, systemID, appID, versionID string) (map[string]any, error) {
					return nil, fmt.Errorf("app not found")
				},
			},
		}
		c, _ := newTestContext(http.MethodPatch, "/quicklaunches/ql-1", `{}`, map[string]string{"id": "ql-1"}, map[string]string{"user": "u"})
		err := h.UpdateQuickLaunchHandler(c)
		assertHTTPError(t, err, http.StatusBadRequest)
	})

	t.Run("merge submission failure", func(t *testing.T) {
		h := &Handlers{
			DB: &mockDB{
				BeginTxFn: defaultBeginTx(),
				GetQuickLaunchFn: func(_ context.Context, _ db.Tx, id, user string) (*db.QuickLaunch, error) {
					return &db.QuickLaunch{
						ID: id, AppID: "app-1", AppVersionID: "v1",
						Submission: json.RawMessage(`{"config":{}}`),
					}, nil
				},
				MergeSubmissionFn: func(_ context.Context, _ db.Tx, qlID, user string, newSubmission json.RawMessage) (json.RawMessage, error) {
					return nil, fmt.Errorf("merge error")
				},
			},
			AppsClient: &mockAppFetcher{
				GetAppVersionFn: func(user, systemID, appID, versionID string) (map[string]any, error) {
					return map[string]any{}, nil
				},
			},
			DataInfoClient: &mockPathChecker{
				PathsAccessibleByFn: func(paths []string, user string) (bool, error) {
					return true, nil
				},
			},
		}
		sub := json.RawMessage(`{"config":{}}`)
		bodyObj := map[string]any{"submission": sub}
		bodyBytes, _ := json.Marshal(bodyObj)
		c, _ := newTestContext(http.MethodPatch, "/quicklaunches/ql-1", string(bodyBytes), map[string]string{"id": "ql-1"}, map[string]string{"user": "u"})
		err := h.UpdateQuickLaunchHandler(c)
		assertHTTPError(t, err, http.StatusInternalServerError)
	})

	t.Run("success without submission change", func(t *testing.T) {
		h := &Handlers{
			DB: &mockDB{
				BeginTxFn: defaultBeginTx(),
				GetQuickLaunchFn: func(_ context.Context, _ db.Tx, id, user string) (*db.QuickLaunch, error) {
					return &db.QuickLaunch{
						ID: id, AppID: "app-1", AppVersionID: "v1",
						Submission: json.RawMessage(`{"config":{}}`),
					}, nil
				},
				UpdateQuickLaunchFn: func(_ context.Context, _ db.Tx, id, user string, uql *db.UpdateQuickLaunchRequest) (*db.QuickLaunch, error) {
					return &db.QuickLaunch{ID: id}, nil
				},
			},
			AppsClient: &mockAppFetcher{
				GetAppVersionFn: func(user, systemID, appID, versionID string) (map[string]any, error) {
					return map[string]any{}, nil
				},
			},
			DataInfoClient: &mockPathChecker{
				PathsAccessibleByFn: func(paths []string, user string) (bool, error) {
					return true, nil
				},
			},
		}
		name := "new-name"
		bodyObj := map[string]any{"name": name}
		bodyBytes, _ := json.Marshal(bodyObj)
		c, rec := newTestContext(http.MethodPatch, "/quicklaunches/ql-1", string(bodyBytes), map[string]string{"id": "ql-1"}, map[string]string{"user": "u"})
		err := h.UpdateQuickLaunchHandler(c)
		assertNoError(t, err)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})
}

func TestDeleteQuickLaunchHandler(t *testing.T) {
	t.Run("missing user", func(t *testing.T) {
		h := &Handlers{}
		c, _ := newTestContext(http.MethodDelete, "/quicklaunches/ql-1", "", map[string]string{"id": "ql-1"}, nil)
		err := h.DeleteQuickLaunchHandler(c)
		assertHTTPError(t, err, http.StatusBadRequest)
	})

	t.Run("begin tx error", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: func(context.Context) (db.Tx, error) { return nil, fmt.Errorf("tx error") },
		}}
		c, _ := newTestContext(http.MethodDelete, "/quicklaunches/ql-1", "", map[string]string{"id": "ql-1"}, map[string]string{"user": "u"})
		err := h.DeleteQuickLaunchHandler(c)
		assertHTTPError(t, err, http.StatusInternalServerError)
	})

	t.Run("db error", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: defaultBeginTx(),
			DeleteQuickLaunchFn: func(_ context.Context, _ db.Tx, id, user string) error {
				return fmt.Errorf("db error")
			},
		}}
		c, _ := newTestContext(http.MethodDelete, "/quicklaunches/ql-1", "", map[string]string{"id": "ql-1"}, map[string]string{"user": "u"})
		err := h.DeleteQuickLaunchHandler(c)
		assertHTTPError(t, err, http.StatusInternalServerError)
	})

	t.Run("success", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: defaultBeginTx(),
			DeleteQuickLaunchFn: func(_ context.Context, _ db.Tx, id, user string) error {
				return nil
			},
		}}
		c, rec := newTestContext(http.MethodDelete, "/quicklaunches/ql-1", "", map[string]string{"id": "ql-1"}, map[string]string{"user": "u"})
		err := h.DeleteQuickLaunchHandler(c)
		assertNoError(t, err)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})
}

func TestQuickLaunchAppInfoHandler(t *testing.T) {
	t.Run("missing user", func(t *testing.T) {
		h := &Handlers{}
		c, _ := newTestContext(http.MethodGet, "/quicklaunches/ql-1/app-info", "", map[string]string{"id": "ql-1"}, nil)
		err := h.QuickLaunchAppInfoHandler(c)
		assertHTTPError(t, err, http.StatusBadRequest)
	})

	t.Run("begin tx error", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: func(context.Context) (db.Tx, error) { return nil, fmt.Errorf("tx error") },
		}}
		c, _ := newTestContext(http.MethodGet, "/quicklaunches/ql-1/app-info", "", map[string]string{"id": "ql-1"}, map[string]string{"user": "u"})
		err := h.QuickLaunchAppInfoHandler(c)
		assertHTTPError(t, err, http.StatusInternalServerError)
	})

	t.Run("ql not found", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: defaultBeginTx(),
			GetQuickLaunchFn: func(_ context.Context, _ db.Tx, id, user string) (*db.QuickLaunch, error) {
				return nil, db.NewNotFoundError("quick launch", id)
			},
		}}
		c, _ := newTestContext(http.MethodGet, "/quicklaunches/ql-1/app-info", "", map[string]string{"id": "ql-1"}, map[string]string{"user": "u"})
		err := h.QuickLaunchAppInfoHandler(c)
		assertHTTPError(t, err, http.StatusNotFound)
	})

	t.Run("app fetch failure", func(t *testing.T) {
		h := &Handlers{
			DB: &mockDB{
				BeginTxFn: defaultBeginTx(),
				GetQuickLaunchFn: func(_ context.Context, _ db.Tx, id, user string) (*db.QuickLaunch, error) {
					return &db.QuickLaunch{
						ID: id, AppID: "app-1", AppVersionID: "v1",
						Submission: json.RawMessage(`{"config":{}}`),
					}, nil
				},
			},
			AppsClient: &mockAppFetcher{
				GetAppVersionFn: func(user, systemID, appID, versionID string) (map[string]any, error) {
					return nil, fmt.Errorf("fetch error")
				},
			},
		}
		c, _ := newTestContext(http.MethodGet, "/quicklaunches/ql-1/app-info", "", map[string]string{"id": "ql-1"}, map[string]string{"user": "u"})
		err := h.QuickLaunchAppInfoHandler(c)
		assertHTTPError(t, err, http.StatusInternalServerError)
	})

	t.Run("success", func(t *testing.T) {
		h := &Handlers{
			DB: &mockDB{
				BeginTxFn: defaultBeginTx(),
				GetQuickLaunchFn: func(_ context.Context, _ db.Tx, id, user string) (*db.QuickLaunch, error) {
					return &db.QuickLaunch{
						ID: id, AppID: "app-1", AppVersionID: "v1",
						Submission: json.RawMessage(`{"config":{"p1":"val"},"debug":false}`),
					}, nil
				},
			},
			AppsClient: &mockAppFetcher{
				GetAppVersionFn: func(user, systemID, appID, versionID string) (map[string]any, error) {
					return map[string]any{
						"groups": []any{
							map[string]any{
								"parameters": []any{
									map[string]any{"id": "p1", "type": "TextInput"},
								},
							},
						},
					}, nil
				},
			},
		}
		c, rec := newTestContext(http.MethodGet, "/quicklaunches/ql-1/app-info", "", map[string]string{"id": "ql-1"}, map[string]string{"user": "u"})
		err := h.QuickLaunchAppInfoHandler(c)
		assertNoError(t, err)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})
}

// ──────────────────────────────────────────────────────────────
// Favorites Handlers
// ──────────────────────────────────────────────────────────────

func TestAddFavoriteHandler(t *testing.T) {
	t.Run("missing user", func(t *testing.T) {
		h := &Handlers{}
		c, _ := newTestContext(http.MethodPost, "/quicklaunch/favorites", `{"quick_launch_id":"ql-1"}`, nil, nil)
		err := h.AddFavoriteHandler(c)
		assertHTTPError(t, err, http.StatusBadRequest)
	})

	t.Run("invalid body", func(t *testing.T) {
		h := &Handlers{}
		c, _ := newTestContext(http.MethodPost, "/quicklaunch/favorites", `{bad`, nil, map[string]string{"user": "u"})
		err := h.AddFavoriteHandler(c)
		assertHTTPError(t, err, http.StatusBadRequest)
	})

	t.Run("begin tx error", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: func(context.Context) (db.Tx, error) { return nil, fmt.Errorf("tx error") },
		}}
		c, _ := newTestContext(http.MethodPost, "/quicklaunch/favorites", `{"quick_launch_id":"ql-1"}`, nil, map[string]string{"user": "u"})
		err := h.AddFavoriteHandler(c)
		assertHTTPError(t, err, http.StatusInternalServerError)
	})

	t.Run("db error", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: defaultBeginTx(),
			AddFavoriteFn: func(_ context.Context, _ db.Tx, user, quickLaunchID string) (*db.QuickLaunchFavorite, error) {
				return nil, fmt.Errorf("db error")
			},
		}}
		c, _ := newTestContext(http.MethodPost, "/quicklaunch/favorites", `{"quick_launch_id":"ql-1"}`, nil, map[string]string{"user": "u"})
		err := h.AddFavoriteHandler(c)
		assertHTTPError(t, err, http.StatusInternalServerError)
	})

	t.Run("success", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: defaultBeginTx(),
			AddFavoriteFn: func(_ context.Context, _ db.Tx, user, quickLaunchID string) (*db.QuickLaunchFavorite, error) {
				return &db.QuickLaunchFavorite{ID: "fav-1", QuickLaunchID: quickLaunchID}, nil
			},
		}}
		c, rec := newTestContext(http.MethodPost, "/quicklaunch/favorites", `{"quick_launch_id":"ql-1"}`, nil, map[string]string{"user": "u"})
		err := h.AddFavoriteHandler(c)
		assertNoError(t, err)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})
}

func TestGetFavoriteHandler(t *testing.T) {
	t.Run("missing user", func(t *testing.T) {
		h := &Handlers{}
		c, _ := newTestContext(http.MethodGet, "/quicklaunch/favorites/fav-1", "", map[string]string{"id": "fav-1"}, nil)
		err := h.GetFavoriteHandler(c)
		assertHTTPError(t, err, http.StatusBadRequest)
	})

	t.Run("begin tx error", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: func(context.Context) (db.Tx, error) { return nil, fmt.Errorf("tx error") },
		}}
		c, _ := newTestContext(http.MethodGet, "/quicklaunch/favorites/fav-1", "", map[string]string{"id": "fav-1"}, map[string]string{"user": "u"})
		err := h.GetFavoriteHandler(c)
		assertHTTPError(t, err, http.StatusInternalServerError)
	})

	t.Run("not found", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: defaultBeginTx(),
			GetFavoriteFn: func(_ context.Context, _ db.Tx, user, favID string) (*db.QuickLaunchFavorite, error) {
				return nil, db.NewNotFoundError("favorite", favID)
			},
		}}
		c, _ := newTestContext(http.MethodGet, "/quicklaunch/favorites/fav-1", "", map[string]string{"id": "fav-1"}, map[string]string{"user": "u"})
		err := h.GetFavoriteHandler(c)
		assertHTTPError(t, err, http.StatusNotFound)
	})

	t.Run("other db error", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: defaultBeginTx(),
			GetFavoriteFn: func(_ context.Context, _ db.Tx, user, favID string) (*db.QuickLaunchFavorite, error) {
				return nil, fmt.Errorf("connection refused")
			},
		}}
		c, _ := newTestContext(http.MethodGet, "/quicklaunch/favorites/fav-1", "", map[string]string{"id": "fav-1"}, map[string]string{"user": "u"})
		err := h.GetFavoriteHandler(c)
		assertHTTPError(t, err, http.StatusInternalServerError)
	})

	t.Run("success", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: defaultBeginTx(),
			GetFavoriteFn: func(_ context.Context, _ db.Tx, user, favID string) (*db.QuickLaunchFavorite, error) {
				return &db.QuickLaunchFavorite{ID: favID}, nil
			},
		}}
		c, rec := newTestContext(http.MethodGet, "/quicklaunch/favorites/fav-1", "", map[string]string{"id": "fav-1"}, map[string]string{"user": "u"})
		err := h.GetFavoriteHandler(c)
		assertNoError(t, err)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})
}

func TestGetAllFavoritesHandler(t *testing.T) {
	t.Run("missing user", func(t *testing.T) {
		h := &Handlers{}
		c, _ := newTestContext(http.MethodGet, "/quicklaunch/favorites", "", nil, nil)
		err := h.GetAllFavoritesHandler(c)
		assertHTTPError(t, err, http.StatusBadRequest)
	})

	t.Run("begin tx error", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: func(context.Context) (db.Tx, error) { return nil, fmt.Errorf("tx error") },
		}}
		c, _ := newTestContext(http.MethodGet, "/quicklaunch/favorites", "", nil, map[string]string{"user": "u"})
		err := h.GetAllFavoritesHandler(c)
		assertHTTPError(t, err, http.StatusInternalServerError)
	})

	t.Run("db error", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: defaultBeginTx(),
			GetAllFavoritesFn: func(_ context.Context, _ db.Tx, user string) ([]db.QuickLaunchFavorite, error) {
				return nil, fmt.Errorf("db error")
			},
		}}
		c, _ := newTestContext(http.MethodGet, "/quicklaunch/favorites", "", nil, map[string]string{"user": "u"})
		err := h.GetAllFavoritesHandler(c)
		assertHTTPError(t, err, http.StatusInternalServerError)
	})

	t.Run("success", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: defaultBeginTx(),
			GetAllFavoritesFn: func(_ context.Context, _ db.Tx, user string) ([]db.QuickLaunchFavorite, error) {
				return []db.QuickLaunchFavorite{}, nil
			},
		}}
		c, rec := newTestContext(http.MethodGet, "/quicklaunch/favorites", "", nil, map[string]string{"user": "u"})
		err := h.GetAllFavoritesHandler(c)
		assertNoError(t, err)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})
}

func TestDeleteFavoriteHandler(t *testing.T) {
	t.Run("missing user", func(t *testing.T) {
		h := &Handlers{}
		c, _ := newTestContext(http.MethodDelete, "/quicklaunch/favorites/fav-1", "", map[string]string{"id": "fav-1"}, nil)
		err := h.DeleteFavoriteHandler(c)
		assertHTTPError(t, err, http.StatusBadRequest)
	})

	t.Run("begin tx error", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: func(context.Context) (db.Tx, error) { return nil, fmt.Errorf("tx error") },
		}}
		c, _ := newTestContext(http.MethodDelete, "/quicklaunch/favorites/fav-1", "", map[string]string{"id": "fav-1"}, map[string]string{"user": "u"})
		err := h.DeleteFavoriteHandler(c)
		assertHTTPError(t, err, http.StatusInternalServerError)
	})

	t.Run("db error", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: defaultBeginTx(),
			DeleteFavoriteFn: func(_ context.Context, _ db.Tx, user, favID string) error {
				return fmt.Errorf("connection refused")
			},
		}}
		c, _ := newTestContext(http.MethodDelete, "/quicklaunch/favorites/fav-1", "", map[string]string{"id": "fav-1"}, map[string]string{"user": "u"})
		err := h.DeleteFavoriteHandler(c)
		assertHTTPError(t, err, http.StatusInternalServerError)
	})

	t.Run("success", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: defaultBeginTx(),
			DeleteFavoriteFn: func(_ context.Context, _ db.Tx, user, favID string) error {
				return nil
			},
		}}
		c, rec := newTestContext(http.MethodDelete, "/quicklaunch/favorites/fav-1", "", map[string]string{"id": "fav-1"}, map[string]string{"user": "u"})
		err := h.DeleteFavoriteHandler(c)
		assertNoError(t, err)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})
}

// ──────────────────────────────────────────────────────────────
// User Default Handlers
// ──────────────────────────────────────────────────────────────

func TestAddUserDefaultHandler(t *testing.T) {
	t.Run("missing user", func(t *testing.T) {
		h := &Handlers{}
		c, _ := newTestContext(http.MethodPost, "/quicklaunch/defaults/user", `{}`, nil, nil)
		err := h.AddUserDefaultHandler(c)
		assertHTTPError(t, err, http.StatusBadRequest)
	})

	t.Run("invalid body", func(t *testing.T) {
		h := &Handlers{}
		c, _ := newTestContext(http.MethodPost, "/quicklaunch/defaults/user", `{bad`, nil, map[string]string{"user": "u"})
		err := h.AddUserDefaultHandler(c)
		assertHTTPError(t, err, http.StatusBadRequest)
	})

	t.Run("begin tx error", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: func(context.Context) (db.Tx, error) { return nil, fmt.Errorf("tx error") },
		}}
		c, _ := newTestContext(http.MethodPost, "/quicklaunch/defaults/user", `{"quick_launch_id":"ql-1","app_id":"a-1"}`, nil, map[string]string{"user": "u"})
		err := h.AddUserDefaultHandler(c)
		assertHTTPError(t, err, http.StatusInternalServerError)
	})

	t.Run("db error", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: defaultBeginTx(),
			AddUserDefaultFn: func(_ context.Context, _ db.Tx, user string, nud *db.NewQuickLaunchUserDefault) (*db.QuickLaunchUserDefault, error) {
				return nil, fmt.Errorf("db error")
			},
		}}
		c, _ := newTestContext(http.MethodPost, "/quicklaunch/defaults/user", `{"quick_launch_id":"ql-1","app_id":"a-1"}`, nil, map[string]string{"user": "u"})
		err := h.AddUserDefaultHandler(c)
		assertHTTPError(t, err, http.StatusInternalServerError)
	})

	t.Run("success", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: defaultBeginTx(),
			AddUserDefaultFn: func(_ context.Context, _ db.Tx, user string, nud *db.NewQuickLaunchUserDefault) (*db.QuickLaunchUserDefault, error) {
				return &db.QuickLaunchUserDefault{ID: "ud-1"}, nil
			},
		}}
		c, rec := newTestContext(http.MethodPost, "/quicklaunch/defaults/user", `{"quick_launch_id":"ql-1","app_id":"a-1"}`, nil, map[string]string{"user": "u"})
		err := h.AddUserDefaultHandler(c)
		assertNoError(t, err)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})
}

func TestGetUserDefaultHandler(t *testing.T) {
	t.Run("missing user", func(t *testing.T) {
		h := &Handlers{}
		c, _ := newTestContext(http.MethodGet, "/quicklaunch/defaults/user/ud-1", "", map[string]string{"id": "ud-1"}, nil)
		err := h.GetUserDefaultHandler(c)
		assertHTTPError(t, err, http.StatusBadRequest)
	})

	t.Run("begin tx error", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: func(context.Context) (db.Tx, error) { return nil, fmt.Errorf("tx error") },
		}}
		c, _ := newTestContext(http.MethodGet, "/quicklaunch/defaults/user/ud-1", "", map[string]string{"id": "ud-1"}, map[string]string{"user": "u"})
		err := h.GetUserDefaultHandler(c)
		assertHTTPError(t, err, http.StatusInternalServerError)
	})

	t.Run("not found", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: defaultBeginTx(),
			GetUserDefaultFn: func(_ context.Context, _ db.Tx, user, id string) (*db.QuickLaunchUserDefault, error) {
				return nil, db.NewNotFoundError("user default", id)
			},
		}}
		c, _ := newTestContext(http.MethodGet, "/quicklaunch/defaults/user/ud-1", "", map[string]string{"id": "ud-1"}, map[string]string{"user": "u"})
		err := h.GetUserDefaultHandler(c)
		assertHTTPError(t, err, http.StatusNotFound)
	})

	t.Run("other db error", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: defaultBeginTx(),
			GetUserDefaultFn: func(_ context.Context, _ db.Tx, user, id string) (*db.QuickLaunchUserDefault, error) {
				return nil, fmt.Errorf("connection refused")
			},
		}}
		c, _ := newTestContext(http.MethodGet, "/quicklaunch/defaults/user/ud-1", "", map[string]string{"id": "ud-1"}, map[string]string{"user": "u"})
		err := h.GetUserDefaultHandler(c)
		assertHTTPError(t, err, http.StatusInternalServerError)
	})

	t.Run("success", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: defaultBeginTx(),
			GetUserDefaultFn: func(_ context.Context, _ db.Tx, user, id string) (*db.QuickLaunchUserDefault, error) {
				return &db.QuickLaunchUserDefault{ID: id}, nil
			},
		}}
		c, rec := newTestContext(http.MethodGet, "/quicklaunch/defaults/user/ud-1", "", map[string]string{"id": "ud-1"}, map[string]string{"user": "u"})
		err := h.GetUserDefaultHandler(c)
		assertNoError(t, err)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})
}

func TestGetAllUserDefaultsHandler(t *testing.T) {
	t.Run("missing user", func(t *testing.T) {
		h := &Handlers{}
		c, _ := newTestContext(http.MethodGet, "/quicklaunch/defaults/user", "", nil, nil)
		err := h.GetAllUserDefaultsHandler(c)
		assertHTTPError(t, err, http.StatusBadRequest)
	})

	t.Run("begin tx error", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: func(context.Context) (db.Tx, error) { return nil, fmt.Errorf("tx error") },
		}}
		c, _ := newTestContext(http.MethodGet, "/quicklaunch/defaults/user", "", nil, map[string]string{"user": "u"})
		err := h.GetAllUserDefaultsHandler(c)
		assertHTTPError(t, err, http.StatusInternalServerError)
	})

	t.Run("db error", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: defaultBeginTx(),
			GetAllUserDefaultsFn: func(_ context.Context, _ db.Tx, user string) ([]db.QuickLaunchUserDefault, error) {
				return nil, fmt.Errorf("db error")
			},
		}}
		c, _ := newTestContext(http.MethodGet, "/quicklaunch/defaults/user", "", nil, map[string]string{"user": "u"})
		err := h.GetAllUserDefaultsHandler(c)
		assertHTTPError(t, err, http.StatusInternalServerError)
	})

	t.Run("success", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: defaultBeginTx(),
			GetAllUserDefaultsFn: func(_ context.Context, _ db.Tx, user string) ([]db.QuickLaunchUserDefault, error) {
				return []db.QuickLaunchUserDefault{}, nil
			},
		}}
		c, rec := newTestContext(http.MethodGet, "/quicklaunch/defaults/user", "", nil, map[string]string{"user": "u"})
		err := h.GetAllUserDefaultsHandler(c)
		assertNoError(t, err)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})
}

func TestUpdateUserDefaultHandler(t *testing.T) {
	t.Run("missing user", func(t *testing.T) {
		h := &Handlers{}
		c, _ := newTestContext(http.MethodPatch, "/quicklaunch/defaults/user/ud-1", `{}`, map[string]string{"id": "ud-1"}, nil)
		err := h.UpdateUserDefaultHandler(c)
		assertHTTPError(t, err, http.StatusBadRequest)
	})

	t.Run("invalid body", func(t *testing.T) {
		h := &Handlers{}
		c, _ := newTestContext(http.MethodPatch, "/quicklaunch/defaults/user/ud-1", `{bad`, map[string]string{"id": "ud-1"}, map[string]string{"user": "u"})
		err := h.UpdateUserDefaultHandler(c)
		assertHTTPError(t, err, http.StatusBadRequest)
	})

	t.Run("missing required fields", func(t *testing.T) {
		h := &Handlers{}
		c, _ := newTestContext(http.MethodPatch, "/quicklaunch/defaults/user/ud-1", `{}`, map[string]string{"id": "ud-1"}, map[string]string{"user": "u"})
		err := h.UpdateUserDefaultHandler(c)
		assertHTTPError(t, err, http.StatusBadRequest)
	})

	t.Run("begin tx error", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: func(context.Context) (db.Tx, error) { return nil, fmt.Errorf("tx error") },
		}}
		c, _ := newTestContext(http.MethodPatch, "/quicklaunch/defaults/user/ud-1", `{"quick_launch_id":"ql-1","app_id":"a-1"}`, map[string]string{"id": "ud-1"}, map[string]string{"user": "u"})
		err := h.UpdateUserDefaultHandler(c)
		assertHTTPError(t, err, http.StatusInternalServerError)
	})

	t.Run("db error", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: defaultBeginTx(),
			UpdateUserDefaultFn: func(_ context.Context, _ db.Tx, id, user string, update *db.UpdateQuickLaunchUserDefaultRequest) (*db.QuickLaunchUserDefault, error) {
				return nil, fmt.Errorf("db error")
			},
		}}
		c, _ := newTestContext(http.MethodPatch, "/quicklaunch/defaults/user/ud-1", `{"quick_launch_id":"ql-1","app_id":"a-1"}`, map[string]string{"id": "ud-1"}, map[string]string{"user": "u"})
		err := h.UpdateUserDefaultHandler(c)
		assertHTTPError(t, err, http.StatusInternalServerError)
	})

	t.Run("success", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: defaultBeginTx(),
			UpdateUserDefaultFn: func(_ context.Context, _ db.Tx, id, user string, update *db.UpdateQuickLaunchUserDefaultRequest) (*db.QuickLaunchUserDefault, error) {
				return &db.QuickLaunchUserDefault{ID: id}, nil
			},
		}}
		c, rec := newTestContext(http.MethodPatch, "/quicklaunch/defaults/user/ud-1", `{"quick_launch_id":"ql-1","app_id":"a-1"}`, map[string]string{"id": "ud-1"}, map[string]string{"user": "u"})
		err := h.UpdateUserDefaultHandler(c)
		assertNoError(t, err)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})
}

func TestDeleteUserDefaultHandler(t *testing.T) {
	t.Run("missing user", func(t *testing.T) {
		h := &Handlers{}
		c, _ := newTestContext(http.MethodDelete, "/quicklaunch/defaults/user/ud-1", "", map[string]string{"id": "ud-1"}, nil)
		err := h.DeleteUserDefaultHandler(c)
		assertHTTPError(t, err, http.StatusBadRequest)
	})

	t.Run("begin tx error", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: func(context.Context) (db.Tx, error) { return nil, fmt.Errorf("tx error") },
		}}
		c, _ := newTestContext(http.MethodDelete, "/quicklaunch/defaults/user/ud-1", "", map[string]string{"id": "ud-1"}, map[string]string{"user": "u"})
		err := h.DeleteUserDefaultHandler(c)
		assertHTTPError(t, err, http.StatusInternalServerError)
	})

	t.Run("db error", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: defaultBeginTx(),
			DeleteUserDefaultFn: func(_ context.Context, _ db.Tx, user, id string) error {
				return fmt.Errorf("connection refused")
			},
		}}
		c, _ := newTestContext(http.MethodDelete, "/quicklaunch/defaults/user/ud-1", "", map[string]string{"id": "ud-1"}, map[string]string{"user": "u"})
		err := h.DeleteUserDefaultHandler(c)
		assertHTTPError(t, err, http.StatusInternalServerError)
	})

	t.Run("success", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: defaultBeginTx(),
			DeleteUserDefaultFn: func(_ context.Context, _ db.Tx, user, id string) error {
				return nil
			},
		}}
		c, rec := newTestContext(http.MethodDelete, "/quicklaunch/defaults/user/ud-1", "", map[string]string{"id": "ud-1"}, map[string]string{"user": "u"})
		err := h.DeleteUserDefaultHandler(c)
		assertNoError(t, err)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})
}

// ──────────────────────────────────────────────────────────────
// Global Default Handlers
// ──────────────────────────────────────────────────────────────

func TestAddGlobalDefaultHandler(t *testing.T) {
	t.Run("missing user", func(t *testing.T) {
		h := &Handlers{}
		c, _ := newTestContext(http.MethodPost, "/quicklaunch/defaults/global", `{}`, nil, nil)
		err := h.AddGlobalDefaultHandler(c)
		assertHTTPError(t, err, http.StatusBadRequest)
	})

	t.Run("invalid body", func(t *testing.T) {
		h := &Handlers{}
		c, _ := newTestContext(http.MethodPost, "/quicklaunch/defaults/global", `{bad`, nil, map[string]string{"user": "u"})
		err := h.AddGlobalDefaultHandler(c)
		assertHTTPError(t, err, http.StatusBadRequest)
	})

	t.Run("begin tx error", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: func(context.Context) (db.Tx, error) { return nil, fmt.Errorf("tx error") },
		}}
		c, _ := newTestContext(http.MethodPost, "/quicklaunch/defaults/global", `{"app_id":"a-1","quick_launch_id":"ql-1"}`, nil, map[string]string{"user": "u"})
		err := h.AddGlobalDefaultHandler(c)
		assertHTTPError(t, err, http.StatusInternalServerError)
	})

	t.Run("db error", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: defaultBeginTx(),
			AddGlobalDefaultFn: func(_ context.Context, _ db.Tx, user string, ngd *db.NewQuickLaunchGlobalDefault) (*db.QuickLaunchGlobalDefault, error) {
				return nil, fmt.Errorf("db error")
			},
		}}
		c, _ := newTestContext(http.MethodPost, "/quicklaunch/defaults/global", `{"app_id":"a-1","quick_launch_id":"ql-1"}`, nil, map[string]string{"user": "u"})
		err := h.AddGlobalDefaultHandler(c)
		assertHTTPError(t, err, http.StatusInternalServerError)
	})

	t.Run("success", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: defaultBeginTx(),
			AddGlobalDefaultFn: func(_ context.Context, _ db.Tx, user string, ngd *db.NewQuickLaunchGlobalDefault) (*db.QuickLaunchGlobalDefault, error) {
				return &db.QuickLaunchGlobalDefault{ID: "gd-1"}, nil
			},
		}}
		c, rec := newTestContext(http.MethodPost, "/quicklaunch/defaults/global", `{"app_id":"a-1","quick_launch_id":"ql-1"}`, nil, map[string]string{"user": "u"})
		err := h.AddGlobalDefaultHandler(c)
		assertNoError(t, err)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})
}

func TestGetGlobalDefaultHandler(t *testing.T) {
	t.Run("missing user", func(t *testing.T) {
		h := &Handlers{}
		c, _ := newTestContext(http.MethodGet, "/quicklaunch/defaults/global/gd-1", "", map[string]string{"id": "gd-1"}, nil)
		err := h.GetGlobalDefaultHandler(c)
		assertHTTPError(t, err, http.StatusBadRequest)
	})

	t.Run("begin tx error", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: func(context.Context) (db.Tx, error) { return nil, fmt.Errorf("tx error") },
		}}
		c, _ := newTestContext(http.MethodGet, "/quicklaunch/defaults/global/gd-1", "", map[string]string{"id": "gd-1"}, map[string]string{"user": "u"})
		err := h.GetGlobalDefaultHandler(c)
		assertHTTPError(t, err, http.StatusInternalServerError)
	})

	t.Run("not found", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: defaultBeginTx(),
			GetGlobalDefaultFn: func(_ context.Context, _ db.Tx, user, id string) (*db.QuickLaunchGlobalDefault, error) {
				return nil, db.NewNotFoundError("global default", id)
			},
		}}
		c, _ := newTestContext(http.MethodGet, "/quicklaunch/defaults/global/gd-1", "", map[string]string{"id": "gd-1"}, map[string]string{"user": "u"})
		err := h.GetGlobalDefaultHandler(c)
		assertHTTPError(t, err, http.StatusNotFound)
	})

	t.Run("other db error", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: defaultBeginTx(),
			GetGlobalDefaultFn: func(_ context.Context, _ db.Tx, user, id string) (*db.QuickLaunchGlobalDefault, error) {
				return nil, fmt.Errorf("connection refused")
			},
		}}
		c, _ := newTestContext(http.MethodGet, "/quicklaunch/defaults/global/gd-1", "", map[string]string{"id": "gd-1"}, map[string]string{"user": "u"})
		err := h.GetGlobalDefaultHandler(c)
		assertHTTPError(t, err, http.StatusInternalServerError)
	})

	t.Run("success", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: defaultBeginTx(),
			GetGlobalDefaultFn: func(_ context.Context, _ db.Tx, user, id string) (*db.QuickLaunchGlobalDefault, error) {
				return &db.QuickLaunchGlobalDefault{ID: id}, nil
			},
		}}
		c, rec := newTestContext(http.MethodGet, "/quicklaunch/defaults/global/gd-1", "", map[string]string{"id": "gd-1"}, map[string]string{"user": "u"})
		err := h.GetGlobalDefaultHandler(c)
		assertNoError(t, err)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})
}

func TestGetAllGlobalDefaultsHandler(t *testing.T) {
	t.Run("missing user", func(t *testing.T) {
		h := &Handlers{}
		c, _ := newTestContext(http.MethodGet, "/quicklaunch/defaults/global", "", nil, nil)
		err := h.GetAllGlobalDefaultsHandler(c)
		assertHTTPError(t, err, http.StatusBadRequest)
	})

	t.Run("begin tx error", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: func(context.Context) (db.Tx, error) { return nil, fmt.Errorf("tx error") },
		}}
		c, _ := newTestContext(http.MethodGet, "/quicklaunch/defaults/global", "", nil, map[string]string{"user": "u"})
		err := h.GetAllGlobalDefaultsHandler(c)
		assertHTTPError(t, err, http.StatusInternalServerError)
	})

	t.Run("db error", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: defaultBeginTx(),
			GetAllGlobalDefaultsFn: func(_ context.Context, _ db.Tx, user string) ([]db.QuickLaunchGlobalDefault, error) {
				return nil, fmt.Errorf("db error")
			},
		}}
		c, _ := newTestContext(http.MethodGet, "/quicklaunch/defaults/global", "", nil, map[string]string{"user": "u"})
		err := h.GetAllGlobalDefaultsHandler(c)
		assertHTTPError(t, err, http.StatusInternalServerError)
	})

	t.Run("success", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: defaultBeginTx(),
			GetAllGlobalDefaultsFn: func(_ context.Context, _ db.Tx, user string) ([]db.QuickLaunchGlobalDefault, error) {
				return []db.QuickLaunchGlobalDefault{}, nil
			},
		}}
		c, rec := newTestContext(http.MethodGet, "/quicklaunch/defaults/global", "", nil, map[string]string{"user": "u"})
		err := h.GetAllGlobalDefaultsHandler(c)
		assertNoError(t, err)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})
}

func TestUpdateGlobalDefaultHandler(t *testing.T) {
	t.Run("missing user", func(t *testing.T) {
		h := &Handlers{}
		c, _ := newTestContext(http.MethodPatch, "/quicklaunch/defaults/global/gd-1", `{}`, map[string]string{"id": "gd-1"}, nil)
		err := h.UpdateGlobalDefaultHandler(c)
		assertHTTPError(t, err, http.StatusBadRequest)
	})

	t.Run("invalid body", func(t *testing.T) {
		h := &Handlers{}
		c, _ := newTestContext(http.MethodPatch, "/quicklaunch/defaults/global/gd-1", `{bad`, map[string]string{"id": "gd-1"}, map[string]string{"user": "u"})
		err := h.UpdateGlobalDefaultHandler(c)
		assertHTTPError(t, err, http.StatusBadRequest)
	})

	t.Run("missing required fields", func(t *testing.T) {
		h := &Handlers{}
		c, _ := newTestContext(http.MethodPatch, "/quicklaunch/defaults/global/gd-1", `{}`, map[string]string{"id": "gd-1"}, map[string]string{"user": "u"})
		err := h.UpdateGlobalDefaultHandler(c)
		assertHTTPError(t, err, http.StatusBadRequest)
	})

	t.Run("begin tx error", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: func(context.Context) (db.Tx, error) { return nil, fmt.Errorf("tx error") },
		}}
		c, _ := newTestContext(http.MethodPatch, "/quicklaunch/defaults/global/gd-1", `{"app_id":"a-1","quick_launch_id":"ql-1"}`, map[string]string{"id": "gd-1"}, map[string]string{"user": "u"})
		err := h.UpdateGlobalDefaultHandler(c)
		assertHTTPError(t, err, http.StatusInternalServerError)
	})

	t.Run("db error", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: defaultBeginTx(),
			UpdateGlobalDefaultFn: func(_ context.Context, _ db.Tx, id, user string, update *db.UpdateQuickLaunchGlobalDefaultRequest) (*db.QuickLaunchGlobalDefault, error) {
				return nil, fmt.Errorf("db error")
			},
		}}
		c, _ := newTestContext(http.MethodPatch, "/quicklaunch/defaults/global/gd-1", `{"app_id":"a-1","quick_launch_id":"ql-1"}`, map[string]string{"id": "gd-1"}, map[string]string{"user": "u"})
		err := h.UpdateGlobalDefaultHandler(c)
		assertHTTPError(t, err, http.StatusInternalServerError)
	})

	t.Run("success", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: defaultBeginTx(),
			UpdateGlobalDefaultFn: func(_ context.Context, _ db.Tx, id, user string, update *db.UpdateQuickLaunchGlobalDefaultRequest) (*db.QuickLaunchGlobalDefault, error) {
				return &db.QuickLaunchGlobalDefault{ID: id}, nil
			},
		}}
		c, rec := newTestContext(http.MethodPatch, "/quicklaunch/defaults/global/gd-1", `{"app_id":"a-1","quick_launch_id":"ql-1"}`, map[string]string{"id": "gd-1"}, map[string]string{"user": "u"})
		err := h.UpdateGlobalDefaultHandler(c)
		assertNoError(t, err)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})
}

func TestDeleteGlobalDefaultHandler(t *testing.T) {
	t.Run("missing user", func(t *testing.T) {
		h := &Handlers{}
		c, _ := newTestContext(http.MethodDelete, "/quicklaunch/defaults/global/gd-1", "", map[string]string{"id": "gd-1"}, nil)
		err := h.DeleteGlobalDefaultHandler(c)
		assertHTTPError(t, err, http.StatusBadRequest)
	})

	t.Run("begin tx error", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: func(context.Context) (db.Tx, error) { return nil, fmt.Errorf("tx error") },
		}}
		c, _ := newTestContext(http.MethodDelete, "/quicklaunch/defaults/global/gd-1", "", map[string]string{"id": "gd-1"}, map[string]string{"user": "u"})
		err := h.DeleteGlobalDefaultHandler(c)
		assertHTTPError(t, err, http.StatusInternalServerError)
	})

	t.Run("db error", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: defaultBeginTx(),
			DeleteGlobalDefaultFn: func(_ context.Context, _ db.Tx, user, id string) error {
				return fmt.Errorf("connection refused")
			},
		}}
		c, _ := newTestContext(http.MethodDelete, "/quicklaunch/defaults/global/gd-1", "", map[string]string{"id": "gd-1"}, map[string]string{"user": "u"})
		err := h.DeleteGlobalDefaultHandler(c)
		assertHTTPError(t, err, http.StatusInternalServerError)
	})

	t.Run("success", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: defaultBeginTx(),
			DeleteGlobalDefaultFn: func(_ context.Context, _ db.Tx, user, id string) error {
				return nil
			},
		}}
		c, rec := newTestContext(http.MethodDelete, "/quicklaunch/defaults/global/gd-1", "", map[string]string{"id": "gd-1"}, map[string]string{"user": "u"})
		err := h.DeleteGlobalDefaultHandler(c)
		assertNoError(t, err)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})
}

// ──────────────────────────────────────────────────────────────
// Settings (concurrent job limits) Handlers
// ──────────────────────────────────────────────────────────────

func TestListConcurrentJobLimitsHandler(t *testing.T) {
	t.Run("begin tx error", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: func(context.Context) (db.Tx, error) { return nil, fmt.Errorf("tx error") },
		}}
		c, _ := newTestContext(http.MethodGet, "/settings/concurrent-job-limits", "", nil, nil)
		err := h.ListConcurrentJobLimitsHandler(c)
		assertHTTPError(t, err, http.StatusInternalServerError)
	})

	t.Run("db error", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: defaultBeginTx(),
			ListConcurrentJobLimitsFn: func(_ context.Context, _ db.Tx) ([]db.ConcurrentJobLimit, error) {
				return nil, fmt.Errorf("db error")
			},
		}}
		c, _ := newTestContext(http.MethodGet, "/settings/concurrent-job-limits", "", nil, nil)
		err := h.ListConcurrentJobLimitsHandler(c)
		assertHTTPError(t, err, http.StatusInternalServerError)
	})

	t.Run("success", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: defaultBeginTx(),
			ListConcurrentJobLimitsFn: func(_ context.Context, _ db.Tx) ([]db.ConcurrentJobLimit, error) {
				return []db.ConcurrentJobLimit{{ConcurrentJobs: 8, IsDefault: true}}, nil
			},
		}}
		c, rec := newTestContext(http.MethodGet, "/settings/concurrent-job-limits", "", nil, nil)
		err := h.ListConcurrentJobLimitsHandler(c)
		assertNoError(t, err)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})
}

func TestGetConcurrentJobLimitHandler(t *testing.T) {
	t.Run("begin tx error", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: func(context.Context) (db.Tx, error) { return nil, fmt.Errorf("tx error") },
		}}
		c, _ := newTestContext(http.MethodGet, "/settings/concurrent-job-limits/testuser", "", map[string]string{"username": "testuser"}, nil)
		err := h.GetConcurrentJobLimitHandler(c)
		assertHTTPError(t, err, http.StatusInternalServerError)
	})

	t.Run("not found", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: defaultBeginTx(),
			GetConcurrentJobLimitFn: func(_ context.Context, _ db.Tx, username string) (*db.ConcurrentJobLimit, error) {
				return nil, db.NewNotFoundError("job limit", username)
			},
		}}
		c, _ := newTestContext(http.MethodGet, "/settings/concurrent-job-limits/testuser", "", map[string]string{"username": "testuser"}, nil)
		err := h.GetConcurrentJobLimitHandler(c)
		assertHTTPError(t, err, http.StatusNotFound)
	})

	t.Run("success", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: defaultBeginTx(),
			GetConcurrentJobLimitFn: func(_ context.Context, _ db.Tx, username string) (*db.ConcurrentJobLimit, error) {
				return &db.ConcurrentJobLimit{ConcurrentJobs: 4}, nil
			},
		}}
		c, rec := newTestContext(http.MethodGet, "/settings/concurrent-job-limits/testuser", "", map[string]string{"username": "testuser"}, nil)
		err := h.GetConcurrentJobLimitHandler(c)
		assertNoError(t, err)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})
}

func TestSetConcurrentJobLimitHandler(t *testing.T) {
	t.Run("invalid body", func(t *testing.T) {
		h := &Handlers{}
		c, _ := newTestContext(http.MethodPut, "/settings/concurrent-job-limits/testuser", `{bad`, map[string]string{"username": "testuser"}, nil)
		err := h.SetConcurrentJobLimitHandler(c)
		assertHTTPError(t, err, http.StatusBadRequest)
	})

	t.Run("begin tx error", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: func(context.Context) (db.Tx, error) { return nil, fmt.Errorf("tx error") },
		}}
		c, _ := newTestContext(http.MethodPut, "/settings/concurrent-job-limits/testuser", `{"concurrent_jobs":10}`, map[string]string{"username": "testuser"}, nil)
		err := h.SetConcurrentJobLimitHandler(c)
		assertHTTPError(t, err, http.StatusInternalServerError)
	})

	t.Run("db error", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: defaultBeginTx(),
			SetConcurrentJobLimitFn: func(_ context.Context, _ db.Tx, username string, limit int) (*db.ConcurrentJobLimit, error) {
				return nil, fmt.Errorf("db error")
			},
		}}
		c, _ := newTestContext(http.MethodPut, "/settings/concurrent-job-limits/testuser", `{"concurrent_jobs":10}`, map[string]string{"username": "testuser"}, nil)
		err := h.SetConcurrentJobLimitHandler(c)
		assertHTTPError(t, err, http.StatusInternalServerError)
	})

	t.Run("success", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: defaultBeginTx(),
			SetConcurrentJobLimitFn: func(_ context.Context, _ db.Tx, username string, limit int) (*db.ConcurrentJobLimit, error) {
				return &db.ConcurrentJobLimit{ConcurrentJobs: limit}, nil
			},
		}}
		c, rec := newTestContext(http.MethodPut, "/settings/concurrent-job-limits/testuser", `{"concurrent_jobs":10}`, map[string]string{"username": "testuser"}, nil)
		err := h.SetConcurrentJobLimitHandler(c)
		assertNoError(t, err)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})
}

func TestRemoveConcurrentJobLimitHandler(t *testing.T) {
	t.Run("begin tx error", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: func(context.Context) (db.Tx, error) { return nil, fmt.Errorf("tx error") },
		}}
		c, _ := newTestContext(http.MethodDelete, "/settings/concurrent-job-limits/testuser", "", map[string]string{"username": "testuser"}, nil)
		err := h.RemoveConcurrentJobLimitHandler(c)
		assertHTTPError(t, err, http.StatusInternalServerError)
	})

	t.Run("db error", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: defaultBeginTx(),
			RemoveConcurrentJobLimitFn: func(_ context.Context, _ db.Tx, username string) (*db.ConcurrentJobLimit, error) {
				return nil, fmt.Errorf("db error")
			},
		}}
		c, _ := newTestContext(http.MethodDelete, "/settings/concurrent-job-limits/testuser", "", map[string]string{"username": "testuser"}, nil)
		err := h.RemoveConcurrentJobLimitHandler(c)
		assertHTTPError(t, err, http.StatusInternalServerError)
	})

	t.Run("success", func(t *testing.T) {
		h := &Handlers{DB: &mockDB{
			BeginTxFn: defaultBeginTx(),
			RemoveConcurrentJobLimitFn: func(_ context.Context, _ db.Tx, username string) (*db.ConcurrentJobLimit, error) {
				return &db.ConcurrentJobLimit{ConcurrentJobs: 8, IsDefault: true}, nil
			},
		}}
		c, rec := newTestContext(http.MethodDelete, "/settings/concurrent-job-limits/testuser", "", map[string]string{"username": "testuser"}, nil)
		err := h.RemoveConcurrentJobLimitHandler(c)
		assertNoError(t, err)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})
}
