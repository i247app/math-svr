#!/usr/bin/env bash
# deploy/scripts/lib/common.sh — shared deployment utilities
# Sourced by all phase scripts. Never executed directly.

set -euo pipefail

# ── Logging ──────────────────────────────────────────────

readonly LOG_FILE="deploy.log"

log()   { local ts; ts=$(date '+%Y-%m-%d %H:%M:%S'); echo "[$ts] $*" | tee -a "$LOG_FILE"; }
info()  { log "INFO  $*"; }
warn()  { log "WARN  $*"; }
error() { log "ERROR $*"; }
fatal() { error "$*"; exit 1; }

phase_start() { info "━━━ PHASE: $1 ━━━"; }
phase_end()   { info "━━━ PHASE: $1 complete ━━━"; }

# ── Credential Loading ───────────────────────────────────

CRED_FILE=".env.ec2-credentials"

load_credentials() {
  [[ -f "$CRED_FILE" ]] || fatal "$CRED_FILE not found"

  # Use exact key matching with awk to avoid HOST matching HOST1
  SSH_KEY=$(awk -F= '/^SSH_KEY=/{print $2; exit}' "$CRED_FILE")
  DEPLOY_USER=$(awk -F= '/^USER=/{print $2; exit}' "$CRED_FILE")

  local host_var
  case "${1:-}" in
    t1) host_var="HOST1" ;;
    t2) host_var="HOST2" ;;
    t3) host_var="HOST3" ;;
    t4) host_var="HOST4" ;;
    *)  host_var="HOST"
        DEPLOY_HOST=$(awk -F= "/^${host_var}=/{print \$2; exit}" "$CRED_FILE")
        [[ -n "$DEPLOY_HOST" ]] || fatal "Unknown target: ${1:-<empty>}. Use t1|t2|t3|t4 or set HOST in $CRED_FILE" ;;
  esac

  DEPLOY_HOST=$(awk -F= "/^${host_var}=/{print \$2; exit}" "$CRED_FILE")

  [[ -n "$SSH_KEY" ]]     || fatal "SSH_KEY not set in $CRED_FILE"
  [[ -n "$DEPLOY_USER" ]] || fatal "USER not set in $CRED_FILE"
  [[ -n "$DEPLOY_HOST" ]] || fatal "$host_var not set in $CRED_FILE"

  export SSH_KEY DEPLOY_USER DEPLOY_HOST
}

# ── Remote Execution Helpers ─────────────────────────────

DEST_DIR="/apps/math"

remote_exec() {
  ssh -o ConnectTimeout=10 -i "$SSH_KEY" "${DEPLOY_USER}@${DEPLOY_HOST}" "$@"
}

remote_rsync() {
  rsync -avz -R -O --no-perms --progress \
    -e "ssh -o ConnectTimeout=10 -i $SSH_KEY" \
    "$@" "${DEPLOY_USER}@${DEPLOY_HOST}:${DEST_DIR}"
}
