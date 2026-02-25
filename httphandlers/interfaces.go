package httphandlers

import (
	"encoding/json"

	"github.com/cyverse-de/analyses/db"
)

// DatabaseStore defines all database operations used by HTTP handlers.
type DatabaseStore interface {
	BeginTx() (db.Tx, error)

	// Quick Launches
	GetQuickLaunch(tx db.Tx, id, user string) (*db.QuickLaunch, error)
	GetAllQuickLaunches(tx db.Tx, user string) ([]db.QuickLaunch, error)
	GetQuickLaunchesByApp(tx db.Tx, appID, user string) ([]db.QuickLaunch, error)
	AddQuickLaunch(tx db.Tx, user string, nql *db.NewQuickLaunch) (*db.QuickLaunch, error)
	UpdateQuickLaunch(tx db.Tx, id, user string, uql *db.UpdateQuickLaunchRequest) (*db.QuickLaunch, error)
	DeleteQuickLaunch(tx db.Tx, id, user string) error
	MergeSubmission(tx db.Tx, qlID, user string, newSubmission json.RawMessage) (json.RawMessage, error)

	// Favorites
	GetAllFavorites(tx db.Tx, user string) ([]db.QuickLaunchFavorite, error)
	GetFavorite(tx db.Tx, user, favID string) (*db.QuickLaunchFavorite, error)
	AddFavorite(tx db.Tx, user, quickLaunchID string) (*db.QuickLaunchFavorite, error)
	DeleteFavorite(tx db.Tx, user, favID string) error

	// User Defaults
	GetUserDefault(tx db.Tx, user, id string) (*db.QuickLaunchUserDefault, error)
	GetAllUserDefaults(tx db.Tx, user string) ([]db.QuickLaunchUserDefault, error)
	AddUserDefault(tx db.Tx, user string, nud *db.NewQuickLaunchUserDefault) (*db.QuickLaunchUserDefault, error)
	UpdateUserDefault(tx db.Tx, id, user string, update *db.UpdateQuickLaunchUserDefaultRequest) (*db.QuickLaunchUserDefault, error)
	DeleteUserDefault(tx db.Tx, user, id string) error

	// Global Defaults
	GetGlobalDefault(tx db.Tx, user, id string) (*db.QuickLaunchGlobalDefault, error)
	GetAllGlobalDefaults(tx db.Tx, user string) ([]db.QuickLaunchGlobalDefault, error)
	AddGlobalDefault(tx db.Tx, user string, ngd *db.NewQuickLaunchGlobalDefault) (*db.QuickLaunchGlobalDefault, error)
	UpdateGlobalDefault(tx db.Tx, id, user string, update *db.UpdateQuickLaunchGlobalDefaultRequest) (*db.QuickLaunchGlobalDefault, error)
	DeleteGlobalDefault(tx db.Tx, user, id string) error

	// Settings
	ListConcurrentJobLimits(tx db.Tx) ([]db.ConcurrentJobLimit, error)
	GetConcurrentJobLimit(tx db.Tx, username string) (*db.ConcurrentJobLimit, error)
	SetConcurrentJobLimit(tx db.Tx, username string, limit int) (*db.ConcurrentJobLimit, error)
	RemoveConcurrentJobLimit(tx db.Tx, username string) (*db.ConcurrentJobLimit, error)
}

// AppFetcher retrieves app definitions from the apps service.
type AppFetcher interface {
	GetApp(user, systemID, appID string) (map[string]any, error)
	GetAppVersion(user, systemID, appID, versionID string) (map[string]any, error)
}

// PathChecker verifies path accessibility for a user.
type PathChecker interface {
	PathsAccessibleBy(paths []string, user string) (bool, error)
}
