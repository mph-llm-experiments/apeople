# apeople

An agent-first contacts management tool using the [Denote](https://protesilaos.com/emacs/denote) file naming convention. CLI commands for agent use, TUI as the default (no-args) mode. Part of the atask/anote/apeople ecosystem.

**Note**: This project is not affiliated with the Denote project. It simply adopts Denote's excellent file naming convention for consistent contact identification.

## Important consideration before using this code or interacting with this codebase

This application is an experiment in using Claude Code as the primary driver the development of a small, focused app that concerns itself with the owner's particular point of view on the task it is accomplishing.

As such, this is not meant to be what people think of as "an open source project," because I don't have a commitment to building a community around it and don't have the bandwidth to maintain it beyond "fix bugs I find in the process of pushing it in a direction that works for me."

It's important to understand this for a few reasons:

1. If you use this code, you'll be using something largely written by an LLM with all the things we know this entails in 2025: Potential inefficiency, security risks, and the risk of data loss.

2. If you use this code, you'll be using something that works for me the way I would like it to work. If it doesn't do what you want it to do, or if it fails in some way particular to your preferred environment, tools, or use cases, your best option is to take advantage of its very liberal license and fork it.

3. I'll make a best effort to only tag the codebase when it is in a working state with no bugs that functional testing has revealed.

While I appreciate and applaud assorted efforts to certify code and projects AI-free, I think it's also helpful to post commentary like this up front: Yes, this was largely written by an LLM so treat it accordingly. Don't think of it like code you can engage with, think of it like someone's take on how to do a task or solve a problem.

## Overview

apeople provides both a CLI and TUI for managing personal and professional contacts. Each contact is stored as a markdown file with YAML frontmatter, using Denote's timestamp-based naming convention for unique identification.

### Key Features

- **Agent-First CLI**: JSON output, subcommands, designed for AI agent workflows
- **TUI Mode**: Full terminal UI when run with no arguments
- **Smart Reminders**: Set contact frequencies based on relationship type
- **Visual Status**: See at a glance who's overdue, due soon, or on track
- **Quick Actions**: Log interactions, bump reviews, edit details
- **Flexible Organization**: Tag and categorize contacts by type, style, and custom tags
- **Task Integration**: Creates tasks in [atask](https://github.com/mph-llm-experiments/atask) when contacts need attention

## Installation

```bash
go install github.com/mph-llm-experiments/apeople@latest
```

Or clone and build:

```bash
git clone https://github.com/mph-llm-experiments/apeople.git
cd apeople
go build
```

## Configuration

apeople uses a TOML configuration file:

```
~/.config/apeople/config.toml
```

### Example Configuration

```toml
# Directory where your contact files are stored
contacts_directory = "~/Documents/denote"
```

### Configuration Priority

1. `--dir` flag (highest priority)
2. `APEOPLE_DIR` environment variable
3. Config file setting `contacts_directory`
4. Legacy config at `~/.config/denote-contacts/config.toml`
5. Default: `~/Documents/denote`

## CLI Usage

```bash
# Launch TUI (default, no arguments)
apeople

# List contacts
apeople list
apeople list --json
apeople list --type close --overdue
apeople list --search "portland" --sort days

# Show contact details
apeople show 1
apeople show 1 --json

# Create a contact
apeople new "Sarah Chen" --type close --email sarah@example.com --company "Acme Corp"

# Update a contact
apeople update 1 --state followup --tags "tech,portland"

# Log an interaction
apeople log 1 --interaction email --note "Discussed project timeline"

# Bump (review without contacting)
apeople bump 1

# Delete a contact
apeople delete 1 --confirm

# Global options
apeople list --dir ~/my-contacts --json --quiet
```

## Contact File Format

Contacts are stored as markdown files with YAML frontmatter:

```yaml
---
title: Jane Smith
identifier: 20240715T093045
index_id: 1
date: 2024-07-15
tags: [contact, work, portland]
email: jane@example.com
phone: 555-0123
company: Acme Corp
role: Senior Engineer
location: Portland, OR
relationship_type: work
contact_style: periodic
state: ok
last_contacted: 2024-07-01T10:30:00Z
---

## Notes

Met at tech conference...
```

### File Naming

Files follow the Denote convention:

```
YYYYMMDDTHHMMSS--kebab-case-name__contact.md
```

Example: `20240715T093045--jane-smith__contact.md`

## TUI Keyboard Controls

### List View

- **Navigation**
  - `j/k` or arrows - Move up/down
  - `g/G` - Go to top/bottom
  - `Ctrl+d/u` - Page down/up

- **Actions**
  - `Enter` - View contact details
  - `d` - Log interaction (contacted)
  - `s` - Quick state change
  - `T` - Quick type change
  - `b` - Bump (mark as reviewed)
  - `e` - Edit contact
  - `c` - Create new contact
  - `/` - Search
  - `f` - Filter
  - `q` - Quit

### Detail View

- `e` - Edit contact
- `d` - Log interaction
- `b` - Bump contact
- `q/Esc` - Back to list

## Contact Types & Default Frequencies

When using `contact_style: periodic`, these defaults apply:

- **close** - 30 days
- **family** - 30 days
- **work** - 60 days
- **network** - 90 days
- **social** - No default
- **providers** - No default
- **recruiters** - No default

Override with `custom_frequency_days` in the frontmatter.

## Contact Styles

- **periodic** - Regular check-ins based on frequency
- **ambient** - Passive monitoring, no reminders
- **triggered** - Event-based contact

## Contact States

- **ok** - Up to date
- **followup** - Need to follow up
- **ping** - Send a quick check-in
- **scheduled** - Meeting/call is scheduled
- **timeout** - No response, needs attention

## Ecosystem

apeople is part of a trio of agent-first tools:

- **[atask](https://github.com/mph-llm-experiments/atask)** - Task management
- **[anote](https://github.com/mph-llm-experiments/anote)** - Idea management
- **apeople** - Contacts management

All use Denote file naming and are designed for AI agent workflows with `--json` output.

## License

MIT
