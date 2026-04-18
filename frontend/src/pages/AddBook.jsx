import { useState, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import * as api from '../services/api';

export default function AddBook() {
  const navigate = useNavigate();
  const fileInputRef = useRef(null);

  const [form, setForm] = useState({ title: '', author: '', isbn: '', genre: '', total_copies: 1 });
  const [imageFile, setImageFile]   = useState(null);
  const [imagePreview, setImagePreview] = useState('');
  const [error, setError]   = useState('');
  const [loading, setLoading] = useState(false);
  const [uploadProgress, setUploadProgress] = useState(''); // status text

  const genres = ['Programming', 'Fiction', 'Science', 'History', 'Philosophy', 'Biography', 'Technology', 'Other'];

  const handleChange = (e) => {
    const val = e.target.name === 'total_copies' ? parseInt(e.target.value) || 1 : e.target.value;
    setForm({ ...form, [e.target.name]: val });
  };

  const handleImageChange = (file) => {
    if (!file) return;
    const allowed = ['image/jpeg', 'image/png', 'image/gif', 'image/webp'];
    if (!allowed.includes(file.type)) {
      setError('Only JPG, PNG, GIF, WEBP images are allowed.');
      return;
    }
    if (file.size > 10 * 1024 * 1024) {
      setError('Image must be smaller than 10 MB.');
      return;
    }
    setError('');
    setImageFile(file);
    setImagePreview(URL.createObjectURL(file));
  };

  const handleDrop = (e) => {
    e.preventDefault();
    const file = e.dataTransfer.files?.[0];
    handleImageChange(file);
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      let coverImageURL = '';

      // 1. Upload image first (if selected)
      if (imageFile) {
        setUploadProgress('📤 Uploading cover image...');
        const result = await api.uploadImage(imageFile);
        coverImageURL = result.url;
        setUploadProgress('✅ Image uploaded!');
      }

      // 2. Create the book with the blob URL
      await api.createBook({ ...form, cover_image_url: coverImageURL });
      navigate('/');
    } catch (err) {
      setError(err.message);
      setUploadProgress('');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-[calc(100vh-4rem)] bg-linear-to-br from-gray-950 via-indigo-950 to-gray-950 px-4 py-8">
      <div className="max-w-2xl mx-auto">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-white">Add New Book</h1>
          <p className="text-gray-400 mt-1">Add a book to the library catalog</p>
        </div>

        <form onSubmit={handleSubmit} className="bg-white/5 backdrop-blur-xl rounded-2xl shadow-2xl border border-white/10 p-8 space-y-5">
          {error && (
            <div className="bg-red-500/10 border border-red-500/30 text-red-300 text-sm rounded-lg px-4 py-3">{error}</div>
          )}
          {uploadProgress && (
            <div className="bg-indigo-500/10 border border-indigo-500/30 text-indigo-300 text-sm rounded-lg px-4 py-3">{uploadProgress}</div>
          )}

          {/* Cover Image Upload */}
          <div>
            <label className="block text-sm font-medium text-gray-300 mb-1.5">Cover Image (optional)</label>
            <div
              id="image-drop-zone"
              onDrop={handleDrop}
              onDragOver={(e) => e.preventDefault()}
              onClick={() => fileInputRef.current?.click()}
              className="relative cursor-pointer border-2 border-dashed border-white/20 rounded-xl overflow-hidden hover:border-indigo-500/60 transition-all group"
              style={{ minHeight: '180px' }}
            >
              {imagePreview ? (
                <div className="relative">
                  <img
                    src={imagePreview}
                    alt="Cover preview"
                    className="w-full h-48 object-cover"
                  />
                  <div className="absolute inset-0 bg-black/50 opacity-0 group-hover:opacity-100 transition-all flex items-center justify-center">
                    <span className="text-white text-sm font-medium">Click to change</span>
                  </div>
                </div>
              ) : (
                <div className="flex flex-col items-center justify-center h-44 text-gray-500 group-hover:text-gray-300 transition-all">
                  <span className="text-4xl mb-2">🖼️</span>
                  <p className="text-sm font-medium">Drop image here or click to browse</p>
                  <p className="text-xs mt-1">JPG, PNG, GIF, WEBP — max 10 MB</p>
                </div>
              )}
            </div>
            <input
              id="image-file-input"
              ref={fileInputRef}
              type="file"
              accept="image/jpeg,image/png,image/gif,image/webp"
              className="hidden"
              onChange={(e) => handleImageChange(e.target.files?.[0])}
            />
            {imageFile && (
              <div className="flex items-center justify-between mt-2 px-1">
                <span className="text-xs text-gray-400 truncate">{imageFile.name}</span>
                <button
                  type="button"
                  onClick={() => { setImageFile(null); setImagePreview(''); }}
                  className="text-xs text-red-400 hover:text-red-300 ml-3 shrink-0"
                >
                  Remove
                </button>
              </div>
            )}
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-300 mb-1.5">Title *</label>
            <input name="title" value={form.title} onChange={handleChange} required
              className="w-full px-4 py-3 bg-white/5 border border-white/10 rounded-xl text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 transition-all"
              placeholder="The Go Programming Language" />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-300 mb-1.5">Author *</label>
            <input name="author" value={form.author} onChange={handleChange} required
              className="w-full px-4 py-3 bg-white/5 border border-white/10 rounded-xl text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 transition-all"
              placeholder="Alan Donovan" />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-300 mb-1.5">ISBN *</label>
              <input name="isbn" value={form.isbn} onChange={handleChange} required
                className="w-full px-4 py-3 bg-white/5 border border-white/10 rounded-xl text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 transition-all"
                placeholder="9780134190440" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-300 mb-1.5">Copies *</label>
              <input name="total_copies" type="number" min="1" value={form.total_copies} onChange={handleChange} required
                className="w-full px-4 py-3 bg-white/5 border border-white/10 rounded-xl text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 transition-all" />
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-300 mb-1.5">Genre</label>
            <select name="genre" value={form.genre} onChange={handleChange}
              className="w-full px-4 py-3 bg-white/5 border border-white/10 rounded-xl text-white focus:outline-none focus:ring-2 focus:ring-indigo-500 transition-all">
              <option value="" className="bg-gray-900">Select genre</option>
              {genres.map((g) => <option key={g} value={g} className="bg-gray-900">{g}</option>)}
            </select>
          </div>

          <div className="flex gap-3 pt-2">
            <button type="submit" disabled={loading}
              className="flex-1 py-3 bg-linear-to-r from-indigo-600 to-purple-600 hover:from-indigo-500 hover:to-purple-500 text-white font-semibold rounded-xl shadow-lg transition-all disabled:opacity-50 cursor-pointer">
              {loading ? (imageFile ? 'Uploading...' : 'Adding...') : 'Add Book'}
            </button>
            <button type="button" onClick={() => navigate('/')}
              className="px-6 py-3 bg-white/10 hover:bg-white/20 text-white rounded-xl transition-all font-medium cursor-pointer">
              Cancel
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
