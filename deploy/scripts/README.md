# Local operator tooling

Shell scripts that run on a **developer machine**, against a `math-svr`
checkout. Nothing here is installed on — or shipped to — the EC2 host:
`lib/deliver.sh` excludes this directory from the rsync.

Everything else under `deploy/` *is* delivered, to `/apps/math/deploy/`:

| Sibling | Runs where | What it is |
|---|---|---|
| [`../pre-deploy/`](../pre-deploy/) | on the host, during PREPARE | Stop hooks this pipeline invokes over SSH |
| [`../post-deploy/`](../post-deploy/) | on the host, during ACTIVATE | Start / health hooks this pipeline invokes over SSH |
| [`../app-service/`](../app-service/) | on the host | systemd unit + the `mai` operator wrapper |
| [`../logrotate/`](../logrotate/) | on the host | logrotate config + hourly timer |
| [`../nginx/`](../nginx/) | on the host | nginx config |

So `deploy/` splits along one line: **`scripts/` is the only thing that stays
local** — it is the machinery that pushes all of the above to the host.

Every script is invoked **from the repository root** (that is what the `make`
targets do). The three that need it re-derive the project root from their own
location, so `./deploy/scripts/deploy.sh t1` works from anywhere; the rest read
`.env.ec2-credentials` and `migrations/` relative to the current directory and
therefore expect the repo root as the working directory.

## Deployment pipeline

| Script | `make` target | Purpose |
|---|---|---|
| [`deploy.sh`](./deploy.sh) | `make deploy RHOST=t1` | Full pipeline: validate → build → prepare → deliver → activate |
| | `make deploy-quick RHOST=t1` | Same, `--skip-build` (reuse `dist/mathsvr`) |
| | `make deploy-rollback RHOST=t1` | `--rollback` — restore the previous binary on the host |
| | `make deploy-amd RHOST=t1` | `BUILD_ARCH=amd64` — force an AMD64 build |
| [`lib/`](./lib/) | — | The six phase libraries `deploy.sh` sources: `common`, `validate`, `build`, `prepare`, `deliver`, `activate`, `rollback` |
| [`init_remote.sh`](./init_remote.sh) | — | One-off bootstrap of a brand-new host (creates the server dir, uploads keys + `.env`) |

Full walkthrough of every phase: [`docs/DeploymentProcessExplain.md`](../../docs/DeploymentProcessExplain.md).

## Day-to-day operations

| Script | `make` target | Purpose |
|---|---|---|
| [`login.sh`](./login.sh) | `make login RHOST=t1` | SSH into the host |
| [`watch-logs.sh`](./watch-logs.sh) | `make watch-logs RHOST=t1` | Tail `/apps/math/mathsvr.log` |
| [`connect-mysql.sh`](./connect-mysql.sh) | `make connect-mysql` | Open a `mysql` shell on the host using its `/apps/math/.env` |
| [`clear-data.sh`](./clear-data.sh) | `make clear-data-local` / `make clear-data-ec2 RHOST=…` | Run `sql/clear_data.sql` — wipes user data, keeps reference data. **Destructive** |
| [`create_migration.sh`](./create_migration.sh) | — | Scaffold a migration pair. ⚠️ Writes into `migrations/up/` + `migrations/down/`, which does **not** match this repo's flat, forward-only `migrations/NNN_*.sql` convention |
| [`verify-graceful-shutdown.sh`](./verify-graceful-shutdown.sh) | — | Prove SIGTERM triggers session serialization (and that SIGHUP does not) |

## Credentials

Every remote-facing script reads `.env.ec2-credentials` from the repository
root — `SSH_KEY`, `USER`, and `HOST` / `HOST1`..`HOST4`. The file is
gitignored and is never synced to the host. `RHOST=t1..t4` selects
`HOST1`..`HOST4`; anything else falls back to `HOST`.
