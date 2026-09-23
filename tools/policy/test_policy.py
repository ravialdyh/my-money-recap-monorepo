import pathlib
import sys
import tempfile
import unittest

sys.path.insert(0, str(pathlib.Path(__file__).parent))

import check_policy as cp
from scanner import C_FAMILY, HASH, SQL, scan, style_for


def problems(name, body):
    """Writes body to a temp file named name inside the repo and returns the checker's problems."""
    directory = pathlib.Path(tempfile.mkdtemp(dir=cp.ROOT))
    try:
        path = directory / name
        path.write_text(body)
        return [p.split(" ", 1)[1] for p in cp.check([str(path)])]
    finally:
        for child in directory.iterdir():
            child.unlink()
        directory.rmdir()


class Scanner(unittest.TestCase):
    def test_hash_needs_whitespace_before_it(self):
        self.assertEqual([c.body for c in scan("echo ${VAR#prefix}", HASH)], [])
        self.assertEqual([c.body for c in scan("url=http://x#frag", HASH)], [])
        self.assertEqual([c.body for c in scan("a: b  # note", HASH)], ["note"])

    def test_yaml_truncates_an_unquoted_scalar_at_a_hash(self):
        self.assertEqual([c.body for c in scan("- name: BFF (#15, #17)", HASH)], ["17)"])

    def test_block_comment_reports_its_closing_line(self):
        comments = scan("/**\n * one\n * two\n */\nconst a = 1;", C_FAMILY)
        self.assertEqual((comments[0].line, comments[0].end_line), (1, 4))

    def test_string_literals_are_not_comments(self):
        self.assertEqual([c.body for c in scan('const u = "http://localhost:8080";', C_FAMILY)], [])

    def test_sql_and_style_lookup(self):
        self.assertEqual([c.body for c in scan("SELECT 1; -- why", SQL)], ["why"])
        self.assertIs(style_for("a/b/Dockerfile"), HASH)
        self.assertIs(style_for("a/b/x.go"), C_FAMILY)
        self.assertIsNone(style_for("a/b/x.md"))

    def test_a_declaration_does_not_leak_its_keywords_into_the_next_type(self):
        go = "package x\n\ntype Handler func(int) error\n\ntype Config struct {\n\t// Port is the port.\n\tPort int\n}\n"
        self.assertEqual(problems("a.go", go), [])
        ts = "export const pick = (s: State) => s.x\n\nexport interface P {\n  /** a is a. */\n  a: string;\n}\n"
        self.assertEqual(problems("a.ts", ts), [])

    def test_string_contents_do_not_open_a_body(self):
        swift = 'import S\n\nprivate let retry = "try again later"\n\nstruct Row {\n    // N counts.\n    let n: Int\n}\n'
        self.assertEqual(problems("a.swift", swift), [])

    def test_a_for_header_semicolon_does_not_end_the_statement(self):
        loop = "for (let i = 0; i < 3; i++) {\n  // a note\n}\n"
        self.assertEqual([(c.body, c.scope) for c in scan(loop)], [("a note", "body")])

    def test_a_statement_does_not_leak_into_the_next_brace(self):
        text = "const f = () => 1;\nconst obj = {\n  // a note\n};\n"
        self.assertEqual([(c.body, c.scope) for c in scan(text)], [("a note", "file")])

    def test_a_regex_literal_is_not_code(self):
        for line in ("const re = /\\(/;", "const parts = s.split(/[(]/);", "const q = /)/;"):
            with self.subTest(line=line):
                text = line + "\n\nexport interface P {\n  /** a is a. */\n  a: string;\n}\n"
                self.assertEqual([c.scope for c in scan(text)], ["type"])

    def test_division_and_jsx_are_not_regex_literals(self):
        go = "package x\n\nfunc F(a, b int) int {\n\t// a note\n\treturn (a + b) / (a - b)\n}\n"
        self.assertEqual([c.scope for c in scan(go)], ["body"])
        tsx = "export function C() {\n  // a note\n  return <div>{x}</div>;\n}\n"
        self.assertEqual([c.scope for c in scan(tsx)], ["body"])

    def test_a_dotfile_that_is_all_extension_resolves(self):
        self.assertIs(style_for(".swiftformat"), HASH)
        self.assertIs(style_for("a/b/.editorconfig"), HASH)
        self.assertIsNone(style_for("Makefile"))


