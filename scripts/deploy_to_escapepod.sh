#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

REMOTE_HOST="${REMOTE_HOST:-pi@escapepod}"
REMOTE_DIR="${REMOTE_DIR:-wire-pod}"
REMOTE_LOG="${REMOTE_LOG:-wire-pod-run.log}"
RSYNC_SSH_OPTS="${RSYNC_SSH_OPTS:-}"

DRY_RUN=0
SKIP_RESTART=0
SKIP_SYNC=0

usage() {
  cat <<'EOF'
Usage:
  ./scripts/deploy_to_escapepod.sh [options]

Options:
  --host <user@host>       Remote SSH target (default: pi@escapepod)
  --remote-dir <path>      Remote project directory (default: wire-pod, relative to remote HOME)
  --remote-log <path>      Remote log file when fallback start.sh is used
  --skip-sync              Do not upload files, only restart remotely
  --skip-restart           Only upload files, do not restart
  --dry-run                Show rsync changes without applying them
  -h, --help               Show this help

Examples:
  ./scripts/deploy_to_escapepod.sh
  ./scripts/deploy_to_escapepod.sh --host pi@192.168.1.38
  ./scripts/deploy_to_escapepod.sh --remote-dir /home/pi/wire-pod
  ./scripts/deploy_to_escapepod.sh --dry-run
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --host)
      REMOTE_HOST="$2"
      shift 2
      ;;
    --remote-dir)
      REMOTE_DIR="$2"
      shift 2
      ;;
    --remote-log)
      REMOTE_LOG="$2"
      shift 2
      ;;
    --skip-sync)
      SKIP_SYNC=1
      shift
      ;;
    --skip-restart)
      SKIP_RESTART=1
      shift
      ;;
    --dry-run)
      DRY_RUN=1
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown option: $1" >&2
      usage >&2
      exit 1
      ;;
  esac
done

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Missing required command: $1" >&2
    exit 1
  fi
}

require_cmd rsync
require_cmd ssh

SSH_CMD=(ssh)
if [[ -n "${RSYNC_SSH_OPTS}" ]]; then
  # shellcheck disable=SC2206
  SSH_EXTRA_OPTS=(${RSYNC_SSH_OPTS})
  SSH_CMD+=("${SSH_EXTRA_OPTS[@]}")
fi

RESTART_SSH_CMD=("${SSH_CMD[@]}" -tt)

echo "==> Deploy target: ${REMOTE_HOST}:${REMOTE_DIR}"

if [[ "${SKIP_SYNC}" -eq 0 ]]; then
  echo "==> Syncing local repository to remote host"

  RSYNC_ARGS=(
    -avz
    --progress
    --exclude
    ".git"
    --exclude
    ".DS_Store"
    --exclude
    ".codegraph"
    --exclude
    ".trae"
    --exclude
    ".specify"
    --exclude
    ".vscode"
    --exclude
    "scripts/.DS_Store"
    --exclude
    "chipper/source.sh"
    --exclude
    "chipper/apiConfig.json"
    --exclude
    "chipper/jdocs/"
    --exclude
    "chipper/session-certs/"
    --exclude
    "certs/server_config.json"
    --exclude
    "chipper/chipper"
    -e
    "${SSH_CMD[*]}"
  )

  if [[ "${DRY_RUN}" -eq 1 ]]; then
    RSYNC_ARGS+=(--dry-run)
  fi

  rsync "${RSYNC_ARGS[@]}" "${REPO_ROOT}/" "${REMOTE_HOST}:${REMOTE_DIR}/"
fi

if [[ "${SKIP_RESTART}" -eq 1 ]]; then
  echo "==> Upload complete, restart skipped"
  exit 0
fi

if [[ "${DRY_RUN}" -eq 1 ]]; then
  echo "==> Dry run complete, remote restart skipped"
  exit 0
fi

echo "==> Restarting remote application"

"${RESTART_SSH_CMD[@]}" "${REMOTE_HOST}" \
  "env REMOTE_DIR='${REMOTE_DIR}' REMOTE_LOG='${REMOTE_LOG}' bash -s" <<'EOF'
set -euo pipefail

expand_remote_path() {
  local path="$1"

  if [[ -z "${path}" || "${path}" == "~" ]]; then
    printf '%s\n' "${HOME}"
    return
  fi

  if [[ "${path}" == "~/"* ]]; then
    printf '%s/%s\n' "${HOME}" "${path#~/}"
    return
  fi

  if [[ "${path}" == "${HOME}/~/"* ]]; then
    printf '%s/%s\n' "${HOME}" "${path#"${HOME}/~/"}"
    return
  fi

  if [[ "${path}" == /* ]]; then
    printf '%s\n' "${path}"
    return
  fi

  printf '%s/%s\n' "${HOME}" "${path}"
}

REMOTE_DIR="$(expand_remote_path "${REMOTE_DIR}")"
REMOTE_LOG="$(expand_remote_path "${REMOTE_LOG}")"

if [[ ! -d "${REMOTE_DIR}" ]]; then
  echo "Remote project directory does not exist: ${REMOTE_DIR}" >&2
  exit 1
fi

cd "${REMOTE_DIR}"

echo "Remote directory: ${REMOTE_DIR}"

if sudo systemctl list-unit-files 2>/dev/null | grep -q '^wire-pod\.service'; then
  echo "Detected wire-pod.service, restarting with systemd"
  sudo systemctl restart wire-pod
  sudo systemctl --no-pager --full status wire-pod || true
  exit 0
fi

echo "wire-pod.service not found, falling back to start.sh"

if pgrep -af '/chipper/start\.sh|/chipper($| )|cmd/vosk/main\.go|cmd/experimental/' >/dev/null 2>&1; then
  echo "Stopping existing wire-pod related processes"
  sudo pkill -f '/chipper/start\.sh|/chipper($| )|cmd/vosk/main\.go|cmd/experimental/' || true
  sleep 2
fi

echo "Starting wire-pod with nohup, log: ${REMOTE_LOG}"
nohup sudo ./chipper/start.sh > "${REMOTE_LOG}" 2>&1 < /dev/null &
sleep 3

echo "Recent log output:"
tail -n 40 "${REMOTE_LOG}" || true
EOF

echo "==> Deploy finished"
