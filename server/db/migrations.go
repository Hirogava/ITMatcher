package db

import (
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func Migrate(db *DBManager) error {
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
    if err != nil {
        return fmt.Errorf("ошибка создания драйвера миграции: %v", err)
    }

    m, err := migrate.NewWithDatabaseInstance(
        "file://db",
        "postgres",
        driver,
    )
    if err != nil {
        return fmt.Errorf("ошибка при создании мигратора: %v", err)
    }

    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        return fmt.Errorf("ошибка при применении миграций: %v", err)
    }

    log.Println("Миграции успешно применены!")
    return nil
}