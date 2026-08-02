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

# Do NOT use -tt (pseudo-terminal). It turns the HEREDOC into interactive
# line-by-line paste, which breaks command execution and hides errors.
RESTART_SSH_CMD=("${SSH_CMD[@]}")

echo "==> Deploy target: ${REMOTE_HOST}:${REMOTE_DIR}"

# ---------------------------------------------------------------------------
# Step 1: Optionally cross-compile chipper for ARM Linux locally.
# This avoids the VERY slow Go compilation on Raspberry Pi, and more
# importantly prevents systemd start timeout (systemd kills long compiles).
# ---------------------------------------------------------------------------
LOCAL_CHIPPER_SRC="${REPO_ROOT}/chipper"
LOCAL_CHIPPER_BIN="${LOCAL_CHIPPER_SRC}/chipper"
CROSS_BUILT=0

if [[ "${SKIP_SYNC}" -eq 0 ]] && command -v go &>/dev/null && [[ -f "${LOCAL_CHIPPER_SRC}/go.mod" ]]; then
  echo "==> Attempting local cross-compile for ARM Linux..."
  # Load STT_SERVICE and friends from source.sh if available
  if [[ -f "${LOCAL_CHIPPER_SRC}/source.sh" ]]; then
    # shellcheck disable=SC1090
    source "${LOCAL_CHIPPER_SRC}/source.sh"
  fi
  export GOTAGS="nolibopusfile"
  if [[ "${USE_INBUILT_BLE:-}" == "true" ]]; then
    GOTAGS="${GOTAGS},inbuiltble"
  fi
  COMMIT_HASH="$(git -C "${REPO_ROOT}" rev-parse --short HEAD 2>/dev/null || echo 'unknown')"
  export GOLDFLAGS="-X 'github.com/kercre123/wire-pod/chipper/pkg/vars.CommitSHA=${COMMIT_HASH}'"

  case "${STT_SERVICE:-vosk}" in
    vosk)
      echo "STT_SERVICE=vosk, building with CGO (may fail without Pi libvosk)..."
      export CGO_ENABLED=1
      if (cd "${LOCAL_CHIPPER_SRC}" && GOOS=linux GOARCH=arm GOARM=7 \
           go build -tags "${GOTAGS}" -ldflags="${GOLDFLAGS}" \
           -o chipper cmd/vosk/main.go 2>&1); then
        echo "Cross-compile OK: ${LOCAL_CHIPPER_BIN}"
        CROSS_BUILT=1
      else
        echo "Cross-compile failed (expected for CGO deps), will build on remote."
        rm -f "${LOCAL_CHIPPER_BIN}"
      fi
      ;;
    *)
      export CGO_ENABLED=0
      if (cd "${LOCAL_CHIPPER_SRC}" && GOOS=linux GOARCH=arm GOARM=7 \
           go build -tags "${GOTAGS}" -ldflags="${GOLDFLAGS}" \
           -o chipper cmd/vosk/main.go 2>&1); then
        echo "Cross-compile OK: ${LOCAL_CHIPPER_BIN}"
        CROSS_BUILT=1
      else
        echo "Cross-compile failed, will build on remote."
        rm -f "${LOCAL_CHIPPER_BIN}"
      fi
      ;;
  esac
fi

# ---------------------------------------------------------------------------
# Step 2: rsync source (and optionally the cross-compiled binary) to remote
# ---------------------------------------------------------------------------
if [[ "${SKIP_SYNC}" -eq 0 ]]; then
  echo "==> Syncing to ${REMOTE_HOST}:${REMOTE_DIR} ..."

  RSYNC_ARGS=(
    -avz
    --progress
    --exclude ".git"
    --exclude ".DS_Store"
    --exclude ".codegraph"
    --exclude ".trae"
    --exclude ".specify"
    --exclude ".vscode"
    --exclude "scripts/.DS_Store"
    --exclude "chipper/source.sh"
    --exclude "chipper/apiConfig.json"
    --exclude "chipper/jdocs/"
    --exclude "chipper/session-certs/"
    --exclude "certs/server_config.json"
    -e "${SSH_CMD[*]}"
  )

  # Include cross-compiled binary if we built one; otherwise exclude old binary
  if [[ "${CROSS_BUILT}" -eq 1 ]]; then
    RSYNC_ARGS+=(--include "chipper/chipper")
    echo "==> Including pre-built chipper binary in upload"
  else
    RSYNC_ARGS+=(--exclude "chipper/chipper")
  fi

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

# ---------------------------------------------------------------------------
# Step 3: Remote restart — stop old processes, build if needed, start fresh
# ---------------------------------------------------------------------------
echo "==> Restarting remote application"

"${RESTART_SSH_CMD[@]}" "${REMOTE_HOST}" \
  "env REMOTE_DIR='${REMOTE_DIR}' REMOTE_LOG='${REMOTE_LOG}' CROSS_BUILT='${CROSS_BUILT}' bash -s" <<'RESTART_EOF'
set -euo pipefail

