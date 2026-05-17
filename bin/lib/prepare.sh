#!/usr/bin/env bash
# bin/lib/prepare.sh — stop server and backup on remote

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

  # Run existing pre-deploy scripts (reuses current scripts)
  info "Running pre-deploy scripts..."
  remote_exec "
    for i in $DEST_DIR/pre-deploy/*.sh; do
      [ -f \"\$i\" ] && { echo \"Running \$i...\"; bash \"\$i\"; }
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
