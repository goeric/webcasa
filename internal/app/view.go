// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/micasa/micasa/internal/data"
)

func (m *Model) buildView() string {
	if m.showHelp {
		return m.helpFullScreen()
	}
	if m.mode == modeForm && m.form != nil && m.formKind == formHouse {
		return m.formFullScreen()
	}

	house := m.houseView()
	var tabs, tabLine string
	if m.detail != nil {
		tabs = m.breadcrumbView()
		tabLine = m.tabUnderline()
	} else {
		tabs = m.tabsView()
		tabLine = m.tabUnderline()
	}
	content := ""
	if m.mode == modeForm && m.form != nil {
		content = m.form.View()
	} else if tab := m.effectiveTab(); tab != nil {
		content = m.tableView(tab)
	}
	status := m.statusView()

	// Assemble upper portion with intentional spacing.
	upper := lipgloss.JoinVertical(lipgloss.Left, house, "", tabs, tabLine)
	if content != "" {
		upper = lipgloss.JoinVertical(lipgloss.Left, upper, content)
	}

	// Anchor the status bar to the terminal bottom.
	upperH := lipgloss.Height(upper)
	statusH := lipgloss.Height(status)
	gap := m.height - upperH - statusH + 1
	if gap < 1 {
		gap = 1
	}

	var b strings.Builder
	b.WriteString(upper)
	b.WriteString(strings.Repeat("\n", gap))
	b.WriteString(status)
	return b.String()
}

func (m *Model) houseView() string {
	if !m.hasHouse {
		content := lipgloss.JoinVertical(
			lipgloss.Left,
			joinInline(
				m.styles.HeaderTitle.Render("House"),
				m.styles.HeaderBadge.Render("setup"),
				m.keycap("H"),
			),
			m.styles.HeaderHint.Render("Complete the form to add a house profile."),
		)
		return m.headerBox(content)
	}
	if m.showHouse {
		return m.headerBox(m.houseExpanded())
	}
	return m.headerBox(m.houseCollapsed())
}

func (m *Model) houseCollapsed() string {
	title := m.styles.HeaderTitle.Render("House")
	badge := m.styles.HeaderBadge.Render("▸")
	sep := m.styles.HeaderHint.Render(" · ")
	hint := m.styles.HeaderHint
	val := m.styles.HeaderValue
	stats := joinStyledParts(sep,
		styledPart(val, m.house.Nickname),
		styledPart(hint, formatCityState(m.house)),
		styledPart(hint, bedBathLabel(m.house)),
		styledPart(hint, sqftLabel(m.house.SquareFeet)),
		styledPart(hint, formatInt(m.house.YearBuilt)),
	)
	return joinInline(title, badge) + "  " + stats
}

func (m *Model) houseExpanded() string {
	title := m.styles.HeaderTitle.Render("House")
	badge := m.styles.HeaderBadge.Render("▾")
	hint := m.styles.HeaderHint
	val := m.styles.HeaderValue
	sep := hint.Render(" · ")

	identity := joinStyledParts(sep,
		styledPart(val, m.house.Nickname),
		styledPart(hint, formatAddress(m.house)),
	)
	titleLine := joinInline(title, badge) + "  " + identity

	structNums := joinStyledParts(sep,
		styledPart(val, formatInt(m.house.YearBuilt)),
		styledPart(val, sqftLabel(m.house.SquareFeet)),
		styledPart(val, lotLabel(m.house.LotSquareFeet)),
		styledPart(val, bedBathLabel(m.house)),
	)
	structMaterials := joinStyledParts(sep,
		m.hlv("fnd", m.house.FoundationType),
		m.hlv("wir", m.house.WiringType),
		m.hlv("roof", m.house.RoofType),
		m.hlv("ext", m.house.ExteriorType),
		m.hlv("bsmt", m.house.BasementType),
	)
	structure := m.houseSection("Structure", structNums, structMaterials)

	utilLine := joinStyledParts(sep,
		m.hlv("heat", m.house.HeatingType),
		m.hlv("cool", m.house.CoolingType),
		m.hlv("water", m.house.WaterSource),
		m.hlv("sewer", m.house.SewerType),
		m.hlv("park", m.house.ParkingType),
	)
	utilities := m.houseSection("Utilities", utilLine)

	finLine1 := joinStyledParts(sep,
		m.hlv("ins", m.house.InsuranceCarrier),
		m.hlv("policy", m.house.InsurancePolicy),
		m.hlv("renew", data.FormatDate(m.house.InsuranceRenewal)),
	)
	finLine2 := joinStyledParts(sep,
		m.hlv("tax", data.FormatOptionalCents(m.house.PropertyTaxCents)),
		m.hlv("hoa", hoaSummary(m.house)),
	)
	financial := m.houseSection("Financial", finLine1, finLine2)

	content := joinVerticalNonEmpty(titleLine, "", structure, utilities, financial)
	art := m.houseArt()
	if art == "" {
		return content
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, content, "   ", art)
}

