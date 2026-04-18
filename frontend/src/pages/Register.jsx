import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

export default function Register() {
  const { register } = useAuth();
  const navigate = useNavigate();
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      await register(name, email, password);
      navigate('/login');
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-[calc(100vh-4rem)] flex items-center justify-center bg-linear-to-br from-gray-950 via-indigo-950 to-gray-950 px-4">
      <div className="w-full max-w-md">
        <div className="text-center mb-8">
          <span className="text-5xl">✨</span>
          <h1 className="mt-4 text-3xl font-bold text-white">Create Account</h1>
          <p className="mt-2 text-gray-400">Join the library community</p>
        </div>

        <form onSubmit={handleSubmit} className="bg-white/5 backdrop-blur-xl rounded-2xl shadow-2xl border border-white/10 p-8 space-y-5">
          {error && (
            <div className="bg-red-500/10 border border-red-500/30 text-red-300 text-sm rounded-lg px-4 py-3">{error}</div>
          )}

          <div>
            <label className="block text-sm font-medium text-gray-300 mb-1.5">Full Name</label>
            <input type="text" value={name} onChange={(e) => setName(e.target.value)} required
              className="w-full px-4 py-3 bg-white/5 border border-white/10 rounded-xl text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent transition-all"
              placeholder="John Doe" />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-300 mb-1.5">Email</label>
            <input type="email" value={email} onChange={(e) => setEmail(e.target.value)} required
              className="w-full px-4 py-3 bg-white/5 border border-white/10 rounded-xl text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent transition-all"
              placeholder="you@example.com" />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-300 mb-1.5">Password</label>
            <input type="password" value={password} onChange={(e) => setPassword(e.target.value)} required minLength={8}
              className="w-full px-4 py-3 bg-white/5 border border-white/10 rounded-xl text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent transition-all"
              placeholder="Minimum 8 characters" />
          </div>

          <button type="submit" disabled={loading}
            className="w-full py-3 bg-linear-to-r from-indigo-600 to-purple-600 hover:from-indigo-500 hover:to-purple-500 text-white font-semibold rounded-xl shadow-lg hover:shadow-indigo-500/25 transition-all disabled:opacity-50 cursor-pointer">
            {loading ? 'Creating account...' : 'Create Account'}
          </button>

          <p className="text-center text-gray-400 text-sm">
            Already have an account?{' '}
            <Link to="/login" className="text-indigo-400 hover:text-indigo-300 font-medium">Sign In</Link>
          </p>
        </form>
      </div>
    </div>
  );
}
