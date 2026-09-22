package misc

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

// ClearDataTablesReq selects which tables POST /misc/clear-data-tables wipes.
// Each name must be one of the clear-data allow-listed tables (e.g. "ma_users").
type ClearDataTablesReq struct {
	Tables []string `json:"tables"`
}
