import React from 'react';
import { BrowserRouter, Routes, Route, Link, useLocation } from 'react-router-dom';
import { LiveView } from '@/pages/LiveView';
import { LiveViewEnhanced } from '@/pages/LiveViewEnhanced';
import { PlaybackView } from '@/pages/PlaybackView';
import { Toaster } from 'react-hot-toast';
import { Video, History, Activity, Settings } from 'lucide-react';
import { cn } from '@/utils/cn';
import '@livekit/components-styles';

function AppLayout({ children }: { children: React.ReactNode }) {
  const location = useLocation();

  const navigation = [
    { name: 'Live View', path: '/', icon: Video },
    { name: 'Playback', path: '/playback', icon: History },
    { name: 'Analytics', path: '/analytics', icon: Activity },
    { name: 'Settings', path: '/settings', icon: Settings },
  ];

  return (
    <div className="flex h-screen bg-gray-100">
      {/* Sidebar */}
      <div className="w-64 bg-gray-900 text-white flex flex-col">
        {/* Logo */}
        <div className="p-6 border-b border-gray-800">
          <h1 className="text-2xl font-bold">RTA CCTV</h1>
          <p className="text-sm text-gray-400 mt-1">Dashboard</p>
        </div>

        {/* Navigation */}
        <nav className="flex-1 p-4 space-y-1">
          {navigation.map((item) => {
            const isActive = location.pathname === item.path;
            const Icon = item.icon;

            return (
              <Link
                key={item.path}
                to={item.path}
                className={cn(
                  'flex items-center gap-3 px-4 py-3 rounded-lg transition-colors',
                  isActive
                    ? 'bg-primary-600 text-white'
                    : 'text-gray-300 hover:bg-gray-800'
                )}
              >
                <Icon className="w-5 h-5" />
                <span className="font-medium">{item.name}</span>
              </Link>
            );
          })}
        </nav>

        {/* Footer */}
        <div className="p-4 border-t border-gray-800">
          <p className="text-xs text-gray-500">
            © 2024 RTA. All rights reserved.
          </p>
        </div>
      </div>

      {/* Main Content */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {children}
      </div>
    </div>
  );
}

function App() {
  return (
    <BrowserRouter>
      <AppLayout>
        <Routes>
          <Route path="/" element={<LiveViewEnhanced />} />
          <Route path="/legacy" element={<LiveView />} />
          <Route path="/playback" element={<PlaybackView />} />
          <Route
            path="/analytics"
            element={
              <div className="flex items-center justify-center h-full text-gray-500">
                <div className="text-center">
                  <Activity className="w-16 h-16 mx-auto mb-4 opacity-30" />
                  <h2 className="text-xl font-medium">Analytics</h2>
                  <p className="text-sm mt-2">
                    Object Detection Service (YOLOv8) - Coming Soon
                  </p>
                </div>
              </div>
            }
          />
          <Route
            path="/settings"
            element={
              <div className="flex items-center justify-center h-full text-gray-500">
                <div className="text-center">
                  <Settings className="w-16 h-16 mx-auto mb-4 opacity-30" />
                  <h2 className="text-xl font-medium">Settings</h2>
                  <p className="text-sm mt-2">Configuration - Coming Soon</p>
                </div>
              </div>
            }
          />
        </Routes>
      </AppLayout>
      <Toaster position="top-right" />
    </BrowserRouter>
  );
}

export default App;
