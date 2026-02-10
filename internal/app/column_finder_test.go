// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestFuzzyMatch_ExactPrefix(t *testing.T) {
	score, positions := fuzzyMatch("Pro", "Projects")
	if score == 0 {
		t.Fatal("expected match")
	}
	if len(positions) != 3 || positions[0] != 0 || positions[1] != 1 || positions[2] != 2 {
		t.Errorf("positions = %v, want [0,1,2]", positions)
	}
}

func TestFuzzyMatch_CaseInsensitive(t *testing.T) {
	score, _ := fuzzyMatch("pro", "Projects")
	if score == 0 {
		t.Fatal("expected case-insensitive match")
	}
}

func TestFuzzyMatch_NonContiguous(t *testing.T) {
	score, positions := fuzzyMatch("pj", "Projects")
	if score == 0 {
		t.Fatal("expected match for non-contiguous chars")
	}
	if len(positions) != 2 {
		t.Fatalf("positions length = %d, want 2", len(positions))
	}
	if positions[0] != 0 {
		t.Errorf("first match at %d, want 0", positions[0])
	}
}

func TestFuzzyMatch_NoMatch(t *testing.T) {
	score, _ := fuzzyMatch("xyz", "Projects")
	if score != 0 {
		t.Fatal("expected no match")
	}
}

func TestFuzzyMatch_EmptyQuery(t *testing.T) {
	score, _ := fuzzyMatch("", "Projects")
	if score == 0 {
		t.Fatal("empty query should match everything")
	}
}

func TestFuzzyMatch_QueryLongerThanTarget(t *testing.T) {
	score, _ := fuzzyMatch("very long query", "ID")
	if score != 0 {
		t.Fatal("query longer than target should not match")
	}
}

func TestFuzzyMatch_PrefixScoresHigher(t *testing.T) {
	prefixScore, _ := fuzzyMatch("na", "Name")
	midScore, _ := fuzzyMatch("na", "Maintenance")
	if prefixScore <= midScore {
		t.Errorf("prefix match (%d) should score higher than mid match (%d)", prefixScore, midScore)
	}
}

func TestSortFuzzyMatches_ScoreDescending(t *testing.T) {
	matches := []columnFinderMatch{
		{Entry: columnFinderEntry{FullIndex: 0, Title: "A"}, Score: 10},
		{Entry: columnFinderEntry{FullIndex: 1, Title: "B"}, Score: 30},
		{Entry: columnFinderEntry{FullIndex: 2, Title: "C"}, Score: 20},
	}
	sortFuzzyMatches(matches)
	if matches[0].Score != 30 || matches[1].Score != 20 || matches[2].Score != 10 {
		t.Errorf("expected descending scores, got %d %d %d",
			matches[0].Score, matches[1].Score, matches[2].Score)
	}
}

func TestSortFuzzyMatches_TiebreakByIndex(t *testing.T) {
	matches := []columnFinderMatch{
		{Entry: columnFinderEntry{FullIndex: 5, Title: "E"}, Score: 10},
		{Entry: columnFinderEntry{FullIndex: 2, Title: "B"}, Score: 10},
	}
	sortFuzzyMatches(matches)
	if matches[0].Entry.FullIndex != 2 || matches[1].Entry.FullIndex != 5 {
		t.Error("equal scores should sort by FullIndex ascending")
	}
}

func TestColumnFinderState_RefilterEmpty(t *testing.T) {
	cf := &columnFinderState{
		All: []columnFinderEntry{
			{FullIndex: 0, Title: "ID"},
			{FullIndex: 1, Title: "Name"},
			{FullIndex: 2, Title: "Status"},
		},
	}
	cf.refilter()
	if len(cf.Matches) != 3 {
		t.Fatalf("empty query should show all %d columns, got %d", 3, len(cf.Matches))
	}
}

