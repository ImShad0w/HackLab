package progress

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"hacklab/internal/store"
)

// Progress tracks lab completion state
type Progress struct {
	Labs map[string]*LabProgress `json:"labs"`
}

// LabProgress tracks progress for a single lab
type LabProgress struct {
	Started    time.Time  `json:"started"`
	LastPlayed time.Time  `json:"last_played"`
	Completed  []int      `json:"completed"` // indices of completed objectives
	Attempts   map[int]int `json:"attempts"`  // objective index -> attempt count
}

// Load reads progress from disk
func Load() (*Progress, error) {
	path, err := store.ProgressFile()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Progress{
			Labs: make(map[string]*LabProgress),
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading progress: %w", err)
	}

	var p Progress
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("parsing progress: %w", err)
	}

	if p.Labs == nil {
		p.Labs = make(map[string]*LabProgress)
	}

	return &p, nil
}

// Save writes progress to disk
func (p *Progress) Save() error {
	path, err := store.ProgressFile()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling progress: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

// StartLab records the start of a lab session
func (p *Progress) StartLab(labName string) {
	if p.Labs[labName] == nil {
		p.Labs[labName] = &LabProgress{
			Started:  time.Now(),
			Attempts: make(map[int]int),
		}
	}
	p.Labs[labName].LastPlayed = time.Now()
}

// RecordAttempt increments the attempt counter
func (p *Progress) RecordAttempt(labName string, objectiveIndex int) {
	if p.Labs[labName] == nil {
		return
	}
	if p.Labs[labName].Attempts == nil {
		p.Labs[labName].Attempts = make(map[int]int)
	}
	p.Labs[labName].Attempts[objectiveIndex]++
}

// CompleteObjective marks an objective as done
func (p *Progress) CompleteObjective(labName string, objectiveIndex int) bool {
	if p.Labs[labName] == nil {
		return false
	}

	// Check if already completed
	for _, idx := range p.Labs[labName].Completed {
		if idx == objectiveIndex {
			return false // already done
		}
	}

	p.Labs[labName].Completed = append(p.Labs[labName].Completed, objectiveIndex)
	p.Labs[labName].LastPlayed = time.Now()
	return true
}

// IsCompleted checks if an objective is done
func (p *Progress) IsCompleted(labName string, objectiveIndex int) bool {
	if p.Labs[labName] == nil {
		return false
	}
	for _, idx := range p.Labs[labName].Completed {
		if idx == objectiveIndex {
			return true
		}
	}
	return false
}

// LabStats returns stats for a lab
func (p *Progress) LabStats(labName string) (completed int, totalAttempts int) {
	if p.Labs[labName] == nil {
		return 0, 0
	}
	completed = len(p.Labs[labName].Completed)
	for _, attempts := range p.Labs[labName].Attempts {
		totalAttempts += attempts
	}
	return
}
