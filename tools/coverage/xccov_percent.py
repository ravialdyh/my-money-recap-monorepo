import json
import sys

TARGET = "MyMoneyRecap.app"


def percent(report):
    """Returns the app target's line coverage as a percentage, or raises if the target is absent."""
    for target in report:
        if target.get("name", "").startswith(TARGET):
            return round(target["lineCoverage"] * 100, 1)
    raise LookupError(f"no {TARGET} target in the coverage report")


def main():
    try:
        print(percent(json.load(sys.stdin)))
    except (LookupError, KeyError, ValueError) as err:
        print(err, file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
