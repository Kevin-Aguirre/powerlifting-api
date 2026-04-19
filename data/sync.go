package data

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"sync/atomic"
	"time"
)

func loadAndTime(dataPath string) (*Database, time.Duration, error) {
	start := time.Now()
	db, err := LoadDatabase(dataPath)
	dur := time.Since(start)
	if err == nil {
		DataLoadDurationSeconds.Set(dur.Seconds())
		DataLoadsTotal.Inc()
	}
	return db, dur, err
}

const (
	DefaultRepoURL  = "https://gitlab.com/openpowerlifting/opl-data.git"
	DefaultRepoPath = "./opl-data"
	DefaultDataPath = "./opl-data/meet-data"
	RefreshInterval = 1 * time.Hour
)

// DataStore holds the current database and supports atomic swaps for live reloads.
type DataStore struct {
	db          atomic.Pointer[Database]
	lastUpdated atomic.Value // stores time.Time
}

func NewDataStore() *DataStore {
	return &DataStore{}
}

func (ds *DataStore) DB() *Database {
	return ds.db.Load()
}

// LastUpdated returns the time the database was last successfully loaded.
func (ds *DataStore) LastUpdated() time.Time {
	v := ds.lastUpdated.Load()
	if v == nil {
		return time.Time{}
	}
	return v.(time.Time)
}

func (ds *DataStore) set(db *Database) {
	ds.db.Store(db)
	ds.lastUpdated.Store(time.Now())
}

// SetForTest is exported for use in tests outside this package.
func (ds *DataStore) SetForTest(db *Database) {
	ds.set(db)
}

// EnsureRepo clones the repo if missing, or pulls if it already exists.
func EnsureRepo(repoPath, repoURL string) error {
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		slog.Info("data not found, cloning (shallow)...")
		cmd := exec.Command("git", "clone", "--depth", "1", repoURL, repoPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	return pullRepo(repoPath)
}

func pullRepo(repoPath string) error {
	slog.Info("pulling latest data...")
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

	slog.Info("loading powerlifting data...")
	db, loadDur, err := loadAndTime(dataPath)
	if err != nil {
		return fmt.Errorf("failed to load data: %w", err)
	}
	ds.set(db)
	slog.Info("successfully loaded data", "duration", loadDur)

	lastHash, _ := getHeadHash(repoPath)

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			slog.Info("checking for data updates...")
			if err := pullRepo(repoPath); err != nil {
				slog.Error("error pulling repo", "error", err)
				continue
			}

			newHash, err := getHeadHash(repoPath)
			if err != nil {
				slog.Error("error getting HEAD hash", "error", err)
				continue
			}

			if newHash == lastHash {
				slog.Info("data is up to date")
				continue
			}

			slog.Info("new data detected, reloading...")
			newDB, loadDur, err := loadAndTime(dataPath)
			if err != nil {
				slog.Error("error reloading data", "error", err)
				continue
			}
			ds.set(newDB)
			lastHash = newHash
			slog.Info("data reloaded successfully", "duration", loadDur)
		}
	}()

	return nil
}
