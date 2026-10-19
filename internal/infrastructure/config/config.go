package config

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"math-ai.com/math-ai/internal/shared/utils"
)

// NewEnv configuration data from .env file
// Returns a configData struct
func NewEnv(envpath string) (*Env, error) {
	if err := godotenv.Load(envpath); err != nil {
		return nil, fmt.Errorf("failed to load env file: %v", err)
	}

	result := Env{
		ServerHost:             getConfig("SERVER_HOST"),
		ServerPort:             getConfig("SERVER_PORT"),
		LogFile:                getConfigOptionalString("LOG_FILE"),
		LogFormat:              getConfigOptionalString("LOG_FORMAT"), // "" = use per-destination formats below
		LogConsoleFormat:       getConfigOptionalStringWithDefault("LOG_CONSOLE_FORMAT", "text"),
		LogFileFormat:          getConfigOptionalStringWithDefault("LOG_FILE_FORMAT", "json"),
		SerializedSessionFile:  getConfigOptionalString("SERIALIZED_SESSION_FILE"),
		SessionPersistOnChange: getBoolConfigWithDefault("SESSION_PERSIST_ON_CHANGE", false),
		PowEnabled:             getBoolConfigWithDefault("POW_ENABLED", false),
		PowDifficulty:          getIntConfigOptionalWithDefault("POW_DIFFICULTY", 4),
		GexSessionDriver:       getConfigOptionalString("GEX_SESSION_DRIVER"),
		SharedKeyBytes:         getFileBytesConfig("GEX_SHARED_KEY"),
		HttpsCertFile:          getConfigOptional("HTTPS_CERT_FILE"),
		HttpsKeyFile:           getConfigOptional("HTTPS_KEY_FILE"),
		EnableOTP:              getBoolConfigWithDefault("ENABLE_OTP", false),
		EmailOTPEnable:         getBoolConfigWithDefault("EMAIL_OTP_ENABLE", false),
		PhoneOTPEnable:         getBoolConfigWithDefault("PHONE_OTP_ENABLE", false),
		OtpBypassEnabled:       getBoolConfigWithDefault("OTP_BYPASS_ENABLED", false),
		OtpBypassCode:          getConfigOptionalString("OTP_BYPASS_CODE"),
		TrustDeviceTTLDays:     getIntConfigOptionalWithDefault("TRUST_DEVICE_TTL", 30),
		DemoNames:              getCSVConfig("DEMO_NAMES"),
		AdminApiKey:            getConfigOptionalString("ADMIN_API_KEY"),
		ExamShuffleEnabled:     getBoolConfigWithDefault("EXAM_SHUFFLE_ENABLED", true),
		RateLimitMode:          getConfigOptionalStringWithDefault("RATE_LIMIT_MODE", "shadow"),

		DBConfig: DBConfig{
			DBHost: getConfig("DB_HOST"),
			DBPort: getConfig("DB_PORT"),
			DBUser: getConfig("DB_USER"),
			DBPass: getConfig("DB_PASS"),
			DBName: getConfig("DB_NAME"),

			MaxOpenConns:    getIntConfigOptional("DB_MAX_OPEN_CONNS"),
			MaxIdleConns:    getIntConfigOptional("DB_MAX_IDLE_CONNS"),
			ConnMaxLifetime: time.Duration(getIntConfigOptional("DB_CONN_MAX_LIFETIME")) * time.Second,
			ConnMaxIdleTime: time.Duration(getIntConfigOptional("DB_CONN_MAX_IDLE_TIME")) * time.Second,
		},

		EmailConfig: EmailConfig{
			EmailProvider: getConfig("EMAIL_PROVIDER"),

			EmailHost:     getConfigOptionalString("EMAIL_HOST"),
			EmailPort:     getIntConfigOptional("EMAIL_PORT"),
			EmailUser:     getConfigOptionalString("EMAIL_USER"),
			EmailPassword: getConfigOptionalString("EMAIL_PASSWORD"),
			EmailFrom:     getConfigOptionalString("EMAIL_FROM"),

			GmailCredentialsFile: getConfigOptionalString("GMAIL_CREDENTIALS_FILE"),
			GmailSenderEmail:     getConfigOptionalString("GMAIL_SENDER_EMAIL"),
		},

		StorageConfig: StorageConfig{
			Provider:  getConfigOptionalString("STORAGE_PROVIDER"),
			AccessKey: getConfigOptionalString("STORAGE_ACCESS_KEY"),
			SecretKey: getConfigOptionalString("STORAGE_SECRET_KEY"),
			Region:    getConfigOptionalString("STORAGE_REGION"),
			Bucket:    getConfigOptionalString("STORAGE_BUCKET"),
		},

		SMSConfig: SMSConfig{
			SMSProvider: getConfigOptionalString("SMS_PROVIDER"),

			TwilioAccountSID:          getConfigOptionalString("TWILIO_ACCOUNT_SID"),
			TwilioAuthToken:           getConfigOptionalString("TWILIO_AUTH_TOKEN"),
			TwilioBaseURL:             getConfigOptionalString("TWILIO_BASE_URL"),
			TwilioFrom:                getConfigOptionalString("TWILIO_FROM"),
			TwilioMessagingServiceSID: getConfigOptionalString("TWILIO_MESSAGING_SERVICE_SID"),

			Timeout:       getDurationConfigOptional("SMS_TIMEOUT"),
			MaxRetries:    getIntConfigOptional("SMS_MAX_RETRIES"),
			RetryDelay:    getDurationConfigOptional("SMS_RETRY_DELAY"),
			RequireAtBoot: getBoolConfigWithDefault("SMS_REQUIRE_AT_BOOT", false),
		},

		LLMConfig: LLMConfig{
			DefaultLLMProvider: getConfigOptionalString("LLM_PROVIDER"),

			OpenRouterAPIKey:      getConfigOptionalString("LLM_OPENROUTER_API_KEY"),
			OpenRouterBaseURL:     getConfigOptionalString("LLM_OPENROUTER_BASE_URL"),
			OpenRouterModel:       getConfigOptionalString("LLM_OPENROUTER_MODEL"),
			OpenRouterSiteURL:     getConfigOptionalString("LLM_OPENROUTER_SITE_URL"),
			OpenRouterAppTitle:    getConfigOptionalString("LLM_OPENROUTER_APP_TITLE"),
			OpenRouterTemperature: getFloatConfigWithDefault("LLM_OPENROUTER_TEMPERATURE", -1),
			OpenRouterTopP:        getFloatConfigWithDefault("LLM_OPENROUTER_TOP_P", -1),
			OpenRouterMaxTokens:   getIntConfigOptional("LLM_OPENROUTER_MAX_TOKENS"),

			OpenAIAPIKey:       getConfigOptionalString("LLM_OPENAI_API_KEY"),
			OpenAIBaseURL:      getConfigOptionalString("LLM_OPENAI_BASE_URL"),
			OpenAIModel:        getConfigOptionalString("LLM_OPENAI_MODEL"),
			OpenAIEmbedModel:   getConfigOptionalString("LLM_OPENAI_EMBED_MODEL"),
			OpenAIOrganization: getConfigOptionalString("LLM_OPENAI_ORGANIZATION"),
			OpenAIProject:      getConfigOptionalString("LLM_OPENAI_PROJECT"),
			OpenAITemperature:  getFloatConfigWithDefault("LLM_OPENAI_TEMPERATURE", -1),
			OpenAITopP:         getFloatConfigWithDefault("LLM_OPENAI_TOP_P", -1),
			OpenAIMaxTokens:    getIntConfigOptional("LLM_OPENAI_MAX_TOKENS"),

			OpenAIReasoningEffort: getConfigOptionalString("LLM_OPENAI_REASONING_EFFORT"),
			OpenAIStore:           getBoolConfigWithDefault("LLM_OPENAI_STORE", false),

			GeminiAPIKey:      getConfigOptionalString("LLM_GEMINI_API_KEY"),
			GeminiBaseURL:     getConfigOptionalString("LLM_GEMINI_BASE_URL"),
			GeminiModel:       getConfigOptionalString("LLM_GEMINI_MODEL"),
			GeminiEmbedModel:  getConfigOptionalString("LLM_GEMINI_EMBED_MODEL"),
			GeminiTemperature: getFloatConfigWithDefault("LLM_GEMINI_TEMPERATURE", -1),
			GeminiTopP:        getFloatConfigWithDefault("LLM_GEMINI_TOP_P", -1),
			GeminiMaxTokens:   getIntConfigOptional("LLM_GEMINI_MAX_TOKENS"),

			DeepSeekAPIKey:          getConfigOptionalString("LLM_DEEPSEEK_API_KEY"),
			DeepSeekBaseURL:         getConfigOptionalString("LLM_DEEPSEEK_BASE_URL"),
			DeepSeekModel:           getConfigOptionalString("LLM_DEEPSEEK_MODEL"),
			DeepSeekThinking:        getConfigOptionalString("LLM_DEEPSEEK_THINKING"),
			DeepSeekReasoningEffort: getConfigOptionalString("LLM_DEEPSEEK_REASONING_EFFORT"),
			DeepSeekTemperature:     getFloatConfigWithDefault("LLM_DEEPSEEK_TEMPERATURE", -1),
			DeepSeekTopP:            getFloatConfigWithDefault("LLM_DEEPSEEK_TOP_P", -1),
			DeepSeekMaxTokens:       getIntConfigOptional("LLM_DEEPSEEK_MAX_TOKENS"),

			Timeout:       getDurationConfigOptional("LLM_TIMEOUT"),
			MaxRetries:    getIntConfigOptional("LLM_MAX_RETRIES"),
			RetryDelay:    getDurationConfigOptional("LLM_RETRY_DELAY"),
			RequireAtBoot: getBoolConfigWithDefault("LLM_REQUIRE_AT_BOOT", false),
		},

		NotificationConfig: NotificationConfig{
			Provider: getConfigOptionalString("NOTIFICATION_PROVIDER"),

			FirebaseCredentialsFile: getConfigOptionalString("FIREBASE_CREDENTIALS_FILE"),
			FirebaseProjectID:       getConfigOptionalString("FIREBASE_PROJECT_ID"),
		},

		SocketConfig: SocketConfig{
			Enabled:         getBoolConfigWithDefault("SOCKET_ENABLED", true),
			AllowedOrigins:  getCSVConfig("SOCKET_ALLOWED_ORIGINS"),
			PingInterval:    getDurationConfigOptional("SOCKET_PING_INTERVAL"),
			WriteTimeout:    getDurationConfigOptional("SOCKET_WRITE_TIMEOUT"),
			ReadLimit:       int64(getIntConfigOptional("SOCKET_READ_LIMIT")),
			WriteBuffer:     getIntConfigOptional("SOCKET_WRITE_BUFFER"),
			MaxConnsPerUser: getIntConfigOptional("SOCKET_MAX_CONNS_PER_USER"),
		},

		ObservabilityConfig: ObservabilityConfig{
			ServiceName:    getConfigOptionalStringWithDefault("OBS_SERVICE_NAME", "math-svr"),
			ServiceVersion: getConfigOptionalStringWithDefault("OBS_SERVICE_VERSION", "dev"),
			Environment:    getConfigOptionalStringWithDefault("OBS_ENV", "local"),

			MetricsEnabled: getBoolConfigWithDefault("OBS_METRICS_ENABLED", true),
			MetricsAddr:    getConfigOptionalStringWithDefault("OBS_METRICS_ADDR", ":9091"),

			TracingEnabled:   getBoolConfigWithDefault("OBS_TRACING_ENABLED", false),
			OTLPEndpoint:     getConfigOptionalStringWithDefault("OBS_OTLP_ENDPOINT", "localhost:4318"),
			OTLPInsecure:     getBoolConfigWithDefault("OBS_OTLP_INSECURE", true),
			TraceSampleRatio: getFloatConfigWithDefault("OBS_TRACE_SAMPLE_RATIO", 1.0),
		},
	}

	return &result, nil
}

