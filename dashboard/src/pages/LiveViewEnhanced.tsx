import React, { useState, useEffect, useRef } from 'react';
import { CameraSidebarNew } from '@/components/CameraSidebarNew';
import { StreamGridEnhanced, StreamGridEnhancedRef } from '@/components/StreamGridEnhanced';
import type { Camera } from '@/types';
import { useFolderStore } from '@/stores/folderStore';
import { useStreamStore } from '@/stores/streamStore';

export function LiveViewEnhanced() {
  const [selectedCameraId, setSelectedCameraId] = useState<string | null>(null);
  const gridRef = useRef<StreamGridEnhancedRef>(null);
  const { initializeDefaultFolders, folders } = useFolderStore();
  const { reservations, releaseStream } = useStreamStore();

  // Initialize folders on mount
  useEffect(() => {
    if (folders.length === 0) {
      initializeDefaultFolders();
    }
  }, [initializeDefaultFolders, folders.length]);

  // Release all active streams on page unload/refresh
  useEffect(() => {
    const handleBeforeUnload = () => {
      // Release all active reservations
      const apiBaseUrl = import.meta.env.VITE_API_URL || 'http://localhost:8000';
      reservations.forEach((reservation) => {
        // Use navigator.sendBeacon for more reliable cleanup on page unload
        const apiUrl = `${apiBaseUrl}/api/v1/stream/release/${reservation.reservation_id}`;
        // sendBeacon uses POST by default, but we need DELETE
        // So we'll use synchronous fetch as fallback (not ideal but necessary)
        try {
          fetch(apiUrl, { method: 'DELETE', keepalive: true });
        } catch (error) {
          console.error('Failed to release stream on unload:', error);
        }
      });
    };

    window.addEventListener('beforeunload', handleBeforeUnload);

    return () => {
      window.removeEventListener('beforeunload', handleBeforeUnload);
    };
  }, [reservations]);

  const handleCameraDoubleClick = (camera: Camera) => {
    // Auto-assign to next available grid cell
    if (gridRef.current) {
      const success = gridRef.current.addCameraToNextAvailableCell(camera);
      if (success) {
        setSelectedCameraId(camera.id);
      } else {
        // Grid is full
        alert('Grid is full. Please clear some cells or change the layout.');
      }
    }
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
        <StreamGridEnhanced ref={gridRef} />
      </div>
    </div>
  );
}
