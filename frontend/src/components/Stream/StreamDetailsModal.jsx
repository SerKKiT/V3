import React, { useState } from 'react';
import { X, Copy, Check, Edit2, Trash2, Radio, Eye, Clock, Video } from 'lucide-react';
import { Button } from '../Common';
import { streamsAPI } from '../../api/streams';

export const StreamDetailsModal = ({ stream, isOpen, onClose, onUpdate, onDelete }) => {
    const [isEditing, setIsEditing] = useState(false);
    const [formData, setFormData] = useState({
        title: stream?.title || '',
        description: stream?.description || '',
    });
    const [copiedId, setCopiedId] = useState(false);
    const [copiedKey, setCopiedKey] = useState(false);
    const [copiedUrl, setCopiedUrl] = useState(false);
    const [copiedSrt, setCopiedSrt] = useState(false);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');

    if (!isOpen || !stream) return null;

    const srtUrl = `srt://localhost:6000?streamid=${stream.stream_key}`;

    const handleCopyId = () => {
        navigator.clipboard.writeText(stream.id);
        setCopiedId(true);
        setTimeout(() => setCopiedId(false), 2000);
    };

    const handleCopyKey = () => {
        navigator.clipboard.writeText(stream.stream_key);
        setCopiedKey(true);
        setTimeout(() => setCopiedKey(false), 2000);
    };

    const handleCopyUrl = () => {
        navigator.clipboard.writeText(stream.hls_url || `http://localhost/hls/${stream.stream_key}/playlist.m3u8`);
        setCopiedUrl(true);
        setTimeout(() => setCopiedUrl(false), 2000);
    };

    const handleCopySrt = () => {
        navigator.clipboard.writeText(srtUrl);
        setCopiedSrt(true);
        setTimeout(() => setCopiedSrt(false), 2000);
    };

    const handleEdit = () => {
        setIsEditing(true);
        setFormData({
            title: stream.title,
            description: stream.description || '',
        });
    };

    const handleUpdate = async () => {
        if (!formData.title) {
            setError('Title is required');
            return;
        }

        setLoading(true);
        setError('');

        try {
            const updated = await streamsAPI.updateStream(stream.id, formData);
            onUpdate(updated);
            setIsEditing(false);
        } catch (err) {
            setError(err.response?.data?.error || 'Failed to update stream');
        } finally {
            setLoading(false);
        }
    };

    const handleDelete = async () => {
        if (!window.confirm('Are you sure you want to delete this stream? This action cannot be undone.')) {
            return;
        }

        setLoading(true);
        try {
            await streamsAPI.deleteStream(stream.id);
            onDelete(stream.id);
            onClose();
        } catch (err) {
            setError(err.response?.data?.error || 'Failed to delete stream');
        } finally {
            setLoading(false);
        }
    };

    const formatDate = (dateString) => {
        if (!dateString) return 'N/A';
        const date = new Date(dateString);
        return date.toLocaleString('en-US', {
            month: 'short',
            day: 'numeric',
            year: 'numeric',
            hour: '2-digit',
            minute: '2-digit',
        });
    };

    const getStreamDuration = () => {
        if (!stream.started_at) return 'N/A';
        const start = new Date(stream.started_at);
        const end = stream.ended_at ? new Date(stream.ended_at) : new Date();
        const diffMs = end - start;
        const diffMins = Math.floor(diffMs / 60000);
        const hours = Math.floor(diffMins / 60);
        const mins = diffMins % 60;
        return hours > 0 ? `${hours}h ${mins}m` : `${mins}m`;
    };

    return (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
            <div className="bg-gray-800 rounded-lg max-w-2xl w-full max-h-[90vh] overflow-y-auto border border-gray-700">
                {/* Header */}
                <div className="flex items-center justify-between p-6 border-b border-gray-700">
                    <h2 className="text-2xl font-bold text-white flex items-center gap-2">
                        {stream.status === 'live' && (
                            <span className="flex items-center gap-1 text-red-500">
                                <Radio className="w-5 h-5 animate-pulse" />
                                LIVE
                            </span>
                        )}
                        Stream Details
                    </h2>
                    <button
                        onClick={onClose}
                        className="text-gray-400 hover:text-white transition"
                    >
                        <X className="w-6 h-6" />
                    </button>
                </div>

                {/* Content */}
                <div className="p-6 space-y-6">
                    {error && (
                        <div className="bg-red-500/10 border border-red-500 text-red-400 p-3 rounded-lg">
                            {error}
                        </div>
                    )}

                    {/* Title & Description */}
                    {isEditing ? (
                        <div className="space-y-4">
                            <div>
                                <label className="block text-sm font-medium text-gray-300 mb-2">
                                    Title
                                </label>
                                <input
                                    type="text"
                                    value={formData.title}
                                    onChange={(e) => setFormData({ ...formData, title: e.target.value })}
                                    className="w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-500"
                                    placeholder="Stream title"
                                />
                            </div>
                            <div>
                                <label className="block text-sm font-medium text-gray-300 mb-2">
                                    Description
                                </label>
                                <textarea
                                    value={formData.description}
                                    onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                                    rows={3}
                                    className="w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-500"
                                    placeholder="Stream description"
                                />
                            </div>
                            <div className="flex gap-2">
                                <Button onClick={handleUpdate} disabled={loading}>
                                    {loading ? 'Saving...' : 'Save Changes'}
                                </Button>
                                <Button
                                    onClick={() => setIsEditing(false)}
                                    variant="secondary"
                                    disabled={loading}
                                >
                                    Cancel
                                </Button>
                            </div>
                        </div>
                    ) : (
                        <div>
                            <div className="flex items-start justify-between mb-2">
                                <h3 className="text-xl font-bold text-white">{stream.title}</h3>
                                <button
                                    onClick={handleEdit}
                                    className="text-gray-400 hover:text-white transition"
                                >
                                    <Edit2 className="w-4 h-4" />
                                </button>
                            </div>
                            {stream.description && (
                                <p className="text-gray-400">{stream.description}</p>
                            )}
                        </div>
                    )}

                    {/* Stats */}
                    <div className="grid grid-cols-3 gap-4">
                        <div className="bg-gray-700 p-4 rounded-lg">
                            <div className="flex items-center gap-2 text-gray-400 text-sm mb-1">
                                <Eye className="w-4 h-4" />
                                Views
                            </div>
                            <div className="text-2xl font-bold text-white">
                                {stream.viewer_count || 0}
                            </div>
                        </div>
                        <div className="bg-gray-700 p-4 rounded-lg">
                            <div className="flex items-center gap-2 text-gray-400 text-sm mb-1">
                                <Clock className="w-4 h-4" />
                                Duration
                            </div>
                            <div className="text-2xl font-bold text-white">
                                {getStreamDuration()}
                            </div>
                        </div>
                        <div className="bg-gray-700 p-4 rounded-lg">
                            <div className="flex items-center gap-2 text-gray-400 text-sm mb-1">
                                <Radio className="w-4 h-4" />
                                Status
                            </div>
                            <div className="text-2xl font-bold text-white capitalize">
                                {stream.status}
                            </div>
                        </div>
                    </div>

                    {/* SRT URL - ДОБАВЛЕНО */}
                    <div>
                        <label className="block text-sm font-medium text-gray-400 mb-2 flex items-center gap-2">
                            <Video className="w-4 h-4" />
                            SRT Server URL
                        </label>
                        <div className="flex gap-2">
                            <input
                                type="text"
                                value={srtUrl}
                                readOnly
                                className="flex-1 px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white font-mono text-sm"
                            />
                            <button
                                onClick={handleCopySrt}
                                className="px-4 py-2 bg-gray-700 hover:bg-gray-600 border border-gray-600 rounded-lg text-white transition flex items-center gap-2"
                            >
                                {copiedSrt ? (
                                    <>
                                        <Check className="w-4 h-4 text-green-400" />
                                        Copied
                                    </>
                                ) : (
                                    <>
                                        <Copy className="w-4 h-4" />
                                        Copy
                                    </>
                                )}
                            </button>
                        </div>
                        <p className="text-xs text-gray-500 mt-2">
                            Use this URL in OBS Studio: Settings → Stream → Custom → Server URL
                        </p>
                    </div>

                    {/* Stream ID */}
                    <div>
                        <label className="block text-sm font-medium text-gray-400 mb-2">
                            Stream ID
                        </label>
                        <div className="flex gap-2">
                            <input
                                type="text"
                                value={stream.id}
                                readOnly
                                className="flex-1 px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white font-mono text-sm"
                            />
                            <button
                                onClick={handleCopyId}
                                className="px-4 py-2 bg-gray-700 hover:bg-gray-600 border border-gray-600 rounded-lg text-white transition flex items-center gap-2"
                            >
                                {copiedId ? (
                                    <>
                                        <Check className="w-4 h-4 text-green-400" />
                                        Copied
                                    </>
                                ) : (
                                    <>
                                        <Copy className="w-4 h-4" />
                                        Copy
                                    </>
                                )}
                            </button>
                        </div>
                    </div>

                    {/* Stream Key */}
                    <div>
                        <label className="block text-sm font-medium text-gray-400 mb-2">
                            Stream Key
                        </label>
                        <div className="flex gap-2">
                            <input
                                type="text"
                                value={stream.stream_key}
                                readOnly
                                className="flex-1 px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white font-mono text-sm"
                            />
                            <button
                                onClick={handleCopyKey}
                                className="px-4 py-2 bg-gray-700 hover:bg-gray-600 border border-gray-600 rounded-lg text-white transition flex items-center gap-2"
                            >
                                {copiedKey ? (
                                    <>
                                        <Check className="w-4 h-4 text-green-400" />
                                        Copied
                                    </>
                                ) : (
                                    <>
                                        <Copy className="w-4 h-4" />
                                        Copy
                                    </>
                                )}
                            </button>
                        </div>
                        <p className="text-xs text-gray-500 mt-2">
                            Keep this key private. Do not share it publicly.
                        </p>
                    </div>

                    {/* HLS URL */}
                    <div>
                        <label className="block text-sm font-medium text-gray-400 mb-2">
                            HLS Playback URL
                        </label>
                        <div className="flex gap-2">
                            <input
                                type="text"
                                value={stream.hls_url || `http://localhost/hls/${stream.stream_key}/playlist.m3u8`}
                                readOnly
                                className="flex-1 px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white font-mono text-sm"
                            />
                            <button
                                onClick={handleCopyUrl}
                                className="px-4 py-2 bg-gray-700 hover:bg-gray-600 border border-gray-600 rounded-lg text-white transition flex items-center gap-2"
                            >
                                {copiedUrl ? (
                                    <>
                                        <Check className="w-4 h-4 text-green-400" />
                                        Copied
                                    </>
                                ) : (
                                    <>
                                        <Copy className="w-4 h-4" />
                                        Copy
                                    </>
                                )}
                            </button>
                        </div>
                    </div>

                    {/* Dates */}
                    <div className="grid grid-cols-2 gap-4 text-sm">
                        <div>
                            <span className="text-gray-400">Created:</span>
                            <span className="text-white ml-2">{formatDate(stream.created_at)}</span>
                        </div>
                        {stream.started_at && (
                            <div>
                                <span className="text-gray-400">Started:</span>
                                <span className="text-white ml-2">{formatDate(stream.started_at)}</span>
                            </div>
                        )}
                    </div>

                    {/* Delete Button */}
                    <div className="pt-4 border-t border-gray-700">
                        <button
                            onClick={handleDelete}
                            disabled={loading}
                            className="w-full bg-red-600 hover:bg-red-700 text-white px-4 py-2 rounded-lg font-medium transition flex items-center justify-center gap-2 disabled:opacity-50"
                        >
                            <Trash2 className="w-4 h-4" />
                            Delete Stream
                        </button>
                        <p className="text-xs text-gray-500 text-center mt-2">
                            This will permanently delete the stream and cannot be undone.
                        </p>
                    </div>
                </div>
            </div>
        </div>
    );
};