func TestColumnFinderState_RefilterNarrows(t *testing.T) {
	cf := &columnFinderState{
		All: []columnFinderEntry{
			{FullIndex: 0, Title: "ID"},
			{FullIndex: 1, Title: "Name"},
			{FullIndex: 2, Title: "Status"},
			{FullIndex: 3, Title: "Maintenance"},
		},
	}
	cf.Query = "na"
	cf.refilter()
	if len(cf.Matches) != 2 {
		t.Fatalf("expected 2 matches for 'na', got %d", len(cf.Matches))
	}
	// "Name" should score higher than "Maintenance" because of prefix.
	if cf.Matches[0].Entry.Title != "Name" {
		t.Errorf("expected Name first, got %s", cf.Matches[0].Entry.Title)
	}
}

func TestColumnFinderState_CursorClamps(t *testing.T) {
	cf := &columnFinderState{
		All: []columnFinderEntry{
			{FullIndex: 0, Title: "ID"},
			{FullIndex: 1, Title: "Name"},
		},
		Cursor: 5,
	}
	cf.refilter()
	if cf.Cursor != 1 {
		t.Errorf("cursor should clamp to %d, got %d", 1, cf.Cursor)
	}
}

func TestOpenColumnFinder(t *testing.T) {
	m := newTestModel()
	m.openColumnFinder()
	if m.columnFinder == nil {
		t.Fatal("columnFinder should be non-nil after open")
	}
	if len(m.columnFinder.All) == 0 {
		t.Fatal("columnFinder.All should have entries")
	}
}

func TestColumnFinderJump(t *testing.T) {
	m := newTestModel()
	tab := m.effectiveTab()
	if tab == nil {
		t.Fatal("no effective tab")
	}
	origCol := tab.ColCursor

	m.openColumnFinder()
	// Move cursor to the last match and jump.
	cf := m.columnFinder
	cf.Cursor = len(cf.Matches) - 1
	targetIdx := cf.Matches[cf.Cursor].Entry.FullIndex

	m.columnFinderJump()
	if m.columnFinder != nil {
		t.Fatal("columnFinder should be nil after jump")
	}
	if tab.ColCursor == origCol && origCol != targetIdx {
		t.Error("ColCursor should have moved")
	}
	if tab.ColCursor != targetIdx {
		t.Errorf("ColCursor = %d, want %d", tab.ColCursor, targetIdx)
	}
}

func TestColumnFinderJump_UnhidesHiddenColumn(t *testing.T) {
	m := newTestModel()
	tab := m.effectiveTab()
	if tab == nil || len(tab.Specs) < 3 {
		t.Skip("need at least 3 columns")
	}

	// Hide column 2.
	tab.Specs[2].HideOrder = 1

	m.openColumnFinder()
	cf := m.columnFinder

	// Find the match for the hidden column.
	found := false
	for i, match := range cf.Matches {
		if match.Entry.FullIndex == 2 {
			cf.Cursor = i
			found = true
			break
		}
	}
	if !found {
		t.Fatal("hidden column should still appear in finder")
	}

	m.columnFinderJump()
	if tab.Specs[2].HideOrder != 0 {
		t.Error("jumping to hidden column should unhide it")
	}
	if tab.ColCursor != 2 {
		t.Errorf("ColCursor = %d, want 2", tab.ColCursor)
	}
}

func TestHandleColumnFinderKey_EscCloses(t *testing.T) {
	m := newTestModel()
	m.openColumnFinder()
	m.handleColumnFinderKey(tea.KeyMsg{Type: tea.KeyEscape})
	if m.columnFinder != nil {
		t.Fatal("esc should close the column finder")
	}
}

func TestHandleColumnFinderKey_Typing(t *testing.T) {
	m := newTestModel()
	m.openColumnFinder()
	cf := m.columnFinder
	initial := len(cf.Matches)

	// Type "st" to filter.
	m.handleColumnFinderKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m.handleColumnFinderKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})

	if cf.Query != "st" {
		t.Errorf("query = %q, want %q", cf.Query, "st")
	}
	if len(cf.Matches) >= initial && initial > 1 {
		t.Error("typing should narrow matches")
	}
}

func TestHandleColumnFinderKey_Backspace(t *testing.T) {
	m := newTestModel()
	m.openColumnFinder()
	cf := m.columnFinder

	m.handleColumnFinderKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m.handleColumnFinderKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	if cf.Query != "ab" {
		t.Fatalf("query = %q, want %q", cf.Query, "ab")
	}

	m.handleColumnFinderKey(tea.KeyMsg{Type: tea.KeyBackspace})
	if cf.Query != "a" {
		t.Errorf("after backspace query = %q, want %q", cf.Query, "a")
	}
}

