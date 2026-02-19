package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mph-llm-experiments/apeople/internal/model"
)

// Colors from denote-tasks
var (
	baseColor     = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	selectedColor = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	headerColor   = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
	overdueColor  = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	attentionColor = lipgloss.NewStyle().Foreground(lipgloss.Color("226"))
	goodColor     = lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
)

// updateList handles input in list view
func (m Model) updateList(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
		
	case "j", "down":
		if m.cursor < len(m.filtered)-1 {
			m.cursor++
		}
		
	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
		}
		
	case "g", "home":
		m.cursor = 0
		
	case "G", "end":
		m.cursor = len(m.filtered) - 1
		
	case "ctrl+d":
		// Page down
		m.cursor += 10
		if m.cursor >= len(m.filtered) {
			m.cursor = len(m.filtered) - 1
		}
		
	case "ctrl+u":
		// Page up
		m.cursor -= 10
		if m.cursor < 0 {
			m.cursor = 0
		}
		
	case "enter":
		if m.cursor < len(m.filtered) {
			m.selectedContact = &m.filtered[m.cursor]
			m.currentView = ViewDetail
		}
		
	case "/":
		m.searchMode = true
		m.searchQuery = ""
		
	case "f", "F":
		// Show filter popup
		m.showFilterPopup = true
		
	case "d":
		// Show interaction type selector
		if m.cursor < len(m.filtered) {
			m.contactToMark = &m.filtered[m.cursor]
			m.entryView = m.currentView  // Capture where we came from
			m.currentView = ViewInteractionType
			m.contactLogStep = 0
			m.interactionType = ""
			m.interactionState = ""
			m.interactionNote = ""
		}
		
	case "b":
		// Bump contact
		if m.cursor < len(m.filtered) {
			contact := m.filtered[m.cursor]
			return m, m.bumpContact(contact)
		}
		
	case "e":
		// Edit contact from list view
		if m.cursor < len(m.filtered) {
			contact := m.filtered[m.cursor]
			m.editingContact = &contact
			m.initializeEditValues(contact)
			m.entryView = m.currentView  // Capture where we came from
			m.currentView = ViewEdit
			m.editField = -1 // Start in field selection mode
		}
		
	case "c":
		// Create new contact
		m.initializeCreateValues()
		m.entryView = m.currentView  // Capture where we came from
		m.currentView = ViewCreate
		m.editField = -1 // Start in field selection mode
		
	case "s":
		// Quick state change
		if m.cursor < len(m.filtered) {
			m.contactToMark = &m.filtered[m.cursor]
			m.entryView = m.currentView  // Capture where we came from
			m.currentView = ViewInteractionType
			m.contactLogStep = 1 // Skip interaction type, go straight to state selection
			m.interactionType = "note" // Default to note for quick state changes
			m.interactionState = ""
			m.interactionNote = ""
		}
		
	case "T":
		// Quick type change
		if m.cursor < len(m.filtered) {
			m.contactToMark = &m.filtered[m.cursor]
			m.entryView = m.currentView  // Capture where we came from
			m.currentView = ViewQuickType
		}
	}
	
	return m, nil
}

