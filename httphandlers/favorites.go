package httphandlers

import (
	"net/http"

	"github.com/cyverse-de/analyses/db"
	"github.com/labstack/echo/v4"
)

// NewQuickLaunchFavoriteRequest is the request body for adding a favorite.
type NewQuickLaunchFavoriteRequest struct {
	QuickLaunchID string `json:"quick_launch_id"`
}

// AddFavoriteHandler adds a quick launch favorite.
//
//	@Summary		Add a Quick Launch favorite
//	@Description	Adds a favorite Quick Launch to the database for the user. The username passed in should already exist. A new UUID will be assigned to the favorite and returned with the rest of the record.
//	@Tags			quicklaunch-favorites
//	@Accept			json
//	@Produce		json
//	@Param			user	query		string							true	"Username"
//	@Param			body	body		NewQuickLaunchFavoriteRequest	true	"Quick Launch to favorite"
//	@Success		200		{object}	db.QuickLaunchFavorite
//	@Failure		400		{object}	common.ErrorResponse
//	@Failure		500		{object}	common.ErrorResponse
//	@Router			/quicklaunch/favorites [post]
func (h *Handlers) AddFavoriteHandler(c echo.Context) error {
	user, err := requireUser(c)
	if err != nil {
		return err
	}

	var req NewQuickLaunchFavoriteRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	tx, err := h.DB.BeginTx()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer tx.Rollback() //nolint:errcheck

	fav, err := h.DB.AddFavorite(tx, user, req.QuickLaunchID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, fav)
}

// GetFavoriteHandler returns a single quick launch favorite.
//
//	@Summary		Get a Quick Launch favorite
//	@Description	Gets information about a favorited Quick Launch. Returns a Quick Launch UUID which can be passed to the /quicklaunches endpoints to grab more information.
//	@Tags			quicklaunch-favorites
//	@Produce		json
//	@Param			id		path		string	true	"Favorite ID"
//	@Param			user	query		string	true	"Username"
//	@Success		200		{object}	db.QuickLaunchFavorite
//	@Failure		400		{object}	common.ErrorResponse
//	@Failure		404		{object}	common.ErrorResponse
//	@Router			/quicklaunch/favorites/{id} [get]
func (h *Handlers) GetFavoriteHandler(c echo.Context) error {
	id := c.Param("id")
	user, err := requireUser(c)
	if err != nil {
		return err
	}

	tx, err := h.DB.BeginTx()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer tx.Rollback() //nolint:errcheck

	fav, err := h.DB.GetFavorite(tx, user, id)
	if err != nil {
		if db.IsNotFound(err) {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, fav)
}

// GetAllFavoritesHandler returns all quick launch favorites for a user.
//
//	@Summary		Get all Quick Launch favorites
//	@Description	Gets all of the user's Quick Launch favorites.
//	@Tags			quicklaunch-favorites
//	@Produce		json
//	@Param			user	query		string	true	"Username"
//	@Success		200		{array}		db.QuickLaunchFavorite
//	@Failure		400		{object}	common.ErrorResponse
//	@Failure		500		{object}	common.ErrorResponse
//	@Router			/quicklaunch/favorites [get]
func (h *Handlers) GetAllFavoritesHandler(c echo.Context) error {
	user, err := requireUser(c)
	if err != nil {
		return err
	}

	tx, err := h.DB.BeginTx()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer tx.Rollback() //nolint:errcheck

	favs, err := h.DB.GetAllFavorites(tx, user)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, favs)
}

// DeleteFavoriteHandler deletes a quick launch favorite.
//
//	@Summary		Delete a Quick Launch favorite
//	@Description	Deletes a Quick Launch favorite. Does not delete the actual Quick Launch, just the entry that listed it as a favorite for the user.
//	@Tags			quicklaunch-favorites
//	@Produce		json
//	@Param			id		path		string	true	"Favorite ID"
//	@Param			user	query		string	true	"Username"
//	@Success		200		{object}	db.DeletionResponse
//	@Failure		400		{object}	common.ErrorResponse
//	@Router			/quicklaunch/favorites/{id} [delete]
func (h *Handlers) DeleteFavoriteHandler(c echo.Context) error {
	id := c.Param("id")
	user, err := requireUser(c)
	if err != nil {
		return err
	}

	tx, err := h.DB.BeginTx()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer tx.Rollback() //nolint:errcheck

	if err := h.DB.DeleteFavorite(tx, user, id); err != nil {
		if db.IsNotFound(err) {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, db.DeletionResponse{ID: id})
}
