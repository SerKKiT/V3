import React from 'react';
import { Video as VideoIcon } from 'lucide-react';
import { VideoCard } from './VideoCard';

export const VideoList = ({ videos, onDelete, onUpdate, loading, error }) => {
  if (loading) {
    return (
      <div className="text-center py-12">
        <div className="w-12 h-12 border-4 border-primary-600 border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
        <p className="text-gray-400">Loading videos...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="text-center py-12">
        <p className="text-red-400">❌ {error}</p>
      </div>
    );
  }

  if (!videos || videos.length === 0) {
    return (
      <div className="text-center py-12">
        <VideoIcon className="w-16 h-16 text-gray-600 mx-auto mb-4" />
        <h3 className="text-lg font-semibold text-white mb-2">No videos yet</h3>
        <p className="text-gray-400">
          Your recorded streams will appear here automatically
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {videos.map(video => (
        <VideoCard
          key={video.id}
          video={video}
          onDelete={onDelete}
          onUpdate={onUpdate}
          showActions={true}
        />
      ))}
    </div>
  );
};