// viewList renders the list view
func (m Model) viewList() string {
	var b strings.Builder
	
	// Header
	b.WriteString(m.renderHeader())
	b.WriteString("\n")
	
	// Show message if present
	if m.message != "" {
		messageStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("82")).
			Bold(true)
		b.WriteString(messageStyle.Render("→ " + m.message))
		b.WriteString("\n")
	}
	
	// Column headers - matching the actual column spacing
	if len(m.filtered) > 0 {
		columnHeaders := fmt.Sprintf("      %-30s  %4s  %-10s  %-8s  %-35s  %s",
			"NAME",
			"DAYS",
			"TYPE",
			"STATE",
			"COMPANY/ROLE",
			"TAGS",
		)
		b.WriteString(headerColor.Render(columnHeaders))
		b.WriteString("\n")
		b.WriteString(headerColor.Render(strings.Repeat("─", m.width)))
		b.WriteString("\n")
	}
	
	// List
	// Account for header (1), column headers (3), and footer (2 lines)
	footerLines := 2
	if m.searchMode {
		footerLines = 2 // Search line + empty line
	}
	extraLines := 0
	if m.message != "" {
		extraLines = 1 // Account for message line
	}
	listHeight := m.height - 3 - 1 - footerLines - extraLines // header lines - header - footer - message
	if listHeight < 1 {
		listHeight = 1
	}
	startIdx := 0
	
	// Ensure cursor is visible
	if m.cursor >= startIdx+listHeight {
		startIdx = m.cursor - listHeight + 1
	}
	if m.cursor < startIdx {
		startIdx = m.cursor
	}
	
	endIdx := startIdx + listHeight
	if endIdx > len(m.filtered) {
		endIdx = len(m.filtered)
	}
	
	for i := startIdx; i < endIdx; i++ {
		contact := m.filtered[i]
		line := m.renderContactLine(contact, i == m.cursor)
		b.WriteString(line)
		b.WriteString("\n")
	}
	
	// Fill empty space
	for i := endIdx - startIdx; i < listHeight; i++ {
		b.WriteString("\n")
	}
	
	// Footer
	b.WriteString(m.renderFooter())
	
	return b.String()
}

// renderHeader renders the header
func (m Model) renderHeader() string {
	title := "Denote Contacts"
	status := ""
	
	if len(m.contacts) == 0 {
		status = "Loading contacts..."
	} else {
		// Show position in list
		position := ""
		if len(m.filtered) > 0 {
			position = fmt.Sprintf("[%d/%d]", m.cursor+1, len(m.filtered))
		}
		
		// Build status based on filter state
		if m.searchQuery != "" {
			status = fmt.Sprintf("%s %d of %d (search: %s)", position, len(m.filtered), len(m.contacts), m.searchQuery)
		} else if m.filterType != "" {
			status = fmt.Sprintf("%s %d of %d (type: %s)", position, len(m.filtered), len(m.contacts), m.filterType)
		} else if m.filterState != "" {
			status = fmt.Sprintf("%s %d of %d (state: %s)", position, len(m.filtered), len(m.contacts), m.filterState)
		} else if m.filterStatus != "" {
			statusLabel := m.filterStatus
			switch m.filterStatus {
			case "overdue":
				statusLabel = "overdue"
			case "needsAttention":
				statusLabel = "due soon"
			case "ok":
				statusLabel = "good timing"
			}
			status = fmt.Sprintf("%s %d of %d (status: %s)", position, len(m.filtered), len(m.contacts), statusLabel)
		} else {
			status = fmt.Sprintf("%s %d contacts", position, len(m.filtered))
		}
	}
	
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214"))
	statusStyle := headerColor
	
	// Calculate padding
	totalWidth := m.width
	titleLen := len(title)
	statusLen := len(status)
	padding := totalWidth - titleLen - statusLen - 2
	if padding < 0 {
		padding = 0
	}
	
	return titleStyle.Render(title) + strings.Repeat(" ", padding) + statusStyle.Render(status)
}

