import { useAuth } from "../context/AuthContext";

export function DashboardPage() {
  const { user } = useAuth();

  const cards = [
    { title: "Classrooms", desc: "View and manage your classrooms", link: "/classrooms", color: "#2563eb" },
    { title: "Materials", desc: "Access learning materials", link: "/classrooms", color: "#059669" },
    { title: "Assignments", desc: "View and submit assignments", link: "/classrooms", color: "#d97706" },
    { title: "Discussions", desc: "Participate in class discussions", link: "/classrooms", color: "#7c3aed" },
  ];

  return (
    <div style={{ padding: 24 }}>
      <h1 style={{ marginBottom: 8, color: "#1f2937" }}>Welcome, {user?.name}!</h1>
      <p style={{ color: "#6b7280", marginBottom: 24 }}>Role: {user?.role}</p>

      <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fill, minmax(260px, 1fr))", gap: 16 }}>
        {cards.map((card) => (
          <a
            key={card.title}
            href={card.link}
            style={{
              display: "block",
              padding: 20,
              background: "#fff",
              borderRadius: 10,
              textDecoration: "none",
              border: "1px solid #e5e7eb",
              boxShadow: "0 1px 3px rgba(0,0,0,0.08)",
            }}
          >
            <h3 style={{ color: card.color, marginBottom: 6 }}>{card.title}</h3>
            <p style={{ color: "#6b7280", fontSize: 14 }}>{card.desc}</p>
          </a>
        ))}
      </div>

      {user?.role === "superadmin" && (
        <div style={{ marginTop: 24, padding: 20, background: "#fff", borderRadius: 10, border: "1px solid #e5e7eb" }}>
          <h3 style={{ color: "#dc2626", marginBottom: 8 }}>Admin Actions</h3>
          <a href="/admin" style={{ color: "#2563eb", textDecoration: "none" }}>
            Manage pending user approvals →
          </a>
        </div>
      )}
    </div>
  );
}