func getIntConfigOptional(key string) int {
	val := getConfigOptional(key)
	if val == nil || *val == "" {
		return 0
	}
	i, err := strconv.Atoi(*val)
	if err != nil {
		panic(fmt.Sprintf("config error: %s must be an integer", key))
	}
	return i
}

func getIntConfigOptionalWithDefault(key string, def int) int {
	raw := getConfigOptional(key)
	if raw == nil || *raw == "" {
		return def
	}
	i, err := strconv.Atoi(*raw)
	if err != nil {
		panic(fmt.Sprintf("config error: %s must be an integer", key))
	}
	return i
}

func getFloatConfigOptional(key string) *float64 {
	val := getConfigOptional(key)
	if val == nil {
		return nil
	}
	floatVal, _ := utils.StringToFloat64Err(*val)
	return &floatVal
}

func getFloatConfigWithDefault(key string, def float64) float64 {
	raw := getConfigOptional(key)
	if raw == nil || *raw == "" {
		return def
	}
	v, err := strconv.ParseFloat(*raw, 64)
	if err != nil {
		panic(fmt.Sprintf("config error: %s must be a float", key))
	}
	return v
}

func getBoolConfig(key string) bool {
	val := getConfigOptional(key)
	if val == nil {
		return false
	}
	return *val == "true"
}

