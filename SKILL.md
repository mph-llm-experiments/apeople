---
name: apeople
description: Agent-first contacts management using the apeople CLI. Use when managing contacts, logging interactions, reviewing overdue contacts, or organizing relationship data. NOT for tasks (use atask) or ideas (use anote).
---

# apeople — Agent-First Contacts Management

Manage contacts using the apeople CLI tool. Contacts are people you want to keep in touch with, organized by relationship type and contact frequency. This is a sibling tool to atask (tasks) and anote (ideas).

## When to Use apeople vs atask vs anote

| Use apeople when... | Use atask when... | Use anote when... |
|---|---|---|
| Managing contact information | Creating actionable work items | Capturing ideas |
| Logging interactions with people | Tracking task completion | Exploring concepts |
| Reviewing overdue contacts | Managing project deliverables | Managing idea maturity |
| Organizing relationships | Setting deadlines | Connecting related ideas |

## Quick Decision Tree for Agents

```
├─ 📖 READ/FIND CONTACTS
│  ├─ List all? → apeople list --json
│  ├─ Filter by type? → apeople list --type close --json
│  ├─ Find overdue? → apeople list --overdue --json
│  ├─ Search? → apeople list --search "term" --json
│  └─ Show one? → apeople show <id> --json
│
├─ ✏️ CREATE CONTACTS
│  └─ New contact → apeople new "Name" --type <type> [--options]
│
├─ 🔄 UPDATE CONTACTS
│  ├─ Update fields → apeople update <id> --email new@example.com
│  ├─ Log interaction → apeople log <id> --interaction email
│  └─ Bump (review) → apeople bump <id>
│
└─ 🗑️ DELETE
   └─ Delete → apeople delete <id> --confirm
```

**Golden Rule:** If the output will be processed → add `--json` to the command!

## Command Reference

### List Contacts

```bash
apeople list --json                              # All contacts (JSON)
apeople list --type close --json                  # Filter by relationship type
apeople list --overdue --json                     # Only overdue contacts
apeople list --state followup --json              # Filter by state
apeople list --style periodic --json              # Filter by contact style
apeople list --search "portland" --json           # Search name/company/email/tags
apeople list --sort days --json                   # Sort by days since contact
apeople list --all --json                         # Include archived contacts
```

**Filter options:**
- `--type`: close, family, network, work, social, providers, recruiters
- `--state`: ok, active, followup, ping, archived
- `--style`: periodic, ambient, triggered
- `--sort`: name (default), days, type, state

### Show Contact

```bash
apeople show 1 --json          # By index_id (preferred)
apeople show 20240715T093045   # By denote identifier
```

### Create Contact

```bash
apeople new "Sarah Chen" --type close --email sarah@example.com --company "Acme Corp"
apeople new "John Smith" --type work --tags "portland,tech" --style periodic
```

**Options:**
- `--type`: Relationship type (default: network)
- `--style`: Contact style (default: periodic)
- `--email`, `--phone`, `--company`, `--role`, `--location`
- `--tags`: Comma-separated tags (in addition to 'contact')
- `--state`: Initial state (default: ok)

### Update Contact

```bash
apeople update 1 --email new@example.com
apeople update 1 --type close --state followup
apeople update 1 --add-tag portland              # Add a tag (preserves existing)
apeople update 1 --remove-tag portland           # Remove a tag
apeople update 1 --tags "tech,portland,friend"   # Replace all non-contact tags (use with care)
```

### Cross-App Relationships

Link contacts to tasks (atask), ideas (anote), or other contacts (apeople) using Denote IDs:

```bash
# Add relationships
apeople update 1 --add-task 20250610T141230
apeople update 1 --add-idea 20250607T093045
apeople update 1 --add-person 20250612T080000

# Remove relationships
apeople update 1 --remove-task 20250610T141230
apeople update 1 --remove-idea 20250607T093045
apeople update 1 --remove-person 20250612T080000
```

Relationships are stored in YAML frontmatter as arrays of Denote IDs:
- `related_people` — linked contacts
- `related_tasks` — linked tasks/projects from atask
- `related_ideas` — linked ideas from anote

Relationships are NOT automatically bidirectional. To link both directions, update both entities.

### Log Interaction

```bash
apeople log 1 --interaction email                           # Log email contact
apeople log 1 --interaction call --state followup           # Log + change state
apeople log 1 --interaction meeting --note "Quarterly sync" # Log with note
```

**Interaction types:** email, call, text, meeting, social, bump, note

### Bump Contact

```bash
apeople bump 1    # Review without actually contacting
```

A bump updates `last_bump_date` but NOT `last_contacted`. Use for reviewing a contact's info without reaching out.

### Delete Contact

```bash
apeople delete 1 --confirm    # --confirm is required
```

## JSON Output Structure

### Contact Object

