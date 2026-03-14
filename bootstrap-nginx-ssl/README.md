# Bootstrap: Ubuntu + Nginx + Docker + Certbot

Go CLI tool that bootstraps an Ubuntu server with Docker containers behind an Nginx reverse proxy, secured with Certbot SSL certificates. Works locally or over SSH.

## What it does

1. **Target selection** — Run locally or remotely via SSH (uses system `ssh` binary)
2. **Clone repositories** — Prompt for repo URLs and clone them into `/projects`
3. **Database setup (optional)** — Spin up MySQL + phpMyAdmin containers
4. **Docker Compose generation** — Generate `docker-compose.yml` wiring up all services, then `docker compose up -d --build`
5. **Nginx reverse proxy** — Install and configure Nginx with per-service reverse proxy configs
6. **SSL certificates** — Install Certbot and issue Let's Encrypt certificates for each domain

## Project structure

```
main.go           — Entry point, runs steps sequentially
config.go         — Config structs + interactive prompts (bufio.Scanner)
executor.go       — Executor interface + LocalExecutor + SSHExecutor
repos.go          — Clone git repositories into /projects
database.go       — Prepare MySQL volume directory
compose.go        — Generate docker-compose.yml + run docker compose up
nginx.go          — Install & configure Nginx reverse proxy
certbot.go        — Install Certbot + issue SSL certs
templates.go      — text/template strings for docker-compose.yml & nginx configs
tests/
  Dockerfile.test — Ubuntu container for integration tests
  test.sh         — Script to build and run tests in Docker
```

## Requirements

- Go 1.23+ (to build)
- Target machine: Ubuntu with `apt`, `git`, `docker` (with compose plugin)
- No external Go dependencies — standard library only

## Usage

```bash
# Build
go build -o bootstrap .

# Run
./bootstrap
```

The tool will interactively prompt for:
- **Target mode** — local or remote (SSH user/host/port)
- **Services** — git repo URL, service name, domain, and port for each
- **Database** — optionally enable MySQL + phpMyAdmin with custom ports/password

All configuration is gathered upfront before any changes are made.

## Testing

```bash
bash tests/test.sh
```

Runs integration tests inside a real Ubuntu container (via Docker). Tests cover:
- Repository cloning with a local bare git repo
- Database directory creation (enabled and disabled)
- Docker Compose template rendering (with and without database)
- Nginx template rendering and config validation (`nginx -t`)
- LocalExecutor (`Run`, `WriteFile`, `Close`)
