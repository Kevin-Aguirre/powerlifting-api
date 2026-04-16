package data

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync/atomic"
	"time"
)

const (
	DefaultRepoURL  = "https://gitlab.com/openpowerlifting/opl-data.git"
	DefaultRepoPath = "./opl-data"
	DefaultDataPath = "./opl-data/meet-data"
	RefreshInterval = 1 * time.Hour
)

// DataStore holds the current database and supports atomic swaps for live reloads.
type DataStore struct {
	db atomic.Pointer[Database]
}

func NewDataStore() *DataStore {
	return &DataStore{}
}

func (ds *DataStore) DB() *Database {
	return ds.db.Load()
}

func (ds *DataStore) set(db *Database) {
	ds.db.Store(db)
}

// SetForTest is exported for use in tests outside this package.
func (ds *DataStore) SetForTest(db *Database) {
	ds.set(db)
}

// EnsureRepo clones the repo if missing, or pulls if it already exists.
func EnsureRepo(repoPath, repoURL string) error {
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		fmt.Println("Data not found, cloning (shallow)...")
		cmd := exec.Command("git", "clone", "--depth", "1", repoURL, repoPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	return pullRepo(repoPath)
}

func pullRepo(repoPath string) error {
	fmt.Println("Pulling latest data...")
	cmd := exec.Command("git", "-C", repoPath, "pull")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func getHeadHash(repoPath string) (string, error) {
	cmd := exec.Command("git", "-C", repoPath, "rev-parse", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// Init clones/pulls the repo, loads the database, and starts a background
// goroutine that pulls every interval and reloads when new commits appear.
func (ds *DataStore) Init(repoPath, repoURL, dataPath string, interval time.Duration) error {
	if err := EnsureRepo(repoPath, repoURL); err != nil {
		return fmt.Errorf("failed to ensure repo: %w", err)
	}

	fmt.Println("Loading powerlifting data...")
	db, err := LoadDatabase(dataPath)
	if err != nil {
		return fmt.Errorf("failed to load data: %w", err)
	}
	ds.set(db)
	fmt.Println("Successfully loaded data")

	lastHash, _ := getHeadHash(repoPath)

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			fmt.Println("Checking for data updates...")
			if err := pullRepo(repoPath); err != nil {
				fmt.Println("Error pulling repo:", err)
				continue
			}

			newHash, err := getHeadHash(repoPath)
			if err != nil {
				fmt.Println("Error getting HEAD hash:", err)
				continue
			}

			if newHash == lastHash {
				fmt.Println("Data is up to date")
				continue
			}

			fmt.Println("New data detected, reloading...")
			newDB, err := LoadDatabase(dataPath)
			if err != nil {
				fmt.Println("Error reloading data:", err)
				continue
			}
			ds.set(newDB)
			lastHash = newHash
			fmt.Println("Data reloaded successfully")
		}
	}()

	return nil
}
