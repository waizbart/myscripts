package main

import (
	"bytes"
	"fmt"
	"text/template"
)

func GenerateCompose(cfg *Config, exec Executor) error {
	fmt.Println("→ Generating docker-compose.yml…")

	tmpl, err := template.New("compose").Parse(dockerComposeTpl)
	if err != nil {
		return fmt.Errorf("failed to parse compose template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, cfg); err != nil {
		return fmt.Errorf("failed to render compose template: %w", err)
	}

	composePath := cfg.ProjectDir + "/docker-compose.yml"
	if err := exec.WriteFile(composePath, buf.String(), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", composePath, err)
	}

	fmt.Println("→ Running docker compose up…")
	out, err := exec.Run(fmt.Sprintf("cd %s && docker compose up -d --build", cfg.ProjectDir))
	if err != nil {
		return fmt.Errorf("docker compose up failed: %s\n%w", out, err)
	}
	fmt.Println(out)

	return nil
}
