// Package config provides configuration management for vmctl
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds the vmctl configuration
type Config struct {
	// Connection settings
	DefaultURI string `yaml:"default_uri"` // e.g., "qemu:///system" or "qemu+ssh://user@host/system"

	// Output settings
	Output struct {
		Format     string `yaml:"format"`     // table, json, yaml
		NoColor    bool   `yaml:"no_color"`
		WideMode   bool   `yaml:"wide_mode"`
		SortBy     string `yaml:"sort_by"`     // name, state, cpu, memory
		SortOrder  string `yaml:"sort_order"`  // asc, desc
	} `yaml:"output"`

	// Display settings
	Display struct {
		ShowIPs      bool `yaml:"show_ips"`
		ShowMACs     bool `yaml:"show_macs"`
		ShowUptime   bool `yaml:"show_uptime"`
		ShowAutostart bool `yaml:"show_autostart"`
		Icons        bool `yaml:"icons"`       // Use emoji icons for states
	} `yaml:"display"`

	// Monitoring defaults
	Monitor struct {
		RefreshInterval int `yaml:"refresh_interval"` // seconds
		CPUHistory      int `yaml:"cpu_history"`      // number of data points to show
		MemHistory      int `yaml:"mem_history"`
	} `yaml:"monitor"`

	// Debug settings
	Debug struct {
		OutputDir     string `yaml:"output_dir"`
		CollectLogs   bool   `yaml:"collect_logs"`
		CollectNetwork bool  `yaml:"collect_network"`
		CollectStorage bool  `yaml:"collect_storage"`
		LogLines      int    `yaml:"log_lines"` // number of log lines to collect
	} `yaml:"debug"`

	// Aliases for common commands
	Aliases map[string]string `yaml:"aliases"`
}

var (
	config *Config
	configPath string
)

// InitConfig initializes the configuration from file or defaults
func InitConfig(cfgFile string) {
	config = &Config{}
	
	// Set default values
	setDefaults()

	// Determine config file path
	if cfgFile != "" {
		configPath = cfgFile
	} else {
		home, err := os.UserHomeDir()
		if err == nil {
			configPath = filepath.Join(home, ".vmctl.yaml")
		}
	}

	// Load config file if it exists
	if configPath != "" {
		if _, err := os.Stat(configPath); err == nil {
			loadConfig(configPath)
		}
	}
}

func setDefaults() {
	config.DefaultURI = "qemu:///system"
	config.Output.Format = "table"
	config.Output.NoColor = false
	config.Output.WideMode = false
	config.Output.SortBy = "name"
	config.Output.SortOrder = "asc"
	
	config.Display.ShowIPs = true
	config.Display.ShowMACs = false
	config.Display.ShowUptime = true
	config.Display.ShowAutostart = false
	config.Display.Icons = true
	
	config.Monitor.RefreshInterval = 2
	config.Monitor.CPUHistory = 60
	config.MemHistory = 60
	
	config.Debug.OutputDir = "/tmp/vmctl-debug"
	config.Debug.CollectLogs = true
	config.Debug.CollectNetwork = true
	config.Debug.CollectStorage = true
	config.Debug.LogLines = 100

	// Default aliases
	config.Aliases = map[string]string{
		"ls":   "list",
		"ps":   "list",
		"rm":   "destroy",
		"info": "describe",
	}
}

func loadConfig(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to parse config file: %v\n", err)
		return
	}
	
	// Merge with defaults (non-zero values from file override defaults)
	if cfg.DefaultURI != "" {
		config.DefaultURI = cfg.DefaultURI
	}
	if cfg.Output.Format != "" {
		config.Output.Format = cfg.Output.Format
	}
	if cfg.Output.NoColor {
		config.Output.NoColor = cfg.Output.NoColor
	}
	if cfg.Monitor.RefreshInterval > 0 {
		config.Monitor.RefreshInterval = cfg.Monitor.RefreshInterval
	}
	if len(cfg.Aliases) > 0 {
		for k, v := range cfg.Aliases {
			config.Aliases[k] = v
		}
	}
}

// Get returns the current configuration
func Get() *Config {
	return config
}

// GetConfigPath returns the path to the config file
func GetConfigPath() string {
	return configPath
}

// SaveDefaultConfig creates a default config file at the default location
func SaveDefaultConfig() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}

	path := filepath.Join(home, ".vmctl.yaml")
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	header := "# vmctl Configuration File\n# https://github.com/SimoneChee/vmctl\n\n"
	data = append([]byte(header), data...)

	return os.WriteFile(path, data, 0644)
}
