import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { MemoryRouter, Routes, Route } from "react-router-dom";
import { WorkspaceListPage } from "../pages/WorkspaceListPage";
import { WorkspaceDetailPage } from "../pages/WorkspaceDetailPage";
import { AuthProvider } from "../context/AuthContext";
import { api } from "../services/api";

const mockNavigate = vi.fn();

vi.mock("react-router-dom", async () => {
  const actual = (await vi.importActual("react-router-dom")) as any;
  return {
    ...actual,
    useNavigate: () => mockNavigate,
    useParams: () => ({ workspaceId: "123" }),
  };
});

vi.mock("../services/api", () => {
  return {
    api: {
      me: vi.fn(),
      listWorkspaces: vi.fn(),
      createWorkspace: vi.fn(),
      getWorkspace: vi.fn(),
      listNotes: vi.fn(),
      updateWorkspace: vi.fn(),
      deleteWorkspace: vi.fn(),
    },
  };
});

describe("Workspace Pages", () => {
  const mockUser = {
    id: 1,
    email: "creator@example.com",
    name: "Creator User",
    role: "creator",
    status: "active",
  };

  beforeEach(() => {
    vi.clearAllMocks();
    localStorage.clear();
    localStorage.setItem("token", "valid-token");
    vi.mocked(api.me).mockResolvedValue(mockUser);
  });

  describe("WorkspaceListPage", () => {
    it("renders workspace list and stat count", async () => {
      const mockWorkspaces = [
        { id: 1, name: "Workspace One", description: "First workspace", created_at: "2026-01-01T00:00:00Z" },
        { id: 2, name: "Workspace Two", description: "Second workspace", created_at: "2026-01-02T00:00:00Z" },
      ];
      vi.mocked(api.listWorkspaces).mockResolvedValue(mockWorkspaces);

      render(
        <MemoryRouter>
          <AuthProvider>
            <WorkspaceListPage />
          </AuthProvider>
        </MemoryRouter>
      );

      await waitFor(() => {
        expect(screen.getByText("Workspace One")).toBeInTheDocument();
        expect(screen.getByText("Workspace Two")).toBeInTheDocument();
        expect(screen.getByText("2")).toBeInTheDocument(); // Stat counter
        expect(screen.getByText("Creator Tools")).toBeInTheDocument(); // Rendered because user is creator
      });
    });

    it("opens workspace creation form and creates workspace", async () => {
      vi.mocked(api.listWorkspaces).mockResolvedValue([]);
      vi.mocked(api.createWorkspace).mockResolvedValue({ id: 3, name: "New Space", description: "A new space" });

      render(
        <MemoryRouter>
          <AuthProvider>
            <WorkspaceListPage />
          </AuthProvider>
        </MemoryRouter>
      );

      // Wait for loading to finish
      await waitFor(() => {
        expect(screen.queryByText("Loading workspaces...")).not.toBeInTheDocument();
      });

      // Open creation form
      const newWsBtn = screen.getByRole("button", { name: "+ New Workspace" });
      fireEvent.click(newWsBtn);

      expect(screen.getByText("Create Workspace")).toBeInTheDocument();

      // Fill in details and submit
      const nameInput = screen.getByPlaceholderText("My Notes");
      const descInput = screen.getByPlaceholderText("Optional description");
      fireEvent.change(nameInput, { target: { value: "New Space" } });
      fireEvent.change(descInput, { target: { value: "A new space" } });

      const createBtn = screen.getByRole("button", { name: "Create" });
      fireEvent.click(createBtn);

      await waitFor(() => {
        expect(api.createWorkspace).toHaveBeenCalledWith({
          name: "New Space",
          description: "A new space",
        });
        expect(screen.queryByText("Create Workspace")).not.toBeInTheDocument();
      });
    });
  });

  describe("WorkspaceDetailPage", () => {
    const mockWorkspace = {
      id: 123,
      name: "Detail Workspace",
      description: "Workspace detail desc",
      created_at: "2026-01-01T00:00:00Z",
      updated_at: "2026-01-01T00:00:00Z",
    };
    const mockNotes = [
      { id: 1, title: "Note One", content: "<p>Note content one</p>", creator_name: "Creator User", created_at: "2026-01-01T00:00:00Z", updated_at: "2026-01-01T00:00:00Z" },
    ];

    it("renders workspace detail information and note list", async () => {
      vi.mocked(api.getWorkspace).mockResolvedValue(mockWorkspace);
      vi.mocked(api.listNotes).mockResolvedValue(mockNotes);

      render(
        <MemoryRouter>
          <AuthProvider>
            <WorkspaceDetailPage />
          </AuthProvider>
        </MemoryRouter>
      );

      await waitFor(() => {
        expect(screen.getByRole("heading", { name: "Detail Workspace" })).toBeInTheDocument();
        expect(screen.getByText("Workspace detail desc")).toBeInTheDocument();
        expect(screen.getByText("Note One")).toBeInTheDocument();
        expect(screen.getByText("Note content one")).toBeInTheDocument();
        expect(screen.getByText("1")).toBeInTheDocument(); // Note count stat
      });
    });

    it("allows editing workspace details", async () => {
      vi.mocked(api.getWorkspace).mockResolvedValue(mockWorkspace);
      vi.mocked(api.listNotes).mockResolvedValue(mockNotes);
      vi.mocked(api.updateWorkspace).mockResolvedValue({ ...mockWorkspace, name: "Updated Name" });

      render(
        <MemoryRouter>
          <AuthProvider>
            <WorkspaceDetailPage />
          </AuthProvider>
        </MemoryRouter>
      );

      // Wait for loading to finish
      await waitFor(() => {
        expect(screen.getByText("Detail Workspace")).toBeInTheDocument();
      });

      // Click Edit
      fireEvent.click(screen.getByRole("button", { name: "Edit Workspace" }));

      // Edit name
      const nameInput = screen.getByPlaceholderText("Workspace name");
      fireEvent.change(nameInput, { target: { value: "Updated Name" } });

      // Click Save Changes
      fireEvent.click(screen.getByRole("button", { name: "Save Changes" }));

      await waitFor(() => {
        expect(api.updateWorkspace).toHaveBeenCalledWith(123, {
          name: "Updated Name",
          description: "Workspace detail desc",
        });
      });
    });

    it("deletes workspace after confirmation", async () => {
      vi.mocked(api.getWorkspace).mockResolvedValue(mockWorkspace);
      vi.mocked(api.listNotes).mockResolvedValue([]);
      vi.mocked(api.deleteWorkspace).mockResolvedValue({ success: true });

      // Mock confirm dialog
      const confirmSpy = vi.spyOn(window, "confirm").mockReturnValue(true);

      render(
        <MemoryRouter>
          <AuthProvider>
            <WorkspaceDetailPage />
          </AuthProvider>
        </MemoryRouter>
      );

      // Wait for loading to finish
      await waitFor(() => {
        expect(screen.getByText("Detail Workspace")).toBeInTheDocument();
      });

      // Click Delete
      fireEvent.click(screen.getByRole("button", { name: "Delete" }));

      await waitFor(() => {
        expect(confirmSpy).toHaveBeenCalledWith("Delete this workspace and all its notes?");
        expect(api.deleteWorkspace).toHaveBeenCalledWith(123);
        expect(mockNavigate).toHaveBeenCalledWith("/workspaces");
      });
    });
  });
});
