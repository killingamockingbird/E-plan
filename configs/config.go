package configs

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config represents the global configuration tree root.
type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	MongoDB    MongoConfig      `mapstructure:"mongodb"`
	LLM        LLMConfig        `mapstructure:"llm"`
	WeatherAPI WeatherAPIConfig `mapstructure:"weather_api"`
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
	Mode string `mapstructure:"mode"` // e.g., "development", "production"
}

type MongoConfig struct {
	URI            string `mapstructure:"uri"`
	Database       string `mapstructure:"database"`
	TimeoutSeconds int    `mapstructure:"timeout_seconds"`
}

type LLMConfig struct {
	BaseURL     string  `mapstructure:"base_url"`
	APIKey      string  `mapstructure:"api_key"`
	Model       string  `mapstructure:"model"`
	Temperature float32 `mapstructure:"temperature"`
}

type WeatherAPIConfig struct {
	BaseURL string `mapstructure:"base_url"`
	APIKey  string `mapstructure:"api_key"`
}

// LoadConfig reads the configuration file and binds it to the Config struct.
func LoadConfig(configPath string) (*Config, error) {
	v := viper.New()

	// 1. Set configuration file naming and search paths.
	v.SetConfigName("config")   // File name (without extension)
	v.SetConfigType("yaml")     // File format
	v.AddConfigPath(configPath) // Search path (e.g., "./configs")

	// 2. Enable Environment Variable support.
	// This is critical for 12-factor app compliance and secure production deployments.
	// For example: the environment variable LLM_API_KEY will override llm.api_key in the YAML.
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 3. Read the configuration file.
	if err := v.ReadInConfig(); err != nil {
		// Allow the application to start without a config file if environment variables are used.
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// 4. Unmarshal the configuration into the struct.
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config structure: %w", err)
	}

	// 5. Basic validation for mandatory fields.
	if cfg.LLM.APIKey == "" {
		fmt.Println("WARNING: LLM API Key is missing. Check your config file or set the LLM_API_KEY environment variable.")
	}

	return &cfg, nil
}
