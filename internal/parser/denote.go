package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/mph-llm-experiments/acore"
	"github.com/mph-llm-experiments/apeople/internal/model"
)

// ParseContactFile parses an acore-format contact file
func ParseContactFile(path string) (model.Contact, error) {
	var contact model.Contact
	content, err := acore.ReadFile(path, &contact)
	if err != nil {
		return model.Contact{}, fmt.Errorf("error parsing contact file: %w", err)
	}

	// Set runtime fields
	contact.FilePath = path
	contact.Content = content

	// Extract ID from filename if not in frontmatter (legacy support during migration)
	if contact.ID == "" {
		basename := strings.TrimSuffix(filepath.Base(path), ".md")
		if idx := strings.Index(basename, "--"); idx >= 0 {
			contact.ID = basename[:idx]
		}
	}

	// Initialize relation slices (ensures JSON outputs [] not null)
	contact.EnsureSlices()

	// Compute runtime fields
	contact.DaysSince = contact.DaysSinceContact()
	if contact.IsOverdue() {
		contact.OverdueStatus = "overdue"
	} else if contact.NeedsAttention() {
		contact.OverdueStatus = "attention"
	} else if contact.IsWithinThreshold() {
		contact.OverdueStatus = "good"
	}

	return contact, nil
}

// SaveContactFile saves a contact to an acore-format file
func SaveContactFile(contact model.Contact) error {
	if contact.FilePath == "" {
		return fmt.Errorf("contact has no file path")
	}

	// Update modified timestamp
	contact.Modified = acore.Now()

	return acore.WriteFile(contact.FilePath, &contact, contact.Content)
}

// GenerateFilePath generates a file path for a new contact using acore conventions.
func GenerateFilePath(dir string, contact model.Contact) string {
	filename := acore.BuildFilename(contact.ID, contact.Title, "contact")
	return filepath.Join(dir, filename)
}

// FindContacts loads all contact files from a directory, sorted alphabetically
func FindContacts(dir string) ([]model.Contact, error) {
	contacts := []model.Contact{}

	if info, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("contacts directory '%s' does not exist", dir)
		}
		return nil, fmt.Errorf("cannot access contacts directory '%s': %v", dir, err)
	} else if !info.IsDir() {
		return nil, fmt.Errorf("contacts path '%s' is not a directory", dir)
	}

	scanner := &acore.Scanner{Dir: dir}
	files, err := scanner.FindByType("contact")
	if err != nil {
		return nil, err
	}

	for _, path := range files {
		contact, err := ParseContactFile(path)
		if err != nil {
			continue // skip unparseable files
		}
		contacts = append(contacts, contact)
	}

	// Sort alphabetically by name
	sort.Slice(contacts, func(i, j int) bool {
		return strings.ToLower(contacts[i].Title) < strings.ToLower(contacts[j].Title)
	})

	return contacts, nil
}

// AssignIndexIDs ensures all contacts have index_id values, assigning new ones as needed
func AssignIndexIDs(dir string, contacts []model.Contact) ([]model.Contact, error) {
	counter, err := acore.NewIndexCounter(dir, "apeople")
	if err != nil {
		return contacts, fmt.Errorf("failed to get ID counter: %w", err)
	}

	for i, c := range contacts {
		if c.IndexID == 0 {
			id, err := counter.Next()
			if err != nil {
				return contacts, fmt.Errorf("failed to assign index_id: %w", err)
			}
			contacts[i].IndexID = id
			if err := SaveContactFile(contacts[i]); err != nil {
				return contacts, fmt.Errorf("failed to save index_id for %s: %w", c.Title, err)
			}
		}
	}

	return contacts, nil
}

// FindContactByID finds a contact by index_id or ULID
func FindContactByID(contacts []model.Contact, id string) *model.Contact {
	// Try as numeric index_id first
	for i, c := range contacts {
		if fmt.Sprintf("%d", c.IndexID) == id {
			return &contacts[i]
		}
	}

	// Try as ULID (or legacy Denote identifier)
	for i, c := range contacts {
		if c.ID == id {
			return &contacts[i]
		}
	}

	return nil
}

// AppendInteractionLog adds a log entry to the content's Interaction Log section.
// If no "## Interaction Log" section exists, one is created.
// New entries are inserted at the top of the log (most recent first).
func AppendInteractionLog(content string, entry string) string {
	const header = "## Interaction Log"
	idx := strings.Index(content, header)
	if idx >= 0 {
		afterHeader := idx + len(header)
		rest := content[afterHeader:]
		insertPos := afterHeader
		for i, ch := range rest {
			if ch == '\n' {
				insertPos = afterHeader + i + 1
			} else {
				break
			}
		}
		return content[:insertPos] + entry + "\n" + content[insertPos:]
	}

	trimmed := strings.TrimRight(content, "\n")
	if trimmed == "" {
		return "\n" + header + "\n\n" + entry + "\n"
	}
	return trimmed + "\n\n" + header + "\n\n" + entry + "\n"
}

// NewContact creates a new contact with acore identity.
func NewContact(title string, dir string) model.Contact {
	now := time.Now()
	id := acore.NewID()

	contact := model.Contact{}
	contact.ID = id
	contact.Title = title
	contact.Type = "contact"
	contact.Created = now.UTC().Format(time.RFC3339)
	contact.Modified = now.UTC().Format(time.RFC3339)

	return contact
}
