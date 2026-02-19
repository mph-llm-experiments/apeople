package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/mph-llm-experiments/apeople/internal/config"
	"github.com/mph-llm-experiments/apeople/internal/model"
	"github.com/mph-llm-experiments/apeople/internal/parser"
)

func listCommand(cfg *config.Config) *Command {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	relType := fs.String("type", "", "Filter by relationship type (close, family, network, work, social, providers, recruiters)")
	state := fs.String("state", "", "Filter by state (ok, followup, ping, active, archived)")
	style := fs.String("style", "", "Filter by contact style (periodic, ambient, triggered)")
	overdue := fs.Bool("overdue", false, "Show only overdue contacts")
	search := fs.String("search", "", "Search contacts by name, company, email, or tags")
	all := fs.Bool("all", false, "Show all contacts including archived")
	sortBy := fs.String("sort", "name", "Sort by: name, days, type, state")

	return &Command{
		Name:        "list",
		Usage:       "apeople list [options]",
		Description: "List contacts with optional filtering",
		Flags:       fs,
		Run: func(cmd *Command, args []string) error {
			contacts, err := parser.FindContacts(cfg.ContactsDirectory)
			if err != nil {
				return err
			}

			contacts, err = parser.AssignIndexIDs(cfg.ContactsDirectory, contacts)
			if err != nil {
				return err
			}

			// Apply filters
			var filtered []model.Contact
			for _, c := range contacts {
				if !*all && c.State == "archived" {
					continue
				}
				if *relType != "" && string(c.RelationshipType) != *relType {
					continue
				}
				if *state != "" && c.State != *state {
					continue
				}
				if *style != "" && string(c.ContactStyle) != *style {
					continue
				}
				if *overdue && !c.IsOverdue() {
					continue
				}
				if *search != "" {
					query := strings.ToLower(*search)
					match := strings.Contains(strings.ToLower(c.Title), query) ||
						strings.Contains(strings.ToLower(c.Company), query) ||
						strings.Contains(strings.ToLower(c.Email), query) ||
						strings.Contains(strings.ToLower(c.Role), query)
					if !match {
						for _, tag := range c.Tags {
							if strings.Contains(strings.ToLower(tag), query) {
								match = true
								break
							}
						}
					}
					if !match {
						continue
					}
				}
				filtered = append(filtered, c)
			}

			// Sort
			switch *sortBy {
			case "days":
				sort.Slice(filtered, func(i, j int) bool {
					return filtered[i].DaysSinceContact() > filtered[j].DaysSinceContact()
				})
			case "type":
				sort.Slice(filtered, func(i, j int) bool {
					return string(filtered[i].RelationshipType) < string(filtered[j].RelationshipType)
				})
			case "state":
				sort.Slice(filtered, func(i, j int) bool {
					return filtered[i].State < filtered[j].State
				})
			default: // "name"
				sort.Slice(filtered, func(i, j int) bool {
					return strings.ToLower(filtered[i].Title) < strings.ToLower(filtered[j].Title)
				})
			}

			if globalFlags.JSON {
				data, err := json.MarshalIndent(filtered, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal JSON: %w", err)
				}
				fmt.Println(string(data))
				return nil
			}

			if len(filtered) == 0 {
				fmt.Println("No contacts found.")
				return nil
			}

			// Text output
			fmt.Printf("%-4s %-22s %5s  %-10s %-10s %-20s %s\n",
				"#", "NAME", "DAYS", "TYPE", "STATE", "COMPANY", "TAGS")
			fmt.Println(strings.Repeat("-", 90))

			for _, c := range filtered {
				days := c.DaysSinceContact()
				daysStr := "-"
				if days >= 0 {
					daysStr = fmt.Sprintf("%d", days)
				}

				name := c.Title
				if len(name) > 22 {
					name = name[:19] + "..."
				}

				company := c.Company
				if len(company) > 20 {
					company = company[:17] + "..."
				}

				var tagStrs []string
				for _, t := range c.Tags {
					if t != "contact" {
						tagStrs = append(tagStrs, "#"+t)
					}
				}

				stateStr := c.State
				if stateStr == "" {
					stateStr = "-"
				}

				typeStr := string(c.RelationshipType)
				if typeStr == "" {
					typeStr = "-"
				}

				fmt.Printf("%-4d %-22s %5s  %-10s %-10s %-20s %s\n",
					c.IndexID, name, daysStr, typeStr, stateStr, company, strings.Join(tagStrs, " "))
			}

			return nil
		},
	}
}

