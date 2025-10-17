import { useState, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { Header } from '../components/Layout';
import { SearchBar } from '../components/Common';
import { videosAPI } from '../api/videos';
import { Video, Clock, Eye, Calendar, ThumbsUp, Lock, Play } from 'lucide-react';
import { useAuth } from '../hooks/useAuth';

export const VideosPage = () => {
  const [videos, setVideos] = useState([]);
  const [filteredVideos, setFilteredVideos] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [searchQuery, setSearchQuery] = useState('');
  const navigate = useNavigate();
  const { isAuthenticated } = useAuth();

  useEffect(() => {
    fetchVideos();
  }, []);

  useEffect(() => {
    // Фильтрация по поисковому запросу
    if (searchQuery) {
      const filtered = videos.filter(video =>
        video.title?.toLowerCase().includes(searchQuery.toLowerCase()) ||
        video.description?.toLowerCase().includes(searchQuery.toLowerCase()) ||
        video.username?.toLowerCase().includes(searchQuery.toLowerCase())
      );
      setFilteredVideos(filtered);
    } else {
      setFilteredVideos(videos);
    }
  }, [searchQuery, videos]);

  const fetchVideos = async () => {
    try {
      setLoading(true);
      setError(null);
      
      // if (!isAuthenticated) {
      //   setVideos([]);
      //   setFilteredVideos([]);
      //   setError('Please login to view your videos');
      //   return;
      // }

      const data = await videosAPI.getAllVideos();

      console.log('Videos response:', data);
      
      // Обрабатываем разные форматы ответа
      if (Array.isArray(data)) {
        setVideos(data);
        setFilteredVideos(data);
      } else if (data.videos && Array.isArray(data.videos)) {
        setVideos(data.videos);
        setFilteredVideos(data.videos);
      } else {
        setVideos([]);
        setFilteredVideos([]);
      }
    } catch (err) {
      console.error('Error fetching videos:', err);
      setError(err.message);
      setVideos([]);
      setFilteredVideos([]);
    } finally {
      setLoading(false);
    }
  };

  const formatDuration = (seconds) => {
    if (!seconds) return 'N/A';
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    const secs = seconds % 60;
    
    if (hours > 0) {
      return `${hours}:${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
    }
    return `${minutes}:${secs.toString().padStart(2, '0')}`;
  };

  const formatDate = (dateString) => {
    if (!dateString) return 'N/A';
    const date = new Date(dateString);
    const now = new Date();
    const diffTime = Math.abs(now - date);
    const diffDays = Math.floor(diffTime / (1000 * 60 * 60 * 24));
    
    if (diffDays === 0) {
      return 'Today';
    } else if (diffDays === 1) {
      return 'Yesterday';
    } else if (diffDays < 7) {
      return `${diffDays} days ago`;
    } else if (diffDays < 30) {
      return `${Math.floor(diffDays / 7)} weeks ago`;
    } else {
      return date.toLocaleDateString('en-US', {
        year: 'numeric',
        month: 'short',
        day: 'numeric'
      });
    }
  };

  if (loading) {
    return (
      <>
        <Header />
        <div className="min-h-screen bg-gray-900 pt-20">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
            <div className="text-center py-12">
              <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600 mx-auto"></div>
              <p className="mt-4 text-gray-400">Loading videos...</p>
            </div>
          </div>
        </div>
      </>
    );
  }

  // Not authenticated
  if (!isAuthenticated) {
    return (
      <>
        <Header />
        <div className="min-h-screen bg-gray-900 pt-20">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
            <div className="text-center py-12">
              <Lock className="h-16 w-16 text-gray-600 mx-auto mb-4" />
              <h3 className="text-xl font-semibold text-white mb-2">
                Authentication Required
              </h3>
              <p className="text-gray-400 mb-6">
                Please login to view your videos
              </p>
              <Link
                to="/login"
                className="inline-block px-6 py-3 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg transition"
              >
                Login
              </Link>
            </div>
          </div>
        </div>
      </>
    );
  }

  return (
    <>
      <Header />
      <div className="min-h-screen bg-gray-900 pt-20">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          {/* Header */}
          <div className="mb-8">
            <h1 className="text-3xl font-bold text-white mb-2">My Videos</h1>
            <p className="text-gray-400">
              {filteredVideos.length} {filteredVideos.length === 1 ? 'video' : 'videos'}
            </p>
          </div>

          {/* Search Bar */}
          <div className="mb-6">
            <SearchBar
              value={searchQuery}
              onChange={setSearchQuery}
              placeholder="Search by title or description..."
            />
          </div>

          {/* Error State */}
          {error && (
            <div className="bg-red-900/20 border border-red-500 text-red-400 px-4 py-3 rounded-lg mb-6">
              <p className="font-medium">Error loading videos</p>
              <p className="text-sm">{error}</p>
            </div>
          )}

          {/* Videos Grid */}
          {filteredVideos.length > 0 ? (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
              {filteredVideos.map((video) => (
                <PublicVideoCard
                  key={video.id}
                  video={video}
                  onClick={() => navigate(`/video/${video.id}`)}
                  formatDuration={formatDuration}
                  formatDate={formatDate}
                />
              ))}
            </div>
          ) : (
            <div className="text-center py-12">
              <Video className="h-16 w-16 text-gray-600 mx-auto mb-4" />
              <h3 className="text-xl font-semibold text-white mb-2">
                {searchQuery ? 'No videos found' : 'No videos yet'}
              </h3>
              <p className="text-gray-400 mb-6">
                {searchQuery 
                  ? 'Try adjusting your search query' 
                  : 'Import recordings from your dashboard to create videos'}
              </p>
              {searchQuery ? (
                <button
                  onClick={() => setSearchQuery('')}
                  className="text-indigo-400 hover:text-indigo-300"
                >
                  Clear search
                </button>
              ) : (
                <Link
                  to="/dashboard"
                  className="inline-block px-6 py-3 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg transition"
                >
                  Go to Dashboard
                </Link>
              )}
            </div>
          )}
        </div>
      </div>
    </>
  );
};

// Упрощенная публичная карточка видео для страницы /videos
const PublicVideoCard = ({ video, onClick, formatDuration, formatDate }) => {
  const thumbnailUrl = video.thumbnail_path 
    ? `http://localhost/api/videos/${video.id}/thumbnail`
    : null;
  
  // ✅ Visibility приходит напрямую из backend
  const isPrivate = video.visibility === 'private';
  
  return (
    <div
      onClick={onClick}
      className="bg-gray-800 rounded-lg overflow-hidden hover:ring-2 hover:ring-indigo-500 transition-all cursor-pointer group"
    >
      {/* Thumbnail */}
      <div className="relative aspect-video bg-gray-700">
        {thumbnailUrl ? (
          <>
            <img
              src={thumbnailUrl}
              alt={video.title}
              className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
              onError={(e) => {
                e.target.style.display = 'none';
                const fallback = e.target.parentElement.querySelector('.fallback-icon');
                if (fallback) fallback.style.display = 'flex';
              }}
            />
            <div className="fallback-icon w-full h-full items-center justify-center absolute inset-0" style={{ display: 'none' }}>
              <Video className="h-16 w-16 text-gray-600" />
            </div>
          </>
        ) : (
          <div className="w-full h-full flex items-center justify-center">
            <Video className="h-16 w-16 text-gray-600" />
          </div>
        )}
        
        {/* Play button overlay */}
        <div className="absolute inset-0 bg-black/0 group-hover:bg-black/40 transition flex items-center justify-center">
          <div className="w-16 h-16 bg-white/0 group-hover:bg-white/20 rounded-full flex items-center justify-center transition">
            <Play className="w-8 h-8 text-white opacity-0 group-hover:opacity-100 transition" />
          </div>
        </div>

        {/* Duration Badge */}
        {video.duration && video.duration > 0 && (
          <div className="absolute bottom-2 right-2 bg-black/80 px-2 py-1 rounded text-xs text-white font-medium">
            {formatDuration(video.duration)}
          </div>
        )}

        {/* Visibility Badge */}
        {isPrivate && (
          <div className="absolute top-2 left-2 bg-yellow-600/90 px-2 py-1 rounded text-xs text-white font-medium flex items-center gap-1">
            <Lock className="h-3 w-3" />
            Private
          </div>
        )}

        {/* Status Badge */}
        {video.status && video.status !== 'ready' && (
          <div className={`absolute top-2 right-2 px-2 py-1 rounded text-xs text-white font-medium ${
            video.status === 'processing' ? 'bg-yellow-600/90' :
            video.status === 'failed' ? 'bg-red-600/90' :
            'bg-gray-600/90'
          }`}>
            {video.status}
          </div>
        )}
      </div>

      {/* Content */}
      <div className="p-4">
        <h3 className="text-white font-semibold mb-2 line-clamp-2 group-hover:text-indigo-400 transition">
          {video.title || 'Untitled Video'}
        </h3>

        {/* ✅ Username приходит напрямую из backend */}
        {video.username && (
          <p className="text-gray-400 text-sm mb-3">
            {video.username}
          </p>
        )}

        {/* Metadata */}
        <div className="flex items-center gap-3 text-xs text-gray-500">
          <div className="flex items-center gap-1">
            <Eye className="h-3 w-3" />
            <span>{video.view_count || 0}</span>
          </div>
          
          {video.like_count !== undefined && (
            <div className="flex items-center gap-1">
              <ThumbsUp className="h-3 w-3" />
              <span>{video.like_count || 0}</span>
            </div>
          )}
          
          <div className="flex items-center gap-1">
            <Calendar className="h-3 w-3" />
            <span>{formatDate(video.created_at)}</span>
          </div>
        </div>
      </div>
    </div>
  );
};

