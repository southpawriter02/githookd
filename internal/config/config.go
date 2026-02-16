package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// LogLevel represents the severity level for logging.
type LogLevel int

const (
	LogDebug LogLevel = iota
	LogInfo
	LogWarn
	LogError
)

// DefaultTimeout is the timeout applied when no timeout is specified at any level.
const DefaultTimeout = 30 * time.Second

// Config represents the main configuration structure from .githooksrc.yml
type Config struct {
	Timeout  string                   `yaml:"timeout"`
	LogLevel string                   `yaml:"log_level"`
	Hooks    map[string][]HookCommand `yaml:"hooks"`
}

// HookCommand represents a single command to be executed for a hook.
type HookCommand struct {
	Run         string `yaml:"run"`
	Description string `yaml:"description"`
	Enabled     *bool  `yaml:"enabled,omitempty"`
	Timeout     string `yaml:"timeout,omitempty"`
	LogLevel    string `yaml:"log_level,omitempty"`
}

// IsEnabled returns true if the command is enabled (nil defaults to true).
func (hc HookCommand) IsEnabled() bool {
	if hc.Enabled == nil {
		return true
	}
	return *hc.Enabled
}

// BoolPtr returns a pointer to the given bool value.
func BoolPtr(b bool) *bool {
	return &b
}

// ResolvedConfig holds validated, runtime-ready configuration.
type ResolvedConfig struct {
	Timeout  time.Duration
	LogLevel LogLevel
	Hooks    map[string][]ResolvedHookCommand
}

// ResolvedHookCommand holds a fully resolved command ready for execution.
type ResolvedHookCommand struct {
	Run         string
	Description string
	Timeout     time.Duration // 0 means no timeout (only via "none")
	LogLevel    LogLevel
	Enabled     bool
}

// Load reads the configuration file from the given path and returns a Config struct.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

// Save writes a Config to the given file path as YAML.
func Save(path string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Resolve validates and resolves a raw Config into a ResolvedConfig.
// Returns ALL validation errors, not just the first.
func (c *Config) Resolve() (*ResolvedConfig, []error) {
	var errs []error

	// Resolve global timeout
	globalTimeout := DefaultTimeout
	if c.Timeout != "" {
		d, err := time.ParseDuration(c.Timeout)
		if err != nil {
			errs = append(errs, fmt.Errorf("invalid global timeout %q: %w", c.Timeout, err))
		} else if d == 0 {
			errs = append(errs, fmt.Errorf("invalid global timeout \"0s\": use \"none\" on individual commands to disable timeout"))
		} else if d < 0 {
			errs = append(errs, fmt.Errorf("invalid global timeout %q: must be positive", c.Timeout))
		} else {
			globalTimeout = d
		}
	}

	// Resolve global log level
	globalLogLevel := LogWarn
	if c.LogLevel != "" {
		ll, err := parseLogLevel(c.LogLevel)
		if err != nil {
			errs = append(errs, fmt.Errorf("invalid global log_level %q: valid levels are debug, info, warn, error", c.LogLevel))
		} else {
			globalLogLevel = ll
		}
	}

	// Resolve hooks
	resolvedHooks := make(map[string][]ResolvedHookCommand)

	for hookName, commands := range c.Hooks {
		// Validate hook name
		if err := ValidateHookName(hookName); err != nil {
			suggestion := SuggestHookName(hookName)
			msg := fmt.Sprintf("invalid hook name %q: not a recognized Git hook name", hookName)
			if suggestion != "" {
				msg += fmt.Sprintf("; did you mean %q?", suggestion)
			}
			errs = append(errs, fmt.Errorf("%s", msg))
			continue
		}

		var resolved []ResolvedHookCommand
		for i, cmd := range commands {
			// Validate run field
			if strings.TrimSpace(cmd.Run) == "" {
				errs = append(errs, fmt.Errorf("hook %q command #%d: 'run' field is required but missing or empty", hookName, i+1))
				continue
			}

			// Resolve timeout
			cmdTimeout := globalTimeout
			if cmd.Timeout != "" {
				if strings.ToLower(cmd.Timeout) == "none" {
					cmdTimeout = 0
				} else {
					d, err := time.ParseDuration(cmd.Timeout)
					if err != nil {
						errs = append(errs, fmt.Errorf("hook %q command #%d: invalid timeout %q: %w", hookName, i+1, cmd.Timeout, err))
						continue
					}
					if d == 0 {
						errs = append(errs, fmt.Errorf("hook %q command #%d: invalid timeout \"0s\": use \"none\" to disable timeout", hookName, i+1))
						continue
					}
					if d < 0 {
						errs = append(errs, fmt.Errorf("hook %q command #%d: invalid timeout %q: must be positive", hookName, i+1, cmd.Timeout))
						continue
					}
					cmdTimeout = d
				}
			}

			// Resolve log level
			cmdLogLevel := globalLogLevel
			if cmd.LogLevel != "" {
				ll, err := parseLogLevel(cmd.LogLevel)
				if err != nil {
					errs = append(errs, fmt.Errorf("hook %q command #%d: invalid log_level %q: valid levels are debug, info, warn, error", hookName, i+1, cmd.LogLevel))
					continue
				}
				cmdLogLevel = ll
			}

			resolved = append(resolved, ResolvedHookCommand{
				Run:         cmd.Run,
				Description: cmd.Description,
				Timeout:     cmdTimeout,
				LogLevel:    cmdLogLevel,
				Enabled:     cmd.IsEnabled(),
			})
		}

		if len(resolved) > 0 {
			resolvedHooks[hookName] = resolved
		}
	}

	if len(errs) > 0 {
		return nil, errs
	}

	return &ResolvedConfig{
		Timeout:  globalTimeout,
		LogLevel: globalLogLevel,
		Hooks:    resolvedHooks,
	}, nil
}

// parseLogLevel converts a string log level to the LogLevel type.
func parseLogLevel(s string) (LogLevel, error) {
	switch strings.ToLower(s) {
	case "debug":
		return LogDebug, nil
	case "info":
		return LogInfo, nil
	case "warn":
		return LogWarn, nil
	case "error":
		return LogError, nil
	default:
		return LogWarn, fmt.Errorf("unknown log level: %s", s)
	}
}
