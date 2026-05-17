import { FormEvent, useEffect, useState } from "react";

type Task = {
  id: number;
  title: string;
  description: string;
  status: string;
};

type Note = {
  id: number;
  title: string;
  content: string;
  tags: string;
};

const API = import.meta.env.VITE_API_BASE_URL || "http://localhost:8080";

async function req<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${API}${path}`, {
    headers: { "Content-Type": "application/json", ...(options?.headers || {}) },
    ...options,
  });
  if (!res.ok) {
    const txt = await res.text();
    throw new Error(txt || `HTTP ${res.status}`);
  }
  if (res.status === 204) return undefined as T;
  return res.json() as Promise<T>;
}

export function App() {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [notes, setNotes] = useState<Note[]>([]);
  const [taskTitle, setTaskTitle] = useState("");
  const [noteTitle, setNoteTitle] = useState("");
  const [noteContent, setNoteContent] = useState("");
  const [topic, setTopic] = useState("");
  const [selectedNote, setSelectedNote] = useState<number | null>(null);
  const [wikiSummary, setWikiSummary] = useState("");
  const [error, setError] = useState("");

  const load = async () => {
    try {
      setError("");
      const [t, n] = await Promise.all([req<Task[]>("/api/tasks"), req<Note[]>("/api/notes")]);
      setTasks(t);
      setNotes(n);
      if (n.length && selectedNote === null) setSelectedNote(n[0].id);
    } catch (e) {
      setError((e as Error).message);
    }
  };

  useEffect(() => {
    void load();
  }, []);

  const addTask = async (e: FormEvent) => {
    e.preventDefault();
    await req<Task>("/api/tasks", { method: "POST", body: JSON.stringify({ title: taskTitle, status: "todo" }) });
    setTaskTitle("");
    void load();
  };

  const addNote = async (e: FormEvent) => {
    e.preventDefault();
    await req<Note>("/api/notes", { method: "POST", body: JSON.stringify({ title: noteTitle, content: noteContent }) });
    setNoteTitle("");
    setNoteContent("");
    void load();
  };

  const completeTask = async (task: Task) => {
    await req<Task>(`/api/tasks/${task.id}`, {
      method: "PUT",
      body: JSON.stringify({ ...task, status: task.status === "done" ? "todo" : "done" }),
    });
    void load();
  };

  const removeTask = async (id: number) => {
    await req<void>(`/api/tasks/${id}`, { method: "DELETE" });
    void load();
  };

  const removeNote = async (id: number) => {
    await req<void>(`/api/notes/${id}`, { method: "DELETE" });
    void load();
  };

  const fetchSummary = async () => {
    const data = await req<{ topic: string; summary: string }>(`/api/wiki/summary?topic=${encodeURIComponent(topic)}`);
    setWikiSummary(data.summary);
  };

  const enrichNote = async () => {
    if (!selectedNote) return;
    await req<Note>(`/api/notes/${selectedNote}/enrich`, { method: "POST", body: JSON.stringify({ topic }) });
    setWikiSummary("");
    void load();
  };

  return (
    <main className="page">
      <h1>SyncSpace Edu</h1>
      <p className="sub">Education & Learning Tools MVP</p>
      {error && <p className="err">{error}</p>}
      <section className="card">
        <h2>Tasks</h2>
        <form onSubmit={addTask} className="row">
          <input value={taskTitle} onChange={(e) => setTaskTitle(e.target.value)} placeholder="Task title" required />
          <button>Add</button>
        </form>
        <ul>
          {tasks.map((t) => (
            <li key={t.id}>
              <span>{t.title} ({t.status})</span>
              <div>
                <button onClick={() => completeTask(t)}>Toggle</button>
                <button onClick={() => removeTask(t.id)}>Delete</button>
              </div>
            </li>
          ))}
        </ul>
      </section>

      <section className="card">
        <h2>Notes</h2>
        <form onSubmit={addNote} className="col">
          <input value={noteTitle} onChange={(e) => setNoteTitle(e.target.value)} placeholder="Note title" required />
          <textarea value={noteContent} onChange={(e) => setNoteContent(e.target.value)} placeholder="Content" />
          <button>Add Note</button>
        </form>
        <ul>
          {notes.map((n) => (
            <li key={n.id}>
              <label>
                <input
                  type="radio"
                  checked={selectedNote === n.id}
                  onChange={() => setSelectedNote(n.id)}
                />
                {n.title}
              </label>
              <button onClick={() => removeNote(n.id)}>Delete</button>
            </li>
          ))}
        </ul>
      </section>

      <section className="card">
        <h2>Wikipedia Enrichment</h2>
        <div className="row">
          <input value={topic} onChange={(e) => setTopic(e.target.value)} placeholder="Topic (e.g. Calculus)" />
          <button onClick={fetchSummary}>Preview Summary</button>
          <button onClick={enrichNote} disabled={!selectedNote}>Enrich Selected Note</button>
        </div>
        {wikiSummary && <p className="summary">{wikiSummary}</p>}
      </section>
    </main>
  );
}
