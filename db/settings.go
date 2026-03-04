package db

import "context"

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
func (d *Database) ListConcurrentJobLimits(ctx context.Context, tx Tx) ([]ConcurrentJobLimit, error) {
	var limits []ConcurrentJobLimit
	err := tx.SelectContext(ctx, &limits, jobLimitSelectSQL+` ORDER BY launcher ASC`)
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
func (d *Database) GetConcurrentJobLimit(ctx context.Context, tx Tx, username string) (*ConcurrentJobLimit, error) {
	var limits []ConcurrentJobLimit
	err := tx.SelectContext(ctx, &limits,
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
func (d *Database) SetConcurrentJobLimit(ctx context.Context, tx Tx, username string, limit int) (*ConcurrentJobLimit, error) {
	current, err := d.GetConcurrentJobLimit(ctx, tx, username)
	if err != nil && !IsNotFound(err) {
		return nil, err
	}

	if current == nil || current.IsDefault {
		_, err = tx.ExecContext(ctx,
			"INSERT INTO job_limits (launcher, concurrent_jobs) VALUES (regexp_replace($1, '-', '_'), $2)",
			username, limit,
		)
	} else {
		_, err = tx.ExecContext(ctx,
			"UPDATE job_limits SET concurrent_jobs = $1 WHERE launcher = regexp_replace($2, '-', '_')",
			limit, username,
		)
	}
	if err != nil {
		return nil, err
	}

	return d.GetConcurrentJobLimit(ctx, tx, username)
}

// RemoveConcurrentJobLimit removes a user's concurrent job limit, returning the default.
func (d *Database) RemoveConcurrentJobLimit(ctx context.Context, tx Tx, username string) (*ConcurrentJobLimit, error) {
	_, err := tx.ExecContext(ctx,
		"DELETE FROM job_limits WHERE launcher = regexp_replace($1, '-', '_')",
		username,
	)
	if err != nil {
		return nil, err
	}

	// Return the default limit.
	return d.GetConcurrentJobLimit(ctx, tx, "")
}
