---
name: deploy-skill
description: Deploy ItsyourTurnRing project to Tencent Cloud server. Use when the user asks to deploy, publish, update the server, or push changes to production. Handles git push, file transfer via SCP, and Docker rebuild on the remote server.
---

# Deploy ItsyourTurnRing to Tencent Cloud

## Connection Info

| Key | Value |
|-----|-------|
| Server | `106.54.50.88:22` |
| User | `ubuntu` |
| SSH Key | `F:/ItsyourTurnMy/backend/deploy/test.pem` |
| Project Path (server) | `/root/ItsyourTurnRing` |
| Frontend Port | `5101` |
| Backend Port | `5102` |

SSH options (required for RSA compatibility):

```
-o HostKeyAlgorithms=+ssh-rsa -o PubkeyAcceptedKeyTypes=+ssh-rsa -o StrictHostKeyChecking=no -o ConnectTimeout=30 -o ServerAliveInterval=10
```

Shorthand used below: `$SSH_OPTS` = the options above, `$KEY` = the SSH key path, `$HOST` = `ubuntu@106.54.50.88`.

## Deployment Workflow

Execute these steps **in order**. Use `required_permissions: ["all"]` for every Shell call.

### Step 1: Git Commit & Push

```powershell
cd F:\ItsyourTurnRing
git add -A
git status
git commit -m "<meaningful commit message>"
git push origin main
```

Use `;` to join commands (PowerShell does not support `&&` reliably).
Use simple `-m "message"` for commit (PowerShell does not support heredoc).

### Step 2: Pack Project for SCP

**CRITICAL**: The server cannot reliably `git pull` from GitHub (GnuTLS TLS errors). Always use SCP.

```powershell
cd F:\ItsyourTurnRing
tar -czf ring-deploy.tar.gz --exclude=node_modules --exclude=.git --exclude=.idea --exclude="*.db" --exclude=data --exclude=logs --exclude=dist --exclude=.output --exclude=ring-deploy.tar.gz .
```

This creates `ring-deploy.tar.gz` (~small, excludes build artifacts and data).

### Step 3: SCP Upload to Server

```powershell
scp $SSH_OPTS -i $KEY ring-deploy.tar.gz $HOST:/tmp/ring-deploy.tar.gz
```

If SCP hangs or fails, retry up to 3 times with 10s sleep between attempts.

### Step 4: SSH – Extract & Rebuild

Run as a single SSH command with `sudo bash -c '...'`:

```bash
sudo bash -c '
  cd /root/ItsyourTurnRing &&
  docker compose down &&
  rm -rf backend frontend miniprogram-wechat miniprogram-alipay config.yaml docker-compose.yaml .env.example &&
  tar -xzf /tmp/ring-deploy.tar.gz -C /root/ItsyourTurnRing &&
  rm /tmp/ring-deploy.tar.gz &&
  docker compose up -d --build
'
```

Set `block_until_ms: 0` to background this command (Docker build takes 2-5 minutes).

### Step 5: Monitor Build

Poll the terminal output file every 30s. Look for:
- `Creating ring-backend ... done` / `Creating ring-frontend ... done` → success
- `ring-backend` and `ring-frontend` containers started → success
- Build errors → read logs, fix, re-deploy

### Step 6: Verify

```bash
sudo docker ps --format 'table {{.Names}}\t{{.Status}}\t{{.Ports}}'
```

Check that `ring-backend` and `ring-frontend` are `Up` and healthy.

Optional API check:

```bash
curl -s -o /dev/null -w '%{http_code}' http://localhost:5102/health
```

### Step 7: Cleanup Local

```powershell
Remove-Item F:\ItsyourTurnRing\ring-deploy.tar.gz -ErrorAction SilentlyContinue
```

## Troubleshooting

| Problem | Solution |
|---------|----------|
| SSH timeout | Retry after 10s. If persistent, check Tencent Cloud security group allows port 22. |
| SCP timeout | Kill the process (`taskkill /PID <pid> /F`), wait 10s, retry. |
| Docker build OOM | SSH in, run `docker system prune -f` then retry. |
| Port conflict | Ensure ItsyourTurnMy uses 5001/5002 and Ring uses 5101/5102. Run `sudo netstat -tlnp | grep 510` to verify. |
| Backend crash | Check logs: `sudo docker logs ring-backend --tail 50` |
| Frontend 502 | Backend may not be healthy yet. Wait for healthcheck or check backend logs. |

## Important Notes

1. **Always use SCP** – `git pull` from the server is unreliable due to China GFW/network issues.
2. **Always use `sudo`** – project lives in `/root/` which requires root access.
3. **Don't conflict with ItsyourTurnMy** – that project uses ports 5001, 5002, 9090 and containers `finance-backend`, `finance-frontend`.
4. **Docker volumes persist data** – `ring_data` and `ring_uploads` are not deleted by `docker compose down` (only `docker compose down -v` removes them).
