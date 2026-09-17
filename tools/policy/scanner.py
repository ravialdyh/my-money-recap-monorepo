"""Comment scanner that understands string literals, so "http://x" is never a comment."""

import re
from dataclasses import dataclass

LINE_MARKERS = ("///", "//")
BLOCK_OPEN = "/*"
BLOCK_CLOSE = "*/"

TYPE_OPENER = re.compile(
    r"\b(?:"
    r"type\s+\w+\s+(?:struct|interface)\b"
    r"|type\s+\w+\s*=\s*\{"
    r"|(?:struct|class|enum|protocol|extension|actor|interface|namespace)\s+\w"
    r")"
)

BODY_OPENER = re.compile(r"\b(?:func|init|deinit|subscript|get|set|willSet|didSet)\b|=>")


@dataclass(frozen=True)
class Comment:
    line: int
    column: int
    marker: str
    body: str
    scope: str


@dataclass(frozen=True)
class Block:
    start: int
    end: int
    scope: str
    lines: tuple


class _Frame:
    def __init__(self, scope):
        self.scope = scope


def _strip_stars(fragment):
    return fragment.strip().lstrip("*").strip()


def _swift_multiline(text, i):
    return text.startswith('"""', i)


def _consume_string(line, i, delim, escape):
    n = len(line)
    while i < n:
        ch = line[i]
        if escape and ch == "\\":
            i += 2
            continue
        if line.startswith(delim, i):
            return i + len(delim), False
        i += 1
    return n, True


def _scope_for(opening_line, parent_scope):
    if parent_scope == "body":
        return "body"
    if BODY_OPENER.search(opening_line):
        return "body"
    if TYPE_OPENER.search(opening_line):
        return "type"
    return "body" if parent_scope == "type" else "file"


def scan(text):
    """Yields every Comment in text, with the scope it sits in: file, type or body."""
    comments = []
    stack = [_Frame("file")]
    in_block = False
    block_start = 0
    block_parts = []
    in_raw = None
    lines = text.split("\n")
    for number, line in enumerate(lines, start=1):
        i = 0
        n = len(line)
        if in_raw is not None:
            end, unfinished = _consume_string(line, 0, in_raw, False)
            if unfinished:
                continue
            in_raw = None
            i = end
        if in_block:
            close = line.find(BLOCK_CLOSE)
            if close < 0:
                block_parts.append(_strip_stars(line))
                continue
            block_parts.append(_strip_stars(line[:close]))
            comments.append(
                Comment(block_start, 0, BLOCK_OPEN, " ".join(part for part in block_parts if part), stack[-1].scope)
            )
            in_block = False
            block_parts = []
            i = close + len(BLOCK_CLOSE)
        opener = line
        while i < n:
            ch = line[i]
            if line.startswith(BLOCK_OPEN, i):
                close = line.find(BLOCK_CLOSE, i + 2)
                if close < 0:
                    in_block = True
                    block_start = number
                    block_parts = [_strip_stars(line[i + 2 :])]
                    break
                comments.append(
                    Comment(number, i, BLOCK_OPEN, _strip_stars(line[i + 2 : close]), stack[-1].scope)
                )
                i = close + len(BLOCK_CLOSE)
                continue
            marker = next((m for m in LINE_MARKERS if line.startswith(m, i)), None)
            if marker:
                comments.append(
                    Comment(number, i, marker, line[i + len(marker) :].strip(), stack[-1].scope)
                )
                break
            if _swift_multiline(line, i):
                end, unfinished = _consume_string(line, i + 3, '"""', True)
                if unfinished:
                    in_raw = '"""'
                    break
                i = end
                continue
            if ch in "\"'`":
                end, unfinished = _consume_string(line, i + 1, ch, ch != "`")
                if unfinished:
                    in_raw = ch if ch == "`" else None
                    if in_raw is None:
                        break
                    break
                i = end
                continue
            if ch == "{":
                stack.append(_Frame(_scope_for(opener, stack[-1].scope)))
            elif ch == "}" and len(stack) > 1:
                stack.pop()
            i += 1
    return comments


def group(comments):
    """Groups consecutive single-line comments into blocks, preserving scope."""
    blocks = []
    run = []
    for comment in comments:
        if comment.marker == BLOCK_OPEN:
            blocks.append(Block(comment.line, comment.line, comment.scope, (comment,)))
            continue
        if run and comment.line == run[-1].line + 1 and comment.scope == run[-1].scope:
            run.append(comment)
            continue
        if run:
            blocks.append(Block(run[0].line, run[-1].line, run[0].scope, tuple(run)))
        run = [comment]
    if run:
        blocks.append(Block(run[0].line, run[-1].line, run[0].scope, tuple(run)))
    return blocks
