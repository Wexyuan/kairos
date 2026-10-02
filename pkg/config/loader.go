package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Load loads the config file into the target.
func Load(path string, target any) error {
	if path == "" {
		return fmt.Errorf("[config] path is required")
	}
	if target == nil {
		return fmt.Errorf("[config] target is required")
	}

	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("[config] read config file %s error: %w", path, err)
	}

	if err := v.Unmarshal(target); err != nil {
		return fmt.Errorf("[config] unmarshal config file %s into %T error: %w", path, target, err)
	}
	return nil
}
