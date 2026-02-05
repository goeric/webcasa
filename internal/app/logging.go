package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/dlclark/regexp2"
	"github.com/dlclark/regexp2/syntax"
)

type logLevel int

const (
	logError logLevel = iota
	logInfo
	logDebug
)

func (l logLevel) String() string {
	switch l {
	case logError:
		return "ERROR"
	case logDebug:
		return "DEBUG"
	default:
		return "INFO"
	}
}

type logEntry struct {
	Time    time.Time
	Level   logLevel
	Message string
}

type logState struct {
	enabled    bool
	focus      bool
	verbosity  int
	maxLevel   logLevel
	maxEntries int
	input      textinput.Model
	filter     *regexp2.Regexp
	filterErr  error
	entries    []logEntry
}

func newLogState(verbosity int) logState {
	if verbosity <= 0 {
		return logState{}
	}
	input := textinput.New()
	input.Prompt = ""
	input.Placeholder = "type a Perl-compatible regex"
	input.CharLimit = 256
	input.Width = 32
	maxLevel := logInfo
	if verbosity >= 2 {
		maxLevel = logDebug
	}
	return logState{
		enabled:    true,
		verbosity:  verbosity,
		maxLevel:   maxLevel,
		maxEntries: 500,
		input:      input,
	}
}

func (l *logState) setFilter(pattern string) {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		l.filter = nil
		l.filterErr = nil
		return
	}
	re, err := regexp2.Compile(pattern, 0)
	if err != nil {
		l.filterErr = err
		l.filter = nil
		return
	}
	l.filter = re
	l.filterErr = nil
}

func (l *logState) append(level logLevel, message string) {
	if !l.enabled || level > l.maxLevel {
		return
	}
	message = strings.TrimSpace(message)
	if message == "" {
		return
	}
	entry := logEntry{
		Time:    time.Now(),
		Level:   level,
		Message: message,
	}
	l.entries = append(l.entries, entry)
	if len(l.entries) > l.maxEntries {
		l.entries = l.entries[len(l.entries)-l.maxEntries:]
	}
}

func (l *logState) matches(line string) bool {
	if l.filterErr != nil || l.filter == nil {
		return true
	}
	ok, err := l.filter.MatchString(line)
	if err != nil {
		return false
	}
	return ok
}

func (l *logState) validityLabel() string {
	if l.filterErr != nil {
		if parseErr, ok := l.filterErr.(*syntax.Error); ok {
			return fmt.Sprintf("invalid: %s", parseErr.Code.String())
		}
		message := l.filterErr.Error()
		message = strings.TrimPrefix(message, "error parsing regexp: ")
		message = strings.TrimPrefix(message, "error parsing regex: ")
		return fmt.Sprintf("invalid: %s", message)
	}
	if strings.TrimSpace(l.input.Value()) == "" {
		return "no filter"
	}
	return "valid"
}
