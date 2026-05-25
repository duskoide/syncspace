# SyncSpace Edu - Collaborative Learning Platform

SyncSpace Edu is a self-hosted collaborative learning platform built for the II2210 Teknologi Platform course assignment (Tugas 2).

## Theme
Education & Learning Tools - Teacher-Student Collaborative Platform

## Stack
- **Backend:** Go 1.23 + REST API + WebSocket
- **Frontend:** React 18 + Vite + TypeScript + React Router
- **Database:** SQLite (WAL mode)
- **Authentication:** JWT with bcrypt password hashing
- **Real-time:** WebSocket (Gorilla)
- **Containerization:** Docker + Docker Compose

## Features

### Three User Roles
1. **Superadmin** - Approve/suspend users, manage all content
2. **Teacher** - Create classrooms, upload materials, create assignments, grade submissions
3. **Student** - Enroll in classrooms, access materials, submit assignments, participate in discussions

### Core Features
- **User Registration & Approval** - New users register with pending status, superadmin approves
- **Classroom Management** - Teachers create classrooms, students request enrollment
- **Learning Materials** - Teachers upload materials with file attachments
- **Assignments & Submissions** - Teachers create assignments with deadlines, students submit work
- **Grading System** - Teachers grade submissions with scores and feedback
- **Collaborative Notes** - Shared notes within classrooms
- **Real-time Discussions** - WebSocket-powered classroom chat with live updates
- **File Uploads** - Support for images, videos, PDFs, documents (max 10MB)

## Quickstart

### With Docker Compose (Recommended)
```bash
docker-compose up --build
```
- Backend: http://localhost:8080
- Frontend: http://localhost:3000

### Manual Development

#### Backend
```bash
cd backend
go mod download
go run ./cmd/syncspace
```

#### Frontend
```bash
cd frontend
npm install
npm run dev
```

## Default Credentials
- **Superadmin:** `admin@syncspace.edu` / `admin123`

## API Endpoints

### Authentication
- `POST /api/auth/register` - Register new user (status: pending)
- `POST /api/auth/login` - Login and receive JWT
- `GET /api/auth/me` - Get current user info

### Admin (Superadmin only)
- `GET /api/admin/users` - List all users
- `PUT /api/admin/users/{id}/approve` - Approve pending user
- `PUT /api/admin/users/{id}/suspend` - Suspend user

### Classrooms
- `GET /api/classrooms` - List my classrooms
- `POST /api/classrooms` - Create classroom (teacher)
- `GET /api/classrooms/{id}` - Get classroom details
- `PUT /api/classrooms/{id}` - Update classroom
- `DELETE /api/classrooms/{id}` - Delete classroom
- `POST /api/classrooms/{id}/enroll` - Request enrollment (student)
- `GET /api/classrooms/{id}/students` - List enrolled students

### Materials
- `GET /api/materials?classroom_id={id}` - List materials
- `POST /api/materials` - Create material (teacher)
- `GET /api/materials/{id}` - Get material
- `PUT /api/materials/{id}` - Update material
- `DELETE /api/materials/{id}` - Delete material

### Assignments
- `GET /api/assignments?classroom_id={id}` - List assignments
- `POST /api/assignments` - Create assignment (teacher)
- `GET /api/assignments/{id}` - Get assignment
- `POST /api/assignments/{id}/submissions` - Submit work (student)
- `GET /api/assignments/{id}/submissions` - List submissions (teacher)
- `PUT /api/submissions/{id}/grade` - Grade submission (teacher)

### Discussions
- `GET /api/discussions?classroom_id={id}` - List discussions
- `POST /api/discussions` - Post message

### WebSocket
- `GET /ws?token={jwt}&room=classroom_{id}` - Real-time discussion

### Files
- `POST /api/upload` - Upload file
- `GET /api/files/{id}` - Download file

## Database Schema
- `users` - Authentication & roles
- `classrooms` - Learning spaces
- `enrollments` - Student-classroom relationships
- `materials` - Learning content
- `attachments` - File metadata
- `assignments` - Tasks with deadlines
- `submissions` - Student work
- `collaborative_notes` - Shared notes
- `discussions` - Chat messages

## Architecture
Layered architecture pattern:
- Presentation Layer (React frontend)
- API Layer (Go HTTP handlers + WebSocket)
- Service Layer (Business logic)
- Data Access Layer (SQLite store)
- Data Layer (SQLite + File system)

## References
- Go net/http documentation - https://pkg.go.dev/net/http
- Gorilla WebSocket - https://github.com/gorilla/websocket
- JWT authentication pattern - https://datatracker.ietf.org/doc/html/rfc7519
- Layered architecture - "Patterns of Enterprise Application Architecture" by Martin Fowler
- SQLite WAL mode - https://sqlite.org/wal.html

## License
Academic project for II2210 Teknologi Platform.