func TestHandleColumnFinderKey_CtrlU(t *testing.T) {
	m := newTestModel()
	m.openColumnFinder()
	cf := m.columnFinder

	m.handleColumnFinderKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	m.handleColumnFinderKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	m.handleColumnFinderKey(tea.KeyMsg{Type: tea.KeyCtrlU})
	if cf.Query != "" {
		t.Errorf("ctrl+u should clear query, got %q", cf.Query)
	}
}

func TestHandleColumnFinderKey_Navigation(t *testing.T) {
	m := newTestModel()
	m.openColumnFinder()
	cf := m.columnFinder
	if len(cf.Matches) < 2 {
		t.Skip("need at least 2 columns")
	}

	if cf.Cursor != 0 {
		t.Fatalf("initial cursor = %d, want 0", cf.Cursor)
	}

	m.handleColumnFinderKey(tea.KeyMsg{Type: tea.KeyDown})
	if cf.Cursor != 1 {
		t.Errorf("after down cursor = %d, want 1", cf.Cursor)
	}

	m.handleColumnFinderKey(tea.KeyMsg{Type: tea.KeyUp})
	if cf.Cursor != 0 {
		t.Errorf("after up cursor = %d, want 0", cf.Cursor)
	}

	// Should clamp at top.
	m.handleColumnFinderKey(tea.KeyMsg{Type: tea.KeyUp})
	if cf.Cursor != 0 {
		t.Errorf("should clamp at 0, got %d", cf.Cursor)
	}
}

func TestBuildColumnFinderOverlay_ShowsColumns(t *testing.T) {
	m := newTestModel()
	m.width = 80
	m.height = 24
	m.openColumnFinder()
	rendered := m.buildColumnFinderOverlay()
	if rendered == "" {
		t.Fatal("overlay should not be empty")
	}
	if !strings.Contains(rendered, "Jump to Column") {
		t.Error("overlay should contain title")
	}
}

func TestHighlightFuzzyMatch(t *testing.T) {
	styles := DefaultStyles()
	match := columnFinderMatch{
		Entry:     columnFinderEntry{Title: "Status"},
		Score:     50,
		Positions: []int{0, 1},
	}
	result := highlightFuzzyMatch(match, styles)
	// Should contain the full title text somewhere in the styled output.
	if !strings.Contains(result, "St") {
		t.Error("highlighted output should contain matched chars")
	}
	if !strings.Contains(result, "atus") {
		t.Error("highlighted output should contain unmatched chars")
	}
}

func TestSlashBlockedOnDashboard(t *testing.T) {
	m := newTestModel()
	m.showDashboard = true
	cmd, handled := m.handleDashboardKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	if !handled {
		t.Error("/ should be blocked on dashboard")
	}
	if cmd != nil {
		t.Error("blocked key should return nil cmd")
	}
}

func TestSlashOpensColumnFinder(t *testing.T) {
	m := newTestModel()
	sendKey(m, "/")
	if m.columnFinder == nil {
		t.Fatal("/ in Normal mode should open column finder")
	}
}

func TestSlashBlockedInEditMode(t *testing.T) {
	m := newTestModel()
	m.mode = modeEdit
	sendKey(m, "/")
	if m.columnFinder != nil {
		t.Fatal("/ should not open column finder in Edit mode")
	}
}

func TestColumnFinderEnterJumps(t *testing.T) {
	m := newTestModel()
	m.openColumnFinder()
	cf := m.columnFinder
	if len(cf.Matches) < 2 {
		t.Skip("need at least 2 columns")
	}
	// Move to second match.
	cf.Cursor = 1
	target := cf.Matches[1].Entry.FullIndex

	m.handleColumnFinderKey(tea.KeyMsg{Type: tea.KeyEnter})
	if m.columnFinder != nil {
		t.Fatal("enter should close column finder")
	}
	tab := m.effectiveTab()
	if tab.ColCursor != target {
		t.Errorf("ColCursor = %d, want %d", tab.ColCursor, target)
	}
}
