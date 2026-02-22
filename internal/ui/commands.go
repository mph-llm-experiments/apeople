package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mph-llm-experiments/acore"
	"github.com/mph-llm-experiments/apeople/internal/model"
	"github.com/mph-llm-experiments/apeople/internal/parser"
)

// Message types
type contactsLoadedMsg struct {
	contacts []model.Contact
}

type contactSelectedMsg struct {
	contact model.Contact
}

type errorMsg struct {
	err error
}

type contactUpdatedMsg struct {
	contact model.Contact
	message string
}

type clearMessageMsg struct{}

// loadContacts returns a command that loads all contacts from the directory
func (m Model) loadContacts() tea.Cmd {
	return func() tea.Msg {
		contacts, err := parser.FindContacts(m.contactsDir)
		if err != nil {
			return errorMsg{err: err}
		}

		contacts, err = parser.AssignIndexIDs(m.contactsDir, contacts)
		if err != nil {
			return errorMsg{err: err}
		}

		return contactsLoadedMsg{contacts: contacts}
	}
}

// logContactInteraction returns a command that logs a complete interaction
func (m Model) logContactInteraction(contact model.Contact) tea.Cmd {
	return func() tea.Msg {
		// Update the contact with all interaction details
		now := time.Now()
		contact.LastContacted = &now
		contact.LastInteractionType = m.interactionType
		oldState := contact.State
		contact.State = m.interactionState
		
		// Add interaction log entry
		logEntry := fmt.Sprintf("- **%s** (%s)", now.Format("2006-01-02"), m.interactionType)
		if m.interactionNote != "" {
			logEntry += fmt.Sprintf(" - %s", m.interactionNote)
		}
		contact.Content = parser.AppendInteractionLog(contact.Content, logEntry)
		
		// Save the updated contact
		err := parser.SaveContactFile(contact)
		if err != nil {
			return errorMsg{err: fmt.Errorf("failed to save interaction for '%s': %v", contact.Title, err)}
		}
		
		// Create task if state changed to one requiring action
		var taskCreated bool
		var taskError string
		if oldState != m.interactionState {
			if err := m.createTaskForContact(contact, m.interactionState); err != nil {
				// Include error in message so user knows what happened
				taskError = fmt.Sprintf(" [task error: %v]", err)
			} else if _, needsTask := map[string]bool{
				"followup": true, "ping": true, "scheduled": true, 
				"timeout": true,
			}[m.interactionState]; needsTask {
				taskCreated = true
			}
		}
		
		// Reload the contact to get the updated state
		updatedContact, err := parser.ParseContactFile(contact.FilePath)
		if err != nil {
			return errorMsg{err: fmt.Errorf("failed to reload contact '%s' after logging interaction: %v", contact.Title, err)}
		}
		
		message := fmt.Sprintf("Logged %s interaction with %s", m.interactionType, contact.Title)
		if m.interactionState != "ok" {
			message += fmt.Sprintf(" (→ %s)", m.interactionState)
		}
		if taskCreated {
			message += " [task created]"
		}
		message += taskError
		
		return contactUpdatedMsg{
			contact: updatedContact,
			message: message,
		}
	}
}

// clearMessageAfter returns a command that clears the message after a delay
func clearMessageAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg {
		return clearMessageMsg{}
	})
}

// bumpContact returns a command that updates a contact's bump date
func (m Model) bumpContact(contact model.Contact) tea.Cmd {
	return func() tea.Msg {
		// Update the bump date and increment count
		now := time.Now()
		contact.LastBumpDate = &now
		contact.BumpCount++
		
		// Save the updated contact
		err := parser.SaveContactFile(contact)
		if err != nil {
			return errorMsg{err: fmt.Errorf("failed to save bump for '%s': %v", contact.Title, err)}
		}
		
		// Reload the contact to get the updated state
		updatedContact, err := parser.ParseContactFile(contact.FilePath)
		if err != nil {
			return errorMsg{err: fmt.Errorf("failed to reload contact '%s' after bump: %v", contact.Title, err)}
		}
		
		return contactUpdatedMsg{
			contact: updatedContact,
			message: fmt.Sprintf("Bumped %s (review #%d)", contact.Title, contact.BumpCount),
		}
	}
}

