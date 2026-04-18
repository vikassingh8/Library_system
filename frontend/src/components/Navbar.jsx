import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

export default function Navbar() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  const handleLogout = async () => {
    await logout();
  };

  return (
    <nav className="bg-linear-to-r from-indigo-900 via-purple-900 to-indigo-900 border-b border-white/10 shadow-lg">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-16">
          <Link to="/" className="flex items-center gap-3">
            <span className="text-2xl">📚</span>
            <span className="text-xl font-bold bg-linear-to-r from-blue-300 to-purple-300 bg-clip-text text-transparent">
              LibraryHub
            </span>
          </Link>

          {user ? (
            <div className="flex items-center gap-4">
              <Link to="/" className="text-white/80 hover:text-white transition-colors text-sm font-medium px-3 py-2 rounded-lg hover:bg-white/10">
                Books
              </Link>
              <Link to="/my-borrows" className="text-white/80 hover:text-white transition-colors text-sm font-medium px-3 py-2 rounded-lg hover:bg-white/10">
                My Borrows
              </Link>
              {user.role === 'admin' && (
                <Link to="/add-book" className="text-white/80 hover:text-white transition-colors text-sm font-medium px-3 py-2 rounded-lg hover:bg-white/10">
                  + Add Book
                </Link>
              )}
              <div className="h-5 w-px bg-white/20"></div>
              <span className="text-white/60 text-sm">{user.email}</span>
              <button onClick={handleLogout}
                className="text-sm bg-white/10 hover:bg-white/20 text-white px-4 py-2 rounded-lg transition-all cursor-pointer">
                Logout
              </button>
            </div>
          ) : (
            <div className="flex items-center gap-3">
              <Link to="/login" className="text-white/80 hover:text-white text-sm font-medium px-4 py-2 rounded-lg hover:bg-white/10 transition-all">
                Login
              </Link>
              <Link to="/register" className="text-sm bg-indigo-500 hover:bg-indigo-400 text-white px-4 py-2 rounded-lg transition-all shadow-md">
                Register
              </Link>
            </div>
          )}
        </div>
      </div>
    </nav>
  );
}
