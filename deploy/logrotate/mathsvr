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
