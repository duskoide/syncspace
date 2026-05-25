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

  // Boards (formerly Classrooms)
  listBoards: () => request("/api/boards"),
  createBoard: (data: { name: string; description: string; visibility: string }) =>
    request("/api/boards", { method: "POST", body: JSON.stringify(data) }),
  getBoard: (id: number) => request(`/api/boards/${id}`),
  updateBoard: (id: number, data: { name: string; description: string; visibility: string }) =>
    request(`/api/boards/${id}`, { method: "PUT", body: JSON.stringify(data) }),
  deleteBoard: (id: number) =>
    request(`/api/boards/${id}`, { method: "DELETE" }),
  joinBoard: (id: number) =>
    request(`/api/boards/${id}/join`, { method: "POST" }),
  approveMembership: (id: number) =>
    request(`/api/memberships/${id}/approve`, { method: "PUT" }),
  getBoardMembers: (id: number) =>
    request(`/api/boards/${id}/members`),

  // Board Images
  listBoardImages: (boardId: number) =>
    request(`/api/boards/${boardId}/images`),
  uploadBoardImage: (boardId: number, file: File) => {
    const form = new FormData();
    form.append("file", file);
    form.append("board_id", String(boardId));
    return request("/api/upload", { method: "POST", body: form, headers: {} });
  },

  // Discussions (board_id instead of classroom_id)
  listDiscussions: (boardId: number) =>
    request(`/api/discussions?board_id=${boardId}`),
  createDiscussion: (data: { board_id: number; message: string; parent_id?: number }) =>
    request("/api/discussions", { method: "POST", body: JSON.stringify(data) }),

  // Files
  uploadFile: (file: File, boardId?: number) => {
    const form = new FormData();
    form.append("file", file);
    if (boardId) form.append("board_id", String(boardId));
    return request("/api/upload", { method: "POST", body: form, headers: {} });
  },

  // Wikipedia API
  wikiSummary: (topic: string) =>
    request(`/api/wiki/summary?topic=${encodeURIComponent(topic)}`),
};

export { ApiError };
