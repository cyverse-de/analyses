package httphandlers

import (
	"context"
	"encoding/json"

	"github.com/cyverse-de/analyses/clients"
	"github.com/cyverse-de/analyses/db"
)

// DatabaseStore defines all database operations used by HTTP handlers.
type DatabaseStore interface {
	BeginTx(ctx context.Context) (db.Tx, error)

	// Quick Launches
	GetQuickLaunch(ctx context.Context, tx db.Tx, id, user string) (*db.QuickLaunch, error)
	GetAllQuickLaunches(ctx context.Context, tx db.Tx, user string) ([]db.QuickLaunch, error)
	GetQuickLaunchesByApp(ctx context.Context, tx db.Tx, appID, user string) ([]db.QuickLaunch, error)
	AddQuickLaunch(ctx context.Context, tx db.Tx, user string, nql *db.NewQuickLaunch) (*db.QuickLaunch, error)
	UpdateQuickLaunch(ctx context.Context, tx db.Tx, id, user string, uql *db.UpdateQuickLaunchRequest) (*db.QuickLaunch, error)
	DeleteQuickLaunch(ctx context.Context, tx db.Tx, id, user string) error
	MergeSubmission(ctx context.Context, tx db.Tx, qlID, user string, newSubmission json.RawMessage) (json.RawMessage, error)

	// Favorites
	GetAllFavorites(ctx context.Context, tx db.Tx, user string) ([]db.QuickLaunchFavorite, error)
	GetFavorite(ctx context.Context, tx db.Tx, user, favID string) (*db.QuickLaunchFavorite, error)
	AddFavorite(ctx context.Context, tx db.Tx, user, quickLaunchID string) (*db.QuickLaunchFavorite, error)
	DeleteFavorite(ctx context.Context, tx db.Tx, user, favID string) error

	// User Defaults
	GetUserDefault(ctx context.Context, tx db.Tx, user, id string) (*db.QuickLaunchUserDefault, error)
	GetAllUserDefaults(ctx context.Context, tx db.Tx, user string) ([]db.QuickLaunchUserDefault, error)
	AddUserDefault(ctx context.Context, tx db.Tx, user string, nud *db.NewQuickLaunchUserDefault) (*db.QuickLaunchUserDefault, error)
	UpdateUserDefault(ctx context.Context, tx db.Tx, id, user string, update *db.UpdateQuickLaunchUserDefaultRequest) (*db.QuickLaunchUserDefault, error)
	DeleteUserDefault(ctx context.Context, tx db.Tx, user, id string) error

	// Global Defaults
	GetGlobalDefault(ctx context.Context, tx db.Tx, user, id string) (*db.QuickLaunchGlobalDefault, error)
	GetAllGlobalDefaults(ctx context.Context, tx db.Tx, user string) ([]db.QuickLaunchGlobalDefault, error)
	AddGlobalDefault(ctx context.Context, tx db.Tx, user string, ngd *db.NewQuickLaunchGlobalDefault) (*db.QuickLaunchGlobalDefault, error)
	UpdateGlobalDefault(ctx context.Context, tx db.Tx, id, user string, update *db.UpdateQuickLaunchGlobalDefaultRequest) (*db.QuickLaunchGlobalDefault, error)
	DeleteGlobalDefault(ctx context.Context, tx db.Tx, user, id string) error

	// Settings
	ListConcurrentJobLimits(ctx context.Context, tx db.Tx) ([]db.ConcurrentJobLimit, error)
	GetConcurrentJobLimit(ctx context.Context, tx db.Tx, username string) (*db.ConcurrentJobLimit, error)
	SetConcurrentJobLimit(ctx context.Context, tx db.Tx, username string, limit int) (*db.ConcurrentJobLimit, error)
	RemoveConcurrentJobLimit(ctx context.Context, tx db.Tx, username string) (*db.ConcurrentJobLimit, error)
}

// AppFetcher retrieves app definitions from the apps service.
type AppFetcher interface {
	GetApp(user, systemID, appID string) (map[string]any, error)
	GetAppVersion(user, systemID, appID, versionID string) (map[string]any, error)
}

// PathChecker is the canonical interface for verifying path accessibility.
// Defined in the clients package.
type PathChecker = clients.PathChecker
