// API Base URL - пустая строка = относительные пути через Nginx
export const API_BASE_URL = import.meta.env.VITE_API_URL || '';

export const ENDPOINTS = {
  // Auth
  REGISTER: '/api/auth/register',
  LOGIN: '/api/auth/login',
  
  // Streams
  STREAMS: '/api/streams',
  STREAM_BY_ID: (id) => `/api/streams/${id}`,
  LIVE_STREAMS: '/api/streams/live',
  
  // Videos
 // VOD/Videos endpoints (based on main.go)
  USER_VIDEOS: '/api/videos/user',              // GET user's videos (protected)
  VIDEO_BY_ID: (id) => `/api/videos/${id}`,     // GET video metadata (public if video is public)
  VIDEO_STREAM: (id) => `/api/videos/${id}/stream`,  // GET HLS/MP4 URL
  VIDEO_PLAY: (id) => `/api/videos/${id}/play`,      // Stream video file
  VIDEO_THUMBNAIL: (id) => `/api/videos/${id}/thumbnail`,  // GET thumbnail
  VIDEO_VIEW: (id) => `/api/videos/${id}/view`,      // POST increment view
  VIDEO_LIKE: (id) => `/api/videos/${id}/like`,      // POST like video (protected)
  IMPORT_RECORDING: '/api/videos/import-recording',  // POST import recording (protected)
  
  // Recordings
  RECORDINGS: '/api/recordings',
  RECORDING_BY_ID: (id) => `/api/recordings/${id}`,
};

export const HLS_BASE_URL = '/hls';

export const VIDEO_VISIBILITY = {
  PUBLIC: 'public',
  PRIVATE: 'private',
  UNLISTED: 'unlisted',
};

export const VIDEO_STATUS = {
  PENDING: 'pending',
  READY: 'ready',
  FAILED: 'failed',
  ARCHIVED: 'archived',
};

export const STREAM_STATUS = {
  OFFLINE: 'offline',
  LIVE: 'live',
  ENDED: 'ended',
};