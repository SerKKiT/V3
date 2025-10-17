import React from 'react';
import { useNavigate } from 'react-router-dom';
import { 
  Play, 
  Edit2, 
  Trash2, 
  Eye, 
  ThumbsUp, 
  Clock,
  HardDrive,
  Image as ImageIcon,
  Share2  // ✅ Добавить
} from 'lucide-react';
import { videosAPI } from '../../api/videos';
import { useToast } from '../../hooks/useToast'; // ✅ Добавить

export const VideoCard = ({ video, onDelete, onUpdate, showActions = true }) => {
  const navigate = useNavigate();
  const toast = useToast(); // ✅ Добавить

  const handleDelete = async () => {
    if (window.confirm(`Delete video "${video.title}"?`)) {
      try {
        await videosAPI.deleteVideo(video.id);
        onDelete(video.id);
        console.log('✅ Video deleted');
      } catch (error) {
        console.error('❌ Failed to delete video:', error);
        alert('Failed to delete video');
      }
    }
  };

  const handlePlay = () => {
    navigate(`/video/${video.id}`);
  };

  // ✅ НОВОЕ: Функция копирования публичной ссылки
  const handleShare = (e) => {
    e.stopPropagation(); // Предотвратить переход на страницу видео
    const shareUrl = `${window.location.origin}/watch/${video.id}`;
    navigator.clipboard.writeText(shareUrl);
    toast.success('Link copied to clipboard!');
  };

  const formatDuration = (seconds) => {
    if (!seconds) return '0:00';
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  };

  const formatFileSize = (bytes) => {
    if (!bytes) return '0 B';
    const mb = (bytes / 1024 / 1024).toFixed(1);
    return `${mb} MB`;
  };

  const formatDate = (dateString) => {
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', { 
      month: 'short', 
      day: 'numeric', 
      year: 'numeric' 
    });
  };

  const getStatusColor = (status) => {
    switch (status) {
      case 'ready': return 'bg-green-600/20 text-green-400';
      case 'processing': return 'bg-yellow-600/20 text-yellow-400';
      case 'failed': return 'bg-red-600/20 text-red-400';
      default: return 'bg-gray-600/20 text-gray-400';
    }
  };

  const getVisibilityColor = (visibility) => {
    switch (visibility) {
      case 'public': return 'bg-blue-600/20 text-blue-400';
      case 'private': return 'bg-purple-600/20 text-purple-400';
      case 'unlisted': return 'bg-gray-600/20 text-gray-400';
      default: return 'bg-gray-600/20 text-gray-400';
    }
  };

  const getThumbnailUrl = () => {
    if (!video.thumbnail_path) return null;
    return `http://localhost/api/videos/${video.id}/thumbnail`;
  };

  const thumbnailUrl = getThumbnailUrl();

  return (
    <div className="card hover:bg-gray-700/50 transition-all group">
      <div className="flex gap-4">
        {/* Thumbnail */}
        <div 
          onClick={handlePlay}
          className="relative w-64 h-36 bg-gray-800 rounded-lg overflow-hidden flex-shrink-0 cursor-pointer group/thumb"
        >
          {thumbnailUrl ? (
            <>
              <img 
                src={thumbnailUrl} 
                alt={video.title}
                className="w-full h-full object-cover"
                onError={(e) => {
                  console.warn('Failed to load thumbnail for video:', video.id);
                  e.target.style.display = 'none';
                  e.target.nextSibling.style.display = 'flex';
                }}
              />
              <div 
                className="w-full h-full items-center justify-center bg-gray-800"
                style={{ display: 'none' }}
              >
                <ImageIcon className="w-12 h-12 text-gray-600" />
              </div>
            </>
          ) : (
            <div className="w-full h-full flex items-center justify-center">
              <ImageIcon className="w-12 h-12 text-gray-600" />
            </div>
          )}
          
          {video.duration > 0 && (
            <div className="absolute bottom-2 right-2 bg-black/80 text-white text-xs font-semibold px-2 py-1 rounded">
              {formatDuration(video.duration)}
            </div>
          )}

          <div className="absolute inset-0 bg-black/0 group-hover/thumb:bg-black/40 transition flex items-center justify-center">
            <div className="w-12 h-12 bg-white/0 group-hover/thumb:bg-white/20 rounded-full flex items-center justify-center transition">
              <Play className="w-6 h-6 text-white opacity-0 group-hover/thumb:opacity-100 transition" />
            </div>
          </div>
        </div>

        {/* Info */}
        <div className="flex-1 min-w-0">
          <div className="flex items-start justify-between gap-4 mb-3">
            <div className="flex-1 min-w-0">
              <h3 
                onClick={handlePlay}
                className="text-lg font-semibold text-white mb-2 truncate cursor-pointer hover:text-primary-400 transition"
              >
                {video.title}
              </h3>
              
              {/* Badges */}
              <div className="flex items-center gap-2 mb-3">
                <span className={`px-2 py-1 rounded text-xs font-semibold uppercase ${getStatusColor(video.status)}`}>
                  {video.status}
                </span>
                <span className={`px-2 py-1 rounded text-xs font-semibold uppercase ${getVisibilityColor(video.visibility)}`}>
                  {video.visibility}
                </span>
              </div>

              {/* Stats */}
              <div className="flex items-center gap-4 text-sm text-gray-400 mb-3">
                <span className="flex items-center gap-1">
                  <Eye className="w-4 h-4" />
                  {video.view_count || 0}
                </span>
                <span className="flex items-center gap-1">
                  <ThumbsUp className="w-4 h-4" />
                  {video.like_count || 0}
                </span>
                <span className="flex items-center gap-1">
                  <HardDrive className="w-4 h-4" />
                  {formatFileSize(video.file_size)}
                </span>
                {video.created_at && (
                  <span className="flex items-center gap-1">
                    <Clock className="w-4 h-4" />
                    {formatDate(video.created_at)}
                  </span>
                )}
              </div>

              {/* Description */}
              {video.description && (
                <p className="text-sm text-gray-400 line-clamp-2 mb-3">
                  {video.description}
                </p>
              )}

              {/* Tags */}
              {video.tags && video.tags.length > 0 && (
                <div className="flex flex-wrap gap-2">
                  {video.tags.map((tag, index) => (
                    <span 
                      key={index}
                      className="text-xs bg-gray-700 text-gray-300 px-2 py-1 rounded-full"
                    >
                      #{tag}
                    </span>
                  ))}
                </div>
              )}
            </div>

            {/* Actions */}
            {showActions && (
              <div className="flex gap-2 flex-shrink-0">
                <button
                  onClick={handlePlay}
                  className="p-2 bg-primary-600 hover:bg-primary-700 text-white rounded-lg transition"
                  title="Play"
                >
                  <Play className="w-4 h-4" />
                </button>
                {/* ✅ НОВОЕ: Кнопка Share */}
                <button
                  onClick={handleShare}
                  className="p-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition"
                  title="Share"
                >
                  <Share2 className="w-4 h-4" />
                </button>
                <button
                  onClick={() => onUpdate(video)}
                  className="p-2 bg-gray-600 hover:bg-gray-500 text-white rounded-lg transition"
                  title="Edit"
                >
                  <Edit2 className="w-4 h-4" />
                </button>
                <button
                  onClick={handleDelete}
                  className="p-2 bg-red-600 hover:bg-red-700 text-white rounded-lg transition"
                  title="Delete"
                >
                  <Trash2 className="w-4 h-4" />
                </button>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};
