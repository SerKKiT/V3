import { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
import { Header } from '../components/Layout';
import { LivePlayer } from '../components/Stream/LivePlayer';
import { Eye, Clock, Share2, Flag } from 'lucide-react';

export const WatchStreamPage = () => {
  const { id } = useParams();
  const [stream, setStream] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    fetchStream();
  }, [id]);

  const fetchStream = async () => {
    try {
      const response = await fetch(`http://localhost/api/streams/${id}/play`);
      if (!response.ok) {
        throw new Error('Stream not found');
      }

      const data = await response.json();
      console.log('📡 Received stream data:', data); // ← ДОБАВЛЕНО: debug log

      // Проверяем что стрим live и есть hls_url
      if (!data.is_live || !data.hls_url) {
        throw new Error('Stream is not currently live');
      }

      setStream(data);
    } catch (err) {
      console.error('Error loading stream:', err);
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleShare = () => {
    navigator.clipboard.writeText(window.location.href);
    alert('Stream link copied!');
  };

  // Loading state
  if (loading) {
    return (
      <>
        <Header />
        <div className="min-h-screen bg-gray-900 flex items-center justify-center">
          <div className="text-center">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500 mx-auto mb-4"></div>
            <p className="text-gray-400">Loading stream...</p>
          </div>
        </div>
      </>
    );
  }

  // Error state
  if (error || !stream) {
    return (
      <>
        <Header />
        <div className="min-h-screen bg-gray-900 flex items-center justify-center">
          <div className="text-center max-w-md mx-auto p-8">
            <div className="text-red-500 text-6xl mb-4">⚠️</div>
            <h2 className="text-2xl font-bold text-white mb-2">Stream Unavailable</h2>
            <p className="text-gray-400 mb-6">{error || 'This stream has ended or is not currently live.'}</p>
            <Link
              to="/live"
              className="inline-flex items-center px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition"
            >
              Browse Live Streams
            </Link>
          </div>
        </div>
      </>
    );
  }

  return (
    <>
      <Header />
      <div className="min-h-screen bg-gray-900">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
            {/* Main Video Player */}
            <div className="lg:col-span-2">
              <LivePlayer
                hlsUrl={stream.hls_url}
                isLive={stream.is_live}
                availableQualities={stream.available_qualities || []}
              />

              {/* Stream Info */}
              <div className="bg-gray-800 rounded-lg p-6 mt-4">
                <div className="flex items-start justify-between mb-4">
                  <div className="flex-1">
                    <h1 className="text-2xl font-bold text-white mb-2">{stream.title}</h1>
                    
                    {/* Username Display */}
                    <div className="flex items-center space-x-4 text-gray-400 mb-3">
                      <p className="text-gray-400">Streamer</p>
                      <p className="text-xl font-semibold text-white">{stream.username || 'Unknown Streamer'}</p>
                    </div>

                    <div className="flex items-center space-x-4 text-gray-400">
                      <div className="flex items-center space-x-2">
                        <Eye className="w-5 h-5" />
                        <span>{stream.viewer_count || 0} viewers</span>
                      </div>
                      {stream.started_at && (
                        <div className="flex items-center space-x-2">
                          <Clock className="w-5 h-5" />
                          <span>Started {new Date(stream.started_at).toLocaleTimeString()}</span>
                        </div>
                      )}
                      {stream.is_live && (
                        <span className="px-2 py-1 bg-red-600 text-white text-sm font-semibold rounded">
                          LIVE
                        </span>
                      )}
                    </div>
                  </div>

                  {/* Action Buttons */}
                  <div className="flex space-x-2">
                    <button
                      onClick={handleShare}
                      className="p-2 bg-gray-700 hover:bg-gray-600 rounded-lg transition"
                      title="Share"
                    >
                      <Share2 className="w-5 h-5 text-white" />
                    </button>
                    <button
                      className="p-2 bg-gray-700 hover:bg-gray-600 rounded-lg transition"
                      title="Report"
                    >
                      <Flag className="w-5 h-5 text-white" />
                    </button>
                  </div>
                </div>

                {/* Description */}
                {stream.description && (
                  <div className="mt-4 pt-4 border-t border-gray-700">
                    <p className="text-gray-300 whitespace-pre-wrap">{stream.description}</p>
                  </div>
                )}
              </div>
            </div>

            {/* Chat Sidebar */}
            <div className="lg:col-span-1">
              <div className="bg-gray-800 rounded-lg p-6 h-[600px] flex items-center justify-center">
                <p className="text-gray-400">Chat feature coming soon...</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </>
  );
};
