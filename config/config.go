package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Load reads configuration from environment variables and optional .env file.
func Load(prefix string) (*viper.Viper, error) {
	v := viper.New()

	// Read from .env file if present
	v.SetConfigFile(".env")
	v.SetConfigType("env")
	_ = v.ReadInConfig() // Ignore error if .env doesn't exist

	// Also read from environment variables
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	return v, nil
}

// DatabaseConfig holds database-specific configuration.
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// LoadDatabaseConfig extracts database config from Viper.
func LoadDatabaseConfig(v *viper.Viper, dbNameKey string) DatabaseConfig {
	return DatabaseConfig{
		Host:     v.GetString("DB_HOST"),
		Port:     v.GetInt("DB_PORT"),
		User:     v.GetString("DB_USER"),
		Password: v.GetString("DB_PASSWORD"),
		DBName:   v.GetString(dbNameKey),
		SSLMode:  v.GetString("DB_SSL_MODE"),
	}
}

// KafkaConfig holds Kafka-specific configuration.
type KafkaConfig struct {
	Brokers     []string
	GroupPrefix string
}

// LoadKafkaConfig extracts Kafka config from Viper.
func LoadKafkaConfig(v *viper.Viper) KafkaConfig {
	brokersStr := v.GetString("KAFKA_BROKERS")
	brokers := strings.Split(brokersStr, ",")
	return KafkaConfig{
		Brokers:     brokers,
		GroupPrefix: v.GetString("KAFKA_GROUP_PREFIX"),
	}
}

// JWTConfig holds JWT-specific configuration.
type JWTConfig struct {
	Secret        string
	AccessExpiry  string
	RefreshExpiry string
}

// LoadJWTConfig extracts JWT config from Viper.
func LoadJWTConfig(v *viper.Viper) JWTConfig {
	return JWTConfig{
		Secret:        v.GetString("JWT_SECRET"),
		AccessExpiry:  v.GetString("JWT_ACCESS_EXPIRY"),
		RefreshExpiry: v.GetString("JWT_REFRESH_EXPIRY"),
	}
}

// GetAppEnv returns the current application environment.
// Defaults to "production" if APP_ENV is not set.
func GetAppEnv(v *viper.Viper) string {
	env := v.GetString("APP_ENV")
	if env == "" {
		return "production"
	}
	return env
}

// GetServicePort returns the port for a specific service.
func GetServicePort(v *viper.Viper, key string) string {
	port := v.GetInt(key)
	if port == 0 {
		port = 8080
	}
	return fmt.Sprintf(":%d", port)
}
