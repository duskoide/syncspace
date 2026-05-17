# Deployment Guide (Tuxedo OS + aaPanel + Cloudflare Tunnel)

## A. Backend local run
1. `cd backend`
2. `go run ./cmd/syncspace`

Default env:
- `SYNCSPACE_ADDR=:8080`
- `SYNCSPACE_DB_PATH=../data/syncspace.db`

## B. Frontend local run
1. `cd frontend`
2. `npm install`
3. `cp .env.example .env`
4. `npm run dev -- --host --port 5173`

## C. aaPanel target mapping
- Backend app route: `http://127.0.0.1:8080`
- Frontend app route: `http://127.0.0.1:5173` or built static files from `frontend/build` via aaPanel site root

## D. Security steps (per assignment)
Run `sudo bt` and configure:
- username/password
- security entrance path
- panel port
- `sudo ufw allow <panel_port>`
- SSL toggle depending on tunnel redirect behavior

## E. Cloudflare Tunnel
1. Create tunnel under Zero Trust.
2. Add public hostname routes to local services/ports.
3. Verify tunnel status is `HEALTHY`.
4. Open public URL and capture screenshots for submission.
