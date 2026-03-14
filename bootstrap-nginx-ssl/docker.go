package main

import "fmt"

func SetupDocker(cfg *Config, exec Executor) error {
	// Check if docker is already installed.
	if _, err := exec.Run("docker --version"); err == nil {
		fmt.Println("→ Docker is already installed, skipping.")
		return nil
	}

	fmt.Println("→ Installing Docker…")

	cmds := []struct {
		desc string
		cmd  string
	}{
		{"Installing prerequisites", "apt-get update -qq && apt-get install -y -qq ca-certificates curl gnupg"},
		{"Adding Docker GPG key", "install -m 0755 -d /etc/apt/keyrings && curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc && chmod a+r /etc/apt/keyrings/docker.asc"},
		{"Adding Docker repository", `echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu $(. /etc/os-release && echo "$VERSION_CODENAME") stable" > /etc/apt/sources.list.d/docker.list`},
		{"Installing Docker Engine", "apt-get update -qq && apt-get install -y -qq docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin"},
	}

	for _, c := range cmds {
		fmt.Printf("  → %s…\n", c.desc)
		out, err := exec.Run(c.cmd)
		if err != nil {
			return fmt.Errorf("%s failed: %s\n%w", c.desc, out, err)
		}
	}

	// Verify installation.
	out, err := exec.Run("docker --version")
	if err != nil {
		return fmt.Errorf("docker not working after install: %s\n%w", out, err)
	}
	fmt.Printf("→ %s\n", out)

	return nil
}
