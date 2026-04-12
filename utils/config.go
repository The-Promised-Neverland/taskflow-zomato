package utils

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	Name     string `env:"NAME" envDefault:"taskflow"`
	Env      string `env:"ENV" envDefault:"development"`
	LogLevel string `env:"LOG_LEVEL" envDefault:"DEBUG"`

	// HTTP server settings
	ServerPort          string        `env:"SERVER_PORT" envDefault:"8080"`
	ServerReadTimeout   time.Duration `env:"SERVER_READ_TIMEOUT" envDefault:"60s"`
	ServerWriteTimeout  time.Duration `env:"SERVER_WRITE_TIMEOUT" envDefault:"60s"`
	ServerIdleTimeout   time.Duration `env:"SERVER_IDLE_TIMEOUT" envDefault:"120s"`
	ServerHeaderTimeout time.Duration `env:"SERVER_HEADER_TIMEOUT" envDefault:"60s"`

	// Database connection settings
	DBConnectionStr string `env:"DB_CONNECTION_STR"`
	DBHost          string `env:"DB_HOST"`
	DBPort          string `env:"DB_PORT"`
	DBName          string `env:"DB_NAME"`
	DBUser          string `env:"DB_USER"`
	DBPassword      string `env:"DB_PASSWORD"`
	DBSSLMode       string `env:"DB_SSLMODE"`

	// JWT settings
	JWTSecret              string        `env:"JWT_SECRET"`
	AccessTokenExpiration  time.Duration `env:"ACCESS_TOKEN_EXPIRATION" envDefault:"15m"`
	RefreshTokenExpiration time.Duration `env:"REFRESH_TOKEN_EXPIRATION" envDefault:"30d"`
}

var appConfig *Config

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(".env"); err != nil {
		fmt.Println("No .env file found, reading from environment")
	}

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	if cfg.DBConnectionStr == "" {
		if cfg.DBHost == "" || cfg.DBPort == "" || cfg.DBName == "" || cfg.DBUser == "" || cfg.DBPassword == "" || cfg.DBSSLMode == "" {
			return nil, fmt.Errorf("database config is incomplete: set DB_CONNECTION_STR or DB_HOST, DB_PORT, DB_NAME, DB_USER, DB_PASSWORD, DB_SSLMODE")
		}
		cfg.DBConnectionStr = fmt.Sprintf(
			"postgresql://%s:%s@%s:%s/%s?sslmode=%s",
			cfg.DBUser,
			cfg.DBPassword,
			cfg.DBHost,
			cfg.DBPort,
			cfg.DBName,
			cfg.DBSSLMode,
		)
	}

	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	return cfg, nil
}

func LoadAndGetConfig() *Config {
	if appConfig != nil {
		return appConfig
	}

	cfg, err := LoadConfig()
	if err != nil {
		panic(err)
	}

	appConfig = cfg
	return appConfig
}
