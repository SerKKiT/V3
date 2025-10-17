import React from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import { Radio, Video, ArrowRight, Sparkles } from 'lucide-react';

export const HomePage = () => {
  const { isAuthenticated, user } = useAuth();

  return (
    <div className="min-h-screen bg-gray-900 flex flex-col">
      {/* Navigation */}
      <nav className="border-b border-gray-800 bg-gray-900/80 backdrop-blur-sm sticky top-0 z-50">
        <div className="container mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            <div className="flex items-center gap-2">
              <div className="w-8 h-8 bg-primary-600 rounded-lg flex items-center justify-center">
                <Radio className="w-5 h-5 text-white" />
              </div>
              <span className="text-xl font-bold text-white">StreamPlatform</span>
            </div>
            
            <div className="flex items-center gap-3">
              {isAuthenticated ? (
                <>
                  <Link to="/live" className="text-gray-300 hover:text-white transition px-4 py-2">
                    Live
                  </Link>
                  <Link to="/dashboard" className="btn-primary">
                    Dashboard
                  </Link>
                </>
              ) : (
                <>
                  <Link to="/live" className="text-gray-300 hover:text-white transition px-4 py-2">
                    Browse
                  </Link>
                  <Link to="/login" className="text-gray-300 hover:text-white transition px-4 py-2">
                    Login
                  </Link>
                  <Link to="/register" className="btn-primary">
                    Sign Up
                  </Link>
                </>
              )}
            </div>
          </div>
        </div>
      </nav>

      {/* Hero Section */}
      <main className="flex-1 flex items-center justify-center px-4 py-20">
        <div className="max-w-4xl mx-auto text-center">
          {/* Badge */}
          <div className="inline-flex items-center gap-2 px-4 py-2 bg-primary-600/10 border border-primary-600/20 rounded-full mb-8">
            <Sparkles className="w-4 h-4 text-primary-400" />
            <span className="text-sm text-primary-400 font-medium">
              Live streaming platform
            </span>
          </div>

          {/* Heading */}
          <h1 className="text-5xl md:text-7xl font-bold text-white mb-6 leading-tight">
            Stream your
            <span className="text-transparent bg-clip-text bg-gradient-to-r from-primary-400 to-primary-600"> passion</span>
          </h1>

          {/* Description */}
          <p className="text-xl text-gray-400 mb-12 max-w-2xl mx-auto leading-relaxed">
            Create live streams, save recordings automatically, and share your content with the world.
          </p>

          {/* CTA Buttons */}
          <div className="flex flex-col sm:flex-row gap-4 justify-center mb-16">
            {isAuthenticated ? (
              <>
                <Link to="/dashboard" className="btn-primary text-lg px-8 py-4 inline-flex items-center gap-2 group">
                  Go to Dashboard
                  <ArrowRight className="w-5 h-5 group-hover:translate-x-1 transition" />
                </Link>
                <Link to="/live" className="btn-secondary text-lg px-8 py-4">
                  Browse Live Streams
                </Link>
              </>
            ) : (
              <>
                <Link to="/register" className="btn-primary text-lg px-8 py-4 inline-flex items-center gap-2 group">
                  Start Streaming
                  <ArrowRight className="w-5 h-5 group-hover:translate-x-1 transition" />
                </Link>
                <Link to="/login" className="btn-secondary text-lg px-8 py-4">
                  Sign In
                </Link>
              </>
            )}
          </div>

          {/* Features Grid */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6 max-w-3xl mx-auto">
            <div className="card text-left hover:border-primary-600/50 transition group">
              <div className="flex items-start gap-4">
                <div className="w-12 h-12 bg-primary-600/10 rounded-lg flex items-center justify-center flex-shrink-0 group-hover:bg-primary-600/20 transition">
                  <Radio className="w-6 h-6 text-primary-500" />
                </div>
                <div>
                  <h3 className="text-lg font-semibold text-white mb-2">Live Streaming</h3>
                  <p className="text-gray-400 text-sm">
                    Stream in real-time with low latency to your audience
                  </p>
                </div>
              </div>
            </div>

            <div className="card text-left hover:border-primary-600/50 transition group">
              <div className="flex items-start gap-4">
                <div className="w-12 h-12 bg-primary-600/10 rounded-lg flex items-center justify-center flex-shrink-0 group-hover:bg-primary-600/20 transition">
                  <Video className="w-6 h-6 text-primary-500" />
                </div>
                <div>
                  <h3 className="text-lg font-semibold text-white mb-2">Auto-Save VODs</h3>
                  <p className="text-gray-400 text-sm">
                    Your streams are automatically saved as videos
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </main>

      {/* Footer */}
      <footer className="border-t border-gray-800 py-8">
        <div className="container mx-auto px-4 text-center">
          <p className="text-gray-500 text-sm">
            © 2025 StreamPlatform. Built for creators.
          </p>
        </div>
      </footer>
    </div>
  );
};
