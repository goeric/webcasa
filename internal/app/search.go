package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/micasa/micasa/internal/data"
)

const maxSearchResults = 50

func newSearchState() searchState {
	input := textinput.New()
	input.Prompt = ""
	input.Placeholder = "search..."
	input.CharLimit = 128
	input.Width = 40
	spin := spinner.New()
	spin.Spinner = spinner.Dot
	return searchState{
		input:   input,
		spinner: spin,
	}
}

func (m *Model) openSearch() tea.Cmd {
	m.search.active = true
	m.mode = modeSearch
	m.search.cursor = 0
	m.search.input.SetValue("")
	m.search.results = nil
	m.search.lastQuery = ""
	cmd := m.search.input.Focus()
	if m.search.dirty || len(m.search.entries) == 0 {
		return tea.Batch(cmd, m.startSearchIndexBuild())
	}
	return cmd
}

func (m *Model) updateSearch(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := msg.(type) {
	case tea.KeyMsg:
		switch typed.String() {
		case "esc":
			m.search.active = false
			m.mode = modeTable
			m.search.input.Blur()
			return m, nil
		case "enter":
			if m.search.indexing || len(m.search.results) == 0 {
				return m, nil
			}
			m.selectSearchResult()
			return m, nil
		case "up", "k":
			m.search.cursor = clampIndex(m.search.cursor-1, len(m.search.results))
			return m, nil
		case "down", "j":
			m.search.cursor = clampIndex(m.search.cursor+1, len(m.search.results))
			return m, nil
		}
	case searchIndexMsg:
		m.search.entries = typed.Entries
		m.search.indexing = false
		m.search.dirty = false
		m.updateSearchResults()
		return m, nil
	case searchIndexErrMsg:
		m.search.indexing = false
		m.logError(fmt.Sprintf("search index error: %v", typed.Err))
		return m, nil
	case spinner.TickMsg:
		if m.search.indexing {
			var cmd tea.Cmd
			m.search.spinner, cmd = m.search.spinner.Update(msg)
			return m, cmd
		}
	}

	var cmd tea.Cmd
	m.search.input, cmd = m.search.input.Update(msg)
	m.updateSearchResults()
	return m, cmd
}

func (m *Model) startSearchIndexBuild() tea.Cmd {
	if m.search.indexing {
		return nil
	}
	m.search.indexing = true
	buildCmd := func() tea.Msg {
		entries, err := buildSearchEntries(m.store)
		if err != nil {
			return searchIndexErrMsg{Err: err}
		}
		return searchIndexMsg{Entries: entries}
	}
	return tea.Batch(buildCmd, m.search.spinner.Tick)
}

func (m *Model) updateSearchResults() {
	query := strings.TrimSpace(m.search.input.Value())
	if query == m.search.lastQuery {
		return
	}
	m.search.lastQuery = query
	if query == "" {
		m.search.results = nil
		m.search.cursor = 0
		return
	}
	tokens := strings.Fields(strings.ToLower(query))
	results := make([]searchEntry, 0, maxSearchResults)
	for _, entry := range m.search.entries {
		if !matchesTokens(entry.Searchable, tokens) {
			continue
		}
		results = append(results, entry)
		if len(results) >= maxSearchResults {
			break
		}
	}
	m.search.results = results
	m.search.cursor = 0
}

func matchesTokens(source string, tokens []string) bool {
	for _, token := range tokens {
		if !strings.Contains(source, token) {
			return false
		}
	}
	return true
}

func buildSearchEntries(store *data.Store) ([]searchEntry, error) {
	projects, err := store.ListProjects(false)
	if err != nil {
		return nil, err
	}
	quotes, err := store.ListQuotes(false)
	if err != nil {
		return nil, err
	}
	maintenance, err := store.ListMaintenance(false)
	if err != nil {
		return nil, err
	}
	entries := make([]searchEntry, 0, len(projects)+len(quotes)+len(maintenance))
	for _, project := range projects {
		title := project.Title
		if title == "" {
			title = fmt.Sprintf("Project %d", project.ID)
		}
		summary := fmt.Sprintf("%s • %s", project.ProjectType.Name, project.Status)
		entries = append(entries, searchEntry{
			Tab:        tabProjects,
			ID:         project.ID,
			Title:      title,
			Summary:    summary,
			Searchable: strings.ToLower(title + " " + summary),
		})
	}
	for _, quote := range quotes {
		title := quote.Vendor.Name
		if title == "" {
			title = fmt.Sprintf("Quote %d", quote.ID)
		}
		projectTitle := quote.Project.Title
		if projectTitle == "" {
			projectTitle = fmt.Sprintf("Project %d", quote.ProjectID)
		}
		summary := fmt.Sprintf("%s • %s", projectTitle, data.FormatCents(quote.TotalCents))
		entries = append(entries, searchEntry{
			Tab:        tabQuotes,
			ID:         quote.ID,
			Title:      title,
			Summary:    summary,
			Searchable: strings.ToLower(title + " " + summary),
		})
	}
	for _, item := range maintenance {
		title := item.Name
		if title == "" {
			title = fmt.Sprintf("Maintenance %d", item.ID)
		}
		summary := fmt.Sprintf("%s • %s", item.Category.Name, dateValue(item.NextDueAt))
		entries = append(entries, searchEntry{
			Tab:        tabMaintenance,
			ID:         item.ID,
			Title:      title,
			Summary:    summary,
			Searchable: strings.ToLower(title + " " + summary),
		})
	}
	return entries, nil
}

func (m *Model) selectSearchResult() {
	if len(m.search.results) == 0 {
		return
	}
	result := m.search.results[m.search.cursor]
	m.search.active = false
	m.mode = modeTable
	m.search.input.Blur()
	m.active = tabIndex(result.Tab)
	_ = m.reloadActiveTab()
	tab := m.activeTab()
	if tab == nil {
		return
	}
	if selectRowByID(tab, result.ID) {
		m.setStatusInfo("Jumped to result.")
		return
	}
	m.setStatusError("Result not found in current list.")
}

func selectRowByID(tab *Tab, id uint) bool {
	for idx, meta := range tab.Rows {
		if meta.ID == id {
			tab.Table.SetCursor(idx)
			return true
		}
	}
	return false
}

func tabIndex(kind TabKind) int {
	switch kind {
	case tabProjects:
		return 0
	case tabQuotes:
		return 1
	case tabMaintenance:
		return 2
	default:
		return 0
	}
}

func clampIndex(value, length int) int {
	if length <= 0 {
		return 0
	}
	if value < 0 {
		return 0
	}
	if value >= length {
		return length - 1
	}
	return value
}
