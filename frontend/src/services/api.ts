const API_BASE = import.meta.env.VITE_API_URL || "http://localhost:8080";

class ApiError extends Error {
  constructor(public status: number, public code: string, message: string) {
    super(message);
  }
}

async function request(path: string, options: RequestInit = {}) {
  const url = `${API_BASE}${path}`;
  const token = localStorage.getItem("token");
  const headers: Record<string, string> = {};
  
  // Only set Content-Type if not FormData (browser will set it automatically for FormData)
  const isFormData = options.body instanceof FormData;
  if (!isFormData) {
    headers["Content-Type"] = "application/json";
  }
  
  // Merge with provided headers
  if (options.headers) {
    Object.assign(headers, options.headers as Record<string, string>);
  }
  
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

  // Admin - Users
  listUsers: (params?: { role?: string; status?: string }) => {
    const qs = new URLSearchParams();
    if (params?.role) qs.append("role", params.role);
    if (params?.status) qs.append("status", params.status);
    return request(`/api/admin/users?${qs}`);
  },
  activateUser: (id: number) =>
    request(`/api/admin/users/${id}/activate`, { method: "PUT" }),
  suspendUser: (id: number) =>
    request(`/api/admin/users/${id}/suspend`, { method: "PUT" }),

  // Admin - Templates
  listAllTemplates: () => request("/api/admin/templates"),
  setTemplateHidden: (id: number, isHidden: boolean) =>
    request(`/api/admin/templates/${id}`, { 
      method: "PATCH", 
      body: JSON.stringify({ is_hidden: isHidden }) 
    }),

  // Workspaces
  listWorkspaces: () => request("/api/workspaces"),
  createWorkspace: (data: { name: string; description: string }) =>
    request("/api/workspaces", { method: "POST", body: JSON.stringify(data) }),
  getWorkspace: (id: number) => request(`/api/workspaces/${id}`),
  updateWorkspace: (id: number, data: { name: string; description: string }) =>
    request(`/api/workspaces/${id}`, { method: "PUT", body: JSON.stringify(data) }),
  deleteWorkspace: (id: number) =>
    request(`/api/workspaces/${id}`, { method: "DELETE" }),

  // Notes
  listNotes: (workspaceId: number) => request(`/api/workspaces/${workspaceId}/notes`),
  createNote: (workspaceId: number, data: { title: string }) =>
    request(`/api/workspaces/${workspaceId}/notes`, { method: "POST", body: JSON.stringify(data) }),
  getNote: (id: number) => request(`/api/notes/${id}`),
  updateNote: (id: number, data: { title: string; content: string }) =>
    request(`/api/notes/${id}`, { method: "PUT", body: JSON.stringify(data) }),
  deleteNote: (id: number) =>
    request(`/api/notes/${id}`, { method: "DELETE" }),

  // Templates
  listTemplates: (search?: string) => {
    const qs = new URLSearchParams();
    if (search) qs.append("search", search);
    return request(`/api/templates?${qs}`);
  },
  listMyTemplates: () => request("/api/templates/my"),
  getTemplate: (id: number) => request(`/api/templates/${id}`),
  createTemplate: (data: { type: string; source_id: number; name: string; description: string; visibility: string }) =>
    request("/api/templates", { method: "POST", body: JSON.stringify(data) }),
  updateTemplate: (id: number, data: { name: string; description: string; visibility: string }) =>
    request(`/api/templates/${id}`, { method: "PUT", body: JSON.stringify(data) }),
  updateTemplateContent: (id: number) =>
    request(`/api/templates/${id}/update-content`, { method: "POST" }),
  deleteTemplate: (id: number) =>
    request(`/api/templates/${id}`, { method: "DELETE" }),
  cloneTemplate: (id: number, targetWorkspaceId?: number) =>
    request(`/api/templates/${id}/clone`, { 
      method: "POST", 
      body: JSON.stringify({ target_workspace_id: targetWorkspaceId }) 
    }),

  // Note Images
  uploadNoteImage: (noteId: number, file: File) => {
    const form = new FormData();
    form.append("file", file);
    form.append("note_id", String(noteId));
    return request("/api/upload", { method: "POST", body: form, headers: {} });
  },
  deleteNoteImage: (id: number) =>
    request(`/api/files/${id}`, { method: "DELETE" }),

  // Wikipedia API
  wikiSummary: (topic: string) =>
    request(`/api/wiki/summary?topic=${encodeURIComponent(topic)}`),
};

export { ApiError };
