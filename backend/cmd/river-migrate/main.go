package main

// river-migrate applies River's own schema (river_job, river_leader, ...) to the database.
// Run it once, and again after upgrading River: go run ./cmd/river-migrate

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"

	"pingwise/config/worker"
)

func main() {
	cfg, err := worker.LoadWorkerConfig()
	if err != nil {
		fmt.Println("failed to load config: " + err.Error())
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	dbPool, err := pgxpool.New(ctx, cfg.PrimaryDbDSN)
	if err != nil {
		fmt.Println("failed to create database pool: " + err.Error())
		os.Exit(1)
	}
	defer dbPool.Close()

	migrator, err := rivermigrate.New[pgx.Tx](riverpgxv5.New(dbPool), nil)
	if err != nil {
		fmt.Println("failed to create River migrator: " + err.Error())
		os.Exit(1)
	}
	res, err := migrator.Migrate(ctx, rivermigrate.DirectionUp, nil)
	if err != nil {
		fmt.Println("River migrations failed: " + err.Error())
		os.Exit(1)
	}
	for _, version := range res.Versions {
		fmt.Printf("Applied River migration version %d\n", version.Version)
	}
	fmt.Println("River migrations complete.")
}
