# SyncSpace Pivot Plan: Note-Taking App

**Date:** May 26, 2026
**Goal:** Pivot from collaborative whiteboard/forum to Notion-like note-taking app with template sharing

---

## Core Requirements

### 1. Main Functionality
- **Note-taking app** like Notion, but simpler
- Rich text editing with inline images
- Multiple workspaces per user (personal organization)
- Each workspace contains multiple notes

### 2. User Roles

| Role | Permissions |
|------|-------------|
| **superadmin** | User management (suspend/ban), template moderation (hide/unhide creator templates). No approval needed for new registrations. |
| **creator** | All user privileges + create/update/delete templates from their workspaces/notes. Templates can be shared publicly or via link only. |
| **user** | Core functionality: create personal workspaces, self-notetaking, browse public templates, clone templates to their own workspace. |

### 3. Registration Flow
- New users register → status: `"active"` (immediate login, no approval needed)
- Superadmin moderates existing users only (suspend bad actors, manage creators)

### 4. Template System

#### 4.1 Template Creation
- Creators can create templates from:
  - **Workspace template**: snapshot of entire workspace + all its notes
  - **Note template**: snapshot of a single note
- Templates have:
  - `visibility`: `"public"` (searchable in discovery) or `"link"` (only accessible via direct link)
  - `name`, `description`, `creator_id`
  - `content_snapshot`: stored separately from live workspace/note

#### 4.2 Template Updates
- Creator can manually trigger "Update Template" → re-snapshot current state of source workspace/note
- This updates the template's content snapshot

#### 4.3 Template Cloning ("git clone")
- Users browse public templates or access link-only templates via URL
- When cloning:
  - Workspace template → creates new workspace in user's account + copies all notes
  - Note template → creates new note in chosen workspace
- Cloned content is **fully independent** (no link to template after clone)

### 5. Wikipedia Integration
- Sidebar on Note Editor page
- Search any topic → fetch summary from Wikipedia API
- "Insert to Note" button → inserts summary text at cursor position in rich text editor

### 6. Data Models

