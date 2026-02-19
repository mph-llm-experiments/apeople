# CLAUDE.md - Project Context for apeople

This file contains important context about the apeople project to help AI assistants understand the codebase, architecture decisions, and current state.

## Project Overview

**apeople** is an agent-first contacts management tool built on the Denote file naming convention. It provides CLI commands for agent use and a TUI as the default (no-args) mode. Part of the atask/anote/apeople ecosystem.

## Important Documents

- `/docs/DENOTE_CONTACTS_SPEC.md` - Contact file format specification
- `/docs/TUI_SPECIFICATION.md` - UI patterns
- `/docs/CONTACTS_TUI_ARCHITECTURE.md` - Technical architecture
- `/SKILL.md` - Agent-facing command reference

## Denote File Format

Contact files follow this naming pattern:
- New files: `YYYYMMDDTHHMMSS--kebab-case-name__contact.md`
- Legacy files: `YYYYMMDD--kebab-case-name__contact.md` (still supported for parsing)
- Example: `20250607T093045--sarah-chen__contact.md`
- Double underscore (`__`) appears before the file type tag
- Filename uses single tag `contact` to identify file type
- Additional tags go in the YAML frontmatter, not the filename

## CLI Architecture

apeople follows the same CLI pattern as atask and anote:

- No args = launch TUI
- Subcommands: `list`, `show`, `new`, `update`, `log`, `bump`, `delete`
- Global flags: `--json`, `--dir`, `--config`, `--quiet`, `--no-color`
- `--json` output for all read commands
- `index_id` system for stable numeric references

### CLI Command Reference

```bash
apeople list [--type X] [--state X] [--overdue] [--search X] [--sort X] [--json]
apeople show <id> [--json]
apeople new "Name" [--type X] [--style X] [--email X] [--company X] [--tags X]
apeople update <id> [--name X] [--type X] [--state X] [--tags X] ...
apeople log <id> --interaction <type> [--state X] [--note X]
apeople bump <id>
apeople delete <id> --confirm
```

## Tag System

### Filename Tags
- Contact files MUST use `__contact` in the filename
- Only ONE tag in the filename (the file type identifier)

### Frontmatter Tags
- The `tags` array in YAML MUST include `contact`
- Additional tags for organization: `tags: [contact, personal, tech, portland]`

## Architecture Principles

1. **Agent-First** - CLI with JSON output, designed for AI agent workflows
2. **Denote Format** - Use Denote naming for consistent IDs
3. **Contacts Focus** - Only contacts files, no general notes
4. **No External Dependencies** - No TaskWarrior, Things, dstask, or SQLite
5. **No Caching** - Always read files fresh from disk
6. **Shared Code Paths** - TUI and CLI use the same parser functions

## Testing Guidelines

### CRITICAL RULE: NEVER MARK FEATURES AS COMPLETE WITHOUT HUMAN TESTING

Any feature implementation MUST be marked as "IMPLEMENTED BUT NOT TESTED" until the human has confirmed it works.

### For TUI Development

**IMPORTANT:** It is IMPOSSIBLE to test TUI applications in this environment. NEVER attempt to run or test the TUI. Ask the user to test instead.

### CLI Testing

CLI commands CAN be tested in this environment using `--dir /tmp/test-dir` to avoid touching user data.

## Contact Management Features

### Relationship Types & Frequencies
- close: 30 days
- family: 30 days
- network: 90 days
- work: 60 days
- social: No default frequency
- Custom frequencies override defaults

### Contact Styles
- periodic: Regular check-ins based on frequency
- ambient: Passive monitoring, no active reminders
- triggered: Event-based contact only

### Interaction Types
email, call, text, meeting, social, bump, note

### The "Bump" Concept
A bump is reviewing a contact without actually contacting them. It updates last_bump_date but NOT last_contacted.

## Common Pitfalls to Avoid

1. **Don't modify user configs** - Always use test configs or `--dir`
2. **Don't assume TUI works** - It needs terminal testing
3. **Don't add non-contacts features** - This is a contacts management tool only
4. **Don't add caching** - Read files fresh from disk always
5. **Don't stray from focus** - If it's not about contacts, it doesn't belong

## Format Mistakes to Prevent

NEVER use these incorrect formats:
- `20250607--sarah-chen__contact__personal.md` (tags go in frontmatter, not filename)
- `20250607--sarah-chen_contact.md` (must use double underscore)
- `20250607-sarah-chen__contact.md` (must use double dash after date)

ALWAYS use:
- `YYYYMMDDTHHMMSS--kebab-case-name__contact.md` (new format with timestamp)
- Additional tags in frontmatter: `tags: [contact, work, engineering]`

## Performance Philosophy

- Always read files fresh from disk - no caching
- Small markdown files (typically < 200 lines)
- File I/O is negligible compared to user interaction time

## Ecosystem Integration

apeople works with:
- **atask**: Generate follow-up tasks, reference contacts
- **anote**: Link contacts to ideas
- All use Denote naming and `--json` output

## Explicit Non-Goals

- NO task backend integrations (TaskWarrior, Things, dstask)
- NO SQLite or database layer
- NO general notes functionality
- NO caching of any kind
- NO modal dialogs (use inline editing in TUI)

## Denote Parser Requirements

The parser accepts:
1. `YYYYMMDDTHHMMSS` (new format) or `YYYYMMDD` (legacy) for identifiers
2. Double dash (`--`) after the identifier
3. Kebab-case name (lowercase, hyphens for spaces)
4. Double underscore (`__`) before the tag
5. Single tag only in filename (multiple tags in frontmatter)
6. `.md` extension

New files are always created with the `YYYYMMDDTHHMMSS` format.
