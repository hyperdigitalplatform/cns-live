import React from 'react';
import { BrowserRouter, Routes, Route, Link, useLocation } from 'react-router-dom';
import { LiveView } from '@/pages/LiveView';
import { LiveViewEnhanced } from '@/pages/LiveViewEnhanced';
import { PlaybackView } from '@/pages/PlaybackView';
import CameraDiscovery from '@/pages/CameraDiscovery';
import { Toaster } from 'react-hot-toast';
import { Video, History, Activity, Settings, Camera } from 'lucide-react';
import { cn } from '@/utils/cn';
import '@livekit/components-styles';

function AppLayout({ children }: { children: React.ReactNode }) {
  const location = useLocation();

  const navigation = [
    { name: 'Live View', path: '/', icon: Video },
    { name: 'Playback', path: '/playback', icon: History },
    { name: 'Discovery', path: '/discovery', icon: Camera },
    { name: 'Analytics', path: '/analytics', icon: Activity },
    { name: 'Settings', path: '/settings', icon: Settings },
  ];

  return (
    <div className="flex h-screen bg-gray-100">
      {/* Main Content - Full Screen */}
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
          <Route path="/discovery" element={<CameraDiscovery />} />
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
