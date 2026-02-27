package httphandlers

import (
	"net/http"

	"github.com/cyverse-de/analyses/db"
	"github.com/labstack/echo/v4"
)

// ListConcurrentJobLimitsHandler lists all concurrent job limits.
//
//	@Summary		List concurrent job limits
//	@Description	Lists all defined concurrent job limits.
//	@Tags			settings
//	@Produce		json
//	@Success		200	{object}	db.ConcurrentJobLimits
//	@Failure		500	{object}	common.ErrorResponse
//	@Router			/settings/concurrent-job-limits [get]
func (h *Handlers) ListConcurrentJobLimitsHandler(c echo.Context) error {
	ctx := c.Request().Context()

	tx, err := h.DB.BeginTx(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer tx.Rollback() //nolint:errcheck

	limits, err := h.DB.ListConcurrentJobLimits(ctx, tx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, db.ConcurrentJobLimits{Limits: limits})
}

// GetConcurrentJobLimitHandler gets the concurrent job limit for a user.
//
//	@Summary		Get concurrent job limit
//	@Description	Gets the concurrent job limit for a user. The returned job limit may either be the limit that was explicitly assigned to the user or the default job limit.
//	@Tags			settings
//	@Produce		json
//	@Param			username	path		string	true	"Username"
//	@Success		200			{object}	db.ConcurrentJobLimit
//	@Failure		404			{object}	common.ErrorResponse
//	@Router			/settings/concurrent-job-limits/{username} [get]
func (h *Handlers) GetConcurrentJobLimitHandler(c echo.Context) error {
	ctx := c.Request().Context()
	username := c.Param("username")

	tx, err := h.DB.BeginTx(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer tx.Rollback() //nolint:errcheck

	limit, err := h.DB.GetConcurrentJobLimit(ctx, tx, username)
	if err != nil {
		if db.IsNotFound(err) {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, limit)
}

// SetConcurrentJobLimitHandler sets the concurrent job limit for a user.
//
//	@Summary		Set concurrent job limit
//	@Description	Sets the concurrent job limit for a user. The user's limit will be updated explicitly in the database even if the requested limit is the same as the default limit.
//	@Tags			settings
//	@Accept			json
//	@Produce		json
//	@Param			username	path		string						true	"Username"
//	@Param			body		body		db.ConcurrentJobLimitUpdate	true	"Job limit to set"
//	@Success		200			{object}	db.ConcurrentJobLimit
//	@Failure		400			{object}	common.ErrorResponse
//	@Failure		500			{object}	common.ErrorResponse
//	@Router			/settings/concurrent-job-limits/{username} [put]
func (h *Handlers) SetConcurrentJobLimitHandler(c echo.Context) error {
	ctx := c.Request().Context()
	username := c.Param("username")

	var update db.ConcurrentJobLimitUpdate
	if err := c.Bind(&update); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	tx, err := h.DB.BeginTx(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer tx.Rollback() //nolint:errcheck

	limit, err := h.DB.SetConcurrentJobLimit(ctx, tx, username, update.ConcurrentJobs)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, limit)
}

// RemoveConcurrentJobLimitHandler removes a user's concurrent job limit.
//
//	@Summary		Remove concurrent job limit
//	@Description	Removes the explicitly set concurrent job limit for a user. This will effectively return the user's limit to whatever the default limit is.
//	@Tags			settings
//	@Produce		json
//	@Param			username	path		string	true	"Username"
//	@Success		200			{object}	db.ConcurrentJobLimit
//	@Failure		500			{object}	common.ErrorResponse
//	@Router			/settings/concurrent-job-limits/{username} [delete]
func (h *Handlers) RemoveConcurrentJobLimitHandler(c echo.Context) error {
	ctx := c.Request().Context()
	username := c.Param("username")

	tx, err := h.DB.BeginTx(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer tx.Rollback() //nolint:errcheck

	limit, err := h.DB.RemoveConcurrentJobLimit(ctx, tx, username)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, limit)
}
