# SyncSpace Edu - Tugas 2 Implementation Plan

## II2210 - Teknologi Platform

---

## Table of Contents

1. [Project Overview](#project-overview)
2. [Requirements Mapping](#requirements-mapping)
3. [System Architecture](#system-architecture)
4. [Database Schema](#database-schema)
5. [Backend Implementation](#backend-implementation)
6. [Frontend Implementation](#frontend-implementation)
7. [Real-time Features](#real-time-features)
8. [Sequence Diagrams](#sequence-diagrams)
9. [System Flowcharts](#system-flowcharts)
10. [Function Table](#function-table)
11. [Unique Feature: Collaborative Learning Suite](#unique-feature-collaborative-learning-suite)
12. [Dockerization](#dockerization)
13. [File Storage Strategy](#file-storage-strategy)
14. [Deployment Plan](#deployment-plan)
15. [Timeline & Checklist](#timeline--checklist)
16. [References](#references)

---

## Project Overview

**Platform Name:** SyncSpace Edu - Collaborative Learning Platform  
**Theme:** Education & Learning Tools (Teacher-Student)  
**Previous State:** Simple task/note manager with Wikipedia enrichment  
**Target State:** Full-featured collaborative learning platform

### Three User Roles

1. **Superadmin**
   - Approve/suspend user registrations
   - Create classrooms
   - Manage all platform content
   - Full system access

2. **Teacher (Producer)**
   - Create and manage classrooms
   - Upload learning materials with file attachments
   - Create assignments with deadlines
   - Grade student submissions
   - Approve student enrollment requests

3. **Student (Consumer)**
   - Request enrollment in classrooms
   - Access learning materials and download files
   - Submit assignments before deadlines
   - Participate in collaborative notes
   - Engage in real-time discussions

---

## Requirements Mapping

| Tugas 2 Requirement | Implementation | Points |
|-------------------|----------------|--------|
| **Access Control & Auth** | JWT-based auth, 3 roles, approval workflow | 10 |
| **Database & Backend** | Self-managed SQLite, file storage, cited architecture | 15 |
| **Interaction Flows** | 2+ sequence diagrams (Producer↔Consumer) | 25 |
| **Functionality** | System flowcharts, function table | 20 |
| **Unique Feature** | Collaborative notes + Real-time discussions | 10 |
| **References** | Architecture citations | 10 |
| **Dockerization (Bonus)** | Docker Compose, multi-container setup | +20 |
| **Public Access** | Deployed platform link | 5 |
| **Screenshots** | Platform documentation | 5 |
| **TOTAL** | | **120** |

---

## System Architecture

### Layered Architecture Pattern

```
┌─────────────────────────────────────────────────────────────┐
│                     Presentation Layer                       │
│              React + Vite + TypeScript Frontend             │
│         ┌─────────────┐  ┌─────────────┐  ┌────────────┐   │
│         │  Auth Pages │  │   Teacher   │  │  Student   │   │
│         │             │  │   Portal    │  │   Portal   │   │
│         └─────────────┘  └─────────────┘  └────────────┘   │
└─────────────────────────────────────────────────────────────┘
                              │ HTTP/WebSocket
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      API Layer (Go)                          │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │  REST API   │  │  WebSocket  │  │  Auth Middleware    │  │
│  │  Handlers   │  │   Server    │  │  (JWT + RBAC)       │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   Service Layer (Business Logic)             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │   Auth      │  │  Classroom  │  │  File Management    │  │
│  │  Service    │  │   Service   │  │     Service         │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │  Material   │  │  Assignment │  │  Real-time (WS)     │  │
│  │  Service    │  │   Service   │  │     Service         │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Data Access Layer                         │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │   User      │  │  Classroom  │  │  Material Store     │  │
│  │   Store     │  │   Store     │  │                     │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │  Assignment │  │ Submission  │  │ Collaborative Note  │  │
│  │   Store     │  │   Store     │  │     Store           │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      Data Layer                              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │   SQLite    │  │   Local     │  │   In-Memory Cache   │  │
│  │  Database   │  │ File System │  │   (WebSocket Hub)   │  │
│  │  (WAL Mode) │  │ (Hierachial)│  │                     │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### Technology Stack

**Backend:**
- Go 1.22
- Standard library `net/http` with `http.ServeMux`
- `modernc.org/sqlite` (pure Go SQLite driver)
- Gorilla WebSocket (or native implementation)
- JWT authentication (`github.com/golang-jwt/jwt/v5`)
- bcrypt password hashing (`golang.org/x/crypto/bcrypt`)

**Frontend:**
- React 18
- TypeScript 5
- Vite (build tool)
- Native CSS (no UI framework)
- WebSocket client API

**Infrastructure:**
- SQLite with WAL mode
- Local file storage (hierarchical)
- Docker + Docker Compose
- Nginx reverse proxy

---

## Database Schema

### Core Tables

```sql
-- Users Table (Authentication & Roles)
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    name TEXT NOT NULL,
    role TEXT NOT NULL CHECK(role IN ('superadmin', 'teacher', 'student')),
    status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending', 'active', 'suspended')),
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

-- Classrooms Table (Learning Spaces)
CREATE TABLE classrooms (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,
    teacher_id INTEGER NOT NULL REFERENCES users(id),
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

-- Enrollments Table (Student-Classroom Relationships)
CREATE TABLE enrollments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    classroom_id INTEGER NOT NULL REFERENCES classrooms(id),
    student_id INTEGER NOT NULL REFERENCES users(id),
    status TEXT DEFAULT 'pending' CHECK(status IN ('pending', 'active', 'inactive')),
    enrolled_at TEXT NOT NULL,
    UNIQUE(classroom_id, student_id)
);

-- Materials Table (Learning Content - Producer creates)
CREATE TABLE materials (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    classroom_id INTEGER NOT NULL REFERENCES classrooms(id),
    teacher_id INTEGER NOT NULL REFERENCES users(id),
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    tags TEXT DEFAULT '',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

-- Attachments Table (File Metadata)
CREATE TABLE attachments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    material_id INTEGER REFERENCES materials(id),
    submission_id INTEGER REFERENCES submissions(id),
    filename TEXT NOT NULL,
    original_name TEXT NOT NULL,
    mime_type TEXT NOT NULL,
    file_size INTEGER NOT NULL,
    file_path TEXT NOT NULL,
    uploaded_by INTEGER NOT NULL REFERENCES users(id),
    created_at TEXT NOT NULL
);

-- Assignments Table (Tasks with Deadlines)
CREATE TABLE assignments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    classroom_id INTEGER NOT NULL REFERENCES classrooms(id),
    teacher_id INTEGER NOT NULL REFERENCES users(id),
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    due_date TEXT NOT NULL,
    max_score INTEGER DEFAULT 100,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

-- Submissions Table (Student Work)
CREATE TABLE submissions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    assignment_id INTEGER NOT NULL REFERENCES assignments(id),
    student_id INTEGER NOT NULL REFERENCES users(id),
    content TEXT NOT NULL,
    score INTEGER,
    feedback TEXT,
    submitted_at TEXT NOT NULL,
    graded_at TEXT,
    UNIQUE(assignment_id, student_id)
);

-- Collaborative Notes Table (Unique Feature)
CREATE TABLE collaborative_notes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    material_id INTEGER REFERENCES materials(id),
    classroom_id INTEGER NOT NULL REFERENCES classrooms(id),
    created_by INTEGER NOT NULL REFERENCES users(id),
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    is_public BOOLEAN DEFAULT 1,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

-- Note Contributors Table (Collaborative Editing Tracking)
CREATE TABLE note_contributors (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    note_id INTEGER NOT NULL REFERENCES collaborative_notes(id),
    user_id INTEGER NOT NULL REFERENCES users(id),
    contributed_at TEXT NOT NULL,
    UNIQUE(note_id, user_id)
);

-- Discussions Table (Real-time Chat)
CREATE TABLE discussions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    classroom_id INTEGER NOT NULL REFERENCES classrooms(id),
    material_id INTEGER REFERENCES materials(id),
    user_id INTEGER NOT NULL REFERENCES users(id),
    message TEXT NOT NULL,
    parent_id INTEGER REFERENCES discussions(id),
    created_at TEXT NOT NULL
);
```

### Database Features

- **WAL Mode:** Enabled for concurrent read/write
- **Foreign Keys:** Referential integrity
- **Busy Timeout:** 5000ms to handle concurrent access
- **Indexes:** Added on frequently queried columns
- **Timestamps:** RFC3339 format for consistency

---

## Backend Implementation

### Models

**User Model** (`internal/models/user.go`):
```go
type User struct {
    ID           int64     `json:"id"`
    Email        string    `json:"email"`
    Name         string    `json:"name"`
    Role         string    `json:"role"`    // superadmin, teacher, student
    Status       string    `json:"status"`  // pending, active, suspended
    PasswordHash string    `json:"-"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}

type LoginRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

type RegisterRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
    Name     string `json:"name"`
    Role     string `json:"role"` // teacher or student
}
```

**Classroom Model** (`internal/models/classroom.go`):
```go
type Classroom struct {
    ID          int64     `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    TeacherID   int64     `json:"teacher_id"`
    TeacherName string    `json:"teacher_name,omitempty"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type Enrollment struct {
    ID          int64     `json:"id"`
    ClassroomID int64     `json:"classroom_id"`
    StudentID   int64     `json:"student_id"`
    StudentName string    `json:"student_name,omitempty"`
    Status      string    `json:"status"`
    EnrolledAt  time.Time `json:"enrolled_at"`
}
```

### API Endpoints

#### Authentication
```
POST /api/auth/register     - Register new user (status: pending)
POST /api/auth/login        - Authenticate and return JWT
POST /api/auth/refresh      - Refresh JWT token
GET  /api/auth/me           - Get current user info
POST /api/auth/logout       - Invalidate token
```

#### User Management (Superadmin)
```
GET    /api/admin/users              - List all users
PUT    /api/admin/users/{id}/approve - Approve pending user
PUT    /api/admin/users/{id}/suspend - Suspend user
DELETE /api/admin/users/{id}         - Delete user
```

#### Classrooms
```
POST   /api/classrooms                    - Create classroom
GET    /api/classrooms                    - List user's classrooms
GET    /api/classrooms/{id}               - Get classroom details
PUT    /api/classrooms/{id}               - Update classroom
DELETE /api/classrooms/{id}               - Delete classroom
POST   /api/classrooms/{id}/enroll        - Request enrollment
PUT    /api/classrooms/{id}/enrollments/{enrollment_id}/approve - Approve enrollment
GET    /api/classrooms/{id}/students      - List enrolled students
DELETE /api/classrooms/{id}/students/{student_id} - Remove student
```

#### Materials
```
POST   /api/materials          - Create material with attachments
GET    /api/materials          - List materials in classroom
GET    /api/materials/{id}     - Get material with attachments
PUT    /api/materials/{id}     - Update material
DELETE /api/materials/{id}     - Delete material
```

#### Assignments
```
POST   /api/assignments                    - Create assignment
GET    /api/assignments                    - List assignments
GET    /api/assignments/{id}               - Get assignment details
PUT    /api/assignments/{id}               - Update assignment
DELETE /api/assignments/{id}               - Delete assignment
POST   /api/assignments/{id}/submissions   - Submit work (student)
GET    /api/assignments/{id}/submissions   - List submissions (teacher)
GET    /api/submissions/{id}               - Get submission
PUT    /api/submissions/{id}/grade         - Grade submission (teacher)
```

#### Files
```
POST /api/upload              - Upload file(s)
GET  /api/files/{id}          - Download file
GET  /api/files/{id}/metadata - Get file metadata
```

#### Collaborative Notes
```
POST   /api/collaborative-notes              - Create note
GET    /api/collaborative-notes              - List notes in classroom
GET    /api/collaborative-notes/{id}         - Get note with contributors
PUT    /api/collaborative-notes/{id}         - Update note content
DELETE /api/collaborative-notes/{id}         - Delete note
GET    /api/collaborative-notes/{id}/history - View version history
WS     /ws/notes/{id}                       - Real-time collaboration
```

#### Discussions
```
GET  /api/classrooms/{id}/discussions        - Get discussion messages
POST /api/classrooms/{id}/discussions        - Post message
GET  /api/discussions/{id}/replies           - Get threaded replies
WS   /ws/discussions/{classroom_id}          - Real-time chat
```

### Middleware

**Authentication Middleware** (`internal/api/middleware.go`):
```go
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := extractToken(r)
        claims, err := validateJWT(token)
        if err != nil {
            writeError(w, 401, "unauthorized", "invalid token")
            return
        }
        ctx := context.WithValue(r.Context(), "user", claims)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func RequireRole(roles ...string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            user := r.Context().Value("user").(*Claims)
            for _, role := range roles {
                if user.Role == role {
                    next.ServeHTTP(w, r)
                    return
                }
            }
            writeError(w, 403, "forbidden", "insufficient permissions")
        })
    }
}
```

### Services

**AuthService**:
- `Register(req RegisterRequest) (User, error)`
- `Login(req LoginRequest) (TokenPair, error)`
- `Refresh(token string) (TokenPair, error)`
- `ApproveUser(adminID, userID int64) error`

**ClassroomService**:
- `CreateClassroom(teacherID int64, req CreateClassroomRequest) (Classroom, error)`
- `RequestEnrollment(studentID, classroomID int64) (Enrollment, error)`
- `ApproveEnrollment(teacherID, enrollmentID int64) error`
- `GetClassroomWithStudents(id int64) (Classroom, []User, error)`

**MaterialService**:
- `CreateMaterial(teacherID int64, req CreateMaterialRequest, files []FileUpload) (Material, error)`
- `GetMaterialWithAttachments(id int64) (Material, []Attachment, error)`
- `DeleteMaterial(teacherID, materialID int64) error`

**AssignmentService**:
- `CreateAssignment(teacherID int64, req CreateAssignmentRequest) (Assignment, error)`
- `SubmitWork(studentID, assignmentID int64, req SubmissionRequest, files []FileUpload) (Submission, error)`
- `GradeSubmission(teacherID, submissionID int64, score int, feedback string) error`

**CollaborativeNoteService**:
- `CreateNote(userID, classroomID int64, req CreateNoteRequest) (CollaborativeNote, error)`
- `UpdateNote(userID, noteID int64, content string, version int) (CollaborativeNote, error)`
- `GetNoteWithContributors(noteID int64) (CollaborativeNote, []User, error)`
- `GetNoteHistory(noteID int64) ([]NoteVersion, error)`

**DiscussionService**:
- `PostMessage(userID, classroomID int64, message string, parentID *int64) (Discussion, error)`
- `GetMessages(classroomID int64, limit, offset int) ([]Discussion, error)`
- `GetThread(parentID int64) ([]Discussion, error)`

---

## Frontend Implementation

### Page Structure

```
src/
├── components/
│   ├── common/
│   │   ├── Button.tsx
│   │   ├── Input.tsx
│   │   ├── Card.tsx
│   │   ├── Modal.tsx
│   │   ├── Loading.tsx
│   │   └── ErrorMessage.tsx
│   ├── auth/
│   │   ├── LoginForm.tsx
│   │   ├── RegisterForm.tsx
│   │   └── ProtectedRoute.tsx
│   ├── layout/
│   │   ├── Navbar.tsx
│   │   ├── Sidebar.tsx
│   │   └── Layout.tsx
│   ├── classroom/
│   │   ├── ClassroomList.tsx
│   │   ├── ClassroomCard.tsx
│   │   ├── ClassroomDetail.tsx
│   │   ├── CreateClassroomModal.tsx
│   │   └── EnrollmentRequests.tsx
│   ├── materials/
│   │   ├── MaterialList.tsx
│   │   ├── MaterialCard.tsx
│   │   ├── MaterialViewer.tsx
│   │   ├── MaterialUpload.tsx
│   │   └── FileUploader.tsx
│   ├── assignments/
│   │   ├── AssignmentList.tsx
│   │   ├── AssignmentCard.tsx
│   │   ├── AssignmentDetail.tsx
│   │   ├── SubmissionForm.tsx
│   │   ├── SubmissionList.tsx
│   │   └── GradingInterface.tsx
│   ├── collaborative-notes/
│   │   ├── NoteList.tsx
│   │   ├── NoteEditor.tsx
│   │   ├── NoteCreateModal.tsx
│   │   ├── ContributorList.tsx
│   │   ├── CursorOverlay.tsx
│   │   └── VersionHistory.tsx
│   └── discussions/
│       ├── DiscussionPanel.tsx
│       ├── MessageList.tsx
│       ├── MessageInput.tsx
│       ├── ThreadView.tsx
│       └── TypingIndicator.tsx
├── pages/
│   ├── LoginPage.tsx
│   ├── RegisterPage.tsx
│   ├── DashboardPage.tsx
│   ├── ClassroomPage.tsx
│   ├── MaterialPage.tsx
│   ├── AssignmentPage.tsx
│   ├── CollaborativeNotesPage.tsx
│   └── AdminPage.tsx
├── hooks/
│   ├── useAuth.ts
│   ├── useWebSocket.ts
│   ├── useClassroom.ts
│   ├── useMaterials.ts
│   ├── useAssignments.ts
│   ├── useNotes.ts
│   └── useDiscussions.ts
├── context/
│   ├── AuthContext.tsx
│   ├── ClassroomContext.tsx
│   └── WebSocketContext.tsx
├── services/
│   ├── api.ts
│   ├── auth.service.ts
│   ├── classroom.service.ts
│   ├── material.service.ts
│   ├── assignment.service.ts
│   ├── note.service.ts
│   └── discussion.service.ts
├── types/
│   ├── user.ts
│   ├── classroom.ts
│   ├── material.ts
│   ├── assignment.ts
│   ├── note.ts
│   └── discussion.ts
└── utils/
    ├── formatters.ts
    ├── validators.ts
    └── ot.ts (Operational Transformation)
```

### Key Features

**Role-Based Dashboard:**
- Superadmin: User management interface, all classrooms overview
- Teacher: My classrooms, pending enrollments, recent submissions
- Student: My classes, upcoming deadlines, recent discussions

**Classroom Tabs:**
1. **Overview** - Description, teacher info, enrolled students
2. **Materials** - List with search, filter by tags, download
3. **Assignments** - List with status, submission interface
4. **Collaborative Notes** - List with real-time editor
5. **Discussions** - Threaded chat with real-time updates

**Collaborative Note Editor:**
- Rich text editor (contenteditable or library)
- Live cursor positions from other users
- User avatars indicating presence
- Version history sidebar
- Conflict resolution UI

**Real-time Discussion:**
- Message list with auto-scroll
- Threaded replies (click to expand)
- Typing indicators ("User is typing...")
- Online presence indicators
- Emoji reactions

---

## Real-time Features

### WebSocket Architecture

**Hub Structure** (`internal/websocket/hub.go`):
```go
type Hub struct {
    clients    map[*Client]bool
    rooms      map[string]map[*Client]bool // room_id -> clients
    register   chan *Client
    unregister chan *Client
    broadcast  chan Message
}

type Client struct {
    hub      *Hub
    conn     *websocket.Conn
    send     chan Message
    userID   int64
    userName string
    rooms    map[string]bool
}

type Message struct {
    Type     string      `json:"type"`
    Room     string      `json:"room"`
    UserID   int64       `json:"user_id"`
    UserName string      `json:"user_name"`
    Data     interface{} `json:"data"`
}
```

### Message Types

**Collaborative Notes:**
- `cursor_move` - Broadcast cursor position
- `selection_change` - Broadcast text selection
- `content_operation` - Broadcast OT operation
- `user_joined` / `user_left` - Presence updates
- `sync_request` - Request full document state
- `sync_response` - Send full document state

**Discussions:**
- `chat_message` - New message posted
- `typing_start` / `typing_stop` - Typing indicators
- `message_reaction` - Emoji reaction added
- `user_online` / `user_offline` - Presence

### Operational Transformation (OT)

**Algorithm** (`internal/ot/transform.go`):
```go
// Operation types
type Operation struct {
    Type string // "retain", "insert", "delete"
    N    int    // length for retain/delete
    Text string // text for insert
}

// Transform two concurrent operations
func Transform(op1, op2 Operation) (Operation, Operation) {
    // Implementation of OT algorithm
    // Handles concurrent edits without conflicts
}

// Compose multiple operations
func Compose(ops []Operation) Operation {
    // Merge operations for efficiency
}
```

**Client-Side OT** (`frontend/src/utils/ot.ts`):
```typescript
interface Operation {
    type: 'retain' | 'insert' | 'delete';
    n?: number;
    text?: string;
}

class OTEngine {
    private revision: number = 0;
    private pendingOps: Operation[] = [];
    
    applyLocalOp(op: Operation): void {
        // Apply to local document
        // Send to server
        this.pendingOps.push(op);
    }
    
    receiveRemoteOp(op: Operation, revision: number): void {
        // Transform against pending ops
        // Apply to document
    }
}
```

---

## Sequence Diagrams

### Diagram 1: Teacher Uploads Material → Student Accesses

```
┌─────────┐          ┌──────────┐          ┌─────────┐          ┌──────────┐          ┌─────────┐
│ Student │          │  System  │          │ Teacher │          │ Database │          │  File   │
└────┬────┘          └────┬─────┘          └────┬────┘          └────┬─────┘          └────┬────┘
     │                    │                     │                    │                     │
     │ 1. Enroll Request  │                     │                    │                     │
     │───────────────────>│                     │                    │                     │
     │                    │ 2. Create Enrollment│                    │                     │
     │                    │────────────────────>│                    │                     │
     │                    │                     │ 3. Store (pending) │                     │
     │                    │                     │───────────────────>│                     │
     │                    │                     │<───────────────────│                     │
     │                    │<────────────────────│                    │                     │
     │                    │ 4. Notify Teacher   │                    │                     │
     │                    │────────────────────>│                    │                     │
     │                    │                     │                    │                     │
     │                    │<────────────────────│ 5. Review Request  │                     │
     │                    │                     │                    │                     │
     │                    │ 6. Approve Student  │                    │                     │
     │                    │────────────────────>│                    │                     │
     │                    │                     │ 7. Update Status   │                     │
     │                    │                     │───────────────────>│                     │
     │                    │                     │<───────────────────│                     │
     │<───────────────────│                     │                    │                     │
     │ 8. Enrolled        │                     │                    │                     │
     │                    │                     │                    │                     │
     │                    │                     │ 9. Upload Material │                     │
     │                    │<────────────────────│                    │                     │
     │                    │ 10. Save Files      │                    │                     │
     │                    │─────────────────────────────────────────────────────────────────>│
     │                    │<─────────────────────────────────────────────────────────────────│
     │                    │                     │                    │                     │
     │                    │ 11. Store Material  │                    │                     │
     │                    │────────────────────>│                    │                     │
     │                    │                     │ 12. Save to DB     │                     │
     │                    │                     │───────────────────>│                     │
     │                    │                     │<───────────────────│                     │
     │                    │<────────────────────│                    │                     │
     │ 13. Notify Students│                     │                    │                     │
     │<───────────────────│                     │                    │                     │
     │                    │                     │                    │                     │
     │ 14. View Material  │                     │                    │                     │
     │───────────────────>│                     │                    │                     │
     │                    │ 15. Fetch Material  │                    │                     │
     │                    │────────────────────>│                    │                     │
     │                    │                     │ 16. Query DB       │                     │
     │                    │                     │───────────────────>│                     │
     │                    │                     │<───────────────────│                     │
     │                    │<────────────────────│                    │                     │
     │<───────────────────│                     │                    │                     │
     │ 17. Content + Files│                     │                    │                     │
     │                    │                     │                    │                     │
     │ 18. Download File  │                     │                    │                     │
     │───────────────────>│                     │                    │                     │
     │                    │ 19. Serve File      │                    │                     │
     │                    │─────────────────────────────────────────────────────────────────>│
     │<─────────────────────────────────────────────────────────────────────────────────────│
     │ 20. File Data      │                     │                    │                     │
     │                    │                     │                    │                     │
```

**Actors:**
- Student (Consumer)
- Teacher (Producer)
- System (API Layer)
- Database (SQLite)
- File Storage

**Interactions:**
1. Student requests enrollment
2. System creates pending enrollment
3. Teacher reviews and approves
4. Student gains access
5. Teacher uploads material with files
6. System stores files and metadata
7. Students are notified
8. Student views and downloads material

---

### Diagram 2: Assignment Lifecycle

```
┌─────────┐          ┌──────────┐          ┌─────────┐          ┌──────────┐          ┌──────────┐
│ Student │          │  System  │          │ Teacher │          │ Database │          │  File    │
└────┬────┘          └────┬─────┘          └────┬────┘          └────┬─────┘          └────┬─────┘
     │                    │                     │                    │                     │
     │                    │                     │ 1. Create Assignment│                    │
     │                    │<────────────────────│                    │                     │
     │                    │ 2. Store Assignment │                    │                     │
     │                    │────────────────────>│                    │                     │
     │                    │                     │ 3. Save to DB      │                     │
     │                    │                     │───────────────────>│                     │
     │                    │                     │<───────────────────│                     │
     │                    │<────────────────────│                    │                     │
     │ 4. Notification    │                     │                    │                     │
     │<───────────────────│                     │                    │                     │
     │                    │                     │                    │                     │
     │ 5. View Assignment │                     │                    │                     │
     │───────────────────>│                     │                    │                     │
     │                    │ 6. Fetch Details    │                    │                     │
     │                    │────────────────────>│                    │                     │
     │                    │                     │ 7. Query DB        │                     │
     │                    │                     │───────────────────>│                     │
     │                    │                     │<───────────────────│                     │
     │                    │<────────────────────│                    │                     │
     │<───────────────────│                     │                    │                     │
     │ 8. Assignment Data │                     │                    │                     │
     │                    │                     │                    │                     │
     │ 9. Submit Work     │                     │                    │                     │
     │───────────────────>│                     │                    │                     │
     │                    │ 10. Validate & Store│                    │                     │
     │                    │─────────────────────────────────────────────────────────────────>│
     │                    │<─────────────────────────────────────────────────────────────────│
     │                    │ 11. Save Submission │                    │                     │
     │                    │────────────────────>│                    │                     │
     │                    │                     │ 12. Store in DB    │                     │
     │                    │                     │───────────────────>│                     │
     │                    │                     │<───────────────────│                     │
     │                    │<────────────────────│                    │                     │
     │                    │ 13. Notify Teacher  │                    │                     │
     │                    │────────────────────>│                    │                     │
     │<───────────────────│                     │                    │                     │
     │ 14. Confirmation   │                     │                    │                     │
     │                    │                     │                    │                     │
     │                    │                     │ 15. View Submission │                    │
     │                    │<────────────────────│                    │                     │
     │                    │ 16. Fetch Submission│                    │                     │
     │                    │────────────────────>│                    │                     │
     │                    │                     │ 17. Query DB       │                     │
     │                    │                     │───────────────────>│                     │
     │                    │                     │<───────────────────│                     │
     │                    │<────────────────────│                    │                     │
     │                    │                     │ 18. Grade & Feedback│                    │
     │                    │<────────────────────│                    │                     │
     │                    │ 19. Store Grade     │                    │                     │
     │                    │────────────────────>│                    │                     │
     │                    │                     │ 20. Update DB      │                     │
     │                    │                     │───────────────────>│                     │
     │                    │                     │<───────────────────│                     │
     │                    │<────────────────────│                    │                     │
     │ 21. Notify Student │                     │                    │                     │
     │<───────────────────│                     │                    │                     │
     │                    │                     │                    │                     │
     │ 22. View Grade     │                     │                    │                     │
     │───────────────────>│                     │                    │                     │
     │<───────────────────│                     │                    │                     │
     │ 23. Score & Feedback│                    │                    │                     │
     │                    │                     │                    │                     │
```

**Actors:**
- Student (Consumer)
- Teacher (Producer)
- System (API Layer)
- Database (SQLite)
- File Storage (for submissions)

**Interactions:**
1. Teacher creates assignment with deadline
2. System stores and notifies students
3. Student views assignment details
4. Student submits work (with optional files)
5. System validates (before deadline) and stores
6. Teacher is notified
7. Teacher reviews and grades
8. Student receives notification
9. Student views grade and feedback

**Validations:**
- Check deadline before accepting submission
- Verify student is enrolled in classroom
- Check assignment ownership before grading

---

## System Flowcharts

### Flowchart 1: User Registration & Approval

```
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                          USER REGISTRATION & APPROVAL                                │
└─────────────────────────────────────────────────────────────────────────────────────┘

    ┌──────────┐
    │  START   │
    └────┬─────┘
         │
         ▼
    ┌──────────────────┐
    │ Fill Registration │
    │  Form with Role   │
    └────┬─────────────┘
         │
         ▼
    ┌──────────────────┐
    │ Validate Input   │
    │ - Email format   │
    │ - Password >= 8  │
    │ - Name not empty │
    │ - Valid role     │
    └────┬─────────────┘
         │
         ▼
    ┌──────────────────┐     ┌──────────────────┐
    │   Valid?         │ NO  │ Show Validation  │
    │                  │────>│     Errors       │
    └────┬─────────────┘     └────────┬─────────┘
         │ YES                        │
         ▼                            │
    ┌──────────────────┐              │
    │ Check Email      │              │
    │   Exists?        │              │
    └────┬─────────────┘              │
         │                            │
         ▼                            │
    ┌──────────────────┐     ┌───────┴──────────┐
    │   Exists?        │ YES │ Show "Email      │
    │                  │────>│ Already Exists"  │
    └────┬─────────────┘     └────────┬─────────┘
         │ NO                           │
         ▼                              │
    ┌──────────────────┐                │
    │ Hash Password    │                │
    │ with bcrypt      │                │
    └────┬─────────────┘                │
         │                              │
         ▼                              │
    ┌──────────────────┐                │
    │ Create User      │                │
    │ Status: PENDING  │                │
    └────┬─────────────┘                │
         │                              │
         ▼                              │
    ┌──────────────────┐                │
    │ Notify Superadmin│                │
    │ (Email/Notif)    │                │
    └────┬─────────────┘                │
         │                              │
         ▼                              │
    ┌──────────────────┐                │
    │ Show "Pending    │◄───────────────┘
    │ Approval" Page   │
    └────┬─────────────┘
         │
         ▼
    ┌──────────────────┐
    │ Superadmin Logs  │
    │       In         │
    └────┬─────────────┘
         │
         ▼
    ┌──────────────────┐
    │ View Pending     │
    │   Users List     │
    └────┬─────────────┘
         │
         ▼
    ┌──────────────────┐
    │ Review User      │
    └────┬─────────────┘
         │
         ▼
    ┌──────────────────┐     ┌──────────────────┐
    │   Approve?       │ NO  │ Suspend User     │
    │                  │────>│ (Rejection)      │
    └────┬─────────────┘     └────┬─────────────┘
         │ YES                     │
         ▼                         │
    ┌──────────────────┐          │
    │ Update Status    │          │
    │   ACTIVE         │          │
    └────┬─────────────┘          │
         │                        │
         ▼                        │
    ┌──────────────────┐          │
    │ Send Welcome     │          │
    │     Email        │          │
    └────┬─────────────┘          │
         │                        │
         ▼                        │
    ┌──────────────────┐          │
    │   END            │◄─────────┘
    └──────────────────┘
```

**Process Description:**
1. User fills registration form selecting role (teacher/student)
2. System validates all inputs
3. If email exists, show error
4. Password is hashed using bcrypt
5. User created with PENDING status
6. Superadmin notified of pending approval
7. User sees "pending approval" message
8. Superadmin reviews and approves/rejects
9. If approved: status set to ACTIVE, welcome email sent
10. If rejected: user suspended

---

### Flowchart 2: Collaborative Note Editing

```
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                      COLLABORATIVE NOTE EDITING                                     │
└─────────────────────────────────────────────────────────────────────────────────────┘

    ┌──────────┐
    │  START   │
    └────┬─────┘
         │
         ▼
    ┌──────────────────┐
    │ User Opens Note  │
    │   in Classroom   │
    └────┬─────────────┘
         │
         ▼
    ┌──────────────────┐     ┌──────────────────┐
    │ Enrolled in      │ NO  │ Show Access      │
    │ Classroom?       │────>│   Denied Error   │
    └────┬─────────────┘     └────────┬─────────┘
         │ YES                        │
         ▼                            │
    ┌──────────────────┐             │
    │ Connect to       │             │
    │ WebSocket Room   │             │
    └────┬─────────────┘             │
         │                           │
         ▼                           │
    ┌──────────────────┐             │
    │ Broadcast        │             │
    │ "User Joined"    │             │
    └────┬─────────────┘             │
         │                           │
         ▼                           │
    ┌──────────────────┐             │
    │ Load Current     │             │
    │ Note Version     │             │
    └────┬─────────────┘             │
         │                           │
         ▼                           │
    ┌──────────────────┐             │
    │ Display Editor   │             │
    │ UI with Cursors  │             │
    └────┬─────────────┘             │
         │                           │
         ▼                           │
    ┌──────────────────┐             │
    │ Wait for User    │             │
    │    Action        │             │
    └────┬─────────────┘             │
         │                           │
         ▼                           │
    ┌──────────────────┐             │
    │ Action Type?     │             │
    └────┬─────────────┘             │
         │                           │
      ┌──┴──┬────────┬────────┐      │
      │     │        │        │      │
      ▼     ▼        ▼        ▼      │
┌─────────┐ ┌────────┐ ┌────────┐ ┌────────┐
│ Cursor  │ │ Edit   │ │ Close  │ │ Other  │
│ Move    │ │ Content│ │  Note  │ │ Action │
└────┬────┘ └────┬───┘ └───┬────┘ └───┬────┘
     │           │         │          │
     ▼           ▼         │          │
┌─────────┐ ┌──────────────────────────┐
│ Broadcast│ │ Create Operation         │
│ Position │ │ (Insert/Delete/Retain)   │
│ to Room  │ └────┬─────────────────────┘
└────┬────┘      │
     │           ▼
     │    ┌──────────────────┐
     │    │ Optimistic Lock  │
     │    │ Check Version    │
     │    └────┬─────────────┘
     │         │
     │         ▼
     │    ┌──────────────────┐     ┌──────────────────┐
     │    │   Conflict?      │ YES │ Show Conflict    │
     │    │                  │────>│ Resolution UI    │
     │    └────┬─────────────┘     └────┬─────────────┘
     │         │ NO                     │
     │         ▼                        │
     │    ┌──────────────────┐          │
     │    │ Broadcast Op to  │          │
     │    │ WebSocket Room   │          │
     │    └────┬─────────────┘          │
     │         │                        │
     │         ▼                        │
     │    ┌──────────────────┐          │
     │    │ Update Database  │          │
     │    │ (content,        │          │
     │    │  contributors)   │          │
     │    └────┬─────────────┘          │
     │         │                        │
     │         ▼                        │
     │    ┌──────────────────┐          │
     │    │ Record Contributor│         │
     │    │  if New          │          │
     │    └────┬─────────────┘          │
     │         │                        │
     │         ▼                        │
     └────►┌──────────────────┐         │
           │ Update UI with   │◄────────┘
           │ Other Users'     │
           │ Changes          │
           └────┬─────────────┘
                │
                ▼
           ┌──────────────────┐
           │ Continue?        │ YES
           │ (Editing)        │────────┐
           └────┬─────────────┘        │
                │ NO                    │
                ▼                       │
           ┌──────────────────┐        │
           │ Save Final       │        │
           │ Version          │        │
           └────┬─────────────┘        │
                │                      │
                ▼                      │
           ┌──────────────────┐        │
           │ Disconnect       │        │
           │ WebSocket        │        │
           └────┬─────────────┘        │
                │                      │
                ▼                      │
           ┌──────────────────┐        │
           │ Broadcast        │        │
           │ "User Left"      │        │
           └────┬─────────────┘        │
                │                      │
                ▼                      │
           ┌──────────────────┐        │
           │      END         │        │
           └──────────────────┘        │
                                       │
                                       ▼
                              ┌──────────────────┐
                              │ Return to        │
                              │ Note List        │
                              └──────────────────┘
```

**Process Description:**
1. User opens a collaborative note
2. System verifies enrollment in classroom
3. WebSocket connection established
4. Other users notified of join
5. Current note content loaded
6. User can:
   - Move cursor (broadcast position)
   - Edit content (OT operation)
   - Close note (disconnect)
7. Edits use Optimistic Locking with version check
8. If conflict: show conflict resolution UI
9. Operations broadcast to all connected users
10. Database updated with new content and contributor record
11. UI updates with real-time changes from others
12. On close: save final version, disconnect, notify others

---

## Function Table

| # | Function | Role | Input | Validation | Output |
|---|----------|------|-------|------------|--------|
| 1 | **Register User** | Any | Email, password, name, role | Email unique, password ≥8 chars, valid role (teacher/student) | Pending user account, superadmin notified |
| 2 | **Login** | Any | Email, password | User exists, status=active, password matches | JWT token pair (access + refresh) |
| 3 | **Approve User** | Superadmin | User ID | User exists, status=pending | Active account, welcome email sent |
| 4 | **Suspend User** | Superadmin | User ID | User exists, not superadmin | Suspended account, sessions invalidated |
| 5 | **Create Classroom** | Superadmin, Teacher | Name, description | Name not empty, user has permission | Classroom created, teacher assigned |
| 6 | **Update Classroom** | Teacher | Classroom ID, name, description | Teacher owns classroom | Updated classroom info |
| 7 | **Delete Classroom** | Teacher | Classroom ID | Teacher owns classroom, confirm delete | Classroom and all data removed |
| 8 | **Request Enrollment** | Student | Classroom ID | Not already enrolled/pending, valid classroom | Pending enrollment request |
| 9 | **Approve Enrollment** | Teacher | Enrollment ID | Teacher owns classroom, status=pending | Student enrolled, access granted |
| 10 | **Remove Student** | Teacher | Classroom ID, Student ID | Teacher owns classroom, student enrolled | Student removed, data preserved |
| 11 | **Upload Material** | Teacher | Title, content, files, classroom ID | Teacher owns classroom, valid files (≤10MB, allowed types) | Material published with attachments |
| 12 | **View Material** | Enrolled User | Material ID | User enrolled in classroom | Material content + file list |
| 13 | **Download File** | Enrolled User | File ID | User has access to parent resource | File stream download |
| 14 | **Delete Material** | Teacher | Material ID | Teacher owns material | Material and attachments removed |
| 15 | **Create Assignment** | Teacher | Title, description, due date, max score | Future due date, teacher owns classroom | Assignment created, students notified |
| 16 | **Update Assignment** | Teacher | Assignment ID, fields | Teacher owns assignment, no submissions yet | Updated assignment |
| 17 | **Delete Assignment** | Teacher | Assignment ID | Teacher owns assignment | Assignment and submissions removed |
| 18 | **View Assignment** | Enrolled User | Assignment ID | User enrolled in classroom | Assignment details + submission status |
| 19 | **Submit Assignment** | Student | Assignment ID, content, files | Before deadline, enrolled student, not submitted | Submission recorded, files stored |
| 20 | **View Submission** | Student | Submission ID | Student owns submission | Submission content + files |
| 21 | **List Submissions** | Teacher | Assignment ID | Teacher owns assignment | List of all student submissions |
| 22 | **Grade Submission** | Teacher | Submission ID, score, feedback | Teacher owns assignment, valid score (0-max_score) | Graded submission, student notified |
| 23 | **Create Collaborative Note** | Enrolled User | Title, content, classroom ID | User enrolled in classroom | Note created, editable by all enrolled |
| 24 | **Update Collaborative Note** | Enrolled User | Note ID, content, version | User enrolled, optimistic lock (version matches) | Updated note, broadcast to room |
| 25 | **Delete Collaborative Note** | Creator | Note ID | User is creator or superadmin | Note removed from system |
| 26 | **View Note History** | Enrolled User | Note ID | User enrolled in classroom | Version history with contributors |
| 27 | **Post Discussion Message** | Enrolled User | Classroom ID, message, parent ID | User enrolled, message not empty, valid parent | Message posted, real-time broadcast |
| 28 | **Get Discussion Messages** | Enrolled User | Classroom ID, limit, offset | User enrolled in classroom | Paginated message list with authors |
| 29 | **Get Thread Replies** | Enrolled User | Parent message ID | User enrolled, parent exists | Threaded replies list |
| 30 | **Get Current User** | Authenticated | JWT token | Valid token, not expired | User profile and permissions |
| 31 | **Refresh Token** | Authenticated | Refresh token | Valid refresh token | New access token |
| 32 | **Logout** | Authenticated | JWT token | Valid token | Token invalidated |

---

## Unique Feature: Collaborative Learning Suite

### Feature Overview

**Name:** SyncSpace Collaborative Learning Suite  
**Components:**
1. Real-time Collaborative Notes
2. Contextual Discussion Forums
3. Contribution Analytics

### Why It's Unique

**Problem with Traditional LMS:**
- Students work in isolation on assignments
- Discussion happens in separate tools (WhatsApp, Discord)
- No visibility into peer learning
- Static content consumption only

**SyncSpace Solution:**
1. **Live Collaboration:** Multiple students edit notes simultaneously with live cursors and conflict-free editing using Operational Transformation
2. **Contextual Discussions:** Discussions tied directly to materials and notes, creating threaded learning conversations
3. **Contribution Tracking:** Every edit attributed, encouraging accountability and collaborative learning
4. **Offline-to-Online Sync:** Changes sync seamlessly when connectivity returns (critical for remote Indonesian areas)

### Data Used

**Real-time Data:**
- User presence (online/offline status)
- Cursor positions and text selections
- Operational Transformation operations
- Edit history with timestamps

**Persistent Data:**
- Collaborative note content versions
- Discussion message threads
- File upload/download analytics
- Contributor frequency per user
- Engagement metrics per classroom

### Benefits

**For Students:**
- Learn from peer perspectives in real-time
- Clarify doubts instantly through contextual discussions
- Build collaborative skills valued in modern workplace
- Study groups can work together remotely

**For Teachers:**
- Monitor student engagement through participation metrics
- Identify struggling students via contribution analytics
- Facilitate peer learning without constant intervention
- Rich feedback on which materials generate most discussion

**For Institution:**
- Higher platform retention due to social learning dynamics
- Differentiation from basic LMS offerings
- Data-driven insights into learning patterns
- Supports hybrid and remote learning models

### Technical Implementation

**Operational Transformation Algorithm:**
- Handles concurrent edits without data loss
- Transforms operations to maintain consistency
- Composable operations for efficiency
- Version vectors for conflict detection

**WebSocket Infrastructure:**
- Sub-100ms latency for real-time feel
- Room-based message routing
- Automatic reconnection with state recovery
- Horizontal scaling ready

**Optimistic Locking:**
- Version-based conflict detection
- Graceful conflict resolution UI
- Automatic retry with exponential backoff
- Last-write-wins fallback for edge cases

---

## Dockerization

### Docker Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           Docker Host                                        │
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                        Docker Compose                                │   │
│  │                                                                     │   │
│  │  ┌──────────────────┐         ┌──────────────────┐                 │   │
│  │  │   Nginx          │         │   Go Backend     │                 │   │
│  │  │   (Frontend)     │────────▶│   (API Server)   │                 │   │
│  │  │                  │         │                  │                 │   │
│  │  │  - Port: 80      │         │  - Port: 8080    │                 │   │
│  │  │  - Static files  │         │  - REST API      │                 │   │
│  │  │  - Reverse proxy │         │  - WebSocket     │                 │   │
│  │  └──────────────────┘         └────────┬─────────┘                 │   │
│  │                                        │                           │   │
│  │                                        │                           │   │
│  │                              ┌─────────▼──────────┐                │   │
│  │                              │   Shared Volumes   │                │   │
│  │                              │                    │                │   │
│  │                              │  ./data:/data      │                │   │
│  │                              │  ./uploads:/uploads│                │   │
│  │                              │                    │                │   │
│  │                              │  SQLite DB         │                │   │
│  │                              │  File uploads      │                │   │
│  │                              └────────────────────┘                │   │
│  │                                                                     │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Dockerfile - Backend

```dockerfile
# backend/Dockerfile
# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o syncspace ./cmd/syncspace

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/syncspace .

# Create directories for data and uploads
RUN mkdir -p /data /uploads

# Expose API port
EXPOSE 8080

# Volume mounts for persistence
VOLUME ["/data", "/uploads"]

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the binary
CMD ["./syncspace"]
```

### Dockerfile - Frontend

```dockerfile
# frontend/Dockerfile
# Build stage
FROM node:20-alpine AS builder

WORKDIR /app

# Copy package files
COPY package*.json ./
RUN npm ci

# Copy source
COPY . .

# Build production bundle
RUN npm run build

# Runtime stage
FROM nginx:alpine

# Copy built assets
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy nginx config
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Expose HTTP port
EXPOSE 80

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost/ || exit 1

# Start nginx
CMD ["nginx", "-g", "daemon off;"]
```

### Nginx Configuration

```nginx
# frontend/nginx.conf
server {
    listen 80;
    server_name localhost;
    root /usr/share/nginx/html;
    index index.html;

    # Frontend routes - serve index.html for SPA
    location / {
        try_files $uri $uri/ /index.html;
    }

    # API proxy to backend
    location /api/ {
        proxy_pass http://backend:8080/api/;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
    }

    # WebSocket proxy
    location /ws/ {
        proxy_pass http://backend:8080/ws/;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_read_timeout 86400;
    }

    # Gzip compression
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_types text/plain text/css application/json application/javascript text/xml;
}
```

### Docker Compose

```yaml
# docker-compose.yml
version: '3.8'

services:
  backend:
    build:
      context: ./backend
      dockerfile: Dockerfile
    container_name: syncspace-backend
    restart: unless-stopped
    ports:
      - "8080:8080"
    volumes:
      - ./data:/data
      - ./uploads:/uploads
    environment:
      - SYNCSPACE_ADDR=:8080
      - SYNCSPACE_DB_PATH=/data/syncspace.db
      - SYNCSPACE_UPLOAD_PATH=/uploads
      - SYNCSPACE_JWT_SECRET=${JWT_SECRET}
      - SYNCSPACE_JWT_EXPIRY=24h
    networks:
      - syncspace-network
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 3s
      retries: 3
      start_period: 10s

  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile
    container_name: syncspace-frontend
    restart: unless-stopped
    ports:
      - "80:80"
    depends_on:
      backend:
        condition: service_healthy
    networks:
      - syncspace-network
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost/"]
      interval: 30s
      timeout: 3s
      retries: 3
      start_period: 10s

networks:
  syncspace-network:
    driver: bridge

volumes:
  data:
    driver: local
  uploads:
    driver: local
```

### Environment Variables

```bash
# .env
JWT_SECRET=your-super-secret-jwt-key-min-32-characters
DB_PATH=./data/syncspace.db
UPLOAD_PATH=./uploads
API_URL=http://localhost:8080
FRONTEND_URL=http://localhost
```

### Docker Benefits

1. **Environment Consistency**
   - Identical behavior across development, staging, and production
   - No "works on my machine" issues
   - Reproducible builds

2. **Isolation**
   - Backend, frontend, and database in separate containers
   - Dependency conflicts eliminated
   - Easy to reset to clean state

3. **Portability**
   - Single `docker-compose up` command deploys entire stack
   - Easy migration between servers
   - Version-controlled infrastructure

4. **Scalability**
   - Horizontal scaling of frontend/backend independently
   - Load balancer ready
   - Microservices migration path

5. **Development Efficiency**
   - New team members up and running in minutes
   - Consistent development environment
   - Easy integration testing

---

## File Storage Strategy

### Hierarchical Directory Structure

```
uploads/
└── classrooms/
    ├── {classroom_id}/
    │   ├── materials/
    │   │   ├── {material_id}/
    │   │   │   ├── lecture1.pdf
    │   │   │   ├── diagram.png
    │   │   │   └── video.mp4
    │   │   └── {material_id}/
    │   │       └── supplementary.docx
    │   └── submissions/
    │       └── {assignment_id}/
    │           ├── {student_id}/
    │           │   └── homework.pdf
    │           ├── {student_id}/
    │           │   └── project.zip
    │           └── {student_id}/
    │               └── essay.docx
    ├── {classroom_id}/
    │   ├── materials/
    │   │   └── {material_id}/
    │   │       └── slides.pptx
    │   └── submissions/
    │       └── {assignment_id}/
    │           └── {student_id}/
    │               └── lab_report.pdf
    └── ...
```

### Storage Configuration

**Path Generation:**
```go
func GenerateFilePath(classroomID, entityID, entityType, userID int64, filename string) string {
    // Format: classrooms/{classroom_id}/{type}/{entity_id}/{user_id}_{timestamp}_{filename}
    timestamp := time.Now().Unix()
    safeFilename := sanitizeFilename(filename)
    
    switch entityType {
    case "material":
        return fmt.Sprintf("classrooms/%d/materials/%d/%s", 
            classroomID, entityID, safeFilename)
    case "submission":
        return fmt.Sprintf("classrooms/%d/submissions/%d/%d/%d_%s", 
            classroomID, entityID, userID, timestamp, safeFilename)
    default:
        return fmt.Sprintf("misc/%d_%s", timestamp, safeFilename)
    }
}
```

**File Metadata Storage:**
```go
type Attachment struct {
    ID           int64     `json:"id"`
    MaterialID   *int64    `json:"material_id,omitempty"`
    SubmissionID *int64    `json:"submission_id,omitempty"`
    Filename     string    `json:"filename"`        // Stored name (UUID)
    OriginalName string    `json:"original_name"`   // User's original name
    MimeType     string    `json:"mime_type"`
    FileSize     int64     `json:"file_size"`
    FilePath     string    `json:"file_path"`       // Relative path from upload root
    UploadedBy   int64     `json:"uploaded_by"`
    CreatedAt    time.Time `json:"created_at"`
}
```

**File Validation:**
```go
var allowedMimeTypes = map[string]bool{
    // Images
    "image/jpeg": true,
    "image/png":  true,
    "image/gif":  true,
    "image/webp": true,
    
    // Documents
    "application/pdf": true,
    "application/msword": true,
    "application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
    "application/vnd.ms-powerpoint": true,
    "application/vnd.openxmlformats-officedocument.presentationml.presentation": true,
    
    // Videos
    "video/mp4":       true,
    "video/webm":      true,
    "video/quicktime": true,
    
    // Archives
    "application/zip": true,
}

const MaxFileSize = 10 * 1024 * 1024 // 10MB
```

**Why Hierarchical is Optimal:**

1. **Classroom-Centric Organization**
   - Education platforms naturally organize around classrooms/courses
   - Easy to find all content for a specific class
   - Logical grouping aligns with user mental models

2. **Lifecycle Management**
   - Delete classroom folder → removes all associated files
   - No orphaned files in system
   - Easy cleanup when courses end

3. **Access Control**
   - Simple validation: `user enrolled in classroom_id?`
   - Directory traversal attacks prevented by validation
   - No direct file access, always through API

4. **Scalability**
   - No single directory with millions of files
   - OS directory listing stays performant
   - Easy to implement storage quotas per classroom

5. **Backup Strategy**
   - Can backup individual classrooms
   - Incremental backups efficient
   - Disaster recovery by classroom

6. **Teacher Management**
   - Teachers can easily find their materials
   - Organized view of student submissions
   - Bulk operations per assignment

---

## Deployment Plan

### Phase 1: Pre-Deployment (Local Testing)

1. **Unit Testing**
   - Run all backend tests: `go test ./...`
   - Verify service layer logic
   - Test database operations

2. **Integration Testing**
   - Test full API workflows
   - Verify authentication flows
   - Test file upload/download
   - Validate WebSocket connections

3. **Build Verification**
   - Build backend binary
   - Build frontend bundle
   - Verify no build errors

### Phase 2: Docker Build

1. **Build Images**
   ```bash
   docker-compose build
   ```

2. **Local Docker Test**
   ```bash
   docker-compose up -d
   ```
   - Verify all services start
   - Test health endpoints
   - Run smoke tests

3. **Push to Registry (Optional)**
   ```bash
   docker tag syncspace-backend your-registry/syncspace-backend:latest
   docker push your-registry/syncspace-backend:latest
   ```

### Phase 3: Production Deployment

1. **Server Preparation**
   - Install Docker and Docker Compose
   - Create directories: `mkdir -p /opt/syncspace/{data,uploads}`
   - Set permissions: `chmod 755 /opt/syncspace`

2. **Deploy Application**
   ```bash
   # Copy docker-compose.yml and .env
   scp docker-compose.yml .env server:/opt/syncspace/
   
   # SSH to server and deploy
   ssh server "cd /opt/syncspace && docker-compose up -d"
   ```

3. **Database Migration**
   - Automatic on first startup
   - Verify tables created
   - Create initial superadmin user

4. **SSL Certificate**
   ```bash
   # Using Let's Encrypt with Certbot
   certbot --nginx -d your-domain.com
   ```

5. **Reverse Proxy (Nginx)**
   ```nginx
   server {
       listen 443 ssl http2;
       server_name syncspace.duskoide.org;
       
       ssl_certificate /path/to/cert.pem;
       ssl_certificate_key /path/to/key.pem;
       
       location / {
           proxy_pass http://localhost:80;
           proxy_set_header Host $host;
           proxy_set_header X-Real-IP $remote_addr;
       }
   }
   ```

6. **Monitoring Setup**
   - Health check endpoint: `GET /health`
   - Log aggregation
   - Disk space monitoring
   - Backup automation

### Phase 4: Post-Deployment

1. **Smoke Tests**
   - Register test users
   - Create test classroom
   - Upload test files
   - Test real-time features

2. **Documentation**
   - Update deployment docs
   - Create user guide
   - Document API endpoints

3. **Backup Strategy**
   ```bash
   # Daily backup script
   #!/bin/bash
   DATE=$(date +%Y%m%d_%H%M%S)
   tar -czf /backups/syncspace_$DATE.tar.gz /opt/syncspace/data /opt/syncspace/uploads
   # Keep only last 7 days
   find /backups -name "syncspace_*.tar.gz" -mtime +7 -delete
   ```

---

## Timeline & Checklist

### Development Timeline

| Day | Focus | Tasks |
|-----|-------|-------|
| **Day 1** | Foundation | Database schema, models, migrations |
| | | User authentication (register, login, JWT) |
| | | Auth middleware, role-based access |
| **Day 2** | Core Features | Classroom CRUD operations |
| | | Enrollment system (request, approve) |
| | | File upload/download system |
| **Day 3** | Producer Flow | Material upload with attachments |
| | | Assignment creation and management |
| | | Teacher grading interface |
| **Day 4** | Consumer Flow | Student submission system |
| | | Material viewing and download |
| | | Assignment list and details |
| **Day 5** | Unique Features | WebSocket server setup |
| | | Operational Transformation implementation |
| | | Collaborative note editor |
| **Day 6** | | Real-time discussion system |
| | | Contribution tracking |
| | | Version history |
| **Day 7** | Frontend | Auth pages and protected routes |
| | | Classroom dashboard |
| | | Material and assignment UI |
| **Day 8** | | Collaborative editor UI |
| | | Discussion panel |
| | | File upload components |
| **Day 9** | Dockerization | Dockerfile for backend |
| | | Dockerfile for frontend |
| | | Docker Compose configuration |
| | | Testing Docker setup |
| **Day 10** | Documentation | Create sequence diagrams |
| | | Create flowcharts |
| | | Write function table |
| | | Document unique feature |
| **Day 11** | Testing | Full integration testing |
| | | Bug fixes |
| | | Performance optimization |
| **Day 12** | Deployment | Production deployment |
| | | Domain configuration |
| | | SSL setup |
| | | Final verification |

### Implementation Checklist

#### Backend

- [ ] Database schema and migrations
- [ ] User model and authentication
- [ ] JWT token generation and validation
- [ ] Auth middleware with role checking
- [ ] User registration with approval workflow
- [ ] Classroom CRUD operations
- [ ] Enrollment request and approval
- [ ] File upload handler with validation
- [ ] File download with auth check
- [ ] Material CRUD with attachments
- [ ] Assignment CRUD with deadlines
- [ ] Submission system with file uploads
- [ ] Grading system with notifications
- [ ] WebSocket server implementation
- [ ] OT algorithm for collaborative editing
- [ ] Discussion message system
- [ ] Threaded replies
- [ ] Contribution tracking
- [ ] Version history

#### Frontend

- [ ] Login page with form validation
- [ ] Registration page with role selection
- [ ] Protected route wrapper
- [ ] Role-based navigation
- [ ] Dashboard for each role
- [ ] Classroom list and creation
- [ ] Classroom detail with tabs
- [ ] Material upload interface
- [ ] Material viewer with downloads
- [ ] Assignment list for students
- [ ] Assignment creation for teachers
- [ ] Submission form with file upload
- [ ] Grading interface for teachers
- [ ] Collaborative note editor
- [ ] Live cursor display
- [ ] Version history sidebar
- [ ] Discussion panel
- [ ] Message threading
- [ ] Typing indicators
- [ ] Online presence indicators

#### Infrastructure

- [ ] Dockerfile for backend
- [ ] Dockerfile for frontend
- [ ] Docker Compose configuration
- [ ] Nginx configuration
- [ ] Environment variable setup
- [ ] Health check endpoints
- [ ] Backup strategy
- [ ] Production deployment

#### Documentation

- [ ] Sequence Diagram 1: Material Upload Flow
- [ ] Sequence Diagram 2: Assignment Lifecycle
- [ ] Flowchart 1: Registration & Approval
- [ ] Flowchart 2: Collaborative Editing
- [ ] Function table (all 32 functions)
- [ ] Unique feature explanation
- [ ] Architecture diagram
- [ ] Docker architecture diagram
- [ ] References and citations

---

## References

### Architecture & Design Patterns

1. **Martin Fowler, "Patterns of Enterprise Application Architecture" (2002)**
   - Layered Architecture pattern for separation of concerns
   - Repository pattern for data access
   - Service layer for business logic

2. **Robert C. Martin, "Clean Architecture" (2017)**
   - Dependency rule and layer independence
   - Use cases and entities separation

### Authentication & Security

3. **RFC 7519 - JSON Web Token (JWT), IETF (2015)**
   - Stateless authentication mechanism
   - Token structure and validation
   - https://tools.ietf.org/html/rfc7519

4. **OWASP Authentication Cheat Sheet**
   - Password storage best practices (bcrypt)
   - Session management guidelines
   - https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html

### Real-time Communication

5. **RFC 6455 - The WebSocket Protocol, IETF (2011)**
   - Full-duplex communication over single TCP connection
   - Frame structure and handshake
   - https://tools.ietf.org/html/rfc6455

6. **Ellis, C. A., and S. J. Gibbs. "Concurrency Control in Groupware Systems." ACM SIGMOD Record (1989)**
   - Operational Transformation algorithm
   - Concurrent editing without conflicts
   - Foundation for collaborative editors

### Database

7. **SQLite Documentation, "Write-Ahead Logging"**
   - WAL mode for concurrent read/write
   - Performance characteristics
   - https://sqlite.org/wal.html

8. **"SQLite vs MySQL vs PostgreSQL: A Comparison of Relational Database Management Systems" - DigitalOcean**
   - Self-managed database rationale
   - SQLite suitability for this use case

### Web Development

9. **React Documentation, "Thinking in React"**
   - Component-based architecture
   - State management patterns
   - https://react.dev/learn/thinking-in-react

10. **Go Documentation, "Standard Library"**
    - `net/http` for REST API
    - `database/sql` for database operations
    - https://pkg.go.dev/std

### File Storage

11. **Filesystem Hierarchy Standard (FHS)**
    - Directory organization best practices
    - Path structure conventions

### Docker & Deployment

12. **Docker Documentation, "Best practices for writing Dockerfiles"**
    - Multi-stage builds
    - Image optimization
    - https://docs.docker.com/develop/dev-best-practices/

13. **Docker Compose Documentation**
    - Service orchestration
    - Volume management
    - https://docs.docker.com/compose/

### Educational Technology

14. **"Collaborative Learning: Higher Education, Interdependence, and the Authority of Knowledge" - Kenneth Bruffee (1993)**
    - Pedagogical basis for collaborative features
    - Peer learning benefits

15. **"The Role of Real-time Collaboration in Online Learning" - Journal of Asynchronous Learning Networks**
    - Impact of synchronous vs asynchronous learning
    - Student engagement metrics

---

## Appendices

### Appendix A: API Response Formats

**Success Response:**
```json
{
  "data": {
    // Response object
  }
}
```

**Error Response:**
```json
{
  "error": {
    "code": "validation_error",
    "message": "Title is required",
    "details": {
      "field": "title",
      "issue": "empty"
    }
  }
}
```

**Standard Error Codes:**
- `bad_request` - Invalid input
- `unauthorized` - Authentication required
- `forbidden` - Insufficient permissions
- `not_found` - Resource doesn't exist
- `validation_error` - Business logic violation
- `conflict` - Resource conflict (e.g., duplicate email)
- `internal_error` - Server error
- `upstream_error` - External API failure

### Appendix B: Environment Variables

**Backend:**
```bash
SYNCSPACE_ADDR=:8080                    # Server listen address
SYNCSPACE_DB_PATH=/data/syncspace.db    # Database file path
SYNCSPACE_UPLOAD_PATH=/uploads          # File upload directory
SYNCSPACE_JWT_SECRET=secret             # JWT signing key
SYNCSPACE_JWT_EXPIRY=24h                # Token expiration
SYNCSPACE_ENV=production                # Environment mode
```

**Frontend:**
```bash
VITE_API_BASE_URL=https://api.syncspace.duskoide.org  # Backend URL
VITE_WS_URL=wss://api.syncspace.duskoide.org          # WebSocket URL
VITE_ENV=production                                   # Environment mode
```

### Appendix C: Testing Commands

**Backend Tests:**
```bash
cd backend
go test ./... -v                    # Run all tests
go test ./... -cover                # With coverage
go test ./internal/service -run TestAuth  # Specific test
```

**Frontend Tests:**
```bash
cd frontend
npm test                            # Run tests
npm run build                       # Production build
npm run lint                        # Lint check
```

**Integration Tests:**
```bash
cd backend
./scripts/integration-test.sh       # Full API tests
```

---

**Document Version:** 1.0  
**Last Updated:** 2026-05-24  
**Prepared for:** II2210 - Teknologi Platform  
**Platform:** SyncSpace Edu - Collaborative Learning Platform

---

*This implementation plan provides a comprehensive roadmap for transforming SyncSpace from a simple task/note manager into a full-featured collaborative learning platform suitable for the Tugas 2 assignment requirements.*
