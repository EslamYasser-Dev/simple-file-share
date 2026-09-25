#!/usr/bin/env bash
#
# Build the landing website (Next.js) and the React app, then publish them
# as one static tree to the gh-pages branch:
#
#   <branch-root>/      landing website (EN/AR)
#   <branch-root>/app/  file-share UI
#
# GitHub Pages serves static files only, so the Go API must be hosted
# separately (Docker/VPS/etc.). Point the UI at it with VITE_API_URL.
#
# Usage:
#   VITE_API_URL=https://api.example.com ./scripts/deploy-pages.sh
#
# Environment:
#   VITE_API_URL        Required. Origin of the deployed backend, no trailing slash.
#   VITE_BASE_PATH      Optional. Sub-path the app is served from. Defaults to
#                       /<repo-name>/app/ (GitHub project pages). Use /app/ for
#                       user/org pages (e.g. <user>.github.io/app/).
#   WEBSITE_BASE_PATH   Optional. Next.js basePath for the landing. Defaults to
#                       /<repo-name> ("" for user/org pages).
#   PAGES_ORIGIN        Optional. Origin the app is reachable at, used for the
#                       landing's "Open app" links. Defaults to
#                       https://<owner>.github.io derived from the remote.
#   PAGES_REMOTE        Optional. Git remote to push to (default: origin).
#   PAGES_BRANCH        Optional. Target branch (default: gh-pages).
#   SKIP_INSTALL        Optional. Set to 1 to skip `yarn install --immutable`.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
FRONTEND_DIR="$REPO_ROOT/frontend"
WEBSITE_DIR="$REPO_ROOT/website"
DIST_DIR="$FRONTEND_DIR/dist"
OUT_DIR="$WEBSITE_DIR/out"

PAGES_REMOTE="${PAGES_REMOTE:-origin}"
PAGES_BRANCH="${PAGES_BRANCH:-gh-pages}"

if [[ -z "${VITE_API_URL:-}" ]]; then
  echo "error: VITE_API_URL is required (origin of the deployed backend)" >&2
  echo "       e.g. VITE_API_URL=https://api.example.com $0" >&2
  exit 1
fi

REPO_NAME="$(basename "$(git -C "$REPO_ROOT" rev-parse --show-toplevel)")"

# Project pages live under /<repo>/; user/org pages set the overrides above.
if [[ -z "${VITE_BASE_PATH:-}" ]]; then
  VITE_BASE_PATH="/$REPO_NAME/app/"
fi
if [[ -z "${WEBSITE_BASE_PATH:-}" ]]; then
  WEBSITE_BASE_PATH="/$REPO_NAME"
fi

# Where the deployed app is reachable, for the landing's "Open app" links.
if [[ -z "${PAGES_ORIGIN:-}" ]]; then
  REMOTE_URL="$(git -C "$REPO_ROOT" remote get-url "$PAGES_REMOTE")"
  OWNER_REPO="${REMOTE_URL##*github.com[:/]}"
  OWNER_REPO="${OWNER_REPO%.git}"
  OWNER="${OWNER_REPO%%/*}"
  PAGES_ORIGIN="https://${OWNER}.github.io"
fi
NEXT_PUBLIC_APP_URL="${PAGES_ORIGIN}${VITE_BASE_PATH}"

export VITE_API_URL VITE_BASE_PATH
export NEXT_PUBLIC_BASE_PATH="$WEBSITE_BASE_PATH" NEXT_PUBLIC_APP_URL

echo "==> Building website (base=$NEXT_PUBLIC_BASE_PATH, app=$NEXT_PUBLIC_APP_URL)"
if [[ "${SKIP_INSTALL:-0}" != "1" ]]; then
  (cd "$WEBSITE_DIR" && corepack yarn install --immutable)
fi
(cd "$WEBSITE_DIR" && corepack yarn build)

echo "==> Building frontend (API=$VITE_API_URL, base=$VITE_BASE_PATH)"
if [[ "${SKIP_INSTALL:-0}" != "1" ]]; then
  (cd "$FRONTEND_DIR" && corepack yarn install --immutable)
fi
(cd "$FRONTEND_DIR" && corepack yarn build)

# SPA fallback for client-side routes + disable Jekyll processing on Pages.
cp "$DIST_DIR/index.html" "$DIST_DIR/404.html"

# Assemble: landing at the branch root, app under /app/.
SITE_DIR="$(mktemp -d)"
trap 'rm -rf "$SITE_DIR"' EXIT
mkdir -p "$SITE_DIR/app"
cp -a "$OUT_DIR/." "$SITE_DIR/"
cp -a "$DIST_DIR/." "$SITE_DIR/app/"
touch "$SITE_DIR/.nojekyll"

REMOTE_URL="$(git -C "$REPO_ROOT" remote get-url "$PAGES_REMOTE")"

echo "==> Publishing to $PAGES_REMOTE/$PAGES_BRANCH (landing at /, app at /app/)"
(
  cd "$SITE_DIR"
  git init -q
  git checkout -q -b "$PAGES_BRANCH"
  git add -A
  git -c user.name="${GIT_AUTHOR_NAME:-pages-bot}" \
      -c user.email="${GIT_AUTHOR_EMAIL:-pages-bot@users.noreply.github.com}" \
      commit -q -m "deploy: site $(date -u +%Y-%m-%dT%H:%M:%SZ)"
  git remote add "$PAGES_REMOTE" "$REMOTE_URL"
  git push -f "$PAGES_REMOTE" "$PAGES_BRANCH"
)

echo "==> Done. Enable Pages (Source: branch '$PAGES_BRANCH') in repo settings if not already."
