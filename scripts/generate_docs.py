#!/usr/bin/env python3
"""
generate_docs.py - Auto-generates documentation for dev-stack
from YAML manifests.

- Commands: scripts/commands.yaml -> docs/reference.md
- Services: services/services.yaml -> docs/services.md

Usage:
    python scripts/generate_docs.py

Requirements:
    pip install pyyaml

Best practices:
- Keep commands.yaml and services.yaml up to date as the single source
  of truth.
- Run this script after updating commands/services to keep docs current.
- Integrate with CI or pre-commit for automation.
"""

import yaml
import os
import sys

DOCS_DIR = os.path.abspath(
    os.path.join(os.path.dirname(__file__), '..', 'docs')
)
COMMANDS_YAML = os.path.abspath(
    os.path.join(os.path.dirname(__file__), 'commands.yaml')
)
SERVICES_YAML = os.path.abspath(
    os.path.join(os.path.dirname(__file__), '..', 'services', 'services.yaml')
)


def update_autogen_section(doc_path, generated_content):
    with open(doc_path, "r") as f:
        doc = f.read()
    start_marker = "<!-- AUTO-GENERATED-START -->"
    end_marker = "<!-- AUTO-GENERATED-END -->"
    start = doc.find(start_marker)
    end = doc.find(end_marker)
    if start == -1 or end == -1 or end < start:
        print(
            f"Markers not found or invalid in {doc_path}. "
            "Skipping update."
        )
        return
    new_doc = (
        doc[: start + len(start_marker)]
        + "\n" + generated_content.strip() + "\n"
        + doc[end:]
    )
    with open(doc_path, "w") as f:
        f.write(new_doc)
    print(f"Updated auto-generated section in {doc_path}")


def generate_command_reference():
    try:
        with open(COMMANDS_YAML) as f:
            commands = yaml.safe_load(f)
    except Exception as e:
        print(f"Error reading {COMMANDS_YAML}: {e}")
        sys.exit(1)
    out_path = os.path.join(DOCS_DIR, 'reference.md')
    autogen = []
    autogen.append("# Command Reference (dev-stack)\n")
    autogen.append(
      "This section is auto-generated from "
      "`scripts/commands.yaml`.\n"
    )
    for script, cmds in commands.items():
        autogen.append(f"## {script}")
        for cmd in cmds:
            autogen.append(
                f"- `{cmd}`"
            )
        autogen.append("")
    update_autogen_section(
        out_path,
        "\n".join(autogen)
    )


def generate_services_guide():
    try:
        with open(SERVICES_YAML) as f:
            services = yaml.safe_load(f)
    except Exception as e:
        print(f"Error reading {SERVICES_YAML}: {e}")
        sys.exit(1)
    out_path = os.path.join(DOCS_DIR, 'services.md')
    autogen = []
    autogen.append("# Services Guide (dev-stack)\n")
    autogen.append(
      "This section is auto-generated from "
      "`services/services.yaml`.\n"
    )
    for svc, info in services.items():
        autogen.append(f"## {svc}")
        autogen.append(f"{info.get('description', '')}\n")
        options = info.get('options', [])
        if options:
            autogen.append("**Options:**")
            for opt in options:
                autogen.append(f"- `{opt}`")
            autogen.append("")
        examples = info.get('examples', [])
        if examples:
            autogen.append("**Examples:**")
            for ex in examples:
                autogen.append(f"- `{ex}`")
            autogen.append("")
        usage_notes = info.get('usage_notes', '')
        if usage_notes:
            autogen.append(f"**Usage Notes:** {usage_notes}\n")
        links = info.get('links', [])
        if links:
            autogen.append("**Links:**")
            for link in links:
                autogen.append(
                    f"- [{link}]({link})"
                )
            autogen.append("")
    update_autogen_section(
        out_path,
        "\n".join(autogen)
    )


def main():
    if not os.path.isdir(DOCS_DIR):
        print(f"Docs directory not found: {DOCS_DIR}")
        sys.exit(1)
    generate_command_reference()
    generate_services_guide()
    print("Documentation generation complete.")


if __name__ == "__main__":
    main()
