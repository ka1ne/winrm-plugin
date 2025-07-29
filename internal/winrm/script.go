package winrm

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf16"
)

// ExecuteScriptFile executes a PowerShell script from a local file
func (c *Client) ExecuteScriptFile(scriptPath string, opts *ExecuteOptions) (*ExecuteResult, error) {
	if opts == nil {
		opts = &ExecuteOptions{}
	}

	c.logger.Debugf("Executing script file: %s", scriptPath)

	// Check if file exists
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("script file not found: %s", scriptPath)
	}

	// Read script content
	scriptContent, err := os.ReadFile(scriptPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read script file: %v", err)
	}

	// Validate it's a PowerShell script
	ext := strings.ToLower(filepath.Ext(scriptPath))
	if ext != ".ps1" {
		c.logger.Warnf("Script file '%s' does not have .ps1 extension", scriptPath)
	}

	c.logger.Debugf("Script content length: %d characters", len(scriptContent))

	// Execute the script content
	return c.ExecuteScript(string(scriptContent), opts)
}

// ExecuteScript executes PowerShell script content
func (c *Client) ExecuteScript(scriptContent string, opts *ExecuteOptions) (*ExecuteResult, error) {
	if opts == nil {
		opts = &ExecuteOptions{}
	}

	c.logger.Debugf("Executing script content (%d characters)", len(scriptContent))

	// Prepare script execution command
	// Use -EncodedCommand for better handling of special characters and multi-line scripts
	encodedScript := encodeToBase64(scriptContent)
	command := fmt.Sprintf("powershell.exe -NonInteractive -EncodedCommand %s", encodedScript)

	// Execute using the command execution path
	return c.executeInternal(command, opts)
}

// executeInternal is the common execution path for both commands and scripts
func (c *Client) executeInternal(command string, opts *ExecuteOptions) (*ExecuteResult, error) {
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
		c.logger.Debugf("Execution attempt %d/%d", attempt, c.opts.MaxRetries)

		if opts.StreamOutput {
			// For streaming, we would implement real-time output capture here
			// For now, fall back to buffered execution
			c.logger.Debug("Streaming output not yet fully implemented, using buffered mode")
		}

		stdout, stderr, exitCode, err = c.client.RunPSWithString(fullCommand, "")
		if err == nil {
			break
		}

		c.logger.Warnf("Execution attempt %d failed: %v", attempt, err)
		if attempt < c.opts.MaxRetries {
			c.logger.Debugf("Retrying in %v", c.opts.RetryInterval)
			time.Sleep(c.opts.RetryInterval)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("execution failed after %d attempts: %v", c.opts.MaxRetries, err)
	}

	result := &ExecuteResult{
		ExitCode: exitCode,
		Stdout:   stdout,
		Stderr:   stderr,
	}

	c.logger.Debugf("Execution completed with exit code: %d", exitCode)
	return result, nil
}

// encodeToBase64 encodes a PowerShell script to base64 for -EncodedCommand
func encodeToBase64(script string) string {
	// Convert to UTF-16LE (required by PowerShell -EncodedCommand)
	utf16le := utf16.Encode([]rune(script))
	bytes := make([]byte, len(utf16le)*2)
	for i, r := range utf16le {
		bytes[i*2] = byte(r)
		bytes[i*2+1] = byte(r >> 8)
	}
	return base64.StdEncoding.EncodeToString(bytes)
}
