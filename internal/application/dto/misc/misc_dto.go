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

// ClearDataRes reports what was wiped by POST /misc/clear-data.
type ClearDataRes struct {
	TablesCleared []string `json:"tables_cleared"`
	SeqsReset     []string `json:"seqs_reset"`
}
