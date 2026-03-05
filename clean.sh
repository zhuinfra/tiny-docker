#!/bin/bash

# 1. clean overlay
MERGED=/var/lib/tiny-docker/overlay2/containerID/merged

if mountpoint -q "$MERGED"; then
    umount "$MERGED"
fi

rm -rf /var/lib/tiny-docker/overlay2/containerID

# 2. kill container processes
for f in /var/run/tiny-docker/*/config.json; do
    [ -f "$f" ] || continue

    pid=$(jq -r '.pid' "$f")

    if [ -n "$pid" ]; then
        kill "$pid" 2>/dev/null
    fi
done

# 3. clean runtime state
rm -rf /var/run/tiny-docker/*
