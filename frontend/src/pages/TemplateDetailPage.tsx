import { useEffect, useState } from "react";
import { useParams, Link, useNavigate } from "react-router-dom";
import { api } from "../services/api";
import { BackButton } from "../components/BackButton";
import { useAuth } from "../context/AuthContext";

interface Template {
  id: number;
  type: string;
  name: string;
  description: string;
  visibility: string;
  creator_id: number;
  creator_name: string;
  content_snapshot: string;
  created_at: string;
}

interface Workspace {
  id: number;
  name: string;
}

export function TemplateDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { user } = useAuth();
  const [template, setTemplate] = useState<Template | null>(null);
  const [workspaces, setWorkspaces] = useState<Workspace[]>([]);
  const [loading, setLoading] = useState(true);
  const [cloning, setCloning] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    const fetchData = async () => {
      try {
        setLoading(true);
        const [tData, wsData] = await Promise.all([
          api.getTemplate(parseInt(id!)),
          api.listWorkspaces(),
        ]);
        setTemplate(tData);
        setWorkspaces(wsData);
      } catch (err: any) {
        setError(err.message);
      } finally {
        setLoading(false);
      }
    };
    fetchData();
  }, [id]);

  const handleClone = async () => {
    if (!template) return;

    try {
      setCloning(true);
      const result = await api.cloneTemplate(template.id);
      navigate(`/workspaces/${result.workspace.id}`);
    } catch (err: any) {
      setError(err.message);
      setCloning(false);
    }
  };

  if (loading) return <div className="page">Loading...</div>;
  if (!template) return <div className="page">Template not found</div>;

  const isOwner = user?.id === template.creator_id;

  return (
    <div className="page">
      <BackButton fallback="/templates" />
      <div className="surfaceBlock">
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", marginBottom: 16 }}>
          <div>
            <p className="eyebrow">Template</p>
            <h1 style={{ margin: 0 }}>{template.name}</h1>
          </div>
          <span className={`tag tag-${template.visibility === "public" ? "success" : "warning"}`}>
            {template.visibility}
          </span>
        </div>

        <p className="text-soft" style={{ fontSize: 16, marginBottom: 16 }}>
          {template.description || "No description provided."}
        </p>

        <div style={{ display: "flex", gap: 16, marginBottom: 24, fontSize: 14 }}>
          <span className="text-soft">Type: <strong>workspace</strong></span>
          <span className="text-soft">Created by: <strong>{template.creator_name}</strong></span>
          <span className="text-soft">
            Created: <strong>{new Date(template.created_at).toLocaleDateString()}</strong>
          </span>
        </div>

        {error && <div className="banner error" style={{ marginBottom: 16 }}>{error}</div>}

        {!isOwner && (
          <div style={{ padding: 24, background: "rgba(255,255,255,0.05)", borderRadius: 12, marginBottom: 24 }}>
            <h3 style={{ marginTop: 0 }}>Use This Template</h3>

            <button
              onClick={handleClone}
              disabled={cloning}
              className="active"
            >
              {cloning ? "Cloning..." : "Clone Workspace"}
            </button>
          </div>
        )}

        {isOwner && (
          <div style={{ display: "flex", gap: 16 }}>
            <Link to="/templates/my" className="button">
              Manage My Templates
            </Link>
            <button
              className="ghost danger"
              onClick={async () => {
                if (confirm("Delete this template?")) {
                  await api.deleteTemplate(template.id);
                  navigate("/templates/my");
                }
              }}
            >
              Delete
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
