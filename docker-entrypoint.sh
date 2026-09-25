#!/bin/sh
set -eu

# The container starts as root only to create/own the storage root (custom
# paths and platform volumes are frequently missing or root-owned), then the
# server process itself always runs unprivileged as nobody.
root="${ROOT_DIR:-/data}"
mkdir -p "$root"
chown nobody:nobody "$root"
export HOME=/

exec su-exec nobody /file-server
