package repositories

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"math-ai.com/math-ai/internal/infrastructure/database"
)

// Integration tests in this package run against a real MySQL. Mocking the
// database here would defeat the purpose: the things most likely to be wrong
// are column names, scan order, and the concurrency semantics of the actual
// engine — none of which a fake reproduces.
//
// They are opt-in via an environment variable, so `go test ./...` stays green
// on a machine with no database:
//
//	MATHSVR_TEST_DSN='user:pass@tcp(127.0.0.1:3306)/mathai?parseTime=true&loc=UTC&collation=utf8mb4_0900_ai_ci' \
//	    go test ./internal/infrastructure/persistence/mysql/repositories/ -run Integration -v
//
// parseTime=true is REQUIRED — without it the driver hands DATETIME back as
// []byte and every scan into time.Time fails.
const testDSNEnv = "MATHSVR_TEST_DSN"

// testIDBase keeps fixtures far away from real rows. Every id these tests mint
// is testIDBase+n, so a failed cleanup leaves obviously-synthetic data that is
// trivial to find and delete by hand rather than something indistinguishable
// from a real conversation.
const testIDBase int64 = 990000000000

// openTestDB connects, or skips the test when no DSN is configured.
func openTestDB(t *testing.T) database.SqlHandler {
	t.Helper()

	dsn := os.Getenv(testDSNEnv)
	if dsn == "" {
		t.Skipf("%s not set — skipping MySQL integration test", testDSNEnv)
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping test db (%s): %v", testDSNEnv, err)
	}

	// Fail loudly rather than mysteriously if the chat migrations were not
	// applied to the database the DSN points at.
	requireTables(t, db, "ma_chat_conversations", "ma_chat_participants",
		"ma_chat_messages", "ma_user_presence", "ma_seqs")

	return database.NewDatabaseWithLogs(db)
}

func requireTables(t *testing.T, db *sql.DB, tables ...string) {
	t.Helper()
	for _, name := range tables {
		var found string
		err := db.QueryRow("SHOW TABLES LIKE ?", name).Scan(&found)
		if err == sql.ErrNoRows {
			t.Fatalf("table %s is missing — apply migrations 023-027 to the test database first", name)
		}
		if err != nil {
			t.Fatalf("checking table %s: %v", name, err)
		}
	}
}

// inTx runs fn inside a committed transaction, which is what NextSeqNo and the
// command handlers require: the increment and the read-back must share one
// connection.
func inTx(t *testing.T, db database.SqlHandler, fn func(ctx context.Context, ex database.Executor) error) {
	t.Helper()
	ctx := context.Background()
	if err := database.WithTransaction(ctx, db, func(txCtx context.Context, tx *sql.Tx) error {
		return fn(txCtx, database.NewTxWithLogs(txCtx, tx))
	}); err != nil {
		t.Fatalf("transaction: %v", err)
	}
}

// cleanupChatRows removes every fixture row for a conversation. Registered
// with t.Cleanup so it runs even when the test fails, keeping a developer
// database from silently filling with test threads.
func cleanupChatRows(t *testing.T, db database.SqlHandler, conversationIDs ...int64) {
	t.Helper()
	t.Cleanup(func() {
		ctx := context.Background()
		for _, id := range conversationIDs {
			for _, stmt := range []string{
				"DELETE FROM ma_chat_messages WHERE conversation_id = ?",
				"DELETE FROM ma_chat_participants WHERE conversation_id = ?",
				"DELETE FROM ma_chat_conversations WHERE conversation_id = ?",
			} {
				if _, err := db.Exec(ctx, stmt, id); err != nil {
					t.Logf("cleanup %q for conversation %d: %v", stmt, id, err)
				}
			}
		}
	})
}

func cleanupPresenceRows(t *testing.T, db database.SqlHandler, userIDs ...int64) {
	t.Helper()
	t.Cleanup(func() {
		ctx := context.Background()
		for _, id := range userIDs {
			if _, err := db.Exec(ctx, "DELETE FROM ma_user_presence WHERE user_id = ?", id); err != nil {
				t.Logf("cleanup presence for user %d: %v", id, err)
			}
		}
	})
}

// uniqueTestID derives a fixture id that is stable within one test run but
// distinct across parallel runs, so two developers testing at once against a
// shared database do not collide on the UNIQUE keys.
func uniqueTestID(offset int64) int64 {
	return testIDBase + (time.Now().UnixNano()/1e6)%1e8*100 + offset
}

// runInTx is inTx's error-returning twin, for tests that assert on the error
// a transaction produces rather than failing on it.
func runInTx(db database.SqlHandler, fn func(ctx context.Context, ex database.Executor) error) error {
	ctx := context.Background()
	return database.WithTransaction(ctx, db, func(txCtx context.Context, tx *sql.Tx) error {
		return fn(txCtx, database.NewTxWithLogs(txCtx, tx))
	})
}

// runOrFail runs fn in a transaction and fails the test on error.
func runOrFail(t *testing.T, db database.SqlHandler, fn func(ctx context.Context, ex database.Executor) error) {
	t.Helper()
	if err := runInTx(db, fn); err != nil {
		t.Fatalf("transaction: %v", err)
	}
}
