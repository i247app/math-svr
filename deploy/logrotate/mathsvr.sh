# logrotate config for the mathsvr systemd service log.
# Install to: /etc/logrotate.d/mathsvr  (see README.md in this directory)
#
# Strategy: copytruncate — systemd opens mathsvr.log with `append:` and holds
# the file descriptor, so we truncate the original inode in place instead of
# renaming it. This avoids restarting the service (which would trigger the
# app's expensive graceful shutdown: drain jobs + serialize sessions).
#
# Rotation trigger: size-based (100M). NOTE: `size` only takes effect when
# logrotate actually runs. The default system timer runs logrotate once a day,
# which is too coarse for a size threshold, so a dedicated hourly timer invokes
# logrotate on this config — see mathsvr-logrotate.{service,timer} and README.md.

/apps/math/mathsvr.log {
    size 100M
    rotate 14
    missingok
    notifempty
    compress
    delaycompress
    copytruncate
    su mot mot
    create 0640 mot mot
}

# The JSON log (LOG_FILE=./logs/app.log) that Grafana Alloy tails and ships to
# Loki. This is a DIFFERENT stream from mathsvr.log above: that one is the
# console (stdout/stderr captured by systemd), this one is the structured log
# the app writes itself.
#
# copytruncate is mandatory here, for a different reason than above: the Go app
# holds its own append-only FD (logger.openLogFile) and has no log-reopen
# signal — SIGHUP does not reopen the file, it kills the process (see
# deploy/scripts/verify-graceful-shutdown.sh). Renaming the file would leave
# the app writing into the rotated inode forever, so we truncate in place.
#
# Alloy tolerates this: on truncation it detects the file shrank and resumes
# from offset 0. A handful of lines can be lost in the rotation window — the
# accepted trade for not restarting the app.
#
# Mode 0644 (not 0640) is deliberate: the Alloy container reads this file
# through the ../logs bind mount as a different uid than `mot`, and needs world
# read. It matches the 0644 the app itself creates the file with.
/apps/math/logs/*.log {
    size 100M
    rotate 7
    missingok
    notifempty
    compress
    delaycompress
    copytruncate
    su mot mot
    create 0644 mot mot
}
