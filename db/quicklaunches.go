package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"maps"

	"github.com/cyverse-de/analyses/common"
	"github.com/google/uuid"
)

// QuickLaunch represents a quick launch record.
type QuickLaunch struct {
	ID           string          `db:"id" json:"id"`
	Creator      string          `db:"creator" json:"creator"`
	AppID        string          `db:"app_id" json:"app_id"`
	AppVersionID string          `db:"app_version_id" json:"app_version_id"`
	Name         string          `db:"name" json:"name"`
	Description  string          `db:"description" json:"description"`
	IsPublic     bool            `db:"is_public" json:"is_public"`
	Submission   json.RawMessage `db:"submission" json:"submission" swaggertype:"object"`
}

// NewQuickLaunch represents the data needed to create a new quick launch.
type NewQuickLaunch struct {
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	AppID        string          `json:"app_id"`
	AppVersionID string          `json:"app_version_id"`
	IsPublic     bool            `json:"is_public"`
	Submission   json.RawMessage `json:"submission" swaggertype:"object"`
}

// UpdateQuickLaunchRequest represents the data for updating a quick launch.
type UpdateQuickLaunchRequest struct {
	Name         *string          `json:"name,omitempty"`
	Description  *string          `json:"description,omitempty"`
	AppID        *string          `json:"app_id,omitempty"`
	AppVersionID *string          `json:"app_version_id,omitempty"`
	IsPublic     *bool            `json:"is_public,omitempty"`
	Creator      *string          `json:"creator,omitempty"`
	Submission   *json.RawMessage `json:"submission,omitempty" swaggertype:"object"`
}

const quickLaunchSelectSQL = `
	SELECT ql.id,
	       u.username AS creator,
	       ql.app_id,
	       ql.app_version_id,
	       ql.name,
	       ql.description,
	       ql.is_public,
	       s.submission
	  FROM quick_launches ql
	  JOIN users u ON ql.creator = u.id
	  JOIN submissions s ON ql.submission_id = s.id`

// GetQuickLaunch returns a quick launch by ID, scoped to the user.
func (d *Database) GetQuickLaunch(ctx context.Context, tx Tx, id, user string) (*QuickLaunch, error) {
	userID, err := d.GetUserID(ctx, tx, user)
	if err != nil {
		return nil, err
	}

	var ql QuickLaunch
	err = tx.QueryRowxContext(ctx,
		quickLaunchSelectSQL+`
		 WHERE ql.id = $1
		   AND (ql.creator = $2 OR ql.is_public = true)`,
		id, userID,
	).StructScan(&ql)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, NewNotFoundError("quick launch", id)
		}
		return nil, fmt.Errorf("failed to get quick launch %s: %w", id, err)
	}
	return &ql, nil
}

