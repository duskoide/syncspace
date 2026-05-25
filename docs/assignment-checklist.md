# II2210 Tugas 2 Checklist (SyncSpace Edu)

## Requirements Mapping

| Requirement | Implementation | Status |
|------------|----------------|--------|
| **Access Control & Auth (10 pts)** | JWT auth, 3 roles, approval workflow | ✅ |
| **Database & Backend (15 pts)** | SQLite, file storage, cited architecture | ✅ |
| **Interaction Flows (25 pts)** | Sequence diagrams in implementation plan | ✅ |
| **Functionality (20 pts)** | System flowcharts, function table | ✅ |
| **Unique Feature (10 pts)** | Collaborative notes + Real-time discussions | ✅ |
| **References (10 pts)** | Architecture citations in README | ✅ |
| **Dockerization +20 pts** | Docker Compose, multi-container | ✅ |
| **Public Access (5 pts)** | Deployed platform link | ⏳ |
| **Screenshots (5 pts)** | Platform documentation | ⏳ |
| **TOTAL** | | **120 pts** |

## Detailed Checklist

### Authentication & Authorization
- [x] User registration with role selection (teacher/student)
- [x] Password hashing with bcrypt
- [x] JWT token generation and validation
- [x] User approval workflow (pending -> active)
- [x] Role-based access control (superadmin/teacher/student)
- [x] Protected routes middleware

### Database
- [x] SQLite with WAL mode
- [x] Foreign key constraints
- [x] Indexes on frequently queried columns
- [x] Self-managed (no external DB service)
- [x] Hierarchical file storage (/uploads/)

### Producer-Consumer Flows
- [x] Teacher creates classroom
- [x] Student requests enrollment
- [x] Teacher approves enrollment
- [x] Teacher uploads materials
- [x] Student accesses materials
- [x] Teacher creates assignments
- [x] Student submits work
- [x] Teacher grades submissions

### Unique Features
- [x] Collaborative notes (classroom-wide shared notes)
- [x] Real-time discussions via WebSocket
- [x] Live connection status indicator
- [x] Room-based message broadcasting

### Docker
- [x] Dockerfile.backend (multi-stage Go build)
- [x] Dockerfile.frontend (Node build + nginx)
- [x] docker-compose.yml (orchestration)
- [x] Successfully tested docker-compose up

### Frontend
- [x] Login page
- [x] Registration page
- [x] Role-based dashboard
- [x] Classroom management
- [x] Admin user approval interface
- [x] Materials/Assignments/Discussions tabs
- [x] WebSocket integration for real-time chat

### Documentation
- [x] Sequence diagrams (in TUGAS2_IMPLEMENTATION_PLAN.md)
- [x] System flowcharts
- [x] Function table
- [x] Architecture citations
- [x] Updated README.md

## Deployment
- [ ] Deploy to production server
- [ ] Configure Cloudflare Tunnel
- [ ] Take screenshots for documentation
- [ ] Submit platform link

## Notes
- Default superadmin: admin@syncspace.edu / admin123
- File upload limit: 10MB
- Supported file types: images, videos, PDFs, documents
- WebSocket endpoint: /ws?token={jwt}&room=classroom_{id}
