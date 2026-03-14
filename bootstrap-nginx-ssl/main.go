package main

import (
	"fmt"
	"os"
)

type step struct {
	name string
	fn   func(*Config, Executor) error
}

func main() {
	fmt.Println("=== Bootstrap: Ubuntu + Nginx + Docker + Certbot ===")
	fmt.Println()

	cfg, err := GatherConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		os.Exit(1)
	}

	exec := NewExecutor(cfg)
	defer exec.Close()

	steps := []step{
		{"Clone repositories", SetupRepos},
		{"Prepare database", SetupDatabase},
		{"Generate docker-compose & start containers", GenerateCompose},
		{"Configure Nginx reverse proxy", SetupNginx},
		{"Install Certbot & issue SSL certificates", SetupCertbot},
	}

	for i, s := range steps {
		fmt.Printf("\n[%d/%d] %s\n", i+1, len(steps), s.name)
		if err := s.fn(cfg, exec); err != nil {
			fmt.Fprintf(os.Stderr, "\n✗ Step failed: %s\n  Error: %v\n", s.name, err)
			os.Exit(1)
		}
	}

	fmt.Println("\n✓ Bootstrap complete!")
}
