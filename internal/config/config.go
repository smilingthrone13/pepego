package config

import (
	"os"
	"path"
	"time"

	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

const (
	DefaultCommandCooldown       = time.Second * 5
	DefaultRequestTimeout        = time.Second * 5
	DefaultLastSentQueueSize     = 10
	DefaultMaxRetries            = 3
	DefaultMinSubscriptionPeriod = time.Minute * 15
	DefaultMaxSubscriptionPeriod = time.Hour * 24
)

type Config struct {
	IsDebug               bool          `yaml:"is_debug"`
	ApiKey                string        `yaml:"api_key"`
	DBPath                string        `yaml:"db_path"`
	CommandCooldown       time.Duration `yaml:"command_cooldown"`
	ImagesDirPath         string        `yaml:"images_dir_path"`
	RequestTimeout        time.Duration `yaml:"request_timeout"`
	LastSentQueueSize     int           `yaml:"last_sent_queue_size"`
	MaxRetries            int           `yaml:"max_retries"`
	MinSubscriptionPeriod time.Duration `yaml:"min_subscription_period"`
	MaxSubscriptionPeriod time.Duration `yaml:"max_subscription_period"`
}

func NewConfig(cfgFolderPath string) (*Config, error) {
	c := &Config{
		IsDebug:               false,
		CommandCooldown:       DefaultCommandCooldown,
		RequestTimeout:        DefaultRequestTimeout,
		LastSentQueueSize:     DefaultLastSentQueueSize,
		MaxRetries:            DefaultMaxRetries,
		MinSubscriptionPeriod: DefaultMinSubscriptionPeriod,
		MaxSubscriptionPeriod: DefaultMaxSubscriptionPeriod,
	}

	cfgPath := path.Join(cfgFolderPath, "config.yaml")

	err := c.loadConfig(cfgPath)
	if err != nil {
		err = errors.Wrap(err, "NewConfig")

		return nil, err
	}

	c.MaxSubscriptionPeriod = c.MaxSubscriptionPeriod.Round(time.Second)
	c.MinSubscriptionPeriod = c.MinSubscriptionPeriod.Round(time.Second)

	envFileName := "prod.env"
	if c.IsDebug {
		envFileName = "dev.env"
	}
	envPath := path.Join(cfgFolderPath, envFileName)

	err = c.loadEnv(envPath)
	if err != nil {
		err = errors.Wrap(err, "NewConfig")

		return nil, err
	}

	err = c.validate()
	if err != nil {
		err = errors.Wrap(err, "NewConfig")

		return nil, err
	}

	return c, nil
}

func (c *Config) loadConfig(filePath string) error {
	configFile, err := os.ReadFile(filePath)
	if err != nil {
		err = errors.Wrap(err, "loadConfig")

		return err
	}

	err = yaml.Unmarshal(configFile, c)
	if err != nil {
		err = errors.Wrap(err, "loadConfig")

		return err
	}

	return nil
}

func (c *Config) loadEnv(filePath string) error {
	err := godotenv.Load(filePath)
	if err != nil {
		err = errors.Wrap(err, "loadEnv")

		return err
	}

	c.ApiKey = os.Getenv("api_key")
	c.DBPath = os.Getenv("db_path")

	return nil
}

func (c *Config) validate() error {
	if c.ApiKey == "" {
		err := errors.New("api_key is required")

		return err
	}

	if c.DBPath == "" {
		err := errors.New("db_path is required")

		return err
	}

	if c.ImagesDirPath == "" {
		err := errors.New("images_dir_path is required")

		return err
	}

	return nil
}
