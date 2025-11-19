package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// R2Config holds the configuration for Cloudflare R2 access.
type R2Config struct {
	AccountID       string `toml:"AccountID"`
	AccessKeyID     string `toml:"AccessKeyID"`
	SecretAccessKey string `toml:"SecretAccessKey"`
	DefaultBucket   string `toml:"DefaultBucket"`
}

const configFilePath = "~/.local/cfg/cfr2.toml"

// LoadConfig loads the R2 configuration from a TOML file or environment variables.
// TOML file takes precedence over environment variables.
func LoadConfig() (*R2Config, error) {
	cfg := &R2Config{}

	// 1. Try to load from TOML file
	expandedPath := expandPath(configFilePath)
	if _, err := os.Stat(expandedPath); err == nil {
		data, err := os.ReadFile(expandedPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file %s: %w", expandedPath, err)
		}
		if err := toml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config file %s: %w", expandedPath, err)
		}
	}

	// 2. Override with environment variables (if set and TOML didn't provide them)
	if os.Getenv("CFR2_ACCOUNT_ID") != "" {
		cfg.AccountID = os.Getenv("CFR2_ACCOUNT_ID")
	}
	if os.Getenv("CFR2_ACCESS_KEY_ID") != "" {
		cfg.AccessKeyID = os.Getenv("CFR2_ACCESS_KEY_ID")
	}
	if os.Getenv("CFR2_SECRET_ACCESS_KEY") != "" {
		cfg.SecretAccessKey = os.Getenv("CFR2_SECRET_ACCESS_KEY")
	}
	if os.Getenv("CFR2_DEFAULT_BUCKET") != "" {
		cfg.DefaultBucket = os.Getenv("CFR2_DEFAULT_BUCKET")
	}

	// 3. Validate required fields
	if cfg.AccountID == "" {
		return nil, fmt.Errorf("AccountID is not set. Please provide it in %s or via CFR2_ACCOUNT_ID environment variable", expandedPath)
	}
	if cfg.AccessKeyID == "" {
		return nil, fmt.Errorf("AccessKeyID is not set. Please provide it in %s or via CFR2_ACCESS_KEY_ID environment variable", expandedPath)
	}
	if cfg.SecretAccessKey == "" {
		return nil, fmt.Errorf("SecretAccessKey is not set. Please provide it in %s or via CFR2_SECRET_ACCESS_KEY environment variable", expandedPath)
	}
	if cfg.DefaultBucket == "" {
		return nil, fmt.Errorf("DefaultBucket is not set. Please provide it in %s or via CFR2_DEFAULT_BUCKET environment variable", expandedPath)
	}

	return cfg, nil
}

// expandPath expands a path that might contain a leading tilde (~).
func expandPath(path string) string {
	if len(path) > 0 && path[0] == '~' {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			// Fallback to original path if home directory cannot be determined
			return path
		}
		return filepath.Join(homeDir, path[1:])
	}
	return path
}
