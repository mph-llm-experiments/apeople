package ui

import (
	"time"
	
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mph-llm-experiments/apeople/internal/model"
)

// ViewMode represents the current view
type ViewMode int

const (
	ViewList ViewMode = iota
	ViewDetail
	ViewEdit
	ViewCreate
	ViewFilter
	ViewInteractionType
	ViewQuickType
)

// Model represents the application state
type Model struct {
	// Core state
	contacts     []model.Contact
	contactsDir  string
	currentView  ViewMode
	
	// List view state
	list         list.Model
	cursor       int
	selected     map[string]bool
	
	// Detail view state
	selectedContact *model.Contact
	
	// Contact logging state
	contactToMark      *model.Contact
	interactionType    string
	interactionState   string
	interactionNote    string
	contactLogStep     int // 0=type, 1=state, 2=note
	
	// Edit view state
	editingContact *model.Contact
	editField      int
	editValues     []string
	
	// Search/filter state
	searchQuery     string
	searchMode      bool              // true when typing search
	filtered        []model.Contact
	filterType      string            // Filter by relationship type
	filterState     string            // Filter by state
	filterStatus    string            // Filter by status (overdue, needsAttention, ok)
	showFilterPopup bool              // Show filter dialog
	
	// UI state
	width        int
	height       int
	ready        bool
	err          error
	message      string
	entryView    ViewMode  // The view to return to after completing an operation
}

// NewModel creates a new application model
func NewModel(contactsDir string) Model {
	return Model{
		contactsDir:  contactsDir,
		currentView:  ViewList,
		entryView:    ViewList, // Default to list view
		selected:     make(map[string]bool),
		contacts:     []model.Contact{},
		filtered:     []model.Contact{},
		width:        80,  // Default width
		height:       24,  // Default height
		ready:        true, // Start ready
		filterType:   "",  // Initialize as empty
		filterState:  "",  // Initialize as empty
		filterStatus: "",  // Initialize as empty
	}
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.loadContacts(),
	)
}

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.width == 0 {
			m.width = 80
		}
		if m.height == 0 {
			m.height = 24
		}
		m.ready = true
		return m, nil
		
	case tea.KeyMsg:
		switch m.currentView {
		case ViewList:
			if m.searchMode {
				return m.updateSearch(msg)
			}
			if m.showFilterPopup {
				return m.updateFilter(msg)
			}
			return m.updateList(msg)
		case ViewDetail:
			return m.updateDetail(msg)
		case ViewEdit:
			return m.updateEdit(msg)
		case ViewCreate:
			return m.updateCreate(msg)
		case ViewInteractionType:
			return m.updateInteractionType(msg)
		case ViewQuickType:
			return m.updateQuickType(msg)
		}
		
	case contactsLoadedMsg:
		m.contacts = msg.contacts
		m.filtered = m.contacts
		return m, nil
		
	case contactUpdatedMsg:
		// Update the contact in our lists
		for i, c := range m.contacts {
			if c.FilePath == msg.contact.FilePath {
				m.contacts[i] = msg.contact
				break
			}
		}
		
		// Update selected contact if it's the same one
		if m.selectedContact != nil && m.selectedContact.FilePath == msg.contact.FilePath {
			m.selectedContact = &msg.contact
		}
		
		// Re-apply filters to update the filtered list
		m.applyFilters()
		
		// Set the success message
		m.message = msg.message
		
		// Return to previous view and reset state
		if m.currentView == ViewInteractionType {
			m.currentView = m.entryView  // Return to where we came from
			m.contactToMark = nil
			m.interactionType = ""
			m.interactionState = ""
			m.interactionNote = ""
			m.contactLogStep = 0
		} else if m.currentView == ViewEdit {
			// Return to previous view after editing
			m.currentView = m.entryView  // Return to where we came from
			m.editingContact = nil
			m.editField = -1
			m.editValues = nil
		} else if m.currentView == ViewCreate {
			// Return to entry view after creating
			m.currentView = m.entryView  // Return to where we came from
			m.editField = -1
			m.editValues = nil
			// Reload contacts to include the new one
			return m, m.loadContacts()
		} else if m.currentView == ViewQuickType {
			// Return to entry view after quick type change
			m.currentView = m.entryView  // Return to where we came from
			m.contactToMark = nil
		}
		
		// Clear message after 3 seconds
		return m, clearMessageAfter(3 * time.Second)
		
	case clearMessageMsg:
		m.message = ""
		return m, nil
		
	case error:
		m.err = msg
		return m, nil
	}
	
	return m, nil
}

// View implements tea.Model
func (m Model) View() string {
	if !m.ready {
		return "Loading..."
	}
	
	if m.err != nil {
		return "Error: " + m.err.Error()
	}
	
	var view string
	switch m.currentView {
	case ViewList:
		view = m.viewList()
	case ViewDetail:
		view = m.viewDetail()
	case ViewEdit:
		view = m.viewEdit()
	case ViewCreate:
		view = m.viewCreate()
	case ViewInteractionType:
		view = m.viewInteractionType()
	case ViewQuickType:
		view = m.viewQuickType()
	default:
		view = m.viewList()
	}
	
	// Show filter screen if active
	if m.showFilterPopup && m.currentView == ViewList {
		view = m.renderFilterPopup()
	}
	
	return view
}