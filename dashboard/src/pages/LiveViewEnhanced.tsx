import React, { useState, useEffect, useRef } from 'react';
import { CameraSidebarNew } from '@/components/CameraSidebarNew';
import { CameraSidebarRecordingSection } from '@/components/CameraSidebarRecordingSection';
import { StreamGridEnhanced, StreamGridEnhancedRef } from '@/components/StreamGridEnhanced';
import { ToastContainer } from '@/components/Toast';
import { X, Info, Trash2, AlertTriangle } from 'lucide-react';
import type { Camera } from '@/types';
import { useFolderStore } from '@/stores/folderStore';
import { useStreamStore } from '@/stores/streamStore';
import { useCameraStore } from '@/stores/cameraStore';
import { cn } from '@/utils/cn';
import { api } from '@/services/api';
import { useToast } from '@/hooks/useToast';

export function LiveViewEnhanced() {
  const [selectedCameraId, setSelectedCameraId] = useState<string | null>(null);
  const [showCameraDetails, setShowCameraDetails] = useState(false);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [deletingCamera, setDeletingCamera] = useState(false);
  const gridRef = useRef<StreamGridEnhancedRef>(null);
  const { initializeDefaultFolders, folders } = useFolderStore();
  const { reservations, releaseStream } = useStreamStore();
  const { cameras, selectCamera, selectedCamera, fetchCameras } = useCameraStore();
  const { toasts, removeToast, error: showError, warning: showWarning } = useToast();

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
        selectCamera(camera);
        // Don't show camera details panel when dragging
        setShowCameraDetails(false);
      } else {
        // Grid is full
        showWarning('Grid is full. Please clear some cells or change the layout.');
      }
    }
  };

  const handleCameraDragStart = (camera: Camera, folderId: string | null) => {
    setSelectedCameraId(camera.id);
    selectCamera(camera);
    // Don't show camera details panel when dragging
    setShowCameraDetails(false);
  };

  const handleCloseCameraDetails = () => {
    setShowCameraDetails(false);
  };

  const handleDeleteCamera = async () => {
    if (!selectedCamera) return;

    setDeletingCamera(true);
    try {
      await api.deleteCamera(selectedCamera.id);

      // Close details panel and refresh camera list
      setShowCameraDetails(false);
      setShowDeleteConfirm(false);
      selectCamera(null);
      await fetchCameras();
    } catch (error) {
      console.error('Failed to delete camera:', error);
      showError('Failed to delete camera. Please try again.');
    } finally {
      setDeletingCamera(false);
    }
  };

  return (
    <div className="flex h-screen overflow-hidden">
      {/* Toast Notifications */}
      <ToastContainer toasts={toasts} onRemove={removeToast} />

      {/* Left Sidebar - Camera Tree */}
      <CameraSidebarNew
        onCameraDoubleClick={handleCameraDoubleClick}
        onCameraDragStart={handleCameraDragStart}
        selectedCameraId={selectedCameraId}
      />

      {/* Main content - Camera Grid */}
      <div className="flex-1 overflow-hidden">
        <StreamGridEnhanced ref={gridRef} />
      </div>

      {/* Right Sidebar - Camera Details */}
      {showCameraDetails && selectedCamera && (
        <div className="w-96 h-full bg-white dark:bg-dark-sidebar border-l border-gray-200 dark:border-dark-border flex flex-col shadow-lg">
          {/* Header */}
          <div className="p-4 border-b border-gray-200 dark:border-dark-border bg-gray-50 dark:bg-dark-secondary">
            <div className="flex items-start justify-between">
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2 mb-1">
                  <Info className="w-5 h-5 text-blue-600 dark:text-primary-500 flex-shrink-0" />
                  <h3 className="text-lg font-semibold text-gray-900 dark:text-text-primary truncate">
                    Camera Details
                  </h3>
                </div>
                <p className="text-sm text-gray-600 dark:text-text-secondary truncate" title={selectedCamera.name}>
                  {selectedCamera.name}
                </p>
              </div>
              <button
                onClick={handleCloseCameraDetails}
                className="p-1.5 hover:bg-gray-200 dark:hover:bg-dark-surface rounded-lg transition-colors flex-shrink-0 ml-2"
                title="Close details"
              >
                <X className="w-5 h-5 text-gray-600 dark:text-text-secondary" />
              </button>
            </div>
          </div>

          {/* Camera Info */}
          <div className="p-4 border-b border-gray-200 dark:border-dark-border space-y-3">
            <div>
              <label className="block text-xs font-medium text-gray-500 dark:text-text-muted mb-1">
                Status
              </label>
              <div className="flex items-center gap-2">
                <span
                  className={cn(
                    'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium',
                    selectedCamera.status === 'ONLINE' &&
                      'bg-green-100 dark:bg-green-500/20 text-green-800 dark:text-green-400',
                    selectedCamera.status === 'OFFLINE' &&
                      'bg-gray-100 dark:bg-gray-500/20 text-gray-800 dark:text-text-secondary',
                    selectedCamera.status === 'MAINTENANCE' &&
                      'bg-yellow-100 dark:bg-yellow-500/20 text-yellow-800 dark:text-yellow-400',
                    selectedCamera.status === 'ERROR' && 'bg-red-100 dark:bg-red-500/20 text-red-800 dark:text-red-400'
                  )}
                >
                  {selectedCamera.status}
                </span>
              </div>
            </div>

            <div>
              <label className="block text-xs font-medium text-gray-500 dark:text-text-muted mb-1">
                Source
              </label>
              <p className="text-sm text-gray-900 dark:text-text-primary">
                {selectedCamera.source.replace('_', ' ')}
              </p>
            </div>

            {selectedCamera.milestone_device_id && (
              <div>
                <label className="block text-xs font-medium text-gray-500 dark:text-text-muted mb-1">
                  Milestone Device ID
                </label>
                <p className="text-sm text-gray-900 dark:text-text-primary font-mono text-xs break-all">
                  {selectedCamera.milestone_device_id}
                </p>
              </div>
            )}

            {selectedCamera.location && (
              <div>
                <label className="block text-xs font-medium text-gray-500 dark:text-text-muted mb-1">
                  Location
                </label>
                <p className="text-sm text-gray-900 dark:text-text-primary">
                  {selectedCamera.location.address}
                </p>
              </div>
            )}
          </div>

          {/* Action Buttons */}
          <div className="p-4 border-b border-gray-200 dark:border-dark-border">
            <button
              onClick={() => setShowDeleteConfirm(true)}
              className="w-full flex items-center justify-center gap-2 px-4 py-2 bg-red-50 dark:bg-red-500/10 hover:bg-red-100 dark:hover:bg-red-500/20 text-red-700 dark:text-red-400 rounded-lg transition-colors border border-red-200 dark:border-red-500/30"
            >
              <Trash2 className="w-4 h-4" />
              <span className="text-sm font-medium">Delete Camera</span>
            </button>
          </div>

          {/* Recording Section */}
          <div className="flex-1 overflow-y-auto">
            <CameraSidebarRecordingSection selectedCamera={selectedCamera} />
          </div>
        </div>
      )}

      {/* Delete Confirmation Dialog */}
      {showDeleteConfirm && selectedCamera && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white dark:bg-dark-secondary rounded-lg shadow-xl max-w-md w-full mx-4">
            <div className="p-6">
              <div className="flex items-start gap-4">
                <div className="flex-shrink-0 w-12 h-12 rounded-full bg-red-100 dark:bg-red-500/20 flex items-center justify-center">
                  <AlertTriangle className="w-6 h-6 text-red-600 dark:text-red-400" />
                </div>
                <div className="flex-1">
                  <h3 className="text-lg font-semibold text-gray-900 dark:text-text-primary mb-2">
                    Delete Camera?
                  </h3>
                  <p className="text-sm text-gray-600 dark:text-text-secondary mb-4">
                    Are you sure you want to delete <strong>{selectedCamera.name}</strong>?
                    This will remove the camera and all its associated data. This action cannot be undone.
                  </p>
                  <div className="flex gap-3 justify-end">
                    <button
                      onClick={() => setShowDeleteConfirm(false)}
                      disabled={deletingCamera}
                      className="px-4 py-2 text-sm font-medium text-gray-700 dark:text-text-secondary bg-white dark:bg-dark-surface border border-gray-300 dark:border-dark-border rounded-lg hover:bg-gray-50 dark:hover:bg-dark-elevated disabled:opacity-50"
                    >
                      Cancel
                    </button>
                    <button
                      onClick={handleDeleteCamera}
                      disabled={deletingCamera}
                      className="px-4 py-2 text-sm font-medium text-white bg-red-600 rounded-lg hover:bg-red-700 disabled:opacity-50 flex items-center gap-2"
                    >
                      {deletingCamera ? (
                        <>
                          <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
                          Deleting...
                        </>
                      ) : (
                        <>
                          <Trash2 className="w-4 h-4" />
                          Delete Camera
                        </>
                      )}
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
