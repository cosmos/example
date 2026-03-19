#!/usr/bin/env python3
"""
Tests for the docs-sync transform script.

Run from repo root:
  python3 -m pytest scripts/docs-sync/test_transform.py -v
  # or without pytest:
  python3 scripts/docs-sync/test_transform.py
"""

import sys
import os
import textwrap
import unittest

# Add the scripts/docs-sync directory to path so we can import transform
sys.path.insert(0, os.path.dirname(__file__))
import transform


class TestToMintlify(unittest.TestCase):
    """Tests for to_mintlify direction (example .md → docs site .mdx)."""

    def test_h1_becomes_frontmatter_title(self):
        md = "# Prerequisites\n\nSome content here.\n"
        result = transform.to_mintlify(md)
        self.assertIn("title: Prerequisites", result)
        self.assertNotIn("# Prerequisites", result)

    def test_frontmatter_format(self):
        md = "# Build a Module\n\nContent.\n"
        result = transform.to_mintlify(md)
        self.assertTrue(result.startswith("---\ntitle: Build a Module\n---\n"))

    def test_body_content_preserved(self):
        md = "# Title\n\nThis is the body.\n\n## Section\n\nMore text.\n"
        result = transform.to_mintlify(md)
        self.assertIn("This is the body.", result)
        self.assertIn("## Section", result)
        self.assertIn("More text.", result)

    def test_absolute_docs_link_rewritten_to_mintlify_path(self):
        md = "# Title\n\nSee [page](https://docs.cosmos.network/sdk/next/learn/concepts/modules).\n"
        result = transform.to_mintlify(md)
        self.assertIn("(/sdk/next/learn/concepts/modules)", result)
        self.assertNotIn("https://docs.cosmos.network", result)

    def test_absolute_link_in_text_rewritten(self):
        md = "# Title\n\nRead more at https://docs.cosmos.network/sdk/next/learn/foo.\n"
        result = transform.to_mintlify(md)
        self.assertIn("/sdk/next/learn/foo", result)
        self.assertNotIn("https://docs.cosmos.network/sdk/next/learn/foo", result)

    def test_local_md_link_becomes_absolute_mintlify_path(self):
        md = "# Title\n\nNext: [Quickstart](./quickstart.md).\n"
        result = transform.to_mintlify(md)
        self.assertIn("(/sdk/next/tutorials/example/quickstart)", result)
        self.assertNotIn("./quickstart", result)

    def test_external_links_not_rewritten(self):
        md = "# Title\n\nSee [Go](https://go.dev) and [Docker](https://docs.docker.com/get-docker).\n"
        result = transform.to_mintlify(md)
        self.assertIn("https://go.dev", result)
        self.assertIn("https://docs.docker.com/get-docker", result)

    def test_strips_existing_frontmatter(self):
        md = "---\nnoindex: true\ntitle: Old Title\n---\n\n# New Title\n\nContent.\n"
        result = transform.to_mintlify(md)
        self.assertIn("title: New Title", result)
        self.assertNotIn("title: Old Title", result)

    def test_preserves_extra_frontmatter_from_existing_mdx(self):
        md = "# Prerequisites\n\nContent.\n"
        existing_mdx = "---\ntitle: Prerequisites\ndescription: Learn how to set up your environment.\n---\n\nOld content.\n"
        result = transform.to_mintlify(md, existing_content=existing_mdx)
        self.assertIn("title: Prerequisites", result)
        self.assertIn("description: Learn how to set up your environment.", result)

    def test_title_always_overwritten_not_preserved(self):
        md = "# New Title\n\nContent.\n"
        existing_mdx = "---\ntitle: Old Title\ndescription: Some description.\n---\n\nOld content.\n"
        result = transform.to_mintlify(md, existing_content=existing_mdx)
        self.assertIn("title: New Title", result)
        self.assertNotIn("title: Old Title", result)
        self.assertIn("description: Some description.", result)

    def test_no_existing_mdx_produces_title_only_frontmatter(self):
        md = "# Quickstart\n\nContent.\n"
        result = transform.to_mintlify(md)
        self.assertEqual(result.count("---"), 2)
        self.assertIn("title: Quickstart", result)
        self.assertNotIn("description", result)

    def test_no_h1_uses_untitled(self):
        md = "No heading here.\n\nJust content.\n"
        result = transform.to_mintlify(md)
        self.assertIn("title: Untitled", result)

    def test_code_blocks_preserved(self):
        md = textwrap.dedent("""\
            # Title

            ```bash
            exampled query counter count
            ```
        """)
        result = transform.to_mintlify(md)
        self.assertIn("```bash", result)
        self.assertIn("exampled query counter count", result)

    def test_multiple_links_all_rewritten(self):
        md = textwrap.dedent("""\
            # Title

            See [prereqs](https://docs.cosmos.network/sdk/next/tutorials/example/prerequisites)
            and [quickstart](https://docs.cosmos.network/sdk/next/tutorials/example/quickstart).
        """)
        result = transform.to_mintlify(md)
        self.assertNotIn("https://docs.cosmos.network", result)
        self.assertIn("sdk/next/tutorials/example/prerequisites", result)
        self.assertIn("sdk/next/tutorials/example/quickstart", result)


