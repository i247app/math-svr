package config

import (
	"fmt"
	"io"
	"os"
	"strconv"
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
		ServerHost:            getConfig("SERVER_HOST"),
		ServerPort:            getConfig("SERVER_PORT"),
		LogFile:               getConfigOptionalString("LOG_FILE"),
		SerializedSessionFile: getConfigOptionalString("SERIALIZED_SESSION_FILE"),
		GexSessionDriver:      getConfigOptionalString("GEX_SESSION_DRIVER"),
		SharedKeyBytes:        getFileBytesConfig("GEX_SHARED_KEY"),
		HttpsCertFile:         getConfigOptional("HTTPS_CERT_FILE"),
		HttpsKeyFile:          getConfigOptional("HTTPS_KEY_FILE"),
		EnableOTP:             getBoolConfigWithDefault("ENABLE_OTP", false),

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

		BotConfig: BotConfig{
			BotProvider: getConfigOptionalString("BOT_PROVIDER"),

			LangChainBackend:     getConfigOptionalString("BOT_LANGCHAIN_BACKEND"),
			LangChainAPIKey:      getConfigOptionalString("BOT_LANGCHAIN_API_KEY"),
			LangChainBaseURL:     getConfigOptionalString("BOT_LANGCHAIN_BASE_URL"),
			LangChainModel:       getConfigOptionalString("BOT_LANGCHAIN_MODEL"),
			LangChainEmbedModel:  getConfigOptionalString("BOT_LANGCHAIN_EMBED_MODEL"),
			LangChainTemperature: getFloatConfigWithDefault("BOT_LANGCHAIN_TEMPERATURE", -1),
			LangChainTopP:        getFloatConfigWithDefault("BOT_LANGCHAIN_TOP_P", -1),
			LangChainMaxTokens:   getIntConfigOptional("BOT_LANGCHAIN_MAX_TOKENS"),

			Timeout:       getDurationConfigOptional("BOT_TIMEOUT"),
			MaxRetries:    getIntConfigOptional("BOT_MAX_RETRIES"),
			RetryDelay:    getDurationConfigOptional("BOT_RETRY_DELAY"),
			RequireAtBoot: getBoolConfigWithDefault("BOT_REQUIRE_AT_BOOT", false),
		},
	}
	return &result, nil
}

func getIntConfigOptional(key string) int {
	val := getConfigOptional(key)
	if val == nil {
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
