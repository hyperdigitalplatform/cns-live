import React, { useState } from 'react';
import { StreamGrid } from '@/components/StreamGrid';
import { CameraSidebar } from '@/components/CameraSidebar';
import type { Camera } from '@/types';

export function LiveView() {
  const [selectedCameras, setSelectedCameras] = useState<Camera[]>([]);

  return (
    <div className="flex h-full">
      {/* Sidebar */}
      <div className="w-80 flex-shrink-0">
        <CameraSidebar
          onCameraSelect={setSelectedCameras}
          selectedCameras={selectedCameras}
        />
      </div>

      {/* Main Content */}
      <div className="flex-1">
        {selectedCameras.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-full bg-gray-100 dark:bg-dark-base text-gray-500 dark:text-text-secondary">
            <p className="text-lg">Select cameras from the sidebar to view</p>
          </div>
        ) : (
          <StreamGrid cameras={selectedCameras} defaultLayout="2x2" />
        )}
      </div>
    </div>
  );
}
