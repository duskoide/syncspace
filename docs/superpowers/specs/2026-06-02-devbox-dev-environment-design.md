# Devbox-based Local Development Environment — Design

**Date:** 2026-06-02
**Status:** Approved
**Scope:** Local development tooling only. Production (Docker) is unchanged.

## Goal

Replace `dev.sh` and the implicit "install Go, Node, and sqlite yourself" setup with a reproducible [devbox](https://www.jetpack.io/devbox) environment at the repo root. The Docker production stack stays exactly as it is.

## What changes

A single new file at the repo root: `devbox.json`. It declares:

- The packages the dev environment provides (Go 1.23, Node 20, sqlite)
- The environment variables the dev services need
- A shell init hook that prepares `./dev-data/`
- Two long-running services (backend and frontend)
- A handful of scripts that wrap common workflows

`dev.sh` is deleted. The README's "Manual Development" section is replaced with devbox instructions. `AGENTS.md` is updated to reference devbox instead of `dev.sh`.

No source code in `backend/` or `frontend/src/` is touched. No Dockerfile, no `docker-compose.yml`, no `nginx.conf` is touched.

## `devbox.json`

```jsonc
{
  "$schema": "https://raw.githubusercontent.com/jetpack-io/devbox/main/.schema/devbox.schema.json",
  "packages": [
    "go@1.23",
    "nodejs@20",
    "sqlite"
  ],
  "env": {
    "SYNCSPACE_ADDR": ":8081",
    "SYNCSPACE_DB_PATH": "../dev-data/syncspace.db",
    "SYNCSPACE_UPLOAD_DIR": "../dev-data/uploads",
    "VITE_API_URL": "http://localhost:8081"
  },
  "shell": {
    "init_hook": "mkdir -p dev-data/uploads"
  },
  "services": {
    "backend": {
      "process": "mkdir -p ../dev-data && go run ./cmd/syncspace",
      "directory": "backend"
    },
    "frontend": {
      "process": "npm run dev",
      "directory": "frontend"
    }
  },
  "scripts": {
    "dev":    "devbox services up",
    "stop":   "devbox services stop",
    "test":   "cd backend && go test ./... && cd ../frontend && npx tsc -b",
    "build":  "cd frontend && npm run build",
    "logs":   "devbox services logs -f"
  }
}
```

### Choices explained

- **Packages** — exact versions match the Dockerfiles (`go:1.23-alpine`, `node:20-alpine`). `sqlite` adds the `sqlite3` CLI for poking at the dev DB, mirroring what the README's Docker troubleshooting section suggests.
- **Env block** — the same env vars that `dev.sh` exported, declared once. `VITE_API_URL` is set in the env block so the Vite dev server picks it up at process start (Vite reads it at boot, not per-request). Path env vars (`SYNCSPACE_DB_PATH`, `SYNCSPACE_UPLOAD_DIR`) use `../dev-data/...` because the backend service runs from the `backend/` working dir; this matches what `dev.sh` does after its `cd backend`.
- **`shell.init_hook`** — idempotent `mkdir -p dev-data/uploads`. Runs when entering `devbox shell`. The `dev-data/uploads` subdirectory is created lazily by the file service on first upload, but creating it eagerly here keeps the directory tree tidy.
- **Services** — `process` is the command, `directory` is its working dir. The backend command prefixes `mkdir -p ../dev-data && go run ./cmd/syncspace` because `shell.init_hook` does NOT run for `devbox services up` — the `mkdir` is idempotent and ensures `dev-data/` exists before SQLite tries to open the DB. The frontend uses Vite's dev script and inherits `VITE_API_URL` from the env block. The backend and frontend are siblings, no `depends_on` (CORS allows localhost either way and Vite starts fast enough that the backend's first request works).
- **Scripts** — `dev` brings both services up with logs in the foreground; Ctrl+C stops both. `stop` works from a separate shell. `test` runs the backend Go suite (the only test layer) then the frontend typecheck. `build` mirrors what the production Dockerfile does. `logs` tails logs from a separate shell.

## User workflow

| Step | Command | Notes |
|------|---------|-------|
| One-time | Install devbox: `curl -fsSL https://get.jetpack.io/devbox \| bash` | Per the official devbox install page. |
| First run | `devbox run dev` | Triggers `devbox install` automatically on first invocation. |
| Daily | `devbox run dev` | Starts backend on `:8081` and frontend on `:5173` with logs interleaved. Ctrl+C stops both. |
| Stop from another shell | `devbox run stop` | Useful if the foreground process is in a different terminal. |
| Tail logs | `devbox run logs` | Streams logs from a separate shell. |
| Run tests | `devbox run test` | Backend Go tests + frontend typecheck. |
| Build frontend | `devbox run build` | Output in `frontend/dist/`. |
| Drop into a shell | `devbox shell` | Go, Node, and sqlite on PATH. Env vars from `devbox.json` exported. |

## Port and data layout

- Backend: `:8081` (unchanged from `dev.sh`).
- Frontend: `:5173` (Vite default, unchanged from `dev.sh`).
- DB: `./dev-data/syncspace.db` (unchanged from `dev.sh`).
- Uploads: `./dev-data/uploads/` (unchanged from `dev.sh`).
- Production Docker stack on `:3000` is unaffected.

## Documentation changes

### README.md

Replace the existing "Manual Development" subsection (lines 93-120 in the current file) with a "Local Development (devbox)" subsection that:

- Links to the devbox install page.
- Lists the `devbox run dev` command as the primary workflow.
- Provides a table of service URLs and the database location.
- Lists the other `devbox run` scripts (stop, test, build, logs, shell).
- Notes that production Docker on `:3000` does not conflict.

Delete the `chmod +x dev.sh` line and the separate-terminals instructions. The new devbox section is the only local-dev story in the README.

### AGENTS.md

Two edits:

- In the "Layout" section, replace the bullet about `dev.sh` with a bullet about `devbox.json` that notes: "Nix-based dev environment. Same ports and DB as the old dev.sh, no Docker required locally."
- In the "Commands" section, replace the `Dev (both services, separate DB): ./dev.sh from repo root.` line with `Dev (both services, separate DB): devbox run dev from repo root.`

## What stays the same

- `docker-compose.yml` — production stack unchanged.
- `Dockerfile.backend`, `Dockerfile.frontend` — unchanged.
- `frontend/nginx.conf` — unchanged.
- `.env`, `.env.example`, `.env.prod` — still relevant for the Docker tunnel profile.
- `.gitignore` — already ignores `dev-data/`, `frontend/node_modules/`, etc.
- All source code in `backend/` and `frontend/src/`.
- `AGENTS.md` quirks sections (CORS, JWT defaults, SQLite single-connection, etc.).

## Out of scope

- **No Go hot-reload.** README explicitly says "Go API with hot-reload disabled" today. Adding `air` is a separate, optional change.
- **No CGO removal or build optimizations.** Dev uses `go run` with CGO defaults. Production's `CGO_ENABLED=0` in `Dockerfile.backend` is unchanged.
- **No frontend test runner.** No frontend tests exist; adding Vitest is separate.
- **No CI / GitHub Actions for devbox.** No CI exists today; adding it is separate.
- **No devbox lockfile.** Devbox resolves Nix packages on first install. `devbox.lock` can be added later if reproducible resolution becomes an issue.
- **No migration tooling for existing devs with Docker.** The README's "Quickstart" stays Docker-first. `devbox run dev` is the new alternative; if a dev has both, they pick one.
