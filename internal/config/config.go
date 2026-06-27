package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Port           string        `env:"PORT" envDefault:"8080"`
	DatabaseURL    string        `env:"DATABASE_URL,required"`
	DBMaxConns     int32         `env:"DB_MAX_CONNS" envDefault:"10"`
	DBIdleMaxTime  time.Duration `env:"DB_MAX_IDLE_TIME" envDefault:"15m"`
	Environment    string        `env:"ENVIRONMENT" envDefault:"development"`
	LogLevel       string        `env:"LOG_LEVEL" envDefault:"info"`
	AllowedOrigins []string      `env:"ALLOWED_ORIGINS" envDefault:"*"`
	JWTSecret      string        `env:"JWT_SECRET,required"`
	JWTExpiryHours int           `env:"JWT_EXPIRY_HOURS" envDefault:"24"`
}

func (c Config) Validate() error {
	switch c.Environment {
	case "development", "staging", "production":
		// ok
	default:
		return fmt.Errorf("Environment must be development, staging or production, got %q", c.Environment)
	}

	switch c.LogLevel {
	case "debug", "info", "warn", "error":
		// ok
	default:
		return fmt.Errorf("LOG_LEVEL must be debug, info, warn, or error, got %q", c.LogLevel)
	}

	return nil
}

func (c Config) IsProduction() bool {
	return c.Environment == "production"
}

func Load() (Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return Config{}, fmt.Errorf("parsing confing: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}
