import { useState, useEffect } from 'react';
import { useAuth } from '../context/AuthContext';
import * as api from '../services/api';

export default function Books() {
  const { user } = useAuth();
  const isAdmin = user?.role === 'admin';

  const [books, setBooks] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [search, setSearch] = useState('');
  const [searchType, setSearchType] = useState('author'); // 'author' | 'title' | 'genre'
  const [borrowMsg, setBorrowMsg] = useState('');

  const fetchBooks = async () => {
    setLoading(true);
    try {
      const data = await api.getBooks();
      setBooks(data.books || []);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { fetchBooks(); }, []);

  const handleSearch = async () => {
    setLoading(true);
    try {
      const genre  = searchType === 'genre'  ? search : '';
      const author = searchType === 'author' ? search : '';
      const title  = searchType === 'title'  ? search : '';
      const data = await api.searchBooks(genre, author, title);
      setBooks(data.books || []);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (bookId) => {
    if (!window.confirm('Delete this book from the catalog?')) return;
    try {
      await api.deleteBook(bookId);
      setBorrowMsg('✅ Book deleted.');
      fetchBooks();
      setTimeout(() => setBorrowMsg(''), 3000);
    } catch (err) {
      setBorrowMsg(`❌ ${err.message}`);
      setTimeout(() => setBorrowMsg(''), 3000);
    }
  };

  const handleBorrow = async (bookId) => {
    try {
      const data = await api.borrowBook(bookId);
      setBorrowMsg(`✅ Book borrowed! Due: ${data.due_date}`);
      fetchBooks();
      setTimeout(() => setBorrowMsg(''), 4000);
    } catch (err) {
      setBorrowMsg(`❌ ${err.message}`);
      setTimeout(() => setBorrowMsg(''), 4000);
    }
  };

  const genreColors = {
    Programming: 'bg-blue-500/20 text-blue-300',
    Fiction: 'bg-purple-500/20 text-purple-300',
    Science: 'bg-green-500/20 text-green-300',
    History: 'bg-amber-500/20 text-amber-300',
    default: 'bg-gray-500/20 text-gray-300',
  };

  return (
    <div className="min-h-[calc(100vh-4rem)] bg-linear-to-br from-gray-950 via-indigo-950 to-gray-950 px-4 py-8">
      <div className="max-w-7xl mx-auto">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-white">Book Catalog</h1>
          <p className="text-gray-400 mt-1">Browse and borrow from our collection</p>
        </div>

        {/* Search */}
        <div className="flex gap-3 mb-8">
          <select value={searchType} onChange={(e) => setSearchType(e.target.value)}
            className="px-3 py-3 bg-white/5 border border-white/10 rounded-xl text-white focus:outline-none focus:ring-2 focus:ring-indigo-500 transition-all">
            <option value="author" className="bg-gray-900">Author</option>
            <option value="title"  className="bg-gray-900">Title</option>
            <option value="genre"  className="bg-gray-900">Genre</option>
          </select>
          <input value={search} onChange={(e) => setSearch(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
            className="flex-1 max-w-md px-4 py-3 bg-white/5 border border-white/10 rounded-xl text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 transition-all"
            placeholder={`Search by ${searchType}...`} />
          <button onClick={handleSearch}
            className="px-6 py-3 bg-indigo-600 hover:bg-indigo-500 text-white rounded-xl transition-all font-medium cursor-pointer">
            Search
          </button>
          <button onClick={fetchBooks}
            className="px-6 py-3 bg-white/10 hover:bg-white/20 text-white rounded-xl transition-all font-medium cursor-pointer">
            Reset
          </button>
        </div>

        {/* Messages */}
        {borrowMsg && (
          <div className={`mb-6 px-4 py-3 rounded-xl text-sm font-medium ${borrowMsg.startsWith('✅') ? 'bg-green-500/10 border border-green-500/30 text-green-300' : 'bg-red-500/10 border border-red-500/30 text-red-300'}`}>
            {borrowMsg}
          </div>
        )}

        {error && <div className="mb-6 bg-red-500/10 border border-red-500/30 text-red-300 text-sm rounded-xl px-4 py-3">{error}</div>}

        {/* Loading */}
        {loading ? (
          <div className="flex justify-center py-20">
            <div className="animate-spin rounded-full h-10 w-10 border-t-2 border-b-2 border-indigo-400"></div>
          </div>
        ) : books.length === 0 ? (
          <div className="text-center py-20">
            <span className="text-6xl">📭</span>
            <p className="text-gray-400 mt-4 text-lg">No books found</p>
            <p className="text-gray-500 text-sm mt-1">Add some books to get started!</p>
          </div>
        ) : (
          /* Book Grid */
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {books.map((book) => (
              <div key={book.id} className="bg-white/5 backdrop-blur-sm border border-white/10 rounded-2xl p-6 hover:bg-white/8 hover:border-white/20 transition-all group">

                {/* Book Cover – real image from Azure Blob or emoji fallback */}
                <div className="h-56 rounded-xl overflow-hidden mb-4 group-hover:opacity-90 transition-all bg-gradient-to-br from-indigo-600/20 to-purple-600/20 flex items-center justify-center">
                  {book.cover_image_url ? (
                    <img
                      src={book.cover_image_url}
                      alt={`${book.title} cover`}
                      className="w-full h-full object-contain"
                      onError={(e) => {
                        e.currentTarget.style.display = 'none';
                      }}
                    />
                  ) : (
                    <span className="text-5xl">📖</span>
                  )}
                </div>

                {/* Genre Badge */}
                {book.genre && (
                  <span className={`inline-block px-3 py-1 rounded-full text-xs font-medium mb-3 ${genreColors[book.genre] || genreColors.default}`}>
                    {book.genre}
                  </span>
                )}

                {/* Title & Author */}
                <h3 className="text-lg font-semibold text-white leading-tight">{book.title}</h3>
                <p className="text-gray-400 text-sm mt-1">by {book.author}</p>
                <p className="text-gray-600 text-xs mt-1">ISBN: {book.isbn}</p>

                {/* Availability + Actions */}
                <div className="flex items-center justify-between mt-4 pt-4 border-t border-white/10">
                  <div>
                    <span className={`text-sm font-medium ${book.available_copies > 0 ? 'text-green-400' : 'text-red-400'}`}>
                      {book.available_copies > 0 ? `${book.available_copies} available` : 'Not available'}
                    </span>
                    <span className="text-gray-600 text-xs ml-2">/ {book.total_copies} total</span>
                  </div>

                  <div className="flex items-center gap-2">
                    {isAdmin && (
                      <button onClick={() => handleDelete(book.id)}
                        className="px-3 py-2 bg-red-600/20 hover:bg-red-600 border border-red-500/30 text-red-400 hover:text-white text-sm rounded-lg transition-all font-medium cursor-pointer">
                        Delete
                      </button>
                    )}
                    {!isAdmin && (
                      <button onClick={() => handleBorrow(book.id)} disabled={book.available_copies <= 0}
                        className="px-4 py-2 bg-indigo-600 hover:bg-indigo-500 disabled:bg-gray-700 disabled:text-gray-500 text-white text-sm rounded-lg transition-all font-medium cursor-pointer disabled:cursor-not-allowed">
                        {book.available_copies > 0 ? 'Borrow' : 'Unavailable'}
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