expand_remote_path() {
  local path="$1"
  if [[ -z "${path}" || "${path}" == "~" ]]; then
    printf '%s\n' "${HOME}"; return
  fi
  if [[ "${path}" == "~/"* ]]; then
    printf '%s/%s\n' "${HOME}" "${path#~/}"; return
  fi
  if [[ "${path}" == "${HOME}/~/"* ]]; then
    printf '%s/%s\n' "${HOME}" "${path#"${HOME}/~/"}"
    return
  fi
  if [[ "${path}" == /* ]]; then
    printf '%s\n' "${path}"; return
  fi
  printf '%s/%s\n' "${HOME}" "${path}"
}

REMOTE_DIR="$(expand_remote_path "${REMOTE_DIR}")"
REMOTE_LOG="$(expand_remote_path "${REMOTE_LOG}")"

if [[ ! -d "${REMOTE_DIR}" ]]; then
  echo "ERROR: Remote directory not found: ${REMOTE_DIR}" >&2
  exit 1
fi

cd "${REMOTE_DIR}"
echo "Remote directory: ${REMOTE_DIR}"

# Helper: check if wire-pod.service exists and is known to systemd.
# Uses 'systemctl cat' (checks unit file) + 'systemctl status' as fallback.
has_wirepod_service() {
  sudo systemctl cat wire-pod.service >/dev/null 2>&1 && return 0
  sudo systemctl status wire-pod.service >/dev/null 2>&1 && return 0
  return 1
}

# ---- Stop everything first ----
if has_wirepod_service; then
  echo "Stopping wire-pod.service..."
  sudo systemctl stop wire-pod 2>/dev/null || true
fi

if pgrep -af '/chipper/start\.sh|/chipper($|[[:space:]])|cmd/vosk/main\.go|cmd/experimental/' >/dev/null 2>&1; then
  echo "Killing lingering chipper processes..."
  sudo pkill -f '/chipper/start\.sh|/chipper($|[[:space:]])|cmd/vosk/main\.go|cmd/experimental/' || true
  sleep 2
fi

# ---- Build if needed ----
CHIPPER_BIN="${REMOTE_DIR}/chipper/chipper"
if [[ "${CROSS_BUILT}" == "1" ]]; then
  echo "Using pre-built binary: ${CHIPPER_BIN}"
  sudo chmod +x "${CHIPPER_BIN}"
else
  echo "Building chipper on remote (this may take 2-5 minutes on Pi)..."
  sudo rm -f "${CHIPPER_BIN}"
  cd "${REMOTE_DIR}/chipper"
  if [[ -f source.sh ]]; then
    # shellcheck disable=SC1091
    source source.sh
  fi
  export GOTAGS="nolibopusfile"
  if [[ "${USE_INBUILT_BLE:-}" == "true" ]]; then
    GOTAGS="${GOTAGS},inbuiltble"
  fi
  COMMIT_HASH="$(git -C "${REMOTE_DIR}" rev-parse --short HEAD 2>/dev/null || echo 'unknown')"
  export GOLDFLAGS="-X 'github.com/kercre123/wire-pod/chipper/pkg/vars.CommitSHA=${COMMIT_HASH}'"
  if [[ "${STT_SERVICE:-}" == "vosk" ]]; then
    export CGO_ENABLED=1
    export CGO_CFLAGS="-I/root/.vosk/libvosk -I${HOME}/.vosk/libvosk"
    export CGO_LDFLAGS="-L/root/.vosk/libvosk -L${HOME}/.vosk/libvosk -lvosk -ldl -lpthread"
    export LD_LIBRARY_PATH="/root/.vosk/libvosk:${HOME}/.vosk/libvosk:${LD_LIBRARY_PATH:-}"
  fi
  if ! sudo -E /usr/local/go/bin/go build -tags "${GOTAGS}" -ldflags="${GOLDFLAGS}" -o chipper cmd/vosk/main.go; then
    echo "ERROR: Build failed on remote!" >&2
    exit 1
  fi
  echo "Build OK: ${CHIPPER_BIN}"
  cd "${REMOTE_DIR}"
fi

# ---- Start with systemd (preferred) or nohup fallback ----
if has_wirepod_service; then
  echo "Starting wire-pod.service..."
  if ! sudo systemctl start wire-pod; then
    echo "ERROR: systemctl start failed!" >&2
    sudo systemctl --no-pager --full status wire-pod || true
    exit 1
  fi
  sleep 2
  echo "--- Service status ---"
  sudo systemctl --no-pager --full status wire-pod || true
  echo ""
  echo "--- Recent logs ---"
  sudo journalctl -u wire-pod.service --no-pager -n 30 || true
else
  echo "wire-pod.service not found, starting with nohup (log: ${REMOTE_LOG})"
  nohup sudo ./chipper/start.sh > "${REMOTE_LOG}" 2>&1 < /dev/null &
  sleep 4
  echo "--- Recent log output ---"
  tail -n 40 "${REMOTE_LOG}" || true
fi
RESTART_EOF

echo "==> Deploy finished"

# Clean up local cross-compiled binary
if [[ "${CROSS_BUILT}" -eq 1 ]]; then
  rm -f "${LOCAL_CHIPPER_BIN}"
fi
