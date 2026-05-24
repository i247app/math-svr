package time

var (
	DefaultFormat       = "2006-01-02T15:04:05.000000Z07:00" // RFC3339 format
	DateOnly            = "2006-01-02"                       // Date only format
	CoreDateTimeFormat  = "2006-01-02T15:04:05Z"             // Core API date-time format (ISO 8601 without fractional seconds)
	DateTimeFormat20FSP = "2006-01-02 15:04:05.000000"       // MySQL DATETIME(6) format
	Format17FSP         = "20060102150405.000"               // Date time input format
	Format20FSP         = "20060102150405.000000"            // Date time input format
)
