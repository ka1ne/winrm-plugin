package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/harness/winrm-plugin/internal/logger"
	"github.com/harness/winrm-plugin/internal/winrm"
)

// Config holds all plugin configuration
type Config struct {
	// Connection settings
	Host     string
	Port     int
	Username string
	Password string
	Protocol string
	AuthType string
	Timeout  time.Duration

	// Execution settings
	Command        string
	ScriptPath     string
	ScriptContent  string
	WorkingDir     string
	PersistSession bool
	StreamOutput   bool

	// Retry configuration
	MaxRetries    int
	RetryInterval time.Duration

	// Logging configuration
	LogLevel  string
	LogFormat string
	Debug     bool

	// Proxy settings
	HTTPProxy  string
	HTTPSProxy string
	NoProxy    string
}

func main() {
	config, err := parseConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log := logger.New(config.LogLevel, config.LogFormat)

	log.Info("Starting WinRM Plugin")
	log.Debugf("Configuration: Host=%s, Port=%d, Protocol=%s, AuthType=%s",
		config.Host, config.Port, config.Protocol, config.AuthType)

	// Create WinRM client
	client, err := winrm.NewClient(config.Host, config.Port, config.Username, config.Password, &winrm.Options{
		Protocol:      config.Protocol,
		AuthType:      config.AuthType,
		Timeout:       config.Timeout,
		MaxRetries:    config.MaxRetries,
		RetryInterval: config.RetryInterval,
		HTTPProxy:     config.HTTPProxy,
		HTTPSProxy:    config.HTTPSProxy,
		NoProxy:       config.NoProxy,
		Logger:        log,
	})
	if err != nil {
		log.Errorf("Failed to create WinRM client: %v", err)
		os.Exit(1)
	}

	// Execute command, script file, or script content
	var result *winrm.ExecuteResult

	execOptions := &winrm.ExecuteOptions{
		WorkingDir:     config.WorkingDir,
		PersistSession: config.PersistSession,
		StreamOutput:   config.StreamOutput,
	}

	if config.Command != "" {
		log.Debugf("Executing command: %s", config.Command)
		result, err = client.ExecuteCommand(config.Command, execOptions)
	} else if config.ScriptPath != "" {
		log.Debugf("Executing script file: %s", config.ScriptPath)
		result, err = client.ExecuteScriptFile(config.ScriptPath, execOptions)
	} else if config.ScriptContent != "" {
		log.Debug("Executing script content")
		result, err = client.ExecuteScript(config.ScriptContent, execOptions)
	} else {
		log.Error("No command, script file, or script content specified")
		log.Info("Use PLUGIN_COMMAND, PLUGIN_SCRIPT_PATH, or PLUGIN_SCRIPT_CONTENT")
		os.Exit(1)
	}

	if err != nil {
		log.Errorf("Execution failed: %v", err)
		os.Exit(1)
	}

	log.Infof("Execution completed successfully. Exit code: %d", result.ExitCode)
	if result.Stdout != "" {
		log.Info("STDOUT:")
		fmt.Print(result.Stdout)
	}
	if result.Stderr != "" {
		log.Warn("STDERR:")
		fmt.Print(result.Stderr)
	}

	if result.ExitCode != 0 {
		os.Exit(result.ExitCode)
	}

	log.Info("WinRM Plugin completed successfully")
}

func parseConfig() (*Config, error) {
	config := &Config{
		// Defaults
		Port:          5986,
		Protocol:      "https",
		AuthType:      "ntlm",
		Timeout:       30 * time.Second,
		MaxRetries:    3,
		RetryInterval: 5 * time.Second,
		LogLevel:      "info",
		LogFormat:     "text",
		WorkingDir:    "C:\\",
	}

	// Required fields
	config.Host = os.Getenv("PLUGIN_HOST")
	if config.Host == "" {
		return nil, fmt.Errorf("PLUGIN_HOST is required")
	}

	config.Username = os.Getenv("PLUGIN_USERNAME")
	if config.Username == "" {
		return nil, fmt.Errorf("PLUGIN_USERNAME is required")
	}

	config.Password = os.Getenv("PLUGIN_PASSWORD")
	if config.Password == "" {
		return nil, fmt.Errorf("PLUGIN_PASSWORD is required")
	}

	config.Command = os.Getenv("PLUGIN_COMMAND")
	config.ScriptPath = os.Getenv("PLUGIN_SCRIPT_PATH")
	config.ScriptContent = os.Getenv("PLUGIN_SCRIPT_CONTENT")

	// Optional fields with defaults
	if port := os.Getenv("PLUGIN_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Port = p
		}
	}

	if protocol := os.Getenv("PLUGIN_PROTOCOL"); protocol != "" {
		config.Protocol = strings.ToLower(protocol)
	}

	if authType := os.Getenv("PLUGIN_AUTH_TYPE"); authType != "" {
		config.AuthType = strings.ToLower(authType)
	}

	if timeout := os.Getenv("PLUGIN_TIMEOUT"); timeout != "" {
		if t, err := strconv.Atoi(timeout); err == nil {
			config.Timeout = time.Duration(t) * time.Second
		}
	}

	if workingDir := os.Getenv("PLUGIN_WORKING_DIR"); workingDir != "" {
		config.WorkingDir = workingDir
	}

	if maxRetries := os.Getenv("PLUGIN_MAX_RETRIES"); maxRetries != "" {
		if r, err := strconv.Atoi(maxRetries); err == nil {
			config.MaxRetries = r
		}
	}

	if retryInterval := os.Getenv("PLUGIN_RETRY_INTERVAL"); retryInterval != "" {
		if r, err := strconv.Atoi(retryInterval); err == nil {
			config.RetryInterval = time.Duration(r) * time.Second
		}
	}

	if logLevel := os.Getenv("PLUGIN_LOG_LEVEL"); logLevel != "" {
		config.LogLevel = strings.ToLower(logLevel)
	}

	if logFormat := os.Getenv("PLUGIN_LOG_FORMAT"); logFormat != "" {
		config.LogFormat = strings.ToLower(logFormat)
	}

	if debug := os.Getenv("PLUGIN_DEBUG_MODE"); debug == "true" {
		config.Debug = true
		config.LogLevel = "debug"
	}

	if persistSession := os.Getenv("PLUGIN_PERSIST_SESSION"); persistSession == "true" {
		config.PersistSession = true
	}

	if streamOutput := os.Getenv("PLUGIN_STREAM_OUTPUT"); streamOutput == "true" {
		config.StreamOutput = true
	}

	// Proxy settings
	config.HTTPProxy = os.Getenv("HTTP_PROXY")
	config.HTTPSProxy = os.Getenv("HTTPS_PROXY")
	config.NoProxy = os.Getenv("NO_PROXY")

	return config, nil
}
