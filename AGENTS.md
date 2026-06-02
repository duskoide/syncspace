# AGENTS.md

Compact notes for OpenCode sessions working in `syncspace`.

## Layout

- `backend/` — Go 1.23 REST API. Module path is `syncspace/backend` (not `backend`). Entry: `backend/cmd/syncspace/main.go`.
- `frontend/` — React 18 + Vite + TypeScript SPA. Uses TipTap editor and React Router v7.
- `docker-compose.yml` — backend (`:8080` internal), frontend (`:3000` host → `:80` internal), optional `cloudflared` (profile `tunnel`).
- `devbox.json` + `process-compose.yaml` — devbox-managed dev environment. Backend `:8081` + frontend `:5173`, separate DB at `./dev-data/syncspace.db`. No Docker required locally. `process-compose.yaml` declares the two long-running services that `devbox services up` launches.

## Backend quirks

- SQLite driver is `modernc.org/sqlite` (pure Go, no CGO). `internal/store/store.go` forces `SetMaxOpenConns(1)` and enables WAL — tests must use `t.TempDir()` for the DB path, not `:memory:`.
- `migrate()` in `internal/store/store.go` runs a `DROP TABLE IF EXISTS` for legacy tables at startup. Editing that block changes destructive behaviour on every boot.
- Routes are registered with Go 1.22+ method-prefixed patterns (`mux.HandleFunc("GET /api/...", ...)`). The dispatcher functions (`handleWorkspace`, `handleNote`, `handleTemplate`, `handleAdminUser`) switch on `r.Method` — do not add a new method without updating the dispatcher and the `mux.Handle` line.
- Auth: `internal/auth/jwt.go` secret is set once via `auth.SetJWTSecret(cfg.JWTSecret)` in `main.go`. JWT is read by `internal/api/middleware.go` (`AuthMiddleware`, `RequireRole`).
- WebSocket hub lives at `/ws` (`internal/websocket/hub.go`). Not documented in README.
- Wikipedia fetch is in `internal/service/service.go::WikiSummary` — sets a custom `User-Agent` (required by Wikipedia's API policy). Reuse the same client/headers for any future external API call.
- CORS in `main.go::withCORS` only allows `localhost` origins and the no-`Origin` case (nginx proxied). Adding a deployed frontend requires editing this function.
- Config env vars: `SYNCSPACE_ADDR`, `SYNCSPACE_DB_PATH`, `SYNCSPACE_UPLOAD_DIR`, `SYNCSPACE_JWT_SECRET`. The JWT secret has an insecure default — set it in any non-dev deploy.
- On first start, `store.Open` seeds three accounts: `admin@syncspace.edu` / `admin123` (superadmin), `creator@syncspace.edu` / `creator123` (creator), `user@syncspace.edu` / `user123` (user). New registrations start `pending` and must be activated by a superadmin.
- Templates are workspace-only (note-type templates were removed in commit `922141e`). `templates.type` is always `'workspace'`.

## Frontend quirks

- API base URL: `import.meta.env.VITE_API_URL ?? "http://localhost:8080"`. In Docker, nginx proxies `/api/` and `/api/files/` to `http://backend:8080`, so the frontend uses same-origin and `VITE_API_URL` is empty at build time.
- Auth: token stored in `localStorage` under key `token`. Login response shape is `{ token: { access_token }, user }` (see `src/context/AuthContext.tsx`). Update both the `api.login` call site and the `AuthContext.login` consumer if this ever changes.
- Route protection: `src/components/ProtectedRoute.tsx` accepts an optional `roles` array. Use it for `creator`-only (`/templates/my`) and `superadmin`-only (`/admin`) routes — see `src/App.tsx`.
- Build: `npm run build` runs `tsc -b && vite build`. Type errors break the build, there is no separate typecheck step.
- No frontend tests, no ESLint, no Prettier. Don't assume the standard Vite + React TS template scripts are present.
- `vite.config.ts` whitelists `allowedHosts: ["syncspaceedu.duskoide.org"]` — required for the Vite dev server to serve the production host. Add new public hostnames here too.

## Commands

- Dev (both services, separate DB): `devbox run dev` from repo root.
- Backend tests: `cd backend && go test ./...` (~40s, all in `internal/service`).
- Run a single test: `cd backend && go test ./internal/service/ -run TestName -v`.
- Frontend typecheck + build: `cd frontend && npx tsc -b && npx vite build`.
- Stack rebuild: `docker-compose up --build -d`. Tunnel profile: `docker-compose --profile tunnel up --build -d`.
- Wipe state: `docker-compose down -v && rm -rf data/`.

## Things an agent typically gets wrong here

- Editing the `migrate` block or `seedDefaultUsers` will re-run destructive SQL / reseed on every container restart. The `data/` volume persists between restarts, so a "wipe" needs `docker-compose down -v` plus `rm -rf data/`.
- Adding a new API route: register on the right mux (`mux` for public, `authMux` for authenticated, then wrap with `AuthMiddleware` / `RequireRole` in the `mux.Handle` lines at the bottom of `Register`). Skipping the wrap leaves the route public.
- The frontend `api.ts` always sets `Content-Type: application/json` unless the body is `FormData`. For new file uploads, pass `FormData` directly.
- `docker-compose.yml` mounts both `./data` and `./uploads`. DB lives under `./data/syncspace.db`; uploads go to `./uploads` (via `SYNCSPACE_UPLOAD_DIR=/uploads`). Don't put uploads in `data/`.
- There is no pre-commit, no CI, no linter, and no test runner for the frontend. Don't try to invoke them.
