import { createContext, useContext, useState } from 'react';
import * as api from '../services/api';

const AuthContext = createContext(null);

export function AuthProvider({ children }) {
  const [user, setUser] = useState(() => {
    const saved = localStorage.getItem('user');
    return saved ? JSON.parse(saved) : null;
    // Only stores { name, email, role } — never the token
    // The JWT lives in the HTTP-only cookie, browser sends it automatically
  });

  const handleLogin = async (email, password) => {
    const data = await api.login(email, password);

    // Store only non-sensitive UI info — NOT the token
    const userData = { name: data.name, email: data.email, role: data.role };
    setUser(userData);
    localStorage.setItem('user', JSON.stringify(userData));
    return data;
  };

  const handleRegister = async (name, email, password) => {
    return await api.register(name, email, password);
  };

  const handleLogout = async () => {
    try {
      await api.logout();
    } catch (err) {
      console.error('Logout error:', err);
    }
    setUser(null);
    localStorage.removeItem('user');
  };

  return (
    <AuthContext.Provider value={{ user, login: handleLogin, register: handleRegister, logout: handleLogout }}>
      {children}
    </AuthContext.Provider>
  );
}

export const useAuth = () => useContext(AuthContext);
