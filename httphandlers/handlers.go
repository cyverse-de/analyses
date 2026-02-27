package httphandlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Handlers holds dependencies for HTTP handlers.
type Handlers struct {
	DB             DatabaseStore
	AppsClient     AppFetcher
	DataInfoClient PathChecker
}

// NewHandlers creates a new Handlers instance.
func NewHandlers(database DatabaseStore, appsClient AppFetcher, dataInfoClient PathChecker) *Handlers {
	return &Handlers{
		DB:             database,
		AppsClient:     appsClient,
		DataInfoClient: dataInfoClient,
	}
}

// requireUser extracts the "user" query parameter from the request,
// returning an HTTP 400 error if it is missing.
func requireUser(c echo.Context) (string, error) {
	user := c.QueryParam("user")
	if user == "" {
		return "", echo.NewHTTPError(http.StatusBadRequest, "user query parameter is required")
	}
	return user, nil
}
