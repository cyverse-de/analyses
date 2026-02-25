package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"maps"

	"github.com/google/uuid"
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
type Tx interface {
	Exec(query string, args ...any) (sql.Result, error)
	QueryRow(query string, args ...any) *sql.Row
	QueryRowx(query string, args ...any) *sqlx.Row
	Select(dest any, query string, args ...any) error
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
func (d *Database) BeginTx() (Tx, error) {
	return d.db.Beginx()
}

// GetUserID returns the user ID for the given username.
func (d *Database) GetUserID(tx Tx, username string) (string, error) {
	var id string
	err := tx.QueryRow("SELECT id FROM users WHERE username = $1", username).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", NewNotFoundError("user", username)
		}
		return "", fmt.Errorf("failed to look up user %s: %w", username, err)
	}
	return id, nil
}

// Submission represents a submission record in the database.
type Submission struct {
	ID         string          `db:"id" json:"id"`
	Submission json.RawMessage `db:"submission" json:"submission"`
}

// AddSubmission adds a new submission and returns its UUID.
func (d *Database) AddSubmission(tx Tx, submission json.RawMessage) (string, error) {
	id := uuid.New().String()
	_, err := tx.Exec(
		"INSERT INTO submissions (id, submission) VALUES ($1, CAST($2 AS JSON))",
		id, string(submission),
	)
	if err != nil {
		return "", err
	}
	return id, nil
}

