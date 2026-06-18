package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/miguelnietoa/stellar-explorer/tui-indexer/internal/store"
	"github.com/miguelnietoa/stellar-explorer/tui-indexer/migrations"
)

func runMigrate(databaseURL string) {
	db, err := store.NewPostgresStore(databaseURL)
	if err != nil {
		log.Fatalf("migrate: connect to database: %v", err)
	}
	defer db.Close()

	source, err := iofs.New(migrations.FS, ".")
	if err != nil {
		log.Fatalf("migrate: create iofs source: %v", err)
	}

	driver, err := postgres.WithInstance(db.DB(), &postgres.Config{})
	if err != nil {
		log.Fatalf("migrate: create postgres driver: %v", err)
	}

	m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		log.Fatalf("migrate: init: %v", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("migrate: run: %v", err)
	}

	version, dirty, err := m.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		log.Fatalf("migrate: get version: %v", err)
	}

	fmt.Printf("✅ Migrations applied. Version: %d, Dirty: %v\n", version, dirty)
}
