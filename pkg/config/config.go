package config

import (
	"encoding/json"
	"os"

	"github.com/ansel1/merry/v2"
)

type HistoryFile struct {
	Year int `json:"Year"`

	Path string `json:"Path"`
}

// Config Represents the backup config from the config file
type Config struct {
	// Items that are in the file
	HistoryFiles []HistoryFile `json:"historyFiles"`
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
