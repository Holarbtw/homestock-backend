package config

import (
	"fmt"
	"sync"

	pkgConfig "github.com/Holarbtw/homestock-backend/pkg/config"
)

type Config struct {
	Port string `env:"CATALOG_SERVICE_PORT" env-default:"9090"`
	DB   DBConfig
}

type DBConfig struct {
	User     string `env:"DB_USER" env-required:"true"`
	Password string `env:"DB_PASSWORD" env-required:"true"`
	Host     string `env:"DB_HOST" env-default:"postgres"`
	Port     string `env:"DB_PORT" env-default:"5432"`
	Name     string `env:"CATALOG_DB_NAME" env-required:"true"`
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
			loadErr = fmt.Errorf("failed to load catalog-service config: %w", err)
		}
	})

	if loadErr != nil {
		return nil, loadErr
	}
	return cfg, nil
}
