package app

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/micasa/micasa/internal/data"
)

func (m *Model) buildView() string {
	house := m.houseView()
	tabs := m.tabsView()
	content := ""
	if m.mode == modeForm && m.form != nil {
		content = m.form.View()
	} else if tab := m.activeTab(); tab != nil {
		content = tab.Table.View()
	}
	status := m.statusView()
	return lipgloss.JoinVertical(lipgloss.Left, house, tabs, content, status)
}

func (m *Model) houseView() string {
	if !m.hasHouse {
		content := lipgloss.JoinVertical(
			lipgloss.Left,
			m.houseTitleLine("setup"),
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
	line1 := m.houseTitleLine("collapsed")
	line2 := joinInline(
		m.chip("🏠", m.house.Nickname),
		m.chip("📍", formatCityState(m.house)),
		m.chip("🗓", formatInt(m.house.YearBuilt)),
		m.chip("📐", formatInt(m.house.SquareFeet)),
		m.chip("🛏", formatInt(m.house.Bedrooms)),
		m.chip("🛁", formatFloat(m.house.Bathrooms)),
	)
	return lipgloss.JoinVertical(lipgloss.Left, line1, line2)
}

func (m *Model) houseExpanded() string {
	address := formatAddress(m.house)
	line1 := m.houseTitleLine("expanded")
	line2 := joinInline(
		m.chip("🏠", m.house.Nickname),
		m.chip("📍", address),
	)
	line3 := m.sectionLine(
		"🧱 Structure",
		m.chip("🗓", formatInt(m.house.YearBuilt)),
		m.chip("📐", formatInt(m.house.SquareFeet)),
		m.chip("🌳", formatInt(m.house.LotSquareFeet)),
		m.chip("🛏", formatInt(m.house.Bedrooms)),
		m.chip("🛁", formatFloat(m.house.Bathrooms)),
		m.chip("🧱", m.house.FoundationType),
		m.chip("🧵", m.house.WiringType),
		m.chip("🏚", m.house.RoofType),
		m.chip("🧱", m.house.ExteriorType),
		m.chip("🕳", m.house.BasementType),
	)
	line4 := m.sectionLine(
		"⚡ Utilities",
		m.chip("🔥", m.house.HeatingType),
		m.chip("❄️", m.house.CoolingType),
		m.chip("💧", m.house.WaterSource),
		m.chip("🚰", m.house.SewerType),
		m.chip("🚗", m.house.ParkingType),
	)
	line5 := m.sectionLine(
		"💰 Financial",
		m.chip("🛡", m.house.InsuranceCarrier),
		m.chip("📄", m.house.InsurancePolicy),
		m.chip("🔁", data.FormatDate(m.house.InsuranceRenewal)),
		m.chip("🧾", data.FormatOptionalCents(m.house.PropertyTaxCents)),
		m.chip("🏘", hoaSummary(m.house)),
	)
	return lipgloss.JoinVertical(
		lipgloss.Left,
		line1,
		line2,
		line3,
		line4,
		line5,
	)
}

func (m *Model) tabsView() string {
	tabs := make([]string, 0, len(m.tabs))
	for i, tab := range m.tabs {
		if i == m.active {
			tabs = append(tabs, m.styles.TabActive.Render(tab.Name))
		} else {
			tabs = append(tabs, m.styles.TabInactive.Render(tab.Name))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, tabs...)
}

func (m *Model) statusView() string {
	if m.mode == modeForm {
		help := joinWithSeparator(
			m.styles.HeaderHint.Render(" | "),
			m.helpItem("esc", "cancel"),
			m.helpItem("ctrl+c", "quit"),
		)
		if m.status.Text == "" {
			return help
		}
		style := m.styles.Info
		if m.status.Kind == statusError {
			style = m.styles.Error
		}
		return lipgloss.JoinVertical(
			lipgloss.Left,
			style.Render(m.status.Text),
			help,
		)
	}
	deleted := "deleted:off"
	tab := m.activeTab()
	if tab != nil && tab.ShowDeleted {
		deleted = "deleted:on"
	}
	help := joinWithSeparator(
		m.styles.HeaderHint.Render(" | "),
		m.helpItem("tab/shift+tab", "switch"),
		m.helpItem("a", "add"),
		m.helpItem("d", "delete"),
		m.helpItem("u", "restore"),
		m.helpItem("x", "deleted"),
		m.helpItem("h", "house"),
		m.helpItem("q", "quit"),
	)
	deletedLabel := m.styles.HeaderHint.Render(deleted)
	helpLine := joinWithSeparator(
		m.styles.HeaderHint.Render(" | "),
		help,
		deletedLabel,
	)
	if m.status.Text == "" {
		return helpLine
	}
	style := m.styles.Info
	if m.status.Kind == statusError {
		style = m.styles.Error
	}
	return lipgloss.JoinVertical(
		lipgloss.Left,
		style.Render(m.status.Text),
		helpLine,
	)
}

func (m *Model) headerBox(content string) string {
	return m.styles.HeaderBox.Render(content)
}

func (m *Model) helpItem(keys, label string) string {
	keycaps := m.renderKeys(keys)
	desc := m.styles.HeaderHint.Render(label)
	return strings.TrimSpace(fmt.Sprintf("%s %s", keycaps, desc))
}

func (m *Model) renderKeys(keys string) string {
	parts := strings.Split(keys, "/")
	rendered := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		rendered = append(rendered, m.keycap(part))
	}
	return joinWithSeparator(m.styles.HeaderHint.Render(" / "), rendered...)
}

func (m *Model) keycap(value string) string {
	return m.styles.Keycap.Render(strings.ToUpper(value))
}

func (m *Model) houseTitleLine(state string) string {
	title := m.styles.HeaderTitle.Render("House Profile")
	badge := ""
	if strings.TrimSpace(state) != "" {
		badge = m.styles.HeaderBadge.Render(state)
	}
	hint := m.styles.HeaderHint.Render("h toggle")
	return joinInline(title, badge, hint)
}

func (m *Model) chip(label, value string) string {
	labelText := m.styles.HeaderLabel.Render(label)
	valueText := m.renderHouseValue(value)
	return m.styles.HeaderChip.Render(fmt.Sprintf("%s %s", labelText, valueText))
}

func (m *Model) sectionLine(label string, chips ...string) string {
	section := m.styles.HeaderSection.Render(label)
	parts := append([]string{section}, chips...)
	return joinInline(parts...)
}

func (m *Model) renderHouseValue(value string) string {
	if strings.TrimSpace(value) == "" {
		return m.styles.Empty.Render("n/a")
	}
	return m.styles.HeaderValue.Render(value)
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

func joinInline(values ...string) string {
	filtered := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			filtered = append(filtered, value)
		}
	}
	if len(filtered) == 0 {
		return ""
	}
	return lipgloss.JoinHorizontal(lipgloss.Center, filtered...)
}

func joinWithSeparator(sep string, values ...string) string {
	filtered := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			filtered = append(filtered, value)
		}
	}
	return strings.Join(filtered, sep)
}

func joinNonEmpty(values []string, sep string) string {
	filtered := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			filtered = append(filtered, value)
		}
	}
	return strings.Join(filtered, sep)
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
