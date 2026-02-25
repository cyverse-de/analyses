package httphandlers

import (
	"encoding/json"

	"github.com/cyverse-de/analyses/db"
)

// DatabaseStore defines all database operations used by HTTP handlers.
type DatabaseStore interface {
	// Quick Launches
	GetQuickLaunch(id, user string) (*db.QuickLaunch, error)
	GetAllQuickLaunches(user string) ([]db.QuickLaunch, error)
	GetQuickLaunchesByApp(appID, user string) ([]db.QuickLaunch, error)
	AddQuickLaunch(user string, nql *db.NewQuickLaunch) (*db.QuickLaunch, error)
	UpdateQuickLaunch(id, user string, uql *db.UpdateQuickLaunchRequest) (*db.QuickLaunch, error)
	DeleteQuickLaunch(id, user string) error
	MergeSubmission(qlID, user string, newSubmission json.RawMessage) (json.RawMessage, error)

	// Favorites
	GetAllFavorites(user string) ([]db.QuickLaunchFavorite, error)
	GetFavorite(user, favID string) (*db.QuickLaunchFavorite, error)
	AddFavorite(user, quickLaunchID string) (*db.QuickLaunchFavorite, error)
	DeleteFavorite(user, favID string) error

	// User Defaults
	GetUserDefault(user, id string) (*db.QuickLaunchUserDefault, error)
	GetAllUserDefaults(user string) ([]db.QuickLaunchUserDefault, error)
	AddUserDefault(user string, nud *db.NewQuickLaunchUserDefault) (*db.QuickLaunchUserDefault, error)
	UpdateUserDefault(id, user string, update *db.UpdateQuickLaunchUserDefaultRequest) (*db.QuickLaunchUserDefault, error)
	DeleteUserDefault(user, id string) error

	// Global Defaults
	GetGlobalDefault(user, id string) (*db.QuickLaunchGlobalDefault, error)
	GetAllGlobalDefaults(user string) ([]db.QuickLaunchGlobalDefault, error)
	AddGlobalDefault(user string, ngd *db.NewQuickLaunchGlobalDefault) (*db.QuickLaunchGlobalDefault, error)
	UpdateGlobalDefault(id, user string, update *db.UpdateQuickLaunchGlobalDefaultRequest) (*db.QuickLaunchGlobalDefault, error)
	DeleteGlobalDefault(user, id string) error

	// Settings
	ListConcurrentJobLimits() ([]db.ConcurrentJobLimit, error)
	GetConcurrentJobLimit(username string) (*db.ConcurrentJobLimit, error)
	SetConcurrentJobLimit(username string, limit int) (*db.ConcurrentJobLimit, error)
	RemoveConcurrentJobLimit(username string) (*db.ConcurrentJobLimit, error)
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
