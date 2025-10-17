import axios from 'axios';
import { storage } from '../utils/storage';
import { API_BASE_URL } from '../utils/constants';

const client = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000,
  // ✅ Следовать редиректам (для presigned URLs)
  maxRedirects: 5,
});

// Добавляем JWT токен ко всем запросам
client.interceptors.request.use(
  (config) => {
    const token = storage.getToken();
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Обработка ошибок
client.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Токен истек или невалиден
      console.log('⚠️ Unauthorized, clearing session');
      storage.clear();
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

export default client;
