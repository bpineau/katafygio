// I'd love a working pure Go implementation. I can't find any though, src-d/go-git
// being innapropriate due to https://github.com/src-d/go-git/issues/793 and
// https://github.com/src-d/go-git/issues/785 .

// Package git stores changes a in a git repository
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
	checkInterval   = 10 * time.Second
)

// Store will maintain a git repository off dumped kube objects
type Store struct {
	Logger   *logrus.Logger
	URL      string
	LocalDir string
	Author   string
	Email    string
	Msg      string
}

// New instantiate a new Store
func New(config *config.KdnConfig) *Store {
	return &Store{
		Logger:   config.Logger,
		URL:      config.GitURL,
		LocalDir: config.LocalDir,
		Author:   "Katafygio", // XXX maybe this could be a cli option
		Email:    "katafygio@localhost",
		Msg:      "Kubernetes cluster change",
	}
}

// Watch maintains a directory content committed
func (s *Store) Watch() {
	checkTick := time.NewTicker(checkInterval).C
	for {
		<-checkTick
		s.commitAndPush()
	}
}

// Git wraps the git command
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

// Status tests the git status of a repository
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

// Clone does git clone, or git init (when there's no GiURL to clone from)
func (s *Store) Clone() error {
	err := os.MkdirAll(s.LocalDir, 0700)
	if err != nil {
		return fmt.Errorf("failed to created %s: %v", s.LocalDir, err)
	}

	if s.URL == "" {
		err = s.Git("init", s.LocalDir)
	} else {
		err = s.Git("clone", s.URL, s.LocalDir)
	}

	if err != nil {
		return fmt.Errorf("failed to init or clone in %s: %v", s.LocalDir, err)
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

// Commit git commit all the directory's changes
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
		return false, fmt.Errorf("failed to git add -A: %v", err)
	}

	err = s.Git("commit", "-m", s.Msg)
	if err != nil {
		return false, fmt.Errorf("failed to git commit: %v", err)
	}

	return true, nil
}

// Push git push to the origin
func (s *Store) Push() error {
	err := s.Git("push")
	if err != nil {
		return fmt.Errorf("failed to git push: %v", err)
	}

	return nil
}

func (s *Store) commitAndPush() {
	changed, err := s.Commit()
	if err != nil {
		s.Logger.Warn(err)
	}

	if !changed || s.URL == "" {
		return
	}

	err = s.Push()
	if err != nil {
		s.Logger.Warn(err)
	}
}
