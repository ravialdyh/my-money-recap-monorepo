"""Coverage ratchet: a language's coverage may never fall, and must reach the target to stop rising."""

import argparse
import json
import pathlib
import sys

BASELINE = pathlib.Path(__file__).parent / "baseline.json"
TOLERANCE = 0.5


def _load():
    return json.loads(BASELINE.read_text())


def _save(data):
    BASELINE.write_text(json.dumps(data, indent=2, sort_keys=True) + "\n")


def gate(language, measured, update):
    """Returns (exit_code, message) for one language's measured percentage."""
    data = _load()
    target = data["target"]
    floors = data["floors"]
    if language not in floors:
        return 1, f"{language}: no floor recorded. Add it to {BASELINE.name}."
    floor = floors[language]
    if measured + TOLERANCE < floor:
        return 1, (
            f"{language}: coverage fell to {measured:.1f}%, below the {floor:.1f}% floor. "
            "Add tests for what you changed, or say why the floor should move."
        )
    if measured >= target:
        if floor < target:
            floors[language] = target
            if update:
                _save(data)
            return 0, f"{language}: {measured:.1f}% reaches the {target:.0f}% target. Floor raised to {target:.0f}%."
        return 0, f"{language}: {measured:.1f}%, at or above the {target:.0f}% target."
    if measured > floor + TOLERANCE:
        floors[language] = round(measured, 1)
        if update:
            _save(data)
            return 0, f"{language}: {measured:.1f}%, up from {floor:.1f}%. Floor raised; commit {BASELINE.name}."
        return 0, (
            f"{language}: {measured:.1f}%, up from {floor:.1f}%. "
            f"Run `python3 tools/coverage/coverage_gate.py {language} {measured:.1f} --update` and commit the floor."
        )
    return 0, f"{language}: {measured:.1f}%, holding the {floor:.1f}% floor (target {target:.0f}%)."


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("language")
    parser.add_argument("measured", type=float)
    parser.add_argument("--update", action="store_true")
    args = parser.parse_args()
    code, message = gate(args.language, args.measured, args.update)
    print(message, file=sys.stderr if code else sys.stdout)
    return code


if __name__ == "__main__":
    sys.exit(main())
