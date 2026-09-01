#!/usr/bin/env bash
# deploy/scripts/lib/prepare.sh — stop server and backup on remote

run_prepare() {
  phase_start "PREPARE"

  # Backup current binary for rollback
  info "Backing up current binary..."
  remote_exec "
    if [ -f $DEST_DIR/dist/mathsvr ]; then
      cp $DEST_DIR/dist/mathsvr $DEST_DIR/dist/mathsvr.bak
      echo 'Backup created: mathsvr.bak'
    else
      echo 'No existing binary to back up'
    fi
  "

  # Run existing pre-deploy scripts (reuses current scripts).
  # nullglob makes an unmatched glob expand to nothing instead of staying
  # literal, so a fresh host with no remote pre-deploy/ directory yields
  # an empty loop instead of a failing `[ -f "<glob>" ]` test that would
  # propagate non-zero back to the local set -e and abort the deploy.
  #
  # The directory is delivered as part of deploy/, so it only appears on the
  # host after the first DELIVER — and PREPARE runs *before* DELIVER. An
  # absent directory therefore means the stop hook did not run and the server
  # is still up, which the force-kill below would paper over silently. Say so
  # loudly instead.
  info "Running pre-deploy scripts..."
  remote_exec "
    if [ ! -d $DEST_DIR/deploy/pre-deploy ]; then
      echo 'WARNING: $DEST_DIR/deploy/pre-deploy is missing — no graceful stop will run.'
      echo 'WARNING: on a host that predates the deploy/ layout, run once:'
      echo 'WARNING:   mkdir -p $DEST_DIR/deploy && mv $DEST_DIR/pre-deploy $DEST_DIR/post-deploy $DEST_DIR/deploy/'
    fi
    shopt -s nullglob
    for i in $DEST_DIR/deploy/pre-deploy/*.sh; do
      echo \"Running \$i...\"
      bash \"\$i\"
    done
  "

  # Wait for process to fully exit
  info "Waiting for server process to exit..."
  remote_exec "
    for attempt in 1 2 3 4 5 6 7 8 9 10; do
      if ! pgrep -x mathsvr > /dev/null 2>&1; then
        echo 'Server stopped'
        exit 0
      fi
      echo \"Still running... attempt \$attempt/10\"
      sleep 1
    done
    echo 'WARNING: Server did not stop within 10s, force killing...'
    pkill -9 -x mathsvr 2>/dev/null || true
    sleep 1
  "

  phase_end "PREPARE"
}
