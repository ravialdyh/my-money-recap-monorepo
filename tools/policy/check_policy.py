"""Enforces the repository's size and comment policy. Run with no arguments to check everything."""

import argparse
import fnmatch
import json
import pathlib
import re
import subprocess
import sys

from scanner import group, scan

ROOT = pathlib.Path(__file__).resolve().parents[2]
CONFIG = json.loads((pathlib.Path(__file__).parent / "policy.json").read_text())

DIRECTIVE = re.compile(
    r"^(?:"
    r"go:\w+"
    r"|nolint(?::|$)"
    r"|lint:\w+"
    r"|MARK:"
    r"|swiftlint:"
    r"|swiftformat:"
    r"|swift-format-ignore"
    r"|sourcery:"
    r"|@ts-(?:expect-error|ignore|nocheck)"
    r"|eslint-(?:disable|enable)"
    r"|biome-ignore"
    r"|prettier-ignore"
    r"|c8 ignore"
    r"|v8 ignore"
    r"|istanbul ignore"
    r"|#"
    r")"
)

TRACKER_TAG = re.compile(r"\b(?:TODO|FIXME|XXX|HACK|WIP)\b")

CODE_SHAPE = re.compile(
    r"^\s*(?:func|var|let|const|return|if|for|while|switch|import|package|type|struct"
    r"|class|guard|defer|throw|export|public|private|static|@\w+)\b.*[{;]\s*$"
    r"|:="
    r"|\)\s*=>"
    r"|^\s*\}"
)


def _relative(path):
    return str(pathlib.Path(path).resolve().relative_to(ROOT)).replace("\\", "/")


def _excluded(rel):
    return any(fnmatch.fnmatch(rel, pattern) for pattern in CONFIG["exclude"])


def _generated(text):
    head = "\n".join(text.split("\n")[:5])
    return any(marker in head for marker in CONFIG["generated_markers"])


def _under_comment_policy(rel):
    suffix = pathlib.Path(rel).suffix
    if suffix not in CONFIG["comment_policy_extensions"]:
        return False
    return any(rel.startswith(root) for root in CONFIG["comment_policy_roots"])


def _block_text(block):
    return " ".join(comment.body for comment in block.lines).strip()


def _is_directive(block):
    return all(DIRECTIVE.match(comment.body) for comment in block.lines)


def _check_length(rel, text, problems):
    count = len(text.rstrip("\n").split("\n"))
    if pathlib.Path(rel).suffix not in CONFIG["line_cap_extensions"]:
        return
    if count > CONFIG["max_lines"]:
        problems.append(
            f"{rel}:1 file is {count} lines, over the {CONFIG['max_lines']}-line cap"
        )


def _check_comments(rel, text, problems):
    for block in group(scan(text)):
        if _is_directive(block):
            continue
        body = _block_text(block)
        if not body:
            continue
        where = f"{rel}:{block.start}"
        if TRACKER_TAG.search(body):
            problems.append(f"{where} tracker tag in a comment; open an issue instead")
            continue
        if block.scope == "body":
            problems.append(
                f"{where} comment inside a function body; name the code or lift it to a doc comment"
            )
            continue
        if any(CODE_SHAPE.search(comment.body) for comment in block.lines):
            problems.append(f"{where} commented-out code; delete it, git remembers")
            continue
        if len(body) > CONFIG["max_doc_chars"]:
            problems.append(
                f"{where} doc comment is {len(body)} chars, over the {CONFIG['max_doc_chars']} cap"
            )


def _tracked_files():
    out = subprocess.run(
        ["git", "ls-files", "-z"], cwd=ROOT, capture_output=True, text=True, check=True
    )
    return [name for name in out.stdout.split("\0") if name]


def check(paths):
    """Returns the list of policy problems found in paths, newest rule first."""
    problems = []
    for name in paths:
        rel = _relative(ROOT / name) if not pathlib.Path(name).is_absolute() else _relative(name)
        if _excluded(rel):
            continue
        full = ROOT / rel
        if not full.is_file():
            continue
        try:
            text = full.read_text(encoding="utf-8")
        except (UnicodeDecodeError, OSError):
            continue
        if _generated(text):
            continue
        _check_length(rel, text, problems)
        if _under_comment_policy(rel):
            _check_comments(rel, text, problems)
    return problems


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("paths", nargs="*")
    args = parser.parse_args()
    problems = check(args.paths or _tracked_files())
    for problem in sorted(problems):
        print(problem)
    if problems:
        print(f"\n{len(problems)} policy problem(s). See tools/policy/README.md.", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
