package database

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"
)

// DatabaseURL builds a postgres:// connection URL from PostgresConfig.
func (c PostgresConfig) DatabaseURL() string {
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(c.User, c.Password),
		Host:     fmt.Sprintf("%s:%d", c.Host, c.Port),
		Path:     c.DBName,
		RawQuery: fmt.Sprintf("sslmode=%s", c.SSLMode),
	}
	return u.String()
}

// RunMigrations applies all pending up migrations from the given path.
func RunMigrations(dbURL, migrationsPath string, logger *zap.Logger) error {
	sourceURL := fmt.Sprintf("file://%s", migrationsPath)
	logger.Info("running database migrations", zap.String("source", sourceURL))

	m, err := migrate.New(sourceURL, dbURL)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer func() {
		srcErr, dbErr := m.Close()
		if srcErr != nil {
			logger.Warn("failed to close migration source", zap.Error(srcErr))
		}
		if dbErr != nil {
			logger.Warn("failed to close migration database", zap.Error(dbErr))
		}
	}()

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			logger.Info("database already up to date")
			return nil
		}
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	version, dirty, _ := m.Version()
	logger.Info("migrations applied successfully",
		zap.Uint("version", version),
		zap.Bool("dirty", dirty),
	)
	return nil
}
