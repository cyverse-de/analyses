package db

import (
	"context"
	"database/sql"
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