class Comments(unittest.TestCase):
    def test_docstring_on_a_declaration_is_allowed(self):
        self.assertEqual(problems("a.go", "// Thing does a thing.\nfunc Thing() {}\n"), [])

    def test_go_package_comment_is_allowed(self):
        self.assertEqual(problems("a.go", "// Package x is the x.\npackage x\n"), [])

    def test_file_header_attached_to_nothing_is_rejected(self):
        found = problems("a.go", "// What this file is about.\n\npackage x\n")
        self.assertEqual(found, ["comment attached to no declaration; delete it"])

    def test_trailing_comment_is_rejected(self):
        found = problems("a.go", "package x\n\nvar N = 1 // why one\n")
        self.assertEqual(found, ["trailing comment after code; delete it"])

    def test_body_comment_is_rejected(self):
        found = problems("a.go", "package x\n\nfunc F() {\n\t// step one\n\t_ = 1\n}\n")
        self.assertEqual(found, ["comment inside a function body; name the code instead"])

    def test_typescript_function_body_is_body_scope(self):
        found = problems("a.ts", "export function f() {\n  // a note\n  return 1;\n}\n")
        self.assertEqual(found, ["comment inside a function body; name the code instead"])

    def test_a_wrapped_signature_still_opens_a_body(self):
        go = "package x\n\nfunc Wide(\n\ta int,\n) int {\n\t// a note\n\treturn a\n}\n"
        self.assertEqual(problems("a.go", go), ["comment inside a function body; name the code instead"])

    def test_a_destructured_props_signature_still_opens_a_body(self):
        tsx = "export function C({\n  a,\n}: Props) {\n  // a note\n  return a;\n}\n"
        self.assertEqual(problems("a.tsx", tsx), ["comment inside a function body; name the code instead"])

    def test_a_wrapped_signature_holding_braces_still_opens_a_body(self):
        go = "package x\n\nfunc F(\n\tdone chan struct{},\n) error {\n\t// a note\n\treturn nil\n}\n"
        self.assertEqual(problems("a.go", go), ["comment inside a function body; name the code instead"])
        ts = "export function f(\n  a: Record<string, string> = {},\n): void {\n  // a note\n  const x = a;\n}\n"
        self.assertEqual(problems("a.ts", ts), ["comment inside a function body; name the code instead"])

    def test_a_wrapped_swift_signature_still_opens_a_body(self):
        swift = "struct S {\n    func wide(\n        a: Int\n    ) -> Int {\n        // a note\n        return a\n    }\n}\n"
        self.assertEqual(problems("a.swift", swift), ["comment inside a function body; name the code instead"])

    def test_tracker_tag_is_rejected(self):
        found = problems("a.go", "// TODO: fix this.\nfunc F() {}\n")
        self.assertEqual(found, ["tracker tag in a comment; open an issue instead"])

    def test_docstring_over_the_cap_is_rejected(self):
        found = problems("a.go", "// " + "x" * 300 + "\nfunc F() {}\n")
        self.assertEqual(len(found), 1)
        self.assertIn("over the 255 cap", found[0])

    def test_member_docstring_inside_a_type_is_allowed(self):
        body = "package x\n\n// S is a struct.\ntype S struct {\n\t// N counts.\n\tN int\n}\n"
        self.assertEqual(problems("a.go", body), [])

    def test_a_conf_file_is_held_to_the_hash_rules(self):
        found = problems("default.conf", "# the edge\nserver {\n    listen 80;\n}\n")
        self.assertEqual(found, ["comment attached to no declaration; delete it"])


