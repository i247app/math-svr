# Deployment Process — `make deploy`

This document explains exactly what happens when a developer runs:

```bash
make deploy RHOST=<target>
```

against a `math-svr` checkout. It covers the local orchestrator (`deploy/scripts/`), the remote pre-deploy hooks (`deploy/pre-deploy/`), and the remote post-deploy hooks (`deploy/post-deploy/`).

---

## 1. Entry point — the Makefile

`make deploy` is a thin wrapper:

```makefile
deploy:
	@./deploy/scripts/deploy.sh $(RHOST)
```

Related targets that hit the same script with flags:

| Target | Command | Purpose |
|---|---|---|
| `make deploy RHOST=t1` | `deploy/scripts/deploy.sh t1` | full pipeline: validate → build → prepare → deliver → activate |
| `make deploy-quick RHOST=t1` | `deploy/scripts/deploy.sh t1 --skip-build` | reuse the binary already in `dist/` |
| `make deploy-rollback RHOST=t1` | `deploy/scripts/deploy.sh t1 --rollback` | restore the previous binary on the remote |
| `make deploy-amd RHOST=t1` | `BUILD_ARCH=amd64 ./deploy/scripts/deploy.sh t1` | force AMD64 build |

`RHOST` selects which EC2 host to target. Valid values: `t1`, `t2`, `t3`, `t4` (mapped to `HOST1..HOST4` in `.env.ec2-credentials`).

---

## 2. The orchestrator — `deploy/scripts/deploy.sh`

`deploy/scripts/deploy.sh` is the single source of truth. It sources six phase libraries from `deploy/scripts/lib/` and runs them in order:

```
deploy/scripts/deploy.sh
├── lib/common.sh    (logging, credentials, remote helpers)
├── lib/validate.sh  → run_validate
├── lib/build.sh     → run_build       (skipped with --skip-build)
├── lib/prepare.sh   → run_prepare
├── lib/deliver.sh   → run_deliver
├── lib/activate.sh  → run_activate
└── lib/rollback.sh  → run_rollback    (only with --rollback)
```

Flow:

1. **Parse args.** Read `<target>` plus optional `--skip-build` / `--rollback`.
2. **Load credentials** via `load_credentials()` from `.env.ec2-credentials`:
   - `SSH_KEY` — path to the private key
   - `USER` (`DEPLOY_USER`) — SSH user
   - `HOST` / `HOST1..4` (`DEPLOY_HOST`) — selected by the `t1..t4` argument
3. **Show confirmation banner** with branch, short SHA, target host, and whether build is enabled. Aborts unless the user answers `y`.
4. **Append a `DEPLOY START` line** to `deploy.log` (timestamped via `common.sh`).
5. **Run phases sequentially.** Any failure exits the script (`set -euo pipefail`).
6. **Print duration + `DEPLOY COMPLETE`.**

`common.sh` also defines the two remote primitives every phase uses:

```bash
remote_exec  "<cmd>"           # ssh into $DEPLOY_USER@$DEPLOY_HOST and run <cmd>
remote_rsync <local paths...>  # rsync -avz -R --no-perms -O --progress to ${DEST_DIR}
```

`DEST_DIR` on the remote is hard-coded to **`/apps/math`**.

---

## 3. Phase-by-phase walkthrough

### Phase 1 — VALIDATE (`deploy/scripts/lib/validate.sh`)

Preflight only — no mutations.

- Records the current branch + commit SHA into `deploy.log`.
- Warns (does not fail) if the working tree has uncommitted changes.
- SSHes to the remote and runs `echo 'SSH OK'` to confirm reachability.
- Confirms `/apps/math` exists on the remote. Fails fast if not (use `deploy/scripts/init_remote.sh` to bootstrap a new host).

### Phase 2 — BUILD (`deploy/scripts/lib/build.sh`)

Skipped when `--skip-build` is passed.

