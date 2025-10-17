import React from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../../hooks/useAuth';
import { 
  Home, 
  Radio, 
  Video, 
  LayoutDashboard, 
  LogOut, 
  Settings,
  User,
  Menu,
  X
} from 'lucide-react';

export const Header = () => {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const [mobileMenuOpen, setMobileMenuOpen] = React.useState(false);
  const [profileMenuOpen, setProfileMenuOpen] = React.useState(false);

  const handleLogout = () => {
    logout();
    navigate('/');
  };

  return (
    <header className="bg-gray-800 border-b border-gray-700 sticky top-0 z-50">
      <div className="container mx-auto px-4">
        <div className="flex items-center justify-between h-16">
          {/* Logo */}
          <Link to="/" className="flex items-center gap-2">
            <div className="w-10 h-10 bg-primary-600 rounded-lg flex items-center justify-center">
              <Radio className="w-6 h-6 text-white" />
            </div>
            <span className="text-xl font-bold text-white hidden sm:block">
              StreamPlatform
            </span>
          </Link>

          {/* Desktop Navigation */}
          <nav className="hidden md:flex items-center gap-6">
            <Link
              to="/"
              className="flex items-center gap-2 text-gray-300 hover:text-white transition"
            >
              <Home className="w-5 h-5" />
              <span>Home</span>
            </Link>
            
            <Link
              to="/live"
              className="flex items-center gap-2 text-gray-300 hover:text-white transition"
            >
              <Radio className="w-5 h-5" />
              <span>Live</span>
            </Link>
            
            <Link
                to="/videos"
                className="flex items-center gap-2 text-gray-300 hover:text-white transition"
            >
                <Video className="w-5 h-5" />
                <span>Videos</span>
            </Link>

            
            <Link
              to="/dashboard"
              className="flex items-center gap-2 text-gray-300 hover:text-white transition"
            >
              <LayoutDashboard className="w-5 h-5" />
              <span>Dashboard</span>
            </Link>
          </nav>

          {/* User Menu */}
          <div className="flex items-center gap-4">
            {/* Profile Dropdown */}
            <div className="relative">
              <button
                onClick={() => setProfileMenuOpen(!profileMenuOpen)}
                className="flex items-center gap-2 bg-gray-700 hover:bg-gray-600 rounded-lg px-3 py-2 transition"
              >
                <div className="w-8 h-8 bg-primary-600 rounded-full flex items-center justify-center">
                  <User className="w-5 h-5 text-white" />
                </div>
                <span className="text-white hidden sm:block">{user?.username}</span>
              </button>

              {/* Dropdown Menu */}
              {profileMenuOpen && (
                <>
                  <div
                    className="fixed inset-0 z-10"
                    onClick={() => setProfileMenuOpen(false)}
                  ></div>
                  <div className="absolute right-0 mt-2 w-48 bg-gray-800 border border-gray-700 rounded-lg shadow-lg py-2 z-20">
                    <Link
                      to="/dashboard"
                      className="flex items-center gap-2 px-4 py-2 text-gray-300 hover:bg-gray-700 transition"
                      onClick={() => setProfileMenuOpen(false)}
                    >
                      <LayoutDashboard className="w-4 h-4" />
                      <span>Dashboard</span>
                    </Link>
                    
                    <Link
                      to="/settings"
                      className="flex items-center gap-2 px-4 py-2 text-gray-300 hover:bg-gray-700 transition"
                      onClick={() => setProfileMenuOpen(false)}
                    >
                      <Settings className="w-4 h-4" />
                      <span>Settings</span>
                    </Link>
                    
                    <hr className="my-2 border-gray-700" />
                    
                    <button
                      onClick={handleLogout}
                      className="w-full flex items-center gap-2 px-4 py-2 text-red-400 hover:bg-gray-700 transition"
                    >
                      <LogOut className="w-4 h-4" />
                      <span>Logout</span>
                    </button>
                  </div>
                </>
              )}
            </div>

            {/* Mobile Menu Button */}
            <button
              onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
              className="md:hidden text-gray-300 hover:text-white"
            >
              {mobileMenuOpen ? (
                <X className="w-6 h-6" />
              ) : (
                <Menu className="w-6 h-6" />
              )}
            </button>
          </div>
        </div>

        {/* Mobile Navigation */}
        {mobileMenuOpen && (
          <nav className="md:hidden border-t border-gray-700 py-4 space-y-2">
            <Link
              to="/"
              className="flex items-center gap-2 px-4 py-2 text-gray-300 hover:bg-gray-700 rounded transition"
              onClick={() => setMobileMenuOpen(false)}
            >
              <Home className="w-5 h-5" />
              <span>Home</span>
            </Link>
            
            <Link
              to="/live"
              className="flex items-center gap-2 px-4 py-2 text-gray-300 hover:bg-gray-700 rounded transition"
              onClick={() => setMobileMenuOpen(false)}
            >
              <Radio className="w-5 h-5" />
              <span>Live</span>
            </Link>
            
            <Link
              to="/videos"
              className="flex items-center gap-2 px-4 py-2 text-gray-300 hover:bg-gray-700 rounded transition"
              onClick={() => setMobileMenuOpen(false)}
            >
              <Video className="w-5 h-5" />
              <span>Videos</span>
            </Link>
            
            <Link
              to="/dashboard"
              className="flex items-center gap-2 px-4 py-2 text-gray-300 hover:bg-gray-700 rounded transition"
              onClick={() => setMobileMenuOpen(false)}
            >
              <LayoutDashboard className="w-5 h-5" />
              <span>Dashboard</span>
            </Link>
          </nav>
        )}
      </div>
    </header>
  );
};
