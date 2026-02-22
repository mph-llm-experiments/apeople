package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/mph-llm-experiments/acore"
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
		Description: "Show contact details by index_id or ULID",
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
				type contactWithContent struct {
					*model.Contact
					Content string `json:"content,omitempty"`
				}
				out := contactWithContent{Contact: contact, Content: strings.TrimSpace(contact.Content)}
				data, err := json.MarshalIndent(out, "", "  ")
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

			if contact.Created != "" {
				fmt.Printf("  Created:        %s\n", formatDate(contact.Created))
			}
			if contact.Modified != "" {
				fmt.Printf("  Updated:        %s\n", formatDate(contact.Modified))
			}

			var tagStrs []string
			for _, t := range contact.Tags {
				if t != "contact" {
					tagStrs = append(tagStrs, "#"+t)
				}
			}
			if len(tagStrs) > 0 {
				fmt.Printf("\n  Tags: %s\n", strings.Join(tagStrs, " "))
			}

			if len(contact.RelatedPeople) > 0 || len(contact.RelatedTasks) > 0 || len(contact.RelatedIdeas) > 0 {
				fmt.Println()
				if len(contact.RelatedPeople) > 0 {
					fmt.Printf("  Related people: %s\n", strings.Join(contact.RelatedPeople, ", "))
				}
				if len(contact.RelatedTasks) > 0 {
					fmt.Printf("  Related tasks:  %s\n", strings.Join(contact.RelatedTasks, ", "))
				}
				if len(contact.RelatedIdeas) > 0 {
					fmt.Printf("  Related ideas:  %s\n", strings.Join(contact.RelatedIdeas, ", "))
				}
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

			// Create contact with acore identity
			contact := parser.NewContact(name, cfg.ContactsDirectory)

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
			contact.Tags = contactTags

			// Set domain fields
			contact.RelationshipType = model.RelationshipType(*relType)
			contact.ContactStyle = model.ContactStyle(*style)
			contact.State = *state
			contact.Email = *email
			contact.Phone = *phone
			contact.Company = *company
			contact.Role = *role
			contact.Location = *location

			// Get index_id
			counter, err := acore.NewIndexCounter(cfg.ContactsDirectory, "apeople")
			if err != nil {
				return fmt.Errorf("failed to get ID counter: %w", err)
			}
			id, err := counter.Next()
			if err != nil {
				return fmt.Errorf("failed to get next ID: %w", err)
			}
			contact.IndexID = id

			// Generate file path
			contact.FilePath = parser.GenerateFilePath(cfg.ContactsDirectory, contact)

			if err := parser.SaveContactFile(contact); err != nil {
				return fmt.Errorf("failed to create contact: %w", err)
			}

			if globalFlags.JSON {
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
	addTag := fs.String("add-tag", "", "Add a tag (preserves existing tags)")
	removeTag := fs.String("remove-tag", "", "Remove a tag")
	state := fs.String("state", "", "Update state")
	location := fs.String("location", "", "Update location")

	// Cross-app relationship flags
	addPerson := fs.String("add-person", "", "Add related contact (ULID)")
	removePerson := fs.String("remove-person", "", "Remove related contact (ULID)")
	addTask := fs.String("add-task", "", "Add related task (ULID)")
	removeTask := fs.String("remove-task", "", "Remove related task (ULID)")
	addIdea := fs.String("add-idea", "", "Add related idea (ULID)")
	removeIdea := fs.String("remove-idea", "", "Remove related idea (ULID)")

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
			if *addTag != "" {
				tag := strings.TrimSpace(*addTag)
				if tag != "" && tag != "contact" {
					acore.AddRelation(&contact.Tags, tag)
				}
			}
			if *removeTag != "" {
				tag := strings.TrimSpace(*removeTag)
				if tag != "contact" {
					acore.RemoveRelation(&contact.Tags, tag)
				}
			}

			// Apply cross-app relationship updates
			if *addPerson != "" {
				acore.AddRelation(&contact.RelatedPeople, *addPerson)
				acore.SyncRelation(contact.Type, contact.ID, *addPerson)
			}
			if *removePerson != "" {
				acore.RemoveRelation(&contact.RelatedPeople, *removePerson)
				acore.UnsyncRelation(contact.Type, contact.ID, *removePerson)
			}
			if *addTask != "" {
				acore.AddRelation(&contact.RelatedTasks, *addTask)
				acore.SyncRelation(contact.Type, contact.ID, *addTask)
			}
			if *removeTask != "" {
				acore.RemoveRelation(&contact.RelatedTasks, *removeTask)
				acore.UnsyncRelation(contact.Type, contact.ID, *removeTask)
			}
			if *addIdea != "" {
				acore.AddRelation(&contact.RelatedIdeas, *addIdea)
				acore.SyncRelation(contact.Type, contact.ID, *addIdea)
			}
			if *removeIdea != "" {
				acore.RemoveRelation(&contact.RelatedIdeas, *removeIdea)
				acore.UnsyncRelation(contact.Type, contact.ID, *removeIdea)
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
					msg += fmt.Sprintf(" [state -> %s]", *state)
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

func migrateCommand(cfg *config.Config) *Command {
	fs := flag.NewFlagSet("migrate", flag.ContinueOnError)
	applyMap := fs.String("apply-map", "", "Apply a migration map from another app")

	return &Command{
		Name:        "migrate",
		Usage:       "apeople migrate [--apply-map <path>]",
		Description: "Migrate contacts from Denote format to acore format",
		Flags:       fs,
		Run: func(cmd *Command, args []string) error {
			if *applyMap != "" {
				// Apply external mapping
				migMap, err := acore.ReadMigrationMap(*applyMap)
				if err != nil {
					return fmt.Errorf("failed to read migration map: %w", err)
				}

				if err := acore.ApplyMappings(cfg.ContactsDirectory, migMap.Mappings); err != nil {
					return fmt.Errorf("failed to apply mappings: %w", err)
				}

				if !globalFlags.Quiet {
					fmt.Printf("Applied %d mappings from %s\n", len(migMap.Mappings), migMap.App)
				}
				return nil
			}

			// Migrate this app's files
			migMap, err := acore.MigrateDirectory(cfg.ContactsDirectory, "contact", "apeople")
			if err != nil {
				return fmt.Errorf("migration failed: %w", err)
			}

			if len(migMap.Mappings) == 0 {
				if !globalFlags.Quiet {
					fmt.Println("No files to migrate.")
				}
				return nil
			}

			// Initialize the index counter from migrated files
			counter, err := acore.NewIndexCounter(cfg.ContactsDirectory, "apeople")
			if err != nil {
				return fmt.Errorf("failed to create counter: %w", err)
			}
			readIndexID := func(path string) (int, error) {
				var entity struct {
					acore.Entity `yaml:",inline"`
				}
				if _, err := acore.ReadFile(path, &entity); err != nil {
					return 0, err
				}
				return entity.IndexID, nil
			}
			if err := counter.InitFromFiles("contact", readIndexID); err != nil {
				return fmt.Errorf("counter init: %w", err)
			}

			// Write mapping file
			mapPath := cfg.ContactsDirectory + "/migration-map.json"
			if err := acore.WriteMigrationMap(mapPath, migMap); err != nil {
				return fmt.Errorf("failed to write migration map: %w", err)
			}

			if globalFlags.JSON {
				data, _ := json.MarshalIndent(migMap, "", "  ")
				fmt.Println(string(data))
				return nil
			}

			if !globalFlags.Quiet {
				fmt.Printf("Migrated %d contacts. Mapping saved to %s\n", len(migMap.Mappings), mapPath)
				fmt.Println("Run 'atask migrate --apply-map " + mapPath + "' and 'anote migrate --apply-map " + mapPath + "' to update cross-references.")
			}
			return nil
		},
	}
}

// formatDate formats an RFC 3339 timestamp string as YYYY-MM-DD for display.
func formatDate(rfc3339 string) string {
	t, err := time.Parse(time.RFC3339, rfc3339)
	if err != nil {
		return rfc3339
	}
	return t.Format("2006-01-02")
}
