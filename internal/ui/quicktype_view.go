package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mph-llm-experiments/apeople/internal/model"
)

// updateQuickType handles input in quick type change view
func (m Model) updateQuickType(msg tea.KeyMsg) (Model, tea.Cmd) {
	if m.contactToMark == nil {
		m.currentView = ViewList
		return m, nil
	}

	switch msg.String() {
	case "esc", "ctrl+c":
		// Cancel and return to entry view
		m.currentView = m.entryView
		m.contactToMark = nil
		return m, nil

	// Type selections
	case "f":
		return m.saveQuickType("family")
	case "c":
		return m.saveQuickType("close")
	case "n":
		return m.saveQuickType("network")
	case "w":
		return m.saveQuickType("work")
	case "r":
		return m.saveQuickType("recruiters")
	case "p":
		return m.saveQuickType("providers")
	case "s":
		return m.saveQuickType("social")
	}

	return m, nil
}

// saveQuickType saves the type change and returns to list
func (m Model) saveQuickType(newType string) (Model, tea.Cmd) {
	if m.contactToMark == nil {
		return m, nil
	}

	// Update the contact's type
	contact := *m.contactToMark
	contact.RelationshipType = model.RelationshipType(newType)

	// Save command
	return m, m.saveQuickTypeChange(contact)
}

// viewQuickType renders the quick type selection interface
func (m Model) viewQuickType() string {
	if m.contactToMark == nil {
		return "No contact selected"
	}

	var b strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("214"))
	b.WriteString(titleStyle.Render("Change Type"))
	b.WriteString("\n\n")

	// Contact name
	b.WriteString(fmt.Sprintf("Contact: %s\n", 
		lipgloss.NewStyle().Bold(true).Render(m.contactToMark.Title)))
	b.WriteString(fmt.Sprintf("Current type: %s\n\n", 
		lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(string(m.contactToMark.RelationshipType))))

	// Options
	b.WriteString("Select new type:\n\n")
	
	hotkeyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	
	options := []struct {
		key   string
		value string
	}{
		{"f", "family"},
		{"c", "close"},
		{"n", "network"},
		{"w", "work"},
		{"r", "recruiters"},
		{"p", "providers"},
		{"s", "social"},
	}

	for _, opt := range options {
		b.WriteString(fmt.Sprintf("  %s  %s\n", 
			hotkeyStyle.Render("("+opt.key+")"),
			labelStyle.Render(opt.value)))
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render("Esc to cancel"))

	return b.String()
}