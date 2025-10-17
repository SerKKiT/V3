import React, { useState, useEffect, useRef } from 'react';
import { useAuth } from '../hooks/useAuth';
import { Header } from '../components/Layout';
import { CreateStreamModal, StreamCard, StreamDetailsModal } from '../components/Stream';
import { VideoList, EditVideoModal } from '../components/Video';
import { Button, useToast } from '../components/Common';
import { 
  Plus, 
  Radio, 
  Video, 
  Eye, 
  RefreshCw
} from 'lucide-react';
import { streamsAPI } from '../api/streams';
import { videosAPI } from '../api/videos';

export const DashboardPage = () => {
  const { user } = useAuth();
  const toast = useToast();
  const isInitialLoad = useRef(true);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [selectedStream, setSelectedStream] = useState(null);
  const [showDetailsModal, setShowDetailsModal] = useState(false);
  const [selectedVideo, setSelectedVideo] = useState(null);
  const [showEditModal, setShowEditModal] = useState(false);
  const [streams, setStreams] = useState([]);
  const [videos, setVideos] = useState([]);
  const [stats, setStats] = useState({
    totalStreams: 0,
    totalVideos: 0,
    totalViews: 0,
  });
  const [loading, setLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [videoError, setVideoError] = useState(null);

  const loadDashboardData = async (showToast = false, silent = false) => {
    if (!silent) {
      setLoading(true);
    }
    setVideoError(null);
    
    try {
      console.log(silent ? '🔄 Silent refresh...' : '📊 Loading dashboard data...');
      
      const streamsResponse = await streamsAPI.getUserStreams();
      const loadedStreams = streamsResponse.streams || [];
      setStreams(loadedStreams);

      let loadedVideos = [];
      try {
        const videosData = await videosAPI.getUserVideos();
        loadedVideos = videosData.videos || [];
        setVideos(loadedVideos);
      } catch (err) {
        console.error('❌ Failed to load videos:', err);
        setVideoError('Failed to load videos');
        if (showToast) {
          toast.error('Failed to load videos');
        }
        setVideos([]);
      }

      const totalViews = loadedVideos.reduce((sum, video) => sum + (video.view_count || 0), 0);

      setStats({
        totalStreams: loadedStreams.length,
        totalVideos: loadedVideos.length,
        totalViews,
      });

      if (showToast) {
        toast.success('Dashboard refreshed');
      }
    } catch (error) {
      console.error('❌ Failed to load dashboard data:', error);
      if (showToast) {
        toast.error('Failed to load dashboard');
      }
    } finally {
      if (!silent) {
        setLoading(false);
      }
    }
  };

  useEffect(() => {
    if (isInitialLoad.current) {
      loadDashboardData(false, false);
      isInitialLoad.current = false;
    }

    const interval = setInterval(() => {
      loadDashboardData(false, true);
    }, 10000);

    return () => clearInterval(interval);
  }, []);

  const handleStreamCreated = (newStream) => {
    const streamData = newStream.stream || newStream;
    setStreams([streamData, ...streams]);
    setStats({
      ...stats,
      totalStreams: stats.totalStreams + 1,
    });
    toast.success('Stream created successfully');
  };

  const handleManageStream = (stream) => {
    setSelectedStream(stream);
    setShowDetailsModal(true);
  };

  const handleStreamUpdated = (updatedStream) => {
    const streamData = updatedStream.stream || updatedStream;
    setStreams(streams.map(s => s.id === streamData.id ? streamData : s));
    setSelectedStream(streamData);
    toast.success('Stream updated');
  };

  const handleStreamDeleted = (streamId) => {
    setStreams(streams.filter(s => s.id !== streamId));
    setStats({
      ...stats,
      totalStreams: stats.totalStreams - 1,
    });
    toast.success('Stream deleted');
  };

  const handleDeleteVideo = async (videoId) => {
    setVideos(videos.filter(v => v.id !== videoId));
    setStats({
      ...stats,
      totalVideos: stats.totalVideos - 1,
    });
    toast.success('Video deleted successfully');
  };

  const handleUpdateVideo = (video) => {
    setSelectedVideo(video);
    setShowEditModal(true);
  };

  const handleVideoUpdated = (updatedVideo) => {
    setVideos(videos.map(v => v.id === updatedVideo.id ? updatedVideo : v));
    toast.success('Video updated successfully');
  };

  const handleRefresh = async () => {
    setIsRefreshing(true);
    await loadDashboardData(true, false);
    setIsRefreshing(false);
  };

  return (
    <div className="min-h-screen bg-gray-900">
      <Header />

      <main className="container mx-auto px-4 py-8">
        {/* Welcome Section */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-white mb-2">
            Welcome back, {user?.username}! 👋
          </h1>
          <p className="text-gray-400">
            Here's what's happening with your streams and content.
          </p>
        </div>

        {/* Stats Cards */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
          <div className="card">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-gray-400 text-sm mb-1">Total Streams</p>
                <p className="text-3xl font-bold text-white">
                  {stats.totalStreams}
                </p>
              </div>
              <div className="w-12 h-12 bg-primary-600/20 rounded-lg flex items-center justify-center">
                <Radio className="w-6 h-6 text-primary-500" />
              </div>
            </div>
          </div>

          <div className="card">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-gray-400 text-sm mb-1">Total Videos</p>
                <p className="text-3xl font-bold text-white">
                  {stats.totalVideos}
                </p>
              </div>
              <div className="w-12 h-12 bg-primary-600/20 rounded-lg flex items-center justify-center">
                <Video className="w-6 h-6 text-primary-500" />
              </div>
            </div>
          </div>

          <div className="card">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-gray-400 text-sm mb-1">Total Views</p>
                <p className="text-3xl font-bold text-white">
                  {stats.totalViews.toLocaleString()}
                </p>
              </div>
              <div className="w-12 h-12 bg-primary-600/20 rounded-lg flex items-center justify-center">
                <Eye className="w-6 h-6 text-primary-500" />
              </div>
            </div>
          </div>
        </div>

        {/* Quick Actions */}
        <div className="card mb-8">
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-xl font-bold text-white">Quick Actions</h2>
          </div>
          
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <button
              onClick={() => setShowCreateModal(true)}
              className="flex items-center gap-4 p-4 bg-primary-600 hover:bg-primary-700 rounded-lg transition text-left"
            >
              <div className="w-12 h-12 bg-white/20 rounded-lg flex items-center justify-center flex-shrink-0">
                <Plus className="w-6 h-6 text-white" />
              </div>
              <div>
                <h3 className="text-white font-semibold mb-1">Create New Stream</h3>
                <p className="text-primary-100 text-sm">Set up a new live stream</p>
              </div>
            </button>

            <button
              onClick={handleRefresh}
              disabled={isRefreshing}
              className="flex items-center gap-4 p-4 bg-gray-700 hover:bg-gray-600 rounded-lg transition text-left disabled:opacity-50"
            >
              <div className="w-12 h-12 bg-gray-600 rounded-lg flex items-center justify-center flex-shrink-0">
                <RefreshCw className={`w-6 h-6 text-gray-300 ${isRefreshing ? 'animate-spin' : ''}`} />
              </div>
              <div>
                <h3 className="text-white font-semibold mb-1">Refresh Data</h3>
                <p className="text-gray-400 text-sm">Update your dashboard stats</p>
              </div>
            </button>
          </div>
        </div>

        {/* My Streams - ✅ ИСПРАВЛЕНИЕ: Добавлен скроллинг */}
        <div className="card mb-8">
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-xl font-bold text-white">My Streams</h2>
            <Button onClick={() => setShowCreateModal(true)}>
              <Plus className="w-4 h-4 mr-2" />
              New Stream
            </Button>
          </div>

          {/* ✅ Контейнер с фиксированной высотой и скроллингом */}
          <div className="max-h-[600px] overflow-y-auto pr-2 custom-scrollbar">
            {loading && streams.length === 0 ? (
              <div className="text-center py-12">
                <div className="w-12 h-12 border-4 border-primary-600 border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
                <p className="text-gray-400">Loading streams...</p>
              </div>
            ) : streams.length > 0 ? (
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                {streams.map((stream) => (
                  <StreamCard
                    key={stream.id}
                    stream={stream}
                    showActions={true}
                    onManage={handleManageStream}
                  />
                ))}
              </div>
            ) : (
              <div className="text-center py-12">
                <Radio className="w-16 h-16 text-gray-600 mx-auto mb-4" />
                <h3 className="text-lg font-semibold text-white mb-2">
                  No streams yet
                </h3>
                <p className="text-gray-400 mb-6">
                  Create your first stream to get started
                </p>
                <Button onClick={() => setShowCreateModal(true)}>
                  <Plus className="w-4 h-4 mr-2" />
                  Create Stream
                </Button>
              </div>
            )}
          </div>
        </div>

        {/* Recent Videos - ✅ ИСПРАВЛЕНИЕ: Добавлен скроллинг */}
        <div className="card">
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-xl font-bold text-white">Recent Videos</h2>
            <span className="text-sm text-gray-400">
              {videos.length} video{videos.length !== 1 ? 's' : ''}
            </span>
          </div>

          {/* ✅ Контейнер с фиксированной высотой и скроллингом */}
          <div className="max-h-[800px] overflow-y-auto pr-2 custom-scrollbar">
            <VideoList
              videos={videos}
              onDelete={handleDeleteVideo}
              onUpdate={handleUpdateVideo}
              loading={loading && videos.length === 0}
              error={videoError}
            />
          </div>
        </div>
      </main>

      <CreateStreamModal
        isOpen={showCreateModal}
        onClose={() => setShowCreateModal(false)}
        onSuccess={handleStreamCreated}
      />

      <StreamDetailsModal
        stream={selectedStream}
        isOpen={showDetailsModal}
        onClose={() => {
          setShowDetailsModal(false);
          setSelectedStream(null);
        }}
        onUpdate={handleStreamUpdated}
        onDelete={handleStreamDeleted}
      />

      <EditVideoModal
        video={selectedVideo}
        isOpen={showEditModal}
        onClose={() => {
          setShowEditModal(false);
          setSelectedVideo(null);
        }}
        onSuccess={handleVideoUpdated}
      />
    </div>
  );
};
