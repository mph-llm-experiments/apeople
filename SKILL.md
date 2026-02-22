---
name: apeople
description: Agent-first contacts management using the apeople CLI. Use when managing contacts, logging interactions, reviewing overdue contacts, or organizing relationship data. NOT for tasks (use atask) or ideas (use anote).
---

# apeople -- Contacts Management

Manage contacts using the apeople CLI. Contacts are people you want to keep in touch with, organized by relationship type and contact frequency. Sibling tools: atask (tasks) and anote (ideas).

All data is stored as plain markdown files with YAML frontmatter. Filename format: `{ulid}--{slug}__contact.md`

## Commands

### list -- List contacts

```bash
apeople list [options] --json
```

Default: excludes archived contacts.

Options:
- `--all` -- Include archived contacts
- `--overdue` -- Show only overdue contacts
- `--type` -- Filter by relationship type: close, family, network, work, social, providers, recruiters
- `--state` -- Filter by state: ok, active, followup, ping, archived
- `--style` -- Filter by contact style: periodic, ambient, triggered
- `--search` -- Search by name, company, email, or tags
- `--sort` -- Sort by: name (default), days, type, state

### show -- Show contact details

```bash
apeople show <index_id_or_ulid> --json
```

Accepts index_id (numeric) or ULID.

### new -- Create a contact

```bash
apeople new "Name" [options]
```

Options:
- `--type` -- Relationship type (default: network)
- `--style` -- Contact style (default: periodic)
- `--state` -- Initial state (default: ok)
- `--email`, `--phone`, `--company`, `--role`, `--location`
- `--tags` -- Comma-separated tags (in addition to 'contact')

### update -- Update contact fields

```bash
apeople update <id> [options]
```

Options:
- `--name` -- Update name
- `--email`, `--phone`, `--company`, `--role`, `--location`
- `--type` -- Update relationship type
- `--state` -- Update state
- `--style` -- Update contact style
- `--tags` -- Replace all non-contact tags (comma-separated)
- `--add-tag <tag>` -- Add a tag (preserves existing)
- `--remove-tag <tag>` -- Remove a tag

Cross-app relationship flags (values are ULIDs):
- `--add-person <ulid>` / `--remove-person <ulid>`
- `--add-task <ulid>` / `--remove-task <ulid>`
- `--add-idea <ulid>` / `--remove-idea <ulid>`

### log -- Log an interaction

```bash
apeople log <id> --interaction <type> [--note "text"] [--state <new-state>]
```

Interaction types: email, call, text, meeting, social, bump, note

Updates `last_contacted` in frontmatter. Appends to an `## Interaction Log` section in the file body (most recent first).

### bump -- Review without contacting

```bash
apeople bump <id>
```

Updates `last_bump_date` but NOT `last_contacted`. Use for reviewing a contact's info without reaching out.

### delete -- Delete a contact

```bash
apeople delete <id> --confirm
```

`--confirm` is required.

## JSON Structure

```json
{
  "id": "01KJ1KHY3XGRPSBE9ZKJYDDKVT",
  "title": "Sarah Chen",
  "index_id": 1,
  "type": "contact",
  "tags": ["contact", "tech"],
  "created": "2026-01-31T10:05:26Z",
  "modified": "2026-02-22T02:41:14Z",
  "related_people": [],
  "related_tasks": ["01KJ1KJ3VFJFNDH5K6VEDS2G6G"],
  "related_ideas": [],
  "file_path": "/path/to/01KJ1KHY3X...--sarah-chen__contact.md",
  "email": "sarah@example.com",
  "phone": "555-0123",
  "relationship_type": "close",
  "state": "ok",
  "label": "@sarahc",
  "contact_style": "periodic",
  "last_contacted": "2026-02-01T10:00:00Z",
  "last_bump_date": "2026-02-10T14:00:00Z",
  "bump_count": 2,
  "company": "Acme Corp",
  "days_since_contact": 14
}
```

Key fields:
- `id` -- ULID, the canonical identifier
- `index_id` -- stable numeric ID for CLI commands
- `label` -- short handle (e.g. `@sarahc`)
- `days_since_contact` -- -1 if never contacted, otherwise days since last contact
- `related_people`, `related_tasks`, `related_ideas` -- arrays of ULIDs (always `[]`, never null)

## Contact Types and Frequencies

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

## Contact States

ok, active, followup, ping, archived

## Agent Workflows

### Contact review

```bash
apeople list --overdue --json
apeople show <id> --json
apeople log <id> --interaction email --note "Checked in"
# or just review without contacting:
apeople bump <id>
```

### Cross-app: log interaction and create follow-up task

```bash
apeople log 5 --interaction call --note "Discussed proposal timeline"
atask new "Follow up on Sarah's proposal" --due "next friday" --json
# Parse id (ULID) from atask output, then link both directions:
apeople update 5 --add-task <task-ulid>
atask update <task-index-id> --add-person <contact-ulid>
```

### Find everything related to a person

```bash
apeople show 5 --json
# Parse related_tasks and related_ideas arrays, then look up each:
atask show <ulid> --json
anote show <ulid> --json
```

## Configuration

Config: `~/.config/acore/config.toml`

```toml
[directories]
apeople = "/path/to/contacts"
```

Override with `--dir` flag. Also supports `--config` for alternate config file.

## Global Options

```
--json         JSON output (always use for programmatic access)
--dir PATH     Override contacts directory
--config PATH  Use specific config file
--quiet, -q    Minimal output
--no-color     Disable color output
```
