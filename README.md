# SyncSpace - Collaborative Note-Taking & Template Sharing Platform

SyncSpace is a self-hosted collaborative note-taking platform with template sharing, built for the II2210 Teknologi Platform course assignment.

## Theme
Education & Learning Tools - Note-Taking & Template Sharing Platform

## Stack
- **Backend:** Go 1.23 + REST API
- **Frontend:** React 18 + Vite + TypeScript + React Router
- **Database:** SQLite (WAL mode)
- **Authentication:** JWT with bcrypt password hashing
- **Containerization:** Docker + Docker Compose

## Quickstart

### With Docker Compose (Recommended)
```bash
git clone <repo-url> && cd syncspace
docker-compose up --build -d
```

The app is available at http://localhost:3000

## Docker Usage

### Basic Commands
```bash
# Start all services
docker-compose up -d

# Start with rebuild (after code changes)
docker-compose up --build -d

# View logs
docker-compose logs -f

# View specific service logs
docker-compose logs -f backend
docker-compose logs -f frontend

# Stop all services
docker-compose down

# Stop and remove volumes (WARNING: deletes database)
docker-compose down -v

# Restart a specific service
docker-compose restart backend
```

### Data Persistence
- **Database:** Stored in `./data/syncspace.db` (SQLite)
- **Uploads:** Files stored in `./data/uploads/`
- **Logs:** Check with `docker-compose logs`

### Environment Variables
Create a `.env` file for configuration:
```env
# JWT Secret (generate a secure random string)
JWT_SECRET=your-secret-key-here

# Cloudflare Tunnel Token (optional)
TUNNEL_TOKEN=your-cloudflare-tunnel-token
```

### Troubleshooting
```bash
# Check container status
docker-compose ps

# Execute commands in container
docker-compose exec backend sh
docker-compose exec frontend sh

# Check database inside container
docker-compose exec backend sh -c "sqlite3 /data/syncspace.db '.tables'"

# Reset everything (delete data)
docker-compose down -v
rm -rf data/
docker-compose up --build -d
```