- Compiles a static Linux binary into `dist/mathsvr`:
  ```bash
  GOOS=linux GOARCH="${BUILD_ARCH:-arm64}" go build -o dist/mathsvr ./cmd/mathsvr
  ```
- Verifies the binary exists and is non-empty.
- Runs `file dist/mathsvr` and asserts the architecture string matches `BUILD_ARCH` (aarch64/arm64 for arm64, x86-64/amd64 for amd64). This catches stale cross-compile caches.

Default architecture is **arm64** because production EC2 hosts are Graviton.

### Phase 3 — PREPARE (`deploy/scripts/lib/prepare.sh`)

Runs on the remote via `remote_exec`. This is where the server is stopped.

1. **Backup current binary.** Copies `/apps/math/dist/mathsvr` → `/apps/math/dist/mathsvr.bak` (if it exists). This is what `--rollback` later restores.
2. **Run remote `/apps/math/deploy/pre-deploy/*.sh` scripts** in lexicographic order, with `shopt -s nullglob` so a fresh host (no scripts yet) does not abort. Currently there is one script — see §5. Because the hooks are delivered *inside* `deploy/`, and PREPARE runs *before* DELIVER, a host that has never received the `deploy/` layout has no stop hook yet — the phase prints a `WARNING` naming the one-time `mv` that fixes it (see §9) rather than silently skipping the graceful stop.
3. **Wait for the `mathsvr` process to exit**, polling `pgrep -x mathsvr` up to 10 times (1 s apart). If it is still alive after 10 s, it is force-killed with `pkill -9 -x mathsvr`. The exact-name match (`-x`) ensures `mathsvr-amd64` is not accidentally killed.

After this phase, the server is guaranteed to be down on the remote.

### Phase 4 — DELIVER (`deploy/scripts/lib/deliver.sh`)

Copies the new artifacts up.

- `rsync -avz -R --no-perms -O --progress` for:
  - `migrations/` — applied by the server on startup (forward-only)
  - `dist/mathsvr` — the new binary itself
  - `keys/` — the HMAC / handshake key material the app reads at boot
  - `deploy/` — everything the host needs, in one subtree:
    - `deploy/pre-deploy/` — stop hooks, invoked over SSH during PREPARE
    - `deploy/post-deploy/` — start / health hooks, invoked over SSH during ACTIVATE
    - `deploy/app-service/` — systemd unit + the `mai` operator wrapper
    - `deploy/logrotate/` — logrotate config + timer
    - `deploy/nginx/` — nginx config

  The hooks ride along inside `deploy/` and are **not** listed separately —
  doing so would make rsync transfer them twice, to two different paths.
- **`--exclude='/deploy/scripts/'`** — `deploy/scripts/` is the *local* orchestration tooling (this pipeline itself: `deploy.sh`, `lib/`, `login.sh`, `connect-mysql.sh`, …). It only ever runs on a developer machine, so it is deliberately kept off the host. The pattern is anchored to the rsync transfer root because `-R` preserves the `deploy/` prefix on the wire.
- `--no-perms` means the binary lands with the remote's default umask (typically 0644). The script then runs `chmod +x /apps/math/dist/mathsvr` so it is executable.

Everything else under `/apps/math` (`.env`, `keys/`, `data/`, `log_file`, `mathsvr.log`, `mathsvr.pid`) is left untouched.

### Phase 5 — ACTIVATE (`deploy/scripts/lib/activate.sh`)

Runs every script under `/apps/math/deploy/post-deploy/*.sh` in lexicographic order via `remote_exec`. The numeric prefixes (`008_`, `012_`, `015_`, `025_`) define the order — see §6.

If any script returns non-zero, the orchestrator logs `Post-deploy scripts failed` and exits non-zero.

### Phase 6 — final summary

`deploy/scripts/deploy.sh` prints the elapsed duration and writes `DEPLOY COMPLETE` to `deploy.log`. There is no implicit smoke test against the public endpoint — health checking is delegated to the post-deploy scripts (currently they only verify the PID is alive; an HTTP probe would be a good future addition).

