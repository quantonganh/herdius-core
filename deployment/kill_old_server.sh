#!/usr/bin/env bash
set -x

echo "attempting to kill old server process"
pkill -f /tmp/go-build
echo "old server killed"
