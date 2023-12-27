package version

import "os/exec"

// Git represents an implementation of VersionControl interface using Git
type Git struct{}

// NewGit returns a new instance of Git
func NewGit() *Git {
	return &Git{}
}

// Add implements the Add method in the VersionControl interface for git
func (g *Git) Add(filePath string) error {
	cmd := exec.Command("git", "add", filePath)
	return cmd.Run()
}

// Commit implements the Commit method in the VersionControl interface for git
func (g *Git) Commit(message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	return cmd.Run()
}

// Tag implements the Tag method in the VersionControl interface for git
func (g *Git) Tag(version string) error {
	cmd := exec.Command("git", "tag", "-m", version, version)
	return cmd.Run()
}
