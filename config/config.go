package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Represents the Wingflow config.
type Config struct {
	// Git repository credentials
	// This currently only uses the git address, meaning there is no support for fetching from private repositories yet.
	Git struct {
		Address string `yaml:"address"`
		Token   string `yaml:"token,omitempty"`
	} `yaml:"git"`

	// Pterodactyl panel credentials
	// URL - The url to your panel. Only put the HTTP scheme and your domain
	//		 (e.g. "https://pterodactyl.domain").
	//
	// Key - This should be a CLIENT API key, not an application API key. To prevent
	//		 request issues, do not restrict this key to an IP address.
	//
	// ID  - The identifier of the server to upload and deploy to. You can find this in
	//		 the URL when going to your server (e.g. "https://pterodacty.domain/servers/84ac52b",
	//		 the last part is the identifier).
	Panel struct {
		URL string `yaml:"url"`
		Key string `yaml:"key"`
		ID  string `yaml:"id"`
	} `yaml:"panel"`

	// Configurations for the repository.
	// Include - A list of file paths or patterns to include in the file upload.
	// Exclude - A list of file paths or patterns to exclude from being uploaded.
	Repository struct {
		Include []string `yaml:"include"`
		Exclude []string `yaml:"exlude"`
	} `yaml:"repository"`

	// Configurations for the file system in the server.
	// ForceKill - Whether to kill the server instead of stopping it.
	// Ignore    - A list of files to ignore from being deleted.
	System struct {
		ForceKill bool     `yaml:"force_kill"`
		Ignore    []string `yaml:"ignore"`
	} `yaml:"system"`

	// Commands to execute before starting the main process.
	PreRun []string `yaml:"pre_run"`

	// Commands to execute after the main process is completed.
	PostRun []string `yaml:"post_run"`
}

func exists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		return !os.IsNotExist(err)
	}

	return true
}

func Fetch(dir string) (*Config, error) {
	path := filepath.Join(filepath.Clean(dir), "wingflow.yml")
	if !exists(path) {
		return nil, errors.New("file path does not exist")
	}

	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg *Config
	err = yaml.Unmarshal(buf, &cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func Create(dir string, force bool) error {
	path := filepath.Join(filepath.Clean(dir), "wingflow.yml")
	if exists(path) && !force {
		return errors.New("config file already exists at this path")
	}

	file, err := os.Create(path)
	// shouldn't happen, but just in case
	if err != nil {
		return err
	}

	buf, _ := yaml.Marshal(Config{})
	defer file.Close()
	file.Write(buf)

	return nil
}