func (m *Model) tabsView() string {
	pinned := m.mode == modeEdit
	tabs := make([]string, 0, len(m.tabs))
	for i, tab := range m.tabs {
		if i == m.active {
			tabs = append(tabs, m.styles.TabActive.Render(tab.Name))
		} else if pinned {
			tabs = append(tabs, m.styles.TabLocked.Render(tab.Name))
		} else {
			tabs = append(tabs, m.styles.TabInactive.Render(tab.Name))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, tabs...)
}

func (m *Model) breadcrumbView() string {
	if m.detail == nil {
		return ""
	}
	arrow := m.styles.BreadcrumbArrow.Render(" > ")
	parts := strings.Split(m.detail.Breadcrumb, " > ")
	rendered := make([]string, len(parts))
	for i, p := range parts {
		if i < len(parts)-1 {
			rendered[i] = m.styles.HeaderHint.Render(p)
		} else {
			rendered[i] = m.styles.Breadcrumb.Render(p)
		}
	}
	crumb := strings.Join(rendered, arrow)
	back := m.styles.HeaderHint.Render(" (")
	back += m.keycap("esc")
	back += m.styles.HeaderHint.Render(" back)")
	return crumb + back
}

func (m *Model) tabUnderline() string {
	return m.styles.TabUnderline.Render(strings.Repeat("━", m.effectiveWidth()))
}

func (m *Model) statusView() string {
	if m.mode == modeForm {
		dirtyIndicator := m.styles.FormClean.Render("○ saved")
		if m.formDirty {
			dirtyIndicator = m.styles.FormDirty.Render("● unsaved")
		}
		help := joinWithSeparator(
			m.helpSeparator(),
			dirtyIndicator,
			m.helpItem("ctrl+s", "save"),
			m.helpItem("esc", "cancel"),
			m.helpItem("ctrl+c", "quit"),
		)
		return m.withStatusMessage(help)
	}

	// Both badges render at the same width to prevent layout shift.
	badgeWidth := lipgloss.Width(m.styles.ModeNormal.Render("NORMAL"))
	modeBadge := m.styles.ModeNormal.Render("NORMAL")
	if m.mode == modeEdit {
		modeBadge = m.styles.ModeEdit.
			Width(badgeWidth).
			Align(lipgloss.Center).
			Render("EDIT")
	}

	var help string
	if m.mode == modeNormal {
		items := []string{modeBadge}
		if m.detail == nil {
			items = append(items, m.helpItem("tab", "switch"))
		}
		items = append(items,
			m.helpItem("h/l", "col"),
			m.helpItem("s", "sort"),
		)
		if hint := m.enterHint(); hint != "" {
			items = append(items, m.helpItem("enter", hint))
		}
		items = append(items,
			m.helpItem("c", "hide col"),
		)
		if hint := m.hiddenHint(); hint != "" {
			items = append(items, hint)
		}
		items = append(items,
			m.helpItem("i", "edit"),
			m.helpItem("H", "house"),
			m.helpItem("?", "help"),
		)
		if m.detail != nil {
			items = append(items, m.helpItem("esc", "back"))
		} else {
			items = append(items, m.helpItem("q", "quit"))
		}
		help = joinWithSeparator(m.helpSeparator(), items...)
	} else {
		help = joinWithSeparator(
			m.helpSeparator(),
			modeBadge,
			m.helpItem("a", "add"),
			m.helpItem("e", m.editHint()),
			m.helpItem("d", "del/restore"),
			m.helpItem("u", "undo"),
			m.helpItem("r", "redo"),
			m.deletedHint(m.effectiveTab()),
			m.helpItem("p", "profile"),
			m.helpItem("esc", "normal"),
		)
	}

	return m.withStatusMessage(help)
}

// hiddenHint returns a status bar label listing hidden column names when any
// are hidden, or "" when all are visible.
func (m *Model) hiddenHint() string {
	tab := m.effectiveTab()
	if tab == nil {
		return ""
	}
	names := hiddenColumnNames(tab.Specs)
	if len(names) == 0 {
		return ""
	}
	label := "hidden: " + strings.Join(names, ", ")
	return m.styles.HeaderHint.Render(label)
}

// hiddenColumnNames returns the titles of all hidden columns.
func hiddenColumnNames(specs []columnSpec) []string {
	var names []string
	for _, s := range specs {
		if s.HideOrder > 0 {
			names = append(names, s.Title)
		}
	}
	return names
}

func (m *Model) deletedHint(tab *Tab) string {
	key := m.keycap("x")
	label := m.styles.HeaderHint.Render("deleted")
	if tab != nil && tab.ShowDeleted {
		label = m.styles.DeletedLabel.Render("deleted")
	}
	return strings.TrimSpace(fmt.Sprintf("%s %s", key, label))
}

// withStatusMessage renders the help line, prepending the status message if set.
func (m *Model) withStatusMessage(helpLine string) string {
	if m.status.Text == "" {
		return helpLine
	}
	style := m.styles.Info
	if m.status.Kind == statusError {
		style = m.styles.Error
	}
	return lipgloss.JoinVertical(lipgloss.Left, style.Render(m.status.Text), helpLine)
}

func (m *Model) editHint() string {
	tab := m.effectiveTab()
	if tab == nil {
		return "edit"
	}
	col := tab.ColCursor
	if col < 0 || col >= len(tab.Specs) {
		return "edit"
	}
	spec := tab.Specs[col]
	// Show "follow link" hint when on a linked cell with a target.
	if spec.Link != nil {
		if c, ok := m.selectedCell(col); ok && c.LinkID > 0 {
			return "follow " + spec.Link.Relation
		}
	}
	if spec.Kind == cellReadonly {
		return "edit"
	}
	return "edit: " + spec.Title
}

// enterHint returns a contextual label for the enter key in Normal mode,
// or "" if enter has no action on the current column.
func (m *Model) enterHint() string {
	tab := m.effectiveTab()
	if tab == nil {
		return ""
	}
	col := tab.ColCursor
	if col < 0 || col >= len(tab.Specs) {
		return ""
	}
	spec := tab.Specs[col]
	if spec.Kind == cellDrilldown && m.detail == nil {
		switch tab.Kind {
		case tabMaintenance:
			return "service log"
		case tabAppliances:
			return "maintenance"
		}
	}
	if spec.Link != nil {
		if c, ok := m.selectedCell(col); ok && c.LinkID > 0 {
			return "follow " + spec.Link.Relation
		}
	}
	return ""
}

func (m *Model) formFullScreen() string {
	formContent := m.form.View()
	status := m.statusView()
	panel := lipgloss.JoinVertical(lipgloss.Left, formContent, "", status)
	return m.centerPanel(panel, 1)
}

func (m *Model) helpFullScreen() string {
	return m.centerPanel(m.helpView(), 0)
}

// centerPanel centers a rendered panel within the terminal dimensions.
// minPadTop sets the minimum top padding (e.g. 1 to keep a gap above forms).
func (m *Model) centerPanel(panel string, minPadTop int) string {
	width := m.effectiveWidth()
	height := m.effectiveHeight()
	panelH := lipgloss.Height(panel)
	panelW := lipgloss.Width(panel)
	padTop := (height - panelH) / 2
	if padTop < minPadTop {
		padTop = minPadTop
	}
	padLeft := (width - panelW) / 2
	if padLeft < 0 {
		padLeft = 0
	}
	lines := strings.Split(panel, "\n")
	var b strings.Builder
	for i := 0; i < padTop; i++ {
		b.WriteString("\n")
	}
	indent := strings.Repeat(" ", padLeft)
	for i, line := range lines {
		b.WriteString(indent)
		b.WriteString(line)
		if i < len(lines)-1 {
			b.WriteString("\n")
		}
	}
	return b.String()
}

func (m *Model) helpView() string {
	type binding struct {
		key  string
		desc string
	}
	sections := []struct {
		title    string
		bindings []binding
	}{
		{
			title: "Normal Mode",
			bindings: []binding{
				{"j/k", "Move through rows"},
				{"h/l", "Move through columns"},
				{"g/G", "Jump to first/last row"},
				{"d/u", "Half page down/up"},
				{"tab/shift+tab", "Switch tabs"},
				{"H", "Toggle house profile"},
				{"s", "Sort by column (cycle asc/desc/off)"},
				{"S", "Clear all sorts"},
				{"enter", "Open detail / follow link"},
				{"c", "Hide current column"},
				{"C", "Show all columns"},
				{"i", "Enter Edit mode"},
				{"?", "Help"},
				{"q", "Quit"},
			},
		},
		{
			title: "Edit Mode",
			bindings: []binding{
				{"a", "Add new entry"},
				{"e", "Edit cell (or full row on ID column)"},
				{"d", "Toggle delete/restore"},
				{"u", "Undo last edit"},
				{"r", "Redo last undo"},
				{"x", "Toggle showing deleted items"},
				{"p", "Edit house profile"},
				{"esc", "Back to Normal mode"},
			},
		},
		{
			title: "Forms",
			bindings: []binding{
				{"1-9", "Jump to Nth option (selects only)"},
				{"ctrl+s", "Save immediately"},
				{"esc", "Cancel and discard"},
			},
		},
	}

	var b strings.Builder
	b.WriteString(m.styles.HeaderTitle.Render(" Keyboard Shortcuts "))
	b.WriteString("\n\n")
	for i, section := range sections {
		b.WriteString(m.styles.HeaderSection.Render(" " + section.title + " "))
		b.WriteString("\n")
		for _, bind := range section.bindings {
			keys := m.renderKeys(bind.key)
			desc := m.styles.HeaderHint.Render(bind.desc)
			b.WriteString(fmt.Sprintf("  %s  %s\n", keys, desc))
		}
		if i < len(sections)-1 {
			b.WriteString("\n")
		}
	}
	b.WriteString("\n")
	b.WriteString(m.styles.HeaderHint.Render("Press "))
	b.WriteString(m.renderKeys("esc"))
	b.WriteString(m.styles.HeaderHint.Render(" or "))
	b.WriteString(m.renderKeys("?"))
	b.WriteString(m.styles.HeaderHint.Render(" to close"))
	if m.dbPath != "" {
		b.WriteString("\n\n")
		b.WriteString(m.styles.HeaderLabel.Render("db"))
		b.WriteString(" ")
		b.WriteString(m.styles.HeaderHint.Render(m.dbPath))
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(accent).
		Padding(1, 2).
		Render(b.String())
	return box
}

// visibleProjection computes the visible-only view of a tab's columns and data.
// It returns projected specs, cell rows, the translated column cursor (-1 if
// the real cursor is hidden), remapped sort entries, and the vis-to-full index map.
func visibleProjection(tab *Tab) (
	specs []columnSpec,
	cellRows [][]cell,
	colCursor int,
	sorts []sortEntry,
	visToFull []int,
) {
	fullToVis := make(map[int]int, len(tab.Specs))
	for i, spec := range tab.Specs {
		if spec.HideOrder > 0 {
			continue
		}
		fullToVis[i] = len(visToFull)
		visToFull = append(visToFull, i)
		specs = append(specs, spec)
	}

	colCursor = -1
	if vis, ok := fullToVis[tab.ColCursor]; ok {
		colCursor = vis
	}

	cellRows = make([][]cell, len(tab.CellRows))
	for r, row := range tab.CellRows {
		projected := make([]cell, 0, len(visToFull))
		for _, fi := range visToFull {
			if fi < len(row) {
				projected = append(projected, row[fi])
			}
		}
		cellRows[r] = projected
	}

	for _, se := range tab.Sorts {
		if vis, ok := fullToVis[se.Col]; ok {
			sorts = append(sorts, sortEntry{Col: vis, Dir: se.Dir})
		}
	}
	return
}

func (m *Model) tableView(tab *Tab) string {
	if tab == nil || len(tab.Specs) == 0 {
		return ""
	}
	visSpecs, visCells, visColCursor, visSorts, visToFull := visibleProjection(tab)
	if len(visSpecs) == 0 {
		return ""
	}

	// Edge ladles: smooth L-shaped borders that cradle the candy pills.
	hasLeading := len(visToFull) > 0 && visToFull[0] > 0
	hasTrailing := len(visToFull) > 0 && visToFull[len(visToFull)-1] < len(tab.Specs)-1
	ladle := ladleChrome(hasLeading, hasTrailing)

	width := m.effectiveWidth() - ladle.width
	normalSep := m.styles.TableSeparator.Render(" │ ")
	normalDiv := m.styles.TableSeparator.Render("─┼─")
	plainSeps, collapsedSeps := gapSeparators(visToFull, len(tab.Specs), normalSep, m.styles)
	widths := columnWidths(visSpecs, visCells, width, lipgloss.Width(normalSep))
	header := renderHeaderRow(visSpecs, widths, collapsedSeps, visColCursor, visSorts, m.styles)
	divider := renderDivider(widths, plainSeps, normalDiv, m.styles.TableSeparator)

	sepWidth := lipgloss.Width(normalSep)
	stacks := computeCollapsedStacks(tab.Specs, visToFull, widths, sepWidth)
	stackLines := renderCollapsedStacks(stacks)
	connector := renderStackConnector(stacks)

	// Column-space width for the ladle bottom curve.
	colSpaceWidth := sumInts(widths)
	if len(widths) > 1 {
		colSpaceWidth += (len(widths) - 1) * sepWidth
	}
	leftWidth := 0
	if hasLeading {
		leftWidth = 2
	}
	bottom := renderLadleBottom(stacks, hasLeading, hasTrailing, leftWidth, colSpaceWidth)

	// Blank spacer for edge stacks when no middle connector already provides
	// breathing room between the table body and the pills.
	needsSpacer := ladle.width > 0 && len(stackLines) > 0 && connector == ""

	// Height accounting: stack lines + connector/spacer + bottom curve.
	stackChrome := len(stackLines)
	if connector != "" || needsSpacer {
		stackChrome++
	}
	if bottom != "" {
		stackChrome++
	}
	effectiveHeight := tab.Table.Height() - stackChrome
	if effectiveHeight < 2 {
		effectiveHeight = 2
	}
	rows := renderRows(
		visSpecs,
		visCells,
		tab.Rows,
		widths,
		plainSeps,
		collapsedSeps,
		tab.Table.Cursor(),
		visColCursor,
		effectiveHeight,
		m.styles,
	)

	// Assemble body (header + divider + data).
	bodyParts := []string{header, divider}
	if len(rows) == 0 {
		bodyParts = append(bodyParts, m.styles.Empty.Render("No entries yet."))
	} else {
		bodyParts = append(bodyParts, strings.Join(rows, "\n"))
	}
	body := joinVerticalNonEmpty(bodyParts...)

	// Right-pad a line to colSpaceWidth so the right │ border aligns.
	padToCol := func(s string) string {
		if ladle.right == "" {
			return s
		}
		gap := colSpaceWidth - lipgloss.Width(s)
		if gap > 0 {
			return s + strings.Repeat(" ", gap)
		}
		return s
	}

	// All wrapped lines share the same ladle border (│) for continuity.
	parts := []string{wrapLines(body, ladle.left, ladle.right)}
	if connector != "" {
		parts = append(parts, ladle.left+padToCol(connector)+ladle.right)
	} else if needsSpacer {
		parts = append(parts, ladle.left+strings.Repeat(" ", colSpaceWidth)+ladle.right)
	}
	for _, sl := range stackLines {
		parts = append(parts, ladle.left+padToCol(sl)+ladle.right)
	}
	if bottom != "" {
		parts = append(parts, bottom)
	}
	return joinVerticalNonEmpty(parts...)
}

// ladleStrings holds the border prefix/suffix for the ladle shape.
type ladleStrings struct {
	left  string // "│ " — applied to body, connector, and stack lines
	right string // " │"
	width int    // total horizontal chars consumed by both sides
}

// ladleChrome computes the ladle border strings for edge hidden columns.
func ladleChrome(hasLeading, hasTrailing bool) ladleStrings {
	ls := ladleStrings{}
	style := lipgloss.NewStyle().Foreground(secondary)
	if hasLeading {
		ls.left = style.Render("│") + " "
		ls.width += 2
	}
	if hasTrailing {
		ls.right = " " + style.Render("│")
		ls.width += 2
	}
	return ls
}

// renderLadleBottom draws the horizontal base of the ladle L-shape, closing
// off the bottom of the candy pill area. ╰──── on the left, ────╯ on the right.
func renderLadleBottom(
	stacks []collapsedStack,
	hasLeading, hasTrailing bool,
	leftWidth, colSpaceWidth int,
) string {
	if !hasLeading && !hasTrailing {
		return ""
	}
	style := lipgloss.NewStyle().Foreground(secondary)

	rightWidth := 0
	if hasTrailing {
		rightWidth = 2
	}
	fullWidth := leftWidth + colSpaceWidth + rightWidth

	// Both edges: single continuous curve spanning the full line width.
	if hasLeading && hasTrailing {
		if fullWidth < 2 {
			return ""
		}
		return style.Render("╰" + strings.Repeat("─", fullWidth-2) + "╯")
	}

	// Left edge only.
	if hasLeading {
		var leadW int
		for _, s := range stacks {
			if s.edge && s.offset == 0 {
				leadW = s.width
				break
			}
		}
		if leadW > 0 {
			return style.Render("╰" + strings.Repeat("─", 1+leadW))
		}
		return ""
	}

	// Right edge only.
	var trailOff int
	for _, s := range stacks {
		if s.edge {
			trailOff = s.offset
			break
		}
	}
	start := leftWidth + trailOff
	dashLen := colSpaceWidth - trailOff + 1
	return strings.Repeat(" ", start) + style.Render(strings.Repeat("─", dashLen)+"╯")
}

// wrapLines prepends prefix and appends suffix to every line in s.
func wrapLines(s, prefix, suffix string) string {
	if prefix == "" && suffix == "" {
		return s
	}
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = prefix + line + suffix
	}
	return strings.Join(lines, "\n")
}

// gapSeparators computes a per-gap separator for the header/data and divider.
// When hidden columns exist between two visible ones, the separator uses ⋯ to
// signal a collapsed region. Returns one separator per gap (len(visToFull)-1).
func gapSeparators(
	visToFull []int,
	totalCols int,
	normalSep string,
	styles Styles,
) (plainSeps, collapsedSeps []string) {
	n := len(visToFull)
	if n <= 1 {
		return nil, nil
	}
	collapsedSep := styles.TableSeparator.Render(" ") +
		lipgloss.NewStyle().Foreground(secondary).Render("⋯") +
		styles.TableSeparator.Render(" ")

	plainSeps = make([]string, n-1)
	collapsedSeps = make([]string, n-1)
	for i := 0; i < n-1; i++ {
		plainSeps[i] = normalSep
		if visToFull[i+1] > visToFull[i]+1 {
			collapsedSeps[i] = collapsedSep
		} else {
			collapsedSeps[i] = normalSep
		}
	}
	return
}

// --- Collapsed column stacks ---

type stackEntry struct {
	name      string
	fullIndex int
	hideOrder int
}

type collapsedStack struct {
	entries []stackEntry // most recent (highest hideOrder) first
	offset  int          // horizontal character offset for pill left edge
	width   int          // pill width including padding
	edge    bool         // true for leading/trailing stacks (ladle handles connector)
}

// Candy palette cycles through the Wong chromatic colors.
var candyPalette = []lipgloss.AdaptiveColor{
	accent, secondary, success, muted, warning, danger,
}

func candyColor(index int) lipgloss.AdaptiveColor {
	return candyPalette[index%len(candyPalette)]
}

// computeCollapsedStacks finds groups of hidden columns between visible ones
// and returns a positioned stack for each group.
func computeCollapsedStacks(
	specs []columnSpec,
	visToFull []int,
	widths []int,
	sepWidth int,
) []collapsedStack {
	n := len(visToFull)
	if n == 0 {
		return nil
	}

	var stacks []collapsedStack

	// Leading hidden columns (before first visible).
	if visToFull[0] > 0 {
		if entries := collectHiddenEntries(specs, 0, visToFull[0]); len(entries) > 0 {
			w := maxEntryWidth(entries) + 2
			stacks = append(stacks, collapsedStack{
				entries: entries, offset: 0, width: w, edge: true,
			})
		}
	}

	// Between visible columns (anchored to ⋯ gaps).
	offset := 0
	for i := 0; i < n; i++ {
		if i > 0 {
			lo := visToFull[i-1] + 1
			hi := visToFull[i]
			if hi > lo {
				if entries := collectHiddenEntries(specs, lo, hi); len(entries) > 0 {
					w := maxEntryWidth(entries) + 2
					gapCenter := offset + sepWidth/2
					pillOff := gapCenter - w/2
					if pillOff < 0 {
						pillOff = 0
					}
					stacks = append(stacks, collapsedStack{
						entries: entries, offset: pillOff, width: w,
					})
				}
			}
			offset += sepWidth
		}
		if i < len(widths) {
			offset += widths[i]
		}
	}

	// Trailing hidden columns (after last visible).
	if last := visToFull[n-1]; last < len(specs)-1 {
		if entries := collectHiddenEntries(specs, last+1, len(specs)); len(entries) > 0 {
			w := maxEntryWidth(entries) + 2
			trailOff := offset - w
			if trailOff < 0 {
				trailOff = 0
			}
			stacks = append(stacks, collapsedStack{
				entries: entries, offset: trailOff, width: w, edge: true,
			})
		}
	}

	// Clamp pill widths so they never exceed the total column space,
	// then merge any stacks that overlap after clamping.
	totalWidth := sumInts(widths)
	if len(widths) > 1 {
		totalWidth += (len(widths) - 1) * sepWidth
	}
	for i := range stacks {
		if stacks[i].width > totalWidth {
			stacks[i].width = totalWidth
		}
		if stacks[i].offset+stacks[i].width > totalWidth {
			stacks[i].offset = totalWidth - stacks[i].width
		}
		if stacks[i].offset < 0 {
			stacks[i].offset = 0
		}
	}

	// Merge overlapping stacks (e.g. leading+trailing both clamped to offset 0).
	merged := stacks[:0]
	for _, s := range stacks {
		if len(merged) > 0 {
			last := &merged[len(merged)-1]
			if s.offset < last.offset+last.width {
				end := last.offset + last.width
				if se := s.offset + s.width; se > end {
					end = se
				}
				last.entries = append(last.entries, s.entries...)
				last.width = end - last.offset
				if last.width > totalWidth {
					last.width = totalWidth
				}
				last.edge = last.edge || s.edge
				continue
			}
		}
		merged = append(merged, s)
	}
	// Re-sort merged entries by column index descending (rightmost on top).
	for i := range merged {
		sort.Slice(merged[i].entries, func(a, b int) bool {
			return merged[i].entries[a].fullIndex > merged[i].entries[b].fullIndex
		})
	}

	return merged
}

// collectHiddenEntries gathers hidden columns in [lo, hi) ordered so that the
// leftmost column is at the bottom of the stack and the rightmost is on top
// (closest to the data rows). This preserves spatial intuition: the column
// nearest the gap edge sits at the top of the visual stack.
func collectHiddenEntries(specs []columnSpec, lo, hi int) []stackEntry {
	var entries []stackEntry
	for i := hi - 1; i >= lo; i-- {
		if specs[i].HideOrder > 0 {
			entries = append(entries, stackEntry{
				name: specs[i].Title, fullIndex: i, hideOrder: specs[i].HideOrder,
			})
		}
	}
	return entries
}

func maxEntryWidth(entries []stackEntry) int {
	w := 0
	for _, e := range entries {
		if ew := lipgloss.Width(e.name); ew > w {
			w = ew
		}
	}
	return w
}

// renderCollapsedStacks renders one line per stack depth, with candy-colored
// pills positioned horizontally at each gap.
func renderCollapsedStacks(stacks []collapsedStack) []string {
	if len(stacks) == 0 {
		return nil
	}
	maxDepth := 0
	for _, s := range stacks {
		if len(s.entries) > maxDepth {
			maxDepth = len(s.entries)
		}
	}
	lines := make([]string, 0, maxDepth)
	for depth := 0; depth < maxDepth; depth++ {
		lines = append(lines, renderStackLine(stacks, depth))
	}
	return lines
}

// renderStackConnector draws a thin vertical line at the center of each
// non-edge stack, connecting the data rows above to the candy pills below.
// Edge stacks are handled by the ladle curves instead.
func renderStackConnector(stacks []collapsedStack) string {
	connStyle := lipgloss.NewStyle().Foreground(secondary)
	var b strings.Builder
	cursor := 0
	any := false
	for _, stack := range stacks {
		if stack.edge {
			continue
		}
		any = true
		center := stack.offset + stack.width/2
		if center > cursor {
			b.WriteString(strings.Repeat(" ", center-cursor))
			cursor = center
		}
		b.WriteString(connStyle.Render("│"))
		cursor++
	}
	if !any {
		return ""
	}
	return b.String()
}

func renderStackLine(stacks []collapsedStack, depth int) string {
	type positioned struct {
		offset int
		text   string
		width  int
	}
	var pills []positioned
	for _, stack := range stacks {
		if depth >= len(stack.entries) {
			continue
		}
		entry := stack.entries[depth]
		style := lipgloss.NewStyle().
			Background(candyColor(entry.fullIndex)).
			Foreground(onAccent).
			Bold(true).
			Width(stack.width).
			Align(lipgloss.Center)
		pills = append(pills, positioned{
			offset: stack.offset,
			text:   style.Render(entry.name),
			width:  stack.width,
		})
	}
	sort.Slice(pills, func(i, j int) bool { return pills[i].offset < pills[j].offset })
	var b strings.Builder
	cursor := 0
	for _, p := range pills {
		if p.offset > cursor {
			b.WriteString(strings.Repeat(" ", p.offset-cursor))
			cursor = p.offset
		}
		b.WriteString(p.text)
		cursor += p.width
	}
	return b.String()
}

func renderHeaderRow(
	specs []columnSpec,
	widths []int,
	separators []string,
	colCursor int,
	sorts []sortEntry,
	styles Styles,
) string {
	cells := make([]string, 0, len(specs))
	for i, spec := range specs {
		width := safeWidth(widths, i)
		title := spec.Title
		if spec.Link != nil {
			title = title + " " + styles.LinkIndicator.Render(spec.Link.Relation)
		}
		indicator := sortIndicator(sorts, i)
		text := formatHeaderCell(title, indicator, width)
		if i == colCursor {
			cells = append(cells, styles.ColActiveHeader.Render(text))
		} else {
			cells = append(cells, styles.TableHeader.Render(text))
		}
	}
	return joinCells(cells, separators)
}

// formatHeaderCell renders a header cell with the title left-aligned and
// the sort indicator right-aligned within the given width. If there's no
// indicator, it falls back to plain left-aligned formatting.
func formatHeaderCell(title, indicator string, width int) string {
	if indicator == "" {
		return formatCell(title, width, alignLeft)
	}
	titleW := lipgloss.Width(title)
	indW := lipgloss.Width(indicator)
	gap := width - titleW - indW
	if gap < 0 {
		// Not enough room; truncate title to make space.
		available := width - indW
		if available < 1 {
			return formatCell(title, width, alignLeft)
		}
		title = ansi.Truncate(title, available, "")
		titleW = lipgloss.Width(title)
		gap = width - titleW - indW
	}
	return title + strings.Repeat(" ", gap) + indicator
}

// sortIndicator returns a string like "▲1" or "▼2" for the column's
// position in the sort stack, or "" if the column is not sorted.
// headerTitleWidth returns the rendered width of a column header including
// any link relation suffix. Sort indicators are rendered within the
// existing column width so toggling sorts never changes the layout.
func headerTitleWidth(spec columnSpec) int {
	w := lipgloss.Width(spec.Title)
	if spec.Link != nil {
		w += 1 + lipgloss.Width(spec.Link.Relation) // " m:1"
	}
	return w
}

func sortIndicator(sorts []sortEntry, col int) string {
	for i, entry := range sorts {
		if entry.Col == col {
			arrow := "\u25b2" // ▲
			if entry.Dir == sortDesc {
				arrow = "\u25bc" // ▼
			}
			if len(sorts) == 1 {
				return arrow
			}
			return fmt.Sprintf("%s%d", arrow, i+1)
		}
	}
	return ""
}

func renderDivider(
	widths []int,
	separators []string,
	divSep string,
	style lipgloss.Style,
) string {
	parts := make([]string, 0, len(widths))
	for _, width := range widths {
		if width < 1 {
			width = 1
		}
		parts = append(parts, style.Render(strings.Repeat("─", width)))
	}
	// Uniform divider separator for all gaps (no ⋯ on the divider line).
	if len(separators) > 0 {
		uniform := make([]string, len(separators))
		for i := range uniform {
			uniform[i] = divSep
		}
		separators = uniform
	}
	return joinCells(parts, separators)
}

func renderRows(
	specs []columnSpec,
	rows [][]cell,
	meta []rowMeta,
	widths []int,
	plainSeps []string,
	collapsedSeps []string,
	cursor int,
	colCursor int,
	height int,
	styles Styles,
) []string {
	total := len(rows)
	if total == 0 {
		return nil
	}
	if height <= 0 {
		height = total
	}
	start, end := visibleRange(total, height, cursor)
	count := end - start
	mid := start + count/2
	rendered := make([]string, 0, count)
	for i := start; i < end; i++ {
		selected := i == cursor
		deleted := i < len(meta) && meta[i].Deleted
		// Show ⋯ on first, middle, and last visible rows only.
		seps := plainSeps
		if i == start || i == mid || i == end-1 {
			seps = collapsedSeps
		}
		row := renderRow(specs, rows[i], widths, seps, selected, deleted, colCursor, styles)
		rendered = append(rendered, row)
	}
	return rendered
}

// cellHighlight describes how a cell should be visually marked.
type cellHighlight int

const (
	highlightNone   cellHighlight = iota
	highlightRow                  // selected row, not the active column
	highlightCursor               // selected row AND active column
)

func renderRow(
	specs []columnSpec,
	row []cell,
	widths []int,
	separators []string,
	selected bool,
	deleted bool,
	colCursor int,
	styles Styles,
) string {
	cells := make([]string, 0, len(specs))
	for i, spec := range specs {
		width := safeWidth(widths, i)
		var cellValue cell
		if i < len(row) {
			cellValue = row[i]
		}
		hl := highlightNone
		if selected && i == colCursor {
			hl = highlightCursor
		} else if selected {
			hl = highlightRow
		}
		rendered := renderCell(cellValue, spec, width, hl, deleted, styles)
		cells = append(cells, rendered)
	}
	return joinCells(cells, separators)
}

func renderCell(
	cellValue cell,
	spec columnSpec,
	width int,
	hl cellHighlight,
	deleted bool,
	styles Styles,
) string {
	if width < 1 {
		width = 1
	}
	value := strings.TrimSpace(cellValue.Value)
	style := cellStyle(cellValue.Kind, styles)
	if value == "" {
		value = "—"
		style = styles.Empty
	} else if cellValue.Kind == cellDrilldown {
		return renderPillCell(value, spec, width, hl, deleted, styles)
	} else if cellValue.Kind == cellStatus {
		if s, ok := styles.StatusStyles[value]; ok {
			style = s
		}
	}

	if deleted {
		style = style.Foreground(textDim).Strikethrough(true).Italic(true)
	}

	// For cursor underline and deleted strikethrough, style just the
	// text and pad separately so the decoration matches text length.
	if hl == highlightCursor || deleted {
		cursorStyle := style
		if hl == highlightCursor {
			cursorStyle = cursorStyle.Underline(true).Bold(true)
		}
		if hl == highlightRow {
			cursorStyle = cursorStyle.Background(surface).Bold(true)
		}
		truncated := ansi.Truncate(value, width, "…")
		styled := cursorStyle.Render(truncated)
		textW := lipgloss.Width(truncated)
		if pad := width - textW; pad > 0 {
			if spec.Align == alignRight {
				return strings.Repeat(" ", pad) + styled
			}
			return styled + strings.Repeat(" ", pad)
		}
		return styled
	}

	if hl == highlightRow {
		style = style.Background(surface).Bold(true)
	}

	aligned := formatCell(value, width, spec.Align)
	return style.Render(aligned)
}

// renderPillCell renders a drilldown value as a compact pill badge,
// right-aligned within the column width.
func renderPillCell(
	value string,
	spec columnSpec,
	width int,
	hl cellHighlight,
	deleted bool,
	styles Styles,
) string {
	style := styles.Drilldown
	if deleted {
		style = lipgloss.NewStyle().
			Foreground(textDim).
			Strikethrough(true).
			Italic(true)
		pill := style.Render(value)
		pillW := lipgloss.Width(pill)
		if pad := width - pillW; pad > 0 {
			return strings.Repeat(" ", pad) + pill
		}
		return pill
	}

	if hl == highlightCursor {
		style = style.Underline(true)
	}

	pill := style.Render(value)
	pillW := lipgloss.Width(pill)

	// Pad to fill the column; pill is always right-aligned.
	if pad := width - pillW; pad > 0 {
		padStyle := lipgloss.NewStyle()
		if hl == highlightRow {
			padStyle = padStyle.Background(surface)
		}
		return padStyle.Render(strings.Repeat(" ", pad)) + pill
	}
	return pill
}

// joinCells joins rendered cell strings using per-gap separators.
// If separators is shorter than needed, falls back to the last separator.
func joinCells(cells []string, separators []string) string {
	if len(cells) == 0 {
		return ""
	}
	var b strings.Builder
	for i, c := range cells {
		if i > 0 {
			idx := i - 1
			if idx < len(separators) {
				b.WriteString(separators[idx])
			} else if len(separators) > 0 {
				b.WriteString(separators[len(separators)-1])
			}
		}
		b.WriteString(c)
	}
	return b.String()
}

func cellStyle(kind cellKind, styles Styles) lipgloss.Style {
	switch kind {
	case cellMoney:
		return styles.Money
	case cellReadonly:
		return styles.Readonly
	case cellDrilldown:
		return styles.Drilldown
	default:
		return lipgloss.NewStyle()
	}
}

func formatCell(value string, width int, align alignKind) string {
	if width < 1 {
		return ""
	}
	truncated := ansi.Truncate(value, width, "…")
	current := lipgloss.Width(truncated)
	if current >= width {
		return truncated
	}
	padding := width - current
	switch align {
	case alignRight:
		return strings.Repeat(" ", padding) + truncated
	default:
		return truncated + strings.Repeat(" ", padding)
	}
}

func visibleRange(total, height, cursor int) (int, int) {
	if total <= height {
		return 0, total
	}
	if cursor < 0 {
		cursor = 0
	}
	if cursor >= total {
		cursor = total - 1
	}
	start := cursor - height/2
	if start < 0 {
		start = 0
	}
	end := start + height
	if end > total {
		end = total
		start = end - height
		if start < 0 {
			start = 0
		}
	}
	return start, end
}

func columnWidths(
	specs []columnSpec,
	rows [][]cell,
	width int,
	separatorWidth int,
) []int {
	columnCount := len(specs)
	if columnCount == 0 {
		return nil
	}
	available := width - separatorWidth*(columnCount-1)
	if available < columnCount {
		available = columnCount
	}

	natural := naturalWidths(specs, rows)

	// If content-driven widths fit, use them — no truncation.
	if sumInts(natural) <= available {
		widths := make([]int, columnCount)
		copy(widths, natural)
		extra := available - sumInts(widths)
		if extra > 0 {
			flex := flexColumns(specs)
			if len(flex) == 0 {
				flex = allColumns(specs)
			}
			distribute(widths, specs, flex, extra, true)
		}
		return widths
	}

	// Content exceeds terminal width — apply Max constraints.
	widths := make([]int, columnCount)
	for i, w := range natural {
		if specs[i].Max > 0 && w > specs[i].Max {
			w = specs[i].Max
		}
		widths[i] = w
	}

	total := sumInts(widths)
	if total <= available {
		// Max-capped fits. Give extra space to truncated columns first
		// so they can show their full content before flex columns grow.
		extra := available - total
		extra = widenTruncated(widths, natural, extra)
		if extra > 0 {
			flex := flexColumns(specs)
			if len(flex) == 0 {
				flex = allColumns(specs)
			}
			distribute(widths, specs, flex, extra, true)
		}
		return widths
	}

	// Still too wide — shrink flex columns.
	deficit := total - available
	flex := flexColumns(specs)
	if len(flex) == 0 {
		flex = allColumns(specs)
	}
	distribute(widths, specs, flex, deficit, false)
	return widths
}

// naturalWidths returns the content-driven width for each column (header,
// fixed values, and actual cell values) floored by Min but NOT capped by Max.
func naturalWidths(specs []columnSpec, rows [][]cell) []int {
	widths := make([]int, len(specs))
	for i, spec := range specs {
		w := headerTitleWidth(spec)
		for _, fv := range spec.FixedValues {
			if fw := lipgloss.Width(fv); fw > w {
				w = fw
			}
		}
		for _, row := range rows {
			if i >= len(row) {
				continue
			}
			value := strings.TrimSpace(row[i].Value)
			if value == "" {
				continue
			}
			if cw := lipgloss.Width(value); cw > w {
				w = cw
			}
		}
		if w < spec.Min {
			w = spec.Min
		}
		widths[i] = w
	}
	return widths
}

// widenTruncated gives extra space to columns whose current width is less than
// their natural (content-driven) width, eliminating truncation where possible.
// Returns the remaining unused extra.
func widenTruncated(widths, natural []int, extra int) int {
	for extra > 0 {
		changed := false
		for i := range widths {
			if extra == 0 {
				break
			}
			if widths[i] < natural[i] {
				widths[i]++
				extra--
				changed = true
			}
		}
		if !changed {
			break
		}
	}
	return extra
}

func distribute(
	widths []int,
	specs []columnSpec,
	indices []int,
	amount int,
	grow bool,
) {
	if amount <= 0 || len(indices) == 0 {
		return
	}
	for amount > 0 {
		changed := false
		for _, idx := range indices {
			if idx >= len(widths) {
				continue
			}
			if grow {
				if specs[idx].Max > 0 && widths[idx] >= specs[idx].Max {
					continue
				}
				widths[idx]++
			} else {
				if widths[idx] <= specs[idx].Min {
					continue
				}
				widths[idx]--
			}
			amount--
			changed = true
			if amount == 0 {
				break
			}
		}
		if !changed {
			return
		}
	}
}

func flexColumns(specs []columnSpec) []int {
	indices := make([]int, 0, len(specs))
	for i, spec := range specs {
		if spec.Flex {
			indices = append(indices, i)
		}
	}
	return indices
}

func allColumns(specs []columnSpec) []int {
	indices := make([]int, len(specs))
	for i := range specs {
		indices[i] = i
	}
	return indices
}

func sumInts(values []int) int {
	total := 0
	for _, value := range values {
		total += value
	}
	return total
}

func safeWidth(widths []int, idx int) int {
	if idx < 0 || idx >= len(widths) {
		return 1
	}
	if widths[idx] < 1 {
		return 1
	}
	return widths[idx]
}

func (m *Model) headerBox(content string) string {
	return m.styles.HeaderBox.Render(content)
}

// houseArt renders a retro pixel-art house for the expanded profile.
// Uses shade characters (░▒▓█) for a classic DOS/BBS aesthetic.
// Returns "" if the terminal is too narrow.
func (m *Model) houseArt() string {
	if m.effectiveWidth() < 80 {
		return ""
	}
	rf := lipgloss.NewStyle().Foreground(accent)    // roof
	wl := lipgloss.NewStyle().Foreground(textMid)   // walls
	wn := lipgloss.NewStyle().Foreground(warning)   // windows (lit)
	dr := lipgloss.NewStyle().Foreground(secondary) // door
	lines := []string{
		rf.Render("      ▄▓▄"),
		rf.Render("    ▄▓▓▓▓▓▄"),
		rf.Render("  ▄▓▓▓▓▓▓▓▓▓▄"),
		wl.Render("  ██ ") + wn.Render("░░") + wl.Render(" ") + wn.Render("░░") + wl.Render(" ██"),
		wl.Render("  ██  ") + dr.Render("████") + wl.Render(" ██"),
		wl.Render("  ██  ") + dr.Render("█  █") + wl.Render(" ██"),
		wl.Render("  ▀▀▀▀▀▀▀▀▀▀▀"),
	}
	return strings.Join(lines, "\n")
}

func (m *Model) helpItem(keys, label string) string {
	keycaps := m.renderKeys(keys)
	desc := m.styles.HeaderHint.Render(label)
	return strings.TrimSpace(fmt.Sprintf("%s %s", keycaps, desc))
}

func (m *Model) helpSeparator() string {
	return m.styles.HeaderHint.Render(" · ")
}

func (m *Model) renderKeys(keys string) string {
	// A bare "/" is a single key, not a separator between two keys.
	if strings.TrimSpace(keys) == "/" {
		return m.keycap("/")
	}
	parts := strings.Split(keys, "/")
	rendered := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		rendered = append(rendered, m.keycap(part))
	}
	return joinWithSeparator(m.styles.HeaderHint.Render(" · "), rendered...)
}