func showCommand(cfg *config.Config) *Command {
	return &Command{
		Name:        "show",
		Usage:       "apeople show <id>",
		Description: "Show contact details by index_id or denote identifier",
		Run: func(cmd *Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("usage: apeople show <id>")
			}

			contacts, err := parser.FindContacts(cfg.ContactsDirectory)
			if err != nil {
				return err
			}
			contacts, err = parser.AssignIndexIDs(cfg.ContactsDirectory, contacts)
			if err != nil {
				return err
			}

			contact := parser.FindContactByID(contacts, args[0])
			if contact == nil {
				return fmt.Errorf("contact not found: %s", args[0])
			}

			if globalFlags.JSON {
				data, err := json.MarshalIndent(contact, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal JSON: %w", err)
				}
				fmt.Println(string(data))
				return nil
			}

			// Text output
			fmt.Printf("# %s (#%d)\n\n", contact.Title, contact.IndexID)

			if contact.Email != "" {
				fmt.Printf("  Email:     %s\n", contact.Email)
			}
			if contact.Phone != "" {
				fmt.Printf("  Phone:     %s\n", contact.Phone)
			}
			if contact.Company != "" {
				fmt.Printf("  Company:   %s\n", contact.Company)
			}
			if contact.Role != "" {
				fmt.Printf("  Role:      %s\n", contact.Role)
			}
			if contact.Location != "" {
				fmt.Printf("  Location:  %s\n", contact.Location)
			}
			if contact.LinkedIn != "" {
				fmt.Printf("  LinkedIn:  %s\n", contact.LinkedIn)
			}
			if contact.Website != "" {
				fmt.Printf("  Website:   %s\n", contact.Website)
			}
			fmt.Println()

			fmt.Printf("  Type:      %s\n", contact.RelationshipType)
			if contact.ContactStyle != "" {
				fmt.Printf("  Style:     %s\n", contact.ContactStyle)
			}
			if contact.State != "" {
				fmt.Printf("  State:     %s\n", contact.State)
			}
			if contact.Label != "" {
				fmt.Printf("  Label:     %s\n", contact.Label)
			}

			freq := contact.GetFrequencyDays()
			if freq > 0 {
				fmt.Printf("  Frequency: %d days\n", freq)
			}
			fmt.Println()

			days := contact.DaysSinceContact()
			if days >= 0 {
				fmt.Printf("  Last contacted: %d days ago", days)
				if contact.LastInteractionType != "" {
					fmt.Printf(" (%s)", contact.LastInteractionType)
				}
				fmt.Println()
			} else {
				fmt.Println("  Last contacted: never")
			}
			if contact.LastBumpDate != nil {
				fmt.Printf("  Last bump:      %s (count: %d)\n", contact.LastBumpDate.Format("2006-01-02"), contact.BumpCount)
			}
			fmt.Printf("  Created:        %s\n", contact.Date.Format("2006-01-02"))
			fmt.Printf("  Updated:        %s\n", contact.UpdatedAt.Format("2006-01-02"))

			var tagStrs []string
			for _, t := range contact.Tags {
				if t != "contact" {
					tagStrs = append(tagStrs, "#"+t)
				}
			}
			if len(tagStrs) > 0 {
				fmt.Printf("\n  Tags: %s\n", strings.Join(tagStrs, " "))
			}

			if strings.TrimSpace(contact.Content) != "" {
				fmt.Printf("\n---\n%s", contact.Content)
			}

			return nil
		},
	}
}

