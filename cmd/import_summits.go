package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/unshade/hikechievement/internal/importer"
	"github.com/unshade/hikechievement/internal/repository"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	importFile        string
	importDatabaseURL string
	importBatchSize   int
)

var importSummitsCmd = &cobra.Command{
	Use:   "import-summits",
	Short: "Import summits (OpenStreetMap natural=peak nodes) from a .osm.pbf file",
	RunE: func(cmd *cobra.Command, _ []string) error {
		f, err := os.Open(importFile)
		if err != nil {
			return fmt.Errorf("open %s: %w", importFile, err)
		}
		defer f.Close()

		db, err := gorm.Open(postgres.Open(importDatabaseURL), &gorm.Config{})
		if err != nil {
			return fmt.Errorf("connect to database: %w", err)
		}

		repo := repository.NewSummitsRepository(db)
		n, err := importer.Import(cmd.Context(), f, repo, importBatchSize)
		if err != nil {
			return fmt.Errorf("import summits (%d imported before the error): %w", n, err)
		}
		fmt.Printf("imported %d summits\n", n)
		return nil
	},
}

func init() {
	importSummitsCmd.Flags().StringVar(&importFile, "file", "", "path to a .osm.pbf extract (required)")
	importSummitsCmd.Flags().StringVar(&importDatabaseURL, "database-url", os.Getenv("DATABASE_URL"), "PostgreSQL connection string (env: DATABASE_URL)")
	importSummitsCmd.Flags().IntVar(&importBatchSize, "batch-size", 5000, "number of summits per upsert batch")
	_ = importSummitsCmd.MarkFlagRequired("file")
	rootCmd.AddCommand(importSummitsCmd)
}
