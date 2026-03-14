package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type ServiceConfig struct {
	RepoURL string
	Name    string
	Domain  string
	Port    int
}

type DatabaseConfig struct {
	Enabled      bool
	RootPassword string
	MySQLPort    int
	AdminPort    int
}

type Config struct {
	TargetMode string // "local" or "remote"
	SSHUser    string
	SSHHost    string
	SSHPort    string
	GitToken   string // optional, for cloning private repos
	Services   []ServiceConfig
	Database   DatabaseConfig
}

// GatherConfig interactively prompts the user for all configuration up front.
func GatherConfig() (*Config, error) {
	scanner := bufio.NewScanner(os.Stdin)
	cfg := &Config{}

	// --- target mode ---
	cfg.TargetMode = prompt(scanner, "Target mode (local/remote)", "local")
	if cfg.TargetMode == "remote" {
		cfg.SSHUser = prompt(scanner, "SSH user", "root")
		cfg.SSHHost = promptRequired(scanner, "SSH host")
		cfg.SSHPort = prompt(scanner, "SSH port", "22")
	}

	// --- git auth ---
	fmt.Println("\n— Git Authentication —")
	cfg.GitToken = prompt(scanner, "Git token for private repos (leave empty for public)", "")

	// --- services ---
	fmt.Println("\n— Services —")
	fmt.Println("Add the git repositories you want to deploy.")
	for {
		repoURL := promptRequired(scanner, "Git repo URL (or 'done' to finish)")
		if strings.ToLower(repoURL) == "done" {
			break
		}

		name := promptRequired(scanner, "Service name (used for container & nginx)")
		domain := promptRequired(scanner, "Domain for this service")
		port := promptInt(scanner, "Application port inside the container", 3000)

		cfg.Services = append(cfg.Services, ServiceConfig{
			RepoURL: repoURL,
			Name:    name,
			Domain:  domain,
			Port:    port,
		})
	}

	if len(cfg.Services) == 0 {
		return nil, fmt.Errorf("at least one service is required")
	}

	// --- database ---
	fmt.Println("\n— Database —")
	dbEnabled := prompt(scanner, "Enable MySQL + phpMyAdmin? (y/n)", "n")
	if strings.ToLower(dbEnabled) == "y" {
		cfg.Database.Enabled = true
		cfg.Database.RootPassword = promptRequired(scanner, "MySQL root password")
		cfg.Database.MySQLPort = promptInt(scanner, "MySQL host port", 3306)
		cfg.Database.AdminPort = promptInt(scanner, "phpMyAdmin host port", 8080)
	}

	return cfg, nil
}

// NewExecutor creates the appropriate Executor from the gathered config.
func NewExecutor(cfg *Config) Executor {
	if cfg.TargetMode == "remote" {
		return &SSHExecutor{
			User: cfg.SSHUser,
			Host: cfg.SSHHost,
			Port: cfg.SSHPort,
		}
	}
	return &LocalExecutor{}
}

// ---------------------------------------------------------------------------
// prompt helpers
// ---------------------------------------------------------------------------

func prompt(scanner *bufio.Scanner, label, defaultVal string) string {
	fmt.Printf("%s [%s]: ", label, defaultVal)
	if scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text != "" {
			return text
		}
	}
	return defaultVal
}

func promptRequired(scanner *bufio.Scanner, label string) string {
	for {
		fmt.Printf("%s: ", label)
		if scanner.Scan() {
			text := strings.TrimSpace(scanner.Text())
			if text != "" {
				return text
			}
		}
		fmt.Println("  This field is required.")
	}
}

func promptInt(scanner *bufio.Scanner, label string, defaultVal int) int {
	for {
		fmt.Printf("%s [%d]: ", label, defaultVal)
		if scanner.Scan() {
			text := strings.TrimSpace(scanner.Text())
			if text == "" {
				return defaultVal
			}
			var n int
			if _, err := fmt.Sscanf(text, "%d", &n); err == nil {
				return n
			}
			fmt.Println("  Please enter a valid number.")
		}
	}
}
