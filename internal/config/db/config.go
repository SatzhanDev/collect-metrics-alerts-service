// Package db предоставляет утилиты для подключения к базе данных и миграций.
package db

import "time"

type Config struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}
