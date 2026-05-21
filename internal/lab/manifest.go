package lab

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Manifest is the lab.yml structure
type Manifest struct {
	Name        string       `yaml:"name"`
	Version     string       `yaml:"version"`
	Description string       `yaml:"description"`
	Difficulty  string       `yaml:"difficulty"`
	Author      string       `yaml:"author"`
	Tags        []string     `yaml:"tags"`

	// Single container setup
	Image string `yaml:"image"`
	Port  int    `yaml:"port"`

	// Docker Compose setup
	ComposeFile string `yaml:"compose_file"`

	// Health check
	WaitFor  string `yaml:"wait_for"`  // URL to poll for readiness
	WaitSecs int    `yaml:"wait_secs"` // max seconds to wait (default: 30)

	// Objectives (the challenge content)
	Objectives []Objective `yaml:"objectives"`

	// Browser behavior
	OpenBrowser bool `yaml:"open_browser"` // default: ask user
}

// Objective represents a single challenge in a lab
type Objective struct {
	Name     string   `yaml:"name"`
	Category string   `yaml:"category"`
	Hint     string   `yaml:"hint"`
	Hints    []string `yaml:"hints"` // progressive hints
}

// LoadManifest reads and validates a lab.yml file
func LoadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading lab.yml: %w", err)
	}

	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing lab.yml: %w", err)
	}

	// Validate
	if m.Name == "" {
		return nil, fmt.Errorf("lab.yml: missing 'name' field")
	}
	if m.Image == "" && m.ComposeFile == "" {
		return nil, fmt.Errorf("lab.yml: must specify 'image' or 'compose_file'")
	}
	if len(m.Objectives) == 0 {
		return nil, fmt.Errorf("lab.yml: no objectives defined")
	}
	if m.WaitSecs <= 0 {
		m.WaitSecs = 30
	}

	return &m, nil
}

// Lab represents a loaded lab with its metadata and path
type Lab struct {
	Name     string
	Path     string
	Manifest *Manifest
}

// LoadLab loads a lab from its directory
func LoadLab(dir string) (*Lab, error) {
	manifestPath := filepath.Join(dir, "lab.yml")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("no lab.yml found in %s", dir)
	}

	m, err := LoadManifest(manifestPath)
	if err != nil {
		return nil, err
	}

	return &Lab{
		Name:     filepath.Base(dir),
		Path:     dir,
		Manifest: m,
	}, nil
}

// DiscoverLabs finds all labs in the labs directory
func DiscoverLabs(labsDir string) ([]Lab, error) {
	entries, err := os.ReadDir(labsDir)
	if err != nil {
		return nil, fmt.Errorf("reading labs dir: %w", err)
	}

	var labs []Lab
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		labPath := filepath.Join(labsDir, e.Name())
		lab, err := LoadLab(labPath)
		if err != nil {
			continue // skip invalid labs
		}
		labs = append(labs, *lab)
	}
	return labs, nil
}
