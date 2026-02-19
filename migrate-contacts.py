#!/usr/bin/env python3
"""
Migrate contacts from basic-memory format to apeople Denote format.

Reads from ~/basic-memory/people/ (subdirectories by relationship type)
Writes to ~/basic-memory/people-migrated/ for review before deploying.

Two source patterns handled:
  1. Double frontmatter: false basic-memory block, then real contact block
  2. Merged frontmatter: basic-memory fields mixed into single block

Output: YYYYMMDDTHHMMSS--kebab-case-name__contact.md with clean apeople frontmatter
"""

import os
import re
import sys
import stat
from datetime import datetime, timezone
from pathlib import Path

SOURCE_DIR = Path.home() / "basic-memory" / "people"
DEST_DIR = Path.home() / "basic-memory" / "people-migrated"

# Files to skip (not contacts)
SKIP_FILES = {
    "mike-hall.md",
    "Nathan - V2MOM Coaching.md",
}

# Basic-memory junk fields to strip
JUNK_FIELDS = {
    "metadata",
    "type",
    "permalink",
    "basic_memory_url",
}

# State mapping
STATE_MAP = {
    "write": "followup",
    "ok": "ok",
    "followup": "followup",
    "ping": "ping",
}

# Contact style mapping
STYLE_MAP = {
    "professional": "periodic",
    "ambient": "ambient",
    "triggered": "triggered",
    "periodic": "periodic",
}


def get_creation_time(filepath):
    """Get file creation time (birthtime on macOS, mtime fallback)."""
    st = os.stat(filepath)
    # macOS has st_birthtime
    birthtime = getattr(st, "st_birthtime", None)
    if birthtime:
        return datetime.fromtimestamp(birthtime)
    return datetime.fromtimestamp(st.st_mtime)


def parse_source_file(filepath):
    """Parse a basic-memory contact file, handling both mangling patterns."""
    with open(filepath, "r") as f:
        raw = f.read()

    # Split on --- delimiters
    parts = raw.split("---\n")

    # Pattern 1: Double frontmatter (5+ parts: empty, junk_fm, empty, real_fm, content)
    # Pattern 2: Merged frontmatter (3 parts: empty, merged_fm, content)
    # Pattern 3: No frontmatter at all (skip)

    frontmatter_lines = []
    content = ""

    if len(parts) >= 5:
        # Double frontmatter: parts[1] is junk, parts[3] is real
        frontmatter_lines = parts[3].strip().split("\n")
        content = "---\n".join(parts[4:]) if len(parts) > 4 else ""
    elif len(parts) >= 3:
        # Single (possibly merged) frontmatter
        frontmatter_lines = parts[1].strip().split("\n")
        content = "---\n".join(parts[2:]) if len(parts) > 2 else ""
    else:
        return None  # No frontmatter

    # Parse frontmatter into dict, handling multiline values (tags)
    data = {}
    current_key = None
    list_values = []

    for line in frontmatter_lines:
        if not line.strip():
            continue

        # Check for list continuation
        list_match = re.match(r"^\s+-\s+(.*)", line)
        if list_match and current_key:
            list_values.append(list_match.group(1).strip().strip("'\""))
            data[current_key] = list_values
            continue

        # Key: value line
        kv_match = re.match(r"^(\w[\w_]*)\s*:\s*(.*)", line)
        if kv_match:
            # Save previous list if any
            current_key = kv_match.group(1)
            value = kv_match.group(2).strip()

            if value == "":
                # Could be start of a list
                list_values = []
            else:
                data[current_key] = value.strip("'\"")
                list_values = []
        else:
            current_key = None

    return data, content.strip()


def clean_company(company):
    """Clean company field: strip semicolons, unescape commas."""
    if not company:
        return ""
    # Strip trailing semicolons
    company = company.rstrip(";").strip()
    # Unescape commas
    company = company.replace("\\,", ",")
    # If it's just a semicolon or empty after cleanup
    if company in (";", ""):
        return ""
    return company


def sanitize_name(name):
    """Convert name to kebab-case slug."""
    slug = name.lower()
    slug = slug.replace(" ", "-")
    # Remove special characters, keep only alphanumeric and hyphens
    slug = re.sub(r"[^a-z0-9-]", "", slug)
    # Collapse multiple hyphens
    slug = re.sub(r"-+", "-", slug)
    return slug.strip("-")