func (m *Model) keycap(value string) string {
	// Detect bare uppercase letters and render as SHIFT+X.
	if len(value) == 1 && value[0] >= 'A' && value[0] <= 'Z' {
		return m.styles.Keycap.Render("SHIFT+" + value)
	}
	return m.styles.Keycap.Render(strings.ToUpper(value))
}

// styledPart returns a styled value, or "" if the underlying value is blank.
func styledPart(style lipgloss.Style, value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}
	return style.Render(value)
}

// joinStyledParts joins pre-styled parts with a separator, skipping empty ones.
func joinStyledParts(sep string, parts ...string) string {
	filtered := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			filtered = append(filtered, p)
		}
	}
	if len(filtered) == 0 {
		return ""
	}
	return strings.Join(filtered, sep)
}

// hlv renders a dim label followed by a bright value, or "" if the value is blank.
func (m *Model) hlv(label, value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}
	return m.styles.HeaderLabel.Render(label) + " " + m.styles.HeaderValue.Render(value)
}

// houseSection renders a section header with values, indenting continuation lines.
func (m *Model) houseSection(header string, lines ...string) string {
	label := m.styles.HeaderSection.Render(header)
	labelWidth := lipgloss.Width(label)
	indent := strings.Repeat(" ", labelWidth+1)
	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			filtered = append(filtered, line)
		}
	}
	if len(filtered) == 0 {
		return ""
	}
	result := make([]string, len(filtered))
	for i, line := range filtered {
		if i == 0 {
			result[i] = label + " " + line
		} else {
			result[i] = indent + line
		}
	}
	return strings.Join(result, "\n")
}

