package migrations

import (
	"context"
	"database/sql"
	"embed"

	"github.com/pressly/goose/v3"
)

//go:embed *.sql
var embedMigrations embed.FS

// RunGoose runs goose migrations with the provided arguments against the database.
// Args should be goose commands like []string{"up", "-v"} or []string{"down"}
func RunGoose(ctx context.Context, args []string, db *sql.DB) error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("clickhouse"); err != nil {
		return err
	}

	// Run goose command
	if len(args) == 0 {
		args = []string{"up"}
	}

	command := args[0]
	arguments := args[1:]

	return goose.RunContext(ctx, command, db, ".", arguments...)
}

