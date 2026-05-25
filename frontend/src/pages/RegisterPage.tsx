import { useState } from "react";
import { Link } from "react-router-dom";
import { api } from "../services/api";

export function RegisterPage() {
  const [form, setForm] = useState({ email: "", password: "", name: "", role: "collaborator" });
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
      <div className="authShell page">
        <div className="card authCard focusCard">
          <p className="eyebrow authTitle">Account Created</p>
          <h2 className="authHeading" style={{ color: "#ccefdc" }}>Registration Successful</h2>
          <p className="authCopy">Your account is pending approval from the administrator.</p>
          <p className="authFooter">
            <Link to="/login" className="textLink">Go to Login</Link>
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="authShell page">
      <div className="card authCard focusCard">
        <p className="eyebrow authTitle">SyncSpace</p>
        <h1 className="authHeading">Register</h1>
        <p className="authCopy">Create a collaborator or moderator account. New accounts require approval before access.</p>
        {error && (
          <div className="banner error" style={{ marginTop: 20, marginBottom: 0 }}>
            {error}
          </div>
        )}
        <form onSubmit={handleSubmit} className="authForm">
          <div className="field">
            <label>Full Name</label>
            <input
              value={form.name}
              onChange={(e) => setForm({ ...form, name: e.target.value })}
              required
            />
          </div>
          <div className="field">
            <label>Email</label>
            <input
              type="email"
              value={form.email}
              onChange={(e) => setForm({ ...form, email: e.target.value })}
              required
            />
          </div>
          <div className="field">
            <label>Password</label>
            <input
              type="password"
              value={form.password}
              onChange={(e) => setForm({ ...form, password: e.target.value })}
              required
              minLength={8}
            />
          </div>
          <div className="field">
            <label>Role</label>
            <select
              value={form.role}
              onChange={(e) => setForm({ ...form, role: e.target.value })}
            >
              <option value="collaborator">Collaborator</option>
              <option value="moderator">Moderator</option>
            </select>
          </div>
          <button type="submit">Register</button>
        </form>
        <p className="authFooter">
          Already have an account? <Link to="/login" className="textLink">Sign In</Link>
        </p>
      </div>
    </div>
  );
}
