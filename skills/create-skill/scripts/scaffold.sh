#!/usr/bin/env bash
set -euo pipefail

HELP_MESSAGE="Usage: scaffold.sh --name <skill-name> --description \"<description>\"

Creates a Maki skill skeleton in skills/<skill-name>/ with:
  - SKILL.md (frontmatter + section headers)
  - scripts/
  - references/
  - assets/

Options:
  --name <name>        Skill name (kebab-case, 1-64 chars, lowercase, hyphens only)
  --description <str>  Trigger-optimized description (required)
  --force              Overwrite existing directory
  --help               Show this help message

Example:
  scaffold.sh --name my-skill --description \"Does X and Y for Z\"
  scaffold.sh --name my-skill --description \"Does X\" --force
"

NAME=""
DESCRIPTION=""
FORCE=false

while [[ $# -gt 0 ]]; do
  case "$1" in
    --name)
      NAME="$2"
      shift 2
      ;;
    --description)
      DESCRIPTION="$2"
      shift 2
      ;;
    --force)
      FORCE=true
      shift
      ;;
    --help)
      echo "$HELP_MESSAGE"
      exit 0
      ;;
    *)
      echo "Unknown option: $1" >&2
      echo "$HELP_MESSAGE" >&2
      exit 1
      ;;
  esac
done

# Validate --name
if [[ -z "$NAME" ]]; then
  echo "Error: --name is required" >&2
  echo "$HELP_MESSAGE" >&2
  exit 1
fi

if ! echo "$NAME" | grep -qE '^[a-z][a-z0-9-]{0,63}$'; then
  echo "Error: name must be 1-64 characters, lowercase, hyphens only, no consecutive hyphens" >&2
  exit 1
fi

if echo "$NAME" | grep -q '\-\-'; then
  echo "Error: name must not contain consecutive hyphens" >&2
  exit 1
fi

# Validate --description
if [[ -z "$DESCRIPTION" ]]; then
  echo "Error: --description is required" >&2
  echo "$HELP_MESSAGE" >&2
  exit 1
fi

TARGET_DIR="skills/$NAME"

if [[ -d "$TARGET_DIR" ]]; then
  if [[ "$FORCE" == "false" ]]; then
    echo "Error: $TARGET_DIR already exists. Use --force to overwrite." >&2
    exit 1
  fi
  rm -rf "$TARGET_DIR"
fi

mkdir -p "$TARGET_DIR/scripts" "$TARGET_DIR/references" "$TARGET_DIR/assets"

cat > "$TARGET_DIR/SKILL.md" <<EOF
---
name: $NAME
description: $DESCRIPTION
compatibility: opencode
---

## Purpose

<!-- Describe what this skill does, when to use it, and when NOT to use it. -->

## Content

<!-- Main instructions go here. Use step-by-step numbering, explicit decision trees,
third-person imperative, and consistent terminology. Keep under 500 lines. -->

## Scripts

<!-- List each script in scripts/ with its purpose and usage examples. -->

## References

<!-- List each file in references/ with what it contains and when to read it. -->

## Assets

<!-- List each file in assets/ with what it contains and when to use it. -->

## Edge Cases

<!-- Document failure modes, error recovery steps, and handling for unusual inputs. -->
EOF

echo "Created skill skeleton at $TARGET_DIR/"
echo "  SKILL.md     — Core instructions (edit this)"
echo "  scripts/     — Executable helpers"
echo "  references/  — Supplementary context"
echo "  assets/      — Templates and static files"
echo "  examples/    — Code examples"
echo ""
echo "Next: edit $TARGET_DIR/SKILL.md with your instructions."
