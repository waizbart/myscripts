package main

import "fmt"

func SetupCertbot(cfg *Config, exec Executor) error {
	fmt.Println("→ Installing Certbot…")
	out, err := exec.Run("apt-get install -y -qq certbot python3-certbot-nginx")
	if err != nil {
		return fmt.Errorf("failed to install certbot: %s\n%w", out, err)
	}

	for _, svc := range cfg.Services {
		fmt.Printf("→ Requesting SSL certificate for %s…\n", svc.Domain)
		cmd := fmt.Sprintf(
			"certbot --nginx -d %s --non-interactive --agree-tos --register-unsafely-without-email",
			svc.Domain,
		)
		out, err := exec.Run(cmd)
		if err != nil {
			return fmt.Errorf("certbot failed for %s: %s\n%w", svc.Domain, out, err)
		}
		fmt.Println(out)
	}

	return nil
}
