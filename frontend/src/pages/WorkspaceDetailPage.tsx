import { useEffect, useState } from "react";
import { Link, useParams, useNavigate } from "react-router-dom";
import { api } from "../services/api";
import { useAuth } from "../context/AuthContext";

interface Workspace {
  id: number;
  name: string;
  description: string;
  created_at: string;
  updated_at: string;
}

interface Note {
  id: number;
  title: string;
  content: string;
  creator_name: string;
  created_at: string;
  updated_at: string;
}

export function WorkspaceDetailPage() {
  const { workspaceId } = useParams<{ workspaceId: string }>();
  const navigate = useNavigate();
  const { user } = useAuth();
  const [workspace, setWorkspace] = useState<Workspace | null>(null);
  const [notes, setNotes] = useState<Note[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [showEdit, setShowEdit] = useState(false);
  const [editForm, setEditForm] = useState({ name: "", description: "" });

  const fetchData = async () => {
    try {
      setLoading(true);
      const id = parseInt(workspaceId!);
      const [wsData, notesData] = await Promise.all([
        api.getWorkspace(id),
        api.listNotes(id),
      ]);
      setWorkspace(wsData);
      setNotes(notesData);
      setEditForm({ name: wsData.name, description: wsData.description });
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, [workspaceId]);

  const handleUpdate = async () => {
    if (!editForm.name.trim()) return;
    try {
      await api.updateWorkspace(parseInt(workspaceId!), editForm);
      setShowEdit(false);
      fetchData();
    } catch (err: any) {
      setError(err.message);
    }
  };

  const handleDelete = async () => {
    if (!confirm("Delete this workspace and all its notes?")) return;
    try {
      await api.deleteWorkspace(parseInt(workspaceId!));
      navigate("/workspaces");
    } catch (err: any) {
      setError(err.message);
    }
  };

  if (loading) return <div className="page">Loading workspace...</div>;
  if (!workspace) return <div className="page">Workspace not found</div>;

  return (
    <div className="page">
      <div className="hero">
        <div style={{ flex: 1 }}>
          <p className="eyebrow">Workspace</p>
          {!showEdit ? (
            <>
              <h1 style={{ margin: "0 0 8px 0" }}>{workspace.name}</h1>
              <p className="subheading">{workspace.description || "No description"}</p>
            </>
          ) : (
            <div style={{ maxWidth: 500 }}>
              <div className="field" style={{ marginBottom: 12 }}>
                <input
                  type="text"
                  value={editForm.name}
                  onChange={(e) => setEditForm({ ...editForm, name: e.target.value })}
                  placeholder="Workspace name"
                  style={{ fontSize: "1.5rem", fontWeight: 600 }}
                />
              </div>
              <div className="field">
                <input
                  type="text"
                  value={editForm.description}
                  onChange={(e) => setEditForm({ ...editForm, description: e.target.value })}
                  placeholder="Description"
                />
              </div>
            </div>
          )}
        </div>
        <div className="stats" style={{ marginTop: 0 }}>
          <div className="statCard">
            <span className="statValue">{notes.length}</span>
            <span className="statLabel">Notes</span>
          </div>
        </div>
      </div>

      {error && <div className="banner error" style={{ marginBottom: 16 }}>{error}</div>}

      {/* Action buttons */}
      <div style={{ display: "flex", gap: 12, marginBottom: 24 }}>
        {!showEdit ? (
          <>
            <Link
              to={`/workspaces/${workspace.id}/notes/new`}
              className="button active"
            >
              + New Note
            </Link>
            <button className="ghost" onClick={() => setShowEdit(true)}>
              Edit Workspace
            </button>
            {user?.role === "creator" && (
              <Link
                to={`/templates/my?type=workspace&source_id=${workspace.id}`}
                className="button"
              >
                Share as Template
              </Link>
            )}
            <button className="ghost danger" onClick={handleDelete}>
              Delete
            </button>
          </>
        ) : (
          <>
            <button onClick={handleUpdate}>Save Changes</button>
            <button className="ghost" onClick={() => setShowEdit(false)}>
              Cancel
            </button>
          </>
        )}
      </div>

      {/* Notes list */}
      <h2 style={{ marginBottom: 16 }}>Notes</h2>
      {notes.length === 0 ? (
        <div className="card" style={{ textAlign: "center", padding: 48 }}>
          <p className="text-soft">No notes yet.</p>
          <Link
            to={`/workspaces/${workspace.id}/notes/new`}
            className="button"
            style={{ display: "inline-block", marginTop: 16 }}
          >
            Create your first note
          </Link>
        </div>
      ) : (
        <div className="grid">
          {notes.map((note) => (
            <Link
              key={note.id}
              to={`/workspaces/${workspace.id}/notes/${note.id}`}
              className="card"
              style={{ textDecoration: "none", color: "inherit" }}
            >
              <h3 style={{ marginTop: 0, marginBottom: 8 }}>{note.title}</h3>
              <p
                className="text-soft"
                style={{
                  fontSize: 14,
                  marginBottom: 12,
                  overflow: "hidden",
                  textOverflow: "ellipsis",
                  display: "-webkit-box",
                  WebkitLineClamp: 2,
                  WebkitBoxOrient: "vertical",
                }}
              >
                {note.content.replace(/<[^>]*>/g, "") || "No content"}
              </p>
              <div style={{ display: "flex", justifyContent: "space-between", fontSize: 12 }}>
                <span className="text-soft">By {note.creator_name}</span>
                <span className="text-soft">
                  {new Date(note.updated_at).toLocaleDateString()}
                </span>
              </div>
            </Link>
          ))}

          {/* Add new note card */}
          <Link
            to={`/workspaces/${workspace.id}/notes/new`}
            className="card"
            style={{
              borderStyle: "dashed",
              opacity: 0.7,
              textDecoration: "none",
              color: "inherit",
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              minHeight: 120,
            }}
          >
            <span style={{ fontSize: 24 }}>+ New Note</span>
          </Link>
        </div>
      )}
    </div>
  );
}
