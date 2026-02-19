package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mph-llm-experiments/apeople/internal/model"
)

// updateSearch handles input in search mode
func (m Model) updateSearch(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEscape:
		m.searchMode = false
		m.searchQuery = ""
		m.filtered = m.contacts
		m.cursor = 0
		return m, nil
		
	case tea.KeyEnter:
		m.searchMode = false
		// Keep the search results
		return m, nil
		
	case tea.KeyBackspace:
		if len(m.searchQuery) > 0 {
			m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
			m.applyFilters()
			m.cursor = 0
		}
		
	case tea.KeyRunes:
		m.searchQuery += string(msg.Runes)
		m.applyFilters()
		m.cursor = 0
	}
	
	return m, nil
}

// contactMatchesSearch checks if a contact matches the search query
func (m *Model) contactMatchesSearch(contact model.Contact, query string) bool {
	query = strings.ToLower(query)
	
	// Search in name
	if strings.Contains(strings.ToLower(contact.Title), query) {
		return true
	}
	// Search in company
	if strings.Contains(strings.ToLower(contact.Company), query) {
		return true
	}
	// Search in email
	if strings.Contains(strings.ToLower(contact.Email), query) {
		return true
	}
	// Search in tags
	for _, tag := range contact.Tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}
	// Search in label
	if strings.Contains(strings.ToLower(contact.Label), query) {
		return true
	}
	// Search in role
	if strings.Contains(strings.ToLower(contact.Role), query) {
		return true
	}
	
	return false
}

// applyFilters applies search and filter criteria
func (m *Model) applyFilters() {
	m.filtered = []model.Contact{}
	
	for _, contact := range m.contacts {
		// Apply search query
		if m.searchQuery != "" && !m.contactMatchesSearch(contact, m.searchQuery) {
			continue
		}
		
		// Apply type filter
		if m.filterType != "" && string(contact.RelationshipType) != m.filterType {
			continue
		}
		
		// Apply state filter
		if m.filterState != "" && contact.State != m.filterState {
			continue
		}
		
		// Apply status filter
		if m.filterStatus != "" {
			switch m.filterStatus {
			case "overdue":
				if !contact.IsOverdue() {
					continue
				}
			case "needsAttention":
				if !contact.NeedsAttention() {
					continue
				}
			case "ok":
				if !contact.IsWithinThreshold() {
					continue
				}
			}
		}
		
		m.filtered = append(m.filtered, contact)
	}
	
	// Reset cursor if it's out of bounds
	if m.cursor >= len(m.filtered) {
		m.cursor = len(m.filtered) - 1
		if m.cursor < 0 {
			m.cursor = 0
		}
	}
}