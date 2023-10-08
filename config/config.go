package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type (
	// Config -.
	Config struct {
		App      `yaml:"http"`
		HTTP     `yaml:"http"`
		Log      `yaml:"logger"`
		PG       `yaml:"postgres"`
		Auth     `yaml:"authentication"`
		OutBound `yaml:"out_bound"`
		Command  `yaml:"command"`
	}

	// App -.
	App struct {
		Name          string `env-required:"true" yaml:"name"    env:"APP_NAME"`
		Version       string `env-required:"true" yaml:"version" env:"APP_VERSION"`
		HostCallback  string `env-required:"true" yaml:"host_callback" env:"APP_HOST_CALLBACK"`
		UpdateTimeOut int    `env-required:"true" env:"TIMEOUT"`
	}

	// HTTP -.
	HTTP struct {
		Port string `env-required:"true" yaml:"port" env:"HTTP_PORT"`
		Host string `env-required:"true" yaml:"port" env:"HOST"`
	}

	// Log -.
	Log struct {
		Level string `env-required:"true" yaml:"log_level"   env:"LOG_LEVEL"`
	}

	// PG -.
	PG struct {
		PoolMax int    `env-required:"true" yaml:"pool_max" env:"PG_POOL_MAX"`
		URL     string `env-required:"true"                 env:"PG_URL"`
	}

	Auth struct {
		GoogleClientID string `env-required:"true" yaml:"google_client_id" env:"GOOGLE_CLIENT_ID"`
		GoogleSecret   string `env-required:"true" yaml:"google_secret" env:"GOOGLE_SECRET"`

		// JWT SECRET
		JWTSecret string `env-required:"true" env:"SECRET"`

		AdminSecret string `env-required:"true" env:"ADMIN_SECRET"`

		XenditAPISecret string `env-required:"true" env:"XENDIT_API_SECRET"`

		AlertID int64 `env:"ALERT_ID"`

		AdaAPISecret string `env-required:"true" env:"ADA_API_SECRET"`
	}

	OutBound struct {
		FrontEndURL string `env-required:"true"  env:"FRONTEND_URL"`
		VertexURL   string `env-required:"true" env:"VERTEX_URL"`
		ChatModel   string `env-required:"true" env:"VERTEX_MODEL_CHAT"`
		TextModel   string `env-required:"true" env:"VERTEX_MODEL_TEXT"`
		AdaHostURL  string `env-required:"true" env:"ADA_HOST_URL"`
	}

	Command struct {
		MonolithURL string `env-required:"true"  env:"MONOLITH_URL"`
		CommandStr  string `env-required:"true" env:"COMMAND_STRING"`
	}
)

// NewConfig returns http config.
func NewConfig() (*Config, error) {
	cfg := &Config{}

	err := cleanenv.ReadEnv(cfg)
	if err == nil && cfg != nil {
		return cfg, nil
	}

	err = cleanenv.ReadConfig(".env", cfg)
	if err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	err = cleanenv.ReadEnv(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
