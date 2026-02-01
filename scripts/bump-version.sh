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
GO_VERSION="1.22.8"
DEFAULT_GVM_ROOT="${HOME}/.gvm/gos/go${GO_VERSION}"
CACHE_GO_ROOT="${REPO_ROOT}/.cache/go${GO_VERSION}"
GOLANGCI_VERSION="v1.63.4"
GOBIN="${REPO_ROOT}/bin"
OS_NAME="$(uname -s)"
DARWIN_MAJOR=""
if [[ "${OS_NAME}" == "Darwin" ]]; then
  DARWIN_MAJOR="$(uname -r | cut -d. -f1)"
fi

abort() { echo "ERROR: $*" >&2; exit 1; }
info()  { echo "==> $*"; }
cmd()   { echo "+ $*"; "$@"; }

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || abort "Missing required command: $1"
}

require_cmd curl

ensure_pinned_go() {
  local go_root=""
  if [[ -x "${DEFAULT_GVM_ROOT}/bin/go" ]]; then
    go_root="${DEFAULT_GVM_ROOT}"
  elif [[ -x "${CACHE_GO_ROOT}/bin/go" ]]; then
    go_root="${CACHE_GO_ROOT}"
  else
    info "Pinned Go ${GO_VERSION} not found; downloading to ${CACHE_GO_ROOT}..."
    local go_os go_arch tarball url tmp
    go_os="$(uname | tr '[:upper:]' '[:lower:]')"
    case "$(uname -m)" in
      x86_64|amd64) go_arch="amd64" ;;
      arm64|aarch64) go_arch="arm64" ;;
      *) abort "Unsupported architecture: $(uname -m)" ;;
    esac
    tarball="go${GO_VERSION}.${go_os}-${go_arch}.tar.gz"
    url="https://go.dev/dl/${tarball}"
    tmp="$(mktemp -t go-toolchain.XXXXXX.tar.gz)"
    cmd curl -L "${url}" -o "${tmp}"
    mkdir -p "${REPO_ROOT}/.cache"
    tar -C "${REPO_ROOT}/.cache" -xzf "${tmp}"
    mv "${REPO_ROOT}/.cache/go" "${CACHE_GO_ROOT}"
    rm -f "${tmp}"
    go_root="${CACHE_GO_ROOT}"
  fi
  echo "${go_root}"
}

GO_ROOT="$(ensure_pinned_go)"

GO_BIN=""

if [[ -x "${GO_ROOT}/bin/go" ]]; then
  if [[ "${OS_NAME}" == "Darwin" && -n "${DARWIN_MAJOR}" && "${DARWIN_MAJOR}" -ge 25 ]]; then
    info "Darwin ${DARWIN_MAJOR} detected; skipping pinned Go ${GO_VERSION} due to LC_UUID issues on macOS 15+."
  else
    GO_BIN="${GO_ROOT}/bin/go"
    info "Using pinned Go ${GO_VERSION}: ${GO_BIN}"
  fi
fi

if [[ -z "${GO_BIN}" ]]; then
  GO_BIN="$(command -v go || true)"
  [[ -n "${GO_BIN}" ]] || abort "Missing required command: go"
  info "Pinned Go not available; falling back to system Go: ${GO_BIN}"
fi

export GOMODCACHE="${GOMODCACHE:-${REPO_ROOT}/.cache/gomod}"
export GOBIN
export PATH="${REPO_ROOT}/bin:$(dirname "${GO_BIN}"):${PATH}"
mkdir -p "${GOMODCACHE}"

require_clean_tree() {
  if ! git diff --quiet || ! git diff --cached --quiet; then
    echo "Working tree not clean."
    read -rp "Enter commit message to auto-commit and push changes (leave empty to abort): " COMMIT_MSG
    if [[ -z "${COMMIT_MSG}" ]]; then
      abort "Working tree not clean. Commit or stash changes first."
    fi
    cmd git add -A
    cmd git commit -m "${COMMIT_MSG}"
    cmd git push origin "${MAIN_BRANCH}"
  fi
}

restore_go_mod() {
  git checkout -- go.mod go.sum >/dev/null 2>&1 || true
}

# --- Repo checks ---
[[ -f "go.mod" ]] || abort "Run from the repo root (go.mod not found)."
[[ "$(git rev-parse --show-toplevel)" == "$(pwd)" ]] || abort "Run from the repo root."

require_cmd git

ensure_golangci() {
  local desired="${GOLANGCI_VERSION}"
  local bin_path="${GOBIN}/golangci-lint"
  if [[ -x "${bin_path}" ]]; then
    local have
    have="$("${bin_path}" --version 2>/dev/null | awk '{print $4}')"
    if [[ "${have}" == "${desired}" ]]; then
      echo "${bin_path}"
      return
    fi
    info "Updating golangci-lint ${have} -> ${desired}..."
  else
    info "Installing golangci-lint ${desired}..."
  fi
  cmd "${GO_BIN}" install "github.com/golangci/golangci-lint/cmd/golangci-lint@${desired}"
  echo "${bin_path}"
}

GOLANGCI_BIN="$(ensure_golangci)"

CURRENT_BRANCH="$(git rev-parse --abbrev-ref HEAD)"
MAIN_BRANCH="main"

if [[ "${CURRENT_BRANCH}" != "${MAIN_BRANCH}" ]]; then
  abort "Release must be run on ${MAIN_BRANCH}. Current branch: ${CURRENT_BRANCH}."
fi

require_clean_tree

# --- Quality gates (fail fast before any commits/tags) ---
info "Checking dependencies (go mod tidy)..."
cmd "${GO_BIN}" mod tidy
if ! git diff --quiet -- go.mod go.sum; then
  restore_go_mod
  abort "go mod tidy produced changes; commit those first."
fi

info "Running tests..."
if [[ -n "${COVERPROFILE:-}" ]]; then
  COVER_FILE="$(mktemp -t onyx-go-cover.XXXXXX)"
  cmd "${GO_BIN}" test ./... -coverprofile="${COVER_FILE}" -covermode=atomic
  rm -f "${COVER_FILE}"
else
  cmd "${GO_BIN}" test ./...
fi

info "Linting..."
cmd "${GOLANGCI_BIN}" run

info "Building..."
cmd "${GO_BIN}" build ./...

info "Running examples (smoke test)..."
EXAMPLE_CONFIG="examples/config/onyx-database.json"
EXAMPLE_SCHEMA="examples/api/onyx.schema.json"
[[ -f "${EXAMPLE_CONFIG}" ]] || abort "Missing ${EXAMPLE_CONFIG}."
[[ -f "${EXAMPLE_SCHEMA}" ]] || abort "Missing ${EXAMPLE_SCHEMA}."
( \
  cd examples
  ONYX_CONFIG_PATH="./config/onyx-database.json" \
  ONYX_SCHEMA_PATH="./api/onyx.schema.json" \
    cmd "${GO_BIN}" run ./cmd/query/basic/main.go
)

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

View Actions runs:
- Tag:    https://github.com/OnyxDevTools/onyx-database-go/actions?query=tag%3A${NEXT_VERSION}
- Branch: https://github.com/OnyxDevTools/onyx-database-go/actions?query=branch%3A${MAIN_BRANCH}
- All runs: https://github.com/OnyxDevTools/onyx-database-go/actions
NOTE
