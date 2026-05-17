# syncspace

SyncSpace Edu is a self-hosted educational productivity platform.

## Stack
- Backend: Go + REST API
- Frontend: React + Vite + TypeScript
- DB: SQLite (WAL)
- External API: Wikipedia

## Quickstart
### Backend
```bash
cd backend
go run ./cmd/syncspace
```

### Frontend
```bash
cd frontend
npm install
cp .env.example .env
npm run dev
```

## How To Use
1. Open the frontend app in your browser.
2. Add tasks in the `Tasks` section, then toggle status or delete as needed.
3. Add notes in the `Notes` section and select one note as the enrichment target.
4. Enter a topic in `Wikipedia Enrichment`, preview summary, then click `Enrich Selected Note`.
5. Refresh task/note lists to verify persisted data from SQLite.

## Production Build
### Backend binary
```bash
cd backend
mkdir -p bin
go build -o bin/syncspace ./cmd/syncspace
```

### Frontend static build
```bash
cd frontend
npm install
npm run build
```

## Core Endpoints
- `GET /health`
- `GET/POST /api/tasks`
- `GET/PUT/DELETE /api/tasks/{id}`
- `GET/POST /api/notes`
- `GET/PUT/DELETE /api/notes/{id}`
- `GET /api/wiki/summary?topic=...`
- `POST /api/notes/{id}/enrich`
