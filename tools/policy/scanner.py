import re
from dataclasses import dataclass

TYPE_OPENER = re.compile(
    r"\b(?:"
    r"type\s+\w+\s+(?:struct|interface)\b"
    r"|type\s+\w+\s*=\s*\{"
    r"|(?:struct|class|enum|protocol|extension|actor|interface|namespace)\s+\w"
    r")"
)

MAX_CONTINUATION_LINES = 100

BODY_OPENER = re.compile(
    r"\b(?:func|function|init|deinit|subscript|get|set|willSet|didSet"
    r"|if|else|for|while|switch|do|try|catch|guard|repeat|defer)\b|=>"
)


@dataclass(frozen=True)
class Style:
    line_markers: tuple
    block: tuple
    quotes: str
    raw_triple: str
    braces: bool


C_FAMILY = Style(("///", "//"), ("/*", "*/"), "\"'`", '"""', True)
HASH = Style(("#",), (), "\"'", "", False)
SQL = Style(("--",), ("/*", "*/"), "'", "", False)
CSS = Style((), ("/*", "*/"), "\"'", "", False)

STYLES = {
    ".go": C_FAMILY,
    ".swift": C_FAMILY,
    ".ts": C_FAMILY,
    ".tsx": C_FAMILY,
    ".js": C_FAMILY,
    ".mjs": C_FAMILY,
    ".proto": C_FAMILY,
    ".css": CSS,
    ".sql": SQL,
    ".sh": HASH,
    ".yml": HASH,
    ".yaml": HASH,
    ".py": HASH,
    ".toml": HASH,
    ".gitignore": HASH,
    ".dockerignore": HASH,
    ".example": HASH,
    ".editorconfig": HASH,
    ".swiftformat": HASH,
}

NAMED = {"Dockerfile": HASH, ".gitignore": HASH, ".dockerignore": HASH}


def style_for(path):
    """Returns the comment Style for a path, or None when the file has no known syntax."""
    name = path.rsplit("/", 1)[-1]
    if name in NAMED:
        return NAMED[name]
    dot = name.rfind(".")
    return STYLES.get(name[dot:] if dot > 0 else name)


@dataclass(frozen=True)
class Comment:
    line: int
    end_line: int
    column: int
    marker: str
    body: str
    scope: str
    trailing: bool


@dataclass(frozen=True)
class Block:
    start: int
    end: int
    scope: str
    trailing: bool
    lines: tuple


def _strip_stars(fragment):
    return fragment.strip().lstrip("*").strip()


def _consume_string(line, i, delim, escape):
    n = len(line)
    while i < n:
        if escape and line[i] == "\\":
            i += 2
            continue
        if line.startswith(delim, i):
            return i + len(delim), False
        i += 1
    return n, True


def _scope_for(opening_text, parent_scope, reopened):
    if parent_scope == "body":
        return "body"
    if BODY_OPENER.search(opening_text):
        return "body"
    if TYPE_OPENER.search(opening_text):
        return "type"
    if reopened is not None:
        return reopened
    return "body" if parent_scope == "type" else "file"


class _State:
    def __init__(self, style):
        self.style = style
        self.stack = ["file"]
        self.in_block = False
        self.block_start = 0
        self.block_parts = []
        self.in_raw = None
        self.pending = ""
        self.pending_lines = 0
        self.reopened = None
        self.depth = 0
        self.out = []

    def scope(self):
        return self.stack[-1]

    def carry(self, code):
        """Keeps code across a line break only while a bracket is open; a finished line starts afresh."""
        self.pending += code
        if self.depth and self.pending_lines < MAX_CONTINUATION_LINES:
            self.pending += "\n"
            self.pending_lines += 1
            return
        self.pending = ""
        self.pending_lines = 0
        self.depth = 0

    def emit(self, line, end_line, column, marker, body, trailing):
        self.out.append(Comment(line, end_line, column, marker, body, self.scope(), trailing))


