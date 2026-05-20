package store

import (
	"fmt"
	"os"
	"path/filepath"
)

// Home returns ~/.hacklab
func Home() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("finding home dir: %w", err)
	}
	return filepath.Join(home, ".hacklab"), nil
}

// LabsDir returns ~/.hacklab/labs
func LabsDir() (string, error) {
	root, err := Home()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "labs"), nil
}

// ProgressFile returns ~/.hacklab/progress.json
func ProgressFile() (string, error) {
	root, err := Home()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "progress.json"), nil
}

// Ensure creates the directory structure
func Ensure() error {
	root, err := Home()
	if err != nil {
		return err
	}
	labsDir := filepath.Join(root, "labs")
	return os.MkdirAll(labsDir, 0755)
}

// LabPath returns the path for a specific lab
func LabPath(name string) (string, error) {
	labsDir, err := LabsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(labsDir, name), nil
}

// LabExists checks if a lab exists
func LabExists(name string) (bool, error) {
	path, err := LabPath(name)
	if err != nil {
		return false, err
	}
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}

// CreateLabDir creates a lab directory
func CreateLabDir(name string) (string, error) {
	if err := Ensure(); err != nil {
		return "", err
	}
	path, err := LabPath(name)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(path, 0755); err != nil {
		return "", fmt.Errorf("creating lab dir: %w", err)
	}
	return path, nil
}

// RemoveLabDir removes a lab directory
func RemoveLabDir(name string) error {
	path, err := LabPath(name)
	if err != nil {
		return err
	}
	return os.RemoveAll(path)
}
