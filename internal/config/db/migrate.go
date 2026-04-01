package db

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	migratepostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(db *sql.DB, migrationsPath string) error {
	driver, err := migratepostgres.WithInstance(db, &migratepostgres.Config{})
	if err != nil {
		return fmt.Errorf("create driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath,
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("create migrations: %w", err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("run migrations: %w", err)
	}
	return nil

}
