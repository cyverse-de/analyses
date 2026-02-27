package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// QuickLaunchFavorite represents a favorited quick launch.
type QuickLaunchFavorite struct {
	ID            string `db:"id" json:"id"`
	QuickLaunchID string `db:"quick_launch_id" json:"quick_launch_id"`
	User          string `db:"user" json:"user"`
}

// GetAllFavorites returns all quick launch favorites for a user.
func (d *Database) GetAllFavorites(ctx context.Context, tx Tx, user string) ([]QuickLaunchFavorite, error) {
	userID, err := d.GetUserID(ctx, tx, user)
	if err != nil {
		return nil, err
	}

	var favs []QuickLaunchFavorite
	err = tx.SelectContext(ctx, &favs,
		`SELECT qlf.id,
		        qlf.quick_launch_id,
		        u.username AS "user"
		   FROM quick_launch_favorites qlf
		   JOIN quick_launches ql ON qlf.quick_launch_id = ql.id
		   JOIN users u ON qlf.user_id = u.id
		  WHERE qlf.user_id = $1`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	if favs == nil {
		favs = []QuickLaunchFavorite{}
	}
	return favs, nil
}

// GetFavorite returns a single quick launch favorite.
func (d *Database) GetFavorite(ctx context.Context, tx Tx, user, favID string) (*QuickLaunchFavorite, error) {
	userID, err := d.GetUserID(ctx, tx, user)
	if err != nil {
		return nil, err
	}

	var fav QuickLaunchFavorite
	err = tx.QueryRowxContext(ctx,
		`SELECT qlf.id,
		        qlf.quick_launch_id,
		        u.username AS "user"
		   FROM quick_launch_favorites qlf
		   JOIN quick_launches ql ON qlf.quick_launch_id = ql.id
		   JOIN users u ON qlf.user_id = u.id
		  WHERE qlf.user_id = $1
		    AND qlf.id = $2`,
		userID, favID,
	).StructScan(&fav)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, NewNotFoundError("favorite", favID)
		}
		return nil, fmt.Errorf("failed to get favorite %s: %w", favID, err)
	}
	return &fav, nil
}

// AddFavorite adds a quick launch favorite.
func (d *Database) AddFavorite(ctx context.Context, tx Tx, user, quickLaunchID string) (*QuickLaunchFavorite, error) {
	userID, err := d.GetUserID(ctx, tx, user)
	if err != nil {
		return nil, err
	}

	id := uuid.New().String()
	_, err = tx.ExecContext(ctx,
		"INSERT INTO quick_launch_favorites (id, quick_launch_id, user_id) VALUES ($1, $2, $3)",
		id, quickLaunchID, userID,
	)
	if err != nil {
		return nil, err
	}

	return d.GetFavorite(ctx, tx, user, id)
}

// DeleteFavorite deletes a quick launch favorite. Idempotent: returns nil if
// the favorite does not exist.
func (d *Database) DeleteFavorite(ctx context.Context, tx Tx, user, favID string) error {
	userID, err := d.GetUserID(ctx, tx, user)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx,
		"DELETE FROM quick_launch_favorites WHERE id = $1 AND user_id = $2",
		favID, userID,
	)
	return err
}
