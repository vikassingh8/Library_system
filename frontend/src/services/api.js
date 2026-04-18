const BASE = import.meta.env.VITE_API_BASE_URL || 'https://librarysystem-b7ayetc0hwepgyag.centralindia-01.azurewebsites.net';

async function request(url, options = {}) {
  const config = {
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    ...options,
  };

  const res = await fetch(`${BASE}${url}`, config);
  const data = await res.json().catch(() => null);

  if (res.status === 401) {
    // Session expired or invalid — clear local storage and redirect to login
    localStorage.removeItem('user');
    window.location.href = '/login';

    return;
  }

  if (!res.ok) {
    const msg = data?.error || `Request failed (${res.status})`;
    throw new Error(msg);

  }

  return data;
}

// Auth
export const register = (name, email, password) =>
  request('/register', { method: 'POST', body: JSON.stringify({ name, email, password }) });

export const login = (email, password) =>
  request('/login', { method: 'POST', credentials: "include", body: JSON.stringify({ email, password }) });

export const logout = () =>
  request('/logout', { method: 'POST' });

// Image Upload – multipart/form-data (no Content-Type header; browser sets boundary)
export const uploadImage = async (file) => {
  const formData = new FormData();
  formData.append('image', file);
  const res = await fetch(`${BASE}/upload-image`, {
    method: 'POST',
    credentials: 'include',
    body: formData,
  });
  const data = await res.json().catch(() => null);
  if (res.status === 401) {
    localStorage.removeItem('user');
    window.location.href = '/login';
    throw new Error('Session expired. Please login again.');
  }
  if (!res.ok) {
    const msg = data?.error || `Upload failed (${res.status})`;
    throw new Error(msg);
  }
  return data; // { url: "https://..." }
};

// Books
export const getBooks = () => request('/books');
export const getBook = (id) => request(`/books/${id}`);
export const searchBooks = (genre, author, title) => {
  const params = new URLSearchParams();
  if (genre) params.set('genre', genre);
  if (author) params.set('author', author);
  if (title) params.set('title', title);
  return request(`/books/search?${params}`);
};
export const createBook = (book) =>
  request('/books', { method: 'POST', credentials: "include", body: JSON.stringify(book) });
export const updateBook = (id, book) =>
  request(`/books/${id}`, { method: 'PUT', body: JSON.stringify(book) });
export const deleteBook = (id) =>
  request(`/books/${id}`, { method: 'DELETE' });

// Borrows
export const borrowBook = (bookId) =>
  request(`/books/${bookId}/borrow`, { method: 'POST' });
export const returnBook = (borrowId) =>
  request(`/borrows/${borrowId}/return`, { method: 'POST' });
export const getMyBorrows = () => request('/my-borrows');
