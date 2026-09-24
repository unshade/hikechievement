package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/unshade/hikechievement/internal/api"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	serveAddr        string
	serveDatabaseURL string
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the HTTP API",
	RunE: func(cmd *cobra.Command, _ []string) error {
		ctx, stop := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		db, err := gorm.Open(postgres.Open(serveDatabaseURL), &gorm.Config{})
		if err != nil {
			return fmt.Errorf("connect to database: %w", err)
		}
		defer func() {
			if sqlDB, err := db.DB(); err == nil {
				_ = sqlDB.Close()
			}
		}()

		srv := &http.Server{
			Addr:              serveAddr,
			Handler:           api.New(db).Handler(),
			ReadHeaderTimeout: 10 * time.Second,
		}

		errCh := make(chan error, 1)
		go func() {
			slog.Info("listening", "addr", serveAddr)
			errCh <- srv.ListenAndServe()
		}()

		select {
		case err := <-errCh:
			return err
		case <-ctx.Done():
		}

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	},
}

func init() {
	serveCmd.Flags().StringVar(&serveAddr, "addr", ":8080", "address to listen on")
	serveCmd.Flags().StringVar(&serveDatabaseURL, "database-url", os.Getenv("DATABASE_URL"), "PostgreSQL connection string (env: DATABASE_URL)")
	rootCmd.AddCommand(serveCmd)
}
