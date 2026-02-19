package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	ContactsDirectory string `toml:"contacts_directory"`
}

func Load(configPath string) (*Config, error) {
	config := &Config{}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	// If explicit config path provided, use it
	if configPath != "" {
		if _, err := toml.DecodeFile(configPath, config); err != nil {
			return nil, err
		}
		expandTilde(config, homeDir)
		return config, nil
	}

	// Try new config path first
	newConfigPath := filepath.Join(homeDir, ".config", "apeople", "config.toml")
	if _, err := os.Stat(newConfigPath); err == nil {
		if _, err := toml.DecodeFile(newConfigPath, config); err != nil {
			return nil, err
		}
		expandTilde(config, homeDir)
		return config, nil
	}

	// Fallback to legacy config path
	legacyConfigPath := filepath.Join(homeDir, ".config", "denote-contacts", "config.toml")
	if _, err := os.Stat(legacyConfigPath); err == nil {
		// Legacy config uses notes_directory key
		var legacyConfig struct {
			NotesDirectory    string `toml:"notes_directory"`
			ContactsDirectory string `toml:"contacts_directory"`
		}
		if _, err := toml.DecodeFile(legacyConfigPath, &legacyConfig); err != nil {
			return nil, err
		}
		if legacyConfig.ContactsDirectory != "" {
			config.ContactsDirectory = legacyConfig.ContactsDirectory
		} else {
			config.ContactsDirectory = legacyConfig.NotesDirectory
		}
		expandTilde(config, homeDir)
		return config, nil
	}

	// Use defaults if no config file
	config.ContactsDirectory = filepath.Join(homeDir, "Documents", "denote")
	return config, nil
}

func expandTilde(config *Config, homeDir string) {
	if len(config.ContactsDirectory) > 0 && config.ContactsDirectory[0] == '~' {
		config.ContactsDirectory = filepath.Join(homeDir, config.ContactsDirectory[1:])
	}
}
