package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// QuickLaunchUserDefault represents a quick launch user default.
type QuickLaunchUserDefault struct {
	ID            string `db:"id" json:"id"`
	User          string `db:"user" json:"user"`
	QuickLaunchID string `db:"quick_launch_id" json:"quick_launch_id"`
	AppID         string `db:"app_id" json:"app_id"`
}

// NewQuickLaunchUserDefault represents the data to create a user default.
type NewQuickLaunchUserDefault struct {
	QuickLaunchID string `json:"quick_launch_id"`
	AppID         string `json:"app_id"`
}

// UpdateQuickLaunchUserDefaultRequest represents an update to a user default.
type UpdateQuickLaunchUserDefaultRequest struct {
	QuickLaunchID string `json:"quick_launch_id"`
	AppID         string `json:"app_id"`
}

// GetUserDefault returns a quick launch user default.
func (d *Database) GetUserDefault(ctx context.Context, tx Tx, user, id string) (*QuickLaunchUserDefault, error) {
	userID, err := d.GetUserID(ctx, tx, user)
	if err != nil {
		return nil, err
	}

	var ud QuickLaunchUserDefault
	err = tx.QueryRowxContext(ctx,
		`SELECT qlud.id,
		        u.username AS "user",
		        qlud.quick_launch_id,
		        qlud.app_id
		   FROM quick_launch_user_defaults qlud
		   JOIN users u ON qlud.user_id = u.id
		  WHERE qlud.user_id = $1
		    AND qlud.id = $2`,
		userID, id,
	).StructScan(&ud)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, NewNotFoundError("user default", id)
		}
		return nil, fmt.Errorf("failed to get user default %s: %w", id, err)
	}
	return &ud, nil
}

// GetAllUserDefaults returns all quick launch user defaults for a user.
func (d *Database) GetAllUserDefaults(ctx context.Context, tx Tx, user string) ([]QuickLaunchUserDefault, error) {
	userID, err := d.GetUserID(ctx, tx, user)
	if err != nil {
		return nil, err
	}

	var uds []QuickLaunchUserDefault
	err = tx.SelectContext(ctx, &uds,
		`SELECT qlud.id,
		        u.username AS "user",
		        qlud.quick_launch_id,
		        qlud.app_id
		   FROM quick_launch_user_defaults qlud
		   JOIN users u ON qlud.user_id = u.id
		  WHERE qlud.user_id = $1`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	if uds == nil {
		uds = []QuickLaunchUserDefault{}
	}
	return uds, nil
}

// AddUserDefault adds a quick launch user default.
func (d *Database) AddUserDefault(ctx context.Context, tx Tx, user string, nud *NewQuickLaunchUserDefault) (*QuickLaunchUserDefault, error) {
	userID, err := d.GetUserID(ctx, tx, user)
	if err != nil {
		return nil, err
	}

	id := uuid.New().String()
	_, err = tx.ExecContext(ctx,
		`INSERT INTO quick_launch_user_defaults (id, user_id, quick_launch_id, app_id)
		 VALUES ($1, $2, $3, $4)`,
		id, userID, nud.QuickLaunchID, nud.AppID,
	)
	if err != nil {
		return nil, err
	}

	return d.GetUserDefault(ctx, tx, user, id)
}

// UpdateUserDefault updates a quick launch user default.
func (d *Database) UpdateUserDefault(ctx context.Context, tx Tx, id, user string, update *UpdateQuickLaunchUserDefaultRequest) (*QuickLaunchUserDefault, error) {
	userID, err := d.GetUserID(ctx, tx, user)
	if err != nil {
		return nil, err
	}

	result, err := tx.ExecContext(ctx,
		`UPDATE quick_launch_user_defaults
		    SET app_id = $1, quick_launch_id = $2
		  WHERE id = $3 AND user_id = $4`,
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
		return nil, NewNotFoundError("user default", id)
	}

	return d.GetUserDefault(ctx, tx, user, id)
}

// DeleteUserDefault deletes a quick launch user default. Idempotent: returns
// nil if the user default does not exist.
func (d *Database) DeleteUserDefault(ctx context.Context, tx Tx, user, id string) error {
	userID, err := d.GetUserID(ctx, tx, user)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx,
		"DELETE FROM quick_launch_user_defaults WHERE id = $1 AND user_id = $2",
		id, userID,
	)
	return err
}
