#!/usr/bin/env bash
# deploy/scripts/deploy.sh — deployment orchestrator
#
# Usage:
#   ./deploy/scripts/deploy.sh <target>              Full deploy (validate -> build -> prepare -> deliver -> activate)
#   ./deploy/scripts/deploy.sh <target> --rollback   Rollback to previous binary
#   ./deploy/scripts/deploy.sh <target> --skip-build Skip build phase (use existing binary)
#
# Examples:
#   ./deploy/scripts/deploy.sh t1
#   ./deploy/scripts/deploy.sh t2 --skip-build
#   ./deploy/scripts/deploy.sh t1 --rollback

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"

cd "$PROJECT_DIR"

# Source shared library and phase scripts
source "$SCRIPT_DIR/lib/common.sh"
source "$SCRIPT_DIR/lib/validate.sh"
source "$SCRIPT_DIR/lib/build.sh"
source "$SCRIPT_DIR/lib/prepare.sh"
source "$SCRIPT_DIR/lib/deliver.sh"
source "$SCRIPT_DIR/lib/activate.sh"
source "$SCRIPT_DIR/lib/rollback.sh"

# ── Parse arguments ──────────────────────────────────────

DEPLOY_TARGET="${1:-}"
SKIP_BUILD=false
DO_ROLLBACK=false

[[ -z "$DEPLOY_TARGET" ]] && fatal "Usage: $0 <p1|p2|t1|t2|t3|t4> [--skip-build|--rollback]"

shift
for arg in "$@"; do
  case "$arg" in
    --skip-build) SKIP_BUILD=true ;;
    --rollback)   DO_ROLLBACK=true ;;
    *)            fatal "Unknown flag: $arg" ;;
  esac
done

# ── Load credentials ─────────────────────────────────────

load_credentials "$DEPLOY_TARGET"

# ── Confirmation ─────────────────────────────────────────

GIT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
GIT_SHA=$(git rev-parse --short HEAD)

if $DO_ROLLBACK; then
  echo ""
  echo "  ROLLBACK: $DEPLOY_TARGET ($DEPLOY_HOST)"
  echo ""
  read -rp "Confirm rollback? [y/N] " confirm
  [[ "$confirm" =~ ^[Yy]$ ]] || { info "Cancelled."; exit 0; }

  run_rollback
  info "Rollback complete."
  exit 0
fi

echo ""
echo "  DEPLOY"
echo "  ├─ Branch:  $GIT_BRANCH ($GIT_SHA)"
echo "  ├─ Target:  $DEPLOY_TARGET → $DEPLOY_HOST"
echo "  └─ Build:   $( $SKIP_BUILD && echo 'SKIP (use existing)' || echo 'yes' )"
echo ""
read -rp "Proceed? [y/N] " confirm
[[ "$confirm" =~ ^[Yy]$ ]] || { info "Cancelled."; exit 0; }

# ── Deploy log header ────────────────────────────────────

info "═══════════════════════════════════════════════════"
info "DEPLOY START  branch=$GIT_BRANCH  commit=$GIT_SHA  target=$DEPLOY_TARGET"
info "═══════════════════════════════════════════════════"

DEPLOY_START=$(date +%s)

# ── Execute phases ───────────────────────────────────────

run_validate

if ! $SKIP_BUILD; then
  run_build
fi

run_prepare
run_deliver
run_activate

# ── Summary ──────────────────────────────────────────────

DEPLOY_END=$(date +%s)
DURATION=$((DEPLOY_END - DEPLOY_START))

info "═══════════════════════════════════════════════════"
info "DEPLOY COMPLETE  duration=${DURATION}s  target=$DEPLOY_TARGET"
info "═══════════════════════════════════════════════════"
