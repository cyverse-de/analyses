package httphandlers

import (
	"net/http"

	"github.com/cyverse-de/analyses/db"
	"github.com/labstack/echo/v4"
)

// AddGlobalDefaultHandler adds a quick launch global default.
//
//	@Summary		Add a Quick Launch global default
//	@Description	Add a new Quick Launch global default. Assigns a new UUID.
//	@Tags			quicklaunch-global-defaults
//	@Accept			json
//	@Produce		json
//	@Param			user	query		string							true	"Username"
//	@Param			body	body		db.NewQuickLaunchGlobalDefault	true	"Global default to add"
//	@Success		200		{object}	db.QuickLaunchGlobalDefault
//	@Failure		400		{object}	common.ErrorResponse
//	@Failure		500		{object}	common.ErrorResponse
//	@Router			/quicklaunch/defaults/global [post]
func (h *Handlers) AddGlobalDefaultHandler(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := requireUser(c)
	if err != nil {
		return err
	}

	var ngd db.NewQuickLaunchGlobalDefault
	if err := c.Bind(&ngd); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	tx, err := h.DB.BeginTx(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer tx.Rollback() //nolint:errcheck

	gd, err := h.DB.AddGlobalDefault(ctx, tx, user, &ngd)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, gd)
}

// GetGlobalDefaultHandler returns a quick launch global default.
//
//	@Summary		Get a Quick Launch global default
//	@Description	Get a Quick Launch global default.
//	@Tags			quicklaunch-global-defaults
//	@Produce		json
//	@Param			id		path		string	true	"Global Default ID"
//	@Param			user	query		string	true	"Username"
//	@Success		200		{object}	db.QuickLaunchGlobalDefault
//	@Failure		400		{object}	common.ErrorResponse
//	@Failure		404		{object}	common.ErrorResponse
//	@Router			/quicklaunch/defaults/global/{id} [get]
func (h *Handlers) GetGlobalDefaultHandler(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	user, err := requireUser(c)
	if err != nil {
		return err
	}

	tx, err := h.DB.BeginTx(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer tx.Rollback() //nolint:errcheck

	gd, err := h.DB.GetGlobalDefault(ctx, tx, user, id)
	if err != nil {
		if db.IsNotFound(err) {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, gd)
}

// GetAllGlobalDefaultsHandler returns all quick launch global defaults.
//
//	@Summary		Get all Quick Launch global defaults
//	@Description	Get all of the Quick Launch global defaults that the user has created.
//	@Tags			quicklaunch-global-defaults
//	@Produce		json
//	@Param			user	query		string	true	"Username"
//	@Success		200		{array}		db.QuickLaunchGlobalDefault
//	@Failure		400		{object}	common.ErrorResponse
//	@Failure		500		{object}	common.ErrorResponse
//	@Router			/quicklaunch/defaults/global [get]
func (h *Handlers) GetAllGlobalDefaultsHandler(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := requireUser(c)
	if err != nil {
		return err
	}

	tx, err := h.DB.BeginTx(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer tx.Rollback() //nolint:errcheck

	gds, err := h.DB.GetAllGlobalDefaults(ctx, tx, user)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, gds)
}

// UpdateGlobalDefaultHandler updates a quick launch global default.
//
//	@Summary		Update a Quick Launch global default
//	@Description	Modifies an existing Quick Launch global default.
//	@Tags			quicklaunch-global-defaults
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string										true	"Global Default ID"
//	@Param			user	query		string										true	"Username"
//	@Param			body	body		db.UpdateQuickLaunchGlobalDefaultRequest	true	"Fields to update"
//	@Success		200		{object}	db.QuickLaunchGlobalDefault
//	@Failure		400		{object}	common.ErrorResponse
//	@Failure		404		{object}	common.ErrorResponse
//	@Router			/quicklaunch/defaults/global/{id} [patch]
func (h *Handlers) UpdateGlobalDefaultHandler(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	user, err := requireUser(c)
	if err != nil {
		return err
	}

	var update db.UpdateQuickLaunchGlobalDefaultRequest
	if err := c.Bind(&update); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if update.AppID == "" || update.QuickLaunchID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "app_id and quick_launch_id are required")
	}

	tx, err := h.DB.BeginTx(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer tx.Rollback() //nolint:errcheck

	gd, err := h.DB.UpdateGlobalDefault(ctx, tx, id, user, &update)
	if err != nil {
		if db.IsNotFound(err) {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, gd)
}

// DeleteGlobalDefaultHandler deletes a quick launch global default.
//
//	@Summary		Delete a Quick Launch global default
//	@Description	Delete the Quick Launch global default.
//	@Tags			quicklaunch-global-defaults
//	@Produce		json
//	@Param			id		path		string	true	"Global Default ID"
//	@Param			user	query		string	true	"Username"
//	@Success		200		{object}	db.DeletionResponse
//	@Failure		400		{object}	common.ErrorResponse
//	@Router			/quicklaunch/defaults/global/{id} [delete]
func (h *Handlers) DeleteGlobalDefaultHandler(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	user, err := requireUser(c)
	if err != nil {
		return err
	}

	tx, err := h.DB.BeginTx(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer tx.Rollback() //nolint:errcheck

	if err := h.DB.DeleteGlobalDefault(ctx, tx, user, id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, db.DeletionResponse{ID: id})
}
