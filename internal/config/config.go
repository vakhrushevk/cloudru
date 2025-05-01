//nolint:revive
package config

import (
	"io"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

// Config конфигурация приложения
type Config struct {
	HTTPConfig     HTTPConfig     `yaml:"http"`
	RetryConfig    RetryConfig    `yaml:"retry"`
	BalancerConfig BalancerConfig `yaml:"balancer"`
	LoggerConfig   LoggerConfig   `yaml:"logger"`
	BucketConfig   BucketConfig   `yaml:"bucket"`
	RedisConfig    RedisConfig    `yaml:"redis"`
}

// HTTPConfig конфигурация HTTP сервера
type HTTPConfig struct {
	ListenPort   int `yaml:"listen_port"`
	ReadTimeout  int `yaml:"read_timeout"`
	WriteTimeout int `yaml:"write_timeout"`
}

// RetryConfig конфигурация повторных попыток
type RetryConfig struct {
	MaxAttempts int           `yaml:"max_attempts"`
	Delay       time.Duration `yaml:"delay"`
	MaxDelay    time.Duration `yaml:"max_delay"`
}

// BalancerConfig конфигурация балансировщика
type BalancerConfig struct {
	Strategy            string          `yaml:"strategy"`
	BackedsFile         string          `yaml:"backends_file"`
	Backends            []BackendConfig `yaml:"-"`
	HealthCheckInterval time.Duration   `yaml:"health_check_interval"`
}

// BackendConfig конфигурация бэкенда
type BackendConfig struct {
	URL string `yaml:"url"`
}

// LoggerConfig конфигурация логгера
type LoggerConfig struct {
	LogLevel  string `yaml:"log_level"`
	LogFormat string `yaml:"log_format"`
	LogOutput string `yaml:"log_output"`
}

func (l *LoggerConfig) Level() string {
	return l.LogLevel
}

func (l *LoggerConfig) Format() string {
	return l.LogFormat
}

func (l *LoggerConfig) Output() io.Writer {
	switch l.LogOutput {
	case "stdout":
		return os.Stdout
	case "stderr":
		return os.Stderr
	default:
		f, err := os.OpenFile(l.LogOutput, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			// Fallback to stdout on error
			return os.Stdout
		}
		return f
	}
}

// BucketConfig конфигурация бакета
type BucketConfig struct {
	Capacity  int           `yaml:"capacity"`   // default Максимальное количество токенов в бакете
	RefilRate int           `yaml:"refil_rate"` // default Дефолтное время заполнения токенов для бакета
	RefilTime time.Duration `yaml:"refil_time"` // Время через которое будет запущено заполнение токенов для бакета
	Tokens    int           `yaml:"tokens"`     // default Количество токенов в бакете
}

type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
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
