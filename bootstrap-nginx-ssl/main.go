package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type step struct {
	key  string // short identifier used with --skip (e.g. "clone", "nginx")
	name string
	fn   func(*Config, Executor) error
}

func main() {
	fmt.Println("=== Bootstrap: Ubuntu + Nginx + Docker + Certbot ===")
	fmt.Println()

	allSteps := []step{
		{"clone", "Clone repositories", SetupRepos},
		{"database", "Prepare database", SetupDatabase},
		{"docker", "Install Docker", SetupDocker},
		{"compose", "Generate docker-compose & start containers", GenerateCompose},
		{"nginx", "Configure Nginx reverse proxy", SetupNginx},
		{"certbot", "Install Certbot & issue SSL certificates", SetupCertbot},
	}

	// Parse --skip flag from CLI args.
	skipped := parseSkipFlag(os.Args[1:])
	if len(skipped) > 0 {
		fmt.Printf("Skipping steps: %s\n", strings.Join(setKeys(skipped), ", "))
	}

	steps := filterSteps(allSteps, skipped)

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

// parseSkipFlag extracts the set of step keys to skip from args.
// Accepts --skip=clone,compose or --skip clone,compose.
func parseSkipFlag(args []string) map[string]bool {
	skipped := map[string]bool{}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		var val string
		if strings.HasPrefix(arg, "--skip=") {
			val = strings.TrimPrefix(arg, "--skip=")
		} else if arg == "--skip" && i+1 < len(args) {
			i++
			val = args[i]
		} else {
			continue
		}
		for _, key := range strings.Split(val, ",") {
			key = strings.TrimSpace(strings.ToLower(key))
			if key != "" {
				skipped[key] = true
			}
		}
	}
	return skipped
}

// filterSteps returns only the steps whose key is not in the skipped set.
func filterSteps(steps []step, skipped map[string]bool) []step {
	if len(skipped) == 0 {
		return steps
	}
	filtered := make([]step, 0, len(steps))
	for _, s := range steps {
		if skipped[s.key] {
			fmt.Printf("  ↷ Skipping: %s\n", s.name)
			continue
		}
		filtered = append(filtered, s)
	}
	return filtered
}

// setKeys returns the keys of a map as a sorted-ish slice for display.
func setKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
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
