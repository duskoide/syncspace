import { useState, useEffect } from "react";
import { useAuth } from "../context/AuthContext";
import { api } from "../services/api";

interface Classroom {
  id: number;
  name: string;
  description: string;
  teacher_id: number;
  teacher_name: string;
  created_at: string;
}

interface Material {
  id: number;
  title: string;
  content: string;
  tags: string;
  created_at: string;
}

interface Assignment {
  id: number;
  title: string;
  description: string;
  due_date: string;
  max_score: number;
  created_at: string;
}

interface Discussion {
  id: number;
  user_name: string;
  message: string;
  created_at: string;
}

interface Enrollment {
  id: number;
  student_name: string;
  student_email: string;
  status: string;
}

export function ClassroomPage() {
  const { user } = useAuth();
  const [classrooms, setClassrooms] = useState<Classroom[]>([]);
  const [selected, setSelected] = useState<Classroom | null>(null);
  const [materials, setMaterials] = useState<Material[]>([]);
  const [assignments, setAssignments] = useState<Assignment[]>([]);
  const [discussions, setDiscussions] = useState<Discussion[]>([]);
  const [enrollments, setEnrollments] = useState<Enrollment[]>([]);
  const [tab, setTab] = useState("materials");
  const [showCreate, setShowCreate] = useState(false);
  const [newClassroom, setNewClassroom] = useState({ name: "", description: "" });
  const [newMaterial, setNewMaterial] = useState({ title: "", content: "", tags: "" });
  const [newAssignment, setNewAssignment] = useState({ title: "", description: "", due_date: "", max_score: 100 });
  const [newDiscussion, setNewDiscussion] = useState("");
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadClassrooms();
  }, []);

  const loadClassrooms = async () => {
    setLoading(true);
    try {
      const data = await api.listClassrooms();
      setClassrooms(data);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const selectClassroom = async (c: Classroom) => {
    setSelected(c);
    setTab("materials");
    try {
      const [mats, assigns, discs] = await Promise.all([
        api.listMaterials(c.id),
        api.listAssignments(c.id),
        api.listDiscussions(c.id),
      ]);
      setMaterials(mats);
      setAssignments(assigns);
      setDiscussions(discs);
      if (c.teacher_id === user?.id) {
        // Load enrollments for teachers
        const enrolls = await fetch(`http://localhost:8080/api/classrooms/${c.id}/students`, {
          headers: { Authorization: `Bearer ${localStorage.getItem("token")}` },
        }).then((r) => (r.ok ? r.json() : []));
        setEnrollments(enrolls);
      }
    } catch (err) {
      console.error(err);
    }
  };

  const createClassroom = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await api.createClassroom(newClassroom);
      setShowCreate(false);
      setNewClassroom({ name: "", description: "" });
      loadClassrooms();
    } catch (err: any) {
      alert(err.message);
    }
  };

  const createMaterial = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selected) return;
    try {
      await api.createMaterial({ ...newMaterial, classroom_id: selected.id });
      setNewMaterial({ title: "", content: "", tags: "" });
      const mats = await api.listMaterials(selected.id);
      setMaterials(mats);
    } catch (err: any) {
      alert(err.message);
    }
  };

  const createAssignment = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selected) return;
    try {
      await api.createAssignment({ ...newAssignment, classroom_id: selected.id });
      setNewAssignment({ title: "", description: "", due_date: "", max_score: 100 });
      const assigns = await api.listAssignments(selected.id);
      setAssignments(assigns);
    } catch (err: any) {
      alert(err.message);
    }
  };

  const createDiscussion = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selected) return;
    try {
      await api.createDiscussion({ classroom_id: selected.id, message: newDiscussion });
      setNewDiscussion("");
      const discs = await api.listDiscussions(selected.id);
      setDiscussions(discs);
    } catch (err: any) {
      alert(err.message);
    }
  };

  const enroll = async (classroomId: number) => {
    try {
      await api.enroll(classroomId);
      alert("Enrollment requested! Waiting for teacher approval.");
      loadClassrooms();
    } catch (err: any) {
      alert(err.message);
    }
  };

  const isTeacher = selected?.teacher_id === user?.id;
  const isStudent = user?.role === "student";

  if (loading) return <div style={{ padding: 24 }}>Loading...</div>;

  return (
    <div style={{ padding: 24, display: "flex", gap: 24 }}>
      {/* Sidebar */}
      <div style={{ width: 280, flexShrink: 0 }}>
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 12 }}>
          <h2 style={{ color: "#1f2937", fontSize: 18 }}>Classrooms</h2>
          {user?.role === "teacher" && (
            <button
              onClick={() => setShowCreate(true)}
              style={{
                padding: "6px 12px",
                background: "#2563eb",
                color: "#fff",
                border: "none",
                borderRadius: 6,
                cursor: "pointer",
                fontSize: 13,
              }}
            >
              + New
            </button>
          )}
        </div>

        {showCreate && (
          <form onSubmit={createClassroom} style={{ background: "#fff", padding: 16, borderRadius: 8, marginBottom: 12, border: "1px solid #e5e7eb" }}>
            <input
              placeholder="Classroom name"
              value={newClassroom.name}
              onChange={(e) => setNewClassroom({ ...newClassroom, name: e.target.value })}
              required
              style={{ width: "100%", padding: 8, marginBottom: 8, border: "1px solid #d1d5db", borderRadius: 4 }}
            />
            <textarea
              placeholder="Description"
              value={newClassroom.description}
              onChange={(e) => setNewClassroom({ ...newClassroom, description: e.target.value })}
              style={{ width: "100%", padding: 8, marginBottom: 8, border: "1px solid #d1d5db", borderRadius: 4, minHeight: 60 }}
            />
            <div style={{ display: "flex", gap: 8 }}>
              <button type="submit" style={{ padding: "6px 12px", background: "#2563eb", color: "#fff", border: "none", borderRadius: 4, cursor: "pointer" }}>
                Create
              </button>
              <button type="button" onClick={() => setShowCreate(false)} style={{ padding: "6px 12px", background: "#e5e7eb", border: "none", borderRadius: 4, cursor: "pointer" }}>
                Cancel
              </button>
            </div>
          </form>
        )}

        <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
          {classrooms.map((c) => (
            <div
              key={c.id}
              onClick={() => selectClassroom(c)}
              style={{
                padding: 12,
                background: selected?.id === c.id ? "#eff6ff" : "#fff",
                borderRadius: 8,
                cursor: "pointer",
                border: "1px solid",
                borderColor: selected?.id === c.id ? "#2563eb" : "#e5e7eb",
              }}
            >
              <div style={{ fontWeight: 600, color: "#1f2937", marginBottom: 4 }}>{c.name}</div>
              <div style={{ fontSize: 12, color: "#6b7280" }}>{c.teacher_name}</div>
            </div>
          ))}
          {classrooms.length === 0 && (
            <div style={{ color: "#6b7280", fontSize: 14, textAlign: "center", padding: 20 }}>
              No classrooms yet
              {user?.role === "student" && <div style={{ marginTop: 8 }}>Ask your teacher for an enrollment</div>}
            </div>
          )}
        </div>
      </div>

      {/* Main Content */}
      <div style={{ flex: 1 }}>
        {selected ? (
          <div>
            <div style={{ marginBottom: 16 }}>
              <h1 style={{ color: "#1f2937", marginBottom: 4 }}>{selected.name}</h1>
              <p style={{ color: "#6b7280", fontSize: 14 }}>{selected.description}</p>
              <p style={{ color: "#9ca3af", fontSize: 12, marginTop: 4 }}>Teacher: {selected.teacher_name}</p>
            </div>

            {isStudent && (
              <button
                onClick={() => enroll(selected.id)}
                style={{
                  padding: "8px 16px",
                  background: "#059669",
                  color: "#fff",
                  border: "none",
                  borderRadius: 6,
                  cursor: "pointer",
                  marginBottom: 16,
                }}
              >
                Request Enrollment
              </button>
            )}

            <div style={{ display: "flex", gap: 8, marginBottom: 16, borderBottom: "1px solid #e5e7eb" }}>
              {["materials", "assignments", "discussions"].map((t) => (
                <button
                  key={t}
                  onClick={() => setTab(t)}
                  style={{
                    padding: "10px 16px",
                    background: "transparent",
                    border: "none",
                    borderBottom: tab === t ? "2px solid #2563eb" : "2px solid transparent",
                    color: tab === t ? "#2563eb" : "#6b7280",
                    fontWeight: tab === t ? 600 : 400,
                    cursor: "pointer",
                    textTransform: "capitalize",
                  }}
                >
                  {t}
                </button>
              ))}
            </div>

            {tab === "materials" && (
              <div>
                {isTeacher && (
                  <form onSubmit={createMaterial} style={{ background: "#f9fafb", padding: 16, borderRadius: 8, marginBottom: 16 }}>
                    <h4 style={{ marginBottom: 8 }}>Add Material</h4>
                    <input
                      placeholder="Title"
                      value={newMaterial.title}
                      onChange={(e) => setNewMaterial({ ...newMaterial, title: e.target.value })}
                      required
                      style={{ width: "100%", padding: 8, marginBottom: 8, border: "1px solid #d1d5db", borderRadius: 4 }}
                    />
                    <textarea
                      placeholder="Content"
                      value={newMaterial.content}
                      onChange={(e) => setNewMaterial({ ...newMaterial, content: e.target.value })}
                      style={{ width: "100%", padding: 8, marginBottom: 8, border: "1px solid #d1d5db", borderRadius: 4, minHeight: 80 }}
                    />
                    <input
                      placeholder="Tags (comma separated)"
                      value={newMaterial.tags}
                      onChange={(e) => setNewMaterial({ ...newMaterial, tags: e.target.value })}
                      style={{ width: "100%", padding: 8, marginBottom: 8, border: "1px solid #d1d5db", borderRadius: 4 }}
                    />
                    <button type="submit" style={{ padding: "6px 12px", background: "#2563eb", color: "#fff", border: "none", borderRadius: 4, cursor: "pointer" }}>
                      Add Material
                    </button>
                  </form>
                )}
                {materials.map((m) => (
                  <div key={m.id} style={{ background: "#fff", padding: 16, borderRadius: 8, marginBottom: 8, border: "1px solid #e5e7eb" }}>
                    <h4 style={{ color: "#1f2937", marginBottom: 4 }}>{m.title}</h4>
                    <p style={{ color: "#4b5563", fontSize: 14, whiteSpace: "pre-wrap" }}>{m.content}</p>
                    {m.tags && (
                      <div style={{ marginTop: 8 }}>
                        {m.tags.split(",").map((tag) => (
                          <span key={tag} style={{ padding: "2px 8px", background: "#e0e7ff", color: "#4338ca", borderRadius: 12, fontSize: 12, marginRight: 4 }}>
                            {tag.trim()}
                          </span>
                        ))}
                      </div>
                    )}
                  </div>
                ))}
                {materials.length === 0 && <div style={{ color: "#6b7280", padding: 20 }}>No materials yet</div>}
              </div>
            )}

            {tab === "assignments" && (
              <div>
                {isTeacher && (
                  <form onSubmit={createAssignment} style={{ background: "#f9fafb", padding: 16, borderRadius: 8, marginBottom: 16 }}>
                    <h4 style={{ marginBottom: 8 }}>Create Assignment</h4>
                    <input
                      placeholder="Title"
                      value={newAssignment.title}
                      onChange={(e) => setNewAssignment({ ...newAssignment, title: e.target.value })}
                      required
                      style={{ width: "100%", padding: 8, marginBottom: 8, border: "1px solid #d1d5db", borderRadius: 4 }}
                    />
                    <textarea
                      placeholder="Description"
                      value={newAssignment.description}
                      onChange={(e) => setNewAssignment({ ...newAssignment, description: e.target.value })}
                      style={{ width: "100%", padding: 8, marginBottom: 8, border: "1px solid #d1d5db", borderRadius: 4, minHeight: 60 }}
                    />
                    <input
                      type="datetime-local"
                      value={newAssignment.due_date}
                      onChange={(e) => setNewAssignment({ ...newAssignment, due_date: e.target.value })}
                      required
                      style={{ width: "100%", padding: 8, marginBottom: 8, border: "1px solid #d1d5db", borderRadius: 4 }}
                    />
                    <button type="submit" style={{ padding: "6px 12px", background: "#2563eb", color: "#fff", border: "none", borderRadius: 4, cursor: "pointer" }}>
                      Create Assignment
                    </button>
                  </form>
                )}
                {assignments.map((a) => (
                  <div key={a.id} style={{ background: "#fff", padding: 16, borderRadius: 8, marginBottom: 8, border: "1px solid #e5e7eb" }}>
                    <h4 style={{ color: "#1f2937", marginBottom: 4 }}>{a.title}</h4>
                    <p style={{ color: "#4b5563", fontSize: 14 }}>{a.description}</p>
                    <p style={{ color: "#dc2626", fontSize: 12, marginTop: 4 }}>Due: {new Date(a.due_date).toLocaleString()}</p>
                    <p style={{ color: "#6b7280", fontSize: 12 }}>Max Score: {a.max_score}</p>
                  </div>
                ))}
                {assignments.length === 0 && <div style={{ color: "#6b7280", padding: 20 }}>No assignments yet</div>}
              </div>
            )}

            {tab === "discussions" && (
              <div>
                <form onSubmit={createDiscussion} style={{ background: "#f9fafb", padding: 16, borderRadius: 8, marginBottom: 16 }}>
                  <textarea
                    placeholder="Write a message..."
                    value={newDiscussion}
                    onChange={(e) => setNewDiscussion(e.target.value)}
                    required
                    style={{ width: "100%", padding: 8, marginBottom: 8, border: "1px solid #d1d5db", borderRadius: 4, minHeight: 60 }}
                  />
                  <button type="submit" style={{ padding: "6px 12px", background: "#2563eb", color: "#fff", border: "none", borderRadius: 4, cursor: "pointer" }}>
                    Post Message
                  </button>
                </form>
                {discussions.map((d) => (
                  <div key={d.id} style={{ background: "#fff", padding: 12, borderRadius: 8, marginBottom: 8, border: "1px solid #e5e7eb" }}>
                    <div style={{ fontWeight: 600, color: "#2563eb", fontSize: 13, marginBottom: 4 }}>{d.user_name}</div>
                    <p style={{ color: "#4b5563", fontSize: 14 }}>{d.message}</p>
                    <p style={{ color: "#9ca3af", fontSize: 11, marginTop: 4 }}>{new Date(d.created_at).toLocaleString()}</p>
                  </div>
                ))}
                {discussions.length === 0 && <div style={{ color: "#6b7280", padding: 20 }}>No discussions yet. Be the first to post!</div>}
              </div>
            )}
          </div>
        ) : (
          <div style={{ color: "#6b7280", textAlign: "center", padding: 60 }}>
            Select a classroom to view its contents
          </div>
        )}
      </div>
    </div>
  );
}
