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

export function AdminPage() {
  const [users, setUsers] = useState<User[]>([]);
  const [filter, setFilter] = useState("pending");
  const [loading, setLoading] = useState(true);
  const { user } = useAuth();

  const loadUsers = async () => {
    setLoading(true);
    try {
      const data = await api.listUsers({ status: filter });
      setUsers(data);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadUsers();
  }, [filter]);

  const approve = async (id: number) => {
    try {
      await api.approveUser(id);
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

  const getRoleTagClass = (role: string) => {
    switch (role) {
      case "superadmin":
        return "tag-danger";
      case "moderator":
        return "tag-info";
      case "collaborator":
        return "tag-success";
      default:
        return "tag";
    }
  };

  const getStatusTagClass = (status: string) => {
    switch (status) {
      case "active":
        return "tag-success";
      case "pending":
        return "tag-warning";
      case "suspended":
        return "tag-danger";
      default:
        return "tag";
    }
  };

  return (
    <div className="page">
      <div className="hero">
        <div>
          <p className="eyebrow">Administration</p>
          <h1>User Management</h1>
          <p className="sub">Manage user accounts, approve pending registrations, and control access levels.</p>
        </div>
        <div className="stats">
          <div className="stat">
            <span className="statValue">{users.length}</span>
            <span className="statLabel">
              {filter === "" ? "Total users" : `${filter} users`}
            </span>
          </div>
          <div className="stat">
            <span className="statValue">{user?.role === "superadmin" ? "Full" : "Limited"}</span>
            <span className="statLabel">Admin access</span>
          </div>
        </div>
      </div>

      <div className="card">
        <div className="sectionHead" style={{ marginBottom: 24 }}>
          <h2>Users</h2>
          <div className="tabRow" style={{ border: "none", padding: 0 }}>
            <button
              onClick={() => setFilter("pending")}
              className={`tabButton${filter === "pending" ? " active" : ""}`}
            >
              Pending
            </button>
            <button
              onClick={() => setFilter("active")}
              className={`tabButton${filter === "active" ? " active" : ""}`}
            >
              Active
            </button>
            <button
              onClick={() => setFilter("suspended")}
              className={`tabButton${filter === "suspended" ? " active" : ""}`}
            >
              Suspended
            </button>
            <button
              onClick={() => setFilter("")}
              className={`tabButton${filter === "" ? " active" : ""}`}
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
                        <button onClick={() => approve(u.id)}>
                          Approve
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
    </div>
  );
}
