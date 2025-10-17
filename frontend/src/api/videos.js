import client from './client';

const API_URL = 'http://localhost/api';

export const videosAPI = {
  // ✅ Получить все публичные видео + свои приватные
  getAllVideos: async (limit = 20, offset = 0) => {
    const response = await client.get(`${API_URL}/videos`, {
      params: { limit, offset }
    });
    return response.data;
  },

  // Получить список видео пользователя
  getUserVideos: async () => {
    const response = await client.get(`${API_URL}/videos/user`);
    return response.data;
  },

  // Получить видео по ID
  getVideo: async (id) => {
    const response = await client.get(`${API_URL}/videos/${id}`);
    return response.data;
  },

  // Получить URL для стриминга
  getStreamUrl: async (id) => {
    const response = await client.get(`${API_URL}/videos/${id}/stream`);
    return response.data;
  },

  // Импортировать запись в VOD
  importRecording: async (recordingId, title, description, visibility = 'public', tags = [], category = '') => {
    const response = await client.post(`${API_URL}/videos/import-recording`, {
      recording_id: recordingId,
      title,
      description,
      visibility,
      tags,
      category,
    });
    return response.data;
  },

  // Обновить метаданные видео
  updateVideo: async (id, data) => {
    const response = await client.put(`${API_URL}/videos/${id}`, data);
    return response.data;
  },

  // Удалить видео
  deleteVideo: async (id) => {
    const response = await client.delete(`${API_URL}/videos/${id}`);
    return response.data;
  },

  // Увеличить счетчик просмотров
  incrementView: async (id) => {
    const response = await client.post(`${API_URL}/videos/${id}/view`);
    return response.data;
  },

  // Лайкнуть видео
  likeVideo: async (id) => {
    const response = await client.post(`${API_URL}/videos/${id}/like`);
    return response.data;
  },
};