// renderContactLine renders a single contact line
func (m Model) renderContactLine(contact model.Contact, selected bool) string {
	// Cursor indicator
	cursor := "  "
	if selected {
		cursor = "> "
	}
	
	// Status indicator (overdue/attention/good/ok)
	var status string
	var statusStyle lipgloss.Style
	if contact.IsOverdue() {
		status = "●"
		statusStyle = overdueColor
	} else if contact.NeedsAttention() {
		status = "!"
		statusStyle = attentionColor
	} else if contact.IsWithinThreshold() {
		status = "●"
		statusStyle = goodColor
	} else {
		status = "○"
		statusStyle = baseColor
	}
	
	// Contact style icon
	styleIcon := " "
	switch contact.ContactStyle {
	case model.StylePeriodic:
		styleIcon = "↻" // periodic/recurring
	case model.StyleAmbient:
		styleIcon = "◦" // ambient/passive
	case model.StyleTriggered:
		styleIcon = "!" // triggered/event-based
	}
	
	
	// Name (fixed width) - FIRST main column
	name := contact.Title
	if len(name) > 30 {
		name = name[:27] + "..."
	}
	name = fmt.Sprintf("%-30s", name)
	
	// Days since contact
	days := contact.DaysSinceContact()
	daysStr := "   -"
	if days >= 0 {
		daysStr = fmt.Sprintf("%4d", days)
	}
	
	// Relationship type - fixed width (10 chars for "recruiters")
	relType := "          "
	if contact.RelationshipType != "" {
		relType = fmt.Sprintf("%-10s", contact.RelationshipType)
		if len(relType) > 10 {
			relType = relType[:10]
		}
	}
	
	// State (active/followup/ping/archived) - only show if not empty or "ok"
	state := "        " // 8 chars for "followup"
	if contact.State != "" && contact.State != "ok" {
		state = fmt.Sprintf("%-8s", contact.State)
	}
	
	// Company/Role
	companyRole := ""
	if contact.Company != "" {
		companyRole = contact.Company
		if contact.Role != "" {
			companyRole += " - " + contact.Role
		}
	} else if contact.Role != "" {
		companyRole = contact.Role
	}
	if len(companyRole) > 35 {
		companyRole = companyRole[:32] + "..."
	}
	companyRole = fmt.Sprintf("%-35s", companyRole)
	
	// Tags (remaining space)
	var displayTags []string
	for _, tag := range contact.Tags {
		if tag != "contact" {
			displayTags = append(displayTags, tag)
		}
	}
	tagStr := ""
	if len(displayTags) > 0 {
		tagStr = "#" + strings.Join(displayTags, " #")
		// Truncate if too long
		if len(tagStr) > 30 {
			tagStr = tagStr[:27] + "..."
		}
	}
	
	// Build columnar line matching header order with proper spacing
	line := fmt.Sprintf("%s%s %s  %s  %s  %s  %s  %s  %s",
		cursor,
		statusStyle.Render(status),
		styleIcon,
		name,
		daysStr,
		relType,
		state,
		companyRole,
		tagStr,
	)
	
	// Apply selection highlighting
	if selected {
		return selectedColor.Render(line)
	}
	return baseColor.Render(line)
}

// renderFooter renders the footer with hotkeys
func (m Model) renderFooter() string {
	// If in search mode, show search input at the bottom
	if m.searchMode {
		searchStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
		prompt := searchStyle.Render("Search: ")
		query := m.searchQuery
		cursor := searchStyle.Render("█")
		
		searchLine := prompt + query + cursor + " " + headerColor.Render("(fuzzy match, #tag for tags, Esc to clear)")
		
		// Pad to full width
		padding := m.width - lipgloss.Width(searchLine)
		if padding > 0 {
			searchLine += strings.Repeat(" ", padding)
		}
		
		return searchLine + "\n" + strings.Repeat(" ", m.width)
	}
	
	// Normal footer
	keys := []string{
		"j/k:navigate",
		"enter:view",
		"d:contacted",
		"s:state",
		"T:type",
		"b:bump",
		"e:edit",
		"c:create",
		"/:search",
		"f:filter",
		"q:quit",
	}
	
	// Add style legend on second line
	// Build colored legend
	var legendParts []string
	legendParts = append(legendParts, headerColor.Render("↻:periodic"))
	legendParts = append(legendParts, headerColor.Render("◦:ambient"))
	legendParts = append(legendParts, headerColor.Render("!:triggered"))
	legendParts = append(legendParts, overdueColor.Render("●:overdue"))
	legendParts = append(legendParts, attentionColor.Render("!:soon"))
	legendParts = append(legendParts, goodColor.Render("●:good"))
	legendParts = append(legendParts, headerColor.Render("○:ok"))
	
	return headerColor.Render(strings.Join(keys, " • ")) + "\n" +
		   strings.Join(legendParts, headerColor.Render(" • "))
}