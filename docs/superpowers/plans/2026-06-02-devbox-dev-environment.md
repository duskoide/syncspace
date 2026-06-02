# Devbox-based Local Development Environment Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace `dev.sh` and the implicit "install Go + Node + sqlite yourself" setup with a `devbox.json`-driven dev environment at the repo root. Production Docker stack is unchanged.

**Architecture:** A single `devbox.json` declares the toolchain (Go 1.23, Node 20, sqlite), env vars for the two dev services, and two long-running services (backend + frontend) that are started together by `devbox run dev`. `dev.sh` is deleted. README and AGENTS.md are updated to point at devbox.

**Tech Stack:** devbox 0.x (Nix-based dev environment), Go 1.23, Node 20, sqlite (CLI).

**Reference spec:** `docs/superpowers/specs/2026-06-02-devbox-dev-environment-design.md`

---

## File Structure

Files this plan creates or modifies:

- `devbox.json` (Create) — devbox environment config. Packages, env vars, services, scripts.
- `README.md` (Modify) — replace the "Manual Development" subsection with a "Local Development (devbox)" subsection.
- `AGENTS.md` (Modify) — update two bullets that mention `dev.sh`; add one new gotcha about devbox path resolution.
- `dev.sh` (Delete) — replaced by `devbox run dev`.

No source files in `backend/` or `frontend/src/` are touched. No Dockerfile, `docker-compose.yml`, or `nginx.conf` is touched.

---

## Task 1: Create `devbox.json`

**Files:**
- Create: `devbox.json`

- [ ] **Step 1: Write the file**

Write `devbox.json` at the repo root with the following exact content:

```json
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
    "dev": "devbox services up",
    "stop": "devbox services stop",
    "test": "cd backend && go test ./... && cd ../frontend && npx tsc -b",
    "build": "cd frontend && npm run build",
    "logs": "devbox services logs -f"
  }
}
```

Notes for the implementer:
- Path env vars (`SYNCSPACE_DB_PATH`, `SYNCSPACE_UPLOAD_DIR`) start with `../dev-data/...` because the backend service's `directory` is `backend/`, so paths are resolved relative to that. This matches what the old `dev.sh` did after its `cd backend`.
- The backend service's `process` starts with `mkdir -p ../dev-data` because `shell.init_hook` only runs on `devbox shell` entry, not on `devbox services up`. The mkdir is idempotent.

- [ ] **Step 2: Verify the JSON parses**

Run: `python3 -m json.tool devbox.json > /dev/null && echo OK`
Expected: `OK`

If it prints anything else, the JSON is malformed. Fix the syntax and re-run.

- [ ] **Step 3: Commit**

```bash
cd /home/pn/Projects/syncspace
git add devbox.json
git commit -m "feat: add devbox.json for local dev environment"
```

---

## Task 2: Replace the README's "Manual Development" section

**Files:**
- Modify: `README.md` (the block from `### Manual Development` through the end of the "To run services manually in separate terminals" code block, currently lines 93-120)

- [ ] **Step 1: Find the exact text to replace**

In `README.md`, locate the subsection that starts with `### Manual Development` and ends right before the next `## Default Credentials` heading. The block to replace reads:

```markdown
### Manual Development

Run both backend and frontend in development mode with a separate database:

```bash
chmod +x dev.sh
./dev.sh
```

| Service | URL | Notes |
|---------|-----|-------|
| Frontend | http://localhost:5173 | Vite dev server with HMR |
| Backend | http://localhost:8081 | Go API with hot-reload disabled |
| Database | `./dev-data/syncspace.db` | Separate from production |

Production (Docker on `:3000`) stays running — no conflict.

To run services manually in separate terminals:

```bash
# Terminal 1: Backend
cd backend
SYNCSPACE_ADDR=:8081 SYNCSPACE_DB_PATH=../dev-data/syncspace.db go run ./cmd/syncspace

# Terminal 2: Frontend
cd frontend
VITE_API_URL=http://localhost:8081 npm run dev
```
```

(Including the surrounding triple-backtick fences.)

- [ ] **Step 2: Replace it with the new devbox section**

Replace the entire block with:

```markdown
### Local Development (devbox)

SyncSpace uses [devbox](https://www.jetpack.io/devbox) to provide a reproducible dev environment. Install it once: `curl -fsSL https://get.jetpack.io/devbox | bash`.

```bash
devbox run dev
```

| Service | URL | Notes |
|---------|-----|-------|
| Frontend | http://localhost:5173 | Vite dev server with HMR |
| Backend | http://localhost:8081 | Go API |
| Database | `./dev-data/syncspace.db` | Auto-created, separate from prod |

Other scripts:

- `devbox run stop` — stop both services
- `devbox run test` — backend Go tests + frontend typecheck
- `devbox run build` — production frontend build
- `devbox run logs` — tail logs from a separate shell
- `devbox shell` — drop into a shell with Go/Node/sqlite on PATH

