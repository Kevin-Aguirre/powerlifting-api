package gitutils

import (
	"fmt"
	"os"
	"path/filepath"
	"github.com/go-git/go-git/v6"
)

func CloneRepo(targetDir, repoURL string) error {
	targetPath := filepath.Clean(targetDir)

	if _, err := os.Stat(targetPath); err == nil {
		os.RemoveAll(targetPath)
	}

	_, err := git.PlainClone(targetPath, &git.CloneOptions{
		URL:      repoURL,
		Progress: os.Stdout,
	})
	if err != nil {
		return err
	}

	fmt.Println("Repo cloned to:", targetPath)
	return nil
}
