import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { api } from "../services/api";
import { useAuth } from "../context/AuthContext";

interface Workspace {
  id: number;
  name: string;
  description: string;
  created_at: string;
}

export function WorkspaceListPage() {
  const { user } = useAuth();
  const [workspaces, setWorkspaces] = useState<Workspace[]>([]);
  const [showCreate, setShowCreate] = useState(false);
  const [newWorkspace, setNewWorkspace] = useState({ name: "", description: "" });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const fetchWorkspaces = async () => {
    try {
      setLoading(true);
      const data = await api.listWorkspaces();
      setWorkspaces(data);
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchWorkspaces();
  }, []);

  const handleCreate = async () => {
    if (!newWorkspace.name.trim()) return;
    try {
      await api.createWorkspace(newWorkspace);
      setNewWorkspace({ name: "", description: "" });
      setShowCreate(false);
      fetchWorkspaces();
    } catch (err: any) {
      setError(err.message);
    }
  };

  if (loading) return <div className="page">Loading workspaces...</div>;

  return (
    <div className="page">
      <div className="hero">
        <div>
          <h1 className="heading">Your Workspaces</h1>
          <p className="subheading">
            Organize your notes in workspaces. Create a new workspace to get started.
          </p>
        </div>
        <div className="stats" style={{ marginTop: 0 }}>
          <div className="statCard">
            <span className="statValue">{workspaces.length}</span>
            <span className="statLabel">Workspaces</span>
          </div>
        </div>
      </div>

      {error && <div className="banner error">{error}</div>}

      <div className="grid" style={{ marginTop: 24 }}>
        {workspaces.map((w) => (
          <Link
            key={w.id}
            to={`/workspaces/${w.id}`}
            className="card"
            style={{ textDecoration: "none", color: "inherit" }}
          >
            <h3 style={{ marginTop: 0 }}>{w.name}</h3>
            <p className="text-soft">{w.description || "No description"}</p>
            <p className="text-soft" style={{ fontSize: 12, marginTop: 16 }}>
              Created: {new Date(w.created_at).toLocaleDateString()}
            </p>
          </Link>
        ))}

        <div className="card" style={{ borderStyle: "dashed", opacity: 0.7 }}>
          {!showCreate ? (
            <button
              className="ghost"
              style={{ width: "100%", height: "100%", minHeight: 100 }}
              onClick={() => setShowCreate(true)}
            >
              + New Workspace
            </button>
          ) : (
            <div>
              <h4 style={{ marginTop: 0 }}>Create Workspace</h4>
              <div className="field">
                <label>Name</label>
                <input
                  value={newWorkspace.name}
                  onChange={(e) =>
                    setNewWorkspace({ ...newWorkspace, name: e.target.value })
                  }
                  placeholder="My Notes"
                />
              </div>
              <div className="field">
                <label>Description</label>
                <input
                  value={newWorkspace.description}
                  onChange={(e) =>
                    setNewWorkspace({ ...newWorkspace, description: e.target.value })
                  }
                  placeholder="Optional description"
                />
              </div>
              <div style={{ display: "flex", gap: 8, marginTop: 16 }}>
                <button onClick={handleCreate}>Create</button>
                <button className="ghost" onClick={() => setShowCreate(false)}>
                  Cancel
                </button>
              </div>
            </div>
          )}
        </div>
      </div>

      {user?.role === "creator" && (
        <div className="surfaceBlock" style={{ marginTop: 24 }}>
          <h3>Creator Tools</h3>
          <p>Share your workspaces or notes as templates for others to use.</p>
          <Link to="/templates/my" className="button" style={{ display: "inline-block", marginTop: 8 }}>
            Manage My Templates
          </Link>
        </div>
      )}
    </div>
  );
}
