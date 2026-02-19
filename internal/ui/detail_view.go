package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mph-llm-experiments/apeople/internal/model"
)

// Detail view styles
var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Bold(true)
	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("250"))
	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))
	sectionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			MarginTop(1)
	emptyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Italic(true)
)

// updateDetail handles input in detail view
func (m Model) updateDetail(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.currentView = ViewList
		m.selectedContact = nil
		
	case "d":
		// Show interaction type selector
		if m.selectedContact != nil {
			m.contactToMark = m.selectedContact
			m.entryView = m.currentView  // Capture where we came from
			m.currentView = ViewInteractionType
			m.contactLogStep = 0
			m.interactionType = ""
			m.interactionState = ""
			m.interactionNote = ""
		}
		
	case "b":
		// Bump contact
		if m.selectedContact != nil {
			return m, m.bumpContact(*m.selectedContact)
		}
		
	case "e":
		// Enter edit mode
		if m.selectedContact != nil {
			m.editingContact = m.selectedContact
			m.initializeEditValues(*m.selectedContact)
			m.entryView = m.currentView  // Capture where we came from
			m.currentView = ViewEdit
			m.editField = -1 // Start in field selection mode
		}
		
	case "x":
		// TODO: Delete contact
	}
	return m, nil
}

// viewDetail renders the detail view
func (m Model) viewDetail() string {
	if m.selectedContact == nil {
		return "No contact selected"
	}
	
	contact := *m.selectedContact
	var b strings.Builder
	
	// Header
	b.WriteString(m.renderDetailHeader(contact))
	b.WriteString("\n\n")
	
	// Show message if present
	if m.message != "" {
		messageStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("82")).
			Bold(true)
		b.WriteString(messageStyle.Render("→ " + m.message))
		b.WriteString("\n\n")
	}
	
	// Basic Information
	b.WriteString(sectionStyle.Render("Contact Information"))
	b.WriteString("\n")
	b.WriteString(m.renderContactInfo(contact))
	b.WriteString("\n")
	
	// Relationship & Frequency
	b.WriteString(sectionStyle.Render("Relationship"))
	b.WriteString("\n")
	b.WriteString(m.renderRelationshipInfo(contact))
	b.WriteString("\n")
	
	// Contact History
	b.WriteString(sectionStyle.Render("Contact History"))
	b.WriteString("\n")
	b.WriteString(m.renderContactHistory(contact))
	b.WriteString("\n")
	
	// Notes/Content
	if contact.Content != "" {
		b.WriteString(sectionStyle.Render("Recent Interactions"))
		b.WriteString("\n")
		b.WriteString(m.renderContactContent(contact))
		b.WriteString("\n")
	}
	
	// Footer
	b.WriteString(m.renderDetailFooter())
	
	return b.String()
}

// renderDetailHeader renders the contact name and status
func (m Model) renderDetailHeader(contact model.Contact) string {
	title := titleStyle.Render(contact.Title)
	
	// Status indicators
	var status []string
	
	if contact.IsOverdue() {
		status = append(status, overdueColor.Render("● Overdue"))
	} else if contact.NeedsAttention() {
		status = append(status, attentionColor.Render("! Needs Attention"))
	} else if contact.IsWithinThreshold() {
		status = append(status, goodColor.Render("● Good"))
	} else {
		status = append(status, baseColor.Render("○ OK"))
	}
	
	// Days since contact
	days := contact.DaysSinceContact()
	if days >= 0 {
		status = append(status, fmt.Sprintf("%d days since contact", days))
	} else {
		status = append(status, "Never contacted")
	}
	
	statusLine := strings.Join(status, " • ")
	
	return title + "\n" + statusLine
}

