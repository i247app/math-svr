# systemd unit for `mathsvr`

The main systemd service unit that runs the MathAI Go server (`mathsvr`) as a
managed, auto-restarting, boot-time service on the EC2 host, plus the operator
wrapper used to drive it by hand.

- Unit file: [`mathsvr.service`](./mathsvr.service)
  → `/etc/systemd/system/mathsvr.service`
- Operator script: [`alias_mai.sh`](./alias_mai.sh) — exposed as the `mai`
  command; see [§ Operator script](#operator-script--alias_maish) below

## Background

`mathsvr` is a long-running Go HTTP server (`Type=simple`). It also runs the
in-process job runtime and the metrics listener in the same process. This unit
governs its lifecycle: start on boot, restart on crash, and — critically — a
**graceful** stop.

The unit is deliberately aligned with the app's runtime behaviour:

- **`KillSignal=SIGTERM` + `TimeoutStopSec=30`** — on stop, systemd sends
  SIGTERM (not SIGKILL) and waits up to 30s. The app uses SIGTERM to run its
  graceful shutdown: drain the job runtime and serialize active sessions to
  `SERIALIZED_SESSION_FILE`. Using SIGKILL would lose in-flight work and active
  sessions.
- **`WorkingDirectory=/apps/math`** — the app resolves `.env`, `hmac.key`,
  `keys/`, `migrations/`, and log paths relative to this directory. A wrong
  working directory makes boot fail (missing `.env` / `hmac.key`).
- **`User=mot` / `Group=mot`** — runs unprivileged (least privilege).
- **Sandboxing** (`NoNewPrivileges`, `ProtectSystem=full`, `ProtectHome`,
  `PrivateTmp`) — keep all runtime paths under `/apps/math`, which is outside the
  read-only/hidden sandbox regions. Pointing `LOG_FILE`, cert paths, or the
  session file outside `/apps/math` (e.g. into `/etc` or `/home`) would be
  blocked by the sandbox.

## Key directives

| Directive | Value | Meaning |
|---|---|---|
| `Type=simple` | — | Considered started as soon as `ExecStart` execs |
| `After` / `Requires=mysql.service` | — | Start after MySQL is up; stop if MySQL stops. NOTE: guarantees MySQL is *started*, not that it *accepts connections* yet — see below |
| `Restart=on-failure` / `RestartSec=5s` | — | Auto-restart on non-zero exit, wait 5s (no restart on a clean `systemctl stop`) |
| `StandardOutput` / `StandardError` | `append:/apps/math/mathsvr.log` | Console stream appended to the log file (rotated by `../logrotate/`) |
| `WantedBy=multi-user.target` | — | Auto-start at boot once enabled |

> **MySQL readiness race:** `Requires`/`After` only ensure MySQL has *started*,
> not that it is ready to accept connections. The app does a 5s `PingContext` at
> boot and **fails boot if the DB is unreachable** — but `Restart=on-failure`
> retries after 5s, so it self-heals in practice.

## Installation (run on EC2)

```bash
# 1. Copy the unit into the systemd directory
sudo cp /apps/math/deploy/app-service/mathsvr.service /etc/systemd/system/mathsvr.service
sudo chown root:root /etc/systemd/system/mathsvr.service
sudo chmod 0644 /etc/systemd/system/mathsvr.service

# 2. Reload systemd so it picks up the new/changed unit
sudo systemctl daemon-reload

# 3. Enable (auto-start at boot) and start it now
sudo systemctl enable --now mathsvr.service

# 4. Verify
sudo systemctl status mathsvr.service --no-pager
```

After editing the unit later, always re-run `sudo systemctl daemon-reload`
before `restart`, otherwise systemd keeps the old definition.

## Graceful-stop test

```bash
# Stop sends SIGTERM; confirm the app logs its shutdown sequence
sudo systemctl stop mathsvr.service
tail -n 50 /apps/math/mathsvr.log     # look for drain / session-serialize lines
```

## Operator script — `alias_mai.sh`

A thin wrapper around the `systemctl` verbs above, so an operator can type
`mai start` instead of `sudo systemctl start mathsvr.service`. It lives next to
the unit it drives.

| Command | Action |
|---|---|
| `mai start` | Start the server (via systemd) |
| `mai stop` | Stop the server gracefully (SIGTERM) |
| `mai restart` | Restart the server |
| `mai status` | Show service status |
| `mai watch` | Tail `/apps/math/mathsvr.log` in real time |

### Installation (run on EC2)

```bash
# 1. Make it executable
chmod +x /apps/math/deploy/app-service/alias_mai.sh

# 2. Expose it as the `mai` command. Either symlink it onto PATH...
sudo ln -sf /apps/math/deploy/app-service/alias_mai.sh /usr/local/bin/mai

# 3. ...or add a shell alias (append to ~/.bashrc, then re-login):
#    alias mai='/apps/math/deploy/app-service/alias_mai.sh'

# 4. Verify
mai status
```

The script calls `sudo systemctl ...` internally, so the operator needs sudo
rights for the `mathsvr.service` control verbs. `mai watch` tails the console
log file directly (`/apps/math/mathsvr.log`) — the same file rotated by
[`../logrotate/`](../logrotate/).

## Related

- Log rotation for `mathsvr.log`: [`../logrotate/`](../logrotate/)
- Local deploy tooling that ships this directory to the host:
  [`../scripts/`](../scripts/) (never installed on the host itself)
