package config

import (
	"encoding/json"
	"os"

	"github.com/ansel1/merry/v2"
)

// BackupConfigItem represents a config item from the config file
type BackupConfigItem struct {
	// Name of the config
	Name string `json:"name"`

	// Source directory to be backed up
	// Can be a "remote" of form "<RCLONE_REMOTE_NAME>:"
	SourceDir string `json:"sourceDir"`

	// Destination directory to be backed up to
	// Can be a "remote" of form "<RCLONE_REMOTE_NAME>:"
	DestDir string `json:"destDir"`
}

// Config Represents the backup config from the config file
type Config struct {
	// Items that are in the file
	BackupConfigItem []BackupConfigItem `json:"items"`
}

// ParseConfigFile reads and marshals the file into a Config type struct
func ParseConfigFile(filePath string) (*Config, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, merry.Errorf("failed to open config file")
	}
	defer file.Close()

	config := &Config{}
	jsonParser := json.NewDecoder(file)
	if err = jsonParser.Decode(config); err != nil {
		return nil, merry.Errorf("failed to parse config file")
	}

	return config, nil
}
