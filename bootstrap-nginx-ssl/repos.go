package main

import (
	"fmt"
	"strings"
)

// cloneURL injects a Git token into HTTPS URLs for private repo access.
// e.g. https://github.com/user/repo → https://<token>@github.com/user/repo
func cloneURL(repoURL, token string) string {
	if token == "" {
		return repoURL
	}
	for _, prefix := range []string{"https://", "http://"} {
		if strings.HasPrefix(repoURL, prefix) {
			return prefix + token + "@" + strings.TrimPrefix(repoURL, prefix)
		}
	}
	// Non-HTTP URL (SSH, local path) — return as-is.
	return repoURL
}

func SetupRepos(cfg *Config, exec Executor) error {
	fmt.Printf("→ Creating %s directory…\n", cfg.ProjectDir)
	if _, err := exec.Run("mkdir -p " + cfg.ProjectDir); err != nil {
		return fmt.Errorf("failed to create %s: %w", cfg.ProjectDir, err)
	}

	for _, svc := range cfg.Services {
		dest := cfg.ProjectDir + "/" + svc.Name
		url := cloneURL(svc.RepoURL, cfg.GitToken)
		fmt.Printf("→ Cloning %s into %s…\n", svc.RepoURL, dest)

		// Remove existing directory to ensure a clean clone.
		if _, err := exec.Run(fmt.Sprintf("rm -rf %s", dest)); err != nil {
			return fmt.Errorf("failed to clean %s: %w", dest, err)
		}

		if _, err := exec.Run(fmt.Sprintf("git clone %s %s", url, dest)); err != nil {
			return fmt.Errorf("failed to clone %s: %w", svc.RepoURL, err)
		}
	}

	return nil
}