def _close_block(st, number, line, close):
    st.block_parts.append(_strip_stars(line[:close]))
    body = " ".join(part for part in st.block_parts if part)
    st.emit(st.block_start, number, 0, st.style.block[0], body, False)
    st.in_block = False
    st.block_parts = []
    return close + len(st.style.block[1])


def _scan_line(st, number, line):
    style = st.style
    i = 0
    n = len(line)
    if st.in_raw is not None:
        end, unfinished = _consume_string(line, 0, st.in_raw, False)
        if unfinished:
            return
        st.in_raw = None
        i = end
    if st.in_block:
        close = line.find(style.block[1])
        if close < 0:
            st.block_parts.append(_strip_stars(line))
            return
        i = _close_block(st, number, line, close)
    st.reopened = None
    code_from = i
    while i < n:
        ch = line[i]
        if style.block and line.startswith(style.block[0], i):
            close = line.find(style.block[1], i + len(style.block[0]))
            st.pending += line[code_from:i]
            if close < 0:
                st.in_block = True
                st.block_start = number
                st.block_parts = [_strip_stars(line[i + len(style.block[0]) :])]
                return
            body = _strip_stars(line[i + len(style.block[0]) : close])
            st.emit(number, number, i, style.block[0], body, bool(line[:i].strip()))
            i = close + len(style.block[1])
            code_from = i
            continue
        marker = next((m for m in style.line_markers if line.startswith(m, i)), None)
        if marker and (marker != "#" or i == 0 or line[i - 1].isspace()):
            body = line[i + len(marker) :].strip()
            st.emit(number, number, i, marker, body, bool(line[:i].strip()))
            st.carry(line[code_from:i])
            return
        if style.raw_triple and line.startswith(style.raw_triple, i):
            end, unfinished = _consume_string(line, i + 3, style.raw_triple, True)
            st.pending += line[code_from:i]
            if unfinished:
                st.in_raw = style.raw_triple
                return
            i = end
            code_from = i
            continue
        if ch in style.quotes:
            end, unfinished = _consume_string(line, i + 1, ch, ch != "`")
            st.pending += line[code_from:i]
            if unfinished:
                st.in_raw = ch if ch == "`" else None
                return
            i = end
            code_from = i
            continue
        if style.braces and ch == "{":
            st.pending += line[code_from : i + 1]
            st.stack.append(_scope_for(st.pending + line[i + 1 :], st.scope(), st.reopened))
            if not st.depth:
                st.pending = ""
                st.reopened = None
            code_from = i + 1
        elif style.braces and ch == "}":
            if not st.depth:
                st.pending = ""
            code_from = i + 1
            if len(st.stack) > 1:
                st.reopened = st.stack.pop()
        elif ch == "(":
            st.depth += 1
        elif ch == ")":
            st.depth = max(0, st.depth - 1)
        elif ch == ";" and not st.depth:
            st.pending = ""
            code_from = i + 1
        i += 1
    st.carry(line[code_from:])


def scan(text, style=C_FAMILY):
    """Yields every Comment in text, with the scope it sits in: file, type or body."""
    st = _State(style)
    for number, line in enumerate(text.split("\n"), start=1):
        _scan_line(st, number, line)
    return st.out


def group(comments):
    """Groups consecutive single-line comments into blocks, preserving scope."""
    blocks = []
    run = []

    def flush():
        if run:
            blocks.append(
                Block(run[0].line, run[-1].end_line, run[0].scope, run[0].trailing, tuple(run))
            )

    for comment in comments:
        if comment.marker == "/*" or comment.trailing:
            flush()
            run = []
            blocks.append(
                Block(comment.line, comment.end_line, comment.scope, comment.trailing, (comment,))
            )
            continue
        if (
            run
            and comment.line == run[-1].end_line + 1
            and comment.scope == run[-1].scope
            and comment.trailing == run[-1].trailing
        ):
            run.append(comment)
            continue
        flush()
        run = [comment]
    flush()
    return sorted(blocks, key=lambda b: b.start)
