package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mph-llm-experiments/apeople/internal/model"
)

// Edit form fields
const (
	fieldTitle = iota
	fieldEmail
	fieldPhone
	fieldCompany
	fieldRole
	fieldLocation
	fieldRelationType
	fieldContactStyle
	fieldState
	fieldTags
	fieldCount
)

var fieldLabels = []string{
	"Name",
	"Email",
	"Phone",
	"Company",
	"Role",
	"Location",
	"Type",
	"Style",
	"State",
	"Tags",
}

// updateEdit handles input in edit view
func (m Model) updateEdit(msg tea.KeyMsg) (Model, tea.Cmd) {
	if m.editingContact == nil {
		return m, nil
	}
	

	// If no field is being edited (field selection mode)
	if m.editField == -1 {
		switch msg.String() {
		case "esc":
			// Cancel without saving - return to entry view
			m.currentView = m.entryView
			m.editingContact = nil
			m.editField = -1
			m.editValues = nil
			return m, nil

		case "q":
			// Save changes and exit
			if m.editingContact != nil {
				return m, m.saveEditedContact()
			}

		// Field selection hotkeys
		case "n":
			m.editField = fieldTitle
		case "e":
			m.editField = fieldEmail
		case "p":
			m.editField = fieldPhone
		case "c":
			m.editField = fieldCompany
		case "r":
			m.editField = fieldRole
		case "l":
			m.editField = fieldLocation
		case "t":
			m.editField = fieldRelationType
		case "s":
			m.editField = fieldContactStyle
		case "S":
			m.editField = fieldState
		case "T":
			m.editField = fieldTags
		}
	} else {
		// Field editing mode
		switch msg.String() {
		case "esc":
			// Cancel field edit, return to field selection
			m.editField = -1
			return m, nil

		case "enter":
			// Save field and return to field selection
			m.editField = -1
			return m, nil

		case "backspace":
			// Delete character from current field
			if len(m.editValues[m.editField]) > 0 {
				m.editValues[m.editField] = m.editValues[m.editField][:len(m.editValues[m.editField])-1]
			}

		default:
			// Handle special fields
			if m.editField == fieldRelationType {
				return m.handleTypeSelection(msg), nil
			} else if m.editField == fieldContactStyle {
				return m.handleStyleSelection(msg), nil
			} else if m.editField == fieldState {
				return m.handleStateSelection(msg), nil
			}
			
			// Add character to current field
			if len(msg.String()) == 1 {
				m.editValues[m.editField] += msg.String()
			}
		}
	}

	return m, nil
}

// handleTypeSelection handles relationship type selection
func (m Model) handleTypeSelection(msg tea.KeyMsg) Model {
	switch msg.String() {
	case "f":
		m.editValues[fieldRelationType] = "family"
		m.editField = -1 // Return to field selection
	case "c":
		m.editValues[fieldRelationType] = "close"
		m.editField = -1 // Return to field selection
	case "n":
		m.editValues[fieldRelationType] = "network"
		m.editField = -1 // Return to field selection
	case "r":
		m.editValues[fieldRelationType] = "recruiters"
		m.editField = -1 // Return to field selection
	case "w":
		m.editValues[fieldRelationType] = "work"
		m.editField = -1 // Return to field selection
	case "p":
		m.editValues[fieldRelationType] = "providers"
		m.editField = -1 // Return to field selection
	case "s":
		m.editValues[fieldRelationType] = "social"
		m.editField = -1 // Return to field selection
	}
	
	return m
}

// handleStyleSelection handles contact style selection
func (m Model) handleStyleSelection(msg tea.KeyMsg) Model {
	switch msg.String() {
	case "p":
		m.editValues[fieldContactStyle] = "periodic"
		m.editField = -1 // Return to field selection
	case "a":
		m.editValues[fieldContactStyle] = "ambient"
		m.editField = -1 // Return to field selection
	case "t":
		m.editValues[fieldContactStyle] = "triggered"
		m.editField = -1 // Return to field selection
	}
	return m
}

// handleStateSelection handles contact state selection
func (m Model) handleStateSelection(msg tea.KeyMsg) Model {
	oldState := m.editValues[fieldState]
	
	switch msg.String() {
	case "o":
		m.editValues[fieldState] = "ok"
		m.editField = -1 // Return to field selection
	case "f":
		m.editValues[fieldState] = "followup"
		m.editField = -1 // Return to field selection
	case "p":
		m.editValues[fieldState] = "ping"
		m.editField = -1 // Return to field selection
	case "s":
		m.editValues[fieldState] = "scheduled"
		m.editField = -1 // Return to field selection
	case "t":
		m.editValues[fieldState] = "timeout"
		m.editField = -1 // Return to field selection
	}
	
	// Show message about task creation
	newState := m.editValues[fieldState]
	if oldState != newState && newState != "" && newState != "ok" {
		actionStates := map[string]bool{
			"followup": true, "ping": true, "scheduled": true,
			"timeout": true,
		}
		if actionStates[newState] {
			m.message = fmt.Sprintf("Task will be created when saved (state → %s)", newState)
		}
	}
	
	return m
}


