package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"hacklab/internal/store"
)

// Session records the most recently played lab so `hacklab resume` can
// relaunch it after the TUI was closed (e.g. quit mid-challenge or the
// terminal was killed).
type Session struct {
	LabName    string    `json:"lab_name"`
	Started    time.Time `json:"started"`
	LastActive time.Time `json:"last_active"`
}

// File returns ~/.hacklab/session.json
func File() (string, error) {
	root, err := store.Home()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "session.json"), nil
}

// Load reads the saved session. It returns (nil, nil) when no session
// has been recorded yet — there is nothing to resume.
func Load() (*Session, error) {
	path, err := File()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading session: %w", err)
	}

	var s Session
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parsing session: %w", err)
	}
	if s.LabName == "" {
		return nil, nil
	}
	return &s, nil
}

// Save records labName as the current session, preserving the original
// start time across repeated resumes.
func Save(labName string) error {
	started := time.Now()
	if existing, err := Load(); err == nil && existing != nil && existing.LabName == labName {
		started = existing.Started
	}

	s := Session{
		LabName:    labName,
		Started:    started,
		LastActive: time.Now(),
	}

	path, err := File()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling session: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

// Clear wipes the session file so the last lab can no longer be resumed.
func Clear() error {
	path, err := File()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("clearing session: %w", err)
	}
	return nil
}
