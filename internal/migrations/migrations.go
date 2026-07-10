package migration

import (
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"

	_ "github.com/go-sql-driver/mysql"
)

func ApplyMigrations(db *sql.DB, dir string) error {
	if err := goose.SetDialect("mysql"); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}

	if err := goose.Up(db, dir); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}