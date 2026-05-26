import { useEffect, useState } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import { api } from "../services/api";
import { BackButton } from "../components/BackButton";
import { useAuth } from "../context/AuthContext";

interface Template {
  id: number;
  type: string;
  name: string;
  description: string;
  visibility: string;
  created_at: string;
  updated_at: string;
}

interface Workspace {
  id: number;
  name: string;
}

interface Note {
  id: number;
  title: string;
  workspace_id: number;
}

export function MyTemplatesPage() {
  const { user } = useAuth();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const [templates, setTemplates] = useState<Template[]>([]);
  const [workspaces, setWorkspaces] = useState<Workspace[]>([]);
  const [notes, setNotes] = useState<Note[]>([]);
  const [loading, setLoading] = useState(true);
  const [showCreate, setShowCreate] = useState(false);
  const [createForm, setCreateForm] = useState({
    type: "note" as "workspace" | "note",
    source_id: 0,
    name: "",
    description: "",
    visibility: "public" as "public" | "link",
  });

  const fetchData = async () => {
    try {
      setLoading(true);
      const [templatesData, workspacesData] = await Promise.all([
        api.listMyTemplates(),
        api.listWorkspaces(),
      ]);
      setTemplates(templatesData);
      setWorkspaces(workspacesData);
      
      // Fetch notes from all workspaces
      const allNotes: Note[] = [];
      for (const ws of workspacesData) {
        const wsNotes = await api.listNotes(ws.id);
        allNotes.push(...wsNotes.map((n: any) => ({ ...n, workspace_id: ws.id })));
      }
      setNotes(allNotes);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (user?.role !== "creator") {
      navigate("/dashboard");
      return;
    }
    fetchData();
    
    // Check for query params to pre-fill create form
    const type = searchParams.get("type") as "workspace" | "note" | null;
    const sourceId = searchParams.get("source_id");
    
    if (type && sourceId) {
      setCreateForm(prev => ({
        ...prev,
        type,
        source_id: parseInt(sourceId),
      }));
      setShowCreate(true);
    }
  }, [user, searchParams]);

  const handleCreate = async () => {
    if (!createForm.name.trim() || !createForm.source_id) {
      alert("Please fill in all fields");
      return;
    }
    try {
      await api.createTemplate(createForm);
      setShowCreate(false);
      setCreateForm({
        type: "note",
        source_id: 0,
        name: "",
        description: "",
        visibility: "public",
      });
      fetchData();
    } catch (err: any) {
      alert(err.message);
    }
  };

  const handleUpdateContent = async (templateId: number) => {
    try {
      await api.updateTemplateContent(templateId);
      alert("Template content updated successfully!");
      fetchData();
    } catch (err: any) {
      alert(err.message);
    }
  };

  if (user?.role !== "creator") {
    return <div className="page">Access denied. Creator role required.</div>;
  }

  return (
    <div className="page">
      <BackButton fallback="/templates" />
      <div className="hero">
        <div>
          <h1 className="heading">My Templates</h1>
          <p className="subheading">
            Manage your shared templates. Update content or change visibility.
          </p>
        </div>
        <button onClick={() => setShowCreate(!showCreate)} className="active">
          {showCreate ? "Cancel" : "+ Create Template"}
        </button>
      </div>

      {showCreate && (
        <div className="card" style={{ marginBottom: 24 }}>
          <h3 style={{ marginTop: 0 }}>Create New Template</h3>
          <div className="grid" style={{ gridTemplateColumns: "1fr 1fr", gap: 16 }}>
            <div className="field">
              <label>Template Type</label>
              <select
                value={createForm.type}
                onChange={(e) => {
                  setCreateForm({
                    ...createForm,
                    type: e.target.value as "workspace" | "note",
                    source_id: 0,
                  });
                }}
              >
                <option value="note">Single Note</option>
                <option value="workspace">Entire Workspace</option>
              </select>
            </div>
            <div className="field">
              <label>Visibility</label>
              <select
                value={createForm.visibility}
                onChange={(e) =>
                  setCreateForm({
                    ...createForm,
                    visibility: e.target.value as "public" | "link",
                  })
                }
              >
                <option value="public">Public (searchable)</option>
                <option value="link">Link-only (unlisted)</option>
              </select>
            </div>
          </div>

          <div className="field">
            <label>
              {createForm.type === "workspace" ? "Select Workspace" : "Select Note"}
            </label>
            <select
              value={createForm.source_id || ""}
              onChange={(e) =>
                setCreateForm({ ...createForm, source_id: parseInt(e.target.value) })
              }
            >
              <option value="">-- Select --</option>
              {createForm.type === "workspace"
                ? workspaces.map((w) => (
                    <option key={w.id} value={w.id}>
                      {w.name}
                    </option>
                  ))
                : notes.map((n) => (
                    <option key={n.id} value={n.id}>
                      {n.title} (from workspace {workspaces.find((w) => w.id === n.workspace_id)?.name})
                    </option>
                  ))}
            </select>
          </div>

          <div className="field">
            <label>Template Name</label>
            <input
              type="text"
              value={createForm.name}
              onChange={(e) => setCreateForm({ ...createForm, name: e.target.value })}
              placeholder="My Awesome Template"
            />
          </div>

          <div className="field">
            <label>Description</label>
            <input
              type="text"
              value={createForm.description}
              onChange={(e) =>
                setCreateForm({ ...createForm, description: e.target.value })
              }
              placeholder="Describe what this template contains..."
            />
          </div>

          <button onClick={handleCreate} className="active">
            Create Template
          </button>
        </div>
      )}

      {loading ? (
        <div>Loading templates...</div>
      ) : templates.length === 0 ? (
        <div className="card" style={{ textAlign: "center", padding: 48 }}>
          <p className="text-soft">You haven't created any templates yet.</p>
          <p style={{ marginTop: 16 }}>
            <button onClick={() => setShowCreate(true)} className="button">
              Create Your First Template
            </button>
          </p>
        </div>
      ) : (
        <div className="grid">
          {templates.map((t) => (
            <div key={t.id} className="card">
              <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start" }}>
                <h3 style={{ marginTop: 0, marginBottom: 8 }}>{t.name}</h3>
                <span className={`tag tag-${t.type === "workspace" ? "info" : "success"}`}>
                  {t.type}
                </span>
              </div>
              <p className="text-soft" style={{ marginBottom: 8 }}>
                {t.description || "No description"}
              </p>
              <div style={{ display: "flex", gap: 8, marginBottom: 12 }}>
                <span className={`tag ${t.visibility === "public" ? "tag-success" : "tag-warning"}`}>
                  {t.visibility}
                </span>
              </div>
              <p className="text-soft" style={{ fontSize: 12, marginBottom: 16 }}>
                Last updated: {new Date(t.updated_at).toLocaleDateString()}
              </p>
              <div style={{ display: "flex", gap: 8 }}>
                <Link to={`/templates/${t.id}`} className="button">
                  View
                </Link>
                <button
                  className="ghost"
                  onClick={() => handleUpdateContent(t.id)}
                  title="Re-snapshot current content"
                >
                  Update Content
                </button>
                <button
                  className="ghost danger"
                  onClick={async () => {
                    if (confirm("Delete this template?")) {
                      await api.deleteTemplate(t.id);
                      fetchData();
                    }
                  }}
                >
                  Delete
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
