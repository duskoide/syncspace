import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useAuth } from "../context/AuthContext";

export function LoginPage() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const { login } = useAuth();
  const navigate = useNavigate();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    try {
      await login(email, password);
      navigate("/dashboard");
    } catch (err: any) {
      setError(err.message || "Login failed");
    }
  };

  return (
    <div style={{ minHeight: "100vh", display: "flex", alignItems: "center", justifyContent: "center", background: "#f3f4f6" }}>
      <div style={{ background: "#fff", padding: 32, borderRadius: 12, width: 360, boxShadow: "0 4px 6px rgba(0,0,0,0.1)" }}>
        <h1 style={{ marginBottom: 24, textAlign: "center", color: "#1f2937" }}>SyncSpace Edu</h1>
        <h2 style={{ marginBottom: 16, textAlign: "center", color: "#4b5563", fontSize: 18, fontWeight: 500 }}>Sign In</h2>
        {error && (
          <div style={{ background: "#fee2e2", color: "#b91c1c", padding: 10, borderRadius: 6, marginBottom: 16, fontSize: 14 }}>
            {error}
          </div>
        )}
        <form onSubmit={handleSubmit}>
          <div style={{ marginBottom: 12 }}>
            <label style={{ display: "block", marginBottom: 4, fontSize: 14, color: "#374151" }}>Email</label>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
              style={{ width: "100%", padding: 10, border: "1px solid #d1d5db", borderRadius: 6, fontSize: 14 }}
            />
          </div>
          <div style={{ marginBottom: 16 }}>
            <label style={{ display: "block", marginBottom: 4, fontSize: 14, color: "#374151" }}>Password</label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              style={{ width: "100%", padding: 10, border: "1px solid #d1d5db", borderRadius: 6, fontSize: 14 }}
            />
          </div>
          <button
            type="submit"
            style={{
              width: "100%",
              padding: 12,
              background: "#2563eb",
              color: "#fff",
              border: "none",
              borderRadius: 6,
              fontSize: 16,
              fontWeight: 600,
              cursor: "pointer",
            }}
          >
            Sign In
          </button>
        </form>
        <p style={{ marginTop: 16, textAlign: "center", fontSize: 14, color: "#6b7280" }}>
          Don't have an account? <Link to="/register" style={{ color: "#2563eb" }}>Register</Link>
        </p>
      </div>
    </div>
  );
}
