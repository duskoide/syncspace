import { Link, useLocation } from "react-router-dom";
import { useAuth } from "../context/AuthContext";

export function Navbar() {
  const { user, logout } = useAuth();
  const location = useLocation();

  if (!user) return null;

  const navLink = (to: string, label: string) => (
    <Link
      to={to}
      className={`topbarLink${location.pathname === to ? " active" : ""}`}
    >
      {label}
    </Link>
  );

  return (
    <nav className="topbar">
      <div className="topbarInner">
        <div className="topbarNav">
          <Link to="/dashboard" className="brand">
            <span className="brandTitle">SyncSpace</span>
            <span className="brandSub">Collaborative whiteboard workspace</span>
          </Link>
          <div className="topbarLinks">
            {navLink("/dashboard", "Dashboard")}
            {navLink("/boards", "Boards")}
            {user.role === "superadmin" && navLink("/admin", "Admin")}
          </div>
        </div>
        <div className="topbarMeta">
          <span className="topbarIdentity">
            {user.name} ({user.role})
          </span>
          <button onClick={logout} className="topbarLogout">
            Logout
          </button>
        </div>
      </div>
    </nav>
  );
}
