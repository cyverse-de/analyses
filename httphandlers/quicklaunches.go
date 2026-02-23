package httphandlers

import (
	"encoding/json"
	"net/http"

	"github.com/cyverse-de/analyses/clients"
	"github.com/cyverse-de/analyses/db"
	"github.com/labstack/echo/v4"
)

// Handlers holds dependencies for HTTP handlers.
type Handlers struct {
	DB             *db.Database
	AppsClient     *clients.AppsClient
	DataInfoClient *clients.DataInfoClient
}

// NewHandlers creates a new Handlers instance.
func NewHandlers(database *db.Database, appsClient *clients.AppsClient, dataInfoClient *clients.DataInfoClient) *Handlers {
	return &Handlers{
		DB:             database,
		AppsClient:     appsClient,
		DataInfoClient: dataInfoClient,
	}
}

const systemID = "de"

// AddQuickLaunchHandler handles creating a new quick launch.
//
//	@Summary		Add a Quick Launch
//	@Description	Adds a Quick Launch and corresponding submission information to the database. The username passed in should already exist. A new UUID will be assigned and returned.
//	@Tags			quicklaunches
//	@Accept			json
//	@Produce		json
//	@Param			user	query		string				true	"Username"
//	@Param			body	body		db.NewQuickLaunch	true	"Quick Launch to add"
//	@Success		200		{object}	db.QuickLaunch
//	@Failure		400		{object}	common.ErrorResponse
//	@Failure		500		{object}	common.ErrorResponse
//	@Router			/quicklaunches [post]
func (h *Handlers) AddQuickLaunchHandler(c echo.Context) error {
	user := c.QueryParam("user")
	if user == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "user query parameter is required")
	}

	var nql db.NewQuickLaunch
	if err := c.Bind(&nql); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Fetch app to validate and get version_id if needed
	var app map[string]interface{}
	var err error
	if nql.AppVersionID != "" {
		app, err = h.AppsClient.GetAppVersion(user, systemID, nql.AppID, nql.AppVersionID)
	} else {
		app, err = h.AppsClient.GetApp(user, systemID, nql.AppID)
		if err == nil {
			if vid, ok := app["version_id"].(string); ok {
				nql.AppVersionID = vid
			}
		}
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to get app: "+err.Error())
	}

	// Validate submission
	var submissionMap map[string]interface{}
	if err := json.Unmarshal(nql.Submission, &submissionMap); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid submission JSON")
	}
	config, _ := submissionMap["config"].(map[string]interface{})

	valReq := &clients.ValidationRequest{
		App:      app,
		Config:   config,
		IsPublic: nql.IsPublic,
		User:     user,
	}
	if err := clients.ValidateSubmission(h.AppsClient, h.DataInfoClient, valReq); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ql, err := h.DB.AddQuickLaunch(user, &nql)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, ql)
}

// GetAllQuickLaunchesHandler returns all quick launches for a user.
//
//	@Summary		Get all Quick Launches
//	@Description	Gets all of the Quick Launches for a user. Includes UUIDs.
//	@Tags			quicklaunches
//	@Produce		json
//	@Param			user	query		string	true	"Username"
//	@Success		200		{array}		db.QuickLaunch
//	@Failure		400		{object}	common.ErrorResponse
//	@Failure		500		{object}	common.ErrorResponse
//	@Router			/quicklaunches [get]
func (h *Handlers) GetAllQuickLaunchesHandler(c echo.Context) error {
	user := c.QueryParam("user")
	if user == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "user query parameter is required")
	}

	qls, err := h.DB.GetAllQuickLaunches(user)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, qls)
}

// GetQuickLaunchesByAppHandler returns quick launches for a given app.
//
//	@Summary		Get Quick Launches by App
//	@Description	Returns a list of Quick Launches that the user can access based on the app's UUID.
//	@Tags			quicklaunches
//	@Produce		json
//	@Param			id		path		string	true	"App ID"
//	@Param			user	query		string	true	"Username"
//	@Success		200		{array}		db.QuickLaunch
//	@Failure		400		{object}	common.ErrorResponse
//	@Failure		500		{object}	common.ErrorResponse
//	@Router			/quicklaunches/apps/{id} [get]
func (h *Handlers) GetQuickLaunchesByAppHandler(c echo.Context) error {
	appID := c.Param("id")
	user := c.QueryParam("user")
	if user == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "user query parameter is required")
	}

	qls, err := h.DB.GetQuickLaunchesByApp(appID, user)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, qls)
}

// GetQuickLaunchHandler returns a single quick launch by ID.
//
//	@Summary		Get a Quick Launch
//	@Description	Gets the Quick Launch information from the database, including its UUID, the name of the user that owns it, and the submission JSON.
//	@Tags			quicklaunches
//	@Produce		json
//	@Param			id		path		string	true	"Quick Launch ID"
//	@Param			user	query		string	true	"Username"
//	@Success		200		{object}	db.QuickLaunch
//	@Failure		400		{object}	common.ErrorResponse
//	@Failure		404		{object}	common.ErrorResponse
//	@Router			/quicklaunches/{id} [get]
func (h *Handlers) GetQuickLaunchHandler(c echo.Context) error {
	id := c.Param("id")
	user := c.QueryParam("user")
	if user == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "user query parameter is required")
	}

	ql, err := h.DB.GetQuickLaunch(id, user)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, ql)
}

