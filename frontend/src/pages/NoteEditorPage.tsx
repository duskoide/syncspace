import { useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { api } from "../services/api";
import { TipTapEditor } from "../components/TipTapEditor";
import { WikipediaSidebar } from "../components/WikipediaSidebar";
import { BackButton } from "../components/BackButton";
import { useAuth } from "../context/AuthContext";

interface Note {
  id: number;
  workspace_id: number;
  title: string;
  content: string;
  created_at: string;
  updated_at: string;
}

interface Workspace {
  id: number;
  name: string;
}

export function NoteEditorPage() {
  const { workspaceId, noteId } = useParams<{ workspaceId: string; noteId: string }>();
  const navigate = useNavigate();
  const { user } = useAuth();
  const [note, setNote] = useState<Note | null>(null);
  const [workspace, setWorkspace] = useState<Workspace | null>(null);
  const [title, setTitle] = useState("");
  const [content, setContent] = useState("");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [showSidebar, setShowSidebar] = useState(false);

  const fetchData = async () => {
    try {
      setLoading(true);
      const wsId = parseInt(workspaceId!);
      
      // Fetch workspace
      const wsData = await api.getWorkspace(wsId);
      setWorkspace(wsData);

      if (noteId) {
        // Fetch existing note
        const noteData = await api.getNote(parseInt(noteId));
        setNote(noteData);
        setTitle(noteData.title);
        setContent(noteData.content);
      }
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, [workspaceId, noteId]);

  const handleSave = async () => {
    if (!title.trim()) {
      setError("Title is required");
      return;
    }

    try {
      setSaving(true);
      const wsId = parseInt(workspaceId!);

      if (noteId) {
        // Update existing note
        await api.updateNote(parseInt(noteId), { title, content });
      } else {
        // Create new note
        const newNote = await api.createNote(wsId, { title });
        // Update with content
        await api.updateNote(newNote.id, { title, content });
        navigate(`/workspaces/${workspaceId}/notes/${newNote.id}`);
      }
      setError("");
    } catch (err: any) {
      setError(err.message);
    } finally {
      setSaving(false);
    }
  };

  const handleInsertFromWiki = (html: string) => {
    setContent((prev) => prev + html);
  };

  if (loading) return <div className="page">Loading...</div>;
  if (!workspace) return <div className="page">Workspace not found</div>;

  return (
    <div className="page">
      <BackButton fallback={`/workspaces/${workspaceId}`} />
      <div style={{ display: "flex", gap: 24, alignItems: "flex-start" }}>
        {/* Main Editor Area */}
        <div style={{ flex: 1, minWidth: 0, maxWidth: 800 }}>
          <div className="surfaceBlock">
            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 16 }}>
              <div>
                <p className="eyebrow">{workspace.name}</p>
                <h1 style={{ margin: 0 }}>{noteId ? "Edit Note" : "New Note"}</h1>
              </div>
              <div style={{ display: "flex", gap: 8 }}>
                <button
                  type="button"
                  className="ghost"
                  onClick={() => setShowSidebar(!showSidebar)}
                >
                  Wikipedia Summary
                </button>
                <button
                  type="button"
                  onClick={handleSave}
                  disabled={saving}
                  className={saving ? "" : "active"}
                >
                  {saving ? "Saving..." : "Save"}
                </button>
              </div>
            </div>

            {error && <div className="banner error" style={{ marginBottom: 16 }}>{error}</div>}

            <div className="field" style={{ marginBottom: 16 }}>
              <label>Title</label>
              <input
                type="text"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                placeholder="Note title"
                style={{ fontSize: "1.25rem", fontWeight: 600 }}
              />
            </div>

            <div className="field">
              <label>Content</label>
              {!noteId && (
                <p className="text-soft" style={{ fontSize: 14, marginBottom: 8 }}>
                  <strong>Tip:</strong> Save the note first, then you can upload images directly into it. For now, paste image URLs or use the Wikipedia sidebar.
                </p>
              )}
              <TipTapEditor
                content={content}
                onChange={setContent}
                placeholder="Start writing your note..."
                noteId={note?.id}
              />
            </div>

            {user?.role === "creator" && note && (
              <div style={{ marginTop: 24, paddingTop: 16, borderTop: "1px solid var(--border)" }}>
                <h4>Share as Template</h4>
                <p className="text-soft" style={{ fontSize: 14, marginBottom: 8 }}>
                  Share this note with the community as a template.
                </p>
                <button
                  type="button"
                  className="ghost"
                  onClick={() => navigate(`/templates/my?type=note&source_id=${note.id}`)}
                >
                  Create Template from Note
                </button>
              </div>
            )}
          </div>
        </div>

        {/* Wikipedia Sidebar */}
        {showSidebar && (
          <div style={{ width: 320, flexShrink: 0 }}>
            <WikipediaSidebar onInsertSummary={handleInsertFromWiki} />
          </div>
        )}
      </div>
    </div>
  );
}