### With Cloudflare Tunnel (Public Access)
1. Copy `.env.example` to `.env`: `cp .env.example .env`
2. Create a Cloudflare Tunnel in [Zero Trust Dashboard](https://one.dash.cloudflare.com/) → Networks → Tunnels
3. Choose "Docker" connector type, copy the token
4. Set the tunnel's public hostname origin to: `http://frontend:80`
5. Paste the token in `.env` as `TUNNEL_TOKEN=your_token_here`
6. Start with tunnel profile: `docker-compose --profile tunnel up --build -d`

### Manual Development
```bash
# Backend
cd backend && go run ./cmd/syncspace

# Frontend
cd frontend && npm install && npm run dev
```

## Default Credentials

The following accounts are automatically created on first startup:

| Role | Email | Password |
|------|-------|----------|
| **Superadmin** | `admin@syncspace.edu` | `admin123` |
| **Creator** | `creator@syncspace.edu` | `creator123` |
| **User** | `user@syncspace.edu` | `user123` |

Use these to test different roles without registration.

## User Roles

| Role | Capabilities |
|------|-------------|
| **Superadmin** | Manage users (approve/suspend), moderate templates, full access |
| **Creator** | Create notes/workspaces, create and share templates |
| **User** | Create notes/workspaces, browse and use templates |

New registrations start with **pending** status and must be approved by a superadmin.

## Features

### Core
- **Three User Roles** - Superadmin, Creator, User with distinct capabilities
- **Registration & Approval** - Users register, superadmin approves before access
- **Workspaces** - Personal workspace for organizing notes
- **Rich Text Notes** - TipTap editor with headings, lists, bold/italic
- **Inline Images** - Upload and embed images in notes
- **Template Sharing** - Creators share workspace/note templates (public or link-only)
- **Template Cloning** - Users clone templates into their own workspaces
- **Template Moderation** - Superadmin can hide/unhide templates
- **Wikipedia Sidebar** - Search Wikipedia and insert summaries into notes

### Producer-Consumer Flow
- **Creator (Producer):** Creates templates from their workspaces/notes, shares them publicly or via link
- **User (Consumer):** Discovers and clones templates to create their own copies
- Templates are snapshots - clones are independent and won't change when the original is updated

## API Endpoints

### Authentication
- `POST /api/auth/register` - Register (status: pending, requires superadmin approval)
- `POST /api/auth/login` - Login (blocked if pending or suspended)
- `GET /api/auth/me` - Get current user

### Admin (Superadmin only)
- `GET /api/admin/users` - List users (filter by role/status)
- `PUT /api/admin/users/{id}/activate` - Approve/activate a pending user
- `PUT /api/admin/users/{id}/suspend` - Suspend a user
- `DELETE /api/admin/users/{id}` - Delete a user
- `GET /api/admin/templates` - List all templates
- `PATCH /api/admin/templates/{id}` - Hide/unhide template

### Workspaces
- `GET /api/workspaces` - List my workspaces
- `POST /api/workspaces` - Create workspace
- `GET /api/workspaces/{id}` - Get workspace
- `PUT /api/workspaces/{id}` - Update workspace
- `DELETE /api/workspaces/{id}` - Delete workspace

### Notes
- `GET /api/workspaces/{id}/notes` - List notes in workspace
- `POST /api/workspaces/{id}/notes` - Create note
- `GET /api/notes/{id}` - Get note
- `PUT /api/notes/{id}` - Update note
- `DELETE /api/notes/{id}` - Delete note

### Templates
- `GET /api/templates` - Browse public templates (searchable)
- `GET /api/templates/my` - List my templates (creator)
- `POST /api/templates` - Create template from workspace/note
- `PUT /api/templates/{id}` - Update template metadata
- `POST /api/templates/{id}/update-content` - Update template content from source
- `DELETE /api/templates/{id}` - Delete template
- `POST /api/templates/{id}/clone` - Clone template to workspace

### File Upload
- `POST /api/upload` - Upload image (multipart/form-data)
- `GET /api/files/{id}` - Download file
- `DELETE /api/files/{id}` - Delete file

### Wikipedia
- `GET /api/wiki/summary?topic={topic}` - Get Wikipedia summary

## Docker Architecture

### Container Layout
```
┌──────────────────┐     ┌──────────────────┐     ┌──────────────────┐
│   cloudflared    │────▶│    frontend      │────▶│    backend       │
│  (optional)      │     │    (nginx)       │     │    (Go API)      │
│  :443 → :80     │     │    :80 → :3000   │     │    :8080         │
└──────────────────┘     └──────────────────┘     └──────────────────┘
                                │                         │
                                │    Docker Network        │
                                │    (syncspace)           │
                                │                         │
                         ┌──────┴─────────────────────────┘
                         │
                    ┌────▼────┐
                    │  data/  │  (SQLite DB + uploads)
                    └─────────┘
```

### How Docker Helps Deployment
1. **Consistent Environment** - Same runtime across dev, staging, and production
2. **Isolated Services** - Backend, frontend, and tunnel run in separate containers
3. **One Command Deploy** - `docker-compose up --build -d` starts everything
4. **Nginx Reverse Proxy** - Frontend container proxies `/api` to backend internally
5. **Persistent Data** - SQLite DB and uploaded files stored in mounted volumes
6. **Easy Updates** - Rebuild and restart without manual configuration

### Container Details
| Container | Image | Purpose | Port |
|-----------|-------|---------|------|
| backend | Custom (Go 1.23-alpine) | REST API + auth + business logic | 8080 (internal) |
| frontend | Custom (Node + nginx) | React SPA served by nginx | 3000:80 |
| cloudflared | cloudflare/cloudflared | Optional public access via tunnel | - |

## Database Schema

- `users` - Authentication & roles (superadmin, creator, user) with status (pending, active, suspended)
- `workspaces` - User workspaces for organizing notes
- `notes` - Rich text notes (HTML content from TipTap editor)
- `templates` - Shareable workspace/note snapshots (public or link visibility)
- `note_images` - Inline image metadata for notes

## References
- Go net/http - https://pkg.go.dev/net/http
- JWT Authentication - https://datatracker.ietf.org/doc/html/rfc7519
- Layered Architecture - "Patterns of Enterprise Application Architecture" by Martin Fowler
- SQLite WAL Mode - https://sqlite.org/wal.html
- TipTap Editor - https://tiptap.dev
- Cloudflare Tunnel - https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/

## License
Academic project for II2210 Teknologi Platform - Institut Teknologi Bandung.