// renderContactInfo renders basic contact information
func (m Model) renderContactInfo(contact model.Contact) string {
	var lines []string
	
	// Email
	if contact.Email != "" {
		lines = append(lines, m.renderField("Email", contact.Email))
	}
	
	// Phone
	if contact.Phone != "" {
		lines = append(lines, m.renderField("Phone", contact.Phone))
	}
	
	// Company
	if contact.Company != "" {
		lines = append(lines, m.renderField("Company", contact.Company))
	}
	
	// Role
	if contact.Role != "" {
		lines = append(lines, m.renderField("Role", contact.Role))
	}
	
	// Location
	if contact.Location != "" {
		lines = append(lines, m.renderField("Location", contact.Location))
	}
	
	// LinkedIn
	if contact.LinkedIn != "" {
		lines = append(lines, m.renderField("LinkedIn", contact.LinkedIn))
	}
	
	// Website
	if contact.Website != "" {
		lines = append(lines, m.renderField("Website", contact.Website))
	}
	
	// Tags
	if len(contact.Tags) > 1 { // More than just "contact"
		var displayTags []string
		for _, tag := range contact.Tags {
			if tag != "contact" {
				displayTags = append(displayTags, tag)
			}
		}
		if len(displayTags) > 0 {
			lines = append(lines, m.renderField("Tags", strings.Join(displayTags, ", ")))
		}
	}
	
	if len(lines) == 0 {
		return emptyStyle.Render("No contact information")
	}
	
	return strings.Join(lines, "\n")
}

// renderRelationshipInfo renders relationship and frequency information
func (m Model) renderRelationshipInfo(contact model.Contact) string {
	var lines []string
	
	// Type
	lines = append(lines, m.renderField("Type", string(contact.RelationshipType)))
	
	// Style
	styleStr := string(contact.ContactStyle)
	if styleStr == "" {
		styleStr = "periodic"
	}
	lines = append(lines, m.renderField("Style", styleStr))
	
	// Frequency
	freq := contact.GetFrequencyDays()
	freqStr := "No default"
	if freq > 0 {
		freqStr = fmt.Sprintf("Every %d days", freq)
	}
	if contact.CustomFrequencyDays > 0 {
		freqStr += " (custom)"
	}
	lines = append(lines, m.renderField("Frequency", freqStr))
	
	// State
	if contact.State != "" && contact.State != "ok" {
		lines = append(lines, m.renderField("State", contact.State))
	}
	
	// Label
	if contact.Label != "" {
		lines = append(lines, m.renderField("Label", contact.Label))
	}
	
	return strings.Join(lines, "\n")
}

// renderContactHistory renders last contact and bump information
func (m Model) renderContactHistory(contact model.Contact) string {
	var lines []string
	
	// Last contacted
	if contact.LastContacted != nil {
		lastStr := contact.LastContacted.Format("January 2, 2006")
		days := contact.DaysSinceContact()
		if days == 0 {
			lastStr += " (today)"
		} else if days == 1 {
			lastStr += " (yesterday)"
		} else if days > 0 {
			lastStr += fmt.Sprintf(" (%d days ago)", days)
		}
		lines = append(lines, m.renderField("Last Contacted", lastStr))
		
		// Last interaction type
		if contact.LastInteractionType != "" {
			lines = append(lines, m.renderField("Via", contact.LastInteractionType))
		}
	} else {
		lines = append(lines, m.renderField("Last Contacted", "Never"))
	}
	
	// Bump information
	if contact.LastBumpDate != nil {
		bumpStr := contact.LastBumpDate.Format("January 2, 2006")
		if contact.BumpCount > 0 {
			bumpStr += fmt.Sprintf(" (%d bumps)", contact.BumpCount)
		}
		lines = append(lines, m.renderField("Last Reviewed", bumpStr))
	}
	
	// Created
	lines = append(lines, m.renderField("Created", contact.Date.Format("January 2, 2006")))
	
	// Updated
	if !contact.UpdatedAt.IsZero() {
		lines = append(lines, m.renderField("Updated", contact.UpdatedAt.Format("January 2, 2006")))
	}
	
	return strings.Join(lines, "\n")
}

// renderContactContent renders the markdown content
func (m Model) renderContactContent(contact model.Contact) string {
	content := strings.TrimSpace(contact.Content)
	if content == "" {
		return emptyStyle.Render("No interaction history")
	}
	
	// Simple rendering - just indent the content
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) != "" {
			lines[i] = "  " + line
		}
	}
	
	return valueStyle.Render(strings.Join(lines, "\n"))
}

// renderField renders a label-value pair
func (m Model) renderField(label, value string) string {
	return fmt.Sprintf("  %s: %s", 
		labelStyle.Render(fmt.Sprintf("%-15s", label)),
		valueStyle.Render(value))
}

// renderDetailFooter renders the footer with available actions
func (m Model) renderDetailFooter() string {
	keys := []string{
		"d:mark contacted",
		"b:bump",
		"e:edit",
		"x:delete",
		"esc:back",
	}
	
	return "\n" + headerColor.Render(strings.Join(keys, " • "))
}