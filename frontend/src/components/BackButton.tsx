import { useNavigate } from "react-router-dom";

interface BackButtonProps {
  fallback?: string;
}

export function BackButton({ fallback }: BackButtonProps) {
  const navigate = useNavigate();

  return (
    <button
      type="button"
      className="ghost"
      onClick={() => (fallback ? navigate(fallback) : navigate(-1))}
      style={{ marginBottom: 24 }}
    >
      ← Back
    </button>
  );
}
