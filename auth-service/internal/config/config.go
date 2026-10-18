package config

import (
	"fmt"
	"sync"
	"time"

	pkgConfig "github.com/Holarbtw/homestock-backend/pkg/config"
)

type Config struct {
	Port string `env:"AUTH_SERVICE_PORT" env-default:"8080"`
	DB   DBConfig
	JWT  JWTConfig
	Security SecurityConfig
}

type DBConfig struct {
	User     string `env:"DB_USER" env-required:"true"`
	Password string `env:"DB_PASSWORD" env-required:"true"`
	Host     string `env:"DB_HOST" env-default:"postgres"`
	Port     string `env:"DB_PORT" env-default:"5432"`
	Name     string `env:"AUTH_DB_NAME" env-required:"true"`
}

type JWTConfig struct {
	Secret                string `env:"JWT_SECRET" env-required:"true"`
	AccessTokenTTLMinutes int    `env:"ACCESS_TOKEN_TTL_MINUTES" env-default:"15"`
	RefreshTokenTTLDays   int    `env:"REFRESH_TOKEN_TTL_DAYS" env-default:"30"`
}

func (j *JWTConfig) AccessTokenTTL() time.Duration {
	return time.Duration(j.AccessTokenTTLMinutes) * time.Minute
}

func (j *JWTConfig) RefreshTokenTTL() time.Duration {
	return time.Duration(j.RefreshTokenTTLDays) * 24 * time.Hour
}

type SecurityConfig struct {
    BcryptCost int `env:"BCRYPT_COST" env-default:"10"`
}

func (c *DBConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.User, c.Password, c.Host, c.Port, c.Name,
	)
}

var (
	cfg  *Config
	once sync.Once
)

func GetConfig() (*Config, error) {
	var loadErr error
	once.Do(func() {
		cfg = &Config{}
		if err := pkgConfig.Load(cfg); err != nil {
			loadErr = fmt.Errorf("failed to load auth-service config: %w", err)
		}
	})

	if loadErr != nil {
		return nil, loadErr
	}
	return cfg, nil
}
