#!/usr/bin/env bash
# bin/lib/rollback.sh — restore previous binary and restart

run_rollback() {
  phase_start "ROLLBACK"

  info "Stopping current server..."
  remote_exec "
    pkill -x monexsvr 2>/dev/null || true
    sleep 2
  "

  info "Restoring previous binary..."
  remote_exec "
    if [ -f $DEST_DIR/dist/monexsvr.bak ]; then
      cp $DEST_DIR/dist/monexsvr.bak $DEST_DIR/dist/monexsvr
      echo 'Binary restored from backup'
    else
      echo 'ERROR: No backup found at $DEST_DIR/dist/monexsvr.bak'
      exit 1
    fi
  "

  info "Starting restored server..."
  remote_exec "bash $DEST_DIR/post-deploy/015_start-server.sh"
  remote_exec "bash $DEST_DIR/post-deploy/025_restart-nginx.sh"

  info "Verifying rollback..."
  sleep 3
  if remote_exec "pgrep -x monexsvr > /dev/null"; then
    info "Rollback successful — server is running"
  else
    error "Rollback failed — server is not running"
    exit 1
  fi

  phase_end "ROLLBACK"
}
