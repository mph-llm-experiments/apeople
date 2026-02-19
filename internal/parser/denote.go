package parser

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/mph-llm-experiments/apeople/internal/model"
	"gopkg.in/yaml.v3"
)

// ParseContactFile parses a Denote-format contact file
func ParseContactFile(path string) (model.Contact, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return model.Contact{}, fmt.Errorf("error reading file: %w", err)
	}

	// Split frontmatter and content
	parts := bytes.SplitN(content, []byte("---\n"), 3)
	if len(parts) < 3 {
		return model.Contact{}, fmt.Errorf("invalid file format: no frontmatter found")
	}

	// Parse YAML frontmatter
	var contact model.Contact
	if err := yaml.Unmarshal(parts[1], &contact); err != nil {
		return model.Contact{}, fmt.Errorf("error parsing frontmatter: %w", err)
	}

	// Validate required fields
	if !containsTag(contact.Tags, "contact") {
		return model.Contact{}, fmt.Errorf("not a contact file: missing 'contact' tag")
	}

	// Set runtime fields
	contact.FilePath = path
	contact.Content = string(parts[2])

	// Parse filename to extract identifier if not set
	// Accept both YYYYMMDD (legacy) and YYYYMMDDTHHMMSS (new) formats
	if contact.Identifier == "" {
		basename := strings.TrimSuffix(filepath.Base(path), ".md")
		if idx := strings.Index(basename, "--"); idx >= 0 {
			contact.Identifier = basename[:idx]
		}
	}

	// Initialize relation slices (ensures JSON outputs [] not null)
	contact.EnsureRelationSlices()

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

// SaveContactFile saves a contact to a Denote-format file
func SaveContactFile(contact model.Contact) error {
	// Generate filename if needed
	if contact.FilePath == "" {
		contact.FilePath = GenerateFilename(contact)
	}

	// Ensure updated_at is set
	contact.UpdatedAt = time.Now()

	// Marshal frontmatter
	frontmatter, err := yaml.Marshal(contact)
	if err != nil {
		return fmt.Errorf("error marshaling frontmatter: %w", err)
	}

	// Build file content
	var content bytes.Buffer
	content.WriteString("---\n")
	content.Write(frontmatter)
	content.WriteString("---\n")
	content.WriteString(contact.Content)

	// Write file
	return os.WriteFile(contact.FilePath, content.Bytes(), 0644)
}

// GenerateFilename generates a Denote-compliant filename for a contact
func GenerateFilename(contact model.Contact) string {
	// Use creation date or current date
	date := contact.Date
	if date.IsZero() {
		date = time.Now()
	}

	// Format: YYYYMMDDTHHMMSS--kebab-case-name__contact.md
	identifier := date.Format("20060102T150405")
	name := strings.ToLower(contact.Title)
	name = strings.ReplaceAll(name, " ", "-")
	name = sanitizeName(name)

	return fmt.Sprintf("%s--%s__contact.md", identifier, name)
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

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip directories and non-contact files
		if info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		if !strings.Contains(filepath.Base(path), "__contact") {
			return nil
		}

		contact, err := ParseContactFile(path)
		if err != nil {
			return nil // skip unparseable files
		}

		contact.FilePath = path
		contacts = append(contacts, contact)
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort alphabetically by name
	sort.Slice(contacts, func(i, j int) bool {
		return strings.ToLower(contacts[i].Title) < strings.ToLower(contacts[j].Title)
	})

	return contacts, nil
}

// AssignIndexIDs ensures all contacts have index_id values, assigning new ones as needed
func AssignIndexIDs(dir string, contacts []model.Contact) ([]model.Contact, error) {
	counter, err := GetIDCounter(dir)
	if err != nil {
		return contacts, fmt.Errorf("failed to get ID counter: %w", err)
	}

	for i, c := range contacts {
		if c.IndexID == 0 {
			id, err := counter.NextID()
			if err != nil {
				return contacts, fmt.Errorf("failed to assign index_id: %w", err)
			}
			contacts[i].IndexID = id
			// Write the index_id back to the file
			if err := SaveContactFile(contacts[i]); err != nil {
				return contacts, fmt.Errorf("failed to save index_id for %s: %w", c.Title, err)
			}
		}
	}

	return contacts, nil
}

// FindContactByID finds a contact by index_id or denote identifier
func FindContactByID(contacts []model.Contact, id string) *model.Contact {
	// Try as numeric index_id first
	for i, c := range contacts {
		if fmt.Sprintf("%d", c.IndexID) == id {
			return &contacts[i]
		}
	}

	// Try as denote identifier
	for i, c := range contacts {
		if c.Identifier == id {
			return &contacts[i]
		}
	}

	return nil
}

// sanitizeName removes special characters and ensures valid filename
func sanitizeName(name string) string {
	// Remove special characters, keep only alphanumeric and hyphens
	var result strings.Builder
	for _, ch := range name {
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-' {
			result.WriteRune(ch)
		}
	}
	return result.String()
}

// AppendInteractionLog adds a log entry to the content's Interaction Log section.
// If no "## Interaction Log" section exists, one is created.
// New entries are inserted at the top of the log (most recent first).
func AppendInteractionLog(content string, entry string) string {
	const header = "## Interaction Log"
	idx := strings.Index(content, header)
	if idx >= 0 {
		// Find the position right after the header line
		afterHeader := idx + len(header)
		// Skip past any newlines after the header
		rest := content[afterHeader:]
		insertPos := afterHeader
		for i, ch := range rest {
			if ch == '\n' {
				insertPos = afterHeader + i + 1
			} else {
				break
			}
		}
		// Insert the new entry at the top of the log
		return content[:insertPos] + entry + "\n" + content[insertPos:]
	}

	// No Interaction Log section — create one
	trimmed := strings.TrimRight(content, "\n")
	if trimmed == "" {
		return "\n" + header + "\n\n" + entry + "\n"
	}
	return trimmed + "\n\n" + header + "\n\n" + entry + "\n"
}

// containsTag checks if a tag exists in the tags slice
func containsTag(tags []string, tag string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}
