package config

import "time"

type Env struct {
	ServerHost string
	ServerPort string
	LogFile    string
	// Log encodings per destination. Default split: console=text (readable),
	// file=json (shipped to Loki). LogFormat, when set, overrides BOTH (a
	// back-compat uniform mode).
	LogFormat        string // env LOG_FORMAT;         "" | "text" | "json" (override both)
	LogConsoleFormat string // env LOG_CONSOLE_FORMAT; default "text"
	LogFileFormat    string // env LOG_FILE_FORMAT;    default "json"

	SerializedSessionFile string
	GexSessionDriver      string
	SharedKeyBytes        []byte
	HttpsCertFile         *string
	HttpsKeyFile          *string
	EnableOTP             bool
	TrustDeviceTTLDays    int

	DBConfig            DBConfig
	EmailConfig         EmailConfig
	StorageConfig       StorageConfig
	SMSConfig           SMSConfig
	BotConfig           BotConfig
	NotificationConfig  NotificationConfig
	ObservabilityConfig ObservabilityConfig
}

// ObservabilityConfig tunes metrics + distributed tracing. Both pillars are
// independently switchable so a deploy can run metrics-only (the safe local
// default) or add tracing when a Tempo/OTLP endpoint is available.
//
// Behaviour:
//
//	MetricsEnabled false → no /metrics route, metrics middleware is a no-op.
//	TracingEnabled false → no TracerProvider is installed; every span is a
//	                       no-op (safe) and nothing is exported.
type ObservabilityConfig struct {
	ServiceName    string // env OBS_SERVICE_NAME;    default "math-svr"
	ServiceVersion string // env OBS_SERVICE_VERSION; default "dev"
	Environment    string // env OBS_ENV;             default "local"

	MetricsEnabled bool   // env OBS_METRICS_ENABLED; default true
	MetricsAddr    string // env OBS_METRICS_ADDR;    default ":9091" (separate listener, no app middleware)

	TracingEnabled   bool    // env OBS_TRACING_ENABLED;    default false
	OTLPEndpoint     string  // env OBS_OTLP_ENDPOINT;      default "localhost:4318"
	OTLPInsecure     bool    // env OBS_OTLP_INSECURE;      default true (local: no TLS)
	TraceSampleRatio float64 // env OBS_TRACE_SAMPLE_RATIO; default 1.0 (sample everything)
}

// Config holds the database connection configuration.
type DBConfig struct {
	DBUser string
	DBPass string
	DBHost string
	DBPort string
	DBName string

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

// BotConfig configures the AI bot adapter (`internal/adapter/bot`). The
// factory translates these into provider-specific configs. Today the only
// supported provider is "langchain", which itself can be backed by one of
// several LLM vendors (googleai, openai, anthropic, ollama) selected by
// LangChainBackend.
//
// LangChainAPIKey is a long-lived secret. It is never logged. The
// libs/langchain package masks it out of every code path.
//
// Behaviour matrix for BotProvider:
//
//	""  or "disabled"    → adapter is nil; module services must nil-guard.
//	"langchain" | "eino" → every framework whose backend key is set gets
//	                       registered; BotProvider names the DEFAULT one.
//	anything else        → MathError(BOT_CONFIG_INVALID) at boot.
type BotConfig struct {
	BotProvider string // env BOT_PROVIDER; "langchain" | "eino" | "" | "disabled"

	// LangChain-backed provider settings. Consumed (and the provider
	// registered) whenever LangChainBackend is non-empty.
	LangChainBackend     string  // env BOT_LANGCHAIN_BACKEND; "googleai" | "openai" | "anthropic" | "ollama"
	LangChainAPIKey      string  // env BOT_LANGCHAIN_API_KEY — SECRET
	LangChainBaseURL     string  // env BOT_LANGCHAIN_BASE_URL; optional vendor override / required for ollama
	LangChainModel       string  // env BOT_LANGCHAIN_MODEL; e.g. "gemini-1.5-flash"
	LangChainEmbedModel  string  // env BOT_LANGCHAIN_EMBED_MODEL; e.g. "text-embedding-004"
	LangChainTemperature float64 // env BOT_LANGCHAIN_TEMPERATURE; <0 means vendor default
	LangChainTopP        float64 // env BOT_LANGCHAIN_TOP_P;        <0 means vendor default
	LangChainMaxTokens   int     // env BOT_LANGCHAIN_MAX_TOKENS;   0  means vendor default

	// Eino-backed provider settings (cloudwego/eino + eino-ext). Consumed
	// (and the provider registered) whenever EinoBackend is non-empty.
	// No embed model: the eino provider reports Embed as unsupported.
	EinoBackend     string  // env BOT_EINO_BACKEND; "googleai" | "openai" | "anthropic" | "ollama"
	EinoAPIKey      string  // env BOT_EINO_API_KEY — SECRET
	EinoBaseURL     string  // env BOT_EINO_BASE_URL; optional vendor override / required for ollama
	EinoModel       string  // env BOT_EINO_MODEL; e.g. "gemini-1.5-flash"
	EinoTemperature float64 // env BOT_EINO_TEMPERATURE; <0 means vendor default
	EinoTopP        float64 // env BOT_EINO_TOP_P;        <0 means vendor default
	EinoMaxTokens   int     // env BOT_EINO_MAX_TOKENS;   0  means vendor default

	// Transport tuning shared by BOTH providers.
	Timeout       time.Duration // env BOT_TIMEOUT,        e.g. "60s"
	MaxRetries    int           // env BOT_MAX_RETRIES,    default 2
	RetryDelay    time.Duration // env BOT_RETRY_DELAY,    e.g. "500ms"
	RequireAtBoot bool          // env BOT_REQUIRE_AT_BOOT, default false
}


// NotificationConfig configures the push-notification adapter
// (`internal/adapter/notification`). Today the only provider is Firebase
// Cloud Messaging.
//
// Behaviour matrix for Provider:
//
//	""  or "disabled" → adapter is nil; module services must nil-guard.
//	"firebase"        → Firebase provider, registered + defaulted.
//	anything else     → MathError(NOTIFICATION_CONFIG_INVALID) at boot.
//
// FirebaseCredentialsFile is a path to a service-account JSON (kept under
// keys/ — never logged, never committed to images).
type NotificationConfig struct {
	Provider string // env NOTIFICATION_PROVIDER; "firebase" | "" | "disabled"

	FirebaseCredentialsFile string // env FIREBASE_CREDENTIALS_FILE
	FirebaseProjectID       string // env FIREBASE_PROJECT_ID; optional
}
