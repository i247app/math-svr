package db

import (
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
	"math-ai.com/math-ai/internal/shared/config"
	"math-ai.com/math-ai/internal/shared/logger"
)

const (
	DatabaseTimeout = time.Second * 5
)

// Database wraps a sql.DB connection.
type Database struct {
	db *sql.DB
}

// NewDatabase initializes a new database connection.
func NewDatabase(config *config.DBConfig) (*Database, error) {
	address := fmt.Sprintf("%s:%s", config.Host, config.Port)

	mysqlCfg := mysql.NewConfig()
	mysqlCfg.User = config.User
	mysqlCfg.Passwd = config.Password
	mysqlCfg.Addr = address
	mysqlCfg.DBName = config.Name
	mysqlCfg.AllowNativePasswords = true
	mysqlCfg.Net = "tcp"
	mysqlCfg.TLS = &tls.Config{MinVersion: tls.VersionTLS12, MaxVersion: tls.VersionTLS12}
	mysqlCfg.ParseTime = true

	db, err := sql.Open("mysql", mysqlCfg.FormatDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the database: %v", err)
	}

	// Set up connection pool
	db.SetMaxOpenConns(200)
	db.SetMaxIdleConns(20)
	db.SetConnMaxLifetime(time.Minute * 2)

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping the database: %v", err)
	}

	logger.Info("Connected to database successfully")
	return &Database{db: db}, nil
}

func (d *Database) GetDB() *sql.DB {
	return d.db
}

func (d *Database) WithTransaction(function func(tx *sql.Tx) error) error {
	logger.Info("Starting transaction")
	tx, err := d.db.Begin()
	if err != nil {
		logger.Errorf("Failed to begin transaction: %v", err)
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	if err := function(tx); err != nil {
		logger.Info("Rolling back transaction due to error: %v", err)
		if rbErr := tx.Rollback(); rbErr != nil {
			logger.Errorf("Failed to rollback transaction: %v", rbErr)
			return fmt.Errorf("failed to rollback transaction: %v (original error: %v)", rbErr, err)
		}
		logger.Info("Transaction rolled back successfully")
		return err
	}

	logger.Info("Committing transaction")
	if err := tx.Commit(); err != nil {
		logger.Errorf("Failed to commit transaction: %v", err)
		return fmt.Errorf("failed to commit transaction: %v", err)
	}
	logger.Info("Transaction committed successfully")
	return nil
}

func (d *Database) Query(ctx context.Context, tx *sql.Tx, query string, args ...any) (*sql.Rows, error) {
	_, cancel := context.WithTimeout(ctx, DatabaseTimeout)
	defer cancel()

	d.logInputSQL(query, args...)
	var rows *sql.Rows
	var err error
	if tx != nil {
		rows, err = tx.QueryContext(ctx, query, args...)
	} else {
		rows, err = d.db.QueryContext(ctx, query, args...)
	}
	if err != nil {
		logger.Error(err)
	}
	logger.Info(rows)
	return rows, err
}

func (d *Database) QueryRow(ctx context.Context, tx *sql.Tx, query string, args ...any) *sql.Row {
	_, cancel := context.WithTimeout(ctx, DatabaseTimeout)
	defer cancel()

	d.logInputSQL(query, args...)
	if tx != nil {
		return tx.QueryRowContext(ctx, query, args...)
	}
	return d.db.QueryRowContext(ctx, query, args...)
}

func (d *Database) Exec(ctx context.Context, tx *sql.Tx, query string, args ...any) (sql.Result, error) {
	_, cancel := context.WithTimeout(ctx, DatabaseTimeout)
	defer cancel()

	d.logInputSQL(query, args...)
	var result sql.Result
	var err error
	if tx != nil {
		result, err = tx.ExecContext(ctx, query, args...)
	} else {
		result, err = d.db.ExecContext(ctx, query, args...)
	}
	if err != nil {
		logger.Error(err)
	}
	if result != nil {
	}
	return result, err
}

func (d *Database) PingContext(ctx context.Context) error {
	_, cancel := context.WithTimeout(ctx, DatabaseTimeout)
	defer cancel()

	d.logInputSQL("PING")
	return d.db.PingContext(ctx)
}

func (d *Database) Close() error {
	return d.db.Close()
}