func newCommand(cfg *config.Config) *Command {
	fs := flag.NewFlagSet("new", flag.ContinueOnError)
	relType := fs.String("type", "network", "Relationship type (close, family, network, work, social, providers, recruiters)")
	style := fs.String("style", "periodic", "Contact style (periodic, ambient, triggered)")
	email := fs.String("email", "", "Email address")
	phone := fs.String("phone", "", "Phone number")
	company := fs.String("company", "", "Company name")
	role := fs.String("role", "", "Role/title")
	tags := fs.String("tags", "", "Comma-separated tags (in addition to 'contact')")
	state := fs.String("state", "ok", "Contact state (ok, active, followup, ping, archived)")
	location := fs.String("location", "", "Location")

	return &Command{
		Name:        "new",
		Usage:       "apeople new \"Name\" [options]",
		Description: "Create a new contact",
		Flags:       fs,
		Run: func(cmd *Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("usage: apeople new \"Name\" [options]")
			}

			name := strings.Join(args, " ")
			now := time.Now()
			dateStr := now.Format("20060102T150405")

			// Build tags
			contactTags := []string{"contact"}
			if *tags != "" {
				for _, t := range strings.Split(*tags, ",") {
					t = strings.TrimSpace(t)
					if t != "" && t != "contact" {
						contactTags = append(contactTags, t)
					}
				}
			}

			contact := model.Contact{
				Title:            name,
				Date:             now,
				Identifier:       dateStr,
				Tags:             contactTags,
				RelationshipType: model.RelationshipType(*relType),
				ContactStyle:     model.ContactStyle(*style),
				State:            *state,
				Email:            *email,
				Phone:            *phone,
				Company:          *company,
				Role:             *role,
				Location:         *location,
				UpdatedAt:        now,
			}

			// Get index_id
			counter, err := parser.GetIDCounter(cfg.ContactsDirectory)
			if err != nil {
				return fmt.Errorf("failed to get ID counter: %w", err)
			}
			id, err := counter.NextID()
			if err != nil {
				return fmt.Errorf("failed to get next ID: %w", err)
			}
			contact.IndexID = id

			// Generate filename and filepath
			nameSlug := strings.ToLower(strings.ReplaceAll(name, " ", "-"))
			nameSlug = strings.ReplaceAll(nameSlug, "'", "")
			nameSlug = strings.ReplaceAll(nameSlug, ".", "")
			filename := fmt.Sprintf("%s--%s__contact.md", dateStr, sanitizeSlug(nameSlug))
			contact.FilePath = filepath.Join(cfg.ContactsDirectory, filename)

			if err := parser.SaveContactFile(contact); err != nil {
				return fmt.Errorf("failed to create contact: %w", err)
			}

			if globalFlags.JSON {
				// Reload to get saved state
				saved, err := parser.ParseContactFile(contact.FilePath)
				if err != nil {
					return fmt.Errorf("created but failed to reload: %w", err)
				}
				saved.IndexID = contact.IndexID
				data, _ := json.MarshalIndent(saved, "", "  ")
				fmt.Println(string(data))
				return nil
			}

			if !globalFlags.Quiet {
				fmt.Printf("Created contact #%d: %s\n", contact.IndexID, name)
			}
			return nil
		},
	}
}