class TestToExample(unittest.TestCase):
    """Tests for to_example direction (docs site .mdx → example .md)."""

    def test_frontmatter_title_becomes_h1(self):
        mdx = "---\nnoindex: true\ntitle: Prerequisites\n---\n\nContent here.\n"
        result = transform.to_example(mdx)
        self.assertIn("# Prerequisites", result)
        self.assertNotIn("title:", result)
        self.assertNotIn("---", result)

    def test_frontmatter_stripped(self):
        mdx = "---\nnoindex: true\ntitle: Quickstart\n---\n\nBody text.\n"
        result = transform.to_example(mdx)
        self.assertNotIn("noindex", result)
        self.assertNotIn("---", result)

    def test_h1_at_top(self):
        mdx = "---\ntitle: Run and Test\n---\n\nContent.\n"
        result = transform.to_example(mdx)
        self.assertTrue(result.startswith("# Run and Test"))

    def test_non_tutorial_mintlify_path_becomes_absolute_url(self):
        mdx = "---\ntitle: Title\n---\n\nSee [modules](/sdk/next/learn/concepts/modules).\n"
        result = transform.to_example(mdx)
        self.assertIn("https://docs.cosmos.network/sdk/next/learn/concepts/modules", result)
        self.assertNotIn("(/sdk/", result)

    def test_tutorial_link_not_expanded_to_full_url(self):
        mdx = "---\ntitle: Title\n---\n\nSee [quickstart](/sdk/next/tutorials/example/quickstart).\n"
        result = transform.to_example(mdx)
        self.assertIn("./quickstart.md", result)
        self.assertNotIn("https://docs.cosmos.network", result)

    def test_mintlify_absolute_link_becomes_relative_md(self):
        mdx = "---\ntitle: Title\n---\n\nNext: [Quickstart](/sdk/next/tutorials/example/quickstart).\n"
        result = transform.to_example(mdx)
        self.assertIn("./quickstart.md", result)
        self.assertNotIn("/sdk/next/tutorials/example/quickstart", result)

    def test_external_links_not_rewritten(self):
        mdx = "---\ntitle: Title\n---\n\nSee [Go](https://go.dev).\n"
        result = transform.to_example(mdx)
        self.assertIn("https://go.dev", result)

    def test_body_content_preserved(self):
        mdx = "---\ntitle: Title\n---\n\nThis is the body.\n\n## Section\n\nMore text.\n"
        result = transform.to_example(mdx)
        self.assertIn("This is the body.", result)
        self.assertIn("## Section", result)

    def test_code_blocks_preserved(self):
        mdx = textwrap.dedent("""\
            ---
            title: Title
            ---

            ```bash
            make start
            ```
        """)
        result = transform.to_example(mdx)
        self.assertIn("```bash", result)
        self.assertIn("make start", result)

    def test_no_frontmatter_returns_content_unchanged(self):
        mdx = "Just plain content with no frontmatter.\n"
        result = transform.to_example(mdx)
        self.assertIn("Just plain content with no frontmatter.", result)

    def test_multiple_relative_links(self):
        mdx = textwrap.dedent("""\
            ---
            title: Title
            ---

            See [prereqs](sdk/next/tutorials/example/prerequisites)
            and [quickstart](sdk/next/tutorials/example/quickstart).
        """)
        result = transform.to_example(mdx)
        self.assertIn("https://docs.cosmos.network/sdk/next/tutorials/example/prerequisites", result)
        self.assertIn("https://docs.cosmos.network/sdk/next/tutorials/example/quickstart", result)


class TestRoundTrip(unittest.TestCase):
    """Round-trip tests: md → mdx → md and mdx → md → mdx should be stable."""

    def test_roundtrip_example_to_mintlify_to_example(self):
        original = textwrap.dedent("""\
            # Prerequisites

            Before starting, make sure you have Go installed.

            See [quickstart](./quickstart.md) to continue.

            External: [Go](https://go.dev).
            Docs link: [learn](https://docs.cosmos.network/sdk/next/learn/foo).
        """)
        mdx = transform.to_mintlify(original)
        recovered = transform.to_example(mdx)

        # Title round-trips
        self.assertIn("# Prerequisites", recovered)
        # External link preserved
        self.assertIn("https://go.dev", recovered)
        # Local link round-trips back to ./file.md
        self.assertIn("./quickstart.md", recovered)
        # Docs link recovered
        self.assertIn("https://docs.cosmos.network/sdk/next/learn/foo", recovered)

    def test_roundtrip_mintlify_to_example_to_mintlify(self):
        original = textwrap.dedent("""\
            ---
            title: Quickstart
            ---

            Run the chain with `make start`.

            See [prerequisites](/sdk/next/tutorials/example/prerequisites) first.

            Non-tutorial docs link: [learn](sdk/next/learn/foo).
        """)
        md = transform.to_example(original)
        recovered = transform.to_mintlify(md)

        # Frontmatter round-trips
        self.assertIn("title: Quickstart", recovered)
        # Tutorial link round-trips back to absolute Mintlify path
        self.assertIn("/sdk/next/tutorials/example/prerequisites", recovered)
        # Non-tutorial docs link round-trips
        self.assertIn("sdk/next/learn/foo", recovered)

    def test_roundtrip_is_stable_on_second_pass(self):
        """Running to_mintlify twice on the same content should produce the same result."""
        md = "# Title\n\nSome content with [a link](https://docs.cosmos.network/sdk/next/foo).\n"
        first = transform.to_mintlify(md)
        # Simulate editing the mdx and re-converting to md then back
        md2 = transform.to_example(first)
        second = transform.to_mintlify(md2)
        self.assertEqual(first, second)


if __name__ == "__main__":
    unittest.main(verbosity=2)
