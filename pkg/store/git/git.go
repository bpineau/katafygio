package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/spf13/afero"
)

var (
	// TimeoutCommands defines the max execution time for git commands
	TimeoutCommands = 60 * time.Second

	// CheckInterval defines the interval between local directory checks
	CheckInterval = 10 * time.Second

	// GitAuthor is the name of the commiter
	GitAuthor = "Katafygio"

	// GitEmail is the email of the commiter
	GitEmail = "katafygio@localhost"

	// GitMsg is the commit message we'll use
	GitMsg = "Kubernetes cluster change"
)

var appFs = afero.NewOsFs()

type logger interface {
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

// Store will maintain a git repository off dumped kube objects
type Store struct {
	Logger   logger
	LocalDir string
	URL      string
	Author   string
	Email    string
	Msg      string
	DryRun   bool
	stopch   chan struct{}
	donech   chan struct{}
}

// New instantiate a new git Store. url is optional.
func New(log logger, dryRun bool, dir, url string) *Store {
	return &Store{
		Logger:   log,
		LocalDir: dir,
		URL:      url,
		Author:   GitAuthor,
		Email:    GitEmail,
		Msg:      GitMsg,
		DryRun:   dryRun,
	}
}

// Start maintains a directory content committed
func (s *Store) Start() (*Store, error) {
	s.Logger.Infof("Starting git repository synchronizer")
	s.stopch = make(chan struct{})
	s.donech = make(chan struct{})

	err := s.CloneOrInit()
	if err != nil {
		return nil, err
	}

	go func() {
		checkTick := time.NewTicker(CheckInterval)
		defer checkTick.Stop()
		defer close(s.donech)

		for {
			select {
			case <-checkTick.C:
				s.commitAndPush()
			case <-s.stopch:
				return
			}
		}
	}()

	return s, nil
}

// Stop stops the git goroutine
func (s *Store) Stop() {
	s.Logger.Infof("Stopping git repository synchronizer")
	close(s.stopch)
	<-s.donech
}

// Git wraps the git command
func (s *Store) Git(args ...string) error {
	if s.DryRun {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutCommands)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", args...) // #nosec
	cmd.Dir = s.LocalDir
	cmd.Env = append(os.Environ(), fmt.Sprintf("GIT_DIR=%s/.git", s.LocalDir))

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git %s failed with code %v: %s", args[0], err, out)
	}

	return nil
}

// Status tests the git status of a repository
func (s *Store) Status() (changed bool, err error) {
	if s.DryRun {
		return false, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutCommands)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "status", "--porcelain") // #nosec
	cmd.Dir = s.LocalDir
	cmd.Env = append(os.Environ(), fmt.Sprintf("GIT_DIR=%s/.git", s.LocalDir))

	out, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("git status failed with code %v: %s", err, out)
	}

	if len(out) != 0 {
		return true, nil
	}

	return false, nil
}

// CloneOrInit create a new local repository, either with "git clone" (if a GitURL
// to clone from is provided), or "git init" (in the absence of GitURL).
func (s *Store) CloneOrInit() (err error) {
	s.LocalDir, err = filepath.Abs(s.LocalDir)
	if err != nil {
		return fmt.Errorf("can't find local dir absolute path (broken cwd?): %v", err)
	}

	if !s.DryRun {
		err = appFs.MkdirAll(s.LocalDir, 0700)
		if err != nil {
			return fmt.Errorf("failed to create %s: %v", s.LocalDir, err)
		}
	}

	if s.URL == "" {
		err = s.Git("init", s.LocalDir)
	} else {
		err = s.Git("clone", "--depth=1", s.URL, s.LocalDir)
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

	err = afero.WriteFile(appFs, s.LocalDir+"/.git/info/exclude", []byte(".temp-katafygio-*"), 0644)
	if err != nil {
		return fmt.Errorf("failed to create a git exclusion: %v", err)
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

	err = s.Git("pull", "-s", "recursive", "-X", "ours")
	if err != nil {
		return false, fmt.Errorf("failed to git pull -s recursive -X ours: %v", err)
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
		s.Logger.Errorf("%v", err)
	}

	if !changed || s.URL == "" {
		return
	}

	err = s.Push()
	if err != nil {
		s.Logger.Errorf("%v", err)
	}
}
