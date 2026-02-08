import { Navigate, Route, Routes } from 'react-router-dom';
import type { ReactElement } from 'react';
import LoginPage from './pages/LoginPage';
import AdminDashboard from './pages/AdminDashboard';
import AdminUsersPage from './pages/AdminUsersPage';
import AdminPassesPage from './pages/AdminPassesPage';
import AdminGuestsPage from './pages/AdminGuestsPage';
import GuardDashboard from './pages/GuardDashboard';
import ResidentDashboard from './pages/ResidentDashboard';
import { authStore } from './store/auth';

function RequireAuth({ role, children }: { role?: string; children: ReactElement }) {
  const user = authStore.getUser();
  if (!user) {
    return <Navigate to="/login" replace />;
  }
  if (role && user.role !== role) {
    if (user.role === 'admin') return <Navigate to="/admin" replace />;
    if (user.role === 'guard') return <Navigate to="/guard" replace />;
    return <Navigate to="/resident" replace />;
  }
  return children;
}

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route
        path="/admin"
        element={
          <RequireAuth role="admin">
            <AdminDashboard />
          </RequireAuth>
        }
      />
      <Route
        path="/admin/users"
        element={
          <RequireAuth role="admin">
            <AdminUsersPage />
          </RequireAuth>
        }
      />
      <Route
        path="/admin/passes"
        element={
          <RequireAuth role="admin">
            <AdminPassesPage />
          </RequireAuth>
        }
      />
      <Route
        path="/admin/guests"
        element={
          <RequireAuth role="admin">
            <AdminGuestsPage />
          </RequireAuth>
        }
      />
      <Route
        path="/guard"
        element={
          <RequireAuth role="guard">
            <GuardDashboard />
          </RequireAuth>
        }
      />
      <Route
        path="/resident"
        element={
          <RequireAuth role="resident">
            <ResidentDashboard />
          </RequireAuth>
        }
      />
      <Route path="/" element={<Navigate to="/login" replace />} />
      <Route path="*" element={<Navigate to="/login" replace />} />
    </Routes>
  );
}
