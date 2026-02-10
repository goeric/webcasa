// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"strings"
	"testing"
	"time"
)

const testDate = "2026-02-15"

func TestCalendarGridRendersMonth(t *testing.T) {
	styles := DefaultStyles()
	cal := calendarState{
		Cursor:   time.Date(2026, 2, 15, 0, 0, 0, 0, time.Local),
		HasValue: false,
	}
	grid := calendarGrid(cal, styles)
	if !strings.Contains(grid, "February 2026") {
		t.Fatalf("expected month header, got:\n%s", grid)
	}
	if !strings.Contains(grid, "Su Mo Tu We Th Fr Sa") {
		t.Fatal("expected day-of-week labels")
	}
	// Feb 2026 has 28 days.
	if !strings.Contains(grid, "28") {
		t.Fatal("expected day 28 in February 2026")
	}
}

func TestCalendarMoveDay(t *testing.T) {
	cal := &calendarState{
		Cursor: time.Date(2026, 2, 15, 0, 0, 0, 0, time.Local),
	}
	calendarMove(cal, 1)
	if cal.Cursor.Day() != 16 {
		t.Fatalf("expected day 16, got %d", cal.Cursor.Day())
	}
	calendarMove(cal, -2)
	if cal.Cursor.Day() != 14 {
		t.Fatalf("expected day 14, got %d", cal.Cursor.Day())
	}
}

func TestCalendarMoveWeek(t *testing.T) {
	cal := &calendarState{
		Cursor: time.Date(2026, 2, 15, 0, 0, 0, 0, time.Local),
	}
	calendarMove(cal, 7)
	if cal.Cursor.Day() != 22 {
		t.Fatalf("expected day 22, got %d", cal.Cursor.Day())
	}
}

func TestCalendarMoveMonth(t *testing.T) {
	cal := &calendarState{
		Cursor: time.Date(2026, 2, 15, 0, 0, 0, 0, time.Local),
	}
	calendarMoveMonth(cal, 1)
	if cal.Cursor.Month() != time.March {
		t.Fatalf("expected March, got %v", cal.Cursor.Month())
	}
	calendarMoveMonth(cal, -2)
	if cal.Cursor.Month() != time.January {
		t.Fatalf("expected January, got %v", cal.Cursor.Month())
	}
}

func TestCalendarMoveCrossesMonthBoundary(t *testing.T) {
	cal := &calendarState{
		Cursor: time.Date(2026, 1, 31, 0, 0, 0, 0, time.Local),
	}
	calendarMove(cal, 1)
	if cal.Cursor.Month() != time.February || cal.Cursor.Day() != 1 {
		t.Fatalf("expected Feb 1, got %s", cal.Cursor.Format("Jan 2"))
	}
}

func TestDaysIn(t *testing.T) {
	tests := []struct {
		year  int
		month time.Month
		want  int
	}{
		{2026, time.February, 28},
		{2024, time.February, 29}, // leap year
		{2026, time.January, 31},
		{2026, time.April, 30},
	}
	for _, tt := range tests {
		got := daysIn(tt.year, tt.month)
		if got != tt.want {
			t.Errorf(
				"daysIn(%d, %v) = %d, want %d",
				tt.year,
				tt.month,
				got,
				tt.want,
			)
		}
	}
}

func TestSameDay(t *testing.T) {
	a := time.Date(2026, 2, 10, 9, 30, 0, 0, time.UTC)
	b := time.Date(2026, 2, 10, 18, 0, 0, 0, time.UTC)
	c := time.Date(2026, 2, 11, 9, 30, 0, 0, time.UTC)
	if !sameDay(a, b) {
		t.Fatal("expected same day")
	}
	if sameDay(a, c) {
		t.Fatal("expected different day")
	}
}

