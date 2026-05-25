# aaPanel Docker Deployment Guide

## Prerequisites
- aaPanel installed on your server
- Docker Manager plugin installed in aaPanel
- Domain pointing to your server (e.g., syncspaceedu.duskoide.org)

## Step 1: Upload Project Files

1. In aaPanel, go to **Files** → Upload your entire project to:
   ```
   /www/wwwroot/syncspace/
   ```

2. Or use SSH/Git:
   ```bash
   cd /www/wwwroot
   git clone <your-repo> syncspace
   cd syncspace
   ```

## Step 2: Build Docker Images

In aaPanel Terminal or SSH:

```bash
cd /www/wwwroot/syncspace

# Build backend image
docker build -f Dockerfile.backend -t syncspace-backend .

# Build frontend image  
docker build -f Dockerfile.frontend -t syncspace-frontend .
```

## Step 3: Create Data Directories

```bash
mkdir -p /www/wwwroot/syncspace/data
mkdir -p /www/wwwroot/syncspace/uploads
chmod 755 /www/wwwroot/syncspace/data
chmod 755 /www/wwwroot/syncspace/uploads
```

## Step 4: Run Containers

### Option A: Using Docker Compose (if aaPanel supports it)

```bash
cd /www/wwwroot/syncspace
docker-compose up -d
```

### Option B: Manual Docker Run (Recommended for aaPanel)

**Backend container:**
```bash
docker run -d \
  --name syncspace-backend \
  --restart unless-stopped \
  -p 127.0.0.1:8080:8080 \
  -v /www/wwwroot/syncspace/data:/data \
  -v /www/wwwroot/syncspace/uploads:/uploads \
  -e SYNCSPACE_ADDR=:8080 \
  -e SYNCSPACE_DB_PATH=/data/syncspace.db \
  syncspace-backend
```

**Frontend container:**
```bash
docker run -d \
  --name syncspace-frontend \
  --restart unless-stopped \
  -p 127.0.0.1:3000:80 \
  syncspace-frontend
```

## Step 5: Configure aaPanel Reverse Proxy

### Backend API (api.yourdomain.com or yourdomain.com/api)

1. In aaPanel, go to **Website** → Add Site
2. Domain: `api.syncspaceedu.duskoide.org` (or your subdomain)
3. **Reverse Proxy** tab:
   - Target URL: `http://127.0.0.1:8080`
   - Send Domain: `$host`
   - Enable **WebSocket Support** (for /ws endpoint)

### Frontend (yourdomain.com)

1. Add another site or use main domain
2. **Reverse Proxy** tab:
   - Target URL: `http://127.0.0.1:3000`
   - Send Domain: `$host`

## Step 6: SSL Certificate

1. In aaPanel, go to your site's **SSL** tab
2. Click **Let's Encrypt**
3. Select your domain
4. Enable **Force HTTPS**

## Step 7: WebSocket Configuration (Important!)

For real-time discussions to work, you need to add WebSocket support in Nginx:

1. Go to your API site's **Config** tab
2. Find the location block and add:

```nginx
location /ws {
    proxy_pass http://127.0.0.1:8080;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}
```

## Step 8: Update Frontend API URL

Before building frontend, update the API URL in frontend code:

**File:** `frontend/src/services/api.ts`

```typescript
const API_BASE = "https://api.syncspaceedu.duskoide.org"; // Your API domain
```

Then rebuild:
```bash
cd /www/wwwroot/syncspace/frontend
docker build -f ../Dockerfile.frontend -t syncspace-frontend .
docker stop syncspace-frontend
docker rm syncspace-frontend
docker run -d --name syncspace-frontend -p 127.0.0.1:3000:80 syncspace-frontend
```

## Step 9: Verify Deployment

```bash
# Check containers are running
docker ps

# Check logs
docker logs syncspace-backend
docker logs syncspace-frontend

# Test API
curl https://api.syncspaceedu.duskoide.org/health

# Test frontend
curl https://syncspaceedu.duskoide.org
```

## Useful Commands

```bash
# View logs
docker logs -f syncspace-backend
docker logs -f syncspace-frontend

# Restart containers
docker restart syncspace-backend
docker restart syncspace-frontend

# Update after code changes
cd /www/wwwroot/syncspace
docker build -f Dockerfile.backend -t syncspace-backend .
docker stop syncspace-backend
docker rm syncspace-backend
# Then re-run the docker run command from Step 4

# Backup database
cp /www/wwwroot/syncspace/data/syncspace.db /www/backup/syncspace-$(date +%Y%m%d).db

# Access container shell
docker exec -it syncspace-backend sh
docker exec -it syncspace-frontend sh
```

## Troubleshooting

### Container won't start
Check logs: `docker logs syncspace-backend`

### Database permission errors
```bash
chmod 777 /www/wwwroot/syncspace/data
```

### API not accessible
- Check if backend is listening: `netstat -tlnp | grep 8080`
- Check aaPanel firewall: Allow port 8080 (if not using reverse proxy)
- Check aaPanel Nginx config for reverse proxy

### WebSocket not working
- Make sure WebSocket upgrade headers are configured in Nginx
- Check browser console for connection errors
- Verify token is being passed in WebSocket URL

### File uploads failing
```bash
chmod 777 /www/wwwroot/syncspace/uploads
```

## File Structure on Server

```
/www/wwwroot/syncspace/
├── backend/
│   ├── Dockerfile.backend
│   └── syncspace (binary - will be built)
├── frontend/
│   └── Dockerfile.frontend
├── data/
│   └── syncspace.db (created on first run)
├── uploads/
│   └── (user uploaded files)
├── docker-compose.yml
└── deploy/
    └── aapanel-docker-guide.md (this file)
```

## Quick Start Checklist

- [ ] Upload project to `/www/wwwroot/syncspace/`
- [ ] Install Docker Manager in aaPanel
- [ ] Build Docker images
- [ ] Create data and uploads directories
- [ ] Run containers
- [ ] Configure reverse proxy in aaPanel
- [ ] Set up SSL certificates
- [ ] Configure WebSocket in Nginx
- [ ] Update frontend API URL and rebuild
- [ ] Test the deployed application
