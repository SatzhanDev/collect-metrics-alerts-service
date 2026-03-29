package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

func NewPostgres(cfg Config) (*sql.DB, error) {
	if cfg.DSN == "" {
		return nil, fmt.Errorf("database dsn is empty")
	}

	db, err := sql.Open("postgres", cfg.DSN)
	if err != nil {
		return nil, err
	}

	if cfg.MaxOpenConns > 0 {
		db.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}
