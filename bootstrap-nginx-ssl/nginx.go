package main

import (
	"bytes"
	"fmt"
	"text/template"
)

func SetupNginx(cfg *Config, exec Executor) error {
	fmt.Println("→ Installing Nginx…")
	out, err := exec.Run("apt-get update -qq && apt-get install -y -qq nginx")
	if err != nil {
		return fmt.Errorf("failed to install nginx: %s\n%w", out, err)
	}

	tmpl, err := template.New("nginx").Parse(nginxSiteTpl)
	if err != nil {
		return fmt.Errorf("failed to parse nginx template: %w", err)
	}

	for _, svc := range cfg.Services {
		fmt.Printf("→ Writing Nginx config for %s (%s)…\n", svc.Name, svc.Domain)

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, svc); err != nil {
			return fmt.Errorf("failed to render nginx config for %s: %w", svc.Name, err)
		}

		confPath := fmt.Sprintf("/etc/nginx/sites-available/%s", svc.Name)
		linkPath := fmt.Sprintf("/etc/nginx/sites-enabled/%s", svc.Name)

		if err := exec.WriteFile(confPath, buf.String(), 0644); err != nil {
			return fmt.Errorf("failed to write nginx config %s: %w", confPath, err)
		}

		if _, err := exec.Run(fmt.Sprintf("ln -sf %s %s", confPath, linkPath)); err != nil {
			return fmt.Errorf("failed to enable site %s: %w", svc.Name, err)
		}
	}

	// Remove default site if it exists.
	exec.Run("rm -f /etc/nginx/sites-enabled/default")

	fmt.Println("→ Testing and reloading Nginx…")
	out, err = exec.Run("nginx -t && systemctl reload nginx")
	if err != nil {
		return fmt.Errorf("nginx reload failed: %s\n%w", out, err)
	}

	return nil
}
