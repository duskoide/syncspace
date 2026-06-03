import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { LoginPage } from "../pages/LoginPage";
import { RegisterPage } from "../pages/RegisterPage";
import { AuthProvider } from "../context/AuthContext";
import { api } from "../services/api";

const mockNavigate = vi.fn();

vi.mock("react-router-dom", async () => {
  const actual = (await vi.importActual("react-router-dom")) as any;
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

vi.mock("../services/api", () => {
  return {
    api: {
      login: vi.fn(),
      register: vi.fn(),
      me: vi.fn(),
    },
  };
});

// Helper to find input/select fields associated with a label text when htmlFor/id are missing
function getFieldByLabel(labelText: string): HTMLElement {
  const label = screen.getByText(labelText);
  const control = label.nextElementSibling || label.parentElement?.querySelector("input, select");
  if (!control) {
    throw new Error(`Could not find control for label: ${labelText}`);
  }
  return control as HTMLElement;
}

describe("Authentication Pages", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    localStorage.clear();
  });

  describe("LoginPage", () => {
    it("renders email and password inputs and the submit button", () => {
      render(
        <MemoryRouter>
          <AuthProvider>
            <LoginPage />
          </AuthProvider>
        </MemoryRouter>
      );

      expect(screen.getByText("SyncSpace")).toBeInTheDocument();
      expect(screen.getByText("Email")).toBeInTheDocument();
      expect(screen.getByText("Password")).toBeInTheDocument();
      expect(screen.getByRole("button", { name: "Sign In" })).toBeInTheDocument();
    });

    it("logs in successfully and navigates to the dashboard", async () => {
      const mockLoginResponse = {
        token: { access_token: "test-jwt-token" },
        user: { id: 1, email: "test@example.com", name: "Test User", role: "user", status: "active" },
      };
      vi.mocked(api.login).mockResolvedValue(mockLoginResponse);

      render(
        <MemoryRouter>
          <AuthProvider>
            <LoginPage />
          </AuthProvider>
        </MemoryRouter>
      );

      fireEvent.change(getFieldByLabel("Email"), { target: { value: "test@example.com" } });
      fireEvent.change(getFieldByLabel("Password"), { target: { value: "password123" } });
      fireEvent.click(screen.getByRole("button", { name: "Sign In" }));

      await waitFor(() => {
        expect(api.login).toHaveBeenCalledWith("test@example.com", "password123");
        expect(localStorage.getItem("token")).toBe("test-jwt-token");
        expect(mockNavigate).toHaveBeenCalledWith("/dashboard");
      });
    });

    it("displays error message on login failure", async () => {
      vi.mocked(api.login).mockRejectedValue(new Error("Invalid credentials"));

      render(
        <MemoryRouter>
          <AuthProvider>
            <LoginPage />
          </AuthProvider>
        </MemoryRouter>
      );

      fireEvent.change(getFieldByLabel("Email"), { target: { value: "wrong@example.com" } });
      fireEvent.change(getFieldByLabel("Password"), { target: { value: "wrongpass" } });
      fireEvent.click(screen.getByRole("button", { name: "Sign In" }));

      await waitFor(() => {
        expect(api.login).toHaveBeenCalledWith("wrong@example.com", "wrongpass");
        expect(screen.getByText("Invalid credentials")).toBeInTheDocument();
      });
    });
  });

  describe("RegisterPage", () => {
    it("renders name, email, password, and role selector inputs", () => {
      render(
        <MemoryRouter>
          <RegisterPage />
        </MemoryRouter>
      );

      expect(screen.getByText("Full Name")).toBeInTheDocument();
      expect(screen.getByText("Email")).toBeInTheDocument();
      expect(screen.getByText("Password")).toBeInTheDocument();
      expect(screen.getByText("Role")).toBeInTheDocument();
      expect(screen.getByRole("button", { name: "Register" })).toBeInTheDocument();
    });

    it("calls register API and displays success message on successful registration", async () => {
      vi.mocked(api.register).mockResolvedValue({ message: "Registration successful" });

      render(
        <MemoryRouter>
          <RegisterPage />
        </MemoryRouter>
      );

      fireEvent.change(getFieldByLabel("Full Name"), { target: { value: "New Creator" } });
      fireEvent.change(getFieldByLabel("Email"), { target: { value: "creator@example.com" } });
      fireEvent.change(getFieldByLabel("Password"), { target: { value: "securepassword" } });
      fireEvent.change(getFieldByLabel("Role"), { target: { value: "creator" } });
      fireEvent.click(screen.getByRole("button", { name: "Register" }));

      await waitFor(() => {
        expect(api.register).toHaveBeenCalledWith({
          name: "New Creator",
          email: "creator@example.com",
          password: "securepassword",
          role: "creator",
        });
        expect(screen.getByText("Account Created")).toBeInTheDocument();
        expect(screen.getByText(/Your account is pending approval/)).toBeInTheDocument();
      });
    });

    it("displays error message on registration failure", async () => {
      vi.mocked(api.register).mockRejectedValue(new Error("Email already exists"));

      render(
        <MemoryRouter>
          <RegisterPage />
        </MemoryRouter>
      );

      fireEvent.change(getFieldByLabel("Full Name"), { target: { value: "Existing User" } });
      fireEvent.change(getFieldByLabel("Email"), { target: { value: "existing@example.com" } });
      fireEvent.change(getFieldByLabel("Password"), { target: { value: "pass12345" } });
      fireEvent.click(screen.getByRole("button", { name: "Register" }));

      await waitFor(() => {
        expect(api.register).toHaveBeenCalledWith({
          name: "Existing User",
          email: "existing@example.com",
          password: "pass12345",
          role: "user",
        });
        expect(screen.getByText("Email already exists")).toBeInTheDocument();
      });
    });
  });
});
