package config

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/spf13/viper"
)

type Config struct {
	Providers []Provider `mapstructure:"providers"`
	Logger    Logger     `mapstructure:"logger"`
}
type Provider struct {
	Name             string           `mapstructure:"name"`
	Enabled          bool             `mapstructure:"enabled"`
	ServiceEndpoints ServiceEndpoints `mapstructure:"api_access"`
	Credentials      *Credentials     `mapstructure:"credentials"`
}

type ServiceEndpoints struct {
	IdentityAPI string `mapstructure:"identity_api"`
	ComputeAPI  string `mapstructure:"compute_api"`
}
type Credentials struct {
	OSAuthType           string `mapstructure:"os_auth_type"`
	AppCredentialsID     string `mapstructure:"app_credentials_id"`
	AppCredentialsSecret string `mapstructure:"app_credentials_secret"`
}

type Logger struct {
	Level     string `mapstructure:"level"`
	Formatter string `mapstructure:"formatter"`
}

func LoadConfig() error {
	// Load config
	setDefaults()
	viper.AutomaticEnv()

	if err := setConfigPath(); err != nil {
		return err
	}

	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %s", err))
	}

	var config Config
	err = viper.Unmarshal(&config)
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %s", err))
	}

	return err
}

// setDefaults defines the default values for the configuration.
func setDefaults() {
	viper.SetDefault("API_PORT", "8080")
	viper.SetDefault("SERVER_IP", "0.0.0.0")
	viper.SetDefault("NEO4J_HOST", "localhost")
	viper.SetDefault("NEO4J_PORT", "7687")
	viper.SetDefault("NEO4J_USER", "neo4j")
	viper.SetDefault("NEO4J_PASS", "1985ycdibiy")
	viper.SetDefault("NEO4J_PROTO", "bolt")
}

func setConfigPath() error {
	viper.SetConfigName("config") // Name of the config file (without extension)
	viper.SetConfigType("yaml")   // Type of the config file

	// Determine the directory of the current file.
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("failed to get current file path")
	}
	dir := filepath.Dir(filename)
	viper.AddConfigPath(dir)

	return nil
}
