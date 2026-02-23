package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/cyverse-de/analyses/common"
	"github.com/cyverse-de/go-mod/cfg"
	"github.com/cyverse-de/go-mod/logging"
	"github.com/jmoiron/sqlx"
	"github.com/knadh/koanf"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

var log = common.Log

func main() {
	log.Logger.SetReportCaller(true)

	var (
		err error
		c   *koanf.Koanf

		configPath = flag.String("config", cfg.DefaultConfigPath, "Path to the config file")
		dotEnvPath = flag.String("dotenv-path", cfg.DefaultDotEnvPath, "Path to the dotenv file")
		envPrefix  = flag.String("env-prefix", cfg.DefaultEnvPrefix, "The prefix for environment variables")
		listenPort = flag.Int("port", 60000, "The port to listen on")
		logLevel   = flag.String("log-level", "warn", "One of trace, debug, info, warn, error, fatal, or panic.")
	)

	flag.Parse()
	logging.SetupLogging(*logLevel)

	log := log.WithFields(logrus.Fields{"context": "main"})

	log.Infof("Reading config from %s", *configPath)

	c, err = cfg.Init(&cfg.Settings{
		EnvPrefix:   *envPrefix,
		ConfigPath:  *configPath,
		DotEnvPath:  *dotEnvPath,
		StrictMerge: false,
		FileType:    cfg.YAML,
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("Done reading config from %s", *configPath)

	dbURI := c.String("db.uri")
	if dbURI == "" {
		log.Fatal("db.uri must be set in the config file")
	}

	log.Info("Connecting to the database...")
	dbconn, err := sqlx.Connect("postgres", dbURI)
	if err != nil {
		log.Fatalf("failed to connect to database: %s", err)
	}
	defer dbconn.Close() //nolint:errcheck
	log.Info("Successfully connected to the database")

	appsBaseURL := c.String("apps.base")
	if appsBaseURL == "" {
		appsBaseURL = "http://apps"
	}

	dataInfoBaseURL := c.String("data-info.base")
	if dataInfoBaseURL == "" {
		dataInfoBaseURL = "http://data-info"
	}

	log.Infof("Apps service base URL: %s", appsBaseURL)
	log.Infof("Data-info service base URL: %s", dataInfoBaseURL)

	log.Info("Registering routes and initializing handlers...")
	app := NewAnalysesApp(dbconn, appsBaseURL, dataInfoBaseURL)

	addr := fmt.Sprintf(":%s", strconv.Itoa(*listenPort))

	go func() {
		log.Infof("Listening on port %d", *listenPort)
		if err := app.router.Start(addr); err != nil {
			log.Infof("Shutting down the server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Info("Received interrupt signal, shutting down gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := app.router.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
	log.Info("Server stopped")
}