class EmDash(unittest.TestCase):
    """The owner's rule (2026-09-23): no em dash in any text file, code, comment or document."""

    said = "em dash; use a full stop, a comma or a colon"

    def test_an_em_dash_in_a_string_is_rejected(self):
        self.assertEqual(problems("a.swift", 'let copy = "It\'s been a while \u2014 sign in again."\n'), [self.said])

    def test_an_em_dash_in_a_comment_is_rejected(self):
        self.assertEqual(problems("a.swift", "/// Signs in \u2014 or says why not.\nfunc signIn() {}\n"), [self.said])

    def test_an_em_dash_in_a_document_the_other_rules_skip_is_rejected(self):
        self.assertEqual(problems("a.md", "# Title \u2014 subtitle\n"), [self.said])

    def test_an_escaped_em_dash_is_rejected(self):
        self.assertEqual(problems("a.swift", 'let copy = "a \\u{2014} b"\n'), [self.said])
        self.assertEqual(problems("a.html", "<p>a &mdash; b</p>\n"), [self.said])

    def test_an_em_dash_in_a_plist_is_rejected(self):
        self.assertEqual(problems("Info.plist", "<string>Unlock \u2014 please</string>\n"), [self.said])

    def test_an_en_dash_is_allowed(self):
        self.assertEqual(problems("a.md", "Payday 1\u201331\n"), [])


class Directives(unittest.TestCase):
    def test_load_bearing_directives_survive(self):
        def in_body(line):
            return "package x\n\nfunc F() {\n\t" + line + "\n\t_ = 1\n}\n"

        found = problems("a.go", in_body("// a note"))
        self.assertEqual(found, ["comment inside a function body; name the code instead"])
        for line in (
            "//go:embed migrations/*.sql",
            "//go:build linux",
            "//nolint:errcheck // explained",
            "// swift-tools-version: 5.9",
            "// swiftlint:disable all",
            "// @ts-expect-error legacy",
            "// biome-ignore lint: on purpose",
        ):
            with self.subTest(line=line):
                self.assertEqual(problems("a.go", in_body(line)), [])

    def test_shebang_survives_and_the_prose_under_it_does_not(self):
        self.assertEqual(problems("a.sh", "#!/bin/sh\nset -eu\n"), [])
        found = problems("a.sh", "#!/bin/sh\n# what this script does\nset -eu\n")
        self.assertEqual(found, ["comment attached to no declaration; delete it"])

    def test_shellcheck_waiver_survives(self):
        body = "#!/bin/sh\n# shellcheck disable=SC1091\n. ./.env\n"
        self.assertEqual(problems("a.sh", body), [])


class Generated(unittest.TestCase):
    def test_marker_in_a_comment_exempts_the_file(self):
        self.assertEqual(problems("a.go", "// Code generated by x. DO NOT EDIT.\n\n// stray\n"), [])

    def test_marker_in_a_string_literal_does_not_exempt_the_file(self):
        body = 'package x\n\nconst h = "Ask before you DO NOT EDIT the ledger."\n\n// stray note\n'
        self.assertEqual(problems("a.go", body), ["comment attached to no declaration; delete it"])


class Length(unittest.TestCase):
    def test_over_the_line_cap_is_rejected(self):
        found = problems("a.go", "package x\n" + "var _ = 1\n" * 400)
        self.assertEqual(len(found), 1)
        self.assertIn("over the 300-line cap", found[0])

    def test_a_file_with_no_comment_syntax_is_still_capped(self):
        self.assertEqual(problems("a.json", "{}\n" * 400), ["file is 400 lines, over the 300-line cap"])

    def test_a_generated_file_is_exempt_from_the_cap(self):
        body = "// Code generated by x. DO NOT EDIT.\n\npackage x\n" + "var _ = 1\n" * 400
        self.assertEqual(problems("a.go", body), [])

    def test_a_blank_line_between_a_comment_and_a_declaration_detaches_it(self):
        self.assertEqual(problems("a.go", "// Thing does a thing.\nfunc Thing() {}\n"), [])
        found = problems("a.go", "// Thing does a thing.\n\nfunc Thing() {}\n")
        self.assertEqual(found, ["comment attached to no declaration; delete it"])

    def test_a_directive_between_a_docstring_and_its_declaration_keeps_it_attached(self):
        body = "package x\n\n// F does a thing.\n//go:noinline\nfunc F() {}\n"
        self.assertEqual(problems("a.go", body), [])


if __name__ == "__main__":
    unittest.main()
