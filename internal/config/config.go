//nolint:revive
package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

// HTTPConfig конфигурация HTTP сервера
type HTTPConfig struct {
	ListenPort   string `yaml:"listen_port"`
	ReadTimeout  int    `yaml:"read_timeout"`
	WriteTimeout int    `yaml:"write_timeout"`
}

// RetryConfig конфигурация повторных попыток
type RetryConfig struct {
	MaxAttempts int           `yaml:"max_attempts"`
	Delay       time.Duration `yaml:"delay"`
	MaxDelay    time.Duration `yaml:"max_delay"`
}

// BackendConfig конфигурация бэкенда
type BackendConfig struct {
	URL string `yaml:"url"`
}

// BalancerConfig конфигурация балансировщика
type BalancerConfig struct {
	Strategy            string          `yaml:"strategy"`
	BackedsFile         string          `yaml:"backends_file"`
	Backends            []BackendConfig `yaml:"-"`
	HealthCheckInterval time.Duration   `yaml:"health_check_interval"`
}

// Config конфигурация приложения
type Config struct {
	HTTPConfig     HTTPConfig     `yaml:"http"`
	RetryConfig    RetryConfig    `yaml:"retry"`
	BalancerConfig BalancerConfig `yaml:"balancer"`
}

// LoadConfig загружает конфигурацию из файла
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	config.BalancerConfig.Backends, err = LoadBackends(config.BalancerConfig.BackedsFile)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// LoadBackends загружает бэкенды из файла
func LoadBackends(path string) ([]BackendConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var backends []BackendConfig
	err = yaml.Unmarshal(data, &backends)
	if err != nil {
		return nil, err
	}
	return backends, nil
}
