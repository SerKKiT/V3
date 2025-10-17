import React, { useState, useEffect } from 'react';
import { User, Lock, Mail, Loader2, AlertCircle, CheckCircle2, ArrowLeft } from 'lucide-react';
import { Link } from 'react-router-dom';
import { Header } from '../components/Layout';

export const SettingsPage = () => {
    const [user, setUser] = useState(null);
    const [loading, setLoading] = useState(true);
    const [activeTab, setActiveTab] = useState('profile');
    
    // Profile form
    const [profileForm, setProfileForm] = useState({
        username: '',
        email: '',
    });
    
    // Password form
    const [passwordForm, setPasswordForm] = useState({
        oldPassword: '',
        newPassword: '',
        confirmPassword: '',
    });
    
    // Alerts
    const [alert, setAlert] = useState(null);
    const [saving, setSaving] = useState(false);

    useEffect(() => {
        fetchProfile();
    }, []);

    const fetchProfile = async () => {
        try {
            const token = localStorage.getItem('token');
            const response = await fetch('http://localhost/api/auth/profile', {
                headers: {
                    'Authorization': `Bearer ${token}`,
                },
            });

            if (response.ok) {
                const data = await response.json();
                setUser(data);
                setProfileForm({
                    username: data.username,
                    email: data.email,
                });
            } else {
                throw new Error('Failed to fetch profile');
            }
        } catch (error) {
            console.error('Failed to fetch profile:', error);
            showAlert('error', 'Failed to load profile');
        } finally {
            setLoading(false);
        }
    };

    const showAlert = (type, message) => {
        setAlert({ type, message });
        setTimeout(() => setAlert(null), 5000);
    };

    const handleUpdateProfile = async (e) => {
        e.preventDefault();
        setSaving(true);

        try {
            const token = localStorage.getItem('token');
            const response = await fetch('http://localhost/api/auth/profile', {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${token}`,
                },
                body: JSON.stringify(profileForm),
            });

            if (response.ok) {
                const data = await response.json();
                setUser(data);
                
                // Update user in localStorage
                const currentUser = JSON.parse(localStorage.getItem('user'));
                localStorage.setItem('user', JSON.stringify({ ...currentUser, ...data }));
                
                showAlert('success', 'Profile updated successfully');
            } else {
                const error = await response.json();
                throw new Error(error.error || 'Failed to update profile');
            }
        } catch (error) {
            console.error('Failed to update profile:', error);
            showAlert('error', error.message);
        } finally {
            setSaving(false);
        }
    };

    const handleChangePassword = async (e) => {
        e.preventDefault();

        if (passwordForm.newPassword !== passwordForm.confirmPassword) {
            showAlert('error', 'New passwords do not match');
            return;
        }

        if (passwordForm.newPassword.length < 6) {
            showAlert('error', 'Password must be at least 6 characters');
            return;
        }

        setSaving(true);

        try {
            const token = localStorage.getItem('token');
            const response = await fetch('http://localhost/api/auth/change-password', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${token}`,
                },
                body: JSON.stringify({
                    old_password: passwordForm.oldPassword,
                    new_password: passwordForm.newPassword,
                }),
            });

            if (response.ok) {
                showAlert('success', 'Password changed successfully');
                setPasswordForm({
                    oldPassword: '',
                    newPassword: '',
                    confirmPassword: '',
                });
            } else {
                const error = await response.json();
                throw new Error(error.error || 'Failed to change password');
            }
        } catch (error) {
            console.error('Failed to change password:', error);
            showAlert('error', error.message);
        } finally {
            setSaving(false);
        }
    };

    if (loading) {
        return (
            <div className="min-h-screen bg-gray-900">
                <Header />
                <div className="flex items-center justify-center py-20">
                    <Loader2 className="w-8 h-8 animate-spin text-primary-600" />
                </div>
            </div>
        );
    }

    return (
        <div className="min-h-screen bg-gray-900">
            <Header />
            
            <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
                {/* Back Button */}
                <Link 
                    to="/dashboard"
                    className="inline-flex items-center gap-2 text-gray-400 hover:text-white transition mb-6"
                >
                    <ArrowLeft size={20} />
                    <span>Back to Dashboard</span>
                </Link>

                {/* Header */}
                <div className="mb-8">
                    <h1 className="text-3xl font-bold text-white mb-2">Settings</h1>
                    <p className="text-gray-400">Manage your account settings and preferences</p>
                </div>

                {/* Alert */}
                {alert && (
                    <div className={`mb-6 p-4 rounded-lg flex items-center gap-3 ${
                        alert.type === 'success' 
                            ? 'bg-green-900/50 text-green-200 border border-green-700'
                            : 'bg-red-900/50 text-red-200 border border-red-700'
                    }`}>
                        {alert.type === 'success' ? (
                            <CheckCircle2 size={20} />
                        ) : (
                            <AlertCircle size={20} />
                        )}
                        <p>{alert.message}</p>
                    </div>
                )}

                {/* Tabs */}
                <div className="bg-gray-800 rounded-lg overflow-hidden">
                    <div className="flex border-b border-gray-700">
                        <button
                            onClick={() => setActiveTab('profile')}
                            className={`flex-1 px-6 py-4 text-sm font-medium transition-colors ${
                                activeTab === 'profile'
                                    ? 'bg-gray-900 text-primary-500 border-b-2 border-primary-500'
                                    : 'text-gray-400 hover:text-white'
                            }`}
                        >
                            <User className="inline-block mr-2" size={18} />
                            Profile
                        </button>
                        <button
                            onClick={() => setActiveTab('security')}
                            className={`flex-1 px-6 py-4 text-sm font-medium transition-colors ${
                                activeTab === 'security'
                                    ? 'bg-gray-900 text-primary-500 border-b-2 border-primary-500'
                                    : 'text-gray-400 hover:text-white'
                            }`}
                        >
                            <Lock className="inline-block mr-2" size={18} />
                            Security
                        </button>
                    </div>

                    <div className="p-6">
                        {/* Profile Tab */}
                        {activeTab === 'profile' && (
                            <form onSubmit={handleUpdateProfile} className="space-y-6">
                                <div>
                                    <label className="block text-sm font-medium text-gray-300 mb-2">
                                        Username
                                    </label>
                                    <input
                                        type="text"
                                        value={profileForm.username}
                                        onChange={(e) => setProfileForm({ ...profileForm, username: e.target.value })}
                                        className="w-full px-4 py-3 bg-gray-700 text-white rounded-lg border border-gray-600 focus:outline-none focus:ring-2 focus:ring-primary-500"
                                        required
                                        minLength="3"
                                        maxLength="30"
                                    />
                                </div>

                                <div>
                                    <label className="block text-sm font-medium text-gray-300 mb-2">
                                        Email
                                    </label>
                                    <div className="relative">
                                        <Mail className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400" size={20} />
                                        <input
                                            type="email"
                                            value={profileForm.email}
                                            onChange={(e) => setProfileForm({ ...profileForm, email: e.target.value })}
                                            className="w-full pl-10 pr-4 py-3 bg-gray-700 text-white rounded-lg border border-gray-600 focus:outline-none focus:ring-2 focus:ring-primary-500"
                                            required
                                        />
                                    </div>
                                </div>

                                <div className="pt-4">
                                    <button
                                        type="submit"
                                        disabled={saving}
                                        className="w-full bg-primary-600 text-white py-3 rounded-lg font-medium hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors flex items-center justify-center gap-2"
                                    >
                                        {saving ? (
                                            <>
                                                <Loader2 className="animate-spin" size={20} />
                                                Saving...
                                            </>
                                        ) : (
                                            'Save Changes'
                                        )}
                                    </button>
                                </div>
                            </form>
                        )}

                        {/* Security Tab */}
                        {activeTab === 'security' && (
                            <form onSubmit={handleChangePassword} className="space-y-6">
                                <div>
                                    <label className="block text-sm font-medium text-gray-300 mb-2">
                                        Current Password
                                    </label>
                                    <input
                                        type="password"
                                        value={passwordForm.oldPassword}
                                        onChange={(e) => setPasswordForm({ ...passwordForm, oldPassword: e.target.value })}
                                        className="w-full px-4 py-3 bg-gray-700 text-white rounded-lg border border-gray-600 focus:outline-none focus:ring-2 focus:ring-primary-500"
                                        required
                                    />
                                </div>

                                <div>
                                    <label className="block text-sm font-medium text-gray-300 mb-2">
                                        New Password
                                    </label>
                                    <input
                                        type="password"
                                        value={passwordForm.newPassword}
                                        onChange={(e) => setPasswordForm({ ...passwordForm, newPassword: e.target.value })}
                                        className="w-full px-4 py-3 bg-gray-700 text-white rounded-lg border border-gray-600 focus:outline-none focus:ring-2 focus:ring-primary-500"
                                        required
                                        minLength="6"
                                    />
                                    <p className="mt-1 text-sm text-gray-400">At least 6 characters</p>
                                </div>

                                <div>
                                    <label className="block text-sm font-medium text-gray-300 mb-2">
                                        Confirm New Password
                                    </label>
                                    <input
                                        type="password"
                                        value={passwordForm.confirmPassword}
                                        onChange={(e) => setProfileForm({ ...passwordForm, confirmPassword: e.target.value })}
                                        className="w-full px-4 py-3 bg-gray-700 text-white rounded-lg border border-gray-600 focus:outline-none focus:ring-2 focus:ring-primary-500"
                                        required
                                        minLength="6"
                                    />
                                </div>

                                <div className="pt-4">
                                    <button
                                        type="submit"
                                        disabled={saving}
                                        className="w-full bg-primary-600 text-white py-3 rounded-lg font-medium hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors flex items-center justify-center gap-2"
                                    >
                                        {saving ? (
                                            <>
                                                <Loader2 className="animate-spin" size={20} />
                                                Changing Password...
                                            </>
                                        ) : (
                                            'Change Password'
                                        )}
                                    </button>
                                </div>
                            </form>
                        )}
                    </div>
                </div>

                {/* Account Info */}
                <div className="mt-6 bg-gray-800 rounded-lg p-6">
                    <h3 className="text-lg font-semibold text-white mb-4">Account Information</h3>
                    <div className="space-y-2 text-sm text-gray-400">
                        <p>Account ID: <span className="text-gray-300 font-mono">{user?.id}</span></p>
                        <p>Member since: <span className="text-gray-300">{new Date(user?.created_at).toLocaleDateString()}</span></p>
                    </div>
                </div>
            </div>
        </div>
    );
};
