//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package migrate

import (
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Config tunes golang-migrate for a schema-per-service Postgres database.
type Config struct {
	// SchemaName is the Postgres schema that owns schema_migrations (e.g. user, payment).
	SchemaName string
	// MigrationsTable defaults to schema_migrations when empty.
	MigrationsTable string
}

// Up applies pending migrations from migrationsDir.
func Up(databaseURL, migrationsDir string, cfg Config) error {
	return run(databaseURL, migrationsDir, cfg, (*migrate.Migrate).Up)
}

// Down rolls back the most recent migration in migrationsDir.
func Down(databaseURL, migrationsDir string, cfg Config) error {
	return run(databaseURL, migrationsDir, cfg, (*migrate.Migrate).Down)
}

type migrateFn func(*migrate.Migrate) error

func run(databaseURL, migrationsDir string, cfg Config, fn migrateFn) error {
	absDir, err := filepath.Abs(migrationsDir)
	if err != nil {
		return fmt.Errorf("resolve migrations dir: %w", err)
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	table := cfg.MigrationsTable
	if table == "" {
		table = "schema_migrations"
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{
		SchemaName:      cfg.SchemaName,
		MigrationsTable: table,
	})
	if err != nil {
		return fmt.Errorf("migrate driver: %w", err)
	}

	sourceURL := "file://" + filepath.ToSlash(absDir)
	m, err := migrate.NewWithDatabaseInstance(sourceURL, "postgres", driver)
	if err != nil {
		return fmt.Errorf("migrate instance: %w", err)
	}

	runErr := fn(m)
	srcErr, dbErr := m.Close()
	if runErr != nil && !errors.Is(runErr, migrate.ErrNoChange) {
		return runErr
	}
	return errors.Join(
		wrapCloseErr("close migrate source", srcErr),
		wrapCloseErr("close migrate database", dbErr),
	)
}

func wrapCloseErr(msg string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", msg, err)
}