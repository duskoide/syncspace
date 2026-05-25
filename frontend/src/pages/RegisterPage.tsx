import { useState } from "react";
import { Link } from "react-router-dom";
import { api } from "../services/api";

export function RegisterPage() {
  const [form, setForm] = useState({ email: "", password: "", name: "", role: "student" });
  const [error, setError] = useState("");
  const [success, setSuccess] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    try {
      await api.register(form);
      setSuccess(true);
    } catch (err: any) {
      setError(err.message || "Registration failed");
    }
  };

  if (success) {
    return (
      <div style={{ minHeight: "100vh", display: "flex", alignItems: "center", justifyContent: "center", background: "#f3f4f6" }}>
        <div style={{ background: "#fff", padding: 32, borderRadius: 12, width: 360, textAlign: "center" }}>
          <h2 style={{ color: "#059669", marginBottom: 12 }}>Registration Successful!</h2>
          <p style={{ color: "#4b5563", marginBottom: 16 }}>Your account is pending approval from the administrator.</p>
          <Link to="/login" style={{ color: "#2563eb", textDecoration: "none" }}>Go to Login</Link>
        </div>
      </div>
    );
  }

  return (
    <div style={{ minHeight: "100vh", display: "flex", alignItems: "center", justifyContent: "center", background: "#f3f4f6" }}>
      <div style={{ background: "#fff", padding: 32, borderRadius: 12, width: 360, boxShadow: "0 4px 6px rgba(0,0,0,0.1)" }}>
        <h1 style={{ marginBottom: 24, textAlign: "center", color: "#1f2937" }}>SyncSpace Edu</h1>
        <h2 style={{ marginBottom: 16, textAlign: "center", color: "#4b5563", fontSize: 18, fontWeight: 500 }}>Register</h2>
        {error && (
          <div style={{ background: "#fee2e2", color: "#b91c1c", padding: 10, borderRadius: 6, marginBottom: 16, fontSize: 14 }}>
            {error}
          </div>
        )}
        <form onSubmit={handleSubmit}>
          <div style={{ marginBottom: 12 }}>
            <label style={{ display: "block", marginBottom: 4, fontSize: 14, color: "#374151" }}>Full Name</label>
            <input
              value={form.name}
              onChange={(e) => setForm({ ...form, name: e.target.value })}
              required
              style={{ width: "100%", padding: 10, border: "1px solid #d1d5db", borderRadius: 6, fontSize: 14 }}
            />
          </div>
          <div style={{ marginBottom: 12 }}>
            <label style={{ display: "block", marginBottom: 4, fontSize: 14, color: "#374151" }}>Email</label>
            <input
              type="email"
              value={form.email}
              onChange={(e) => setForm({ ...form, email: e.target.value })}
              required
              style={{ width: "100%", padding: 10, border: "1px solid #d1d5db", borderRadius: 6, fontSize: 14 }}
            />
          </div>
          <div style={{ marginBottom: 12 }}>
            <label style={{ display: "block", marginBottom: 4, fontSize: 14, color: "#374151" }}>Password</label>
            <input
              type="password"
              value={form.password}
              onChange={(e) => setForm({ ...form, password: e.target.value })}
              required
              minLength={8}
              style={{ width: "100%", padding: 10, border: "1px solid #d1d5db", borderRadius: 6, fontSize: 14 }}
            />
          </div>
          <div style={{ marginBottom: 16 }}>
            <label style={{ display: "block", marginBottom: 4, fontSize: 14, color: "#374151" }}>Role</label>
            <select
              value={form.role}
              onChange={(e) => setForm({ ...form, role: e.target.value })}
              style={{ width: "100%", padding: 10, border: "1px solid #d1d5db", borderRadius: 6, fontSize: 14 }}
            >
              <option value="student">Student</option>
              <option value="teacher">Teacher</option>
            </select>
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
            Register
          </button>
        </form>
        <p style={{ marginTop: 16, textAlign: "center", fontSize: 14, color: "#6b7280" }}>
          Already have an account? <Link to="/login" style={{ color: "#2563eb" }}>Sign In</Link>
        </p>
      </div>
    </div>
  );
}
