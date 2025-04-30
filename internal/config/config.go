package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

type HttpConfig struct {
	ListenPort   string `yaml:"listen_port"`
	ReadTimeout  int    `yaml:"read_timeout"`
	WriteTimeout int    `yaml:"write_timeout"`
}

type RetryConfig struct {
	MaxAttempts int           `yaml:"max_attempts"`
	Delay       time.Duration `yaml:"delay"`
	MaxDelay    time.Duration `yaml:"max_delay"`
}

type BackendConfig struct {
	Url string `yaml:"url"`
}

type BalancerConfig struct {
	Strategy    string          `yaml:"strategy"`
	BackedsFile string          `yaml:"backends_file"`
	Backends    []BackendConfig `yaml:"-"`
}

type Config struct {
	HttpConfig     `yaml:"http"`
	RetryConfig    `yaml:"retry"`
	BalancerConfig `yaml:"balancer"`
}

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

	config.Backends, err = LoadBackends(config.BackedsFile)
	if err != nil {
		return nil, err
	}
	Watcher, err := NewWatcher(config.BackedsFile)
	Watcher.DoRun(func() {
		config.Backends, err = LoadBackends(config.BackedsFile)
		if err != nil {
			return
		}
	})

	return &config, nil
}

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
