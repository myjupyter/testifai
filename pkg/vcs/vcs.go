package vcs

import (
	"fmt"
	"os/exec"
	"strings"
)

func GetGitRemoteURL() (string, error) {
	cmd := exec.Command("git", "config", "--get", "remote.origin.url")

	output, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() != 0 {
			return "", fmt.Errorf("git command failed (Exit Code %d). Check if you are in a git repository or if 'origin' is set. Error: %w", exitError.ExitCode(), err)
		}
		return "", fmt.Errorf("failed to execute git command: %w", err)
	}
	url := strings.TrimSpace(string(output))

	if url == "" {
		return "", fmt.Errorf("git remote URL not found (output was empty). Ensure 'remote.origin.url' is configured")
	}

	return url, nil
}
