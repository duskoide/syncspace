import { Link } from "react-router-dom";
import { useAuth } from "../context/AuthContext";

export function DashboardPage() {
  const { user } = useAuth();

  const cards = [
    { title: "Boards", desc: "View and manage your collaborative boards", link: "/boards", color: "#2563eb" },
    { title: "Whiteboard", desc: "Create and collaborate on visual workspaces", link: "/boards", color: "#059669" },
    { title: "Discussions", desc: "Participate in board discussions", link: "/boards", color: "#7c3aed" },
  ];

  return (
    <div className="page">
      <div className="hero">
        <div>
          <p className="eyebrow">Dashboard</p>
          <h1>Welcome, {user?.name}!</h1>
          <p className="sub">Jump into boards, whiteboards, and discussion threads from one place.</p>
        </div>
        <div className="stats">
          <div className="stat">
            <span className="statValue">{cards.length}</span>
            <span className="statLabel">Workspace areas</span>
          </div>
          <div className="stat">
            <span className="statValue">{user?.role === "superadmin" ? "Admin" : user?.role === "moderator" ? "Moderator" : "Member"}</span>
            <span className="statLabel">Access level</span>
          </div>
          <div className="stat">
            <span className="statValue">Live</span>
            <span className="statLabel">Realtime sync ready</span>
          </div>
        </div>
      </div>

      <div className="grid">
        {cards.map((card) => (
          <Link
            key={card.title}
            to={card.link}
            style={{
              display: "block",
              textDecoration: "none",
            }}
          >
            <div className="card">
              <h3 style={{ color: card.color, marginBottom: 8 }}>{card.title}</h3>
              <p className="muted" style={{ marginBottom: 0 }}>{card.desc}</p>
            </div>
          </Link>
        ))}
      </div>

      {user?.role === "superadmin" && (
        <div className="card lower focusCard">
          <h3 style={{ color: "#ffd2cf", marginBottom: 8 }}>Admin Actions</h3>
          <Link to="/admin" className="textLink">
            Manage pending user approvals
          </Link>
        </div>
      )}
    </div>
  );
}
