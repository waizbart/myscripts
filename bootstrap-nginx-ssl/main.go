package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type step struct {
	name string
	fn   func(*Config, Executor) error
}

func main() {
	fmt.Println("=== Bootstrap: Ubuntu + Nginx + Docker + Certbot ===")
	fmt.Println()

	steps := []step{
		{"Clone repositories", SetupRepos},
		{"Prepare database", SetupDatabase},
		{"Install Docker", SetupDocker},
		{"Generate docker-compose & start containers", GenerateCompose},
		{"Configure Nginx reverse proxy", SetupNginx},
		{"Install Certbot & issue SSL certificates", SetupCertbot},
	}

	var cfg *Config
	startStep := 0

	// Check for a saved state to resume from.
	// We need to peek at a project directory — try common locations or prompt.
	cfg, startStep = tryResume(steps)

	if cfg == nil {
		// Fresh run — gather config interactively.
		var err error
		cfg, err = GatherConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
			os.Exit(1)
		}
	}

	exec := NewExecutor(cfg)
	defer exec.Close()

	for i := startStep; i < len(steps); i++ {
		s := steps[i]
		fmt.Printf("\n[%d/%d] %s\n", i+1, len(steps), s.name)
		if err := s.fn(cfg, exec); err != nil {
			// Save state so user can resume after fixing the issue.
			if saveErr := SaveState(cfg, i-1); saveErr != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not save state: %v\n", saveErr)
			} else {
				fmt.Fprintf(os.Stderr, "\n  State saved. Re-run the command to resume from this step.\n")
			}
			fmt.Fprintf(os.Stderr, "\n✗ Step failed: %s\n  Error: %v\n", s.name, err)
			os.Exit(1)
		}
	}

	// All done — remove state file.
	ClearState(cfg.ProjectDir)
	fmt.Println("\n✓ Bootstrap complete!")
}

// tryResume looks for a saved state file and offers to resume.
func tryResume(steps []step) (*Config, int) {
	// Try the most common locations for a state file.
	home, _ := os.UserHomeDir()
	candidates := []string{
		home + "/projects",
		"/home",
	}

	// Also scan /home/*/projects for non-root users running with sudo.
	entries, _ := os.ReadDir("/home")
	for _, e := range entries {
		if e.IsDir() {
			candidates = append(candidates, "/home/"+e.Name()+"/projects")
		}
	}

	for _, dir := range candidates {
		state, err := LoadState(dir)
		if err != nil || state == nil || state.Config == nil {
			continue
		}

		resumeStep := state.CompletedStep + 1
		if resumeStep >= len(steps) {
			continue
		}

		fmt.Printf("Found saved state in %s (completed %d/%d steps).\n", dir, resumeStep, len(steps))
		fmt.Printf("Next step: [%d/%d] %s\n", resumeStep+1, len(steps), steps[resumeStep].name)
		fmt.Print("Resume? (y/n) [y]: ")

		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
			if answer == "" || answer == "y" {
				fmt.Println("→ Resuming…")
				return state.Config, resumeStep
			}
		}

		// User chose not to resume — clear old state and start fresh.
		ClearState(dir)
		break
	}

	return nil, 0
}
