package database

import (
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"

	"math-ai.com/math-ai/internal/infrastructure/logger"

	"math-ai.com/math-ai/internal/infrastructure/config"
)

// NewSqlDB creates and returns a production-ready *sql.DB connection.
func NewSqlDB(dbConfig config.DBConfig) (*sql.DB, error) {
	db, err := NewDatabaseConnect(dbConfig)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// NewDatabaseConnect establishes a MySQL connection with pooling and timeouts.
func NewDatabaseConnect(dbConfig config.DBConfig) (*sql.DB, error) {
	address := fmt.Sprintf("%s:%s", dbConfig.DBHost, dbConfig.DBPort)

	dsn := mysql.Config{
		User:                 dbConfig.DBUser,
		Passwd:               dbConfig.DBPass,
		Net:                  "tcp",
		Addr:                 address,
		DBName:               dbConfig.DBName,
		ParseTime:            true,
		Collation:            "utf8mb4_0900_ai_ci",
		Loc:                  time.UTC,
		MaxAllowedPacket:     4 << 20, // 4MB
		AllowNativePasswords: true,
		InterpolateParams:    true,
		TLS:                  &tls.Config{MinVersion: tls.VersionTLS12, MaxVersion: tls.VersionTLS12},
	}

	db, err := sql.Open("mysql", dsn.FormatDSN())
	if err != nil {
		return nil, fmt.Errorf("database: failed to open connection: %w", err)
	}

	// Apply pool settings with defaults
	maxOpen := dbConfig.MaxOpenConns
	if maxOpen <= 0 {
		maxOpen = 25
	}
	maxIdle := dbConfig.MaxIdleConns
	if maxIdle <= 0 {
		maxIdle = 10
	}
	connMaxLifetime := dbConfig.ConnMaxLifetime
	if connMaxLifetime <= 0 {
		connMaxLifetime = 5 * time.Minute
	}
	connMaxIdleTime := dbConfig.ConnMaxIdleTime
	if connMaxIdleTime <= 0 {
		connMaxIdleTime = 3 * time.Minute
	}

	db.SetMaxOpenConns(maxOpen)
	db.SetMaxIdleConns(maxIdle)
	db.SetConnMaxLifetime(connMaxLifetime)
	db.SetConnMaxIdleTime(connMaxIdleTime)

	// Verify connectivity
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("database: failed to ping: %w", err)
	}

	logger.From(ctx).Info("[DB] Connected to MySQL successfully",
		"host", address, "db", dbConfig.DBName,
	)

	return db, nil
}
