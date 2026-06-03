import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { MemoryRouter, Routes, Route } from "react-router-dom";
import { AuthProvider } from "../context/AuthContext";
import { ProtectedRoute } from "../components/ProtectedRoute";
import { Navbar } from "../components/Navbar";
import { api } from "../services/api";

vi.mock("../services/api", () => {
  return {
    api: {
      me: vi.fn(),
      login: vi.fn(),
    },
  };
});

describe("ProtectedRoute and Navbar", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    localStorage.clear();
  });

  describe("ProtectedRoute", () => {
    it("redirects unauthenticated users to /login", async () => {
      render(
        <MemoryRouter initialEntries={["/dashboard"]}>
          <AuthProvider>
            <Routes>
              <Route path="/login" element={<div>Login Page</div>} />
              <Route
                path="/dashboard"
                element={
                  <ProtectedRoute>
                    <div>Dashboard Content</div>
                  </ProtectedRoute>
                }
              />
            </Routes>
          </AuthProvider>
        </MemoryRouter>
      );

      // Since there is no token in localStorage, it should immediately navigate to /login
      await waitFor(() => {
        expect(screen.getByText("Login Page")).toBeInTheDocument();
        expect(screen.queryByText("DashboardContent")).not.toBeInTheDocument();
      });
    });

    it("renders children when the user is authenticated", async () => {
      localStorage.setItem("token", "valid-token");
      vi.mocked(api.me).mockResolvedValue({
        id: 1,
        email: "user@example.com",
        name: "Regular User",
        role: "user",
        status: "active",
      });

      render(
        <MemoryRouter initialEntries={["/dashboard"]}>
          <AuthProvider>
            <Routes>
              <Route path="/login" element={<div>Login Page</div>} />
              <Route
                path="/dashboard"
                element={
                  <ProtectedRoute>
                    <div>Dashboard Content</div>
                  </ProtectedRoute>
                }
              />
            </Routes>
          </AuthProvider>
        </MemoryRouter>
      );

      // Wait for api.me to load and user state to populate
      await waitFor(() => {
        expect(screen.getByText("Dashboard Content")).toBeInTheDocument();
        expect(screen.queryByText("Login Page")).not.toBeInTheDocument();
      });
    });

    it("redirects authenticated users to /dashboard if they lack required roles", async () => {
      localStorage.setItem("token", "valid-token");
      vi.mocked(api.me).mockResolvedValue({
        id: 1,
        email: "user@example.com",
        name: "Regular User",
        role: "user",
        status: "active",
      });

      render(
        <MemoryRouter initialEntries={["/admin"]}>
          <AuthProvider>
            <Routes>
              <Route path="/dashboard" element={<div>Dashboard Content</div>} />
              <Route
                path="/admin"
                element={
                  <ProtectedRoute roles={["superadmin"]}>
                    <div>Admin Content</div>
                  </ProtectedRoute>
                }
              />
            </Routes>
          </AuthProvider>
        </MemoryRouter>
      );

      await waitFor(() => {
        expect(screen.getByText("Dashboard Content")).toBeInTheDocument();
        expect(screen.queryByText("Admin Content")).not.toBeInTheDocument();
      });
    });

    it("renders children when user possesses one of the required roles", async () => {
      localStorage.setItem("token", "valid-token");
      vi.mocked(api.me).mockResolvedValue({
        id: 2,
        email: "admin@example.com",
        name: "Admin User",
        role: "superadmin",
        status: "active",
      });

      render(
        <MemoryRouter initialEntries={["/admin"]}>
          <AuthProvider>
            <Routes>
              <Route path="/dashboard" element={<div>Dashboard Content</div>} />
              <Route
                path="/admin"
                element={
                  <ProtectedRoute roles={["superadmin"]}>
                    <div>Admin Content</div>
                  </ProtectedRoute>
                }
              />
            </Routes>
          </AuthProvider>
        </MemoryRouter>
      );

      await waitFor(() => {
        expect(screen.getByText("Admin Content")).toBeInTheDocument();
      });
    });
  });

  describe("Navbar", () => {
    it("renders nothing when user is not logged in", async () => {
      render(
        <MemoryRouter>
          <AuthProvider>
            <Navbar />
          </AuthProvider>
        </MemoryRouter>
      );

      // No user is logged in, so Navbar shouldn't render anything
      await new Promise((r) => setTimeout(r, 100)); // wait brief moment
      expect(screen.queryByRole("navigation")).not.toBeInTheDocument();
    });

    it("renders correct navigation links for regular users (no Admin link)", async () => {
      localStorage.setItem("token", "valid-token");
      vi.mocked(api.me).mockResolvedValue({
        id: 1,
        email: "user@example.com",
        name: "Jane Doe",
        role: "user",
        status: "active",
      });

      render(
        <MemoryRouter>
          <AuthProvider>
            <Navbar />
          </AuthProvider>
        </MemoryRouter>
      );

      await waitFor(() => {
        expect(screen.getByRole("navigation")).toBeInTheDocument();
        expect(screen.getByText("Dashboard")).toBeInTheDocument();
        expect(screen.getByText("Workspaces")).toBeInTheDocument();
        expect(screen.getByText("Templates")).toBeInTheDocument();
        expect(screen.queryByText("Admin")).not.toBeInTheDocument();
        expect(screen.getByText("Jane Doe (user)")).toBeInTheDocument();
      });
    });

    it("renders Admin link for superadmin users", async () => {
      localStorage.setItem("token", "valid-token");
      vi.mocked(api.me).mockResolvedValue({
        id: 3,
        email: "admin@example.com",
        name: "Super Boss",
        role: "superadmin",
        status: "active",
      });

      render(
        <MemoryRouter>
          <AuthProvider>
            <Navbar />
          </AuthProvider>
        </MemoryRouter>
      );

      await waitFor(() => {
        expect(screen.getByRole("navigation")).toBeInTheDocument();
        expect(screen.getByText("Admin")).toBeInTheDocument();
        expect(screen.getByText("Super Boss (superadmin)")).toBeInTheDocument();
      });
    });

    it("triggers logout when the logout button is clicked", async () => {
      localStorage.setItem("token", "valid-token");
      vi.mocked(api.me).mockResolvedValue({
        id: 1,
        email: "user@example.com",
        name: "Jane Doe",
        role: "user",
        status: "active",
      });

      render(
        <MemoryRouter>
          <AuthProvider>
            <Navbar />
          </AuthProvider>
        </MemoryRouter>
      );

      await waitFor(() => {
        expect(screen.getByRole("button", { name: "Logout" })).toBeInTheDocument();
      });

      fireEvent.click(screen.getByRole("button", { name: "Logout" }));

      expect(localStorage.getItem("token")).toBeNull();
    });
  });
});
