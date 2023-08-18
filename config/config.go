package config

import (
	"fmt"

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
	viper.SetDefault("API_PORT", "8080")
	viper.SetDefault("NEO4J_HOST", "localhost")
	viper.SetDefault("NEO4J_PORT", "7687")
	viper.SetDefault("NEO4J_USER", "neo4j")
	viper.SetDefault("NEO4J_PASS", "testingshit")
	viper.SetDefault("NEO4J_PROTO", "bolt")

	viper.SetConfigName("config") // Name of the config file (without extension)
	viper.SetConfigType("yaml")   // Type of the config file
	viper.AddConfigPath("config") // Look for the config in the working directory
	err := viper.ReadInConfig()   // Find and read the config file
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	viper.AutomaticEnv()
	var config Config
	err = viper.Unmarshal(&config)
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	return err
}