---

## 4. Rollback path — `make deploy-rollback`

Driven by `deploy/scripts/lib/rollback.sh`:

1. Asks for a `y/N` confirmation.
2. `pkill -x mathsvr` on the remote, then 2 s sleep.
3. Restores `dist/mathsvr.bak` over `dist/mathsvr`. If the backup is missing, the rollback aborts.
4. Re-runs the two key post-deploy scripts directly: `015_start-server.sh` (start), `025_restart-nginx.sh` (reload nginx).
5. After 3 s, verifies `pgrep -x mathsvr` returns a PID; otherwise exits with `Rollback failed`.

Note: rollback **does not** re-run `008_` (session-file perms) or `012_` (cert sync). That is intentional — those steps are idempotent and the file state has not changed since the last successful deploy.

---

## 5. The `deploy/pre-deploy/` directory (runs on the remote during PREPARE)

| File | What it does |
|---|---|
| `0_stop-server.sh` | Stops the running server **gracefully**. |

`0_stop-server.sh` mechanics:

- Reads PID from `/apps/math/mathsvr.pid` (written by `015_start-server.sh` on the previous deploy). If that PID is dead or the file is missing, falls back to `pgrep -x mathsvr` (exact name match — never matches `mathsvr-amd64`).
- Sends **SIGTERM** (`kill -TERM`). The comment in the script is load-bearing: the `gex` server only subscribes to `SIGINT`/`SIGTERM` via `signal.NotifyContext`. SIGHUP would skip the `onShutdown` hooks that serialize sessions to `data/*.dat` — using SIGHUP would corrupt session state.
- Prints `ps -ef | grep mathsvr` before and after for log evidence.

The hard kill (`pkill -9`) only happens later, in `prepare.sh`, after 10 s of waiting for graceful exit.

---

## 6. The `deploy/post-deploy/` directory (runs on the remote during ACTIVATE)

Lexicographic order matters. Numeric prefixes leave gaps so new steps can slot in (e.g. a future `020_run-migrations.sh`).

| File | Order | What it does |
|---|---|---|
| `008_fix-sessionfile-permissions.sh` | 1st | `chmod 775 /apps/math/data/*.dat` so the next process can read/write the session files written by the previous run. |
| `012_sync-certs.sh` | 2nd | Copies the latest Let's Encrypt certs into `/apps/math/keys/`. Loops over six slots (`t0..t5`); sources `fullchain.pem` and `privkey.pem` from `/etc/letsencrypt/live/x21.i247.com`. Uses `sudo rsync --chown=mot:mot --chmod=400` so the new server boots against fresh certs without manual intervention. |
| `015_start-server.sh` | 3rd | Starts the server: `./dist/mathsvr >> /apps/math/mathsvr.log 2>&1 & echo $! > /apps/math/mathsvr.pid`. Then `ps -p $(cat mathsvr.pid)` confirms the process is alive and exits non-zero if not. This PID file is the one `0_stop-server.sh` consults on the next deploy. |
| `025_restart-nginx.sh` | 4th | `sudo systemctl restart nginx` — picks up any nginx config changes and any cert rotation from `012_`. |

There is **no migration step** in this list. Migrations are run automatically by the Go binary itself on boot (`database.Migrate(ctx, sqlDB, "migrations")` in `internal/bootstrap`), reading the freshly-rsynced `migrations/` directory. The `schema_migrations` table makes the runner idempotent.

---

## 7. End-to-end timeline

