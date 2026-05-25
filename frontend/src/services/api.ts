const API_BASE = import.meta.env.VITE_API_URL || "http://localhost:8080";

class ApiError extends Error {
  constructor(public status: number, public code: string, message: string) {
    super(message);
  }
}

async function request(path: string, options: RequestInit = {}) {
  const url = `${API_BASE}${path}`;
  const token = localStorage.getItem("token");
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...((options.headers as Record<string, string>) || {}),
  };
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  const resp = await fetch(url, { ...options, headers });
  const data = await resp.json().catch(() => null);

  if (!resp.ok) {
    throw new ApiError(
      resp.status,
      data?.error?.code || "unknown",
      data?.error?.message || `HTTP ${resp.status}`
    );
  }
  return data;
}

export const api = {
  // Auth
  login: (email: string, password: string) =>
    request("/api/auth/login", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    }),
  register: (data: { email: string; password: string; name: string; role: string }) =>
    request("/api/auth/register", { method: "POST", body: JSON.stringify(data) }),
  me: () => request("/api/auth/me"),

  // Admin
  listUsers: (params?: { role?: string; status?: string }) => {
    const qs = new URLSearchParams();
    if (params?.role) qs.append("role", params.role);
    if (params?.status) qs.append("status", params.status);
    return request(`/api/admin/users?${qs}`);
  },
  approveUser: (id: number) =>
    request(`/api/admin/users/${id}/approve`, { method: "PUT" }),
  suspendUser: (id: number) =>
    request(`/api/admin/users/${id}/suspend`, { method: "PUT" }),

  // Classrooms
  listClassrooms: () => request("/api/classrooms"),
  createClassroom: (data: { name: string; description: string }) =>
    request("/api/classrooms", { method: "POST", body: JSON.stringify(data) }),
  getClassroom: (id: number) => request(`/api/classrooms/${id}`),
  updateClassroom: (id: number, data: { name: string; description: string }) =>
    request(`/api/classrooms/${id}`, { method: "PUT", body: JSON.stringify(data) }),
  deleteClassroom: (id: number) =>
    request(`/api/classrooms/${id}`, { method: "DELETE" }),
  enroll: (id: number) =>
    request(`/api/classrooms/${id}/enroll`, { method: "POST" }),
  approveEnrollment: (id: number) =>
    request(`/api/enrollments/${id}/approve`, { method: "PUT" }),

  // Materials
  listMaterials: (classroomId: number) =>
    request(`/api/materials?classroom_id=${classroomId}`),
  createMaterial: (data: { classroom_id: number; title: string; content: string; tags: string }) =>
    request("/api/materials", { method: "POST", body: JSON.stringify(data) }),
  getMaterial: (id: number) => request(`/api/materials/${id}`),
  updateMaterial: (id: number, data: { title: string; content: string; tags: string }) =>
    request(`/api/materials/${id}`, { method: "PUT", body: JSON.stringify(data) }),
  deleteMaterial: (id: number) =>
    request(`/api/materials/${id}`, { method: "DELETE" }),

  // Assignments
  listAssignments: (classroomId: number) =>
    request(`/api/assignments?classroom_id=${classroomId}`),
  createAssignment: (data: { classroom_id: number; title: string; description: string; due_date: string; max_score: number }) =>
    request("/api/assignments", { method: "POST", body: JSON.stringify(data) }),
  getAssignment: (id: number) => request(`/api/assignments/${id}`),
  deleteAssignment: (id: number) =>
    request(`/api/assignments/${id}`, { method: "DELETE" }),
  submitWork: (id: number, content: string) =>
    request(`/api/assignments/${id}/submissions`, {
      method: "POST",
      body: JSON.stringify({ content }),
    }),
  listSubmissions: (id: number) =>
    request(`/api/assignments/${id}/submissions`),
  gradeSubmission: (id: number, score: number, feedback: string) =>
    request(`/api/submissions/${id}/grade`, {
      method: "PUT",
      body: JSON.stringify({ score, feedback }),
    }),

  // Collaborative Notes
  listNotes: (classroomId: number) =>
    request(`/api/collaborative-notes?classroom_id=${classroomId}`),
  createNote: (data: { classroom_id: number; title: string; content: string }) =>
    request("/api/collaborative-notes", { method: "POST", body: JSON.stringify(data) }),
  updateNote: (id: number, data: { title: string; content: string }) =>
    request(`/api/collaborative-notes/${id}`, { method: "PUT", body: JSON.stringify(data) }),
  deleteNote: (id: number) =>
    request(`/api/collaborative-notes/${id}`, { method: "DELETE" }),

  // Discussions
  listDiscussions: (classroomId: number) =>
    request(`/api/discussions?classroom_id=${classroomId}`),
  createDiscussion: (data: { classroom_id: number; message: string; parent_id?: number }) =>
    request("/api/discussions", { method: "POST", body: JSON.stringify(data) }),

  // Files
  uploadFile: (file: File, materialId?: number) => {
    const form = new FormData();
    form.append("file", file);
    if (materialId) form.append("material_id", String(materialId));
    return request("/api/upload", { method: "POST", body: form, headers: {} });
  },
};

export { ApiError };
