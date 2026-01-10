#!/usr/bin/env bash
# Interactive one-step release script for the Go SDK.
# Prompts for bump type and message, then:
#   - runs quality gates (fail fast before any git mutations)
#   - creates a single release commit on main
#   - tags and pushes the annotated release tag
# CI handles publishing on tag push.

set -euo pipefail

# Pin tools to the repo-local Go toolchain without mutating global shell/gvm state.
REPO_ROOT="$(pwd)"
GO_ROOT="/Users/cosborn/.gvm/gos/go1.22.8"
# Always override to avoid leaking a newer gvm/default toolchain into lint/typecheck.
export GOROOT="${GO_ROOT}"
export GOTOOLCHAIN="go1.22.8"
export GOMODCACHE="${GOMODCACHE:-${REPO_ROOT}/.cache/gomod}"
export PATH="${REPO_ROOT}/bin:${GOROOT}/bin:${PATH}"
mkdir -p "${GOMODCACHE}"

abort() { echo "ERROR: $*" >&2; exit 1; }
info()  { echo "==> $*"; }
cmd()   { echo "+ $*"; "$@"; }

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || abort "Missing required command: $1"
}

require_clean_tree() {
  if ! git diff --quiet || ! git diff --cached --quiet; then
    abort "Working tree not clean. Commit or stash changes first."
  fi
}

restore_go_mod() {
  git checkout -- go.mod go.sum >/dev/null 2>&1 || true
}

# --- Repo checks ---
[[ -f "go.mod" ]] || abort "Run from the repo root (go.mod not found)."
[[ "$(git rev-parse --show-toplevel)" == "$(pwd)" ]] || abort "Run from the repo root."

require_cmd git
require_cmd go
require_cmd golangci-lint

CURRENT_BRANCH="$(git rev-parse --abbrev-ref HEAD)"
MAIN_BRANCH="main"

if [[ "${CURRENT_BRANCH}" != "${MAIN_BRANCH}" ]]; then
  abort "Release must be run on ${MAIN_BRANCH}. Current branch: ${CURRENT_BRANCH}."
fi

require_clean_tree

# --- Quality gates (fail fast before any commits/tags) ---
info "Checking dependencies (go mod tidy)..."
cmd go mod tidy
if ! git diff --quiet -- go.mod go.sum; then
  restore_go_mod
  abort "go mod tidy produced changes; commit those first."
fi

info "Running tests..."
if [[ -n "${COVERPROFILE:-}" ]]; then
  COVER_FILE="$(mktemp -t onyx-go-cover.XXXXXX)"
  cmd go test ./... -coverprofile="${COVER_FILE}" -covermode=atomic
  rm -f "${COVER_FILE}"
else
  cmd go test ./...
fi

info "Linting..."
cmd golangci-lint run

info "Building..."
cmd go build ./...

info "Running examples (smoke test)..."
EXAMPLE_CONFIG="examples/config/onyx-database.json"
EXAMPLE_SCHEMA="examples/api/onyx.schema.json"
[[ -f "${EXAMPLE_CONFIG}" ]] || abort "Missing ${EXAMPLE_CONFIG}."
[[ -f "${EXAMPLE_SCHEMA}" ]] || abort "Missing ${EXAMPLE_SCHEMA}."
ONYX_CONFIG_PATH="${EXAMPLE_CONFIG}" \
ONYX_SCHEMA_PATH="${EXAMPLE_SCHEMA}" \
  cmd go run ./examples/cmd/query/basic/main.go

require_clean_tree

# --- Prompt for bump type ---
read -rp "Bump type (patch/minor/major) [patch]: " BUMP_TYPE
BUMP_TYPE="${BUMP_TYPE:-patch}"
case "${BUMP_TYPE}" in
  patch|minor|major) ;;
  *) abort "Invalid bump type: ${BUMP_TYPE}" ;;
esac

read -rp "Release message [${BUMP_TYPE} release]: " MESSAGE
MESSAGE="${MESSAGE:-"${BUMP_TYPE} release"}"

LATEST_TAG="$(git tag --list "v*.*.*" --sort=-v:refname | head -n1)"
if [[ -z "${LATEST_TAG}" ]]; then
  # First release baseline
  NEXT_VERSION="v0.0.1"
else
  if [[ ! "${LATEST_TAG}" =~ ^v([0-9]+)\.([0-9]+)\.([0-9]+)$ ]]; then
    abort "Latest tag ${LATEST_TAG} is not a valid semver tag."
  fi
  MAJOR="${BASH_REMATCH[1]}"
  MINOR="${BASH_REMATCH[2]}"
  PATCH="${BASH_REMATCH[3]}"
  case "${BUMP_TYPE}" in
    patch) PATCH=$((PATCH + 1)) ;;
    minor) MINOR=$((MINOR + 1)); PATCH=0 ;;
    major) MAJOR=$((MAJOR + 1)); MINOR=0; PATCH=0 ;;
  esac
  NEXT_VERSION="v${MAJOR}.${MINOR}.${PATCH}"
fi

if git rev-parse "${NEXT_VERSION}" >/dev/null 2>&1; then
  abort "Tag ${NEXT_VERSION} already exists."
fi

COMMIT_MESSAGE="chore(release): ${NEXT_VERSION} – ${MESSAGE}"
TAG_MESSAGE="${NEXT_VERSION} – ${MESSAGE}"

info "Creating release commit..."
cmd git add -A
cmd git commit --allow-empty -m "${COMMIT_MESSAGE}"

info "Pushing ${MAIN_BRANCH}..."
cmd git push origin "${MAIN_BRANCH}"

info "Creating annotated tag ${NEXT_VERSION}..."
cmd git tag -a "${NEXT_VERSION}" -m "${TAG_MESSAGE}"

info "Pushing tag ${NEXT_VERSION}..."
cmd git push origin "${NEXT_VERSION}"

cat <<NOTE

Done.

- Bump type: ${BUMP_TYPE}
- Message:   ${MESSAGE}
- Version:   ${NEXT_VERSION}
- Tag:       ${NEXT_VERSION}

CI will publish from tag ${NEXT_VERSION}.
NOTE
