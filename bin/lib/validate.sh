#!/usr/bin/env bash
# bin/lib/validate.sh — preflight checks

run_validate() {
  phase_start "VALIDATE"

  # Check git status
  GIT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
  GIT_SHA=$(git rev-parse --short HEAD)
  info "Branch: $GIT_BRANCH  Commit: $GIT_SHA"

  # Warn on uncommitted changes
  if [[ -n "$(git status --porcelain)" ]]; then
    warn "Working tree has uncommitted changes"
  fi

  # Verify SSH connectivity
  info "Testing SSH to ${DEPLOY_USER}@${DEPLOY_HOST}..."
  remote_exec "echo 'SSH OK'" || fatal "Cannot reach remote host"

  # Verify remote directory exists
  remote_exec "test -d $DEST_DIR" || fatal "Remote directory $DEST_DIR does not exist"

  phase_end "VALIDATE"
}
