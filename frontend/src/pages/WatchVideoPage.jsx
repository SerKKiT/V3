import { useState, useEffect } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { Header } from '../components/Layout';
import { VODPlayer } from '../components/Video/VODPlayer';
import { videosAPI } from '../api/videos';
import { ArrowLeft, Calendar, Eye, Clock, Share2, Download, Trash2, ThumbsUp, Lock } from 'lucide-react';
import { useAuth } from '../hooks/useAuth';

export const WatchVideoPage = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const { user } = useAuth();
  const [video, setVideo] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [liked, setLiked] = useState(false);
  const [likesCount, setLikesCount] = useState(0);
  const [viewsCount, setViewsCount] = useState(0);

  useEffect(() => {
    fetchVideo();
  }, [id]);

  const fetchVideo = async () => {
    try {
      setLoading(true);
      const data = await videosAPI.getVideo(id);
      
      const videoData = data.video || data;
      console.log('📊 Extracted video:', videoData);
      
      if (!videoData) {
        throw new Error('Video not found');
      }
      
      setVideo(videoData);
      setLikesCount(videoData.like_count || 0);
      setViewsCount(videoData.view_count || 0);
      
      // Увеличиваем счетчик просмотров
      try {
        await videosAPI.incrementView(id);
        setViewsCount(prev => prev + 1);
      } catch (err) {
        console.error('Failed to increment view:', err);
      }
    } catch (err) {
      console.error('Error loading video:', err);
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleLike = async () => {
    try {
      await videosAPI.likeVideo(id);
      setLiked(!liked);
      setLikesCount(prev => liked ? prev - 1 : prev + 1);
    } catch (err) {
      console.error('Failed to like video:', err);
      alert('Failed to like video. Please login first.');
    }
  };

  const handleShare = () => {
    const url = window.location.href;
    if (navigator.share) {
      navigator.share({
        title: video.title,
        text: video.description || video.title,
        url: url,
      }).catch(err => {
        console.log('Share failed:', err);
        copyToClipboard(url);
      });
    } else {
      copyToClipboard(url);
    }
  };

  const copyToClipboard = (text) => {
    navigator.clipboard.writeText(text);
    alert('Video link copied to clipboard!');
  };

  const handleDelete = async () => {
    if (!confirm('Are you sure you want to delete this video? This action cannot be undone.')) {
      return;
    }

    try {
      await videosAPI.deleteVideo(id);
      alert('Video deleted successfully');
      navigate('/videos');
    } catch (err) {
      console.error('Failed to delete video:', err);
      alert('Failed to delete video: ' + err.message);
    }
  };

  const formatDate = (dateString) => {
    if (!dateString) return 'N/A';
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
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

  const formatNumber = (num) => {
    if (num >= 1000000) {
      return (num / 1000000).toFixed(1) + 'M';
    } else if (num >= 1000) {
      return (num / 1000).toFixed(1) + 'K';
    }
    return num;
  };

  if (loading) {
    return (
      <>
        <Header />
        <div className="min-h-screen bg-gray-900 pt-20">
          <div className="max-w-7xl mx-auto px-4 py-8">
            <div className="text-center py-12">
              <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600 mx-auto"></div>
              <p className="mt-4 text-gray-400">Loading video...</p>
            </div>
          </div>
        </div>
      </>
    );
  }

  if (error || !video) {
    return (
      <>
        <Header />
        <div className="min-h-screen bg-gray-900 pt-20">
          <div className="max-w-4xl mx-auto px-4 py-8">
            <div className="text-center py-12">
              <div className="bg-red-900/20 border border-red-500 rounded-lg p-6">
                <p className="text-red-400 text-lg mb-4">
                  {error || 'Video not found'}
                </p>
                <Link 
                  to="/videos"
                  className="text-indigo-400 hover:text-indigo-300"
                >
                  ← Back to Videos
                </Link>
              </div>
            </div>
          </div>
        </div>
      </>
    );
  }

  const isOwner = user && video.user_id === user.id;
  const isPublic = video.visibility === 'public';

  // ✅ ИСПРАВЛЕНО: Используем endpoint который вернет presigned URL
  // Backend проверит JWT в заголовке и сделает редирект на presigned URL от MinIO
  const playUrl = video.video_url || `http://localhost/api/videos/${video.id}/play`;
  console.log('🎥 Play URL:', playUrl);

  const creatorName = video.username || 'Unknown Creator';

  return (
    <>
      <Header />
      <div className="min-h-screen bg-gray-900 pt-20">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <Link 
            to="/videos"
            className="inline-flex items-center text-gray-400 hover:text-white mb-4 transition"
          >
            <ArrowLeft className="h-4 w-4 mr-2" />
            Back to Videos
          </Link>

          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            <div className="lg:col-span-2">
              <VODPlayer 
                videoUrl={playUrl}
                autoplay={false}
                startTime={0}
              />

              <div className="bg-gray-800 rounded-lg p-6 mt-4">
                <div className="flex items-start justify-between mb-4">
                  <h1 className="text-2xl font-bold text-white flex-1">
                    {video.title || 'Untitled Video'}
                  </h1>
                  {!isPublic && (
                    <div className="ml-4 flex items-center gap-1 px-3 py-1 bg-yellow-600/20 border border-yellow-600 rounded-lg text-yellow-400 text-sm">
                      <Lock className="h-4 w-4" />
                      Private
                    </div>
                  )}
                </div>

                <div className="flex items-center justify-between mb-4 pb-4 border-b border-gray-700">
                  <div className="flex items-center gap-4 text-sm text-gray-400">
                    <div className="flex items-center gap-1">
                      <Eye className="h-4 w-4" />
                      <span>{formatNumber(viewsCount)} views</span>
                    </div>
                    
                    <div className="flex items-center gap-1">
                      <Calendar className="h-4 w-4" />
                      <span>{formatDate(video.created_at)}</span>
                    </div>

                    {video.duration && video.duration > 0 && (
                      <div className="flex items-center gap-1">
                        <Clock className="h-4 w-4" />
                        <span>{formatDuration(video.duration)}</span>
                      </div>
                    )}
                  </div>

                  <div className="flex items-center gap-2">
                    <button
                      onClick={handleLike}
                      className={`flex items-center gap-2 px-4 py-2 rounded-lg transition ${
                        liked 
                          ? 'bg-indigo-600 text-white' 
                          : 'bg-gray-700 hover:bg-gray-600 text-gray-300'
                      }`}
                    >
                      <ThumbsUp className="h-5 w-5" />
                      <span>{formatNumber(likesCount)}</span>
                    </button>

                    <button
                      onClick={handleShare}
                      className="p-2 bg-gray-700 hover:bg-gray-600 rounded-lg transition"
                      title="Share"
                    >
                      <Share2 className="h-5 w-5 text-gray-300" />
                    </button>

                    <a
                      href={playUrl}
                      download
                      className="p-2 bg-gray-700 hover:bg-gray-600 rounded-lg transition"
                      title="Download"
                    >
                      <Download className="h-5 w-5 text-gray-300" />
                    </a>

                    {isOwner && (
                      <button
                        onClick={handleDelete}
                        className="p-2 bg-red-900/20 hover:bg-red-900/40 rounded-lg transition"
                        title="Delete"
                      >
                        <Trash2 className="h-5 w-5 text-red-400" />
                      </button>
                    )}
                  </div>
                </div>

                <div className="flex items-center justify-between mb-4">
                  <div className="flex items-center gap-3">
                    <div className="w-12 h-12 bg-indigo-600 rounded-full flex items-center justify-center">
                      <span className="text-white font-semibold text-lg">
                        {creatorName[0].toUpperCase()}
                      </span>
                    </div>
                    <div>
                      <p className="text-white font-medium">
                        {creatorName}
                      </p>
                      <p className="text-gray-400 text-sm">
                        {isPublic ? 'Public' : 'Private'} Video
                      </p>
                    </div>
                  </div>
                </div>

                {video.description && (
                  <div className="mt-4 pt-4 border-t border-gray-700">
                    <h3 className="text-white font-semibold mb-2">Description</h3>
                    <p className="text-gray-400 whitespace-pre-wrap">
                      {video.description}
                    </p>
                  </div>
                )}
              </div>

              <div className="bg-gray-800 rounded-lg p-6 mt-4">
                <h3 className="text-white font-semibold mb-4">Comments</h3>
                <p className="text-gray-500 text-center py-8">
                  Comments feature coming soon...
                </p>
              </div>
            </div>

            <div className="lg:col-span-1">
              <div className="bg-gray-800 rounded-lg p-6 sticky top-24">
                <h3 className="text-white font-semibold mb-4">Video Details</h3>
                
                <div className="space-y-3 text-sm">
                  <div>
                    <span className="text-gray-400">Creator:</span>
                    <p className="text-white">{creatorName}</p>
                  </div>

                  <div>
                    <span className="text-gray-400">Video ID:</span>
                    <p className="text-white font-mono text-xs break-all">{video.id}</p>
                  </div>

                  {video.stream_id && (
                    <div>
                      <span className="text-gray-400">Stream ID:</span>
                      <p className="text-white font-mono text-xs break-all">{video.stream_id}</p>
                    </div>
                  )}

                  {video.recording_id && (
                    <div>
                      <span className="text-gray-400">Recording ID:</span>
                      <p className="text-white font-mono text-xs break-all">{video.recording_id}</p>
                    </div>
                  )}

                  {video.file_size && video.file_size > 0 && (
                    <div>
                      <span className="text-gray-400">File Size:</span>
                      <p className="text-white">
                        {(video.file_size / (1024 * 1024)).toFixed(2)} MB
                      </p>
                    </div>
                  )}

                  <div>
                    <span className="text-gray-400">Visibility:</span>
                    <p className={`font-medium ${
                      isPublic ? 'text-green-400' : 'text-yellow-400'
                    }`}>
                      {video.visibility || 'public'}
                    </p>
                  </div>

                  <div>
                    <span className="text-gray-400">Uploaded:</span>
                    <p className="text-white">{formatDate(video.created_at)}</p>
                  </div>

                  {video.updated_at && video.updated_at !== video.created_at && (
                    <div>
                      <span className="text-gray-400">Last Updated:</span>
                      <p className="text-white">{formatDate(video.updated_at)}</p>
                    </div>
                  )}
                </div>

                {isOwner && (
                  <div className="mt-6 pt-6 border-t border-gray-700">
                    <Link
                      to={`/dashboard`}
                      className="block w-full text-center px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg transition"
                    >
                      Manage Video
                    </Link>
                  </div>
                )}
              </div>
            </div>
          </div>
        </div>
      </div>
    </>
  );
};