func bedBathLabel(profile data.HouseProfile) string {
	var parts []string
	if profile.Bedrooms > 0 {
		parts = append(parts, fmt.Sprintf("%dbd", profile.Bedrooms))
	}
	if profile.Bathrooms > 0 {
		parts = append(parts, fmt.Sprintf("%sba", formatFloat(profile.Bathrooms)))
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, " / ")
}

func sqftLabel(sqft int) string {
	if sqft == 0 {
		return ""
	}
	return fmt.Sprintf("%d ft\u00B2", sqft)
}

func lotLabel(sqft int) string {
	if sqft == 0 {
		return ""
	}
	return fmt.Sprintf("%d ft\u00B2 lot", sqft)
}

func formatInt(value int) string {
	if value == 0 {
		return ""
	}
	return fmt.Sprintf("%d", value)
}

func formatFloat(value float64) string {
	if value == 0 {
		return ""
	}
	if value == math.Trunc(value) {
		return fmt.Sprintf("%.0f", value)
	}
	return fmt.Sprintf("%.1f", value)
}

func formatCityState(profile data.HouseProfile) string {
	parts := []string{
		strings.TrimSpace(profile.City),
		strings.TrimSpace(profile.State),
	}
	return joinNonEmpty(parts, ", ")
}

func formatAddress(profile data.HouseProfile) string {
	parts := []string{
		strings.TrimSpace(profile.AddressLine1),
		strings.TrimSpace(profile.AddressLine2),
		strings.TrimSpace(profile.City),
		strings.TrimSpace(profile.State),
		strings.TrimSpace(profile.PostalCode),
	}
	return joinNonEmpty(parts, ", ")
}

// filterNonBlank returns only the values that have visible content.
func filterNonBlank(values []string) []string {
	filtered := make([]string, 0, len(values))
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			filtered = append(filtered, v)
		}
	}
	return filtered
}

func joinInline(values ...string) string {
	if f := filterNonBlank(values); len(f) > 0 {
		return lipgloss.JoinHorizontal(lipgloss.Center, f...)
	}
	return ""
}

func joinVerticalNonEmpty(values ...string) string {
	if f := filterNonBlank(values); len(f) > 0 {
		return lipgloss.JoinVertical(lipgloss.Left, f...)
	}
	return ""
}

func joinWithSeparator(sep string, values ...string) string {
	return strings.Join(filterNonBlank(values), sep)
}

func joinNonEmpty(values []string, sep string) string {
	return joinWithSeparator(sep, values...)
}

func hoaSummary(profile data.HouseProfile) string {
	if profile.HOAName == "" && profile.HOAFeeCents == nil {
		return ""
	}
	fee := data.FormatOptionalCents(profile.HOAFeeCents)
	if profile.HOAName == "" {
		return fee
	}
	if fee == "" {
		return profile.HOAName
	}
	return fmt.Sprintf("%s (%s)", profile.HOAName, fee)
}
