import React, { useEffect } from 'react'; // ✅ Добавьте useEffect
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider } from './context/AuthContext';
import { ToastProvider } from './components/Common';
import { useAuth } from './hooks/useAuth';

// Pages
import { 
  HomePage, 
  LoginPage, 
  RegisterPage, 
  DashboardPage,
  LiveStreamsPage,
  VideosPage,
  WatchStreamPage,
  WatchVideoPage,
  SettingsPage
} from './pages';

// Protected Route Component
const ProtectedRoute = ({ children }) => {
  const { isAuthenticated, loading } = useAuth();

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-900 flex items-center justify-center">
        <div className="text-center">
          <div className="w-16 h-16 border-4 border-primary-600 border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
          <p className="text-gray-400">Loading...</p>
        </div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  return children;
};

function App() {
  // ✅ ДОБАВЬТЕ ЭТО: Логирование localStorage
  useEffect(() => {
    console.log('🔧 Setting up localStorage monitoring...');
    
    const originalSetItem = localStorage.setItem;
    const originalRemoveItem = localStorage.removeItem;
    const originalClear = localStorage.clear;

    localStorage.setItem = function(key, value) {
      if (key === 'token') {
        console.log('🔵 localStorage.setItem("token", ...)');
        console.trace();
      }
      return originalSetItem.apply(this, arguments);
    };

    localStorage.removeItem = function(key) {
      if (key === 'token') {
        console.log('🔴 localStorage.removeItem("token")');
        console.trace();
      }
      return originalRemoveItem.apply(this, arguments);
    };

    localStorage.clear = function() {
      console.log('💥 localStorage.clear()');
      console.trace();
      return originalClear.apply(this, arguments);
    };

    // Проверяем токен при загрузке
    const currentToken = localStorage.getItem('token');
    console.log('🔑 Initial token:', currentToken ? 'EXISTS' : 'NULL');

    return () => {
      localStorage.setItem = originalSetItem;
      localStorage.removeItem = originalRemoveItem;
      localStorage.clear = originalClear;
    };
  }, []);

  return (
    <BrowserRouter>
      <AuthProvider>
        <ToastProvider>
          <Routes>
            <Route path="/" element={<HomePage />} />
            <Route path="/login" element={<LoginPage />} />
            <Route path="/register" element={<RegisterPage />} />
            
            {/* Public Routes */}
            <Route path="/live" element={<LiveStreamsPage />} />
            <Route path="/videos" element={<VideosPage />} />
            <Route path="/watch/:id" element={<WatchStreamPage />} />
            <Route path="/video/:id" element={<WatchVideoPage />} /> 
            
            {/* Protected Routes */}
            <Route
              path="/dashboard"
              element={
                <ProtectedRoute>
                  <DashboardPage />
                </ProtectedRoute>
              }
            />
            
            <Route
              path="/settings"
              element={
                <ProtectedRoute>
                  <SettingsPage />
                </ProtectedRoute>
              }
            />
            
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </ToastProvider>
      </AuthProvider>
    </BrowserRouter>
  );
}

export default App;
