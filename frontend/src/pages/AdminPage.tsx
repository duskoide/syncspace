import { useState, useEffect } from "react";
import { useAuth } from "../context/AuthContext";
import { api } from "../services/api";

interface User {
  id: number;
  email: string;
  name: string;
  role: string;
  status: string;
  created_at: string;
}

interface Template {
  id: number;
  type: string;
  name: string;
  visibility: string;
  creator_name: string;
  is_hidden: boolean;
  created_at: string;
}

export function AdminPage() {
  const [users, setUsers] = useState<User[]>([]);
  const [templates, setTemplates] = useState<Template[]>([]);
  const [userFilter, setUserFilter] = useState("active");
  const [activeTab, setActiveTab] = useState<"users" | "templates">("users");
  const [loading, setLoading] = useState(true);
  const { user } = useAuth();

  const loadUsers = async () => {
    try {
      const data = await api.listUsers({ status: userFilter });
      setUsers(data);
    } catch (err) {
      console.error(err);
    }
  };

  const loadTemplates = async () => {
    try {
      const data = await api.listAllTemplates();
      setTemplates(data);
    } catch (err) {
      console.error(err);
    }
  };

  useEffect(() => {
    const loadData = async () => {
      setLoading(true);
      await Promise.all([loadUsers(), loadTemplates()]);
      setLoading(false);
    };
    loadData();
  }, [userFilter]);

  const activate = async (id: number) => {
    try {
      await api.activateUser(id);
      loadUsers();
    } catch (err: any) {
      alert(err.message);
    }
  };

  const suspend = async (id: number) => {
    try {
      await api.suspendUser(id);
      loadUsers();
    } catch (err: any) {
      alert(err.message);
    }
  };

  const toggleTemplateVisibility = async (id: number, currentHidden: boolean) => {
    try {
      await api.setTemplateHidden(id, !currentHidden);
      loadTemplates();
    } catch (err: any) {
      alert(err.message);
    }
  };

  const getRoleTagClass = (role: string) => {
    switch (role) {
      case "superadmin":
        return "tag-danger";
      case "creator":
        return "tag-info";
      case "user":
        return "tag-success";
      default:
        return "tag";
    }
  };

  const getStatusTagClass = (status: string) => {
    switch (status) {
      case "active":
        return "tag-success";
      case "suspended":
        return "tag-danger";
      case "pending":
        return "tag-warning";
      default:
        return "tag";
    }
  };

  return (
    <div className="page">
      <div className="hero">
        <div>
          <p className="eyebrow">Administration</p>
          <h1>Admin Dashboard</h1>
          <p className="sub">Manage users and moderate community templates.</p>
        </div>
        <div className="stats">
          <div className="stat">
            <span className="statValue">{users.length}</span>
            <span className="statLabel">Users</span>
          </div>
          <div className="stat">
            <span className="statValue">{templates.length}</span>
            <span className="statLabel">Templates</span>
          </div>
        </div>
      </div>

      {/* Tab Navigation */}
      <div className="tabRow" style={{ marginBottom: 24 }}>
        <button
          onClick={() => setActiveTab("users")}
          className={`tabButton${activeTab === "users" ? " active" : ""}`}
        >
          Users
        </button>
        <button
          onClick={() => setActiveTab("templates")}
          className={`tabButton${activeTab === "templates" ? " active" : ""}`}
        >
          Templates
        </button>
      </div>

      {activeTab === "users" && (
        <div className="card">
          <div className="sectionHead" style={{ marginBottom: 24 }}>
            <h2>Users</h2>
            <div className="tabRow" style={{ border: "none", padding: 0 }}>
              <button
                onClick={() => setUserFilter("active")}
                className={`tabButton${userFilter === "active" ? " active" : ""}`}
              >
                Active
              </button>
              <button
                onClick={() => setUserFilter("pending")}
                className={`tabButton${userFilter === "pending" ? " active" : ""}`}
              >
                Pending
              </button>
              <button
                onClick={() => setUserFilter("suspended")}
                className={`tabButton${userFilter === "suspended" ? " active" : ""}`}
              >
                Suspended
              </button>
              <button
                onClick={() => setUserFilter("")}
                className={`tabButton${userFilter === "" ? " active" : ""}`}
              >
                All
              </button>
            </div>
          </div>

          {loading ? (
            <div className="emptyState">
              <p>Loading users...</p>
            </div>
          ) : users.length === 0 ? (
            <div className="emptyState">
              <p>No users found</p>
            </div>
          ) : (
            <div className="adminTable">
              <div className="adminTableHeader">
                <div className="adminTableCell">Name</div>
                <div className="adminTableCell">Email</div>
                <div className="adminTableCell">Role</div>
                <div className="adminTableCell">Status</div>
                <div className="adminTableCell">Actions</div>
              </div>
              <div className="adminTableBody">
                {users.map((u) => (
                  <div key={u.id} className="adminTableRow">
                    <div className="adminTableCell">
                      <strong>{u.name}</strong>
                    </div>
                    <div className="adminTableCell metaText">{u.email}</div>
                    <div className="adminTableCell">
                      <span className={`tag ${getRoleTagClass(u.role)}`}>
                        {u.role}
                      </span>
                    </div>
                    <div className="adminTableCell">
                      <span className={`tag ${getStatusTagClass(u.status)}`}>
                        {u.status}
                      </span>
                    </div>
                    <div className="adminTableCell">
                      <div className="actions">
                        {u.status === "pending" && (
                          <button onClick={() => activate(u.id)}>
                            Approve
                          </button>
                        )}
                        {u.status === "suspended" && (
                          <button onClick={() => activate(u.id)}>
                            Activate
                          </button>
                        )}
                        {u.status === "active" && u.id !== user?.id && (
                          <button
                            className="danger"
                            onClick={() => suspend(u.id)}
                          >
                            Suspend
                          </button>
                        )}
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      )}

      {activeTab === "templates" && (
        <div className="card">
          <div className="sectionHead" style={{ marginBottom: 24 }}>
            <h2>Template Moderation</h2>
            <p className="text-soft">Hide or unhide templates from public discovery.</p>
          </div>

          {loading ? (
            <div className="emptyState">
              <p>Loading templates...</p>
            </div>
          ) : templates.length === 0 ? (
            <div className="emptyState">
              <p>No templates found</p>
            </div>
          ) : (
            <div className="adminTable">
              <div className="adminTableHeader">
                <div className="adminTableCell">Name</div>
                <div className="adminTableCell">Type</div>
                <div className="adminTableCell">Creator</div>
                <div className="adminTableCell">Visibility</div>
                <div className="adminTableCell">Actions</div>
              </div>
              <div className="adminTableBody">
                {templates.map((t) => (
                  <div key={t.id} className="adminTableRow">
                    <div className="adminTableCell">
                      <strong>{t.name}</strong>
                    </div>
                    <div className="adminTableCell">
                      <span className={`tag tag-${t.type === "workspace" ? "info" : "success"}`}>
                        {t.type}
                      </span>
                    </div>
                    <div className="adminTableCell metaText">{t.creator_name}</div>
                    <div className="adminTableCell">
                      <span className={`tag ${t.is_hidden ? "tag-danger" : "tag-success"}`}>
                        {t.is_hidden ? "Hidden" : "Visible"}
                      </span>
                    </div>
                    <div className="adminTableCell">
                      <button
                        onClick={() => toggleTemplateVisibility(t.id, t.is_hidden)}
                        className={t.is_hidden ? "" : "danger"}
                      >
                        {t.is_hidden ? "Unhide" : "Hide"}
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
