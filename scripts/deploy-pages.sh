#!/usr/bin/env bash
#
# Build the React frontend and publish it to GitHub Pages (gh-pages branch).
#
# GitHub Pages serves static files only, so the Go API must be hosted
# separately (Docker/VPS/etc.). Point the UI at it with VITE_API_URL.
#
# Usage:
#   VITE_API_URL=https://api.example.com ./scripts/deploy-pages.sh
#
# Environment:
#   VITE_API_URL    Required. Origin of the deployed backend, no trailing slash.
#   VITE_BASE_PATH  Optional. Sub-path the app is served from. Defaults to
#                   /<repo-name>/ (GitHub project pages). Use "/" for user/org
#                   pages (e.g. <user>.github.io).
#   PAGES_REMOTE    Optional. Git remote to push to (default: origin).
#   PAGES_BRANCH    Optional. Target branch (default: gh-pages).
#   SKIP_INSTALL    Optional. Set to 1 to skip `yarn install --immutable`.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
FRONTEND_DIR="$REPO_ROOT/frontend"
DIST_DIR="$FRONTEND_DIR/dist"

PAGES_REMOTE="${PAGES_REMOTE:-origin}"
PAGES_BRANCH="${PAGES_BRANCH:-gh-pages}"

if [[ -z "${VITE_API_URL:-}" ]]; then
  echo "error: VITE_API_URL is required (origin of the deployed backend)" >&2
  echo "       e.g. VITE_API_URL=https://api.example.com $0" >&2
  exit 1
fi

# Project pages live under /<repo>/; user/org pages use /.
if [[ -z "${VITE_BASE_PATH:-}" ]]; then
  REPO_NAME="$(basename "$(git -C "$REPO_ROOT" rev-parse --show-toplevel)")"
  VITE_BASE_PATH="/$REPO_NAME/"
fi

export VITE_API_URL VITE_BASE_PATH

echo "==> Building frontend (API=$VITE_API_URL, base=$VITE_BASE_PATH)"
if [[ "${SKIP_INSTALL:-0}" != "1" ]]; then
  (cd "$FRONTEND_DIR" && corepack yarn install --immutable)
fi
(cd "$FRONTEND_DIR" && corepack yarn build)

# SPA fallback for client-side routes + disable Jekyll processing on Pages.
cp "$DIST_DIR/index.html" "$DIST_DIR/404.html"
touch "$DIST_DIR/.nojekyll"

REMOTE_URL="$(git -C "$REPO_ROOT" remote get-url "$PAGES_REMOTE")"

echo "==> Publishing $DIST_DIR to $PAGES_REMOTE/$PAGES_BRANCH"
(
  cd "$DIST_DIR"
  rm -rf .git
  git init -q
  git checkout -q -b "$PAGES_BRANCH"
  git add -A
  git -c user.name="${GIT_AUTHOR_NAME:-pages-bot}" \
      -c user.email="${GIT_AUTHOR_EMAIL:-pages-bot@users.noreply.github.com}" \
      commit -q -m "deploy: frontend $(date -u +%Y-%m-%dT%H:%M:%SZ)"
  git remote add "$PAGES_REMOTE" "$REMOTE_URL"
  git push -f "$PAGES_REMOTE" "$PAGES_BRANCH"
)

echo "==> Done. Enable Pages (Source: branch '$PAGES_BRANCH') in repo settings if not already."
