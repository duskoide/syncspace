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

function getErrorMessage(error: unknown): string {
  if (error instanceof Error) {
    try {
      const parsed = JSON.parse(error.message) as { error?: { message?: string } };
      if (parsed.error?.message) {
        return parsed.error.message;
      }
    } catch {
      return error.message;
    }
    return error.message;
  }
  return "Unexpected error";
}

export function App() {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [notes, setNotes] = useState<Note[]>([]);
  const [taskTitle, setTaskTitle] = useState("");
  const [noteTitle, setNoteTitle] = useState("");
  const [noteContent, setNoteContent] = useState("");
  const [editTitle, setEditTitle] = useState("");
  const [editContent, setEditContent] = useState("");
  const [isEditing, setIsEditing] = useState(false);
  const [topic, setTopic] = useState("");
  const [selectedNote, setSelectedNote] = useState<number | null>(null);
  const [wikiSummary, setWikiSummary] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);
  const [working, setWorking] = useState("");

  const load = async () => {
    try {
      setLoading(true);
      setError("");
      const [t, n] = await Promise.all([req<Task[]>("/api/tasks"), req<Note[]>("/api/notes")]);
      setTasks(t);
      setNotes(n);
      setSelectedNote((current) => {
        if (!n.length) {
          setIsEditing(false);
          return null;
        }
        if (current !== null && n.some((note) => note.id === current)) {
          return current;
        }
        setIsEditing(false);
        return n[0].id;
      });
    } catch (e) {
      setError(getErrorMessage(e));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void load();
  }, []);

  const addTask = async (e: FormEvent) => {
    e.preventDefault();
    try {
      setWorking("task");
      await req<Task>("/api/tasks", { method: "POST", body: JSON.stringify({ title: taskTitle, status: "todo" }) });
      setTaskTitle("");
      await load();
    } catch (e) {
      setError(getErrorMessage(e));
    } finally {
      setWorking("");
    }
  };

  const addNote = async (e: FormEvent) => {
    e.preventDefault();
    try {
      setWorking("note");
      await req<Note>("/api/notes", { method: "POST", body: JSON.stringify({ title: noteTitle, content: noteContent }) });
      setNoteTitle("");
      setNoteContent("");
      await load();
    } catch (e) {
      setError(getErrorMessage(e));
    } finally {
      setWorking("");
    }
  };

  const completeTask = async (task: Task) => {
    try {
      setWorking(`task-${task.id}`);
      await req<Task>(`/api/tasks/${task.id}`, {
        method: "PUT",
        body: JSON.stringify({ ...task, status: task.status === "done" ? "todo" : "done" }),
      });
      await load();
    } catch (e) {
      setError(getErrorMessage(e));
    } finally {
      setWorking("");
    }
  };

  const removeTask = async (id: number) => {
    try {
      setWorking(`delete-task-${id}`);
      await req<void>(`/api/tasks/${id}`, { method: "DELETE" });
      await load();
    } catch (e) {
      setError(getErrorMessage(e));
    } finally {
      setWorking("");
    }
  };

  const removeNote = async (id: number) => {
    try {
      setWorking(`delete-note-${id}`);
      await req<void>(`/api/notes/${id}`, { method: "DELETE" });
      if (selectedNote === id) setIsEditing(false);
      await load();
    } catch (e) {
      setError(getErrorMessage(e));
    } finally {
      setWorking("");
    }
  };

  const startEditing = () => {
    const note = notes.find((n) => n.id === selectedNote);
    if (note) {
      setEditTitle(note.title);
      setEditContent(note.content);
      setIsEditing(true);
    }
  };

  const saveEdit = async () => {
    if (!selectedNote || !selectedNoteData) return;
    try {
      setWorking("edit");
      await req<Note>(`/api/notes/${selectedNote}`, {
        method: "PUT",
        body: JSON.stringify({ ...selectedNoteData, title: editTitle, content: editContent }),
      });
      setIsEditing(false);
      await load();
    } catch (e) {
      setError(getErrorMessage(e));
    } finally {
      setWorking("");
    }
  };

  const fetchSummary = async () => {
    try {
      setWorking("summary");
      setError("");
      const data = await req<{ topic: string; summary: string }>(`/api/wiki/summary?topic=${encodeURIComponent(topic)}`);
      setWikiSummary(data.summary);
    } catch (e) {
      setError(getErrorMessage(e));
    } finally {
      setWorking("");
    }
  };

  const enrichNote = async () => {
    if (!selectedNote) return;
    try {
      setWorking("enrich");
      await req<Note>(`/api/notes/${selectedNote}/enrich`, { method: "POST", body: JSON.stringify({ topic }) });
      setWikiSummary("");
      await load();
    } catch (e) {
      setError(getErrorMessage(e));
    } finally {
      setWorking("");
    }
  };

  const selectedNoteData = notes.find((note) => note.id === selectedNote) || null;
  const doneTasks = tasks.filter((task) => task.status === "done").length;

  return (
    <main className="page">
      <section className="hero">
        <div>
          <p className="eyebrow">Study Workspace</p>
          <h1>SyncSpace Edu</h1>
          <p className="sub">
            Keep coursework, learning notes, and quick research in one lightweight dashboard.
          </p>
        </div>
        <div className="stats">
          <article className="stat">
            <span className="statValue">{tasks.length}</span>
            <span className="statLabel">Tasks tracked</span>
          </article>
          <article className="stat">
            <span className="statValue">{doneTasks}</span>
            <span className="statLabel">Completed</span>
          </article>
          <article className="stat">
            <span className="statValue">{notes.length}</span>
            <span className="statLabel">Notes stored</span>
          </article>
        </div>
      </section>

      {error && <p className="banner error">{error}</p>}
      {loading && <p className="banner">Loading workspace...</p>}

      <section className="grid">
        <section className="card">
          <div className="sectionHead">
            <div>
              <p className="sectionTag">Planner</p>
              <h2>Tasks</h2>
            </div>
            <span className="muted">{doneTasks}/{tasks.length || 0} done</span>
          </div>
          <form onSubmit={addTask} className="stack">
            <input value={taskTitle} onChange={(e) => setTaskTitle(e.target.value)} placeholder="Add a new task" required />
            <button disabled={working === "task"}>{working === "task" ? "Adding..." : "Add task"}</button>
          </form>
          <ul className="list">
            {tasks.map((task) => (
              <li key={task.id} className={`listItem ${task.status === "done" ? "done" : ""}`}>
                <div>
                  <strong>{task.title}</strong>
                  <p>{task.status === "done" ? "Completed" : "In progress"}</p>
                </div>
                <div className="actions">
                  <button className="ghost" onClick={() => completeTask(task)} disabled={working === `task-${task.id}`}>
                    {task.status === "done" ? "Reopen" : "Complete"}
                  </button>
                  <button className="danger" onClick={() => removeTask(task.id)} disabled={working === `delete-task-${task.id}`}>
                    Delete
                  </button>
                </div>
              </li>
            ))}
            {!tasks.length && !loading && <li className="empty">No tasks yet. Add one to start the study queue.</li>}
          </ul>
        </section>

        <section className="card">
          <div className="sectionHead">
            <div>
              <p className="sectionTag">Notebook</p>
              <h2>Notes</h2>
            </div>
            <span className="muted">{selectedNoteData ? "1 selected" : "Select a note"}</span>
          </div>
          <form onSubmit={addNote} className="stack">
            <input value={noteTitle} onChange={(e) => setNoteTitle(e.target.value)} placeholder="Note title" required />
            <textarea value={noteContent} onChange={(e) => setNoteContent(e.target.value)} placeholder="Capture key ideas, formulas, or revision notes" rows={5} />
            <button disabled={working === "note"}>{working === "note" ? "Saving..." : "Add note"}</button>
          </form>
          <ul className="list">
            {notes.map((note) => (
              <li key={note.id} className={`listItem selectable ${selectedNote === note.id ? "selected" : ""}`}>
                <label className="noteChoice">
                  <input
                    type="radio"
                    checked={selectedNote === note.id}
                    onChange={() => {
                      setSelectedNote(note.id);
                      setIsEditing(false);
                    }}
                  />
                  <span>
                    <strong>{note.title}</strong>
                    <p>{note.content.slice(0, 88) || "Empty note"}</p>
                  </span>
                </label>
                <button className="danger" onClick={() => removeNote(note.id)} disabled={working === `delete-note-${note.id}`}>
                  Delete
                </button>
              </li>
            ))}
            {!notes.length && !loading && <li className="empty">No notes yet. Create one before enrichment.</li>}
          </ul>
        </section>
      </section>

      <section className="grid lower">
        <section className="card">
          <div className="sectionHead">
            <div>
              <p className="sectionTag">Research</p>
              <h2>Wikipedia Enrichment</h2>
            </div>
            <span className="muted">{selectedNoteData ? selectedNoteData.title : "No note selected"}</span>
          </div>
          <div className="stack">
            <input value={topic} onChange={(e) => setTopic(e.target.value)} placeholder="Try: Calculus, Photosynthesis, Linear algebra" />
            <div className="actions">
              <button onClick={fetchSummary} disabled={!topic || working === "summary"}>
                {working === "summary" ? "Fetching..." : "Preview summary"}
              </button>
              <button className="ghost" onClick={enrichNote} disabled={!selectedNote || !topic || working === "enrich"}>
                {working === "enrich" ? "Enriching..." : "Enrich note"}
              </button>
            </div>
          </div>
          {wikiSummary ? <p className="summary">{wikiSummary}</p> : <p className="empty">Preview a topic summary before inserting it into a note.</p>}
        </section>

        <section className="card focusCard">
          <div className="sectionHead">
            <div style={{ flex: 1 }}>
              <p className="sectionTag">Selected Note</p>
              {isEditing ? (
                <input
                  value={editTitle}
                  onChange={(e) => setEditTitle(e.target.value)}
                  placeholder="Note title"
                  required
                />
              ) : (
                <h2>{selectedNoteData?.title || "Nothing selected"}</h2>
              )}
            </div>
            {selectedNoteData && !isEditing && (
              <button className="ghost" onClick={startEditing}>
                Edit Note
              </button>
            )}
          </div>
          <div className={isEditing ? "stack" : "notePreview"}>
            {isEditing ? (
              <>
                <textarea
                  value={editContent}
                  onChange={(e) => setEditContent(e.target.value)}
                  placeholder="Note content"
                  rows={10}
                />
                <div className="actions">
                  <button onClick={saveEdit} disabled={working === "edit"}>
                    {working === "edit" ? "Saving..." : "Save Changes"}
                  </button>
                  <button className="ghost" onClick={() => setIsEditing(false)} disabled={working === "edit"}>
                    Cancel
                  </button>
                </div>
              </>
            ) : (
              selectedNoteData ? selectedNoteData.content || "This note is empty." : "Choose a note from the notebook to preview it here."
            )}
          </div>
        </section>
      </section>
    </main>
  );
}