// UpdateQuickLaunchHandler updates an existing quick launch.
//
//	@Summary		Update a Quick Launch
//	@Description	Modifies an existing Quick Launch, allowing the caller to change owners and the contents of the submission JSON.
//	@Tags			quicklaunches
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string						true	"Quick Launch ID"
//	@Param			user	query		string						true	"Username"
//	@Param			body	body		db.UpdateQuickLaunchRequest	true	"Fields to update"
//	@Success		200		{object}	db.QuickLaunch
//	@Failure		400		{object}	common.ErrorResponse
//	@Failure		404		{object}	common.ErrorResponse
//	@Router			/quicklaunches/{id} [patch]
func (h *Handlers) UpdateQuickLaunchHandler(c echo.Context) error {
	id := c.Param("id")
	user := c.QueryParam("user")
	if user == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "user query parameter is required")
	}

	var uql db.UpdateQuickLaunchRequest
	if err := c.Bind(&uql); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Get the existing QL to merge with update and validate
	existing, err := h.DB.GetQuickLaunch(id, user)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	appID := existing.AppID
	if uql.AppID != nil {
		appID = *uql.AppID
	}
	appVersionID := existing.AppVersionID
	if uql.AppVersionID != nil {
		appVersionID = *uql.AppVersionID
	}

	app, err := h.AppsClient.GetAppVersion(user, systemID, appID, appVersionID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to get app version: "+err.Error())
	}

	// Determine the submission config for validation
	var submissionForValidation map[string]interface{}
	if uql.Submission != nil {
		merged, merr := h.DB.MergeSubmission(id, user, *uql.Submission)
		if merr != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, merr.Error())
		}
		json.Unmarshal(merged, &submissionForValidation) //nolint:errcheck
	} else {
		json.Unmarshal(existing.Submission, &submissionForValidation) //nolint:errcheck
	}
	config, _ := submissionForValidation["config"].(map[string]interface{})

	isPublic := existing.IsPublic
	if uql.IsPublic != nil {
		isPublic = *uql.IsPublic
	}

	valReq := &clients.ValidationRequest{
		App:      app,
		Config:   config,
		IsPublic: isPublic,
		User:     user,
	}
	if err := clients.ValidateSubmission(h.AppsClient, h.DataInfoClient, valReq); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ql, err := h.DB.UpdateQuickLaunch(id, user, &uql)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, ql)
}

// DeleteQuickLaunchHandler deletes a quick launch.
//
//	@Summary		Delete a Quick Launch
//	@Description	Deletes a Quick Launch from the database. Will return a success even if called on a Quick Launch that has either already been deleted or never existed.
//	@Tags			quicklaunches
//	@Produce		json
//	@Param			id		path		string	true	"Quick Launch ID"
//	@Param			user	query		string	true	"Username"
//	@Success		200		{object}	db.DeletionResponse
//	@Failure		400		{object}	common.ErrorResponse
//	@Router			/quicklaunches/{id} [delete]
func (h *Handlers) DeleteQuickLaunchHandler(c echo.Context) error {
	id := c.Param("id")
	user := c.QueryParam("user")
	if user == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "user query parameter is required")
	}

	if err := h.DB.DeleteQuickLaunch(id, user); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, db.DeletionResponse{ID: id})
}

// QuickLaunchAppInfoHandler returns the app info populated with submission values.
//
//	@Summary		Get Quick Launch app info
//	@Description	Returns the app info needed to create and populate the app launcher in the UI. Populates the parameters with the values from the submission stored for the quick launch.
//	@Tags			quicklaunches
//	@Produce		json
//	@Param			id		path		string	true	"Quick Launch ID"
//	@Param			user	query		string	true	"Username"
//	@Success		200		{object}	map[string]interface{}
//	@Failure		400		{object}	common.ErrorResponse
//	@Failure		404		{object}	common.ErrorResponse
//	@Router			/quicklaunches/{id}/app-info [get]
func (h *Handlers) QuickLaunchAppInfoHandler(c echo.Context) error {
	id := c.Param("id")
	user := c.QueryParam("user")
	if user == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "user query parameter is required")
	}

	ql, err := h.DB.GetQuickLaunch(id, user)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	app, err := h.AppsClient.GetAppVersion(user, systemID, ql.AppID, ql.AppVersionID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get app version: "+err.Error())
	}

	var submission map[string]interface{}
	if err := json.Unmarshal(ql.Submission, &submission); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to parse submission: "+err.Error())
	}

	result := clients.QuickLaunchAppInfo(submission, app, systemID)
	return c.JSON(http.StatusOK, result)
}