func updateCommand(cfg *config.Config) *Command {
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	name := fs.String("name", "", "Update name")
	relType := fs.String("type", "", "Update relationship type")
	style := fs.String("style", "", "Update contact style")
	email := fs.String("email", "", "Update email")
	phone := fs.String("phone", "", "Update phone")
	company := fs.String("company", "", "Update company")
	role := fs.String("role", "", "Update role")
	tags := fs.String("tags", "", "Set tags (comma-separated, replaces existing non-contact tags)")
	state := fs.String("state", "", "Update state")
	location := fs.String("location", "", "Update location")

	return &Command{
		Name:        "update",
		Usage:       "apeople update <id> [options]",
		Description: "Update contact fields",
		Flags:       fs,
		Run: func(cmd *Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("usage: apeople update <id> [options]")
			}

			contacts, err := parser.FindContacts(cfg.ContactsDirectory)
			if err != nil {
				return err
			}
			contacts, err = parser.AssignIndexIDs(cfg.ContactsDirectory, contacts)
			if err != nil {
				return err
			}

			contact := parser.FindContactByID(contacts, args[0])
			if contact == nil {
				return fmt.Errorf("contact not found: %s", args[0])
			}

			// Apply updates
			if *name != "" {
				contact.Title = *name
			}
			if *relType != "" {
				contact.RelationshipType = model.RelationshipType(*relType)
			}
			if *style != "" {
				contact.ContactStyle = model.ContactStyle(*style)
			}
			if *email != "" {
				contact.Email = *email
			}
			if *phone != "" {
				contact.Phone = *phone
			}
			if *company != "" {
				contact.Company = *company
			}
			if *role != "" {
				contact.Role = *role
			}
			if *location != "" {
				contact.Location = *location
			}
			if *state != "" {
				contact.State = *state
			}
			if *tags != "" {
				contactTags := []string{"contact"}
				for _, t := range strings.Split(*tags, ",") {
					t = strings.TrimSpace(t)
					if t != "" && t != "contact" {
						contactTags = append(contactTags, t)
					}
				}
				contact.Tags = contactTags
			}

			if err := parser.SaveContactFile(*contact); err != nil {
				return fmt.Errorf("failed to update contact: %w", err)
			}

			if globalFlags.JSON {
				saved, err := parser.ParseContactFile(contact.FilePath)
				if err != nil {
					return fmt.Errorf("updated but failed to reload: %w", err)
				}
				saved.IndexID = contact.IndexID
				data, _ := json.MarshalIndent(saved, "", "  ")
				fmt.Println(string(data))
				return nil
			}

			if !globalFlags.Quiet {
				fmt.Printf("Updated contact #%d: %s\n", contact.IndexID, contact.Title)
			}
			return nil
		},
	}
}

func logCommand(cfg *config.Config) *Command {
	fs := flag.NewFlagSet("log", flag.ContinueOnError)
	interaction := fs.String("interaction", "", "Interaction type (required: email, call, text, meeting, social, bump, note)")
	state := fs.String("state", "", "Set new state after interaction")
	note := fs.String("note", "", "Add a note about the interaction")

	return &Command{
		Name:        "log",
		Usage:       "apeople log <id> --interaction <type> [options]",
		Description: "Log an interaction with a contact",
		Flags:       fs,
		Run: func(cmd *Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("usage: apeople log <id> --interaction <type>")
			}
			if *interaction == "" {
				return fmt.Errorf("--interaction is required (email, call, text, meeting, social, bump, note)")
			}

			contacts, err := parser.FindContacts(cfg.ContactsDirectory)
			if err != nil {
				return err
			}
			contacts, err = parser.AssignIndexIDs(cfg.ContactsDirectory, contacts)
			if err != nil {
				return err
			}

			contact := parser.FindContactByID(contacts, args[0])
			if contact == nil {
				return fmt.Errorf("contact not found: %s", args[0])
			}

			now := time.Now()
			contact.LastContacted = &now
			contact.LastInteractionType = *interaction

			if *state != "" {
				contact.State = *state
			}

			// Build interaction log entry
			logEntry := fmt.Sprintf("- **%s** (%s)", now.Format("2006-01-02"), *interaction)
			if *note != "" {
				logEntry += fmt.Sprintf(" - %s", *note)
			}
			contact.Content = parser.AppendInteractionLog(contact.Content, logEntry)

			if err := parser.SaveContactFile(*contact); err != nil {
				return fmt.Errorf("failed to log interaction: %w", err)
			}

			if globalFlags.JSON {
				saved, err := parser.ParseContactFile(contact.FilePath)
				if err != nil {
					return fmt.Errorf("logged but failed to reload: %w", err)
				}
				saved.IndexID = contact.IndexID
				data, _ := json.MarshalIndent(saved, "", "  ")
				fmt.Println(string(data))
				return nil
			}

			if !globalFlags.Quiet {
				msg := fmt.Sprintf("Logged %s interaction with %s (#%d)", *interaction, contact.Title, contact.IndexID)
				if *state != "" {
					msg += fmt.Sprintf(" [state → %s]", *state)
				}
				fmt.Println(msg)
			}
			return nil
		},
	}
}

