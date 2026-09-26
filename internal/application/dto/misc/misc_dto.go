package misc

import "time"

type LogTimeFormatReq struct {
	Time string `json:"time"`
}

type LogTimeFormatRes struct {
	DefaultFormat       string `json:"default_format"`
	DateOnlyFormat      string `json:"date_only_format"`
	CoreDateTimeFormat  string `json:"core_date_time_format"`
	DateTimeFormat20FSP string `json:"date_time_format_20_fsp"`
	Format17FSP         string `json:"format_17_fsp"`
	Format20FSP         string `json:"format_20_fsp"`
}

// ClearDataReq is the (optional) body of POST /misc/clear-data. When
// whitelist_id is empty the endpoint wipes everything (the original behaviour);
// when it carries uids, those users and all their data are preserved.
type ClearDataReq struct {
	WhitelistId []int64 `json:"whitelist_id"`
}

// ClearDataRes reports what was wiped by POST /misc/clear-data and
// POST /misc/clear-data-tables.
type ClearDataRes struct {
	TablesCleared []string `json:"tables_cleared"`
	SeqsReset     []string `json:"seqs_reset"`
	// KeptUids echoes the whitelisted users preserved (whitelist path only).
	KeptUids []int64 `json:"kept_uids,omitempty"`
}

// DBPoolStatsRes is a point-in-time snapshot of the MySQL connection pool
// (sql.DB.Stats). Gauges (open/in_use/idle) are current values; the wait and
// *_closed fields are cumulative since process start — diff two snapshots
// taken captured_at apart to get a rate.
type DBPoolStatsRes struct {
	CapturedAt time.Time `json:"captured_at"`

	MaxOpenConnections int `json:"max_open_connections"` // 0 = unlimited
	OpenConnections    int `json:"open_connections"`     // in_use + idle
	InUse              int `json:"in_use"`
	Idle               int `json:"idle"`

	WaitCount          int64   `json:"wait_count"`           // cumulative waits for a free connection
	WaitDurationMs     float64 `json:"wait_duration_ms"`     // cumulative time spent waiting
	AvgWaitDurationMs  float64 `json:"avg_wait_duration_ms"` // wait_duration_ms / wait_count
	MaxIdleClosed      int64   `json:"max_idle_closed"`      // closed by SetMaxIdleConns
	MaxIdleTimeClosed  int64   `json:"max_idle_time_closed"` // closed by SetConnMaxIdleTime
	MaxLifetimeClosed  int64   `json:"max_lifetime_closed"`  // closed by SetConnMaxLifetime
	UtilizationPercent float64 `json:"utilization_percent"`  // in_use / max_open * 100 (0 when unlimited)
}

// ClearDataTablesReq selects which tables POST /misc/clear-data-tables wipes.
// Each name must be one of the clear-data allow-listed tables (e.g. "ma_users").
type ClearDataTablesReq struct {
	Tables []string `json:"tables"`
}
