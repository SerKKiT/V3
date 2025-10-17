import client from './client';
import { ENDPOINTS } from '../utils/constants';

export const streamsAPI = {
  createStream: async (title, description) => {
    const response = await client.post(ENDPOINTS.STREAMS, {
      title,
      description,
    });
    return response.data;
  },

  getLiveStreams: async () => {
    const response = await client.get(`${ENDPOINTS.STREAMS}/live`);
    return response.data;
  },

  getStream: async (id) => {
    const response = await client.get(`${ENDPOINTS.STREAMS}/${id}`);
    return response.data;
  },

  // ← ДОБАВЬТЕ этот метод
  getUserStreams: async () => {
    const response = await client.get(ENDPOINTS.STREAMS);
    return response.data;
  },

  updateStream: async (id, data) => {
    const response = await client.put(`${ENDPOINTS.STREAMS}/${id}`, data);
    return response.data;
  },

  deleteStream: async (id) => {
    const response = await client.delete(`${ENDPOINTS.STREAMS}/${id}`);
    return response.data;
  },
};
