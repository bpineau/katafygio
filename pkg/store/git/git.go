// git stores changes a in a git repository
// I'd love a working pure Go implementation. I can't find any though, src-d/go-git
// being innapropriate due to https://github.com/src-d/go-git/issues/793 and
// https://github.com/src-d/go-git/issues/785 .
package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/bpineau/katafygio/config"
	"github.com/sirupsen/logrus"
)

const (
	timeoutCommands = 60 * time.Second
	CheckDelay      = 10 * time.Second
)

type Store struct {
	Logger   *logrus.Logger
	URL      string
	LocalDir string
	Author   string
	Email    string
	Msg      string
}

func New(config *config.KdnConfig) *Store {
	return &Store{
		Logger:   config.Logger,
		URL:      config.GitUrl,
		LocalDir: config.LocalDir,
		Author:   "Katafygio", // XXX maybe this could be a cli option
		Email:    "katafygio@localhost",
		Msg:      "Kubernetes cluster change",
	}
}

func (s *Store) Watch() {
	checkTick := time.NewTicker(CheckDelay).C
	for {
		select {
		case <-checkTick:
			s.commandAndPush()
		}

	}
}

func (s *Store) Git(args ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutCommands)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = s.LocalDir

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git %s failed with code %v: %s", args[0], err, out)
	}

	return nil
}

func (s *Store) Status() (changed bool, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutCommands)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "status", "--porcelain")
	cmd.Dir = s.LocalDir

	out, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("git status failed with code %v: %s", err, out)
	}

	if len(out) != 0 {
		return true, nil
	}

	return false, nil
}

func (s *Store) Clone() error {
	err := os.MkdirAll(s.LocalDir, 0700)
	if err != nil {
		return fmt.Errorf("failed to created %s: %v", s.LocalDir, err)
	}

	err = s.Git("clone", s.URL, s.LocalDir)
	if err != nil {
		return fmt.Errorf("failed to clone %s in %s: %v", s.URL, s.LocalDir, err)
	}

	err = s.Git("config", "user.name", s.Author)
	if err != nil {
		return fmt.Errorf("failed to config git user.name %s in %s: %v",
			s.Author, s.LocalDir, err)
	}

	err = s.Git("config", "user.email", s.Email)
	if err != nil {
		return fmt.Errorf("failed to config git user.email %s in %s: %v",
			s.Email, s.LocalDir, err)
	}

	return nil
}

func (s *Store) Commit() (changed bool, err error) {
	changed, err = s.Status()
	if err != nil {
		return changed, err
	}

	if !changed {
		return false, nil
	}

	err = s.Git("add", "-A")
	if err != nil {
		return false, fmt.Errorf("failed to add -A: %v", err)
	}

	err = s.Git("commit", "-m", s.Msg)
	if err != nil {
		return false, fmt.Errorf("failed to commit: %v", err)
	}

	return true, nil
}

func (s *Store) Push() error {
	err := s.Git("push")
	if err != nil {
		return fmt.Errorf("failed to push: %v", err)
	}

	return nil
}

func (s *Store) commandAndPush() {
	changed, err := s.Commit()
	if err != nil {
		s.Logger.Warnf("failed to commit: %v", err)
	}
	if changed {
		err := s.Push()
		if err != nil {
			s.Logger.Warnf("failed to push: %v", err)
		}
	}
}
