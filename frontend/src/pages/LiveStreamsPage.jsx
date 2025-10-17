import React, { useState, useEffect } from 'react';
import { Header } from '../components/Layout';
import { SearchBar } from '../components/Common';
import { LiveStreamCard } from '../components/Stream';
import { Radio, RefreshCw, Filter } from 'lucide-react';
import { streamsAPI } from '../api/streams';

export const LiveStreamsPage = () => {
  const [streams, setStreams] = useState([]);
  const [filteredStreams, setFilteredStreams] = useState([]);
  const [searchQuery, setSearchQuery] = useState('');
  const [loading, setLoading] = useState(true);
  const [sortBy, setSortBy] = useState('viewers'); // 'viewers' | 'recent'

  const loadLiveStreams = async () => {
    setLoading(true);
    try {
      const data = await streamsAPI.getLiveStreams();
      setStreams(data.streams || []);
      setFilteredStreams(data.streams || []);
    } catch (error) {
      console.error('Failed to load live streams:', error);
      setStreams([]);
      setFilteredStreams([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadLiveStreams();
    
    // Auto-refresh every 30 seconds
    const interval = setInterval(loadLiveStreams, 30000);
    
    return () => clearInterval(interval);
  }, []);

  // Filter and sort streams
  useEffect(() => {
    let filtered = [...streams];

    // Search filter
    if (searchQuery) {
      filtered = filtered.filter((stream) =>
        stream.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
        stream.description?.toLowerCase().includes(searchQuery.toLowerCase())
      );
    }

    // Sort
    if (sortBy === 'viewers') {
      filtered.sort((a, b) => (b.viewer_count || 0) - (a.viewer_count || 0));
    } else if (sortBy === 'recent') {
      filtered.sort((a, b) => 
        new Date(b.started_at || b.created_at) - new Date(a.started_at || a.created_at)
      );
    }

    setFilteredStreams(filtered);
  }, [streams, searchQuery, sortBy]);

  return (
    <div className="min-h-screen bg-gray-900">
      <Header />

      <main className="container mx-auto px-4 py-8">
        {/* Header */}
        <div className="mb-8">
          <div className="flex items-center gap-3 mb-4">
            <div className="w-12 h-12 bg-red-600 rounded-full flex items-center justify-center">
              <Radio className="w-6 h-6 text-white" />
            </div>
            <div>
              <h1 className="text-3xl font-bold text-white">Live Streams</h1>
              <p className="text-gray-400">
                {streams.length} {streams.length === 1 ? 'stream' : 'streams'} currently live
              </p>
            </div>
          </div>

          {/* Search and Filters */}
          <div className="flex flex-col sm:flex-row gap-4 items-start sm:items-center justify-between">
            <SearchBar
              value={searchQuery}
              onChange={setSearchQuery}
              onClear={() => setSearchQuery('')}
              placeholder="Search live streams..."
              className="w-full sm:max-w-md"
            />

            <div className="flex gap-3">
              {/* Sort Dropdown */}
              <div className="flex items-center gap-2">
                <Filter className="w-5 h-5 text-gray-400" />
                <select
                  value={sortBy}
                  onChange={(e) => setSortBy(e.target.value)}
                  className="input-field py-2 px-3"
                >
                  <option value="viewers">Most Viewers</option>
                  <option value="recent">Recently Started</option>
                </select>
              </div>

              {/* Refresh Button */}
              <button
                onClick={loadLiveStreams}
                disabled={loading}
                className="btn-secondary flex items-center gap-2"
              >
                <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
                <span className="hidden sm:inline">Refresh</span>
              </button>
            </div>
          </div>
        </div>

        {/* Streams Grid */}
        {loading ? (
          <div className="text-center py-20">
            <div className="w-16 h-16 border-4 border-primary-600 border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
            <p className="text-gray-400">Loading live streams...</p>
          </div>
        ) : filteredStreams.length > 0 ? (
          <>
            {/* Stats Bar */}
            <div className="bg-gray-800 rounded-lg px-6 py-4 mb-6 flex items-center justify-between">
              <div className="text-gray-400 text-sm">
                Showing <span className="text-white font-semibold">{filteredStreams.length}</span> of{' '}
                <span className="text-white font-semibold">{streams.length}</span> live streams
              </div>
              <div className="flex items-center gap-4 text-sm text-gray-400">
                <div className="flex items-center gap-2">
                  <div className="w-2 h-2 bg-red-600 rounded-full animate-pulse"></div>
                  <span>Live</span>
                </div>
              </div>
            </div>

            {/* Grid */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
              {filteredStreams.map((stream) => (
                <LiveStreamCard key={stream.id} stream={stream} />
              ))}
            </div>
          </>
        ) : (
          <div className="text-center py-20">
            <Radio className="w-16 h-16 text-gray-600 mx-auto mb-4" />
            <h3 className="text-xl font-semibold text-white mb-2">
              {searchQuery ? 'No streams found' : 'No live streams'}
            </h3>
            <p className="text-gray-400 mb-6">
              {searchQuery
                ? 'Try adjusting your search query'
                : 'Check back later when streamers go live'}
            </p>
            {searchQuery && (
              <button
                onClick={() => setSearchQuery('')}
                className="btn-secondary"
              >
                Clear Search
              </button>
            )}
          </div>
        )}
      </main>
    </div>
  );
};
