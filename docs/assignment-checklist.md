# II2210 Tugas 1 Checklist (SyncSpace Edu)

## 1) Frontend aplikasi
- React + Vite app at `frontend/`
- Public UI for task/note CRUD and wiki enrichment.

## 2) Backend service
- Go REST API at `backend/cmd/syncspace`.

## 3) Database dikelola sendiri
- SQLite database file in `data/syncspace.db`.
- WAL mode enabled via startup pragma.

## 4) API eksternal/publik
- Wikipedia summary API integration in service layer.

## 5) Akses public via internet
- Planned via Cloudflare Tunnel to backend/frontend service ports.
- aaPanel used for service/web management in native Tuxedo OS.