```
LOCAL                                                  REMOTE (/apps/math)

make deploy RHOST=t1
└─ deploy/scripts/deploy.sh t1
   ├─ load .env.ec2-credentials  →  SSH_KEY / USER / HOST1
   ├─ confirm prompt
   │
   ├─ VALIDATE  ──ssh──▶  echo SSH OK  /  test -d /apps/math
   │
   ├─ BUILD     (local)   GOOS=linux GOARCH=arm64 go build -o dist/mathsvr
   │
   ├─ PREPARE  ──ssh──▶  cp dist/mathsvr dist/mathsvr.bak
   │           ──ssh──▶  bash deploy/pre-deploy/0_stop-server.sh   (SIGTERM)
   │           ──ssh──▶  poll pgrep -x mathsvr, force-kill after 10 s
   │
   ├─ DELIVER  ──rsync──▶  migrations/ keys/ dist/mathsvr deploy/
   │                       (deploy/ minus deploy/scripts/)
   │           ──ssh──▶  chmod +x dist/mathsvr
   │
   └─ ACTIVATE ──ssh──▶  bash deploy/post-deploy/008_fix-sessionfile-permissions.sh
               ──ssh──▶  bash deploy/post-deploy/012_sync-certs.sh
               ──ssh──▶  bash deploy/post-deploy/015_start-server.sh   (writes mathsvr.pid)
               ──ssh──▶  bash deploy/post-deploy/025_restart-nginx.sh
                          │
                          └─ mathsvr boots:
                             - load .env
                             - apply pending migrations/*.sql
                             - open MySQL pool (TLS 1.2)
                             - start gex.Server on $SERVER_PORT
```

Total per-deploy downtime is bounded by (graceful shutdown wait + rsync + start) — typically a few seconds.

---

## 8. Files / state that the deploy depends on

**Local (must exist):**

- `.env.ec2-credentials` — `SSH_KEY`, `USER`, `HOST1..4`
- `dist/mathsvr` — produced by BUILD (or pre-existing for `--skip-build`)
- `migrations/`, `keys/` — synced verbatim
- `deploy/` — synced verbatim **except** `deploy/scripts/`, which is the local tooling that drives this pipeline. The lifecycle hooks (`deploy/pre-deploy/`, `deploy/post-deploy/`) live inside it and are delivered along with it

**Remote (must exist on the host):**

- `/apps/math` — the destination root
- `/apps/math/.env` — runtime config; **never** synced by deploy
- `/apps/math/keys/` — TLS material; refreshed by `012_sync-certs.sh`
- `/apps/math/data/` — session files; permissions fixed by `008_`
- `/apps/math/mathsvr.pid`, `/apps/math/mathsvr.log` — owned by `015_start-server.sh`

**Logs:**

- Local: `deploy.log` (one line per phase, timestamped)
- Remote: `/apps/math/mathsvr.log` — tail via `make watch-logs RHOST=t1`

---

## 9. One-time host migration (`deploy/` layout)

The lifecycle hooks used to sit at `/apps/math/pre-deploy/` and
`/apps/math/post-deploy/`. They now live under `/apps/math/deploy/`.

This matters because **PREPARE runs before DELIVER**: the very first deploy
after this change looks for the stop hook at the new path, which the host does
not have yet. The graceful stop would be skipped and `prepare.sh` would fall
back to `pkill -9` — killing the server without serializing sessions.

Run this **once per host**, before the first deploy on the new layout. It both
installs the hooks at the new path and clears the stale copies (rsync never
deletes, so they would otherwise linger forever):

```bash
make login RHOST=t1     # or: ssh -i "$SSH_KEY" "$USER@$HOST"

mkdir -p /apps/math/deploy
mv /apps/math/pre-deploy /apps/math/post-deploy /apps/math/deploy/
ls /apps/math/deploy/    # expect: pre-deploy  post-deploy
```

If you skip it, PREPARE prints a `WARNING` naming this exact command — the
deploy still completes, but that one release stops ungracefully and loses
active sessions.

---

## 10. Quick reference

| Task | Command |
|---|---|
| Full deploy (arm64) | `make deploy RHOST=t1` |
| Deploy without rebuilding | `make deploy-quick RHOST=t1` |
| Deploy AMD64 | `make deploy-amd RHOST=t1` |
| Rollback last deploy | `make deploy-rollback RHOST=t1` |
| Tail prod logs | `make watch-logs RHOST=t1` |
| SSH into host | `make login RHOST=t1` |
