#!/usr/bin/env python3
"""
docs-sync transform script

Converts docs between the example repo format and the Mintlify docs site format.

Directions:
  to-mintlify  — example repo .md → docs site .mdx
                 - Strips # H1, injects YAML frontmatter (title, noindex)
                 - Rewrites https://docs.cosmos.network/sdk/next/... → sdk/next/...
                 - Rewrites local .md links → .mdx links

  to-example   — docs site .mdx → example repo .md
                 - Strips YAML frontmatter, injects # H1
                 - Rewrites sdk/next/... → https://docs.cosmos.network/sdk/next/...
                 - Rewrites .mdx links → .md links

Usage:
  python3 transform.py --direction to-mintlify --input docs/prerequisites.md --output out.mdx
  python3 transform.py --direction to-example  --input prerequisites.mdx    --output out.md

  # Batch: transform all docs/ files to a target directory
  python3 transform.py --direction to-mintlify --input docs/ --output-dir /path/to/docs-site/sdk/next/tutorials/example/
  python3 transform.py --direction to-example  --input /path/to/docs-site/sdk/next/tutorials/example/ --output-dir docs/
"""

import argparse
import os
import re
import sys

DOCS_BASE_URL = "https://docs.cosmos.network"
# The base path for example tutorials on the docs site (no leading slash)
TUTORIAL_BASE_PATH = "sdk/next/tutorials/example"


def strip_frontmatter(content: str) -> tuple[dict, str]:
    """Remove YAML frontmatter from content. Returns (fields dict, remaining content)."""
    fields = {}
    if not content.startswith("---"):
        return fields, content
    end = content.find("\n---", 3)
    if end == -1:
        return fields, content
    front = content[3:end].strip()
    for line in front.splitlines():
        if ":" in line:
            k, _, v = line.partition(":")
            fields[k.strip()] = v.strip()
    remaining = content[end + 4:].lstrip("\n")
    return fields, remaining


def extract_h1(content: str) -> tuple[str | None, str]:
    """Remove the first # H1 line. Returns (title, remaining content)."""
    lines = content.split("\n")
    for i, line in enumerate(lines):
        if line.startswith("# "):
            title = line[2:].strip()
            remaining = "\n".join(lines[i + 1:]).lstrip("\n")
            return title, remaining
    return None, content


def to_mintlify(content: str, existing_content: str = "") -> str:
    # If the destination file already exists, preserve any extra frontmatter fields
    # (e.g. description, sidebarTitle) — title is always overwritten from the H1.
    existing_fields: dict = {}
    if existing_content:
        existing_fields, _ = strip_frontmatter(existing_content)
        existing_fields.pop("title", None)  # title is always sourced from the H1

    # Strip any existing frontmatter from source
    _, content = strip_frontmatter(content)
    # Extract H1 title
    title, content = extract_h1(content)

    # Rewrite absolute docs.cosmos.network links to Mintlify paths (leading slash, no domain)
    # e.g. https://docs.cosmos.network/sdk/next/learn/foo → /sdk/next/learn/foo
    content = re.sub(
        r'https://docs\.cosmos\.network/([^\s)"\']+)',
        r'/\1',
        content,
    )

    # Rewrite local .md links to full Mintlify paths (no extension, leading slash)
    # e.g. [text](./prerequisites.md) → [text](/sdk/next/tutorials/example/prerequisites)
    content = re.sub(
        r'\(\./([^)]+)\.md\)',
        lambda m: f'(/{TUTORIAL_BASE_PATH}/{m.group(1)})',
        content,
    )

    # Build frontmatter: title first, then any preserved docs-site-only fields
    fm_title = title if title else "Untitled"
    fm_lines = [f"title: {fm_title}"]
    for k, v in existing_fields.items():
        fm_lines.append(f"{k}: {v}")
    frontmatter = "---\n" + "\n".join(fm_lines) + "\n---\n\n"

    return frontmatter + content.lstrip("\n")


def to_example(content: str) -> str:
    # Extract frontmatter
    fields, content = strip_frontmatter(content)
    title = fields.get("title", "")

    # Rewrite tutorial-local Mintlify links back to relative .md links
    # e.g. /sdk/next/tutorials/example/quickstart → ./quickstart.md
    content = re.sub(
        rf'\(/{re.escape(TUTORIAL_BASE_PATH)}/([^)]+)\)',
        r'(./\1.md)',
        content,
    )

    # Rewrite other /sdk/... Mintlify paths (non-tutorial) to full URLs
    # e.g. (/sdk/next/learn/foo) → (https://docs.cosmos.network/sdk/next/learn/foo)
    content = re.sub(
        rf'\(/(?!{re.escape(TUTORIAL_BASE_PATH)})([^\s)"\']+)\)',
        lambda m: f'({DOCS_BASE_URL}/{m.group(1)})',
        content,
    )

    # Rewrite bare sdk/... paths (no leading slash) to full URLs for backwards compat
    # e.g. (sdk/next/learn/foo) → (https://docs.cosmos.network/sdk/next/learn/foo)
    content = re.sub(
        r'\((?!http)(?!/)(sdk/[^\s)"\']+)\)',
        lambda m: f'({DOCS_BASE_URL}/{m.group(1)})',
        content,
    )

    # Inject H1
    header = f"# {title}\n\n" if title else ""
    return header + content.lstrip("\n")


def transform_file(direction: str, input_path: str, output_path: str) -> None:
    with open(input_path, "r", encoding="utf-8") as f:
        content = f.read()

    if direction == "to-mintlify":
        existing_content = ""
        if os.path.exists(output_path):
            with open(output_path, "r", encoding="utf-8") as f:
                existing_content = f.read()
        result = to_mintlify(content, existing_content)
    elif direction == "to-example":
        result = to_example(content)
    else:
        raise ValueError(f"Unknown direction: {direction}")

    os.makedirs(os.path.dirname(output_path) or ".", exist_ok=True)
    with open(output_path, "w", encoding="utf-8") as f:
        f.write(result)

    print(f"  {input_path} → {output_path}")


def main():
    parser = argparse.ArgumentParser(description="Transform docs between example repo and Mintlify format")
    parser.add_argument("--direction", required=True, choices=["to-mintlify", "to-example"])
    parser.add_argument("--input", required=True, help="Input file or directory")
    parser.add_argument("--output", help="Output file (single-file mode)")
    parser.add_argument("--output-dir", help="Output directory (batch mode)")
    args = parser.parse_args()

    if os.path.isdir(args.input):
        # Batch mode
        if not args.output_dir:
            print("ERROR: --output-dir required when --input is a directory", file=sys.stderr)
            sys.exit(1)
        ext_in = ".md" if args.direction == "to-mintlify" else ".mdx"
        ext_out = ".mdx" if args.direction == "to-mintlify" else ".md"
        for fname in sorted(os.listdir(args.input)):
            if fname.endswith(ext_in):
                in_path = os.path.join(args.input, fname)
                out_name = fname[: -len(ext_in)] + ext_out
                out_path = os.path.join(args.output_dir, out_name)
                transform_file(args.direction, in_path, out_path)
    else:
        # Single file mode
        if not args.output:
            print("ERROR: --output required for single-file mode", file=sys.stderr)
            sys.exit(1)
        transform_file(args.direction, args.input, args.output)


if __name__ == "__main__":
    main()
