# Logrotate for `mathsvr`

Log rotation config for the `mathsvr` systemd service, to stop
`/apps/math/mathsvr.log` from growing unbounded.

- Config file: [`mathsvr`](./mathsvr)
- Install target on EC2: `/etc/logrotate.d/mathsvr`

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

| Directive | Value | Meaning |
|---|---|---|
| `daily` | — | Rotate every day |
| `rotate 14` | 14 | Keep ~2 weeks of history, then drop the oldest |
| `compress` + `delaycompress` | — | gzip old files, delay compression by one cycle (safe with `copytruncate`) |
| `copytruncate` | — | Truncate in place, no service restart |
| `dateext` | — | Name rotated files by date: `mathsvr.log-YYYYMMDD` |
| `su mot mot` / `create 0640 mot mot` | — | Operate on / create files as user `mot` |

Tune `rotate N` to the actual disk capacity on the EC2 host.

## Installation (run on EC2)

```bash
# 1. Copy the config into the directory logrotate scans
sudo cp /apps/math/deploy/logrotate/mathsvr /etc/logrotate.d/mathsvr

# 2. logrotate refuses config files that are group/other writable
sudo chown root:root /etc/logrotate.d/mathsvr
sudo chmod 0644 /etc/logrotate.d/mathsvr

# 3. Dry-run: check syntax + see what logrotate would do (no execution)
sudo logrotate -d /etc/logrotate.d/mathsvr

# 4. Force one real run to confirm
sudo logrotate -fv /etc/logrotate.d/mathsvr

# 5. Check the result: rotated file created, original truncated to ~0
ls -lh /apps/math/mathsvr.log*
```

## Decisive test

After a forced rotation, confirm the app **still writes to the original file** —
this proves `copytruncate` works correctly (no "missed" rotation):

```bash
tail -f /apps/math/mathsvr.log
```

If new log lines still flow into `mathsvr.log`, it works.

## Confirm logrotate runs on a schedule

logrotate does not run itself; it is triggered by a system timer/cron:

```bash
# Systems using a systemd timer (Amazon Linux 2023, recent Ubuntu)
systemctl list-timers | grep logrotate

# Older systems using cron.daily
ls -l /etc/cron.daily/logrotate
```

## Later: switch to size-based rotation

If log volume spikes unpredictably, add `maxsize` to rotate early once a
threshold is exceeded (while keeping the `daily` cadence):

```
    maxsize 200M
```

Note: size-based rotation depends on logrotate being invoked often enough. The
default timer usually runs once a day — add a dedicated hourly timer if you need
finer granularity.