// GetSubmission returns a submission by ID.
func (d *Database) GetSubmission(tx Tx, id string) (*Submission, error) {
	var sub Submission
	err := tx.QueryRowx(
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
func (d *Database) UpdateSubmission(tx Tx, id string, submission json.RawMessage) (*Submission, error) {
	_, err := tx.Exec(
		"UPDATE submissions SET submission = CAST($1 AS JSON) WHERE id = $2",
		string(submission), id,
	)
	if err != nil {
		return nil, err
	}
	return d.GetSubmission(tx, id)
}

// DeleteSubmission deletes a submission by ID.
func (d *Database) DeleteSubmission(tx Tx, id string) error {
	result, err := tx.Exec("DELETE FROM submissions WHERE id = $1", id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return NewNotFoundError("submission", id)
	}
	return nil
}

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
func (d *Database) GetQuickLaunch(tx Tx, id, user string) (*QuickLaunch, error) {
	userID, err := d.GetUserID(tx, user)
	if err != nil {
		return nil, err
	}

	var ql QuickLaunch
	err = tx.QueryRowx(
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
func (d *Database) GetAllQuickLaunches(tx Tx, user string) ([]QuickLaunch, error) {
	userID, err := d.GetUserID(tx, user)
	if err != nil {
		return nil, err
	}

	var qls []QuickLaunch
	err = tx.Select(&qls,
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
func (d *Database) GetQuickLaunchesByApp(tx Tx, appID, user string) ([]QuickLaunch, error) {
	userID, err := d.GetUserID(tx, user)
	if err != nil {
		return nil, err
	}

	var qls []QuickLaunch
	err = tx.Select(&qls,
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
func (d *Database) AddQuickLaunch(tx Tx, user string, nql *NewQuickLaunch) (*QuickLaunch, error) {
	userID, err := d.GetUserID(tx, user)
	if err != nil {
		return nil, err
	}

	submissionID, err := d.AddSubmission(tx, nql.Submission)
	if err != nil {
		return nil, err
	}

	id := uuid.New().String()
	_, err = tx.Exec(
		`INSERT INTO quick_launches (id, name, description, app_id, app_version_id, is_public, submission_id, creator)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		id, nql.Name, nql.Description, nql.AppID, nql.AppVersionID, nql.IsPublic, submissionID, userID,
	)
	if err != nil {
		return nil, err
	}

	return d.GetQuickLaunch(tx, id, user)
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
func (d *Database) GetUnjoinedQuickLaunch(tx Tx, id, user string) (*UnjoinedQuickLaunch, error) {
	userID, err := d.GetUserID(tx, user)
	if err != nil {
		return nil, err
	}

	var uql UnjoinedQuickLaunch
	err = tx.QueryRowx(
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
func (d *Database) MergeSubmission(tx Tx, qlID, user string, newSubmission json.RawMessage) (json.RawMessage, error) {
	uql, err := d.GetUnjoinedQuickLaunch(tx, qlID, user)
	if err != nil {
		return nil, err
	}

	oldSub, err := d.GetSubmission(tx, uql.SubmissionID)
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

	maps.Copy(oldMap, newMap)

	merged, err := json.Marshal(oldMap)
	if err != nil {
		return nil, err
	}
	return merged, nil
}

// UpdateQuickLaunch updates an existing quick launch.
func (d *Database) UpdateQuickLaunch(tx Tx, id, user string, uql *UpdateQuickLaunchRequest) (*QuickLaunch, error) {
	userID, err := d.GetUserID(tx, user)
	if err != nil {
		return nil, err
	}

	// Verify the quick launch exists and is owned by user
	var exists bool
	err = tx.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM quick_launches WHERE id = $1 AND creator = $2)",
		id, userID,
	).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check quick launch existence: %w", err)
	}
	if !exists {
		return nil, NewNotFoundError("quick launch", id)
	}

	// Handle submission merging
	var newSubmissionID string
	var oldSubmissionID string
	if uql.Submission != nil {
		unjoined, uerr := d.GetUnjoinedQuickLaunch(tx, id, user)
		if uerr != nil {
			return nil, uerr
		}
		oldSubmissionID = unjoined.SubmissionID

		merged, merr := d.MergeSubmission(tx, id, user, *uql.Submission)
		if merr != nil {
			return nil, merr
		}
		newSubmissionID, err = d.AddSubmission(tx, merged)
		if err != nil {
			return nil, err
		}
	} else {
		unjoined, uerr := d.GetUnjoinedQuickLaunch(tx, id, user)
		if uerr != nil {
			return nil, uerr
		}
		newSubmissionID = unjoined.SubmissionID
	}

	// Resolve the target creator
	targetCreator := userID
	if uql.Creator != nil {
		targetCreator, err = d.GetUserID(tx, *uql.Creator)
		if err != nil {
			return nil, err
		}
	}

	// Build dynamic UPDATE using positional args
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

	_, err = tx.Exec(query, args...)
	if err != nil {
		return nil, err
	}

	// Clean up the old submission row if it was replaced.
	if oldSubmissionID != "" && oldSubmissionID != newSubmissionID {
		d.DeleteSubmission(tx, oldSubmissionID) //nolint:errcheck
	}

	return d.GetQuickLaunch(tx, id, user)
}

// DeleteQuickLaunch deletes a quick launch and its associated submission by ID.
func (d *Database) DeleteQuickLaunch(tx Tx, id, user string) error {
	userID, err := d.GetUserID(tx, user)
	if err != nil {
		return err
	}

	// Look up the submission_id before deleting so we can clean it up.
	var submissionID string
	err = tx.QueryRow(
		"SELECT submission_id FROM quick_launches WHERE id = $1 AND creator = $2",
		id, userID,
	).Scan(&submissionID)
	if errors.Is(err, sql.ErrNoRows) {
		// Nothing to delete; treat as success per the idempotent delete contract.
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to look up quick launch %s: %w", id, err)
	}

	_, err = tx.Exec(
		"DELETE FROM quick_launches WHERE id = $1 AND creator = $2",
		id, userID,
	)
	if err != nil {
		return err
	}

	// Clean up the orphaned submission row.
	return d.DeleteSubmission(tx, submissionID)
}

// QuickLaunchFavorite represents a favorited quick launch.
type QuickLaunchFavorite struct {
	ID            string `db:"id" json:"id"`
	QuickLaunchID string `db:"quick_launch_id" json:"quick_launch_id"`
	User          string `db:"user" json:"user"`
}

// GetAllFavorites returns all quick launch favorites for a user.
func (d *Database) GetAllFavorites(tx Tx, user string) ([]QuickLaunchFavorite, error) {
	userID, err := d.GetUserID(tx, user)
	if err != nil {
		return nil, err
	}

	var favs []QuickLaunchFavorite
	err = tx.Select(&favs,
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
func (d *Database) GetFavorite(tx Tx, user, favID string) (*QuickLaunchFavorite, error) {
	userID, err := d.GetUserID(tx, user)
	if err != nil {
		return nil, err
	}

	var fav QuickLaunchFavorite
	err = tx.QueryRowx(
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
func (d *Database) AddFavorite(tx Tx, user, quickLaunchID string) (*QuickLaunchFavorite, error) {
	userID, err := d.GetUserID(tx, user)
	if err != nil {
		return nil, err
	}

	id := uuid.New().String()
	_, err = tx.Exec(
		"INSERT INTO quick_launch_favorites (id, quick_launch_id, user_id) VALUES ($1, $2, $3)",
		id, quickLaunchID, userID,
	)
	if err != nil {
		return nil, err
	}

	return d.GetFavorite(tx, user, id)
}

// DeleteFavorite deletes a quick launch favorite.
func (d *Database) DeleteFavorite(tx Tx, user, favID string) error {
	userID, err := d.GetUserID(tx, user)
	if err != nil {
		return err
	}

	result, err := tx.Exec(
		"DELETE FROM quick_launch_favorites WHERE id = $1 AND user_id = $2",
		favID, userID,
	)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return NewNotFoundError("favorite", favID)
	}
	return nil
}

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
func (d *Database) GetUserDefault(tx Tx, user, id string) (*QuickLaunchUserDefault, error) {
	userID, err := d.GetUserID(tx, user)
	if err != nil {
		return nil, err
	}

	var ud QuickLaunchUserDefault
	err = tx.QueryRowx(
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
func (d *Database) GetAllUserDefaults(tx Tx, user string) ([]QuickLaunchUserDefault, error) {
	userID, err := d.GetUserID(tx, user)
	if err != nil {
		return nil, err
	}

	var uds []QuickLaunchUserDefault
	err = tx.Select(&uds,
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
func (d *Database) AddUserDefault(tx Tx, user string, nud *NewQuickLaunchUserDefault) (*QuickLaunchUserDefault, error) {
	userID, err := d.GetUserID(tx, user)
	if err != nil {
		return nil, err
	}

	id := uuid.New().String()
	_, err = tx.Exec(
		`INSERT INTO quick_launch_user_defaults (id, user_id, quick_launch_id, app_id)
		 VALUES ($1, $2, $3, $4)`,
		id, userID, nud.QuickLaunchID, nud.AppID,
	)
	if err != nil {
		return nil, err
	}

	return d.GetUserDefault(tx, user, id)
}

// UpdateUserDefault updates a quick launch user default.
func (d *Database) UpdateUserDefault(tx Tx, id, user string, update *UpdateQuickLaunchUserDefaultRequest) (*QuickLaunchUserDefault, error) {
	userID, err := d.GetUserID(tx, user)
	if err != nil {
		return nil, err
	}

	result, err := tx.Exec(
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

	return d.GetUserDefault(tx, user, id)
}

// DeleteUserDefault deletes a quick launch user default.
func (d *Database) DeleteUserDefault(tx Tx, user, id string) error {
	userID, err := d.GetUserID(tx, user)
	if err != nil {
		return err
	}

	result, err := tx.Exec(
		"DELETE FROM quick_launch_user_defaults WHERE id = $1 AND user_id = $2",
		id, userID,
	)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return NewNotFoundError("user default", id)
	}
	return nil
}

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
func (d *Database) GetGlobalDefault(tx Tx, user, id string) (*QuickLaunchGlobalDefault, error) {
	userID, err := d.GetUserID(tx, user)
	if err != nil {
		return nil, err
	}

	var gd QuickLaunchGlobalDefault
	err = tx.QueryRowx(
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
func (d *Database) GetAllGlobalDefaults(tx Tx, user string) ([]QuickLaunchGlobalDefault, error) {
	userID, err := d.GetUserID(tx, user)
	if err != nil {
		return nil, err
	}

	var gds []QuickLaunchGlobalDefault
	err = tx.Select(&gds,
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
func (d *Database) AddGlobalDefault(tx Tx, user string, ngd *NewQuickLaunchGlobalDefault) (*QuickLaunchGlobalDefault, error) {
	id := uuid.New().String()
	_, err := tx.Exec(
		`INSERT INTO quick_launch_global_defaults (id, app_id, quick_launch_id)
		 VALUES ($1, $2, $3)`,
		id, ngd.AppID, ngd.QuickLaunchID,
	)
	if err != nil {
		return nil, err
	}

	return d.GetGlobalDefault(tx, user, id)
}

// UpdateGlobalDefault updates a quick launch global default.
func (d *Database) UpdateGlobalDefault(tx Tx, id, user string, update *UpdateQuickLaunchGlobalDefaultRequest) (*QuickLaunchGlobalDefault, error) {
	userID, err := d.GetUserID(tx, user)
	if err != nil {
		return nil, err
	}

	result, err := tx.Exec(
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

	return d.GetGlobalDefault(tx, user, id)
}

// DeleteGlobalDefault deletes a quick launch global default.
func (d *Database) DeleteGlobalDefault(tx Tx, user, id string) error {
	userID, err := d.GetUserID(tx, user)
	if err != nil {
		return err
	}

	result, err := tx.Exec(
		`DELETE FROM quick_launch_global_defaults
		  WHERE id = $1
		    AND quick_launch_id IN (
		        SELECT ql.id FROM quick_launches ql
		        JOIN quick_launch_global_defaults gd ON ql.id = gd.quick_launch_id
		        WHERE ql.creator = $2
		    )`,
		id, userID,
	)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return NewNotFoundError("global default", id)
	}
	return nil
}

// ConcurrentJobLimit represents a concurrent job limit record.
type ConcurrentJobLimit struct {
	Username       *string `db:"username" json:"username,omitempty"`
	ConcurrentJobs int     `db:"concurrent_jobs" json:"concurrent_jobs"`
	IsDefault      bool    `db:"is_default" json:"is_default,omitempty"`
}

// ConcurrentJobLimits wraps a list of limits for JSON response.
type ConcurrentJobLimits struct {
	Limits []ConcurrentJobLimit `json:"limits"`
}

// ConcurrentJobLimitUpdate represents a request to set a job limit.
type ConcurrentJobLimitUpdate struct {
	ConcurrentJobs int `json:"concurrent_jobs"`
}

const jobLimitSelectSQL = `SELECT launcher AS username,
	        concurrent_jobs,
	        (launcher IS NULL) AS is_default
	   FROM job_limits`

// ListConcurrentJobLimits returns all defined concurrent job limits.
func (d *Database) ListConcurrentJobLimits(tx Tx) ([]ConcurrentJobLimit, error) {
	var limits []ConcurrentJobLimit
	err := tx.Select(&limits, jobLimitSelectSQL+` ORDER BY launcher ASC`)
	if err != nil {
		return nil, err
	}
	if limits == nil {
		limits = []ConcurrentJobLimit{}
	}
	return limits, nil
}

// GetConcurrentJobLimit returns the concurrent job limit for a user.
// Falls back to the default limit if no user-specific limit is found.
func (d *Database) GetConcurrentJobLimit(tx Tx, username string) (*ConcurrentJobLimit, error) {
	var limits []ConcurrentJobLimit
	err := tx.Select(&limits,
		jobLimitSelectSQL+`
		  WHERE launcher = regexp_replace($1, '-', '_')
		     OR launcher IS NULL
		  ORDER BY is_default ASC`,
		username,
	)
	if err != nil {
		return nil, err
	}
	if len(limits) == 0 {
		return nil, NewNotFoundError("job limit", username)
	}
	return &limits[0], nil
}

// SetConcurrentJobLimit sets the concurrent job limit for a user.
func (d *Database) SetConcurrentJobLimit(tx Tx, username string, limit int) (*ConcurrentJobLimit, error) {
	current, err := d.GetConcurrentJobLimit(tx, username)
	if err != nil && !IsNotFound(err) {
		return nil, err
	}

	if current == nil || current.IsDefault {
		_, err = tx.Exec(
			"INSERT INTO job_limits (launcher, concurrent_jobs) VALUES (regexp_replace($1, '-', '_'), $2)",
			username, limit,
		)
	} else {
		_, err = tx.Exec(
			"UPDATE job_limits SET concurrent_jobs = $1 WHERE launcher = regexp_replace($2, '-', '_')",
			limit, username,
		)
	}
	if err != nil {
		return nil, err
	}

	return d.GetConcurrentJobLimit(tx, username)
}

// RemoveConcurrentJobLimit removes a user's concurrent job limit, returning the default.
func (d *Database) RemoveConcurrentJobLimit(tx Tx, username string) (*ConcurrentJobLimit, error) {
	_, err := tx.Exec(
		"DELETE FROM job_limits WHERE launcher = regexp_replace($1, '-', '_')",
		username,
	)
	if err != nil {
		return nil, err
	}

	// Return the default limit.
	return d.GetConcurrentJobLimit(tx, "")
}

// DeletionResponse is the standard response for deletion operations.
type DeletionResponse struct {
	ID string `json:"id"`
}
