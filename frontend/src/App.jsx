import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider, useAuth } from './context/AuthContext';
import Navbar from './components/Navbar';
import Login from './pages/Login';
import Register from './pages/Register';
import Books from './pages/Books';
import MyBorrows from './pages/MyBorrows';
import AddBook from './pages/AddBook';

function ProtectedRoute({ children }) {
  const { user } = useAuth();
  return user ? children : <Navigate to="/login" />;
}

function PublicRoute({ children }) {
  const { user } = useAuth();
  return !user ? children : <Navigate to="/" />;
}

function AppRoutes() {
  return (
    <Routes>
      <Route path="/login" element={<PublicRoute><Login /></PublicRoute>} />
      <Route path="/register" element={<PublicRoute><Register /></PublicRoute>} />
      <Route path="/" element={<ProtectedRoute><Books /></ProtectedRoute>} />
      <Route path="/my-borrows" element={<ProtectedRoute><MyBorrows /></ProtectedRoute>} />
      <Route path="/add-book" element={<ProtectedRoute><AddBook /></ProtectedRoute>} />
    </Routes>
  );
}

export default function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <div className="min-h-screen bg-gray-950">
          <Navbar />
          <AppRoutes />
        </div>
      </AuthProvider>
    </BrowserRouter>
  );
}