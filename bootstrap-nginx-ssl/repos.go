package main

import "fmt"

func SetupRepos(cfg *Config, exec Executor) error {
	fmt.Println("→ Creating /projects directory…")
	if _, err := exec.Run("mkdir -p /projects"); err != nil {
		return fmt.Errorf("failed to create /projects: %w", err)
	}

	for _, svc := range cfg.Services {
		dest := "/projects/" + svc.Name
		fmt.Printf("→ Cloning %s into %s…\n", svc.RepoURL, dest)

		// Remove existing directory to ensure a clean clone.
		if _, err := exec.Run(fmt.Sprintf("rm -rf %s", dest)); err != nil {
			return fmt.Errorf("failed to clean %s: %w", dest, err)
		}

		if _, err := exec.Run(fmt.Sprintf("git clone %s %s", svc.RepoURL, dest)); err != nil {
			return fmt.Errorf("failed to clone %s: %w", svc.RepoURL, err)
		}
	}

	return nil
}
