import argparse
import fnmatch
import json
import pathlib
import re
import subprocess
import sys

ROOT = pathlib.Path(__file__).resolve().parents[2]
CONFIG = json.loads((pathlib.Path(__file__).parent / "policy.json").read_text())

EM_DASH = re.compile("\u2014" + r"|&mdash;|&#8212;|&#x2014;|\\u\{?2014\}?", re.IGNORECASE)
EM_DASH_NAMED_BY = ("tools/policy/check_policy.py", "tools/policy/test_policy.py")


def _relative(path):
    return str(pathlib.Path(path).resolve().relative_to(ROOT)).replace("\\", "/")


def _excluded(rel):
    return any(fnmatch.fnmatch(rel, pattern) for pattern in CONFIG["exclude"])


def _generated(text):
    head = "\n".join(text.split("\n")[:5])
    return any(marker in head for marker in CONFIG["generated_markers"])


def _check_em_dash(rel, text, problems):
    for number, line in enumerate(text.split("\n"), 1):
        if EM_DASH.search(line):
            problems.append(f"{rel}:{number} em dash; use a full stop, a comma or a colon")


def _check_length(rel, text, problems):
    count = len(text.rstrip("\n").split("\n"))
    if count > CONFIG["max_lines"]:
        problems.append(f"{rel}:1 file is {count} lines, over the {CONFIG['max_lines']}-line cap")


def _tracked_files():
    out = subprocess.run(
        ["git", "ls-files", "-z"], cwd=ROOT, capture_output=True, text=True, check=True
    )
    return [name for name in out.stdout.split("\0") if name]


def check(paths):
    """Returns the list of policy problems found in paths."""
    problems = []
    for name in paths:
        rel = _relative(ROOT / name) if not pathlib.Path(name).is_absolute() else _relative(name)
        full = ROOT / rel
        if not full.is_file():
            continue
        try:
            text = full.read_text(encoding="utf-8")
        except (UnicodeDecodeError, OSError):
            continue
        if rel not in EM_DASH_NAMED_BY:
            _check_em_dash(rel, text, problems)
        if _excluded(rel) or _generated(text):
            continue
        _check_length(rel, text, problems)
    return problems


def main():
    parser = argparse.ArgumentParser(description="Enforces the repository size policy.")
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
