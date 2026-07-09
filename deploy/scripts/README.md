# Operator scripts

Convenience shell scripts for operating the `mathsvr` service on the EC2 host.

- [`alias_mai`](./alias_mai) — a thin wrapper around `systemctl` for the
  `mathsvr.service`, exposed as the `mai` command.

## `alias_mai`

Wraps the common lifecycle commands so an operator can type `mai start` instead
of `sudo systemctl start mathsvr.service`.

| Command | Action |
|---|---|
| `mai start` | Start the server (via systemd) |
| `mai stop` | Stop the server gracefully (SIGTERM) |
| `mai restart` | Restart the server |
| `mai status` | Show service status |
| `mai watch` | Tail `/apps/math/mathsvr.log` in real time |

It targets `mathsvr.service` — the unit installed from
[`../app-service/`](../app-service/).

## Installation (run on EC2)

```bash
# 1. Make it executable
chmod +x /apps/math/deploy/scripts/alias_mai

# 2. Expose it as the `mai` command. Either symlink it onto PATH...
sudo ln -sf /apps/math/deploy/scripts/alias_mai /usr/local/bin/mai

# 3. ...or add a shell alias (append to ~/.bashrc, then re-login):
#    alias mai='/apps/math/deploy/scripts/alias_mai'

# 4. Verify
mai status
```

## Notes

- The script calls `sudo systemctl ...` internally, so the operator needs sudo
  rights for the `mathsvr.service` control verbs.
- `mai watch` tails the console log file directly (`/apps/math/mathsvr.log`),
  the same file rotated by [`../logrotate/`](../logrotate/).
- **Shebang caveat:** the first line is `#!/usr/local/bin bash`, which is not a
  valid interpreter path. Run it as `bash alias_mai <cmd>` or fix the shebang to
  `#!/usr/bin/env bash` before relying on direct execution / the symlink above.
