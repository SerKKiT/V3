import React, { useEffect } from 'react';
import { CheckCircle, XCircle, Info, AlertTriangle, X } from 'lucide-react';

export const Toast = ({ type = 'info', message, onClose, duration = 5000 }) => {
  useEffect(() => {
    const timer = setTimeout(() => {
      onClose();
    }, duration);

    return () => clearTimeout(timer);
  }, [duration, onClose]);

  const icons = {
    success: <CheckCircle className="w-5 h-5" />,
    error: <XCircle className="w-5 h-5" />,
    info: <Info className="w-5 h-5" />,
    warning: <AlertTriangle className="w-5 h-5" />,
  };

  const colors = {
    success: 'bg-green-600 border-green-500',
    error: 'bg-red-600 border-red-500',
    info: 'bg-blue-600 border-blue-500',
    warning: 'bg-yellow-600 border-yellow-500',
  };

  return (
    <div className={`${colors[type]} border-l-4 rounded-lg shadow-lg p-4 flex items-start gap-3 min-w-[300px] max-w-md animate-slide-in`}>
      <div className="flex-shrink-0 text-white">
        {icons[type]}
      </div>
      <p className="text-white flex-1 text-sm">
        {message}
      </p>
      <button
        onClick={onClose}
        className="flex-shrink-0 text-white/80 hover:text-white transition"
      >
        <X className="w-4 h-4" />
      </button>
    </div>
  );
};
