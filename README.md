# SyncSpace - Collaborative Whiteboard & Forum Platform

SyncSpace is a self-hosted collaborative whiteboard and forum platform built for the II2210 Teknologi Platform course assignment (Tugas 2).

## Theme
Education & Learning Tools - Collaborative Whiteboard Forum

## Stack
- **Backend:** Go 1.23 + REST API + WebSocket
- **Frontend:** React 18 + Vite + TypeScript + React Router
- **Database:** SQLite (WAL mode)
- **Authentication:** JWT with bcrypt password hashing
- **Real-time:** WebSocket for live collaboration
- **Containerization:** Docker + Docker Compose

## Features

### Three User Roles (Assignment Requirement)
1. **Superadmin** - Platform administrator, approve/suspend users, manage all boards
2. **Moderator** - Create and manage boards, invite collaborators, moderate content
3. **Collaborator** - Join boards, add text elements, participate in discussions

### Core Features
- **User Registration & Approval** - New users register with pending status, superadmin approves before access
- **Board Management** - Moderators create collaborative whiteboards with descriptions
- **Board Memberships** - Join boards as viewer or editor, leave boards anytime
- **Collaborative Text Elements** - Add draggable sticky notes anywhere on the whiteboard
- **Real-time Collaboration** - WebSocket-powered live updates for text elements and discussions
- **Discussion Threads** - Threaded conversations attached to each board
- **File Uploads** - Support for images, videos, PDFs, documents (max 10MB)

### Producer-Consumer Interaction Flow
**Moderator (Producer):**
- Creates a board (whiteboard space)
- Manages board memberships and collaborator roles
- Can edit/delete any content on their boards

**Collaborator (Consumer):**
- Joins boards via membership
- Adds text elements (sticky notes) to the whiteboard
- Participates in board discussions
- Can edit their own content

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
- `POST /api/auth/register` - Register new user (roles: moderator/collaborator, status: pending)
- `POST /api/auth/login` - Login and receive JWT
- `GET /api/auth/me` - Get current user info

### Admin (Superadmin only)
- `GET /api/admin/users` - List all users
- `PUT /api/admin/users/{id}/approve` - Approve pending user
- `PUT /api/admin/users/{id}/suspend` - Suspend user
- `DELETE /api/admin/users/{id}` - Delete user

### Boards
- `GET /api/boards` - List my boards (all boards for superadmin)
- `POST /api/boards` - Create board (moderator/superadmin)
- `GET /api/boards/{id}` - Get board details
- `PUT /api/boards/{id}` - Update board (moderator only)
- `DELETE /api/boards/{id}` - Delete board (moderator only)
- `POST /api/boards/{id}/join` - Join board (collaborator)
- `DELETE /api/boards/{id}/leave` - Leave board
- `GET /api/boards/{id}/members` - List board members

### Board Memberships
- `PUT /api/memberships/{id}/role` - Update member role (moderator only)
- `DELETE /api/boards/{id}/members/{member_id}` - Remove member (moderator only)

### Text Elements (Whiteboard Notes)
- `GET /api/text-elements?board_id={id}` - List all notes on a board
- `POST /api/text-elements` - Create new note (editor role required)
- `GET /api/text-elements/{id}` - Get note details
- `PUT /api/text-elements/{id}` - Update note content/position
- `DELETE /api/text-elements/{id}` - Delete note

### Discussions
- `GET /api/discussions?board_id={id}` - List board discussions
- `POST /api/discussions` - Post message to board
- `GET /api/discussions/{id}/replies` - Get discussion replies
- `DELETE /api/discussions/{id}` - Delete discussion

### WebSocket
- `GET /ws?token={jwt}&room=board_{id}` - Real-time collaboration

### Files
- `POST /api/upload` - Upload file
- `GET /api/files/{id}` - Download file

## Database Schema

### Core Tables
- `users` - Authentication & roles (superadmin, moderator, collaborator)
- `boards` - Collaborative whiteboard spaces
- `board_memberships` - User-board relationships with roles (viewer/editor)
- `text_elements` - Draggable sticky notes (x, y, content, color)
- `discussions` - Threaded chat messages per board
- `attachments` - File metadata

## Interaction Sequences

### Sequence 1: Moderator Creates Board, Collaborator Joins and Adds Content
```
1. Moderator creates Board
2. Moderator shares board access (or it's public)
3. Collaborator discovers/joins Board
4. Collaborator adds TextElement to whiteboard
5. Real-time sync broadcasts to all connected users
6. Moderator and Collaborator see updates live
```

### Sequence 2: Discussion Thread
```
1. Any member posts Discussion message
2. WebSocket broadcasts to board room
3. All connected members see message instantly
4. Members can reply creating threaded discussions
```

## Architecture
Layered architecture pattern:
- **Presentation Layer** - React frontend with whiteboard canvas
- **API Layer** - Go HTTP handlers + WebSocket hub
- **Service Layer** - Business logic (boards, memberships, text elements)
- **Data Access Layer** - SQLite store with WAL mode
- **Data Layer** - SQLite + File system for uploads

## UI Design
- Dark theme with glass-morphism cards
- Collaborative whiteboard with grid background
- Draggable sticky notes with color coding
- Real-time status indicators
- Responsive layout for desktop and mobile

## References
- Go net/http documentation - https://pkg.go.dev/net/http
- JWT authentication pattern - https://datatracker.ietf.org/doc/html/rfc7519
- Layered architecture - "Patterns of Enterprise Application Architecture" by Martin Fowler
- SQLite WAL mode - https://sqlite.org/wal.html
- WebSocket API - https://developer.mozilla.org/en-US/docs/Web/API/WebSocket

## License
Academic project for II2210 Teknologi Platform - Institut Teknologi Bandung.