// saveEditedContact returns a command that saves the edited contact
func (m Model) saveEditedContact() tea.Cmd {
	return func() tea.Msg {
		if m.editingContact == nil {
			return errorMsg{err: fmt.Errorf("no contact being edited")}
		}
		
		// Apply edited values to the contact
		contact := *m.editingContact
		oldState := contact.State
		
		// Update basic fields
		contact.Title = strings.TrimSpace(m.editValues[fieldTitle])
		contact.Email = strings.TrimSpace(m.editValues[fieldEmail])
		contact.Phone = strings.TrimSpace(m.editValues[fieldPhone])
		contact.Company = strings.TrimSpace(m.editValues[fieldCompany])
		contact.Role = strings.TrimSpace(m.editValues[fieldRole])
		contact.Location = strings.TrimSpace(m.editValues[fieldLocation])
		contact.RelationshipType = model.RelationshipType(strings.TrimSpace(m.editValues[fieldRelationType]))
		contact.ContactStyle = model.ContactStyle(strings.TrimSpace(m.editValues[fieldContactStyle]))
		contact.State = strings.TrimSpace(m.editValues[fieldState])
		
		// Parse and update tags
		tagStr := strings.TrimSpace(m.editValues[fieldTags])
		tags := []string{"contact"} // Always include the contact tag
		if tagStr != "" {
			for _, tag := range strings.Fields(tagStr) {
				tag = strings.TrimPrefix(tag, "#")
				if tag != "" && tag != "contact" {
					tags = append(tags, tag)
				}
			}
		}
		contact.Tags = tags
		
		// Update the updated_at timestamp
		now := time.Now()
		contact.Modified = now.UTC().Format(time.RFC3339)
		
		// Save the updated contact
		err := parser.SaveContactFile(contact)
		if err != nil {
			return errorMsg{err: fmt.Errorf("failed to save changes to '%s': %v", contact.Title, err)}
		}
		
		// Create task if state changed to one requiring action
		var taskCreated bool
		if oldState != contact.State {
			if err := m.createTaskForContact(contact, contact.State); err != nil {
				// Log error but don't fail the edit
				// The contact update was successful even if task creation failed
			} else if _, needsTask := map[string]bool{
				"followup": true, "ping": true, "scheduled": true,
				"timeout": true,
			}[contact.State]; needsTask {
				taskCreated = true
			}
		}
		
		// Reload the contact to get the updated state
		updatedContact, err := parser.ParseContactFile(contact.FilePath)
		if err != nil {
			return errorMsg{err: fmt.Errorf("failed to reload contact '%s' after editing: %v", contact.Title, err)}
		}
		
		message := fmt.Sprintf("Updated %s", contact.Title)
		if taskCreated {
			message += " [task created]"
		}
		
		return contactUpdatedMsg{
			contact: updatedContact,
			message: message,
		}
	}
}

// createTaskForContact creates a task when a contact changes to an action-requiring state
func (m Model) createTaskForContact(contact model.Contact, newState string) error {
	// Only create tasks for states that require action
	actionStates := map[string]string{
		"followup":  "Follow up with",
		"ping":      "Ping",
		"scheduled": "Meeting with",
		"timeout":   "Follow up with",
	}
	
	taskPrefix, needsTask := actionStates[newState]
	if !needsTask {
		return nil // No task needed for this state
	}
	
	// Generate task title
	taskTitle := fmt.Sprintf("%s %s", taskPrefix, contact.Title)
	if newState == "timeout" {
		taskTitle += " (no response)"
	}
	
	// Generate task using acore identity
	now := time.Now()
	taskID := acore.NewID()

	// Create tags based on contact
	tags := []string{fmt.Sprintf("contact-%s", newState)}

	// Create task content
	var taskContent strings.Builder
	taskContent.WriteString("---\n")
	taskContent.WriteString(fmt.Sprintf("id: %s\n", taskID))
	taskContent.WriteString(fmt.Sprintf("title: %s\n", taskTitle))
	taskContent.WriteString(fmt.Sprintf("type: task\n"))
	taskContent.WriteString(fmt.Sprintf("tags: [%s]\n", strings.Join(tags, ", ")))
	taskContent.WriteString(fmt.Sprintf("created: %s\n", now.UTC().Format(time.RFC3339)))
	taskContent.WriteString(fmt.Sprintf("modified: %s\n", now.UTC().Format(time.RFC3339)))
	taskContent.WriteString("status: open\n")
	if contact.Label != "" {
		taskContent.WriteString(fmt.Sprintf("label: %s\n", contact.Label))
	}
	taskContent.WriteString(fmt.Sprintf("related_people:\n  - %s\n", contact.ID))
	taskContent.WriteString("---\n\n")
	
	// Add task description
	switch newState {
	case "followup":
		taskContent.WriteString(fmt.Sprintf("Follow up with %s regarding previous conversation.\n", contact.Title))
	case "ping":
		taskContent.WriteString(fmt.Sprintf("Send a quick check-in message to %s.\n", contact.Title))
	case "scheduled":
		taskContent.WriteString(fmt.Sprintf("Scheduled meeting or call with %s.\n", contact.Title))
	case "timeout":
		taskContent.WriteString(fmt.Sprintf("%s has not responded. Consider following up or closing the loop.\n", contact.Title))
	}
	
	// Save task file using acore filename convention
	filename := acore.BuildFilename(taskID, taskTitle, "task")
	// Save tasks to the atask directory via acore config
	acoreCfg, _ := acore.LoadConfig()
	notesDir := acoreCfg.DirFor("atask")
	
	// Create notes directory if it doesn't exist
	if err := os.MkdirAll(notesDir, 0755); err != nil {
		return fmt.Errorf("failed to create notes directory: %v", err)
	}
	
	taskPath := filepath.Join(notesDir, filename)
	
	if err := os.WriteFile(taskPath, []byte(taskContent.String()), 0644); err != nil {
		return fmt.Errorf("failed to create task file '%s': %v", filename, err)
	}
	
	return nil
}