func TestCalendarKeyNavigation(t *testing.T) {
	m := newTestModel()
	dateVal := testDate
	m.openCalendar(&dateVal, nil)
	if m.calendar == nil {
		t.Fatal("expected calendar to be open")
	}
	if m.calendar.Cursor.Day() != 15 {
		t.Fatalf("expected cursor on 15, got %d", m.calendar.Cursor.Day())
	}

	// Move right (l).
	sendKey(m, "l")
	if m.calendar.Cursor.Day() != 16 {
		t.Fatalf("expected 16 after l, got %d", m.calendar.Cursor.Day())
	}

	// Move down (j) = +7 days.
	sendKey(m, "j")
	if m.calendar.Cursor.Day() != 23 {
		t.Fatalf("expected 23 after j, got %d", m.calendar.Cursor.Day())
	}

	// Move left (h).
	sendKey(m, "h")
	if m.calendar.Cursor.Day() != 22 {
		t.Fatalf("expected 22 after h, got %d", m.calendar.Cursor.Day())
	}

	// Move up (k) = -7 days.
	sendKey(m, "k")
	if m.calendar.Cursor.Day() != 15 {
		t.Fatalf("expected 15 after k, got %d", m.calendar.Cursor.Day())
	}
}

func TestCalendarConfirmWritesDate(t *testing.T) {
	m := newTestModel()
	dateVal := ""
	confirmed := false
	m.openCalendar(&dateVal, func() { confirmed = true })

	// Navigate to a specific date.
	m.calendar.Cursor = time.Date(
		2026, 3, 20, 0, 0, 0, 0, time.Local,
	)
	sendKey(m, "enter")

	if dateVal != "2026-03-20" {
		t.Fatalf("expected date 2026-03-20, got %q", dateVal)
	}
	if !confirmed {
		t.Fatal("expected onConfirm to be called")
	}
	if m.calendar != nil {
		t.Fatal("expected calendar to be closed")
	}
}

func TestCalendarEscCancels(t *testing.T) {
	m := newTestModel()
	dateVal := testDate
	m.openCalendar(&dateVal, nil)
	sendKey(m, "esc")
	if m.calendar != nil {
		t.Fatal("expected calendar to be nil after esc")
	}
	if dateVal != testDate {
		t.Fatalf("expected value unchanged, got %q", dateVal)
	}
}

func TestCalendarRendersInView(t *testing.T) {
	m := newTestModel()
	m.width = 120
	m.height = 40
	dateVal := testDate
	m.openCalendar(&dateVal, nil)

	view := m.buildView()
	if !strings.Contains(view, "February 2026") {
		t.Fatal("expected calendar month in view")
	}
	if !strings.Contains(view, "enter pick") {
		t.Fatal("expected calendar hints in view")
	}
}

func TestCalendarMonthNavigation(t *testing.T) {
	m := newTestModel()
	dateVal := testDate
	m.openCalendar(&dateVal, nil)

	// H = previous month.
	sendKey(m, "H")
	if m.calendar.Cursor.Month() != time.January {
		t.Fatalf("expected January after H, got %v", m.calendar.Cursor.Month())
	}

	// L = next month.
	sendKey(m, "L")
	sendKey(m, "L")
	if m.calendar.Cursor.Month() != time.March {
		t.Fatalf(
			"expected March after L+L, got %v",
			m.calendar.Cursor.Month(),
		)
	}
}

func TestCalendarYearNavigation(t *testing.T) {
	cal := &calendarState{
		Cursor: time.Date(2026, 2, 15, 0, 0, 0, 0, time.Local),
	}
	calendarMoveYear(cal, 1)
	if cal.Cursor.Year() != 2027 {
		t.Fatalf("expected 2027, got %d", cal.Cursor.Year())
	}
	if cal.Cursor.Month() != time.February || cal.Cursor.Day() != 15 {
		t.Fatalf("expected Feb 15, got %s", cal.Cursor.Format("Jan 2"))
	}
	calendarMoveYear(cal, -2)
	if cal.Cursor.Year() != 2025 {
		t.Fatalf("expected 2025, got %d", cal.Cursor.Year())
	}
}

func TestOpenCalendarWithEmptyValue(t *testing.T) {
	m := newTestModel()
	dateVal := ""
	m.openCalendar(&dateVal, nil)
	if m.calendar == nil {
		t.Fatal("expected calendar to be open")
	}
	// Should default to today.
	if !sameDay(m.calendar.Cursor, time.Now()) {
		t.Fatalf(
			"expected cursor on today, got %s",
			m.calendar.Cursor.Format("2006-01-02"),
		)
	}
	if m.calendar.HasValue {
		t.Fatal("expected HasValue to be false for empty input")
	}
}
