import React from 'react';
import { Link } from 'react-router-dom';
import { Eye, User, Clock } from 'lucide-react';

export const LiveStreamCard = ({ stream }) => {
    const formatViewCount = (count) => {
        if (count >= 1000) {
            return `${(count / 1000).toFixed(1)}K`;
        }
        return count;
    };

    const getStreamDuration = () => {
        if (!stream.started_at) return 'Just started';

        const start = new Date(stream.started_at);
        const now = new Date();
        const diffMs = now - start;
        const diffMins = Math.floor(diffMs / 60000);

        if (diffMins < 60) {
            return `${diffMins}m`;
        }

        const hours = Math.floor(diffMins / 60);
        const mins = diffMins % 60;
        return `${hours}h ${mins}m`;
    };

    // ✅ Исправлено: правильная генерация thumbnail URL
    const getThumbnailUrl = () => {
        if (!stream.thumbnail_url) {
            // ✅ Пытаемся загрузить thumbnail по stream ID
            const timestamp = Math.floor(Date.now() / 30000);
            return `http://localhost/api/streams/${stream.id}/thumbnail?t=${timestamp}`;
        }
        // Обновляем каждые 30 секунд
        const timestamp = Math.floor(Date.now() / 30000);
        return `${stream.thumbnail_url}?t=${timestamp}`;
    };

    return (
        <Link
            to={`/watch/${stream.id}`}
            className="group bg-gray-800 rounded-lg overflow-hidden hover:ring-2 hover:ring-primary-500 transition-all duration-200"
        >
            {/* Thumbnail */}
            <div className="relative aspect-video bg-gray-900">
                <img
                    src={getThumbnailUrl()}
                    alt={stream.title}
                    className="w-full h-full object-cover"
                    onError={(e) => {
                        // ✅ Fallback на SVG placeholder вместо несуществующего файла
                        if (!e.target.dataset.fallback) {
                            e.target.dataset.fallback = 'true';
                            e.target.src = `data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='640' height='360'%3E%3Crect fill='%23374151' width='640' height='360'/%3E%3Ctext fill='%239CA3AF' font-family='system-ui' font-size='24' x='50%25' y='50%25' text-anchor='middle' dominant-baseline='middle'%3ENo Thumbnail%3C/text%3E%3C/svg%3E`;
                        }
                    }}
                />

                {/* LIVE Badge */}
                <div className="absolute top-2 left-2">
                    <span className="bg-red-600 text-white px-2 py-1 rounded text-xs font-bold animate-pulse flex items-center gap-1">
                        <span className="w-2 h-2 bg-white rounded-full"></span>
                        LIVE
                    </span>
                </div>

                {/* Viewer Count */}
                <div className="absolute top-2 right-2 bg-black/70 text-white px-2 py-1 rounded text-xs flex items-center gap-1">
                    <Eye size={12} />
                    {formatViewCount(stream.viewer_count)}
                </div>

                {/* Duration */}
                <div className="absolute bottom-2 right-2 bg-black/70 text-white px-2 py-1 rounded text-xs flex items-center gap-1">
                    <Clock size={12} />
                    {getStreamDuration()}
                </div>
            </div>

            {/* Content */}
            <div className="p-4">
                <h3 className="font-semibold text-white text-lg mb-2 group-hover:text-primary-400 transition-colors line-clamp-2">
                    {stream.title}
                </h3>

                {stream.description && (
                    <p className="text-gray-400 text-sm line-clamp-2 mb-3">
                        {stream.description}
                    </p>
                )}

                {/* Streamer Info */}
                <div className="flex items-center gap-2 text-gray-400 text-sm">
                    <User size={16} />
                    <span>Streamer</span>
                </div>
            </div>
        </Link>
    );
};
