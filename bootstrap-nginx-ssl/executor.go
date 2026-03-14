package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Executor abstracts command execution and file writing so the same
// bootstrap logic works both locally and over SSH.
type Executor interface {
	Run(cmd string) (string, error)
	WriteFile(path, content string, perm os.FileMode) error
	Close() error
}

// ---------------------------------------------------------------------------
// LocalExecutor
// ---------------------------------------------------------------------------

type LocalExecutor struct{}

func (l *LocalExecutor) Run(cmd string) (string, error) {
	c := exec.Command("bash", "-c", cmd)
	out, err := c.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func (l *LocalExecutor) WriteFile(path, content string, perm os.FileMode) error {
	return os.WriteFile(path, []byte(content), perm)
}

func (l *LocalExecutor) Close() error { return nil }

// ---------------------------------------------------------------------------
// SSHExecutor — uses the system ssh binary so it inherits the user's
// SSH agent, config, and known_hosts without pulling in external Go deps.
// ---------------------------------------------------------------------------

type SSHExecutor struct {
	User string
	Host string
	Port string
}

func (s *SSHExecutor) sshArgs() []string {
	args := []string{"-o", "StrictHostKeyChecking=accept-new"}
	if s.Port != "" && s.Port != "22" {
		args = append(args, "-p", s.Port)
	}
	args = append(args, fmt.Sprintf("%s@%s", s.User, s.Host))
	return args
}

func (s *SSHExecutor) Run(cmd string) (string, error) {
	args := append(s.sshArgs(), cmd)
	c := exec.Command("ssh", args...)
	out, err := c.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func (s *SSHExecutor) WriteFile(path, content string, perm os.FileMode) error {
	// Use a heredoc over SSH to write the file.
	cmd := fmt.Sprintf("cat > %s << 'BOOTSTRAP_EOF'\n%s\nBOOTSTRAP_EOF\nchmod %o %s",
		path, content, perm, path)
	_, err := s.Run(cmd)
	return err
}

func (s *SSHExecutor) Close() error { return nil }
