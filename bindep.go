package bindep

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Global `bindep` directory, so deps can be reused. Default behaviour is a shared temporary directory.
var Path = filepath.Join(os.TempDir(), ".bindep")

var Debug = false

type Config struct {
	Repo   string
	Commit string

	Args []string

	Builder func(path string, cmd func(name string, args ...string) error) error
}

func New(cfg *Config) (string, error) {
	// Ensure the global `bindep` directory exists.
	err := os.Mkdir(Path, 0777)
	if err != nil && !os.IsExist(err) {
		return "", err
	}

	// Hash the git repo + commit.
	hash := sha256.New()
	hash.Write([]byte(cfg.Repo))
	hash.Write([]byte(cfg.Commit))
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

		if Debug {
			fmt.Println(name, args)

			c.Stderr = os.Stderr
			c.Stdout = os.Stdout
		}

		return c.Run()
	}

	err = cmd("git", "clone", cfg.Repo, ".")
	if err != nil {
		return "", err
	}

	err = cmd("git", "checkout", cfg.Commit)
	if err != nil {
		return "", err
	}

	if cfg.Builder != nil {
		return path, cfg.Builder(path, cmd)
	}

	args := append([]string{"build", "-o", path}, cfg.Args...)

	err = cmd("go", args...)
	if err != nil {
		return "", err
	}

	return path, nil
}
