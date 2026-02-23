package httphandlers

import (
	"net/http"

	"github.com/cyverse-de/analyses/db"
	"github.com/labstack/echo/v4"
)

// AddUserDefaultHandler adds a quick launch user default.
//
//	@Summary		Add a Quick Launch user default
//	@Description	Add a Quick Launch user default. A new UUID will be assigned to the user default and will be returned in the response.
//	@Tags			quicklaunch-user-defaults
//	@Accept			json
//	@Produce		json
//	@Param			user	query		string							true	"Username"
//	@Param			body	body		db.NewQuickLaunchUserDefault	true	"User default to add"
//	@Success		200		{object}	db.QuickLaunchUserDefault
//	@Failure		400		{object}	common.ErrorResponse
//	@Failure		500		{object}	common.ErrorResponse
//	@Router			/quicklaunch/defaults/user [post]
func (h *Handlers) AddUserDefaultHandler(c echo.Context) error {
	user := c.QueryParam("user")
	if user == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "user query parameter is required")
	}

	var nud db.NewQuickLaunchUserDefault
	if err := c.Bind(&nud); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ud, err := h.DB.AddUserDefault(user, &nud)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, ud)
}

// GetUserDefaultHandler returns a quick launch user default.
//
//	@Summary		Get a Quick Launch user default
//	@Description	Get a Quick Launch user default.
//	@Tags			quicklaunch-user-defaults
//	@Produce		json
//	@Param			id		path		string	true	"User Default ID"
//	@Param			user	query		string	true	"Username"
//	@Success		200		{object}	db.QuickLaunchUserDefault
//	@Failure		400		{object}	common.ErrorResponse
//	@Failure		404		{object}	common.ErrorResponse
//	@Router			/quicklaunch/defaults/user/{id} [get]
func (h *Handlers) GetUserDefaultHandler(c echo.Context) error {
	id := c.Param("id")
	user := c.QueryParam("user")
	if user == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "user query parameter is required")
	}

	ud, err := h.DB.GetUserDefault(user, id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, ud)
}

// GetAllUserDefaultsHandler returns all quick launch user defaults.
//
//	@Summary		Get all Quick Launch user defaults
//	@Description	Get all of the Quick Launch user defaults for the logged in user.
//	@Tags			quicklaunch-user-defaults
//	@Produce		json
//	@Param			user	query		string	true	"Username"
//	@Success		200		{array}		db.QuickLaunchUserDefault
//	@Failure		400		{object}	common.ErrorResponse
//	@Failure		500		{object}	common.ErrorResponse
//	@Router			/quicklaunch/defaults/user [get]
func (h *Handlers) GetAllUserDefaultsHandler(c echo.Context) error {
	user := c.QueryParam("user")
	if user == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "user query parameter is required")
	}

	uds, err := h.DB.GetAllUserDefaults(user)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, uds)
}

// UpdateUserDefaultHandler updates a quick launch user default.
//
//	@Summary		Update a Quick Launch user default
//	@Description	Modifies an existing Quick Launch user default.
//	@Tags			quicklaunch-user-defaults
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string									true	"User Default ID"
//	@Param			user	query		string									true	"Username"
//	@Param			body	body		db.UpdateQuickLaunchUserDefaultRequest	true	"Fields to update"
//	@Success		200		{object}	db.QuickLaunchUserDefault
//	@Failure		400		{object}	common.ErrorResponse
//	@Failure		404		{object}	common.ErrorResponse
//	@Router			/quicklaunch/defaults/user/{id} [patch]
func (h *Handlers) UpdateUserDefaultHandler(c echo.Context) error {
	id := c.Param("id")
	user := c.QueryParam("user")
	if user == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "user query parameter is required")
	}

	var update db.UpdateQuickLaunchUserDefaultRequest
	if err := c.Bind(&update); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ud, err := h.DB.UpdateUserDefault(id, user, &update)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, ud)
}

// DeleteUserDefaultHandler deletes a quick launch user default.
//
//	@Summary		Delete a Quick Launch user default
//	@Description	Delete the Quick Launch user default.
//	@Tags			quicklaunch-user-defaults
//	@Produce		json
//	@Param			id		path		string	true	"User Default ID"
//	@Param			user	query		string	true	"Username"
//	@Success		200		{object}	db.DeletionResponse
//	@Failure		400		{object}	common.ErrorResponse
//	@Router			/quicklaunch/defaults/user/{id} [delete]
func (h *Handlers) DeleteUserDefaultHandler(c echo.Context) error {
	id := c.Param("id")
	user := c.QueryParam("user")
	if user == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "user query parameter is required")
	}

	if err := h.DB.DeleteUserDefault(user, id); err != nil {
		if db.IsNotFound(err) {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, db.DeletionResponse{ID: id})
}
