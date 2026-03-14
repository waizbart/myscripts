package main

import "fmt"

func SetupDatabase(cfg *Config, exec Executor) error {
	if !cfg.Database.Enabled {
		fmt.Println("→ Database not enabled, skipping.")
		return nil
	}

	dataDir := cfg.ProjectDir + "/mysql-data"
	fmt.Printf("→ Preparing MySQL data volume directory (%s)…\n", dataDir)
	if _, err := exec.Run("mkdir -p " + dataDir); err != nil {
		return fmt.Errorf("failed to create mysql-data directory: %w", err)
	}

	return nil
}
