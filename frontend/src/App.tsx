import { BrowserRouter, Routes, Route } from "react-router-dom";
import { AuthProvider } from "./context/AuthContext";
import { ProtectedRoute } from "./components/ProtectedRoute";
import { Navbar } from "./components/Navbar";
import { LoginPage } from "./pages/LoginPage";
import { RegisterPage } from "./pages/RegisterPage";
import { DashboardPage } from "./pages/DashboardPage";
import { AdminPage } from "./pages/AdminPage";
import { WorkspaceListPage } from "./pages/WorkspaceListPage";
import { WorkspaceDetailPage } from "./pages/WorkspaceDetailPage";
import { NoteEditorPage } from "./pages/NoteEditorPage";
import { TemplateDiscoveryPage } from "./pages/TemplateDiscoveryPage";
import { TemplateDetailPage } from "./pages/TemplateDetailPage";
import { MyTemplatesPage } from "./pages/MyTemplatesPage";
import { NotFoundPage } from "./pages/NotFoundPage";

function Layout({ children }: { children: React.ReactNode }) {
  return (
    <div className="appShell">
      <Navbar />
      <main className="appMain">{children}</main>
    </div>
  );
}

export function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/register" element={<RegisterPage />} />
          <Route
            path="/dashboard"
            element={
              <ProtectedRoute>
                <Layout>
                  <DashboardPage />
                </Layout>
              </ProtectedRoute>
            }
          />
          <Route
            path="/workspaces"
            element={
              <ProtectedRoute>
                <Layout>
                  <WorkspaceListPage />
                </Layout>
              </ProtectedRoute>
            }
          />
          <Route
            path="/workspaces/:workspaceId"
            element={
              <ProtectedRoute>
                <Layout>
                  <WorkspaceDetailPage />
                </Layout>
              </ProtectedRoute>
            }
          />
          <Route
            path="/workspaces/:workspaceId/notes/new"
            element={
              <ProtectedRoute>
                <Layout>
                  <NoteEditorPage />
                </Layout>
              </ProtectedRoute>
            }
          />
          <Route
            path="/workspaces/:workspaceId/notes/:noteId"
            element={
              <ProtectedRoute>
                <Layout>
                  <NoteEditorPage />
                </Layout>
              </ProtectedRoute>
            }
          />
          <Route
            path="/templates"
            element={
              <ProtectedRoute>
                <Layout>
                  <TemplateDiscoveryPage />
                </Layout>
              </ProtectedRoute>
            }
          />
          <Route
            path="/templates/:id"
            element={
              <ProtectedRoute>
                <Layout>
                  <TemplateDetailPage />
                </Layout>
              </ProtectedRoute>
            }
          />
          <Route
            path="/templates/my"
            element={
              <ProtectedRoute roles={["creator"]}>
                <Layout>
                  <MyTemplatesPage />
                </Layout>
              </ProtectedRoute>
            }
          />
          <Route
            path="/admin"
            element={
              <ProtectedRoute roles={["superadmin"]}>
                <Layout>
                  <AdminPage />
                </Layout>
              </ProtectedRoute>
            }
          />
          <Route path="/" element={<LoginPage />} />
          <Route path="*" element={<NotFoundPage />} />
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  );
}
