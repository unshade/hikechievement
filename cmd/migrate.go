package cmd

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib" // registers the "pgx" database/sql driver
	"github.com/spf13/cobra"
	"github.com/unshade/hikechievement/migrations"
)

var migrateDatabaseURL string

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Apply pending database migrations",
	RunE: func(cmd *cobra.Command, _ []string) error {
		db, err := sql.Open("pgx", migrateDatabaseURL)
		if err != nil {
			return fmt.Errorf("connect to database: %w", err)
		}
		defer db.Close()

		if err := migrations.Up(cmd.Context(), db); err != nil {
			return fmt.Errorf("apply migrations: %w", err)
		}
		fmt.Println("migrations applied")
		return nil
	},
}

func init() {
	migrateCmd.Flags().StringVar(&migrateDatabaseURL, "database-url", os.Getenv("DATABASE_URL"), "PostgreSQL connection string (env: DATABASE_URL)")
	rootCmd.AddCommand(migrateCmd)
}
