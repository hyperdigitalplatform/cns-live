import React, { useEffect, useState } from 'react';
import { useCameraStore } from '@/stores/cameraStore';
import type { Camera, CameraSource, CameraStatus } from '@/types';
import {
  Search,
  Filter,
  Video,
  VideoOff,
  Circle,
  ChevronDown,
} from 'lucide-react';
import { cn } from '@/utils/cn';

interface CameraSidebarProps {
  onCameraSelect: (cameras: Camera[]) => void;
  selectedCameras: Camera[];
}

export function CameraSidebar({
  onCameraSelect,
  selectedCameras,
}: CameraSidebarProps) {
  const { cameras, loading, fetchCameras } = useCameraStore();
  const [searchQuery, setSearchQuery] = useState('');
  const [sourceFilter, setSourceFilter] = useState<CameraSource | 'ALL'>('ALL');
  const [statusFilter, setStatusFilter] = useState<CameraStatus | 'ALL'>('ALL');
  const [showFilters, setShowFilters] = useState(false);

  useEffect(() => {
    fetchCameras();
  }, [fetchCameras]);

  const filteredCameras = cameras.filter((camera) => {
    const matchesSearch =
      camera.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      camera.name_ar?.includes(searchQuery);
    const matchesSource =
      sourceFilter === 'ALL' || camera.source === sourceFilter;
    const matchesStatus =
      statusFilter === 'ALL' || camera.status === statusFilter;

    return matchesSearch && matchesSource && matchesStatus;
  });

  const toggleCamera = (camera: Camera) => {
    const isSelected = selectedCameras.some((c) => c.id === camera.id);
    if (isSelected) {
      onCameraSelect(selectedCameras.filter((c) => c.id !== camera.id));
    } else {
      onCameraSelect([...selectedCameras, camera]);
    }
  };

  const sources: (CameraSource | 'ALL')[] = [
    'ALL',
    'DUBAI_POLICE',
    'SHARJAH_POLICE',
    'ABU_DHABI_POLICE',
    'METRO',
    'TAXI',
    'PARKING',
  ];

  const statuses: (CameraStatus | 'ALL')[] = [
    'ALL',
    'ONLINE',
    'OFFLINE',
    'MAINTENANCE',
    'ERROR',
  ];

  return (
    <div className="flex flex-col h-full bg-white dark:bg-dark-sidebar border-r border-gray-200 dark:border-dark-border">
      {/* Header */}
      <div className="p-4 border-b border-gray-200 dark:border-dark-border">
        <h2 className="text-lg font-semibold text-gray-900 dark:text-text-primary mb-3">Cameras</h2>

        {/* Search */}
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 dark:text-text-muted" />
          <input
            type="text"
            placeholder="Search cameras..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full pl-9 pr-3 py-2 border border-gray-300 dark:border-dark-border rounded-lg text-sm bg-white dark:bg-dark-secondary text-gray-900 dark:text-text-primary placeholder:text-gray-400 dark:placeholder:text-text-muted focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent"
          />
        </div>

        {/* Filters Toggle */}
        <button
          onClick={() => setShowFilters(!showFilters)}
          className="mt-2 flex items-center gap-2 text-sm text-gray-600 dark:text-text-secondary hover:text-gray-900 dark:hover:text-text-primary transition-colors"
        >
          <Filter className="w-4 h-4" />
          <span>Filters</span>
          <ChevronDown
            className={cn(
              'w-4 h-4 transition-transform',
              showFilters && 'rotate-180'
            )}
          />
        </button>

        {/* Filters */}
        {showFilters && (
          <div className="mt-3 space-y-3">
            <div>
              <label className="block text-xs font-medium text-gray-700 dark:text-text-secondary mb-1">
                Source
              </label>
              <select
                value={sourceFilter}
                onChange={(e) =>
                  setSourceFilter(e.target.value as CameraSource | 'ALL')
                }
                className="w-full px-3 py-1.5 border border-gray-300 dark:border-dark-border rounded text-sm bg-white dark:bg-dark-secondary text-gray-900 dark:text-text-primary focus:outline-none focus:ring-2 focus:ring-primary-500"
              >
                {sources.map((source) => (
                  <option key={source} value={source}>
                    {source.replace('_', ' ')}
                  </option>
                ))}
              </select>
            </div>

            <div>
              <label className="block text-xs font-medium text-gray-700 dark:text-text-secondary mb-1">
                Status
              </label>
              <select
                value={statusFilter}
                onChange={(e) =>
                  setStatusFilter(e.target.value as CameraStatus | 'ALL')
                }
                className="w-full px-3 py-1.5 border border-gray-300 dark:border-dark-border rounded text-sm bg-white dark:bg-dark-secondary text-gray-900 dark:text-text-primary focus:outline-none focus:ring-2 focus:ring-primary-500"
              >
                {statuses.map((status) => (
                  <option key={status} value={status}>
                    {status}
                  </option>
                ))}
              </select>
            </div>
          </div>
        )}
      </div>

      {/* Camera List */}
      <div className="flex-1 overflow-y-auto p-2">
        {loading ? (
          <div className="flex items-center justify-center py-8 text-gray-500 dark:text-text-secondary">
            <Circle className="w-5 h-5 animate-spin" />
          </div>
        ) : filteredCameras.length === 0 ? (
          <div className="text-center py-8 text-gray-500 dark:text-text-secondary text-sm">
            No cameras found
          </div>
        ) : (
          <div className="space-y-1">
            {filteredCameras.map((camera) => {
              const isSelected = selectedCameras.some(
                (c) => c.id === camera.id
              );
              const isOnline = camera.status === 'ONLINE';

              return (
                <button
                  key={camera.id}
                  onClick={() => toggleCamera(camera)}
                  className={cn(
                    'w-full p-3 rounded-lg text-left transition-colors',
                    isSelected
                      ? 'bg-primary-50 dark:bg-primary-500/10 border-2 border-primary-500'
                      : 'bg-white dark:bg-dark-secondary border-2 border-transparent hover:bg-gray-50 dark:hover:bg-dark-surface'
                  )}
                >
                  <div className="flex items-start gap-2">
                    {isOnline ? (
                      <Video className="w-5 h-5 text-green-600 dark:text-green-400 flex-shrink-0 mt-0.5" />
                    ) : (
                      <VideoOff className="w-5 h-5 text-gray-400 dark:text-text-muted flex-shrink-0 mt-0.5" />
                    )}
                    <div className="flex-1 min-w-0">
                      <p className="font-medium text-sm text-gray-900 dark:text-text-primary truncate">
                        {camera.name}
                      </p>
                      {camera.name_ar && (
                        <p className="text-xs text-gray-500 dark:text-text-muted truncate">
                          {camera.name_ar}
                        </p>
                      )}
                      <div className="flex items-center gap-2 mt-1">
                        <span
                          className={cn(
                            'inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium',
                            isOnline
                              ? 'bg-green-100 dark:bg-green-500/20 text-green-700 dark:text-green-400'
                              : 'bg-gray-100 dark:bg-gray-500/20 text-gray-700 dark:text-text-secondary'
                          )}
                        >
                          <span
                            className={cn(
                              'w-1.5 h-1.5 rounded-full',
                              isOnline ? 'bg-green-500 dark:bg-green-400' : 'bg-gray-400 dark:bg-text-muted'
                            )}
                          />
                          {camera.status}
                        </span>
                        <span className="text-xs text-gray-500 dark:text-text-muted">
                          {camera.source.replace('_', ' ')}
                        </span>
                      </div>
                    </div>
                  </div>
                </button>
              );
            })}
          </div>
        )}
      </div>

      {/* Footer */}
      <div className="p-4 border-t border-gray-200 dark:border-dark-border bg-gray-50 dark:bg-dark-secondary">
        <p className="text-sm text-gray-600 dark:text-text-secondary">
          <span className="font-medium">{selectedCameras.length}</span> of{' '}
          <span className="font-medium">{filteredCameras.length}</span> selected
        </p>
      </div>
    </div>
  );
}
