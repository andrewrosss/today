package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"time"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func run() error {
	// bootstrap the config for the application
	config, err := initializeConfig()
	if err != nil {
		return fmt.Errorf("initializing config: %w", err)
	}

	// make sure the directory for entries exists
	if err = os.MkdirAll(config.EntriesDir, 0755); err != nil {
		return fmt.Errorf("creating entries directory (%s): %w", config.EntriesDir, err)
	}

	// find the entry/file for the most recent day, either:
	// - it doesn't exist, so we create a blank file for today
	// - it does exist and it's not today, so we find the penultimate day's
	//   entry/file and copy the contents to today's entry
	// - it does exist and it's today, do nothing
	entryPaths, err := listEntryPaths(config.EntriesDir)
	if err != nil {
		return fmt.Errorf("listing entries in directory (%s): %w", config.EntriesDir, err)
	}

	todayPath := getTodayPath(config.EntriesDir)

	// sort the entries so we can easily find the most recent one
	slices.Sort(entryPaths)

	// create or forward the entry for today
	if len(entryPaths) == 0 {
		log.Println("No previous entries found, creating a new one for today")
		if err = createTodayFile(todayPath); err != nil {
			return fmt.Errorf("creating today's entry file (%s): %w", todayPath, err)
		}
	} else if prevPath := entryPaths[len(entryPaths)-1]; prevPath != todayPath {
		log.Printf("Forwarding previous entry (%s) to today", prevPath)
		if err = forwardPreviousFile(prevPath, todayPath); err != nil {
			return fmt.Errorf("forwarding previous entry (%s) to today (%s): %w", prevPath, todayPath, err)
		}
	} else {
		log.Println("Today's entry already exists")
	}

	// we always write the today's entry path to STDOUT
	fmt.Println(todayPath)

	return nil
}

type Config struct {
	// The directory where the entries are stored
	EntriesDir string `json:"entries_dir"`
}

func configWithDefaults(appDir string) Config {
	return Config{
		EntriesDir: filepath.Join(appDir, "entries"),
	}
}

func initializeConfig() (Config, error) {
	appDir, err := getAppDir()
	if err != nil {
		// if we can't get the "today dir", we can't do anything ...
		return Config{}, err
	}
	configPath := getConfigPath(appDir)
	configFile, err := os.Open(configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// ok if the file doesn't exist, we'll just use the defaults
			return configWithDefaults(appDir), nil
		}
		// if it's not a "file not found" error, we _also_ can't do anything
		return Config{}, err
	}
	defer configFile.Close()

	var config Config
	if err = json.NewDecoder(configFile).Decode(&config); err != nil {
		// at this point the file _does_ exist, but we can't parse it.
		// the user should probably know about this, so we'll return the error
		return Config{}, err
	}

	// fully resolve the entries directory
	entriesDir, err := expanduser(config.EntriesDir)
	if err != nil {
		return Config{}, err
	}
	config.EntriesDir = entriesDir

	return config, nil
}

func getConfigPath(appDir string) string {
	configFile := filepath.Join(appDir, "config.json")
	return configFile
}

func getAppDir() (string, error) {
	// user has set the TODAY_DIR environment variable so use that
	appDir := os.Getenv("TODAY_DIR")
	if appDir != "" {
		return appDir, nil
	}

	// otherwise use the default
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	appDir = filepath.Join(home, ".today")

	return appDir, nil
}

func listEntryPaths(entriesDir string) ([]string, error) {
	return filepath.Glob(filepath.Join(entriesDir, "????-??-??.md"))
}

func getTodayPath(entriesDir string) string {
	formatted := time.Now().Format("2006-01-02")
	return filepath.Join(entriesDir, fmt.Sprintf("%s.md", formatted))
}

func createTodayFile(todayPath string) error {
	heading := makeHeading(time.Now())
	if err := os.WriteFile(todayPath, []byte(heading), 0664); err != nil {
		return err
	}
	return nil
}

func forwardPreviousFile(prevPath, todayPath string) error {
	// Read the contents of the previous file
	contentBytes, err := os.ReadFile(prevPath)
	if err != nil {
		return err
	}

	// If the main header is a date, then replace it with today's date
	re := regexp.MustCompile(`^# \d{4}-\d{2}-\d{2}\n`)
	heading := makeHeading(time.Now())
	contents := re.ReplaceAllString(string(contentBytes), heading)

	// Write the contents to today's file
	if err = os.WriteFile(todayPath, []byte(contents), 0644); err != nil {
		return err
	}

	return nil
}

func makeHeading(date time.Time) string {
	return fmt.Sprintf("# %s\n", date.Format("2006-01-02"))
}

func expanduser(path string) (string, error) {
	if len(path) == 0 || path[0] != '~' {
		return path, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, path[1:]), nil
}