#### User (updated)
```go
type User struct {
    ID           int64
    Email        string
    Name         string
    Role         string // "superadmin" | "creator" | "user"
    Status       string // "active" | "suspended" (default: active on register)
    PasswordHash string
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

#### Workspace (new - replaces Board)
```go
type Workspace struct {
    ID          int64
    Name        string
    Description string
    UserID      int64 // owner
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

#### Note (new - replaces TextElement/Discussion)
```go
type Note struct {
    ID        int64
    WorkspaceID int64
    Title     string
    Content   string // HTML content from TipTap editor
    CreatedBy int64  // user who created the note
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

#### Template (new)
```go
type Template struct {
    ID              int64
    Type            string // "workspace" | "note"
    SourceID        int64  // workspace_id or note_id
    CreatorID       int64
    Name            string
    Description     string
    Visibility      string // "public" | "link"
    ContentSnapshot string // JSON: for workspace -> []Note; for note -> Note content
    CreatedAt       time.Time
    UpdatedAt       time.Time
    IsHidden        bool   // superadmin can hide templates
}
```

#### NoteImage (replaces Attachment)
```go
type NoteImage struct {
    ID           int64
    NoteID       int64
    Filename     string
    OriginalName string
    MimeType     string
    FileSize     int64
    FilePath     string // json:"-"
    UploadedBy   int64
    CreatedAt    time.Time
}
```

### 7. File Storage
- Images uploaded via `/api/notes/:id/images` endpoint
- Stored in `uploads/user_{userID}/notes/{filename}`
- Images are inline in note content (HTML `<img>` tags referencing the file endpoint)

---

## Backend API Specification

### Auth Routes (unchanged structure, updated roles)
| Method | Route | Description |
|--------|-------|-------------|
| POST | `/api/auth/register` | Register (role: creator or user, status: active) |
| POST | `/api/auth/login` | Login |
| GET | `/api/auth/me` | Get current user |

### Admin Routes (superadmin only)
| Method | Route | Description |
|--------|-------|-------------|
| GET | `/api/admin/users` | List users with filters |
| PUT | `/api/admin/users/:id/suspend` | Suspend user |
| PUT | `/api/admin/users/:id/activate` | Reactivate suspended user |
| GET | `/api/admin/templates` | List all templates for moderation |
| PATCH | `/api/admin/templates/:id/hide` | Hide template from public |
| PATCH | `/api/admin/templates/:id/unhide` | Unhide template |

### Workspace Routes
| Method | Route | Access |
|--------|-------|--------|
| GET | `/api/workspaces` | List my workspaces (user only sees their own) |
| POST | `/api/workspaces` | Create workspace (any authenticated) |
| GET | `/api/workspaces/:id` | Get workspace (owner only) |
| PUT | `/api/workspaces/:id` | Update workspace (owner only) |
| DELETE | `/api/workspaces/:id` | Delete workspace + all notes (owner only) |

### Note Routes
| Method | Route | Access |
|--------|-------|--------|
| GET | `/api/workspaces/:id/notes` | List notes in workspace (owner only) |
| POST | `/api/workspaces/:id/notes` | Create note in workspace (owner only) |
| GET | `/api/notes/:id` | Get note (workspace owner only) |
| PUT | `/api/notes/:id` | Update note (workspace owner only) |
| DELETE | `/api/notes/:id` | Delete note (workspace owner only) |
| POST | `/api/notes/:id/images` | Upload inline image for note (owner only) |
| GET | `/api/files/:id` | Download/view image file |

### Template Routes (Creator only for creation)
| Method | Route | Access |
|--------|-------|--------|
| GET | `/api/templates` | Search public templates (any authenticated) |
| GET | `/api/templates/:id` | Get template details (public/link, check access) |
| POST | `/api/templates` | Create template from workspace/note (creator only) |
| PUT | `/api/templates/:id` | Update template metadata (creator only) |
| POST | `/api/templates/:id/update-content` | Re-snapshot content (creator only) |
| DELETE | `/api/templates/:id` | Delete template (creator or superadmin) |
| POST | `/api/templates/:id/clone` | Clone template to my workspace (any authenticated) |

### Wikipedia Routes
| Method | Route | Description |
|--------|-------|-------------|
| GET | `/api/wiki/summary?topic=` | Get Wikipedia summary (reuse existing) |

---

## Frontend Page Structure

### Public Pages (no auth required)
- `/login` - LoginPage (existing, keep dark theme)
- `/register` - RegisterPage (updated: creator/user roles, remove "pending approval")

### Protected Pages (any authenticated)
- `/dashboard` - DashboardPage (updated: show workspaces, stats, template discovery link)
- `/workspaces` - WorkspaceListPage (list my workspaces, create new)
- `/workspaces/:id` - WorkspaceDetailPage (view workspace + list notes, open editor)
- `/workspaces/:id/notes/:noteId` - NoteEditorPage (TipTap editor + Wikipedia sidebar)
- `/templates` - TemplateDiscoveryPage (browse/search public templates)
- `/templates/:id` - TemplateDetailPage (view template, clone button)

### Admin Pages (superadmin only)
- `/admin` - AdminPage (updated: user management + template moderation tabs)

### Removed Pages
- `/boards` - BoardPage (completely removed)
- All board-related routes

---

## Frontend Component Plan

### New Components Needed
1. **TipTapEditor** - Rich text editor with:
   - Bold, italic, headers, lists
   - Image insertion (upload + inline)
   - Content persistence to/from API

2. **WikipediaSidebar** - Sidebar panel with:
   - Search input
   - Loading state
   - Summary display
   - "Insert to Note" button

3. **TemplateCard** - Card displaying template preview (name, description, creator, visibility badge)

4. **TemplateModal** - Modal for creating template from current workspace/note

### Updated Components
1. **Navbar** - Update links (remove boards, add workspaces, templates)
2. **DashboardPage** - New cards: Workspaces, Templates, Recent Notes

---

## Database Schema Changes

### New Tables
```sql
-- Workspaces (personal containers)
CREATE TABLE workspaces (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT DEFAULT '',
    user_id INTEGER NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Notes (rich text content)
CREATE TABLE notes (
    id INTEGER PRIMARY KEY,
    workspace_id INTEGER NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    content TEXT DEFAULT '', -- HTML content
    created_by INTEGER NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Templates (shared templates)
CREATE TABLE templates (
    id INTEGER PRIMARY KEY,
    type TEXT NOT NULL CHECK(type IN ('workspace', 'note')),
    source_id INTEGER NOT NULL, -- workspace_id or note_id
    creator_id INTEGER NOT NULL REFERENCES users(id),
    name TEXT NOT NULL,
    description TEXT DEFAULT '',
    visibility TEXT NOT NULL CHECK(visibility IN ('public', 'link')) DEFAULT 'public',
    content_snapshot TEXT NOT NULL, -- JSON string
    is_hidden BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Note Images (inline images)
CREATE TABLE note_images (
    id INTEGER PRIMARY KEY,
    note_id INTEGER NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
    filename TEXT NOT NULL,
    original_name TEXT NOT NULL,
    mime_type TEXT NOT NULL,
    file_size INTEGER NOT NULL,
    file_path TEXT NOT NULL,
    uploaded_by INTEGER NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Dropped Tables
- `boards`
- `board_memberships`
- `text_elements`
- `discussions`
- `attachments` (replaced by note_images)

### Updated Tables
- `users`: Change role CHECK constraint, default status to "active"

---

## Implementation Order

### Phase 1: Backend Foundation
1. Update User model (roles: superadmin/creator/user, status default active)
2. Create new models: Workspace, Note, Template, NoteImage
3. Update Store (schema migration, new CRUD methods)
4. Update Services (Workspace, Note, Template, File for images)
5. Create new Handlers (Workspace, Note, Template)
6. Update routes (remove old, add new)
7. Test backend API

### Phase 2: Frontend Foundation
1. Install TipTap dependencies
2. Create TipTapEditor component
3. Create WikipediaSidebar component
4. Update AuthContext (new role names)
5. Update API service (new endpoints)

### Phase 3: Frontend Pages
1. Update RegisterPage (roles: creator/user, remove pending message)
2. Create WorkspaceListPage
3. Create NoteEditorPage (editor + Wikipedia sidebar)
4. Create TemplateDiscoveryPage
5. Create TemplateDetailPage
6. Update DashboardPage
7. Update AdminPage (add template moderation)
8. Remove BoardPage

### Phase 4: Cleanup & Testing
1. Remove websocket infrastructure
2. Remove old unused code
3. Clean up database (delete old .db file)
4. Full integration test
5. Docker rebuild

---

## UI Style Constraints

**DO NOT CHANGE:**
- Color scheme (dark navy/teal background, amber accent #efb449)
- Typography (IBM Plex Sans)
- Glassmorphism effects (backdrop-filter: blur(18px))
- Border radius patterns (14-22px for cards)
- Layout grid patterns
- Button styles
- Input styles

**CAN ADD:**
- New utility classes for TipTap editor styling
- New layout classes for workspace/note organization
- Template card styles (must follow existing card pattern)

---

## Key Design Decisions

1. **Personal Workspaces Only**: No collaboration on live workspaces. Sharing is via templates only.
2. **Snapshot Templates**: Templates are decoupled from live content. Manual update required to sync.
3. **HTML Content Storage**: TipTap outputs HTML. Store as HTML string in database.
4. **Inline Images**: Images are uploaded separately, referenced in HTML via `<img src="/api/files/{id}">`
5. **No Real-time Collaboration**: WebSocket removed. Single-user editing only.
6. **No Approval Workflow**: Registration is immediate (status: active).
7. **Superadmin Template Moderation**: Can hide/unhide creator templates from public discovery.

---

## Notes for Implementation

- Keep existing authentication flow (JWT)
- Keep existing file upload infrastructure (just repurpose for note images)
- Keep existing Wikipedia API integration
- Keep existing superadmin user management (just remove approval requirement)
- Database must be deleted and recreated (no migration system)
- All pages must maintain dark theme consistency
- Test image upload/inline display carefully
- Ensure template cloning creates deep copies (not references)
