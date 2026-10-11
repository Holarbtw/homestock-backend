package config

import (
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

func Load(cfg any) error {
	if _, err := os.Stat(".env"); err == nil {
		if err := cleanenv.ReadConfig(".env", cfg); err != nil {
			return fmt.Errorf("failed to read .env file: %w", err)
		}
	}

	if err := cleanenv.ReadEnv(cfg); err != nil {
		return fmt.Errorf("failed to read environment variables: %w", err)
	}

	return nil
}