```json
{
  "title": "Sarah Chen",
  "date": "2024-07-15T09:30:45Z",
  "tags": ["contact", "tech", "portland"],
  "identifier": "20240715T093045",
  "index_id": 1,
  "email": "sarah@example.com",
  "phone": "555-0123",
  "relationship_type": "close",
  "state": "ok",
  "contact_style": "periodic",
  "last_contacted": "2024-08-01T10:00:00Z",
  "last_bump_date": "2024-08-10T14:00:00Z",
  "bump_count": 2,
  "updated_at": "2024-08-10T14:00:00Z",
  "company": "Acme Corp",
  "role": "Senior Engineer",
  "location": "Portland, OR",
  "related_people": [],
  "related_tasks": ["20250610T141230"],
  "related_ideas": [],
  "file_path": "/path/to/20240715T093045--sarah-chen__contact.md",
  "days_since_contact": 14,
  "overdue_status": "good"
}
```

### Key Fields for Agents

- `index_id`: Stable numeric ID for CLI commands (preferred for referencing)
- `identifier`: Denote timestamp ID (also works for referencing)
- `days_since_contact`: -1 if never contacted, otherwise days since last contact
- `overdue_status`: "overdue", "attention", "good", or empty
- `related_people`, `related_tasks`, `related_ideas`: Arrays of Denote IDs linking to other entities (always `[]`, never null)

## Global Options

```bash
--json         # JSON output (always use for programmatic access)
--dir PATH     # Override contacts directory
--config PATH  # Use specific config file
--quiet, -q    # Minimal output
--no-color     # Disable color output
```

## Agent Workflow Patterns

### Contact Review Workflow

```bash
# 1. Find overdue contacts
apeople list --overdue --json

# 2. Review each overdue contact
apeople show <id> --json

# 3. Either log an interaction or bump
apeople log <id> --interaction email --note "Checked in"
# or
apeople bump <id>
```

### Finding Contacts That Need Attention

```bash
# Overdue periodic contacts
apeople list --overdue --style periodic --json

# Contacts in followup state
apeople list --state followup --json

# Search for someone
apeople list --search "acme" --json
```

### Cross-App Workflow: Log Interaction and Create Follow-Up Task

```bash
# 1. Log the interaction
apeople log 5 --interaction call --note "Discussed proposal timeline"

# 2. Create a follow-up task in atask
atask new "Follow up on Sarah's proposal" --due "next friday" --json
# Note the denote_id from the output

# 3. Link the contact to the task (both directions)
apeople update 5 --add-task <task-denote-id>
atask update <task-index-id> --add-person 20240715T093045
```

### Cross-App Workflow: Find Everything Related to a Person

```bash
# 1. Get the contact's relationships
apeople show 5 --json | jq '{tasks: .related_tasks, ideas: .related_ideas}'

# 2. Look up each linked task
atask list --json | jq '[.tasks[] | select(.denote_id == "20250610T141230")]'

# 3. Look up each linked idea
anote --json list | jq '[.[] | select(.denote_id == "20250607T093045")]'
```

### Creating a Contact from Conversation

```bash
apeople new "Jane Doe" \
  --type work \
  --email jane@company.com \
  --company "Company Inc" \
  --role "VP Engineering" \
  --tags "portland,hiring" \
  --location "Portland, OR"
```

## Contact Types & Frequencies

| Type | Default Frequency | Description |
|------|-------------------|-------------|
| close | 30 days | Close friends/contacts |
| family | 30 days | Family members |
| work | 60 days | Work colleagues |
| network | 90 days | Professional network |
| social | No default | Social acquaintances |
| providers | No default | Service providers |
| recruiters | No default | Recruiters |

## Contact Styles

- **periodic**: Regular check-ins based on frequency (generates overdue alerts)
- **ambient**: Passive monitoring, no active reminders
- **triggered**: Event-based contact only

## Interaction Log Format

The `log` command appends entries to an `## Interaction Log` section in the contact's markdown content. The format is:

```markdown
## Interaction Log

- **2026-02-18** (email) - Discussed project timeline
- **2026-02-15** (call)
- **2026-01-20** (meeting) - Quarterly sync
```

- Most recent entries appear first
- The `--note` text appears after the interaction type, separated by ` - `
- If no note is provided, only the date and type are recorded
- If no `## Interaction Log` section exists, one is created automatically
- The `log` command also updates `last_contacted` and `last_interaction_type` in frontmatter

## Best Practices

1. Always use `--json` when processing output programmatically
2. Use `index_id` (numeric) to reference contacts in commands
3. Log interactions to keep contact timelines accurate
4. Use `bump` for reviewing contacts without actual outreach
5. Set `contact_style: ambient` for contacts that don't need regular check-ins
6. Use tags for grouping (e.g., `#portland`, `#conference`, `#client`)
