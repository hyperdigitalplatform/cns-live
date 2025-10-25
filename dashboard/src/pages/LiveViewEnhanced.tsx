import React, { useState, useEffect } from 'react';
import { CameraSidebarNew } from '@/components/CameraSidebarNew';
import { StreamGridEnhanced } from '@/components/StreamGridEnhanced';
import type { Camera } from '@/types';
import { useFolderStore } from '@/stores/folderStore';

export function LiveViewEnhanced() {
  const [selectedCameraId, setSelectedCameraId] = useState<string | null>(null);
  const [gridCameras, setGridCameras] = useState<(Camera | null)[]>([]);
  const { initializeDefaultFolders, folders } = useFolderStore();

  // Initialize folders on mount
  useEffect(() => {
    if (folders.length === 0) {
      initializeDefaultFolders();
    }
  }, [initializeDefaultFolders, folders.length]);

  const handleCameraDoubleClick = (camera: Camera) => {
    // Auto-assign to next available grid cell
    setGridCameras((prev) => {
      const firstEmptyIndex = prev.findIndex((c) => c === null);
      if (firstEmptyIndex === -1) {
        // Grid is full, add to end (will expand grid)
        return [...prev, camera];
      }
      const newCameras = [...prev];
      newCameras[firstEmptyIndex] = camera;
      return newCameras;
    });
    setSelectedCameraId(camera.id);
  };

  const handleCameraDragStart = (camera: Camera, folderId: string | null) => {
    setSelectedCameraId(camera.id);
  };

  return (
    <div className="flex h-screen overflow-hidden">
      {/* Sidebar */}
      <CameraSidebarNew
        onCameraDoubleClick={handleCameraDoubleClick}
        onCameraDragStart={handleCameraDragStart}
        selectedCameraId={selectedCameraId}
      />

      {/* Main content */}
      <div className="flex-1 overflow-hidden">
        <StreamGridEnhanced />
      </div>
    </div>
  );
}
