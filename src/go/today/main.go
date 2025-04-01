package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"time"

	"github.com/alexflint/go-arg"
)

const (
	Version = "0.2.4"
)

type Args struct {
	ListEntries    *ListEntriesCmd    `arg:"subcommand:list" help:"list all entries"`
	ShowEntriesDir *ShowEntriesDirCmd `arg:"subcommand:dir" help:"print the configured directory where entries are stored"`
	Options
}

func (Args) Version() string {
	return fmt.Sprintf("today version %s", Version)
}

type ListEntriesCmd struct{}
type ShowEntriesDirCmd struct{}

type Options struct {
	EntriesDir        string `arg:"-d,--today-dir,env:TODAY_DIR" help:"directory where entries are stored" placeholder:"PATH" default:"~/.today"`
	Quiet             bool   `arg:"-q,env:TODAY_QUIET" help:"suppress logs written to STDERR"`
	ToStdout          bool   `arg:"--stdout" help:"write the contents of today's entry to STDOUT"`
	DeclareBankruptcy *int   `arg:"-b,--declare-bankruptcy" help:"forward only the content in sections up to the specified level"`
}

func main() {
	args := Args{}
	arg.MustParse(&args)

	if err := run(args); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func run(args Args) error {
	if args.Quiet {
		log.SetOutput(io.Discard)
	}

	switch {
	case args.ListEntries != nil:
		return handleListEntries(args.Options)
	case args.ShowEntriesDir != nil:
		return handleShowEntriesDir(args.Options)
	default:
		return handleCreate(args.Options)
	}
}

func handleListEntries(options Options) error {
	entriesDir, err := expanduser(options.EntriesDir)
	if err != nil {
		return fmt.Errorf("invalid entries directory (%s): %w", options.EntriesDir, err)
	}

	// make sure the directory for entries exists
	if err := os.MkdirAll(entriesDir, 0755); err != nil {
		return fmt.Errorf("creating entries directory (%s): %w", entriesDir, err)
	}

	entryPaths, err := listEntryPaths(entriesDir)
	if err != nil {
		return fmt.Errorf("listing entries in directory (%s): %w", entriesDir, err)
	}

	slices.Sort(entryPaths)

	for _, entryPath := range entryPaths {
		fmt.Println(entryPath)
	}

	return nil
}

func handleShowEntriesDir(options Options) error {
	entriesDir, err := expanduser(options.EntriesDir)
	if err != nil {
		return fmt.Errorf("invalid entries directory (%s): %w", options.EntriesDir, err)
	}
	fmt.Println(entriesDir)
	return nil
}

func handleCreate(options Options) error {
	entriesDir, err := expanduser(options.EntriesDir)
	if err != nil {
		return fmt.Errorf("invalid entries directory (%s): %w", options.EntriesDir, err)
	}

	// make sure the directory for entries exists
	if err := os.MkdirAll(entriesDir, 0755); err != nil {
		return fmt.Errorf("creating entries directory (%s): %w", entriesDir, err)
	}

	// find the entry/file for the most recent day, either:
	// - it doesn't exist, so we create a blank file for today
	// - it does exist and it's not today, so we find the penultimate day's
	//   entry/file and copy the contents to today's entry
	// - it does exist and it's today, do nothing
	entryPaths, err := listEntryPaths(entriesDir)
	if err != nil {
		return fmt.Errorf("listing entries in directory (%s): %w", entriesDir, err)
	}

	todayPath := getTodayPath(entriesDir)

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
		fwdOptions := ForwardingOptions{
			DeclareBankruptcy: options.DeclareBankruptcy,
		}
		if err = forwardPreviousFile(prevPath, todayPath, fwdOptions); err != nil {
			return fmt.Errorf("forwarding previous entry (%s) to today (%s): %w", prevPath, todayPath, err)
		}
	} else {
		log.Println("Today's entry already exists")
	}

	if options.ToStdout {
		// write today's entry's contents to STDOUT
		content, err := os.ReadFile(todayPath)
		if err != nil {
			return fmt.Errorf("reading today's entry file (%s): %w", todayPath, err)
		}
		fmt.Print(string(content))
	} else {
		// write today's entry's path to STDOUT
		fmt.Println(todayPath)
	}

	return nil
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
	if err := os.WriteFile(todayPath, []byte(heading), OS_ALL_R|OS_USER_RW); err != nil {
		return err
	}
	return nil
}

func forwardPreviousFile(prevPath, todayPath string, options ForwardingOptions) error {
	// Read the contents of the previous file
	prevContent, err := os.ReadFile(prevPath)
	if err != nil {
		return err
	}

	// If the main header is a date, then replace it with today's date
	re := regexp.MustCompile(`^# \d{4}-\d{2}-\d{2}\n`)
	if !re.Match(prevContent) {
		log.Println("Skipping update of previous entry's heading")
		log.Println("Note: to automatically update the heading, the first line must be of the form '# YYYY-MM-DD'")
	}
	heading := makeHeading(time.Now())
	content := re.ReplaceAll(prevContent, []byte(heading))

	if options.DeclareBankruptcy != nil {
		// If the user specified a level, then we need to clear all lines
		// in sections deeper than that level
		log.Printf("Declaring bankruptcy for sections deeper than level %d", *options.DeclareBankruptcy)
		content = UndergoBankruptcy(content, *options.DeclareBankruptcy)
	}

	// Write the contents to today's file
	if err = os.WriteFile(todayPath, content, OS_ALL_R|OS_USER_RW); err != nil {
		return err
	}

	return nil
}

// ForwardingOptions is a struct that contains options for forwarding
type ForwardingOptions struct {
	DeclareBankruptcy *int
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

// see: https://stackoverflow.com/a/42718395/2889677
const (
	OS_READ        = 04
	OS_WRITE       = 02
	OS_EX          = 01
	OS_USER_SHIFT  = 6
	OS_GROUP_SHIFT = 3
	OS_OTH_SHIFT   = 0

	OS_USER_R   = OS_READ << OS_USER_SHIFT
	OS_USER_W   = OS_WRITE << OS_USER_SHIFT
	OS_USER_X   = OS_EX << OS_USER_SHIFT
	OS_USER_RW  = OS_USER_R | OS_USER_W
	OS_USER_RWX = OS_USER_RW | OS_USER_X

	OS_GROUP_R   = OS_READ << OS_GROUP_SHIFT
	OS_GROUP_W   = OS_WRITE << OS_GROUP_SHIFT
	OS_GROUP_X   = OS_EX << OS_GROUP_SHIFT
	OS_GROUP_RW  = OS_GROUP_R | OS_GROUP_W
	OS_GROUP_RWX = OS_GROUP_RW | OS_GROUP_X

	OS_OTH_R   = OS_READ << OS_OTH_SHIFT
	OS_OTH_W   = OS_WRITE << OS_OTH_SHIFT
	OS_OTH_X   = OS_EX << OS_OTH_SHIFT
	OS_OTH_RW  = OS_OTH_R | OS_OTH_W
	OS_OTH_RWX = OS_OTH_RW | OS_OTH_X

	OS_ALL_R   = OS_USER_R | OS_GROUP_R | OS_OTH_R
	OS_ALL_W   = OS_USER_W | OS_GROUP_W | OS_OTH_W
	OS_ALL_X   = OS_USER_X | OS_GROUP_X | OS_OTH_X
	OS_ALL_RW  = OS_ALL_R | OS_ALL_W
	OS_ALL_RWX = OS_ALL_RW | OS_GROUP_X
)
