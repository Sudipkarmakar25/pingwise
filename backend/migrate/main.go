package main

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/amacneil/dbmate/v2/pkg/dbmate"
	_ "github.com/amacneil/dbmate/v2/pkg/driver/postgres"

	"pingwise/config/api"
)

// one folder per table, each file holding "-- migrate:up" and "-- migrate:down"
const migrationsPath = "migrate/migrations/*"

// usage (from the project root): go run ./migrate up | down | status
func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: go run ./migrate up | down | status")
		os.Exit(1)
	}

	cfg, err := api.LoadApiConfig()
	if err != nil {
		fmt.Println("failed to load config: " + err.Error())
		os.Exit(1)
	}
	dbUrl, err := url.Parse(cfg.PrimaryDbDSN)
	if err != nil {
		fmt.Println("invalid database dsn: " + err.Error())
		os.Exit(1)
	}
	migrationDirs, err := filepath.Glob(migrationsPath)
	if err != nil || len(migrationDirs) == 0 {
		fmt.Println("no migrations found in " + migrationsPath)
		os.Exit(1)
	}

	db := dbmate.New(dbUrl)
	db.MigrationsDir = migrationDirs
	db.AutoDumpSchema = false

	switch os.Args[1] {
	case "up":
		err = db.Migrate()
	case "down":
		// rolls back only the latest migration
		err = db.Rollback()
	case "status":
		_, err = db.Status(false)
	default:
		err = fmt.Errorf("unknown command %q", os.Args[1])
	}
	if err != nil {
		fmt.Println("migration failed: " + err.Error())
		os.Exit(1)
	}
}
