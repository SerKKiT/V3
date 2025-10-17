import React, { useState } from 'react';
import { X, Radio, AlertCircle } from 'lucide-react';
import { Button, Input } from '../Common';
import { streamsAPI } from '../../api/streams';

export const CreateStreamModal = ({ isOpen, onClose, onSuccess }) => {
  const [formData, setFormData] = useState({
    title: '',
    description: '',
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleChange = (e) => {
    setFormData({
      ...formData,
      [e.target.name]: e.target.value,
    });
    setError('');
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError('');

    if (!formData.title) {
      setError('Please enter a stream title');
      setLoading(false);
      return;
    }

    try {
      const response = await streamsAPI.createStream(
        formData.title,
        formData.description
      );
      
      // Reset form
      setFormData({ title: '', description: '' });
      
      // Call success callback
      if (onSuccess) {
        onSuccess(response);
      }
      
      // Close modal
      onClose();
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to create stream');
    } finally {
      setLoading(false);
    }
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      {/* Backdrop */}
      <div
        className="absolute inset-0 bg-black/70"
        onClick={onClose}
      ></div>

      {/* Modal */}
      <div className="relative bg-gray-800 rounded-lg shadow-xl max-w-md w-full">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-gray-700">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 bg-primary-600 rounded-full flex items-center justify-center">
              <Radio className="w-5 h-5 text-white" />
            </div>
            <h2 className="text-xl font-bold text-white">Create Stream</h2>
          </div>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-white transition"
          >
            <X className="w-6 h-6" />
          </button>
        </div>

        {/* Body */}
        <form onSubmit={handleSubmit} className="p-6 space-y-4">
          {/* Error Alert */}
          {error && (
            <div className="bg-red-500/10 border border-red-500 rounded-lg p-4 flex items-start gap-3">
              <AlertCircle className="w-5 h-5 text-red-500 flex-shrink-0 mt-0.5" />
              <div className="flex-1">
                <p className="text-red-500 text-sm">{error}</p>
              </div>
            </div>
          )}

          {/* Title */}
          <Input
            label="Stream Title"
            name="title"
            value={formData.title}
            onChange={handleChange}
            placeholder="What are you streaming today?"
            disabled={loading}
          />

          {/* Description */}
          <div>
            <label className="block text-sm font-medium text-gray-300 mb-2">
              Description (optional)
            </label>
            <textarea
              name="description"
              value={formData.description}
              onChange={handleChange}
              rows={3}
              className="input-field resize-none"
              placeholder="Tell viewers what to expect..."
              disabled={loading}
            />
          </div>

          {/* Actions */}
          <div className="flex gap-3 pt-4">
            <Button
              type="button"
              variant="secondary"
              onClick={onClose}
              disabled={loading}
              className="flex-1"
            >
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={loading}
              className="flex-1"
            >
              {loading ? 'Creating...' : 'Create Stream'}
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
};
