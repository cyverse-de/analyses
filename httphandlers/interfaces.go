package httphandlers

import (
	"github.com/cyverse-de/analyses/clients"
	"github.com/cyverse-de/analyses/db"
)

// DatabaseStore is the canonical interface for database operations,
// defined in the db package near its implementation.
type DatabaseStore = db.Store

// AppFetcher is the canonical interface for retrieving app definitions,
// defined in the clients package near its implementation.
type AppFetcher = clients.AppFetcher

// PathChecker is the canonical interface for verifying path accessibility,
// defined in the clients package near its implementation.
type PathChecker = clients.PathChecker
