package main

import "fmt"

func SetupDatabase(cfg *Config, exec Executor) error {
	if !cfg.Database.Enabled {
		fmt.Println("→ Database not enabled, skipping.")
		return nil
	}

	fmt.Println("→ Preparing MySQL data volume directory…")
	if _, err := exec.Run("mkdir -p /projects/mysql-data"); err != nil {
		return fmt.Errorf("failed to create mysql-data directory: %w", err)
	}

	return nil
}