// GetAllQuickLaunches returns all quick launches accessible to the user.
func (d *Database) GetAllQuickLaunches(ctx context.Context, tx Tx, user string) ([]QuickLaunch, error) {
	userID, err := d.GetUserID(ctx, tx, user)
	if err != nil {
		return nil, err
	}

	var qls []QuickLaunch
	err = tx.SelectContext(ctx, &qls,
		quickLaunchSelectSQL+`
		 WHERE ql.creator = $1 OR ql.is_public = true`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	if qls == nil {
		qls = []QuickLaunch{}
	}
	return qls, nil
}

// GetQuickLaunchesByApp returns quick launches for an app, scoped to the user.
func (d *Database) GetQuickLaunchesByApp(ctx context.Context, tx Tx, appID, user string) ([]QuickLaunch, error) {
	userID, err := d.GetUserID(ctx, tx, user)
	if err != nil {
		return nil, err
	}

	var qls []QuickLaunch
	err = tx.SelectContext(ctx, &qls,
		quickLaunchSelectSQL+`
		 WHERE (ql.creator = $1 OR ql.is_public = true)
		   AND ql.app_id = $2`,
		userID, appID,
	)
	if err != nil {
		return nil, err
	}
	if qls == nil {
		qls = []QuickLaunch{}
	}
	return qls, nil
}

// AddQuickLaunch adds a new quick launch.
func (d *Database) AddQuickLaunch(ctx context.Context, tx Tx, user string, nql *NewQuickLaunch) (*QuickLaunch, error) {
	userID, err := d.GetUserID(ctx, tx, user)
	if err != nil {
		return nil, err
	}

	submissionID, err := d.AddSubmission(ctx, tx, nql.Submission)
	if err != nil {
		return nil, err
	}

	id := uuid.New().String()
	_, err = tx.ExecContext(ctx,
		`INSERT INTO quick_launches (id, name, description, app_id, app_version_id, is_public, submission_id, creator)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		id, nql.Name, nql.Description, nql.AppID, nql.AppVersionID, nql.IsPublic, submissionID, userID,
	)
	if err != nil {
		return nil, err
	}

	return d.GetQuickLaunch(ctx, tx, id, user)
}

// UnjoinedQuickLaunch is a raw quick_launches row without joins.
type UnjoinedQuickLaunch struct {
	ID           string `db:"id"`
	Name         string `db:"name"`
	Description  string `db:"description"`
	AppID        string `db:"app_id"`
	AppVersionID string `db:"app_version_id"`
	IsPublic     bool   `db:"is_public"`
	SubmissionID string `db:"submission_id"`
	Creator      string `db:"creator"`
}

// GetUnjoinedQuickLaunch returns a raw quick launch row owned by user.
func (d *Database) GetUnjoinedQuickLaunch(ctx context.Context, tx Tx, id, user string) (*UnjoinedQuickLaunch, error) {
	userID, err := d.GetUserID(ctx, tx, user)
	if err != nil {
		return nil, err
	}

	var uql UnjoinedQuickLaunch
	err = tx.QueryRowxContext(ctx,
		`SELECT id, name, description, app_id, app_version_id, is_public, submission_id, creator
		   FROM quick_launches WHERE id = $1 AND creator = $2`,
		id, userID,
	).StructScan(&uql)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, NewNotFoundError("quick launch", id)
		}
		return nil, fmt.Errorf("failed to get quick launch %s: %w", id, err)
	}
	return &uql, nil
}

// MergeSubmission merges new submission data into the existing one.
func (d *Database) MergeSubmission(ctx context.Context, tx Tx, qlID, user string, newSubmission json.RawMessage) (json.RawMessage, error) {
	uql, err := d.GetUnjoinedQuickLaunch(ctx, tx, qlID, user)
	if err != nil {
		return nil, err
	}

	oldSub, err := d.GetSubmission(ctx, tx, uql.SubmissionID)
	if err != nil {
		return nil, err
	}

	var oldMap map[string]any
	if err := json.Unmarshal(oldSub.Submission, &oldMap); err != nil {
		return nil, err
	}

	var newMap map[string]any
	if err := json.Unmarshal(newSubmission, &newMap); err != nil {
		return nil, err
	}

	// Shallow copy is intentional: submission values are JSON primitives,
	// arrays, and nested objects that we treat as opaque. A deep copy is
	// unnecessary because we re-marshal immediately and discard the maps.
	maps.Copy(oldMap, newMap)

	merged, err := json.Marshal(oldMap)
	if err != nil {
		return nil, err
	}
	return merged, nil
}

// UpdateQuickLaunch updates an existing quick launch.
func (d *Database) UpdateQuickLaunch(ctx context.Context, tx Tx, id, user string, uql *UpdateQuickLaunchRequest) (*QuickLaunch, error) {
	userID, err := d.GetUserID(ctx, tx, user)
	if err != nil {
		return nil, err
	}

	// Verify the quick launch exists and is owned by user.
	var exists bool
	err = tx.QueryRowContext(ctx,
		"SELECT EXISTS(SELECT 1 FROM quick_launches WHERE id = $1 AND creator = $2)",
		id, userID,
	).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check quick launch existence: %w", err)
	}
	if !exists {
		return nil, NewNotFoundError("quick launch", id)
	}

	// Handle submission merging.
	var newSubmissionID string
	var oldSubmissionID string
	if uql.Submission != nil {
		unjoined, uerr := d.GetUnjoinedQuickLaunch(ctx, tx, id, user)
		if uerr != nil {
			return nil, uerr
		}
		oldSubmissionID = unjoined.SubmissionID

		merged, merr := d.MergeSubmission(ctx, tx, id, user, *uql.Submission)
		if merr != nil {
			return nil, merr
		}
		newSubmissionID, err = d.AddSubmission(ctx, tx, merged)
		if err != nil {
			return nil, err
		}
	} else {
		unjoined, uerr := d.GetUnjoinedQuickLaunch(ctx, tx, id, user)
		if uerr != nil {
			return nil, uerr
		}
		newSubmissionID = unjoined.SubmissionID
	}

	// Resolve the target creator.
	targetCreator := userID
	if uql.Creator != nil {
		targetCreator, err = d.GetUserID(ctx, tx, *uql.Creator)
		if err != nil {
			return nil, err
		}
	}

	// Build dynamic UPDATE using positional args.
	query := "UPDATE quick_launches SET submission_id = $1, creator = $2"
	args := []any{newSubmissionID, targetCreator}
	argIdx := 3

	if uql.Name != nil {
		query += fmt.Sprintf(", name = $%d", argIdx)
		args = append(args, *uql.Name)
		argIdx++
	}
	if uql.Description != nil {
		query += fmt.Sprintf(", description = $%d", argIdx)
		args = append(args, *uql.Description)
		argIdx++
	}
	if uql.AppID != nil {
		query += fmt.Sprintf(", app_id = $%d", argIdx)
		args = append(args, *uql.AppID)
		argIdx++
	}
	if uql.AppVersionID != nil {
		query += fmt.Sprintf(", app_version_id = $%d", argIdx)
		args = append(args, *uql.AppVersionID)
		argIdx++
	}
	if uql.IsPublic != nil {
		query += fmt.Sprintf(", is_public = $%d", argIdx)
		args = append(args, *uql.IsPublic)
		argIdx++
	}

	query += fmt.Sprintf(" WHERE id = $%d AND creator = $%d", argIdx, argIdx+1)
	args = append(args, id, userID)

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	// Clean up the old submission row if it was replaced.
	if oldSubmissionID != "" && oldSubmissionID != newSubmissionID {
		if err := d.DeleteSubmission(ctx, tx, oldSubmissionID); err != nil {
			common.Log.Warnf("failed to delete old submission %s during quick launch update: %v", oldSubmissionID, err)
		}
	}

	return d.GetQuickLaunch(ctx, tx, id, user)
}

// DeleteQuickLaunch deletes a quick launch and its associated submission by ID.
func (d *Database) DeleteQuickLaunch(ctx context.Context, tx Tx, id, user string) error {
	userID, err := d.GetUserID(ctx, tx, user)
	if err != nil {
		return err
	}

	// Look up the submission_id before deleting so we can clean it up.
	var submissionID string
	err = tx.QueryRowContext(ctx,
		"SELECT submission_id FROM quick_launches WHERE id = $1 AND creator = $2",
		id, userID,
	).Scan(&submissionID)
	if errors.Is(err, sql.ErrNoRows) {
		// Nothing to delete; treat as success per the idempotent delete contract.
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to look up quick launch %s: %w", id, err)
	}

	_, err = tx.ExecContext(ctx,
		"DELETE FROM quick_launches WHERE id = $1 AND creator = $2",
		id, userID,
	)
	if err != nil {
		return err
	}

	// Clean up the orphaned submission row.
	return d.DeleteSubmission(ctx, tx, submissionID)
}