func getBoolConfigWithDefault(key string, def bool) bool {
	raw := getConfigOptional(key)
	if raw == nil {
		return def
	}
	return *raw == "true"
}

func getDurationConfigOptional(key string) time.Duration {
	raw := getConfigOptional(key)
	if raw == nil || *raw == "" {
		return 0
	}
	d, err := time.ParseDuration(*raw)
	if err != nil {
		panic(fmt.Sprintf("config error: %s must be a Go duration string (e.g. \"2s\", \"100ms\"): %v", key, err))
	}
	return d
}

func getConfigOptionalStringWithDefault(key, def string) string {
	val := getConfigOptionalString(key)
	if val == "" {
		return def
	}
	return val
}

// getCSVConfig parses a comma-separated env value into a trimmed, non-empty
// slice. Returns nil when the key is unset or blank.
func getCSVConfig(key string) []string {
	raw := getConfigOptionalString(key)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func getConfigOptionalString(key string) string {
	val := getConfigOptional(key)
	if val == nil {
		return ""
	}
	return *val
}

func getConfigOptional(key string) *string {
	if os.Getenv(key) == "" {
		return nil
	}
	val := os.Getenv(key)
	return &val
}

func getConfig(key string) string {
	val := getConfigOptional(key)
	if val == nil || *val == "" {
		panic(fmt.Sprintf("config error: config error: %s is not set", key))
	}
	return *val
}

func getFileBytesConfig(key string) []byte {
	path := getConfig(key)
	bytes, err := loadFile(path)
	if err != nil || bytes == nil {
		panic(fmt.Errorf("config error: %s failed to load file: %w", path, err))
	}
	return bytes
}

func loadFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("config error: %s failed to open file: %w", path, err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("config error: %s failed to get file info: %w", path, err)
	}

	data := make([]byte, fileInfo.Size())
	_, err = io.ReadFull(file, data)
	if err != nil {
		return nil, fmt.Errorf("config error: %s failed to read file: %w", path, err)
	}

	return data, nil
}
