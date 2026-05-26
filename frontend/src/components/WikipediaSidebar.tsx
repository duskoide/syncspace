import { useState } from "react";
import { api } from "../services/api";

interface WikipediaSidebarProps {
  onInsertSummary: (summary: string) => void;
}

export function WikipediaSidebar({ onInsertSummary }: WikipediaSidebarProps) {
  const [query, setQuery] = useState("");
  const [result, setResult] = useState<{ topic: string; summary: string } | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const search = async () => {
    if (!query.trim()) return;
    setLoading(true);
    setError("");
    setResult(null);
    try {
      const data = await api.wikiSummary(query);
      setResult(data);
    } catch (err: any) {
      setError(err.message || "Failed to fetch summary");
    } finally {
      setLoading(false);
    }
  };

  const handleInsert = () => {
    if (result) {
      const formatted = `<p><strong>Wikipedia: ${result.topic}</strong></p><p>${result.summary}</p>`;
      onInsertSummary(formatted);
    }
  };

  return (
    <div className="wikiSidebar">
      <h3>Research</h3>
      <div className="wikiSearch">
        <input
          type="text"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="Search Wikipedia..."
          onKeyDown={(e) => e.key === "Enter" && search()}
          className="input"
        />
        <button
          type="button"
          onClick={search}
          disabled={loading || !query.trim()}
          className="button"
        >
          {loading ? "..." : "Search"}
        </button>
      </div>

      {error && <div className="banner error">{error}</div>}

      {result && (
        <div className="wikiResult">
          <h4>{result.topic}</h4>
          <p className="wikiContent">{result.summary}</p>
          <button type="button" onClick={handleInsert} className="button active">
            Insert to Note
          </button>
        </div>
      )}
    </div>
  );
}
