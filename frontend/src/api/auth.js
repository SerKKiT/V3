const API_URL = 'http://localhost/api';
import { storage } from '../utils/storage';

export const authAPI = {
  // Login
  login: async (username, password) => {
    const response = await fetch(`${API_URL}/auth/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ username, password }),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Login failed');
    }

    const data = await response.json();
    
    storage.setToken(data.token);
    storage.setUser(data.user);
    
    return data;
  },

  // Register
  register: async (username, email, password) => {
    const response = await fetch(`${API_URL}/auth/register`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ username, email, password }),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Registration failed');
    }

    const data = await response.json();
    
    storage.setToken(data.token);
    storage.setUser(data.user);
    
    return data;
  },

  // ✅ НОВОЕ: Получить профиль текущего пользователя
  getProfile: async () => {
    const token = storage.getToken();
    if (!token) {
      throw new Error('Not authenticated');
    }

    const response = await fetch(`${API_URL}/auth/profile`, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
    });

    if (!response.ok) {
      throw new Error('Failed to fetch profile');
    }

    return await response.json();
  },

  // ✅ НОВОЕ: Получить информацию о пользователе по ID (публичный endpoint)
  getUserInfo: async (userId) => {
    try {
      const response = await fetch(`${API_URL}/auth/users/${userId}`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        console.warn(`User info not found for ID: ${userId}`);
        return null;
      }

      const data = await response.json();
      return data;
    } catch (error) {
      console.error('Failed to fetch user info:', error);
      return null;
    }
  },

  // ✅ НОВОЕ: Обновить профиль
  updateProfile: async (profileData) => {
    const token = storage.getToken();
    if (!token) {
      throw new Error('Not authenticated');
    }

    const response = await fetch(`${API_URL}/auth/profile`, {
      method: 'PUT',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(profileData),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Failed to update profile');
    }

    return await response.json();
  },

  // ✅ НОВОЕ: Изменить пароль
  changePassword: async (currentPassword, newPassword) => {
    const token = storage.getToken();
    if (!token) {
      throw new Error('Not authenticated');
    }

    const response = await fetch(`${API_URL}/auth/change-password`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        current_password: currentPassword,
        new_password: newPassword,
      }),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Failed to change password');
    }

    return await response.json();
  },

  // Logout
  logout: () => {
    storage.clear();
  },

  // Get current user from localStorage
  getCurrentUser: () => {
    return storage.getUser();
  },

  // Check if user is authenticated
  isAuthenticated: () => {
    return !!storage.getToken();
  },
};
