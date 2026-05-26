import { Link } from "react-router-dom";

export function NotFoundPage() {
  return (
    <div className="authShell page">
      <div className="card authCard focusCard" style={{ textAlign: "center" }}>
        <p className="eyebrow authTitle">404</p>
        <h1 className="authHeading">Page Not Found</h1>
        <p className="authCopy">
          The page you're looking for doesn't exist.
        </p>
        <p className="authFooter">
          <Link to="/dashboard" className="textLink">
            Go to Dashboard
          </Link>
        </p>
      </div>
    </div>
  );
}
