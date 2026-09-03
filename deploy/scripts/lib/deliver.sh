#!/usr/bin/env bash
# deploy/scripts/lib/deliver.sh — sync files to remote

run_deliver() {
  phase_start "DELIVER"

  info "Syncing files to ${DEPLOY_HOST}:${DEST_DIR}..."
  # deploy/ carries everything the remote needs: the lifecycle hooks this
  # pipeline invokes over SSH (pre-deploy/, post-deploy/) plus the artifacts an
  # operator installs on the host (systemd units, logrotate config, nginx
  # config, the `mai` wrapper). They ride along inside deploy/ — do not list
  # them separately or rsync writes them twice.
  #
  # deploy/scripts/ is the one exception: it is the LOCAL orchestration tooling
  # — this very script, its siblings and lib/ — which only ever runs on a
  # developer machine, so it is excluded rather than shipped to the host. The
  # pattern is anchored to the rsync transfer root because -R (relative)
  # preserves the `deploy/` prefix on the wire.
  remote_rsync \
    --exclude='/deploy/scripts/' \
    migrations/ \
    dist/mathsvr \
    keys/ \
    deploy/ \
    docker/

  # docker/ carries the observability stack as code (compose files + Prometheus
  # / Loki / Tempo / Alloy / Grafana configs). It lands at /apps/math/docker/,
  # which is what makes the compose file's `../logs` bind mount resolve to
  # /apps/math/logs — the directory the app writes app.log into.
  #
  # Shipping it means host-side edits to these configs are OVERWRITTEN on the
  # next deploy: treat the repo as the source of truth and re-run
  # `docker compose up -d` on the host to apply changes. Nothing secret lives
  # here — Grafana's admin password comes from the host env at compose time.

  info "Files synced successfully"

  # remote_rsync passes --no-perms so the binary lands with the receiver's
  # default umask (0644). Restore +x on the executable explicitly; the
  # shell scripts are run via `bash <path>` so their mode bit is irrelevant.
  info "Setting executable bit on binary..."
  remote_exec "chmod +x ${DEST_DIR}/dist/mathsvr"

  phase_end "DELIVER"
}
