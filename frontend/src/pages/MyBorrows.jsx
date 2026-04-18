import { useState, useEffect } from 'react';
import * as api from '../services/api';

export default function MyBorrows() {
  const [borrows, setBorrows] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [msg, setMsg] = useState('');

  const fetchBorrows = async () => {
    setLoading(true);
    try {
      const data = await api.getMyBorrows();
      setBorrows(data.borrows || []);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { fetchBorrows(); }, []);

  const handleReturn = async (borrowId) => {
    try {
      await api.returnBook(borrowId);
      setMsg('✅ Book returned successfully!');
      fetchBorrows();
      setTimeout(() => setMsg(''), 4000);
    } catch (err) {
      setMsg(`❌ ${err.message}`);
      setTimeout(() => setMsg(''), 4000);
    }
  };

  const formatDate = (dateStr) => {
    if (!dateStr) return '-';
    return new Date(dateStr).toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' });
  };

  const isOverdue = (dueDate, status) => {
    return status === 'borrowed' && new Date(dueDate) < new Date();
  };

  return (
    <div className="min-h-[calc(100vh-4rem)] bg-linear-to-br from-gray-950 via-indigo-950 to-gray-950 px-4 py-8">
      <div className="max-w-5xl mx-auto">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-white">My Borrows</h1>
          <p className="text-gray-400 mt-1">Track your borrowed books and due dates</p>
        </div>

        {msg && (
          <div className={`mb-6 px-4 py-3 rounded-xl text-sm font-medium ${msg.startsWith('✅') ? 'bg-green-500/10 border border-green-500/30 text-green-300' : 'bg-red-500/10 border border-red-500/30 text-red-300'}`}>
            {msg}
          </div>
        )}

        {error && <div className="mb-6 bg-red-500/10 border border-red-500/30 text-red-300 text-sm rounded-xl px-4 py-3">{error}</div>}

        {loading ? (
          <div className="flex justify-center py-20">
            <div className="animate-spin rounded-full h-10 w-10 border-t-2 border-b-2 border-indigo-400"></div>
          </div>
        ) : borrows.length === 0 ? (
          <div className="text-center py-20">
            <span className="text-6xl">📋</span>
            <p className="text-gray-400 mt-4 text-lg">No borrows yet</p>
            <p className="text-gray-500 text-sm mt-1">Browse the catalog and borrow a book!</p>
          </div>
        ) : (
          <div className="space-y-4">
            {borrows.map((borrow) => (
              <div key={borrow.id} className="bg-white/5 backdrop-blur-sm border border-white/10 rounded-2xl p-6 hover:bg-white/8 transition-all">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-4">
                    <div className="w-12 h-12 bg-linear-to-br from-indigo-600/30 to-purple-600/30 rounded-xl flex items-center justify-center">
                      <span className="text-2xl">📖</span>
                    </div>
                    <div>
                      <h3 className="text-white font-semibold">{borrow.book_title}</h3>
                      <p className="text-gray-500 text-xs mt-0.5">by {borrow.book_author}</p>
                    </div>
                  </div>

                  <div className="flex items-center gap-6">
                    {/* Status */}
                    <div className="text-right">
                      <span className={`inline-block px-3 py-1 rounded-full text-xs font-medium ${
                        borrow.status === 'returned' ? 'bg-green-500/20 text-green-300' :
                        isOverdue(borrow.due_date, borrow.status) ? 'bg-red-500/20 text-red-300' :
                        'bg-amber-500/20 text-amber-300'
                      }`}>
                        {borrow.status === 'returned' ? '✅ Returned' :
                         isOverdue(borrow.due_date, borrow.status) ? '🔴 Overdue' :
                         '📘 Borrowed'}
                      </span>
                    </div>

                    {/* Dates */}
                    <div className="text-right min-w-[140px]">
                      <p className="text-gray-400 text-sm">Borrowed: {formatDate(borrow.borrowed_at)}</p>
                      <p className={`text-sm ${isOverdue(borrow.due_date, borrow.status) ? 'text-red-400 font-medium' : 'text-gray-500'}`}>
                        Due: {formatDate(borrow.due_date)}
                      </p>
                      {borrow.returned_at && (
                        <p className="text-green-400 text-sm">Returned: {formatDate(borrow.returned_at)}</p>
                      )}
                    </div>

                    {/* Return Button */}
                    {borrow.status === 'borrowed' && (
                      <button onClick={() => handleReturn(borrow.id)}
                        className="px-5 py-2 bg-emerald-600 hover:bg-emerald-500 text-white text-sm rounded-lg transition-all font-medium cursor-pointer shadow-lg shadow-emerald-600/20">
                        Return
                      </button>
                    )}
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
