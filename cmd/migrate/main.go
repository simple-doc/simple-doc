package main

import (
	"log/slog"
	"os"

	"docgen"
	"docgen/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

func main() {
	config.InitLogging()

	migrationsFS := docgen.ResolveFS(config.MigrationsDir(), docgen.EmbeddedMigrations())
	d, err := iofs.New(migrationsFS, ".")
	if err != nil {
		slog.Error("failed to create migration source", "error", err)
		os.Exit(1)
	}

	connStr := config.PostgreSQLConnString()
	m, err := migrate.NewWithSourceInstance("iofs", d, "pgx5://"+connStr[len("postgres://"):]+"&x-migrations-table=simpledoc_version")
	if err != nil {
		slog.Error("failed to initialize migrations", "error", err)
		os.Exit(1)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	slog.Info("migrations applied")
}
