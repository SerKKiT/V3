const TOKEN_KEY = 'auth_token';
const USER_KEY = 'user_data';

export const storage = {
  // Token methods
  getToken: () => localStorage.getItem(TOKEN_KEY),
  
  setToken: (token) => localStorage.setItem(TOKEN_KEY, token),
  
  removeToken: () => localStorage.removeItem(TOKEN_KEY),
  
  // User methods
  getUser: () => {
    const user = localStorage.getItem(USER_KEY);
    return user ? JSON.parse(user) : null;
  },
  
  setUser: (user) => localStorage.setItem(USER_KEY, JSON.stringify(user)),
  
  removeUser: () => localStorage.removeItem(USER_KEY),
  
  // Clear all
  clear: () => {
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(USER_KEY);
  },
};
