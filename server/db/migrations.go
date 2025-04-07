package db

import (
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func Migrate(manager *Manager) {
	driver, err := postgres.WithInstance(manager.Conn, &postgres.Config{})
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать драйвер миграции: %v", err))
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://db//migrations",
		"postgres",
		driver,
	)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать мигратора: %v", err))
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		panic(fmt.Sprintf("Не удалось применить миграции: %v", err))
	}
}