def build_apeople_frontmatter(data, directory_type, filepath):
    """Build clean apeople YAML frontmatter from parsed data."""
    name = data.get("name", "")
    if not name:
        return None

    creation_time = get_creation_time(filepath)
    identifier = creation_time.strftime("%Y%m%dT%H%M%S")

    # Relationship type from field or directory
    rel_type = data.get("relationship", directory_type)

    # Contact style
    raw_style = data.get("contact_style", "")
    contact_style = STYLE_MAP.get(raw_style, "periodic")

    # State
    raw_state = data.get("state", "ok")
    state = STATE_MAP.get(raw_state, "ok")

    # Archived overrides state
    archived = data.get("archived", "false")
    if archived in ("true", "True", "'True'"):
        state = "archived"

    # Extract label from tags (@handle)
    label = ""
    raw_tags = data.get("tags", [])
    if isinstance(raw_tags, str):
        raw_tags = [raw_tags]
    for tag in raw_tags:
        if tag.startswith("@"):
            label = tag
            break

    # Build tags list: always include contact, add relationship type
    tags = ["contact"]
    # Add meaningful tags (skip @handles, 'person', 'relationship')
    skip_tags = {"person", "relationship"}
    for tag in raw_tags:
        if tag.startswith("@"):
            continue
        if tag in skip_tags:
            continue
        if tag and tag not in tags:
            tags.append(tag)

    # Build frontmatter
    lines = []
    lines.append(f"title: {name}")
    lines.append(f"date: {creation_time.strftime('%Y-%m-%dT%H:%M:%S')}Z")
    lines.append(f"tags: [{', '.join(tags)}]")
    lines.append(f"identifier: {identifier}")

    if data.get("email"):
        lines.append(f"email: {data['email']}")
    if data.get("phone"):
        phone = data["phone"].strip("'\"")
        lines.append(f"phone: \"{phone}\"")

    lines.append(f"relationship_type: {rel_type}")
    lines.append(f"contact_style: {contact_style}")
    lines.append(f"state: {state}")

    if label:
        lines.append(f"label: \"{label}\"")

    company = clean_company(data.get("company", ""))
    if company:
        lines.append(f"company: \"{company}\"")

    # Last contacted
    last_contact = data.get("last_contact", "")
    if last_contact:
        last_contact = last_contact.strip("'\"")
        # Parse the date and add time component
        try:
            dt = datetime.strptime(last_contact, "%Y-%m-%d")
            lines.append(f"last_contacted: {dt.strftime('%Y-%m-%dT%H:%M:%S')}Z")
        except ValueError:
            pass

    lines.append(f"updated_at: {datetime.now().strftime('%Y-%m-%dT%H:%M:%S')}Z")

    return identifier, name, "\n".join(lines)


def migrate_file(filepath, directory_type):
    """Migrate a single contact file. Returns (dest_filename, content) or None."""
    result = parse_source_file(filepath)
    if result is None:
        return None

    data, content = result

    # Must have a name field to be a contact
    if "name" not in data:
        return None

    fm_result = build_apeople_frontmatter(data, directory_type, filepath)
    if fm_result is None:
        return None

    identifier, name, frontmatter = fm_result
    slug = sanitize_name(name)
    dest_filename = f"{identifier}--{slug}__contact.md"

    # Build final file content
    output = f"---\n{frontmatter}\n---\n"
    if content:
        output += f"\n{content}\n"

    return dest_filename, output


def main():
    if not SOURCE_DIR.exists():
        print(f"Source directory not found: {SOURCE_DIR}")
        sys.exit(1)

    DEST_DIR.mkdir(parents=True, exist_ok=True)

    migrated = 0
    skipped = 0
    errors = []
    seen_filenames = {}

    # Walk all subdirectories
    for item in sorted(SOURCE_DIR.iterdir()):
        if item.is_file() and item.suffix == ".md":
            # Root-level files
            if item.name in SKIP_FILES:
                print(f"  SKIP (non-contact): {item.name}")
                skipped += 1
                continue

            result = migrate_file(item, "")
            if result:
                dest_filename, content = result
                # Handle filename collisions
                if dest_filename in seen_filenames:
                    # Append a counter
                    base = dest_filename.replace("__contact.md", "")
                    counter = 2
                    while f"{base}-{counter}__contact.md" in seen_filenames:
                        counter += 1
                    dest_filename = f"{base}-{counter}__contact.md"
                seen_filenames[dest_filename] = item
                dest_path = DEST_DIR / dest_filename
                with open(dest_path, "w") as f:
                    f.write(content)
                print(f"  OK: {item.name} -> {dest_filename}")
                migrated += 1
            else:
                print(f"  SKIP (no name): {item.name}")
                skipped += 1

        elif item.is_dir():
            directory_type = item.name
            print(f"\n=== {directory_type}/ ===")
            for md_file in sorted(item.glob("*.md")):
                if md_file.name in SKIP_FILES:
                    print(f"  SKIP (non-contact): {md_file.name}")
                    skipped += 1
                    continue

                try:
                    result = migrate_file(md_file, directory_type)
                    if result:
                        dest_filename, content = result
                        # Handle filename collisions
                        if dest_filename in seen_filenames:
                            base = dest_filename.replace("__contact.md", "")
                            counter = 2
                            while f"{base}-{counter}__contact.md" in seen_filenames:
                                counter += 1
                            dest_filename = f"{base}-{counter}__contact.md"
                        seen_filenames[dest_filename] = md_file
                        dest_path = DEST_DIR / dest_filename
                        with open(dest_path, "w") as f:
                            f.write(content)
                        print(f"  OK: {md_file.name} -> {dest_filename}")
                        migrated += 1
                    else:
                        print(f"  SKIP (parse failed): {md_file.name}")
                        skipped += 1
                except Exception as e:
                    print(f"  ERROR: {md_file.name}: {e}")
                    errors.append((md_file.name, str(e)))

    print(f"\n{'='*60}")
    print(f"Migrated: {migrated}")
    print(f"Skipped:  {skipped}")
    print(f"Errors:   {len(errors)}")
    print(f"Output:   {DEST_DIR}")

    if errors:
        print("\nErrors:")
        for name, err in errors:
            print(f"  {name}: {err}")

    print(f"\nReview the output, then copy to your contacts directory:")
    print(f"  cp {DEST_DIR}/*.md <your-contacts-dir>/")


if __name__ == "__main__":
    main()