// viewEdit renders the edit view
func (m Model) viewEdit() string {
	if m.editingContact == nil {
		return "No contact selected for editing"
	}

	var b strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("214"))
	b.WriteString(titleStyle.Render("Edit Contact"))
	b.WriteString("\n\n")

	// Contact name
	b.WriteString(fmt.Sprintf("Editing: %s\n\n", 
		lipgloss.NewStyle().Bold(true).Render(m.editingContact.Title)))
	
	// Show message if present
	if m.message != "" {
		messageStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("82")).
			Bold(true)
		b.WriteString(messageStyle.Render("→ " + m.message))
		b.WriteString("\n\n")
	}

	if m.editField == -1 {
		// Field selection mode - show all fields with hotkeys
		b.WriteString(m.renderFieldSelectionView())
	} else {
		// Field editing mode - show the active field
		b.WriteString(m.renderFieldEditingView())
	}

	// Pad to fill screen
	content := b.String()
	lines := strings.Split(content, "\n")
	for len(lines) < m.height-1 {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}


// renderTypeOptions shows relationship type options when editing that field
func (m Model) renderTypeOptions(current string) string {
	options := []string{
		"(f)amily",
		"(c)lose", 
		"(n)etwork",
		"(w)ork",
		"(r)ecruiters",
		"(p)roviders",
		"(s)ocial",
	}
	
	return strings.Join(options, " ")
}

// renderStyleOptions shows contact style options when editing that field
func (m Model) renderStyleOptions(current string) string {
	options := []string{
		"(p)eriodic",
		"(a)mbient",
		"(t)riggered",
	}
	
	return strings.Join(options, " ")
}

// renderStateOptions shows contact state options when editing that field
func (m Model) renderStateOptions(current string) string {
	options := []string{
		"(o)k",
		"(f)ollowup",
		"(p)ing",
		"(s)cheduled",
		"(t)imeout",
	}
	
	return strings.Join(options, " ")
}

// renderFieldSelectionView shows all fields with hotkeys for selection
func (m Model) renderFieldSelectionView() string {
	var b strings.Builder
	
	// Show all fields with their current values and hotkeys
	fieldHotkeys := []string{"n", "e", "p", "c", "r", "l", "t", "s", "S", "T"}
	
	for i := 0; i < fieldCount; i++ {
		label := fieldLabels[i]
		value := m.editValues[i]
		hotkey := fieldHotkeys[i]
		
		if value == "" {
			value = lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Italic(true).Render("(empty)")
		}
		
		b.WriteString(fmt.Sprintf("  (%s) %s: %s\n", 
			lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render(hotkey),
			lipgloss.NewStyle().Foreground(lipgloss.Color("250")).Render(fmt.Sprintf("%-15s", label)),
			lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render(value)))
	}
	
	b.WriteString("\n")
	
	// Instructions
	instructionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
	b.WriteString(instructionStyle.Render("Select field to edit • q: save & exit • Esc: cancel"))
	
	return b.String()
}

// renderFieldEditingView shows the active field being edited
func (m Model) renderFieldEditingView() string {
	var b strings.Builder
	
	label := fieldLabels[m.editField]
	value := m.editValues[m.editField]
	
	// Special rendering for certain fields
	if m.editField == fieldRelationType {
		b.WriteString(fmt.Sprintf("Editing %s:\n\n", label))
		b.WriteString(m.renderTypeOptions(value))
		b.WriteString("\n\n")
	} else if m.editField == fieldContactStyle {
		b.WriteString(fmt.Sprintf("Editing %s:\n\n", label))
		b.WriteString(m.renderStyleOptions(value))
		b.WriteString("\n\n")
	} else if m.editField == fieldState {
		b.WriteString(fmt.Sprintf("Editing %s:\n\n", label))
		b.WriteString(m.renderStateOptions(value))
		b.WriteString("\n\n")
	} else {
		b.WriteString(fmt.Sprintf("Editing %s:\n\n", label))
		// Show current value with cursor
		displayValue := value + "█"
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render(displayValue))
		b.WriteString("\n\n")
	}
	
	// Instructions
	instructionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
	b.WriteString(instructionStyle.Render("Type to edit • Enter: save field • Esc: cancel field"))
	
	return b.String()
}

// initializeEditValues populates edit form with current contact values
func (m *Model) initializeEditValues(contact model.Contact) {
	m.editValues = make([]string, fieldCount)
	
	m.editValues[fieldTitle] = contact.Title
	m.editValues[fieldEmail] = contact.Email
	m.editValues[fieldPhone] = contact.Phone
	m.editValues[fieldCompany] = contact.Company
	m.editValues[fieldRole] = contact.Role
	m.editValues[fieldLocation] = contact.Location
	m.editValues[fieldRelationType] = string(contact.RelationshipType)
	m.editValues[fieldContactStyle] = string(contact.ContactStyle)
	m.editValues[fieldState] = contact.State
	
	// Join tags (excluding "contact")
	var tags []string
	for _, tag := range contact.Tags {
		if tag != "contact" {
			tags = append(tags, tag)
		}
	}
	m.editValues[fieldTags] = strings.Join(tags, " ")
}