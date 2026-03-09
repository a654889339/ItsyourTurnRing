---
name: deploy-skill
description: Deploy ItsyourTurnRing project to Tencent Cloud server. Use when the user asks to deploy, publish, update the server, or push changes to production. Handles git push, SSH into server, git pull, and Docker rebuild.
---

# Deploy ItsyourTurnRing to Tencent Cloud

## Connection Info

| Key | Value |
|-----|-------|
| Server | `106.54.50.88:22` |
| User | `ubuntu` |
| SSH Key | `F:/ItsyourTurnMy/backend/deploy/test.pem` |
| Project Path (server) | `/root/ItsyourTurnRing` |
| GitHub Repo | `https://github.com/a654889339/ItsyourTurnRing.git` |
| Frontend Port | `5101` |
| Backend Port | `5102` |

## SSH Command Template

All Shell calls must use `required_permissions: ["all"]`.

```
ssh -o HostKeyAlgorithms=+ssh-rsa -o PubkeyAcceptedKeyTypes=+ssh-rsa -o StrictHostKeyChecking=no -o ConnectTimeout=30 -o ServerAliveInterval=10 -i F:/ItsyourTurnMy/backend/deploy/test.pem ubuntu@106.54.50.88 "sudo bash -c '<COMMAND>'"
```

## Deployment Workflow

### Step 1: Git Commit & Push (本地)

```powershell
cd F:\ItsyourTurnRing
git add -A; git status; git commit -m "<message>"; git push origin main
```

Use `;` to join commands (PowerShell `&&` unreliable). Use simple `-m "message"` (no heredoc).

### Step 2: SSH – Git Pull (服务器)

SSH into the server, pull latest code:

```bash
sudo bash -c 'cd /root/ItsyourTurnRing && git pull origin main'
```

**GitHub 网络不稳定处理**：服务器在中国大陆，访问 GitHub 可能超时或 TLS 断开。处理策略：
1. 设置 `block_until_ms: 120000`（2分钟等待）
2. 如果失败（GnuTLS error / timeout），等待 15 秒后重试
3. 最多重试 5 次
4. 如果连续失败，尝试先配置 git proxy 或换时间段再试

### Step 3: SSH – Docker Rebuild (服务器)

```bash
sudo bash -c 'cd /root/ItsyourTurnRing && docker compose down && docker compose up -d --build'
```

Set `block_until_ms: 0` to background（Docker build 需 2-5 分钟）。

### Step 4: Monitor Build

Poll terminal output file every 30-60s. Look for:
- `Container ring-backend ... Started` + `Container ring-frontend ... Started` → success
- Build error → read logs, fix, re-deploy

### Step 5: Verify

```bash
sudo docker ps --format 'table {{.Names}}\t{{.Status}}\t{{.Ports}}'
```

Check `ring-backend` and `ring-frontend` are `Up`.

Optional health check:

```bash
curl -s -o /dev/null -w '%{http_code}' http://localhost:5102/health
```

## Troubleshooting

| Problem | Solution |
|---------|----------|
| SSH timeout | Retry after 10-15s. Check security group allows port 22. |
| git pull GnuTLS error | GitHub 网络不稳定，等 15s 重试，最多 5 次 |
| git pull timeout | 同上，增大等待时间 |
| Docker build OOM | `docker system prune -f` then retry |
| Port conflict | Ring uses 5101/5102, ItsyourTurnMy uses 5001/5002/9090 |
| Backend crash | `sudo docker logs ring-backend --tail 50` |

## Important Notes

1. **服务器本地工程部署** – 服务器上已 clone 了 `/root/ItsyourTurnRing`，通过 `git pull` 更新后用本地代码构建 Docker。
2. **Always use `sudo`** – 项目在 `/root/` 下，需要 root 权限。
3. **Don't conflict with ItsyourTurnMy** – 那个项目用 5001/5002/9090 端口，容器名 `finance-backend`/`finance-frontend`。
4. **Docker volumes 持久化数据** – `ring_data` 和 `ring_uploads` 在 `docker compose down` 时不会被删除。
