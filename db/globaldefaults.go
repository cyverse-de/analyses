package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// QuickLaunchGlobalDefault represents a quick launch global default.
type QuickLaunchGlobalDefault struct {
	ID            string `db:"id" json:"id"`
	AppID         string `db:"app_id" json:"app_id"`
	QuickLaunchID string `db:"quick_launch_id" json:"quick_launch_id"`
}

// NewQuickLaunchGlobalDefault represents the data to create a global default.
type NewQuickLaunchGlobalDefault struct {
	AppID         string `json:"app_id"`
	QuickLaunchID string `json:"quick_launch_id"`
}

// UpdateQuickLaunchGlobalDefaultRequest represents an update to a global default.
type UpdateQuickLaunchGlobalDefaultRequest struct {
	AppID         string `json:"app_id"`
	QuickLaunchID string `json:"quick_launch_id"`
}

// GetGlobalDefault returns a quick launch global default.
func (d *Database) GetGlobalDefault(ctx context.Context, tx Tx, user, id string) (*QuickLaunchGlobalDefault, error) {
	userID, err := d.GetUserID(ctx, tx, user)
	if err != nil {
		return nil, err
	}

	var gd QuickLaunchGlobalDefault
	err = tx.QueryRowxContext(ctx,
		`SELECT gd.id, gd.app_id, gd.quick_launch_id
		   FROM quick_launch_global_defaults gd
		   JOIN quick_launches ql ON gd.quick_launch_id = ql.id
		  WHERE gd.id = $1
		    AND ql.creator = $2`,
		id, userID,
	).StructScan(&gd)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, NewNotFoundError("global default", id)
		}
		return nil, fmt.Errorf("failed to get global default %s: %w", id, err)
	}
	return &gd, nil
}

// GetAllGlobalDefaults returns all quick launch global defaults for a user.
func (d *Database) GetAllGlobalDefaults(ctx context.Context, tx Tx, user string) ([]QuickLaunchGlobalDefault, error) {
	userID, err := d.GetUserID(ctx, tx, user)
	if err != nil {
		return nil, err
	}

	var gds []QuickLaunchGlobalDefault
	err = tx.SelectContext(ctx, &gds,
		`SELECT gd.id, gd.app_id, gd.quick_launch_id
		   FROM quick_launch_global_defaults gd
		   JOIN quick_launches ql ON gd.quick_launch_id = ql.id
		  WHERE ql.creator = $1`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	if gds == nil {
		gds = []QuickLaunchGlobalDefault{}
	}
	return gds, nil
}

// AddGlobalDefault adds a quick launch global default.
func (d *Database) AddGlobalDefault(ctx context.Context, tx Tx, user string, ngd *NewQuickLaunchGlobalDefault) (*QuickLaunchGlobalDefault, error) {
	id := uuid.New().String()
	_, err := tx.ExecContext(ctx,
		`INSERT INTO quick_launch_global_defaults (id, app_id, quick_launch_id)
		 VALUES ($1, $2, $3)`,
		id, ngd.AppID, ngd.QuickLaunchID,
	)
	if err != nil {
		return nil, err
	}

	return d.GetGlobalDefault(ctx, tx, user, id)
}

// UpdateGlobalDefault updates a quick launch global default.
func (d *Database) UpdateGlobalDefault(ctx context.Context, tx Tx, id, user string, update *UpdateQuickLaunchGlobalDefaultRequest) (*QuickLaunchGlobalDefault, error) {
	userID, err := d.GetUserID(ctx, tx, user)
	if err != nil {
		return nil, err
	}

	result, err := tx.ExecContext(ctx,
		`UPDATE quick_launch_global_defaults
		    SET app_id = $1, quick_launch_id = $2
		  WHERE id = $3
		    AND quick_launch_id IN (
		        SELECT ql.id FROM quick_launches ql
		        JOIN quick_launch_global_defaults gd ON ql.id = gd.quick_launch_id
		        WHERE ql.creator = $4
		    )`,
		update.AppID, update.QuickLaunchID, id, userID,
	)
	if err != nil {
		return nil, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rows == 0 {
		return nil, NewNotFoundError("global default", id)
	}

	return d.GetGlobalDefault(ctx, tx, user, id)
}

// DeleteGlobalDefault deletes a quick launch global default. Idempotent:
// returns nil if the global default does not exist.
func (d *Database) DeleteGlobalDefault(ctx context.Context, tx Tx, user, id string) error {
	userID, err := d.GetUserID(ctx, tx, user)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx,
		`DELETE FROM quick_launch_global_defaults
		  WHERE id = $1
		    AND quick_launch_id IN (
		        SELECT ql.id FROM quick_launches ql
		        JOIN quick_launch_global_defaults gd ON ql.id = gd.quick_launch_id
		        WHERE ql.creator = $2
		    )`,
		id, userID,
	)
	return err
}
