package config

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Git struct {
		Address string `valdiate:"uri" yaml:"address"`
		Token   string `yaml:"token,omitempty"`
		Files   struct {
			Include []string `yaml:"include"`
			Exclude []string `yaml:"exclude"`
		} `yaml:"files"`
	} `yaml:"git"`

	Panel struct {
		URL   string `validate:"url" yaml:"url"`
		Key   string `yaml:"key"`
		ID    string `yaml:"id"`
		Files struct {
			Backup   bool `yaml:"backup"`
			Truncate bool `yaml:"truncate"`
		} `yaml:"files"`
		Signal struct {
			Kill    bool  `yaml:"kill"`
			Timeout int64 `yaml:"timeout"`
		} `yaml:"signal"`
	} `yaml:"panel"`
}

func exists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		return !os.IsNotExist(err)
	}

	return true
}

func Get(check bool) (*Config, error) {
	cwd, _ := os.Getwd()
	path := filepath.Join(cwd, ".wflow")
	if !exists(path) {
		return nil, errors.New("file path does not exist")
	}

	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg *Config
	if err = yaml.Unmarshal(buf, &cfg); err != nil {
		return nil, err
	}

	if check {
		if err = validator.New().Struct(cfg); err != nil {
			return nil, err
		}
	}

	return cfg, nil
}

func Create(force bool) error {
	cwd, _ := os.Getwd()
	path := filepath.Join(cwd, ".wflow")
	if exists(path) && !force {
		return errors.New("exists")
	}

	fd, err := os.Create(path)
	if err != nil {
		return err
	}
	defer fd.Close()

	cfg := Config{}
	cfg.Git.Files.Include = []string{"*"}
	cfg.Panel.Files.Backup = true
	cfg.Panel.Signal.Timeout = 10_000

	enc := yaml.NewEncoder(fd)
	enc.SetIndent(2)
	enc.Encode(cfg)
	enc.Close()

	return nil
}
