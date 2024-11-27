package bindep

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"os/exec"
	"path/filepath"
)

// Global `bindep` directory, so deps can be reused. Default behaviour is a shared temporary directory.
var Path = filepath.Join(os.TempDir(), ".bindep")

func New(repo, commit string, flags []string) (string, error) {
	// Ensure the global `bindep` directory exists.
	err := os.Mkdir(Path, 0777)
	if err != nil && !os.IsExist(err) {
		return "", err
	}

	// Hash the git repo + commit.
	hash := sha256.New()
	hash.Write([]byte(repo))
	hash.Write([]byte(commit))
	sum := hash.Sum(nil)

	// Derive the path.
	path := filepath.Join(Path, hex.EncodeToString(sum))

	// Check if the file already exists.
	_, err = os.Stat(path)
	if err != nil && !os.IsNotExist(err) {
		// Ensure the error is a non-existence issue.
		if !os.IsExist(err) {
			return "", err
		}

		// Binary is already built.
		return path, nil
	}

	tmp, err := os.MkdirTemp("", "")
	if err != nil {
		return "", err
	}

	cmd := func(name string, args ...string) error {
		c := exec.Command(name, args...)
		c.Dir = tmp

		return c.Run()
	}

	err = cmd("git", "clone", repo, ".")
	if err != nil {
		return "", err
	}

	err = cmd("git", "checkout", commit)
	if err != nil {
		return "", err
	}

	args := []string{"go", "build"}

	if len(flags) != 0 {
		args = append(args, flags...)
	}

	args = append(args, "-o", path)

	err = cmd("go", "build", "-o", path)
	if err != nil {
		return "", err
	}

	return path, nil
}
