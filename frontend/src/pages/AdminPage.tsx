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

  return (
    <div style={{ padding: 24 }}>
      <h1 style={{ marginBottom: 16, color: "#1f2937" }}>User Management</h1>

      <div style={{ marginBottom: 16 }}>
        <button
          onClick={() => setFilter("pending")}
          style={{
            padding: "8px 16px",
            marginRight: 8,
            background: filter === "pending" ? "#2563eb" : "#e5e7eb",
            color: filter === "pending" ? "#fff" : "#374151",
            border: "none",
            borderRadius: 6,
            cursor: "pointer",
          }}
        >
          Pending
        </button>
        <button
          onClick={() => setFilter("active")}
          style={{
            padding: "8px 16px",
            marginRight: 8,
            background: filter === "active" ? "#2563eb" : "#e5e7eb",
            color: filter === "active" ? "#fff" : "#374151",
            border: "none",
            borderRadius: 6,
            cursor: "pointer",
          }}
        >
          Active
        </button>
        <button
          onClick={() => setFilter("suspended")}
          style={{
            padding: "8px 16px",
            marginRight: 8,
            background: filter === "suspended" ? "#2563eb" : "#e5e7eb",
            color: filter === "suspended" ? "#fff" : "#374151",
            border: "none",
            borderRadius: 6,
            cursor: "pointer",
          }}
        >
          Suspended
        </button>
        <button
          onClick={() => setFilter("")}
          style={{
            padding: "8px 16px",
            background: filter === "" ? "#2563eb" : "#e5e7eb",
            color: filter === "" ? "#fff" : "#374151",
            border: "none",
            borderRadius: 6,
            cursor: "pointer",
          }}
        >
          All
        </button>
      </div>

      {loading ? (
        <p>Loading...</p>
      ) : (
        <div style={{ background: "#fff", borderRadius: 10, border: "1px solid #e5e7eb", overflow: "hidden" }}>
          <table style={{ width: "100%", borderCollapse: "collapse" }}>
            <thead>
              <tr style={{ background: "#f9fafb" }}>
                <th style={{ padding: 12, textAlign: "left", fontSize: 14, color: "#4b5563", borderBottom: "1px solid #e5e7eb" }}>Name</th>
                <th style={{ padding: 12, textAlign: "left", fontSize: 14, color: "#4b5563", borderBottom: "1px solid #e5e7eb" }}>Email</th>
                <th style={{ padding: 12, textAlign: "left", fontSize: 14, color: "#4b5563", borderBottom: "1px solid #e5e7eb" }}>Role</th>
                <th style={{ padding: 12, textAlign: "left", fontSize: 14, color: "#4b5563", borderBottom: "1px solid #e5e7eb" }}>Status</th>
                <th style={{ padding: 12, textAlign: "left", fontSize: 14, color: "#4b5563", borderBottom: "1px solid #e5e7eb" }}>Actions</th>
              </tr>
            </thead>
            <tbody>
              {users.map((u) => (
                <tr key={u.id} style={{ borderBottom: "1px solid #e5e7eb" }}>
                  <td style={{ padding: 12, fontSize: 14 }}>{u.name}</td>
                  <td style={{ padding: 12, fontSize: 14, color: "#6b7280" }}>{u.email}</td>
                  <td style={{ padding: 12, fontSize: 14 }}>
                    <span
                      style={{
                        padding: "4px 8px",
                        borderRadius: 12,
                        fontSize: 12,
                        fontWeight: 600,
                        background: u.role === "superadmin" ? "#fee2e2" : u.role === "teacher" ? "#dbeafe" : "#d1fae5",
                        color: u.role === "superadmin" ? "#b91c1c" : u.role === "teacher" ? "#2563eb" : "#059669",
                      }}
                    >
                      {u.role}
                    </span>
                  </td>
                  <td style={{ padding: 12, fontSize: 14 }}>
                    <span
                      style={{
                        padding: "4px 8px",
                        borderRadius: 12,
                        fontSize: 12,
                        fontWeight: 600,
                        background: u.status === "active" ? "#d1fae5" : u.status === "pending" ? "#fef3c7" : "#fee2e2",
                        color: u.status === "active" ? "#059669" : u.status === "pending" ? "#d97706" : "#b91c1c",
                      }}
                    >
                      {u.status}
                    </span>
                  </td>
                  <td style={{ padding: 12 }}>
                    {u.status === "pending" && (
                      <button
                        onClick={() => approve(u.id)}
                        style={{
                          padding: "6px 12px",
                          background: "#059669",
                          color: "#fff",
                          border: "none",
                          borderRadius: 6,
                          cursor: "pointer",
                          fontSize: 13,
                          marginRight: 8,
                        }}
                      >
                        Approve
                      </button>
                    )}
                    {u.status === "active" && u.id !== user?.id && (
                      <button
                        onClick={() => suspend(u.id)}
                        style={{
                          padding: "6px 12px",
                          background: "#dc2626",
                          color: "#fff",
                          border: "none",
                          borderRadius: 6,
                          cursor: "pointer",
                          fontSize: 13,
                        }}
                      >
                        Suspend
                      </button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          {users.length === 0 && (
            <div style={{ padding: 24, textAlign: "center", color: "#6b7280" }}>No users found</div>
          )}
        </div>
      )}
    </div>
  );
}
