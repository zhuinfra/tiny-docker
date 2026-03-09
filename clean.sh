#!/bin/bash

echo "== tiny-docker cleanup start =="

found=false

# 1. kill container processes
for f in /var/run/tiny-docker/*/config.json; do
    [ -f "$f" ] || continue
    found=true

    cid=$(basename "$(dirname "$f")")
    pid=$(jq -r '.pid' "$f")

    echo "[container] $cid"

    if [ -n "$pid" ]; then
        echo "  -> killing pid $pid"
        kill "$pid" 2>/dev/null
    else
        echo "  -> pid not found"
    fi

    merged="/var/lib/tiny-docker/overlay2/$cid/merged"

    # 2. umount overlay
    if mountpoint -q "$merged"; then
        echo "  -> umount $merged"
        umount "$merged"
    else
        echo "  -> overlay not mounted"
    fi

    # 3. remove overlay dir
    echo "  -> remove overlay dir /var/lib/tiny-docker/overlay2/$cid"
    rm -rf "/var/lib/tiny-docker/overlay2/$cid"
done

if [ "$found" = false ]; then
    echo "no containers found"
fi

# 4. clean runtime state
echo "clean runtime metadata"
rm -rf /var/run/tiny-docker/containers/*

echo "== tiny-docker cleanup done =="
