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
    <div className="authShell page">
      <div className="card authCard focusCard">
        <p className="eyebrow authTitle">SyncSpace</p>
        <h1 className="authHeading">Sign In</h1>
        <p className="authCopy">Enter your account details to continue to your collaborative workspace.</p>
        {error && (
          <div className="banner error" style={{ marginTop: 24, marginBottom: 0 }}>
            {error}
          </div>
        )}
        <form onSubmit={handleSubmit} className="authForm">
          <div className="field">
            <label>Email</label>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
            />
          </div>
          <div className="field">
            <label>Password</label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
            />
          </div>
          <button type="submit">Sign In</button>
        </form>
        <p className="authFooter">
          Don't have an account? <Link to="/register" className="textLink">Register</Link>
        </p>
      </div>
    </div>
  );
}
