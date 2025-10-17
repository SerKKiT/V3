import React, { createContext, useState, useEffect } from 'react';
import { authAPI } from '../api/auth';
import { storage } from '../utils/storage';

export const AuthContext = createContext();

export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null);
  const [token, setToken] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const currentUser = authAPI.getCurrentUser();
    const currentToken = storage.getToken();
    
    setUser(currentUser);
    setToken(currentToken);
    setLoading(false);
  }, []);

  const login = async (username, password) => {
    console.log('🔐 AuthContext: Starting login...');
    const data = await authAPI.login(username, password);
    
    // ✅ Сохраняем в localStorage
    // (уже делается в authAPI.login)
    
    // ✅ Сохраняем в cookie для video element
    document.cookie = `auth_token=${data.token}; path=/; max-age=86400; SameSite=Lax`;
    
    console.log('✅ AuthContext: Login successful');
    console.log('👤 User:', data.user.username);
    console.log('🔑 Token saved: header & cookie');
    
    setUser(data.user);
    setToken(data.token);
    
    return data;
  };

  const register = async (username, email, password) => {
    const data = await authAPI.register(username, email, password);
    
    // ✅ Сохраняем в cookie
    document.cookie = `auth_token=${data.token}; path=/; max-age=86400; SameSite=Lax`;
    
    setUser(data.user);
    setToken(data.token);
    
    return data;
  };

  const logout = () => {
    authAPI.logout();
    
    // ✅ Удаляем cookie
    document.cookie = 'auth_token=; path=/; max-age=0';
    
    setUser(null);
    setToken(null);
  };

  const value = {
    user,
    token,
    loading,
    isAuthenticated: !!user,
    login,
    register,
    logout,
  };

  return (
    <AuthContext.Provider value={value}>
      {!loading && children}
    </AuthContext.Provider>
  );
};
