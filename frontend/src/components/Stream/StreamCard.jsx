import React from 'react';
import { Radio, Eye, Clock, Copy, Check, Key, MonitorPlay } from 'lucide-react'; // ✅ Добавить MonitorPlay

export const StreamCard = ({ stream, showActions = false, onManage, onCopyKey }) => {
    const [copied, setCopied] = React.useState(false);

    const handleCopyKey = (e) => {
        e.stopPropagation();
        navigator.clipboard.writeText(stream.stream_key);
        setCopied(true);
        if (onCopyKey) onCopyKey();
        setTimeout(() => setCopied(false), 2000);
    };

    const handleManage = (e) => {
        e.stopPropagation();
        if (onManage) {
            onManage(stream);
        }
    };

    const formatDate = (dateString) => {
        const date = new Date(dateString);
        return date.toLocaleDateString('en-US', { 
            month: 'short', 
            day: 'numeric', 
            year: 'numeric' 
        });
    };

    // ✅ Функция для рендера ABR качеств
    const renderQualityBadges = (qualities) => {
        if (!qualities || qualities.length === 0) return null;
        
        return (
            <div className="flex items-center gap-1 mt-2">
                <MonitorPlay size={12} className="text-gray-400" />
                <div className="flex gap-1">
                    {qualities.slice(0, 4).map((quality) => (
                        <span
                            key={quality}
                            className="px-1.5 py-0.5 text-xs font-medium bg-black/50 text-white rounded"
                        >
                            {quality}
                        </span>
                    ))}
                </div>
            </div>
        );
    };

    const isLive = stream.status === 'live';

    // Thumbnail URL через API вместо прямого MinIO
    const thumbnailUrl = stream.id 
        ? `http://localhost/api/streams/${stream.id}/thumbnail` 
        : null;

    return (
        <div className="bg-gray-800 rounded-lg shadow-lg overflow-hidden hover:shadow-xl transition-shadow border border-gray-700">
            {/* Thumbnail */}
            <div className="relative aspect-video bg-gray-900">
                {thumbnailUrl ? (
                    <img 
                        src={thumbnailUrl}
                        alt={stream.title}
                        className="w-full h-full object-cover"
                        onError={(e) => {
                            e.target.style.display = 'none';
                            e.target.nextElementSibling.style.display = 'flex';
                        }}
                    />
                ) : null}
                
                {/* Fallback placeholder */}
                <div 
                    className="absolute inset-0 flex items-center justify-center text-gray-500"
                    style={{ display: thumbnailUrl ? 'none' : 'flex' }}
                >
                    <svg className="w-12 h-12" fill="currentColor" viewBox="0 0 20 20">
                        <path d="M2 6a2 2 0 012-2h6a2 2 0 012 2v8a2 2 0 01-2 2H4a2 2 0 01-2-2V6zM14.553 7.106A1 1 0 0014 8v4a1 1 0 00.553.894l2 1A1 1 0 0018 13V7a1 1 0 00-1.447-.894l-2 1z" />
                    </svg>
                </div>

                {/* Status Badge */}
                {isLive ? (
                    <div className="absolute top-2 left-2">
                        <span className="bg-red-600 text-white px-2 py-1 rounded text-xs font-bold flex items-center gap-1 animate-pulse">
                            <Radio className="w-3 h-3" />
                            LIVE
                        </span>
                    </div>
                ) : (
                    <div className="absolute top-2 left-2">
                        <span className="bg-gray-700 text-gray-300 px-2 py-1 rounded text-xs">
                            Offline
                        </span>
                    </div>
                )}

                {/* ✅ ABR Quality Badges (только для live стримов) */}
                {isLive && stream.available_qualities && stream.available_qualities.length > 0 && (
                    <div className="absolute top-2 right-2 flex gap-1">
                        {stream.available_qualities.slice(0, 2).map((quality) => (
                            <span
                                key={quality}
                                className="bg-black/75 text-white px-1.5 py-0.5 rounded text-xs font-medium"
                            >
                                {quality}
                            </span>
                        ))}
                        {stream.available_qualities.length > 2 && (
                            <span className="bg-black/75 text-white px-1.5 py-0.5 rounded text-xs font-medium">
                                +{stream.available_qualities.length - 2}
                            </span>
                        )}
                    </div>
                )}

                {/* Viewer Count (only for live streams) */}
                {isLive && (
                    <div className="absolute bottom-2 right-2 bg-black bg-opacity-75 text-white px-2 py-1 rounded text-xs flex items-center gap-1">
                        <Eye className="w-3 h-3" />
                        {stream.viewer_count || 0}
                    </div>
                )}
            </div>

            {/* Content */}
            <div className="p-4">
                <h3 className="font-semibold text-lg mb-2 line-clamp-2 text-white">
                    {stream.title}
                </h3>
                
                {/* Stream Key вместо description */}
                <div className="mb-3">
                    <div className="flex items-center gap-2 text-xs text-gray-500 mb-1">
                        <Key className="w-3 h-3" />
                        <span>Stream Key</span>
                    </div>
                    <div className="flex items-center gap-2">
                        <code className="flex-1 text-xs text-gray-400 bg-gray-900 px-2 py-1 rounded font-mono truncate">
                            {stream.stream_key}
                        </code>
                        <button
                            onClick={handleCopyKey}
                            className="text-gray-400 hover:text-white transition"
                            title="Copy Stream Key"
                        >
                            {copied ? (
                                <Check className="w-4 h-4 text-green-400" />
                            ) : (
                                <Copy className="w-4 h-4" />
                            )}
                        </button>
                    </div>
                </div>

                {/* Stats */}
                <div className="flex items-center justify-between text-sm text-gray-400 mb-3">
                    <div className="flex items-center gap-1">
                        <Clock className="w-4 h-4" />
                        {formatDate(stream.created_at)}
                    </div>
                    <div className="flex items-center gap-1">
                        <Eye className="w-4 h-4" />
                        {stream.viewer_count || 0} views
                    </div>
                </div>

                {/* ✅ ABR Quality Info (под stats) */}
                {renderQualityBadges(stream.available_qualities)}

                {/* Actions */}
                {showActions && (
                    <div className="flex gap-2 mt-3">
                        <button
                            onClick={handleManage}
                            className="flex-1 bg-primary-600 hover:bg-primary-700 text-white px-4 py-2 rounded-lg text-sm font-medium transition"
                        >
                            Manage
                        </button>
                    </div>
                )}
            </div>
        </div>
    );
};
