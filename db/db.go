package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// NotFoundError indicates that a requested resource was not found.
type NotFoundError struct {
	Type string
	ID   string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s not found: %s", e.Type, e.ID)
}

// NewNotFoundError creates a new NotFoundError.
func NewNotFoundError(typeName, id string) *NotFoundError {
	return &NotFoundError{Type: typeName, ID: id}
}

// IsNotFound returns true if the error indicates a resource was not found.
func IsNotFound(err error) bool {
	var nfe *NotFoundError
	return errors.As(err, &nfe)
}

// Tx is the interface for database operations within a transaction.
// All query methods require a context for cancellation and tracing.
type Tx interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryRowxContext(ctx context.Context, query string, args ...any) *sqlx.Row
	SelectContext(ctx context.Context, dest any, query string, args ...any) error
	Commit() error
	Rollback() error
}

// Database provides access to the analyses database.
type Database struct {
	db *sqlx.DB
}

// New returns a new Database instance.
func New(db *sqlx.DB) *Database {
	return &Database{db: db}
}

// BeginTx starts a new database transaction.
func (d *Database) BeginTx(ctx context.Context) (Tx, error) {
	return d.db.BeginTxx(ctx, nil)
}

// GetUserID returns the user ID for the given username.
func (d *Database) GetUserID(ctx context.Context, tx Tx, username string) (string, error) {
	var id string
	err := tx.QueryRowContext(ctx, "SELECT id FROM users WHERE username = $1", username).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", NewNotFoundError("user", username)
		}
		return "", fmt.Errorf("failed to look up user %s: %w", username, err)
	}
	return id, nil
}

// DeletionResponse is the standard response for deletion operations.
type DeletionResponse struct {
	ID string `json:"id"`
}

// Store defines all database operations used by HTTP handlers.
type Store interface {
	BeginTx(ctx context.Context) (Tx, error)

	// Quick Launches
	GetQuickLaunch(ctx context.Context, tx Tx, id, user string) (*QuickLaunch, error)
	GetAllQuickLaunches(ctx context.Context, tx Tx, user string) ([]QuickLaunch, error)
	GetQuickLaunchesByApp(ctx context.Context, tx Tx, appID, user string) ([]QuickLaunch, error)
	AddQuickLaunch(ctx context.Context, tx Tx, user string, nql *NewQuickLaunch) (*QuickLaunch, error)
	UpdateQuickLaunch(ctx context.Context, tx Tx, id, user string, uql *UpdateQuickLaunchRequest) (*QuickLaunch, error)
	DeleteQuickLaunch(ctx context.Context, tx Tx, id, user string) error
	MergeSubmission(ctx context.Context, tx Tx, qlID, user string, newSubmission json.RawMessage) (json.RawMessage, error)

	// Favorites
	GetAllFavorites(ctx context.Context, tx Tx, user string) ([]QuickLaunchFavorite, error)
	GetFavorite(ctx context.Context, tx Tx, user, favID string) (*QuickLaunchFavorite, error)
	AddFavorite(ctx context.Context, tx Tx, user, quickLaunchID string) (*QuickLaunchFavorite, error)
	DeleteFavorite(ctx context.Context, tx Tx, user, favID string) error

	// User Defaults
	GetUserDefault(ctx context.Context, tx Tx, user, id string) (*QuickLaunchUserDefault, error)
	GetAllUserDefaults(ctx context.Context, tx Tx, user string) ([]QuickLaunchUserDefault, error)
	AddUserDefault(ctx context.Context, tx Tx, user string, nud *NewQuickLaunchUserDefault) (*QuickLaunchUserDefault, error)
	UpdateUserDefault(ctx context.Context, tx Tx, id, user string, update *UpdateQuickLaunchUserDefaultRequest) (*QuickLaunchUserDefault, error)
	DeleteUserDefault(ctx context.Context, tx Tx, user, id string) error

	// Global Defaults
	GetGlobalDefault(ctx context.Context, tx Tx, user, id string) (*QuickLaunchGlobalDefault, error)
	GetAllGlobalDefaults(ctx context.Context, tx Tx, user string) ([]QuickLaunchGlobalDefault, error)
	AddGlobalDefault(ctx context.Context, tx Tx, user string, ngd *NewQuickLaunchGlobalDefault) (*QuickLaunchGlobalDefault, error)
	UpdateGlobalDefault(ctx context.Context, tx Tx, id, user string, update *UpdateQuickLaunchGlobalDefaultRequest) (*QuickLaunchGlobalDefault, error)
	DeleteGlobalDefault(ctx context.Context, tx Tx, user, id string) error

	// Settings
	ListConcurrentJobLimits(ctx context.Context, tx Tx) ([]ConcurrentJobLimit, error)
	GetConcurrentJobLimit(ctx context.Context, tx Tx, username string) (*ConcurrentJobLimit, error)
	SetConcurrentJobLimit(ctx context.Context, tx Tx, username string, limit int) (*ConcurrentJobLimit, error)
	RemoveConcurrentJobLimit(ctx context.Context, tx Tx, username string) (*ConcurrentJobLimit, error)
}
