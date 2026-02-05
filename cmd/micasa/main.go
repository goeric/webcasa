package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/micasa/micasa/internal/app"
	"github.com/micasa/micasa/internal/data"
)

func main() {
	dbOverride, verbosity, showHelp, err := parseArgs(os.Args[1:])
	if err != nil {
		fail("parse args", err)
	}
	if showHelp {
		printHelp()
		return
	}
	dbPath, err := resolveDBPath(dbOverride)
	if err != nil {
		fail("resolve db path", err)
	}
	store, err := data.Open(dbPath)
	if err != nil {
		fail("open database", err)
	}
	if err := store.AutoMigrate(); err != nil {
		fail("migrate database", err)
	}
	if err := store.SeedDefaults(); err != nil {
		fail("seed defaults", err)
	}
	model, err := app.NewModel(store, app.Options{Verbosity: verbosity})
	if err != nil {
		fail("initialize app", err)
	}
	if _, err := tea.NewProgram(model, tea.WithAltScreen()).Run(); err != nil {
		fail("run app", err)
	}
}

func parseArgs(args []string) (string, int, bool, error) {
	var dbPath string
	verbosity := 0
	for _, arg := range args {
		switch arg {
		case "-h", "--help":
			return "", verbosity, true, nil
		case "-v", "--verbose":
			verbosity++
		default:
			trimmed := strings.TrimLeft(arg, "-")
			if strings.HasPrefix(arg, "-") && strings.Trim(trimmed, "v") == "" {
				verbosity += len(trimmed)
				continue
			}
			if strings.HasPrefix(arg, "-") {
				return "", 0, false, fmt.Errorf("unknown flag: %s", arg)
			}
			if dbPath != "" {
				return "", 0, false, fmt.Errorf("too many arguments")
			}
			dbPath = arg
		}
	}
	return dbPath, verbosity, false, nil
}

func resolveDBPath(override string) (string, error) {
	if override != "" {
		return override, nil
	}
	return data.DefaultDBPath()
}

func printHelp() {
	lines := []string{
		"micasa - home improvement tracker",
		"",
		"Usage:",
		"  micasa [db-path] [--help] [-v|-vv|-vvv]",
		"",
		"Options:",
		"  -h, --help    Show help and exit.",
		"  -v, --verbose Show logs (repeat for more detail).",
		"",
		"Args:",
		"  db-path       Override default sqlite path.",
		"",
		"Environment:",
		"  MICASA_DB_PATH  Override default sqlite path.",
	}
	_, _ = fmt.Fprintln(os.Stdout, strings.Join(lines, "\n"))
}

func fail(context string, err error) {
	fmt.Fprintf(os.Stderr, "micasa: %s: %v\n", context, err)
	os.Exit(1)
}
