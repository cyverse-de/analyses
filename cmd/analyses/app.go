package main

import (
	"net/http"

	"github.com/cyverse-de/analyses/clients"
	"github.com/cyverse-de/analyses/common"
	"github.com/cyverse-de/analyses/db"
	"github.com/cyverse-de/analyses/httphandlers"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"

	_ "github.com/cyverse-de/analyses/docs"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// AnalysesApp encapsulates the application, tying together the REST API with the database.
type AnalysesApp struct {
	router   *echo.Echo
	db       *sqlx.DB
	handlers *httphandlers.Handlers
}

//	@title			Analyses API
//	@version		1.0
//	@description	API for the Discovery Environment's analyses, including Quick Launches and settings.
//
//	@license.name	3-Clause BSD License
//	@license.url	https://github.com/cyverse-de/analyses?tab=License-1-ov-file#readme
//
//	@host			localhost:60000
//	@BasePath		/

// NewAnalysesApp creates and returns a new AnalysesApp.
func NewAnalysesApp(database *sqlx.DB, appsBaseURL, dataInfoBaseURL string) *AnalysesApp {
	appsClient := clients.NewAppsClient(appsBaseURL)
	dataInfoClient := clients.NewDataInfoClient(dataInfoBaseURL)
	handlers := httphandlers.NewHandlers(db.New(database), appsClient, dataInfoClient)

	app := &AnalysesApp{
		router:   echo.New(),
		db:       database,
		handlers: handlers,
	}

	app.router.HTTPErrorHandler = func(err error, c echo.Context) {
		code := http.StatusInternalServerError
		var body interface{}

		switch err := err.(type) {
		case common.ErrorResponse:
			code = http.StatusBadRequest
			body = err
		case *common.ErrorResponse:
			code = http.StatusBadRequest
			body = err
		case *echo.HTTPError:
			echoErr := err
			code = echoErr.Code
			body = common.NewErrorResponse(err)
		default:
			body = common.NewErrorResponse(err)
		}

		c.JSON(code, body) //nolint:errcheck
	}

	app.router.GET("/", app.Greeting).Name = "greeting"
	app.router.GET("/docs/*", echoSwagger.WrapHandler)

	// Quick Launches
	ql := app.router.Group("/quicklaunches")
	ql.POST("", handlers.AddQuickLaunchHandler)
	ql.POST("/", handlers.AddQuickLaunchHandler)
	ql.GET("", handlers.GetAllQuickLaunchesHandler)
	ql.GET("/", handlers.GetAllQuickLaunchesHandler)
	ql.GET("/apps/:id", handlers.GetQuickLaunchesByAppHandler)
	ql.GET("/apps/:id/", handlers.GetQuickLaunchesByAppHandler)
	ql.GET("/:id", handlers.GetQuickLaunchHandler)
	ql.GET("/:id/", handlers.GetQuickLaunchHandler)
	ql.PATCH("/:id", handlers.UpdateQuickLaunchHandler)
	ql.PATCH("/:id/", handlers.UpdateQuickLaunchHandler)
	ql.DELETE("/:id", handlers.DeleteQuickLaunchHandler)
	ql.DELETE("/:id/", handlers.DeleteQuickLaunchHandler)
	ql.GET("/:id/app-info", handlers.QuickLaunchAppInfoHandler)
	ql.GET("/:id/app-info/", handlers.QuickLaunchAppInfoHandler)

	// Quick Launch Favorites
	fav := app.router.Group("/quicklaunch/favorites")
	fav.POST("", handlers.AddFavoriteHandler)
	fav.POST("/", handlers.AddFavoriteHandler)
	fav.GET("", handlers.GetAllFavoritesHandler)
	fav.GET("/", handlers.GetAllFavoritesHandler)
	fav.GET("/:id", handlers.GetFavoriteHandler)
	fav.DELETE("/:id", handlers.DeleteFavoriteHandler)

	// Quick Launch User Defaults
	ud := app.router.Group("/quicklaunch/defaults/user")
	ud.POST("", handlers.AddUserDefaultHandler)
	ud.POST("/", handlers.AddUserDefaultHandler)
	ud.GET("", handlers.GetAllUserDefaultsHandler)
	ud.GET("/", handlers.GetAllUserDefaultsHandler)
	ud.GET("/:id", handlers.GetUserDefaultHandler)
	ud.PATCH("/:id", handlers.UpdateUserDefaultHandler)
	ud.DELETE("/:id", handlers.DeleteUserDefaultHandler)

	// Quick Launch Global Defaults
	gd := app.router.Group("/quicklaunch/defaults/global")
	gd.POST("", handlers.AddGlobalDefaultHandler)
	gd.POST("/", handlers.AddGlobalDefaultHandler)
	gd.GET("", handlers.GetAllGlobalDefaultsHandler)
	gd.GET("/", handlers.GetAllGlobalDefaultsHandler)
	gd.GET("/:id", handlers.GetGlobalDefaultHandler)
	gd.PATCH("/:id", handlers.UpdateGlobalDefaultHandler)
	gd.DELETE("/:id", handlers.DeleteGlobalDefaultHandler)

	// Settings
	settings := app.router.Group("/settings/concurrent-job-limits")
	settings.GET("", handlers.ListConcurrentJobLimitsHandler)
	settings.GET("/", handlers.ListConcurrentJobLimitsHandler)
	settings.GET("/:username", handlers.GetConcurrentJobLimitHandler)
	settings.GET("/:username/", handlers.GetConcurrentJobLimitHandler)
	settings.PUT("/:username", handlers.SetConcurrentJobLimitHandler)
	settings.PUT("/:username/", handlers.SetConcurrentJobLimitHandler)
	settings.DELETE("/:username", handlers.RemoveConcurrentJobLimitHandler)
	settings.DELETE("/:username/", handlers.RemoveConcurrentJobLimitHandler)

	return app
}

// Greeting lets the caller know the service is running.
func (a *AnalysesApp) Greeting(c echo.Context) error {
	return c.String(http.StatusOK, "Hello from analyses.\n")
}
