// Package git makes a git repository out of a local directory, keeps the
// content committed when the directory content changes, and optionaly (if
// a remote repos url is provided), keep it in sync with a remote repository.
//
// It requires the git command in $PATH, since the pure Go git implementations
// aren't up to the task (see go-git issues #793 and #785 for instance).
package git
