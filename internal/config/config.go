package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds the application configuration.
type Config struct {
	IndexUpdate  bool
	GithubUpdate bool
	Path         string
	Color        bool
}

const (
	configDirName  = ".red-tldr"
	configFileName = "config"
	configFileType = "toml"
	databaseDir    = "red-tldr-db"
	databaseFile   = "db" + string(os.PathSeparator) + "db.json"
)

// Load reads the configuration from disk, generating a default if absent.
func Load() *Config {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot determine home directory: %v\n", err)
		os.Exit(1)
	}

	configDir := filepath.Join(homeDir, configDirName)

	viper.SetConfigName(configFileName)
	viper.SetConfigType(configFileType)
	viper.AddConfigPath(configDir)

	if err := viper.ReadInConfig(); err != nil {
		generateConfig(configDir, homeDir)
	}

	return &Config{
		IndexUpdate:  viper.GetBool("red-tldr.index-update"),
		GithubUpdate: viper.GetBool("red-tldr.github-update"),
		Path:         viper.GetString("red-tldr.path"),
		Color:        viper.GetBool("red-tldr.color"),
	}
}

// GetDatabasePath returns the root directory of the database.
func (c *Config) GetDatabasePath() string {
	if c.Path != "" {
		return c.Path
	}
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, databaseDir)
}

// GetDatabaseFilePath returns the full path to the JSON index file.
func (c *Config) GetDatabaseFilePath() string {
	return filepath.Join(c.GetDatabasePath(), databaseFile)
}

// DatabaseExists checks whether the database index file exists on disk.
func DatabaseExists(c *Config) bool {
	_, err := os.Stat(c.GetDatabaseFilePath())
	return err == nil
}

func generateConfig(configDir, homeDir string) {
	if err := os.MkdirAll(configDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot create config directory: %v\n", err)
		return
	}

	viper.SetConfigName(configFileName)
	viper.SetConfigType(configFileType)
	viper.AddConfigPath(configDir)

	viper.SetDefault("red-tldr.path", filepath.Join(homeDir, databaseDir))
	viper.SetDefault("red-tldr.index-update", false)
	viper.SetDefault("red-tldr.github-update", false)
	viper.SetDefault("red-tldr.color", true)

	configFile := filepath.Join(configDir, configFileName+"."+configFileType)
	if err := viper.WriteConfigAs(configFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot write config: %v\n", err)
		return
	}
	log.Printf("[Config generated: %s]\n", configFile)
}
