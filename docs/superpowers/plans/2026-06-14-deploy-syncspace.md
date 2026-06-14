# Deploy SyncSpace on Local Machine

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Install Docker on openSUSE Tumbleweed, deploy SyncSpace via docker-compose, and expose it publicly through the existing Cloudflare Tunnel.

**Architecture:** Install Docker + docker-compose plugin, rebuild the Go backend binary for Linux amd64, start all services with `docker-compose --profile tunnel up`, and verify the tunnel is serving traffic.

**Tech Stack:** Docker, docker-compose, Alpine Linux containers, Cloudflare Tunnel

---

### Task 1: Install Docker on openSUSE Tumbleweed

- [ ] **Step 1: Add Docker repository and install**

```bash
sudo zypper install -y docker docker-compose compose-plugin
```

- [ ] **Step 2: Start and enable Docker daemon**

```bash
sudo systemctl enable --now docker
```

- [ ] **Step 3: Add current user to docker group (免 sudo)**

```bash
sudo usermod -aG docker $USER
newgrp docker
```

- [ ] **Step 4: Verify Docker works**

```bash
docker run --rm hello-world
```

Expected: "Hello from Docker!" message appears.

---

### Task 2: Rebuild backend binary for Linux amd64

The Dockerfile.backend volume-mounts `./backend/bin/syncspace` into the container. The existing binary was built for the host, but we need to ensure it's linux/amd64.

- [ ] **Step 1: Check existing binary architecture**

```bash
file backend/bin/syncspace
```

Expected: Should show `ELF 64-bit LSB executable, x86-64` (Linux amd64). If already correct, skip rebuild.

- [ ] **Step 2: If not linux/amd64, rebuild**

```bash
cd backend && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/syncspace ./cmd/syncspace
```

---

### Task 3: Configure environment

- [ ] **Step 1: Verify .env file has required values**

The `.env` file should contain:
```
TUNNEL_TOKEN=<token>
JWT_SECRET=anu
```

Confirm `TUNNEL_TOKEN` is set (it is — value present in `.env`).

- [ ] **Step 2: Ensure data and uploads directories exist**

```bash
mkdir -p data uploads
```

---

### Task 4: Start all services with Docker Compose

- [ ] **Step 1: Build and start with tunnel profile**

```bash
docker-compose --profile tunnel up --build -d
```

Expected: Three containers start — `backend`, `frontend`, `cloudflared`.

- [ ] **Step 2: Verify all containers are running**

```bash
docker-compose ps
```

Expected: All three services show "Up" status.

- [ ] **Step 3: Check container logs for errors**

```bash
docker-compose logs --tail=20
```

Expected: No crash loops or connection errors.

---

### Task 5: Verify deployment

- [ ] **Step 1: Test backend health**

```bash
curl -s http://localhost:3000/api/auth/me
```

Expected: JSON response (likely `{"error":"unauthorized"}` — confirms backend is running).

- [ ] **Step 2: Test frontend serves HTML**

```bash
curl -s http://localhost:3000/ | head -5
```

Expected: HTML response with `<div id="root">`.

- [ ] **Step 3: Verify Cloudflare Tunnel is connected**

```bash
docker-compose logs cloudflared | tail -5
```

Expected: Log shows "Connection registered" or similar success message. The app should be accessible at the public hostname configured in Cloudflare dashboard (`syncspaceedu.duskoide.org`).

- [ ] **Step 4: Test public access**

Open `https://syncspaceedu.duskoide.org` in a browser or curl it:

```bash
curl -sI https://syncspaceedu.duskoide.org
```

Expected: HTTP 200 response.
