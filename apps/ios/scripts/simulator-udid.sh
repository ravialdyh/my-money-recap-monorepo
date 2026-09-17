#!/bin/sh
# Prints the UDID of an iPhone simulator to use as an xcodebuild destination: a booted one if any,
# else the first available. No device name or OS version is written anywhere (issue #16, "Device
# and OS coverage"); a new Xcode that renames its default devices cannot break the test run.
set -eu
xcrun simctl list devices available -j | python3 -c '
import json, sys
runtimes = json.load(sys.stdin)["devices"]
phones = [d for runtime, devices in runtimes.items() if "iOS" in runtime
          for d in devices if d["isAvailable"] and "iPhone" in d["deviceTypeIdentifier"]]
if not phones:
    sys.exit("no available iPhone simulator; install an iOS simulator runtime in Xcode")
phones.sort(key=lambda d: d["state"] != "Booted")
print(phones[0]["udid"])
'
