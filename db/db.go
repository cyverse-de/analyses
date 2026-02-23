package db

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Database provides access to the analyses database.
type Database struct {
	db *sqlx.DB
}

// New returns a new Database instance.
func New(db *sqlx.DB) *Database {
	return &Database{db: db}
}

// GetUserID returns the user ID for the given username.
func (d *Database) GetUserID(username string) (string, error) {
	var id string
	err := d.db.QueryRow("SELECT id FROM users WHERE username = $1", username).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("user not found: %s", username)
	}
	return id, nil
}

// Submission represents a submission record in the database.
type Submission struct {
	ID         string          `db:"id" json:"id"`
	Submission json.RawMessage `db:"submission" json:"submission"`
}

// AddSubmission adds a new submission and returns its UUID.
func (d *Database) AddSubmission(submission json.RawMessage) (string, error) {
	id := uuid.New().String()
	_, err := d.db.Exec(
		"INSERT INTO submissions (id, submission) VALUES ($1, CAST($2 AS JSON))",
		id, string(submission),
	)
	if err != nil {
		return "", err
	}
	return id, nil
}

// GetSubmission returns a submission by ID.
func (d *Database) GetSubmission(id string) (*Submission, error) {
	var sub Submission
	err := d.db.QueryRowx(
		"SELECT id, submission FROM submissions WHERE id = $1", id,
	).StructScan(&sub)
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

// UpdateSubmission updates a submission and returns the updated record.
func (d *Database) UpdateSubmission(id string, submission json.RawMessage) (*Submission, error) {
	_, err := d.db.Exec(
		"UPDATE submissions SET submission = CAST($1 AS JSON) WHERE id = $2",
		string(submission), id,
	)
	if err != nil {
		return nil, err
	}
	return d.GetSubmission(id)
}

// DeleteSubmission deletes a submission by ID.
func (d *Database) DeleteSubmission(id string) error {
	_, err := d.db.Exec("DELETE FROM submissions WHERE id = $1", id)
	return err
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
	Submission   json.RawMessage `db:"submission" json:"submission"`
}

// NewQuickLaunch represents the data needed to create a new quick launch.
type NewQuickLaunch struct {
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	AppID        string          `json:"app_id"`
	AppVersionID string          `json:"app_version_id"`
	IsPublic     bool            `json:"is_public"`
	Submission   json.RawMessage `json:"submission"`
}

// UpdateQuickLaunchRequest represents the data for updating a quick launch.
type UpdateQuickLaunchRequest struct {
	Name         *string          `json:"name,omitempty"`
	Description  *string          `json:"description,omitempty"`
	AppID        *string          `json:"app_id,omitempty"`
	AppVersionID *string          `json:"app_version_id,omitempty"`
	IsPublic     *bool            `json:"is_public,omitempty"`
	Creator      *string          `json:"creator,omitempty"`
	Submission   *json.RawMessage `json:"submission,omitempty"`
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
func (d *Database) GetQuickLaunch(id, user string) (*QuickLaunch, error) {
	userID, err := d.GetUserID(user)
	if err != nil {
		return nil, err
	}

	var ql QuickLaunch
	err = d.db.QueryRowx(
		quickLaunchSelectSQL+`
		 WHERE ql.id = $1
		   AND (ql.creator = $2 OR ql.is_public = true)`,
		id, userID,
	).StructScan(&ql)
	if err != nil {
		return nil, fmt.Errorf("quick launch not found: %s", id)
	}
	return &ql, nil
}

// GetAllQuickLaunches returns all quick launches accessible to the user.
func (d *Database) GetAllQuickLaunches(user string) ([]QuickLaunch, error) {
	userID, err := d.GetUserID(user)
	if err != nil {
		return nil, err
	}

	var qls []QuickLaunch
	err = d.db.Select(&qls,
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
func (d *Database) GetQuickLaunchesByApp(appID, user string) ([]QuickLaunch, error) {
	userID, err := d.GetUserID(user)
	if err != nil {
		return nil, err
	}

	var qls []QuickLaunch
	err = d.db.Select(&qls,
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
func (d *Database) AddQuickLaunch(user string, nql *NewQuickLaunch) (*QuickLaunch, error) {
	userID, err := d.GetUserID(user)
	if err != nil {
		return nil, err
	}

	submissionID, err := d.AddSubmission(nql.Submission)
	if err != nil {
		return nil, err
	}

	id := uuid.New().String()
	_, err = d.db.Exec(
		`INSERT INTO quick_launches (id, name, description, app_id, app_version_id, is_public, submission_id, creator)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		id, nql.Name, nql.Description, nql.AppID, nql.AppVersionID, nql.IsPublic, submissionID, userID,
	)
	if err != nil {
		return nil, err
	}

	return d.GetQuickLaunch(id, user)
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
func (d *Database) GetUnjoinedQuickLaunch(id, user string) (*UnjoinedQuickLaunch, error) {
	userID, err := d.GetUserID(user)
	if err != nil {
		return nil, err
	}

	var uql UnjoinedQuickLaunch
	err = d.db.QueryRowx(
		"SELECT * FROM quick_launches WHERE id = $1 AND creator = $2",
		id, userID,
	).StructScan(&uql)
	if err != nil {
		return nil, err
	}
	return &uql, nil
}

// MergeSubmission merges new submission data into the existing one.
func (d *Database) MergeSubmission(qlID, user string, newSubmission json.RawMessage) (json.RawMessage, error) {
	uql, err := d.GetUnjoinedQuickLaunch(qlID, user)
	if err != nil {
		return nil, err
	}

	oldSub, err := d.GetSubmission(uql.SubmissionID)
	if err != nil {
		return nil, err
	}

	var oldMap map[string]interface{}
	if err := json.Unmarshal(oldSub.Submission, &oldMap); err != nil {
		return nil, err
	}

	var newMap map[string]interface{}
	if err := json.Unmarshal(newSubmission, &newMap); err != nil {
		return nil, err
	}

	for k, v := range newMap {
		oldMap[k] = v
	}

	merged, err := json.Marshal(oldMap)
	if err != nil {
		return nil, err
	}
	return merged, nil
}

// UpdateQuickLaunch updates an existing quick launch.
func (d *Database) UpdateQuickLaunch(id, user string, uql *UpdateQuickLaunchRequest) (*QuickLaunch, error) {
	userID, err := d.GetUserID(user)
	if err != nil {
		return nil, err
	}

	// Verify the quick launch exists and is owned by user
	var exists bool
	err = d.db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM quick_launches WHERE id = $1 AND creator = $2)",
		id, userID,
	).Scan(&exists)
	if err != nil || !exists {
		return nil, fmt.Errorf("quick launch not found: %s", id)
	}

	// Handle submission merging
	var newSubmissionID string
	if uql.Submission != nil {
		merged, merr := d.MergeSubmission(id, user, *uql.Submission)
		if merr != nil {
			return nil, merr
		}
		newSubmissionID, err = d.AddSubmission(merged)
		if err != nil {
			return nil, err
		}
	} else {
		unjoined, uerr := d.GetUnjoinedQuickLaunch(id, user)
		if uerr != nil {
			return nil, uerr
		}
		newSubmissionID = unjoined.SubmissionID
	}

	// Resolve the target creator
	targetCreator := userID
	if uql.Creator != nil {
		targetCreator, err = d.GetUserID(*uql.Creator)
		if err != nil {
			return nil, err
		}
	}

	// Build dynamic UPDATE using positional args
	query := "UPDATE quick_launches SET submission_id = $1, creator = $2"
	args := []interface{}{newSubmissionID, targetCreator}
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

	_, err = d.db.Exec(query, args...)
	if err != nil {
		return nil, err
	}

	return d.GetQuickLaunch(id, user)
}

// DeleteQuickLaunch deletes a quick launch by ID.
func (d *Database) DeleteQuickLaunch(id, user string) error {
	userID, err := d.GetUserID(user)
	if err != nil {
		return err
	}

	_, err = d.db.Exec(
		"DELETE FROM quick_launches WHERE id = $1 AND creator = $2",
		id, userID,
	)
	return err
}

// QuickLaunchFavorite represents a favorited quick launch.
type QuickLaunchFavorite struct {
	ID            string `db:"id" json:"id"`
	QuickLaunchID string `db:"quick_launch_id" json:"quick_launch_id"`
	User          string `db:"user" json:"user"`
}

// GetAllFavorites returns all quick launch favorites for a user.
func (d *Database) GetAllFavorites(user string) ([]QuickLaunchFavorite, error) {
	userID, err := d.GetUserID(user)
	if err != nil {
		return nil, err
	}

	var favs []QuickLaunchFavorite
	err = d.db.Select(&favs,
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
func (d *Database) GetFavorite(user, favID string) (*QuickLaunchFavorite, error) {
	userID, err := d.GetUserID(user)
	if err != nil {
		return nil, err
	}

	var fav QuickLaunchFavorite
	err = d.db.QueryRowx(
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
		return nil, fmt.Errorf("favorite not found: %s", favID)
	}
	return &fav, nil
}

// AddFavorite adds a quick launch favorite.
func (d *Database) AddFavorite(user, quickLaunchID string) (*QuickLaunchFavorite, error) {
	userID, err := d.GetUserID(user)
	if err != nil {
		return nil, err
	}

	id := uuid.New().String()
	_, err = d.db.Exec(
		"INSERT INTO quick_launch_favorites (id, quick_launch_id, user_id) VALUES ($1, $2, $3)",
		id, quickLaunchID, userID,
	)
	if err != nil {
		return nil, err
	}

	return d.GetFavorite(user, id)
}

// DeleteFavorite deletes a quick launch favorite.
func (d *Database) DeleteFavorite(user, favID string) error {
	userID, err := d.GetUserID(user)
	if err != nil {
		return err
	}

	_, err = d.db.Exec(
		"DELETE FROM quick_launch_favorites WHERE id = $1 AND user_id = $2",
		favID, userID,
	)
	return err
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
	QuickLaunchID *string `json:"quick_launch_id,omitempty"`
	AppID         *string `json:"app_id,omitempty"`
}

// GetUserDefault returns a quick launch user default.
func (d *Database) GetUserDefault(user, id string) (*QuickLaunchUserDefault, error) {
	userID, err := d.GetUserID(user)
	if err != nil {
		return nil, err
	}

	var ud QuickLaunchUserDefault
	err = d.db.QueryRowx(
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
		return nil, fmt.Errorf("user default not found: %s", id)
	}
	return &ud, nil
}

// GetAllUserDefaults returns all quick launch user defaults for a user.
func (d *Database) GetAllUserDefaults(user string) ([]QuickLaunchUserDefault, error) {
	userID, err := d.GetUserID(user)
	if err != nil {
		return nil, err
	}

	var uds []QuickLaunchUserDefault
	err = d.db.Select(&uds,
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
func (d *Database) AddUserDefault(user string, nud *NewQuickLaunchUserDefault) (*QuickLaunchUserDefault, error) {
	userID, err := d.GetUserID(user)
	if err != nil {
		return nil, err
	}

	id := uuid.New().String()
	_, err = d.db.Exec(
		`INSERT INTO quick_launch_user_defaults (id, user_id, quick_launch_id, app_id)
		 VALUES ($1, $2, $3, $4)`,
		id, userID, nud.QuickLaunchID, nud.AppID,
	)
	if err != nil {
		return nil, err
	}

	return d.GetUserDefault(user, id)
}

// UpdateUserDefault updates a quick launch user default.
func (d *Database) UpdateUserDefault(id, user string, update *UpdateQuickLaunchUserDefaultRequest) (*QuickLaunchUserDefault, error) {
	userID, err := d.GetUserID(user)
	if err != nil {
		return nil, err
	}

	// Verify it exists
	var exists bool
	err = d.db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM quick_launch_user_defaults WHERE id = $1)",
		id,
	).Scan(&exists)
	if err != nil || !exists {
		return nil, fmt.Errorf("user default not found: %s", id)
	}

	if update.AppID != nil {
		_, err = d.db.Exec(
			"UPDATE quick_launch_user_defaults SET app_id = $1 WHERE id = $2 AND user_id = $3",
			*update.AppID, id, userID,
		)
		if err != nil {
			return nil, err
		}
	}

	if update.QuickLaunchID != nil {
		_, err = d.db.Exec(
			"UPDATE quick_launch_user_defaults SET quick_launch_id = $1 WHERE id = $2 AND user_id = $3",
			*update.QuickLaunchID, id, userID,
		)
		if err != nil {
			return nil, err
		}
	}

	return d.GetUserDefault(user, id)
}

// DeleteUserDefault deletes a quick launch user default.
func (d *Database) DeleteUserDefault(user, id string) error {
	userID, err := d.GetUserID(user)
	if err != nil {
		return err
	}

	_, err = d.db.Exec(
		"DELETE FROM quick_launch_user_defaults WHERE id = $1 AND user_id = $2",
		id, userID,
	)
	return err
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
	AppID         *string `json:"app_id,omitempty"`
	QuickLaunchID *string `json:"quick_launch_id,omitempty"`
}

// GetGlobalDefault returns a quick launch global default.
func (d *Database) GetGlobalDefault(user, id string) (*QuickLaunchGlobalDefault, error) {
	userID, err := d.GetUserID(user)
	if err != nil {
		return nil, err
	}

	var gd QuickLaunchGlobalDefault
	err = d.db.QueryRowx(
		`SELECT gd.id, gd.app_id, gd.quick_launch_id
		   FROM quick_launch_global_defaults gd
		   JOIN quick_launches ql ON gd.quick_launch_id = ql.id
		  WHERE gd.id = $1
		    AND ql.creator = $2`,
		id, userID,
	).StructScan(&gd)
	if err != nil {
		return nil, fmt.Errorf("global default not found: %s", id)
	}
	return &gd, nil
}

// GetAllGlobalDefaults returns all quick launch global defaults for a user.
func (d *Database) GetAllGlobalDefaults(user string) ([]QuickLaunchGlobalDefault, error) {
	userID, err := d.GetUserID(user)
	if err != nil {
		return nil, err
	}

	var gds []QuickLaunchGlobalDefault
	err = d.db.Select(&gds,
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
func (d *Database) AddGlobalDefault(user string, ngd *NewQuickLaunchGlobalDefault) (*QuickLaunchGlobalDefault, error) {
	id := uuid.New().String()
	_, err := d.db.Exec(
		`INSERT INTO quick_launch_global_defaults (id, app_id, quick_launch_id)
		 VALUES ($1, $2, $3)`,
		id, ngd.AppID, ngd.QuickLaunchID,
	)
	if err != nil {
		return nil, err
	}

	return d.GetGlobalDefault(user, id)
}

// UpdateGlobalDefault updates a quick launch global default.
func (d *Database) UpdateGlobalDefault(id, user string, update *UpdateQuickLaunchGlobalDefaultRequest) (*QuickLaunchGlobalDefault, error) {
	userID, err := d.GetUserID(user)
	if err != nil {
		return nil, err
	}

	subquery := `SELECT ql.id FROM quick_launches ql
	             JOIN quick_launch_global_defaults gd ON ql.id = gd.quick_launch_id
	             WHERE ql.creator = $3`

	if update.AppID != nil {
		_, err = d.db.Exec(
			fmt.Sprintf("UPDATE quick_launch_global_defaults SET app_id = $1 WHERE id = $2 AND quick_launch_id IN (%s)", subquery),
			*update.AppID, id, userID,
		)
		if err != nil {
			return nil, err
		}
	}

	if update.QuickLaunchID != nil {
		_, err = d.db.Exec(
			fmt.Sprintf("UPDATE quick_launch_global_defaults SET quick_launch_id = $1 WHERE id = $2 AND quick_launch_id IN (%s)", subquery),
			*update.QuickLaunchID, id, userID,
		)
		if err != nil {
			return nil, err
		}
	}

	return d.GetGlobalDefault(user, id)
}

// DeleteGlobalDefault deletes a quick launch global default.
func (d *Database) DeleteGlobalDefault(user, id string) error {
	userID, err := d.GetUserID(user)
	if err != nil {
		return err
	}

	_, err = d.db.Exec(
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

// ListConcurrentJobLimits returns all defined concurrent job limits.
func (d *Database) ListConcurrentJobLimits() ([]ConcurrentJobLimit, error) {
	var limits []ConcurrentJobLimit
	err := d.db.Select(&limits,
		`SELECT launcher AS username,
		        concurrent_jobs,
		        (launcher IS NULL) AS is_default
		   FROM job_limits
		  ORDER BY launcher ASC`,
	)
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
func (d *Database) GetConcurrentJobLimit(username string) (*ConcurrentJobLimit, error) {
	var limits []ConcurrentJobLimit
	err := d.db.Select(&limits,
		`SELECT launcher AS username,
		        concurrent_jobs,
		        (launcher IS NULL) AS is_default
		   FROM job_limits
		  WHERE launcher = regexp_replace($1, '-', '_')
		     OR launcher IS NULL
		  ORDER BY is_default ASC`,
		username,
	)
	if err != nil {
		return nil, err
	}
	if len(limits) == 0 {
		return nil, fmt.Errorf("no job limit found for user: %s", username)
	}
	return &limits[0], nil
}

// SetConcurrentJobLimit sets the concurrent job limit for a user.
func (d *Database) SetConcurrentJobLimit(username string, limit int) (*ConcurrentJobLimit, error) {
	tx, err := d.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() //nolint:errcheck

	// Check if the user already has a specific limit
	current, _ := d.getConcurrentJobLimitTx(tx, username)

	if current == nil || current.IsDefault {
		// Insert new limit
		_, err = tx.Exec(
			"INSERT INTO job_limits (launcher, concurrent_jobs) VALUES (regexp_replace($1, '-', '_'), $2)",
			username, limit,
		)
	} else {
		// Update existing limit
		_, err = tx.Exec(
			"UPDATE job_limits SET concurrent_jobs = $1 WHERE launcher = regexp_replace($2, '-', '_')",
			limit, username,
		)
	}
	if err != nil {
		return nil, err
	}

	result, err := d.getConcurrentJobLimitTx(tx, username)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return result, nil
}

func (d *Database) getConcurrentJobLimitTx(tx *sqlx.Tx, username string) (*ConcurrentJobLimit, error) {
	var limits []ConcurrentJobLimit
	err := tx.Select(&limits,
		`SELECT launcher AS username,
		        concurrent_jobs,
		        (launcher IS NULL) AS is_default
		   FROM job_limits
		  WHERE launcher = regexp_replace($1, '-', '_')
		     OR launcher IS NULL
		  ORDER BY is_default ASC`,
		username,
	)
	if err != nil {
		return nil, err
	}
	if len(limits) == 0 {
		return nil, fmt.Errorf("no job limit found for user: %s", username)
	}
	return &limits[0], nil
}

// RemoveConcurrentJobLimit removes a user's concurrent job limit, returning the default.
func (d *Database) RemoveConcurrentJobLimit(username string) (*ConcurrentJobLimit, error) {
	tx, err := d.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() //nolint:errcheck

	_, err = tx.Exec(
		"DELETE FROM job_limits WHERE launcher = regexp_replace($1, '-', '_')",
		username,
	)
	if err != nil {
		return nil, err
	}

	// Return the default limit
	result, err := d.getConcurrentJobLimitTx(tx, "")
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return result, nil
}

// DeletionResponse is the standard response for deletion operations.
type DeletionResponse struct {
	ID string `json:"id"`
}
