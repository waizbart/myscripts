package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const stateFileName = ".bootstrap-state.json"

type State struct {
	Config        *Config `json:"config"`
	CompletedStep int     `json:"completed_step"` // -1 means no step completed yet
}

// statePath returns the path to the state file inside the project directory.
func statePath(projectDir string) string {
	return filepath.Join(projectDir, stateFileName)
}

// SaveState writes the current config and completed step index to disk.
func SaveState(cfg *Config, completedStep int) error {
	s := &State{Config: cfg, CompletedStep: completedStep}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}
	path := statePath(cfg.ProjectDir)
	os.MkdirAll(filepath.Dir(path), 0755)
	return os.WriteFile(path, data, 0644)
}

// LoadState tries to load a previously saved state from the given directory.
// Returns nil if no state file exists.
func LoadState(projectDir string) (*State, error) {
	path := statePath(projectDir)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("failed to parse state file: %w", err)
	}
	return &s, nil
}

// ClearState removes the state file after successful completion.
func ClearState(projectDir string) {
	os.Remove(statePath(projectDir))
}
