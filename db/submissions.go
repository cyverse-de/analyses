package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// Submission represents a submission record in the database.
type Submission struct {
	ID         string          `db:"id" json:"id"`
	Submission json.RawMessage `db:"submission" json:"submission"`
}

// AddSubmission adds a new submission and returns its UUID.
func (d *Database) AddSubmission(ctx context.Context, tx Tx, submission json.RawMessage) (string, error) {
	id := uuid.New().String()
	_, err := tx.ExecContext(ctx,
		"INSERT INTO submissions (id, submission) VALUES ($1, CAST($2 AS JSON))",
		id, string(submission),
	)
	if err != nil {
		return "", err
	}
	return id, nil
}

// GetSubmission returns a submission by ID.
func (d *Database) GetSubmission(ctx context.Context, tx Tx, id string) (*Submission, error) {
	var sub Submission
	err := tx.QueryRowxContext(ctx,
		"SELECT id, submission FROM submissions WHERE id = $1", id,
	).StructScan(&sub)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, NewNotFoundError("submission", id)
		}
		return nil, fmt.Errorf("failed to get submission %s: %w", id, err)
	}
	return &sub, nil
}

// UpdateSubmission updates a submission and returns the updated record.
func (d *Database) UpdateSubmission(ctx context.Context, tx Tx, id string, submission json.RawMessage) (*Submission, error) {
	_, err := tx.ExecContext(ctx,
		"UPDATE submissions SET submission = CAST($1 AS JSON) WHERE id = $2",
		string(submission), id,
	)
	if err != nil {
		return nil, err
	}
	return d.GetSubmission(ctx, tx, id)
}

// DeleteSubmission deletes a submission by ID. Idempotent: returns nil if
// the submission does not exist.
func (d *Database) DeleteSubmission(ctx context.Context, tx Tx, id string) error {
	_, err := tx.ExecContext(ctx, "DELETE FROM submissions WHERE id = $1", id)
	return err
}
