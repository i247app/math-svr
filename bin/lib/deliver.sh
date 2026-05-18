#!/usr/bin/env bash
# bin/lib/deliver.sh — sync files to remote

run_deliver() {
  phase_start "DELIVER"

  info "Syncing files to ${DEPLOY_HOST}:${DEST_DIR}..."
  remote_rsync \
    migrations/ \
    pre-deploy/ \
    post-deploy/ \
    dist/mathsvr

  info "Files synced successfully"

  # remote_rsync passes --no-perms so the binary lands with the receiver's
  # default umask (0644). Restore +x on the executable explicitly; the
  # shell scripts are run via `bash <path>` so their mode bit is irrelevant.
  info "Setting executable bit on binary..."
  remote_exec "chmod +x ${DEST_DIR}/dist/mathsvr"

  phase_end "DELIVER"
}
