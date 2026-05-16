package config

import "time"

type Env struct {
	ServerHost string
	ServerPort string
	LogFile    string

	DBConfig      DBConfig
	EmailConfig   EmailConfig
	StorageConfig StorageConfig
	SMSConfig     SMSConfig
}

// Config holds the database connection configuration.
type DBConfig struct {
	DBUser   string
	DBPasswd string
	DBHost   string
	DBPort   string
	DBName   string

	// Connection pool settings (optional, sensible defaults applied)
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

type EmailConfig struct {
	EmailProvider string

	// gomail
	EmailHost     string
	EmailPort     int
	EmailUser     string
	EmailPassword string
	EmailFrom     string

	// google
	GmailCredentialsFile string
	GmailSenderEmail     string
}

type StorageConfig struct {
	Provider string // env STORAGE_PROVIDER, default "s3"

	AccessKey string // env STORAGE_ACCESS_KEY
	SecretKey string // env STORAGE_SECRET_KEY
	Region    string // env STORAGE_REGION
	Bucket    string // env STORAGE_BUCKET
}

// SMSConfig configures the SMS adapter (`internal/adapter/sms`). The
// factory translates these into provider-specific configs — today the
// only supported provider is "twilio", consumed by libs/twilio.
//
// TwilioAuthToken is a long-lived secret. It is never logged. The
// libs/twilio package masks it out of every code path; see the
// package doc in libs/twilio/config.go.
//
// Either TwilioFrom or TwilioMessagingServiceSID may be set, both may
// be set, or neither — in the last case every Send call must supply
// Message.From or Twilio responds 21603.
type SMSConfig struct {
	SMSProvider string // env SMS_PROVIDER; "twilio" | "" | "disabled"

	TwilioAccountSID          string // env TWILIO_ACCOUNT_SID
	TwilioAuthToken           string // env TWILIO_AUTH_TOKEN — SECRET
	TwilioBaseURL             string // env TWILIO_BASE_URL; default "https://api.twilio.com"
	TwilioFrom                string // env TWILIO_FROM; optional
	TwilioMessagingServiceSID string // env TWILIO_MESSAGING_SERVICE_SID; optional

	Timeout       time.Duration // env SMS_TIMEOUT,     e.g. "10s"
	MaxRetries    int           // env SMS_MAX_RETRIES, default 2
	RetryDelay    time.Duration // env SMS_RETRY_DELAY, e.g. "250ms"
	RequireAtBoot bool          // env SMS_REQUIRE_AT_BOOT, default false
}
