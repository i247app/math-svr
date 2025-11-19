package config

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Env struct {
	DBEnv                 *DBConfig
	ServerEnv             *ServerConfig
	HostConfig            *HostConfig
	TwilioConfig          *TwilioConfig
	MailerConfig          *MailerConfig
	SharedKeyBytes        []byte
	RootSessionDriver     string
	SerializedSessionFile string
}

func NewEnv(envpath string) (*Env, error) {
	err := godotenv.Load(envpath)
	if err != nil {
		return nil, fmt.Errorf("failed to load env file: %v", err)
	}

	result := &Env{
		DBEnv: &DBConfig{
			Host:     getConfig("DB_HOST"),
			Port:     getConfig("DB_PORT"),
			User:     getConfig("DB_USER"),
			Password: getConfig("DB_PASSWORD"),
			Name:     getConfig("DB_NAME"),
			SSLMode:  getConfig("DB_SSL_MODE"),
		},
		ServerEnv: &ServerConfig{
			Port:        getConfig("SERVER_PORT"),
			LogFilePath: getConfig("LOG_FILE_PATH"),
		},
		HostConfig: &HostConfig{
			ServerMode:    getConfig("SERVER_MODE"),
			ServerHost:    getConfig("SERVER_HOST"),
			ServerPort:    getConfig("SERVER_PORT"),
			HttpsCertFile: getConfigOptional("HTTPS_CERT_FILE"),
			HttpsKeyFile:  getConfigOptional("HTTPS_KEY_FILE"),
		},
		TwilioConfig: &TwilioConfig{
			AccountSID:          getConfigOptional("TWILIO_ACCOUNT_SID"),
			AuthToken:           getConfigOptional("TWILIO_AUTH_TOKEN"),
			FromPhoneNumber:     getConfigOptional("TWILIO_FROM_PHONE_NUMBER"),
			MessagingServiceSID: getConfigOptional("TWILIO_MESSAGING_SERVICE_SID"),
		},
		MailerConfig: &MailerConfig{
			SMTPHost: getConfigOptional("MAIL_HOST"),
			SMTPPort: getIntConfigOptional("MAIL_PORT"),
			Username: getConfigOptional("MAIL_USER"),
			Password: getConfigOptional("MAIL_PASSWORD"),
			FromName: getConfigOptional("MAIL_FROM"),
		},
		SharedKeyBytes:        getFileBytesConfig("ROOT_SHARED_KEY"),
		RootSessionDriver:     getConfig("ROOT_SESSION_DRIVER"),
		SerializedSessionFile: getConfig("SERIALIZED_SESSION_FILE"),
	}

	return result, nil
}

func getFloatConfigOptional(key string) *float64 {
	if os.Getenv(key) == "" {
		return nil
	}
	val := os.Getenv(key)
	floatVal, _ := strconv.ParseFloat(val, 64)
	return &floatVal
}

func getConfigOptional(key string) *string {
	if os.Getenv(key) == "" {
		return nil
	}
	val := os.Getenv(key)
	return &val
}

func getIntConfigOptional(key string) *int {
	if os.Getenv(key) == "" {
		return nil
	}
	val := os.Getenv(key)
	intVal, _ := strconv.Atoi(val)
	return &intVal
}

func getConfig(key string) string {
	val := getConfigOptional(key)
	if val == nil {
		return ""
	}
	return *val
}

func getBoolConfig(key string) bool {
	val := getConfigOptional(key)
	if val == nil {
		return false
	}
	return *val == "true"
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
