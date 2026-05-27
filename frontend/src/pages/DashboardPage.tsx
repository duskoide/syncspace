import { Link } from "react-router-dom";
import { useAuth } from "../context/AuthContext";

export function DashboardPage() {
  const { user } = useAuth();

  const cards = [
    { 
      title: "My Workspaces", 
      desc: "Organize and access your personal notes", 
      link: "/workspaces", 
      color: "#2563eb" 
    },
    { 
      title: "Template Gallery", 
      desc: "Discover templates shared by the community", 
      link: "/templates", 
      color: "#059669" 
    },
    ...(user?.role !== "superadmin" ? [{
      title: "Research", 
      desc: "Search Wikipedia and enrich your notes", 
      link: "/workspaces", 
      color: "#7c3aed" 
    }] : []),
    ...(user?.role === "superadmin" ? [{
      title: "Admin Actions",
      desc: "Manage users and moderate community templates",
      link: "/admin",
      color: "#ffd2cf",
    }] : []),
    ...(user?.role === "creator" ? [{
      title: "Creator Tools",
      desc: "Manage and share your templates with the community",
      link: "/templates/my",
      color: "#efb449",
    }] : []),
  ];

  return (
    <div className="page">
      <div className="hero" style={{ gridTemplateColumns: "1fr" }}>
        <div>
          <p className="eyebrow">Dashboard</p>
          <h1>Welcome, {user?.name}!</h1>
          <p className="sub">
            Your personal note-taking space. Create workspaces, take notes, and discover templates.
          </p>
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
    </div>
  );
}
