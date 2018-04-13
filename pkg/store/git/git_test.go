package git

import (
	"io/ioutil"
	"os"
	"os/exec"
	"testing"

	"github.com/spf13/afero"

	"github.com/bpineau/katafygio/config"
	"github.com/bpineau/katafygio/pkg/log"
)

var testHasGit bool

func init() {
	// Thanks to Mitchell Hashimoto!
	if _, err := exec.LookPath("git"); err == nil {
		testHasGit = true
	}
}

func TestGitDryRun(t *testing.T) {
	appFs = afero.NewMemMapFs()

	conf := &config.KfConfig{
		DryRun:   true,
		Logger:   log.New("info", "", "test"),
		LocalDir: "/tmp/ktest", // fake dir (in memory fs provided by Afero)
	}

	repo, err := New(conf).Start()
	if err != nil {
		t.Errorf("failed to start git: %v", err)
	}

	_, err = repo.Status()
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

	conf := &config.KfConfig{
		Logger:   log.New("info", "", "test"),
		LocalDir: dir,
	}

	repo, err := New(conf).Start()
	if err != nil {
		t.Errorf("failed to start git: %v", err)
	}

	changed, err := repo.Status()
	if changed || err != nil {
		t.Errorf("Status should return false on empty new repos (%v)", err)
	}

	_ = ioutil.WriteFile(dir+"/t.yaml", []byte{42}, 0600)

	changed, err = repo.Status()
	if !changed || err != nil {
		t.Errorf("Status should return true on non committed files (%v)", err)
	}

	changed, err = repo.Commit()
	if !changed || err != nil {
		t.Errorf("Commit should notify changes and not fail (%v)", err)
	}

	changed, err = repo.Status()
	if changed || err != nil {
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

	err = repo.Clone()
	if err != nil {
		t.Errorf("clone failed: %v", err)
	}

	_ = ioutil.WriteFile(newdir+"/t2.yaml", []byte{42}, 0600)
	repo.commitAndPush()

	changed, err = repo.Status()
	if changed || err != nil {
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
	_, err = repo.Status()
	if err == nil {
		t.Error("Status should fail on a non-repos")
	}
	repo.commitAndPush()
	_, err = repo.Commit()
	if err == nil {
		t.Error("Commit should fail on a non-repos")
	}
}
