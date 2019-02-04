package git

import (
	"io/ioutil"
	"os"
	"os/exec"
	"testing"

	"github.com/spf13/afero"
)

var testHasGit bool

func init() {
	// Thanks to Mitchell Hashimoto!
	if _, err := exec.LookPath("git"); err == nil {
		testHasGit = true
	}
}

type mockLog struct{}

func (m *mockLog) Infof(format string, args ...interface{})  {}
func (m *mockLog) Errorf(format string, args ...interface{}) {}

func TestGitDryRun(t *testing.T) {
	if !testHasGit {
		t.Log("git not found, skipping")
		t.Skip()
	}

	appFs = afero.NewMemMapFs()

	repo, err := New(new(mockLog), true, "/tmp/ktest", "").Start()
	if err != nil {
		t.Errorf("failed to start git: %v", err)
	}

	_, _, err = repo.Status()
	if err != nil {
		t.Error(err)
	}

	repo.Stop()
}

// testing with real git repositories and commands
func TestGit(t *testing.T) {
	if !testHasGit {
		t.Log("git not found, skipping")
		t.Skip()
	}

	dir, err := ioutil.TempDir("", "katafygio-tests")
	if err != nil {
		t.Fatal("failed to create a temp dir for tests")
	}

	defer os.RemoveAll(dir)

	repo, err := New(new(mockLog), false, dir, "").Start()
	if err != nil {
		t.Errorf("failed to start git: %v", err)
	}

	changed, status, err := repo.Status()
	if changed || status != "" || err != nil {
		t.Errorf("Status should return false on empty new repos (%v)", err)
	}

	_ = ioutil.WriteFile(dir+"/t.yaml", []byte{42}, 0600)

	changed, status, err = repo.Status()
	if !changed || status == "" || err != nil {
		t.Errorf("Status should return true on non committed files (%v)", err)
	}

	changed, err = repo.Commit()
	if !changed || err != nil {
		t.Errorf("Commit should notify changes and not fail (%v)", err)
	}

	changed, status, err = repo.Status()
	if changed || status != "" || err != nil {
		t.Errorf("Status should return false after a add+commit (%v)", err)
	}

	changed, err = repo.Commit()
	if changed || err != nil {
		t.Errorf("Commit shouldn't notify changes on unchanged repos (%v)", err)
	}

	// re-use the previous repos for clone tests

	newdir, err := ioutil.TempDir("", "katafygio-tests")
	if err != nil {
		t.Fatal("failed to create a temp dir for tests")
	}

	defer os.RemoveAll(newdir)

	repo.LocalDir = newdir
	repo.URL = dir

	err = repo.CloneOrInit()
	if err != nil {
		t.Errorf("clone failed: %v", err)
	}

	_ = ioutil.WriteFile(newdir+"/t2.yaml", []byte{42}, 0600)
	repo.commitAndPush()

	changed, status, err = repo.Status()
	if changed || status != "" || err != nil {
		t.Errorf("Status should return false after a add+commit+push (%v)", err)
	}

	repo.Stop()

	// test various failure modes

	_, err = repo.Start()
	if err == nil {
		t.Error("Start/Clone on an existing repository should fail")
	}

	err = repo.Git("fortzob", "42")
	if err == nil {
		t.Error("Git should fail with unknown subcommands")
	}

	if err == nil {
		t.Error("clone should fail on existing repos")
	}

	notrepo, err := ioutil.TempDir("", "katafygio-tests")
	if err != nil {
		t.Fatal("failed to create a temp dir for tests")
	}

	defer os.RemoveAll(notrepo)

	repo.LocalDir = notrepo
	_, _, err = repo.Status()
	if err == nil {
		t.Error("Status should fail on a non-repos")
	}
	repo.commitAndPush()
	_, err = repo.Commit()
	if err == nil {
		t.Error("Commit should fail on a non-repos")
	}
}
