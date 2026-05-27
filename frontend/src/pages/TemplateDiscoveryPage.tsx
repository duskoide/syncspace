import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { api } from "../services/api";
import { useAuth } from "../context/AuthContext";

interface Template {
  id: number;
  type: string;
  name: string;
  description: string;
  visibility: string;
  creator_name: string;
  created_at: string;
}

export function TemplateDiscoveryPage() {
  const { user } = useAuth();
  const [templates, setTemplates] = useState<Template[]>([]);
  const [search, setSearch] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const fetchTemplates = async () => {
    try {
      setLoading(true);
      const data = await api.listTemplates(search);
      setTemplates(data);
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchTemplates();
  }, []);

  const handleSearch = () => {
    fetchTemplates();
  };

  return (
    <div className="page">
      <div className="hero">
        <div>
          <h1 className="heading">Template Gallery</h1>
          <p className="subheading">
            Discover and use templates created by the community.
          </p>
        </div>
      </div>

      <div className="surfaceBlock" style={{ marginBottom: 24 }}>
        <div style={{ display: "flex", gap: 16 }}>
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search templates..."
            style={{ flex: 1 }}
            onKeyDown={(e) => e.key === "Enter" && handleSearch()}
          />
          <button onClick={handleSearch}>Search</button>
        </div>
      </div>

      {error && <div className="banner error">{error}</div>}

      {loading ? (
        <div>Loading templates...</div>
      ) : templates.length === 0 ? (
        <div className="card" style={{ textAlign: "center", padding: 48 }}>
          <p className="text-soft">No templates found.</p>
          {user?.role === "creator" && (
            <p style={{ marginTop: 16 }}>
              <Link to="/templates/my" className="button">
                Create Your First Template
              </Link>
            </p>
          )}
        </div>
      ) : (
        <div className="grid">
          {templates.map((t) => (
            <div key={t.id} className="card">
              <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start" }}>
                <h3 style={{ marginTop: 0, marginBottom: 8 }}>{t.name}</h3>
                <span className="tag tag-info">
                  workspace
                </span>
              </div>
              <p className="text-soft" style={{ marginBottom: 16 }}>
                {t.description || "No description"}
              </p>
              <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginTop: 16 }}>
                <span className="text-soft" style={{ fontSize: 12 }}>
                  by {t.creator_name}
                </span>
                <Link to={`/templates/${t.id}`} className="button">
                  View
                </Link>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
