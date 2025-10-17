import React, { createContext, useContext, useState } from 'react';
import { Toast } from './Toast';

// ✅ ЭКСПОРТИРУЕМ ToastContext
export const ToastContext = createContext();

export const useToast = () => {
  const context = useContext(ToastContext);
  if (!context) {
    throw new Error('useToast must be used within ToastProvider');
  }
  return context;
};

export const ToastProvider = ({ children }) => {
  const [toasts, setToasts] = useState([]);

  const addToast = (type, message, duration = 5000) => {
    const id = Date.now();
    setToasts(prev => [...prev, { id, type, message, duration }]);
  };

  const removeToast = (id) => {
    setToasts(prev => prev.filter(toast => toast.id !== id));
  };

  const toast = {
    success: (message, duration) => addToast('success', message, duration),
    error: (message, duration) => addToast('error', message, duration),
    info: (message, duration) => addToast('info', message, duration),
    warning: (message, duration) => addToast('warning', message, duration),
  };

  return (
    <ToastContext.Provider value={toast}>
      {children}
      
      {/* Toast Container */}
      <div className="fixed top-4 right-4 z-50 space-y-2">
        {toasts.map(t => (
          <Toast
            key={t.id}
            type={t.type}
            message={t.message}
            duration={t.duration}
            onClose={() => removeToast(t.id)}
          />
        ))}
      </div>
    </ToastContext.Provider>
  );
};
