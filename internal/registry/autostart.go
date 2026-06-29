package registry

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	registryKey = `SOFTWARE\Microsoft\Windows\CurrentVersion\Run`
	appName     = "ai-manager"
)

func IsAutoStartEnabled() (bool, error) {
	if runtime.GOOS != "windows" {
		return false, fmt.Errorf("auto-start only supported on Windows")
	}

	value, err := getRegistryValue(registryKey, appName)
	if err != nil {
		return false, err
	}

	return value != "", nil
}

func EnableAutoStart() error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("auto-start only supported on Windows")
	}

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	if err := setRegistryValue(registryKey, appName, exePath); err != nil {
		return fmt.Errorf("failed to set registry value: %w", err)
	}

	return nil
}

func DisableAutoStart() error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("auto-start only supported on Windows")
	}

	if err := deleteRegistryValue(registryKey, appName); err != nil {
		return fmt.Errorf("failed to delete registry value: %w", err)
	}

	return nil
}

func getRegistryValue(key, valueName string) (string, error) {
	cmd := exec.Command("reg", "query", key, "/v", valueName)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.Contains(line, valueName) {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				return parts[len(parts)-1], nil
			}
		}
	}

	return "", fmt.Errorf("value not found")
}

func setRegistryValue(key, valueName, value string) error {
	cmd := exec.Command("reg", "add", key, "/v", valueName, "/t", "REG_SZ", "/d", value, "/f")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("reg add failed: %s: %w", string(out), err)
	}

	return nil
}

func deleteRegistryValue(key, valueName string) error {
	cmd := exec.Command("reg", "delete", key, "/v", valueName, "/f")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("reg delete failed: %s: %w", string(out), err)
	}

	return nil
}

func GetExePath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Abs(exe)
}