// saveQuickTypeChange returns a command that saves a quick type change
func (m Model) saveQuickTypeChange(contact model.Contact) tea.Cmd {
	return func() tea.Msg {
		// Update the updated_at timestamp
		now := time.Now()
		contact.Modified = now.UTC().Format(time.RFC3339)

		// Save the updated contact
		err := parser.SaveContactFile(contact)
		if err != nil {
			return errorMsg{err: fmt.Errorf("failed to update type for '%s': %v", contact.Title, err)}
		}

		// Reload the contact to get the updated state
		updatedContact, err := parser.ParseContactFile(contact.FilePath)
		if err != nil {
			return errorMsg{err: fmt.Errorf("failed to reload contact '%s' after type change: %v", contact.Title, err)}
		}

		return contactUpdatedMsg{
			contact: updatedContact,
			message: fmt.Sprintf("Changed %s to %s", contact.Title, contact.RelationshipType),
		}
	}
}

// saveNewContact returns a command that creates and saves a new contact
func (m Model) saveNewContact() tea.Cmd {
	return func() tea.Msg {
		// Validate required fields
		name := strings.TrimSpace(m.editValues[fieldTitle])
		if name == "" {
			return errorMsg{err: fmt.Errorf("name is required")}
		}
		
		// Create new contact from form values
		contact := parser.NewContact(name, m.contactsDir)
		contact.Email = strings.TrimSpace(m.editValues[fieldEmail])
		contact.Phone = strings.TrimSpace(m.editValues[fieldPhone])
		contact.Company = strings.TrimSpace(m.editValues[fieldCompany])
		contact.Role = strings.TrimSpace(m.editValues[fieldRole])
		contact.Location = strings.TrimSpace(m.editValues[fieldLocation])
		contact.RelationshipType = model.RelationshipType(strings.TrimSpace(m.editValues[fieldRelationType]))
		contact.ContactStyle = model.ContactStyle(strings.TrimSpace(m.editValues[fieldContactStyle]))
		contact.State = strings.TrimSpace(m.editValues[fieldState])
		
		// Parse and set tags
		tagStr := strings.TrimSpace(m.editValues[fieldTags])
		tags := []string{"contact"} // Always include the contact tag
		if tagStr != "" {
			for _, tag := range strings.Fields(tagStr) {
				tag = strings.TrimPrefix(tag, "#")
				if tag != "" && tag != "contact" {
					tags = append(tags, tag)
				}
			}
		}
		contact.Tags = tags
		
		// Check if contacts directory exists before trying to save
		if _, err := os.Stat(m.contactsDir); os.IsNotExist(err) {
			return errorMsg{err: fmt.Errorf("cannot create contact: directory '%s' does not exist. Please create it first", m.contactsDir)}
		} else if err != nil {
			return errorMsg{err: fmt.Errorf("cannot access contacts directory '%s': %v", m.contactsDir, err)}
		}
		
		// Generate filename using acore conventions
		contact.FilePath = parser.GenerateFilePath(m.contactsDir, contact)
		
		// Save the new contact
		err := parser.SaveContactFile(contact)
		if err != nil {
			return errorMsg{err: fmt.Errorf("failed to save contact '%s': %v", name, err)}
		}
		
		// Create task if new contact has an action-requiring state
		var taskCreated bool
		if contact.State != "" && contact.State != "ok" {
			if err := m.createTaskForContact(contact, contact.State); err != nil {
				// Log error but don't fail the contact creation
				// The contact was created successfully even if task creation failed
			} else if _, needsTask := map[string]bool{
				"followup": true, "ping": true, "scheduled": true,
				"timeout": true,
			}[contact.State]; needsTask {
				taskCreated = true
			}
		}
		
		// Reload the contact to get the saved state
		savedContact, err := parser.ParseContactFile(contact.FilePath)
		if err != nil {
			return errorMsg{err: fmt.Errorf("created contact '%s' but failed to reload it: %v", contact.Title, err)}
		}
		
		message := fmt.Sprintf("Created %s", contact.Title)
		if taskCreated {
			message += " [task created]"
		}
		
		return contactUpdatedMsg{
			contact: savedContact,
			message: message,
		}
	}
}