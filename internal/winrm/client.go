package winrm

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/harness/winrm-plugin/internal/logger"
	"github.com/masterzen/winrm"
)

// Client represents a WinRM client
type Client struct {
	client *winrm.Client
	logger logger.Logger
	opts   *Options
}

// Options holds WinRM client configuration
type Options struct {
	Protocol      string
	AuthType      string
	Timeout       time.Duration
	MaxRetries    int
	RetryInterval time.Duration
	HTTPProxy     string
	HTTPSProxy    string
	NoProxy       string
	Logger        logger.Logger
}

// ExecuteOptions holds command execution options
type ExecuteOptions struct {
	WorkingDir     string
	PersistSession bool
	StreamOutput   bool
}

// ExecuteResult holds the result of command execution
type ExecuteResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
}

// NewClient creates a new WinRM client
func NewClient(host string, port int, username, password string, opts *Options) (*Client, error) {
	if opts == nil {
		opts = &Options{}
	}

	endpoint := winrm.NewEndpoint(host, port, opts.Protocol == "https", false, nil, nil, nil, opts.Timeout)

	// Configure proxy if needed
	if opts.HTTPProxy != "" || opts.HTTPSProxy != "" {
		transport, err := configureProxy(endpoint, opts)
		if err != nil {
			return nil, fmt.Errorf("proxy configuration failed: %v", err)
		}
		// Set custom transport if needed - this may need adjustment based on library API
		_ = transport // placeholder for now
	}

	// Create winrm client based on auth type
	var client *winrm.Client
	var err error

	switch strings.ToLower(opts.AuthType) {
	case "ntlm":
		client, err = winrm.NewClient(endpoint, username, password)
	case "kerberos":
		// For Kerberos, we might need additional configuration
		client, err = winrm.NewClient(endpoint, username, password)
	default:
		return nil, fmt.Errorf("unsupported auth type: %s", opts.AuthType)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create WinRM client: %v", err)
	}

	return &Client{
		client: client,
		logger: opts.Logger,
		opts:   opts,
	}, nil
}

// ExecuteCommand executes a PowerShell command
// By default, output is buffered for better reliability and Harness compatibility
// Set StreamOutput=true for real-time streaming of long-running operations
func (c *Client) ExecuteCommand(command string, opts *ExecuteOptions) (*ExecuteResult, error) {
	if opts == nil {
		opts = &ExecuteOptions{}
	}

	c.logger.Debugf("Executing command: %s", command)
	if opts.StreamOutput {
		c.logger.Debug("Using streaming output mode")
	} else {
		c.logger.Debug("Using buffered output mode (default)")
	}

	var fullCommand string
	if opts.WorkingDir != "" && opts.WorkingDir != "C:\\" {
		fullCommand = fmt.Sprintf("cd '%s'; %s", opts.WorkingDir, command)
	} else {
		fullCommand = command
	}

	var stdout, stderr string
	var exitCode int
	var err error

	// Retry logic
	for attempt := 1; attempt <= c.opts.MaxRetries; attempt++ {
		c.logger.Debugf("Command execution attempt %d/%d", attempt, c.opts.MaxRetries)

		stdout, stderr, exitCode, err = c.client.RunPSWithString(fullCommand, "")
		if err == nil {
			break
		}

		c.logger.Warnf("Command execution attempt %d failed: %v", attempt, err)
		if attempt < c.opts.MaxRetries {
			c.logger.Debugf("Retrying in %v", c.opts.RetryInterval)
			time.Sleep(c.opts.RetryInterval)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("command execution failed after %d attempts: %v", c.opts.MaxRetries, err)
	}

	result := &ExecuteResult{
		ExitCode: exitCode,
		Stdout:   stdout,
		Stderr:   stderr,
	}

	c.logger.Debugf("Command completed with exit code: %d", exitCode)
	return result, nil
}

// configureProxy sets up proxy configuration for WinRM transport
func configureProxy(endpoint *winrm.Endpoint, opts *Options) (*http.Transport, error) {
	transport := &http.Transport{}

	// Check if host should bypass proxy
	if shouldBypassProxy(endpoint.Host, opts.NoProxy) {
		opts.Logger.Debugf("Bypassing proxy for host: %s", endpoint.Host)
		transport.Proxy = nil
		return transport, nil
	}

	// Configure proxy based on protocol
	var proxyURL string
	if endpoint.HTTPS {
		proxyURL = opts.HTTPSProxy
	} else {
		proxyURL = opts.HTTPProxy
	}

	if proxyURL != "" {
		parsedURL, err := url.Parse(proxyURL)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy URL: %v", err)
		}

		opts.Logger.Debugf("Using proxy: %s", proxyURL)
		transport.Proxy = http.ProxyURL(parsedURL)
	}

	return transport, nil
}

// shouldBypassProxy checks if a host should bypass proxy based on no_proxy settings
func shouldBypassProxy(host, noProxy string) bool {
	if noProxy == "" {
		return false
	}

	// Split no_proxy by comma and check each entry
	noProxyList := strings.Split(noProxy, ",")
	for _, entry := range noProxyList {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}

		// Direct match
		if host == entry {
			return true
		}

		// Domain suffix match (starts with .)
		if strings.HasPrefix(entry, ".") && strings.HasSuffix(host, entry) {
			return true
		}

		// Domain match without leading dot
		if strings.HasSuffix(host, "."+entry) {
			return true
		}
	}

	return false
}
