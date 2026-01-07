package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/golang-migrate/migrate"
	_ "github.com/golang-migrate/migrate/database/sqlite3"
	_ "github.com/golang-migrate/migrate/source/file"
)

func main() {
	var storagePath, migrationPath string

	flag.StringVar(&storagePath, "storage-path", "", "path to store the migration files")
	flag.StringVar(&migrationPath, "migrations-path", "", "path to store the migration files")
	flag.Parse()

	if storagePath == "" || migrationPath == "" {
		panic("storagePath, migrationPath is required")
	}

	m, err := migrate.New(
		"file://"+migrationPath,
		fmt.Sprintf("sqlite3://%s", storagePath),
	)
	if err != nil {
		panic(err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no migrations to apply")

			return
		}

		panic(err)
	}

	fmt.Println("applied migrations")
}
