# Logrotate for `mathsvr`

Log rotation config for the `mathsvr` systemd service, to stop
`/apps/math/mathsvr.log` from growing unbounded.

- Config file: [`mathsvr`](./mathsvr) → `/etc/logrotate.d/mathsvr`
- Timer unit: [`mathsvr-logrotate.timer`](./mathsvr-logrotate.timer) → `/etc/systemd/system/`
- Service unit: [`mathsvr-logrotate.service`](./mathsvr-logrotate.service) → `/etc/systemd/system/`

## Background

The `mathsvr` service writes stdout/stderr to `/apps/math/mathsvr.log` via
systemd:

```ini
StandardOutput=append:/apps/math/mathsvr.log
StandardError=append:/apps/math/mathsvr.log
```

Because systemd opens the file with `append:` and **holds the file
descriptor**, this config uses `copytruncate` (copy the contents to the rotated
file, then truncate the original in place) instead of renaming the file. That
avoids restarting the service — which would trigger the app's expensive
graceful shutdown (drain the job runtime + serialize sessions to
`SERIALIZED_SESSION_FILE`).

> This log is the **console** stream (stdout/stderr). It is DIFFERENT from
> `LOG_FILE` in `.env` (the JSON file consumed by Loki/Alloy). This config only
> rotates `mathsvr.log`.

## Current policy

Rotation is **size-based (100M)**, checked hourly by a dedicated systemd timer.

| Directive | Value | Meaning |
|---|---|---|
| `size 100M` | 100M | Rotate when the file reaches 100M (only when logrotate runs — see below) |
| `rotate 14` | 14 | Keep 14 old files, then drop the oldest (up to ~1.4G total: 14 × 100M before compression) |
| `compress` + `delaycompress` | — | gzip old files, delay compression by one cycle (safe with `copytruncate`) |
| `copytruncate` | — | Truncate in place, no service restart |
| `su mot mot` / `create 0640 mot mot` | — | Operate on / create files as user `mot` |

Rotated files use numbered suffixes (`mathsvr.log.1`, `mathsvr.log.2.gz`, …).
`dateext` is intentionally omitted: with sub-daily size-based rotation the
default date suffix (`-YYYYMMDD`) would collide when the file rotates more than
once a day.

Tune `rotate N` to the actual disk capacity on the EC2 host.

### Why a dedicated timer

`size` only triggers a rotation **when logrotate actually runs**. The system-wide
logrotate timer runs once a day, which is too coarse for a size threshold — the
file could grow well past 100M between daily checks. So `mathsvr-logrotate.timer`
runs `mathsvr-logrotate.service` **every hour**, which invokes logrotate against
this config. It uses the default shared logrotate state file, so it stays
consistent with the daily system run (no double rotation).

## Installation (run on EC2)

```bash
# 1. Copy the logrotate config into the directory logrotate scans
sudo cp /apps/math/deploy/logrotate/mathsvr /etc/logrotate.d/mathsvr

# 2. logrotate refuses config files that are group/other writable
sudo chown root:root /etc/logrotate.d/mathsvr
sudo chmod 0644 /etc/logrotate.d/mathsvr

# 3. Install the timer + service units that enforce the size threshold hourly
sudo cp /apps/math/deploy/logrotate/mathsvr-logrotate.service /etc/systemd/system/
sudo cp /apps/math/deploy/logrotate/mathsvr-logrotate.timer   /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now mathsvr-logrotate.timer

# 4. Dry-run: check syntax + see what logrotate would do (no execution)
sudo logrotate -d /etc/logrotate.d/mathsvr

# 5. Confirm the timer is scheduled and the service works
systemctl list-timers | grep mathsvr-logrotate
sudo systemctl start mathsvr-logrotate.service   # run the rotation check once now
journalctl -u mathsvr-logrotate.service --no-pager

# 6. Check the result once the file has exceeded 100M and a rotation has run:
#    rotated file created, original truncated to ~0
ls -lh /apps/math/mathsvr.log*
```

> `logrotate -fv` (force) still works for a manual one-off rotation regardless
> of size, but is not needed for normal operation — the timer handles it.

## Decisive test

After a forced rotation, confirm the app **still writes to the original file** —
this proves `copytruncate` works correctly (no "missed" rotation):

```bash
tail -f /apps/math/mathsvr.log
```

If new log lines still flow into `mathsvr.log`, it works.

## Tuning the cadence

The hourly timer bounds worst-case overshoot to ~1 hour of logs above 100M. If
the log grows fast enough to overshoot within an hour, tighten the timer:

```ini
# in mathsvr-logrotate.timer
OnCalendar=*:0/15   # every 15 minutes
```

Then `sudo systemctl daemon-reload && sudo systemctl restart mathsvr-logrotate.timer`.

Conversely, `rotate 14` at 100M each can use up to ~1.4G before compression —
lower it if disk is tight on the EC2 host.
