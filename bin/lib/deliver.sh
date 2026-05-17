#!/usr/bin/env bash
# bin/lib/deliver.sh — sync files to remote

run_deliver() {
  phase_start "DELIVER"

  info "Syncing files to ${DEPLOY_HOST}:${DEST_DIR}..."
  remote_rsync \
    migrations/ \
    assets/views/ \
    assets/images/ \
    pre-deploy/ \
    post-deploy/ \
    dist/mathsvr

  info "Files synced successfully"

  phase_end "DELIVER"
}
