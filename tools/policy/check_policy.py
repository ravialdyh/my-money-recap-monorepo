import argparse
import ast
import fnmatch
import json
import pathlib
import re
import subprocess
import sys

from scanner import C_FAMILY, group, scan, style_for

ROOT = pathlib.Path(__file__).resolve().parents[2]
CONFIG = json.loads((pathlib.Path(__file__).parent / "policy.json").read_text())

DIRECTIVE = re.compile(
    r"^(?:"
    r"!"
    r"|go:\w+|\+build\b|nolint(?::|$)"
    r"|swift-tools-version:|swiftlint:|swiftformat:|swift-format-ignore|sourcery:|periphery:"
    r"|@ts-(?:expect-error|ignore|nocheck)|<reference\b"
    r"|eslint-(?:disable|enable)|biome-ignore|prettier-ignore"
    r"|c8 ignore|v8 ignore|istanbul ignore|@vite-ignore|webpackChunkName"
    r"|sourceMappingURL|sourceURL"
    r"|shellcheck (?:disable|shell|source|enable)="
    r"|type:\s|noqa|pragma:|pylint:|mypy:|ruff:"
    r"|syntax=|escape=|hadolint (?:ignore|global)"
    r"|yaml-language-server:|yamllint"
    r"|lint:\w+"
    r")"
)

EM_DASH = "\u2014"

PIN = re.compile(r"^v?\d+(?:\.\d+)*$")
TRACKER_TAG = re.compile(r"\b(?:TODO|FIXME|XXX|HACK|WIP)\b")

CODE_SHAPE = re.compile(
    r"^\s*(?:func|var|let|const|return|if|for|while|switch|import|package|type|struct"
    r"|class|guard|defer|throw|export|public|private|static|@\w+)\b.*[{;]\s*$"
    r"|:="
    r"|\)\s*=>"
    r"|^\s*\}"
)

_TS = re.compile(
    r"^\s*(?:@\w+|export|declare|default|abstract|async|function|class|interface|type|enum"
    r"|const|let|var|readonly|public|private|protected|static|get|set|import"
    r"|[A-Za-z_$][\w$]*\??\s*[:(<])"
)

DECL = {
    ".go": re.compile(r"^\s*(?:package|func|type|var|const|import)\b"),
    ".swift": re.compile(
        r"^\s*(?:@\w+(?:\([^)]*\))?\s*)*(?:$|public|private|internal|fileprivate|open|final"
        r"|static|nonisolated|class|struct|enum|protocol|extension|actor|func|var|let|init"
        r"|deinit|case|typealias|subscript|associatedtype|import|indirect|mutating|override)\b"
    ),
    ".proto": re.compile(
        r"^\s*(?:syntax|edition|package|import|option|message|enum|service|rpc|extend|oneof"
        r"|reserved|repeated|optional|required|map\s*<|[A-Za-z_][\w.]*\s+[A-Za-z_]\w*\s*=)"
    ),
    ".ts": _TS,
    ".tsx": _TS,
    ".js": _TS,
    ".mjs": _TS,
}


def _relative(path):
    return str(pathlib.Path(path).resolve().relative_to(ROOT)).replace("\\", "/")


def _excluded(rel):
    return any(fnmatch.fnmatch(rel, pattern) for pattern in CONFIG["exclude"])


def _generated(text, style):
    head = "\n".join(text.split("\n")[:5])
    for comment in scan(head, style):
        if any(marker in comment.body for marker in CONFIG["generated_markers"]):
            return True
    return False


def _block_text(block):
    return " ".join(comment.body for comment in block.lines).strip()


def _is_action_pin(rel, lines, block):
    if not rel.startswith(".github/") or not block.trailing:
        return False
    source = lines[block.start - 1]
    return "uses:" in source and PIN.match(_block_text(block)) is not None


def _comment_lines(comments):
    covered = set()
    for comment in comments:
        covered.update(range(comment.line, comment.end_line + 1))
    return covered


def _next_code_line(lines, after, covered):
    for number in range(after + 1, len(lines) + 1):
        if number in covered:
            continue
        return lines[number - 1] if lines[number - 1].strip() else None
    return None


def _attached(rel, lines, block, covered):
    following = _next_code_line(lines, block.end, covered)
    if following is None:
        return False
    if block.scope == "type":
        return not following.strip().startswith(("}", ")", "]"))
    pattern = DECL.get(pathlib.Path(rel).suffix)
    return pattern is not None and pattern.match(following) is not None


def _check_comments(rel, text, style, problems):
    lines = text.split("\n")
    comments = scan(text, style)
    covered = _comment_lines(comments)
    prose = [c for c in comments if not DIRECTIVE.match(c.body)]
    for block in group(prose):
        if _is_action_pin(rel, lines, block):
            continue
        body = _block_text(block)
        if not body:
            continue
        where = f"{rel}:{block.start}"
        if TRACKER_TAG.search(body):
            problems.append(f"{where} tracker tag in a comment; open an issue instead")
            continue
        if block.trailing:
            problems.append(f"{where} trailing comment after code; delete it")
            continue
        if block.scope == "body":
            problems.append(f"{where} comment inside a function body; name the code instead")
            continue
        if any(CODE_SHAPE.search(comment.body) for comment in block.lines):
            problems.append(f"{where} commented-out code; delete it, git remembers")
            continue
        if not _attached(rel, lines, block, covered):
            problems.append(f"{where} comment attached to no declaration; delete it")
            continue
        if len(body) > CONFIG["max_doc_chars"]:
            problems.append(
                f"{where} doc comment is {len(body)} chars, over the {CONFIG['max_doc_chars']} cap"
            )


def _check_em_dash(rel, text, style, problems):
    masked = {}
    for comment in scan(text, style) if style is C_FAMILY else ():
        for number in range(comment.line, comment.end_line + 1):
            start = comment.column if number == comment.line else 0
            masked[number] = min(masked.get(number, start), start)
    for number, line in enumerate(text.split("\n"), 1):
        if EM_DASH in line[: masked.get(number, len(line))]:
            problems.append(f"{rel}:{number} em dash in a string or a page; the app never shows one to a Member, use a full stop or a comma")


def _check_python(rel, text, problems):
    try:
        tree = ast.parse(text)
    except SyntaxError:
        return
    if ast.get_docstring(tree):
        problems.append(f"{rel}:1 module docstring; the file's name and code say what it is")
    for node in ast.walk(tree):
        if not isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef, ast.ClassDef)):
            continue
        doc = ast.get_docstring(node)
        if doc and len(doc) > CONFIG["max_doc_chars"]:
            problems.append(
                f"{rel}:{node.lineno} docstring is {len(doc)} chars, "
                f"over the {CONFIG['max_doc_chars']} cap"
            )


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
        if _excluded(rel):
            continue
        full = ROOT / rel
        if not full.is_file():
            continue
        try:
            text = full.read_text(encoding="utf-8")
        except (UnicodeDecodeError, OSError):
            continue
        style = style_for(rel)
        if style is not None and _generated(text, style):
            continue
        _check_length(rel, text, problems)
        if rel.endswith(".py"):
            _check_python(rel, text, problems)
        if style is not None:
            _check_comments(rel, text, style, problems)
        if style is C_FAMILY or rel.endswith(".html"):
            _check_em_dash(rel, text, style, problems)
    return problems


def main():
    parser = argparse.ArgumentParser(description="Enforces the repository size and comment policy.")
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
