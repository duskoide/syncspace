import { Link, useLocation } from "react-router-dom";
import { useAuth } from "../context/AuthContext";

export function Navbar() {
  const { user, logout } = useAuth();
  const location = useLocation();

  if (!user) return null;

  const navLink = (to: string, label: string) => (
    <Link
      to={to}
      style={{
        marginRight: 16,
        textDecoration: "none",
        color: location.pathname === to ? "#2563eb" : "#374151",
        fontWeight: location.pathname === to ? 600 : 400,
      }}
    >
      {label}
    </Link>
  );

  return (
    <nav
      style={{
        background: "#fff",
        borderBottom: "1px solid #e5e7eb",
        padding: "12px 24px",
        display: "flex",
        justifyContent: "space-between",
        alignItems: "center",
      }}
    >
      <div style={{ display: "flex", alignItems: "center" }}>
        <Link to="/dashboard" style={{ textDecoration: "none", fontSize: 20, fontWeight: 700, color: "#2563eb", marginRight: 24 }}>
          SyncSpace Edu
        </Link>
        {navLink("/dashboard", "Dashboard")}
        {navLink("/classrooms", "Classrooms")}
        {user.role === "superadmin" && navLink("/admin", "Admin")}
      </div>
      <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
        <span style={{ color: "#6b7280", fontSize: 14 }}>
          {user.name} ({user.role})
        </span>
        <button
          onClick={logout}
          style={{
            padding: "6px 12px",
            background: "#ef4444",
            color: "#fff",
            border: "none",
            borderRadius: 6,
            cursor: "pointer",
            fontSize: 14,
          }}
        >
          Logout
        </button>
      </div>
    </nav>
  );
}