func bumpCommand(cfg *config.Config) *Command {
	return &Command{
		Name:        "bump",
		Usage:       "apeople bump <id>",
		Description: "Bump a contact (review without contacting)",
		Run: func(cmd *Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("usage: apeople bump <id>")
			}

			contacts, err := parser.FindContacts(cfg.ContactsDirectory)
			if err != nil {
				return err
			}
			contacts, err = parser.AssignIndexIDs(cfg.ContactsDirectory, contacts)
			if err != nil {
				return err
			}

			contact := parser.FindContactByID(contacts, args[0])
			if contact == nil {
				return fmt.Errorf("contact not found: %s", args[0])
			}

			now := time.Now()
			contact.LastBumpDate = &now
			contact.BumpCount++

			if err := parser.SaveContactFile(*contact); err != nil {
				return fmt.Errorf("failed to bump contact: %w", err)
			}

			if globalFlags.JSON {
				saved, err := parser.ParseContactFile(contact.FilePath)
				if err != nil {
					return fmt.Errorf("bumped but failed to reload: %w", err)
				}
				saved.IndexID = contact.IndexID
				data, _ := json.MarshalIndent(saved, "", "  ")
				fmt.Println(string(data))
				return nil
			}

			if !globalFlags.Quiet {
				fmt.Printf("Bumped %s (#%d) — review #%d\n", contact.Title, contact.IndexID, contact.BumpCount)
			}
			return nil
		},
	}
}

func deleteCommand(cfg *config.Config) *Command {
	fs := flag.NewFlagSet("delete", flag.ContinueOnError)
	confirm := fs.Bool("confirm", false, "Skip confirmation prompt")

	return &Command{
		Name:        "delete",
		Usage:       "apeople delete <id> [--confirm]",
		Description: "Delete a contact file",
		Flags:       fs,
		Run: func(cmd *Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("usage: apeople delete <id> [--confirm]")
			}

			contacts, err := parser.FindContacts(cfg.ContactsDirectory)
			if err != nil {
				return err
			}
			contacts, err = parser.AssignIndexIDs(cfg.ContactsDirectory, contacts)
			if err != nil {
				return err
			}

			contact := parser.FindContactByID(contacts, args[0])
			if contact == nil {
				return fmt.Errorf("contact not found: %s", args[0])
			}

			if !*confirm {
				return fmt.Errorf("use --confirm to delete contact '%s' (%s)", contact.Title, contact.FilePath)
			}

			if err := os.Remove(contact.FilePath); err != nil {
				return fmt.Errorf("failed to delete contact: %w", err)
			}

			if globalFlags.JSON {
				result := map[string]interface{}{
					"deleted":  true,
					"index_id": contact.IndexID,
					"title":    contact.Title,
					"file":     contact.FilePath,
				}
				data, _ := json.MarshalIndent(result, "", "  ")
				fmt.Println(string(data))
				return nil
			}

			if !globalFlags.Quiet {
				fmt.Printf("Deleted %s (#%d)\n", contact.Title, contact.IndexID)
			}
			return nil
		},
	}
}

// sanitizeSlug removes special characters from a filename slug
func sanitizeSlug(name string) string {
	var result strings.Builder
	for _, ch := range name {
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-' {
			result.WriteRune(ch)
		}
	}
	return result.String()
}
