// Package gormdb build connection to db with gorm and ddtrace
package gormdb

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"go.uber.org/multierr"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	customLogger "github.com/Haevnen/audit-logging-api/pkg/logger"
)

const (
	defaultPostgresPort = 5432
)

// Config the least enough config for gorm
type Config struct {
	DBUser              string
	DBPass              string
	DBPort              int
	DBHost              string
	DBName              string
	DBCollation         string
	DBInterpolateParams bool
	DBLocation          string
	DBTimezone          string
	DBSSLMode           string

	LogSQL      bool
	NotifyQuery string

	MaxOpenConn       int
	MaxIdleConn       int
	MaxLifetimeSecond int
}

// New build gorm DB
func New(cfg *Config) (db *gorm.DB, close func() error, err error) {
	// Open raw *sql.DB
	conn, err := newDBConn(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("new db conn: %w", err)
	}

	// Set connection pool settings
	conn.SetMaxOpenConns(cfg.MaxOpenConn)
	conn.SetMaxIdleConns(cfg.MaxIdleConn)
	conn.SetConnMaxLifetime(time.Duration(cfg.MaxLifetimeSecond) * time.Second)

	logfile := customLogger.GetLogger()
	// Open GORM with Postgres driver
	gormDB, err := gorm.Open(
		postgres.New(postgres.Config{Conn: conn}),
		&gorm.Config{
			PrepareStmt: true, // cache prepared statements
			Logger:      NewGormLogger(logger.Info, logfile),
		},
	)
	if err != nil {
		if closeErr := conn.Close(); closeErr != nil {
			err = multierr.Append(err, closeErr)
		}
		return nil, nil, err
	}

	return gormDB, conn.Close, nil
}

func newDBConn(cfg *Config) (*sql.DB, error) {
	dsn, err := BuildPostgresConnectionString(cfg)
	if err != nil {
		return nil, err
	}
	return sql.Open("pgx", dsn)
}

// BuildPostgresConnectionString build postgres connection string
func BuildPostgresConnectionString(c *Config) (string, error) {
	if c.DBUser == "" {
		return "", errors.New("db user is not set")
	}
	if c.DBName == "" {
		return "", errors.New("db name is not set")
	}

	port := defaultPostgresPort
	if c.DBPort != 0 {
		port = c.DBPort
	}

	sslMode := "disable"
	if c.DBSSLMode != "" {
		sslMode = c.DBSSLMode
	}

	// Construct connection string (DSN) for pgx/gorm
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, port, c.DBUser, c.DBPass, c.DBName, sslMode,
	)

	if c.DBTimezone != "" {
		dsn = fmt.Sprintf("%s TimeZone=%s", dsn, c.DBTimezone)
	}

	return dsn, nil
}
