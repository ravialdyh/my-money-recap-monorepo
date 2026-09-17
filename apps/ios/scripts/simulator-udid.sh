#!/bin/sh
# Prints the UDID of an iPhone simulator to use as an xcodebuild destination. No device name or OS
# version is written anywhere (issues #16 and #18, "Device and OS coverage"): everything is read from
# `simctl` when the script runs, so a new Xcode that renames its default devices cannot break a test run.
#
#   simulator-udid.sh          a booted iPhone if any, else the first available one
#   simulator-udid.sh newest   an iPhone on the newest installed iOS runtime
#   simulator-udid.sh oldest   an iPhone on the oldest installed iOS runtime (the deployment target's floor)
#
# When the chosen runtime has no iPhone yet, one is created from the runtime's own supported device types.
set -eu
xcrun simctl list -j devices devicetypes runtimes | python3 -c '
import json, subprocess, sys

want = sys.argv[1] if len(sys.argv) > 1 else ""
data = json.load(sys.stdin)
runtimes = [r for r in data["runtimes"] if r["platform"] == "iOS" and r["isAvailable"]]
if not runtimes:
    sys.exit("no available iOS simulator runtime; install one through Xcode > Settings > Components")

def version(r): return tuple(int(p) for p in r["version"].split("."))
def phones(runtime_id):
    return [d for d in data["devices"].get(runtime_id, []) if d["isAvailable"] and "iPhone" in d["deviceTypeIdentifier"]]

if want in ("newest", "oldest"):
    runtime = sorted(runtimes, key=version)[-1 if want == "newest" else 0]
    candidates = phones(runtime["identifier"])
    if not candidates:
        types = [t for t in runtime.get("supportedDeviceTypes", []) if t.get("productFamily") == "iPhone"]
        if not types:
            sys.exit("runtime %s supports no iPhone device type" % runtime["name"])
        name = "mmr-" + runtime["identifier"].rsplit(".", 1)[-1]
        udid = subprocess.check_output(["xcrun", "simctl", "create", name, types[-1]["identifier"], runtime["identifier"]], text=True).strip()
        print(udid); sys.exit()
else:
    candidates = [d for r in runtimes for d in phones(r["identifier"])]
    if not candidates:
        sys.exit("no available iPhone simulator; install an iOS simulator runtime in Xcode")
candidates.sort(key=lambda d: d["state"] != "Booted")
print(candidates[0]["udid"])
' "$@"
