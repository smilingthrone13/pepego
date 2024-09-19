package config

import (
	"os"
	"path"
	"time"

	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type Config struct {
	IsDebug        bool          `yaml:"is_debug"`
	ApiKey         string        `yaml:"api_key"`
	GetterCooldown time.Duration `yaml:"getter_cooldown"`
	ImagesDirPath  string        `yaml:"images_dir_path"`
	RequestTimeout time.Duration `yaml:"request_timeout"`
}

func NewConfig(cfgFolderPath string) (*Config, error) {
	cfgPath := path.Join(cfgFolderPath, "config.yaml")
	envPath := path.Join(cfgFolderPath, ".env")

	c := &Config{
		IsDebug:        false,
		GetterCooldown: 5 * time.Second,
		RequestTimeout: 5 * time.Second,
	}

	err := c.loadConfig(cfgPath)
	if err != nil {
		err = errors.Wrap(err, "NewConfig")

		return nil, err
	}

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

	c.ApiKey = os.Getenv("BOT_TOKEN")

	return nil
}

func (c *Config) validate() error {
	if c.ApiKey == "" {
		err := errors.New("api_key is required")

		return err
	}

	if c.ImagesDirPath == "" {
		err := errors.New("images_dir_path is required")

		return err
	}

	return nil
}