Production (Docker on `:3000`) stays running — no conflict.
```

- [ ] **Step 3: Verify no `dev.sh` or `chmod` references remain in the README**

Run: `grep -n "dev.sh\|chmod" README.md`
Expected: no output (exit code 1).

If anything matches, you missed a spot. Edit and re-run.

- [ ] **Step 4: Commit**

```bash
cd /home/pn/Projects/syncspace
git add README.md
git commit -m "docs: replace Manual Development section with devbox instructions"
```

---

## Task 3: Update `AGENTS.md` to reference devbox

**Files:**
- Modify: `AGENTS.md` (two bullets in the "Layout" and "Commands" sections)

- [ ] **Step 1: Update the Layout bullet**

In `AGENTS.md`, find the bullet:

```markdown
- `dev.sh` — local dev: backend `:8081` + frontend `:5173`, separate DB at `./dev-data/syncspace.db`. Does not conflict with the Docker stack on `:3000`.
```

Replace it with:

```markdown
- `devbox.json` — Nix-based dev environment. Same ports and DB layout as before (`:8081` / `:5173` / `./dev-data/`), no Docker required locally.
```

- [ ] **Step 2: Update the Commands bullet**

In `AGENTS.md`, find the line:

```markdown
- Dev (both services, separate DB): `./dev.sh` from repo root.
```

Replace it with:

```markdown
- Dev (both services, separate DB): `devbox run dev` from repo root.
```

- [ ] **Step 3: Verify no `dev.sh` references remain in AGENTS.md**

Run: `grep -n "dev.sh" AGENTS.md`
Expected: no output (exit code 1).

If anything matches, you missed a spot. Edit and re-run.

- [ ] **Step 4: Commit**

```bash
cd /home/pn/Projects/syncspace
git add AGENTS.md
git commit -m "docs: update AGENTS.md to reference devbox instead of dev.sh"
```

---

## Task 4: Add a devbox gotcha to AGENTS.md

**Files:**
- Modify: `AGENTS.md` (append a bullet to the "Things an agent typically gets wrong here" section)

- [ ] **Step 1: Find the "Things an agent typically gets wrong here" section**

In `AGENTS.md`, find the closing list of that section. It ends with the bullet about the absence of pre-commit, CI, linter, and frontend test runner.

- [ ] **Step 2: Append a new bullet**

Add a new bullet at the end of the list:

```markdown
- `devbox.json` services run with `directory` as the working dir, so path env vars (`SYNCSPACE_DB_PATH`, `SYNCSPACE_UPLOAD_DIR`) are resolved relative to that dir, not the repo root. The backend service runs from `backend/`, so paths use `../dev-data/...`. `shell.init_hook` runs only on `devbox shell` entry, not on `devbox services up` — that's why the backend service's `process` starts with `mkdir -p ../dev-data`.
```

- [ ] **Step 3: Commit**

```bash
cd /home/pn/Projects/syncspace
git add AGENTS.md
git commit -m "docs: document devbox path-resolution gotcha in AGENTS.md"
```

---

## Task 5: Delete `dev.sh`

**Files:**
- Delete: `dev.sh`

- [ ] **Step 1: Remove the file**

Run: `rm dev.sh`

- [ ] **Step 2: Verify it's gone**

Run: `ls dev.sh 2>&1`
Expected: `ls: cannot access 'dev.sh': No such file or directory` (exit code 2).

- [ ] **Step 3: Verify no other tracked file references `dev.sh`**

Run:
```bash
git grep -l "dev.sh" -- ':!docs/superpowers/specs/*' ':!docs/superpowers/plans/*'
```
Expected: no output (exit code 1, meaning no tracked file outside the spec/plan dirs references it).

If any file matches, edit it to remove the reference, then re-run the grep.

- [ ] **Step 4: Commit**

```bash
cd /home/pn/Projects/syncspace
git add -u dev.sh
git commit -m "chore: remove dev.sh, replaced by devbox run dev"
```

---

## Task 6: Final verification

**Files:** none (read-only checks)

- [ ] **Step 1: Backend tests still pass**

Run: `cd backend && go test ./...`
Expected: ends with `ok  syncspace/backend/internal/service  <duration>`. Other packages show `[no test files]`, which is fine.

- [ ] **Step 2: Frontend typecheck still passes**

Run: `cd frontend && npx tsc -b`
Expected: exit code 0, no output.

- [ ] **Step 3: Git status is clean**

Run: `cd /home/pn/Projects/syncspace && git status`
Expected: `nothing to commit, working tree clean` (or only untracked files like `dev-data/`, `frontend/node_modules/`, etc., which are gitignored).

- [ ] **Step 4: Review the commit log**

Run: `git log --oneline -10`
Expected: the five new commits from this plan appear at the top:
1. `feat: add devbox.json for local dev environment`
2. `docs: replace Manual Development section with devbox instructions`
3. `docs: update AGENTS.md to reference devbox instead of dev.sh`
4. `docs: document devbox path-resolution gotcha in AGENTS.md`
5. `chore: remove dev.sh, replaced by devbox run dev`

(Plus the earlier `docs: add design spec for devbox-based dev environment` commit if you count the spec.)

If any step fails, stop and fix before claiming the plan is done.
