import React, { useEffect, useState } from 'react';
import { useCameraStore } from '@/stores/cameraStore';
import { useFolderStore } from '@/stores/folderStore';
import { CameraTreeView } from './CameraTreeView';
import type { Camera, CameraSource, CameraStatus } from '@/types';
import {
  Search,
  Filter,
  FolderPlus,
  Settings,
  ChevronDown,
  LayoutGrid,
  List,
  Maximize2,
  Minimize2,
  Folder,
  Video,
} from 'lucide-react';
import { cn } from '@/utils/cn';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogBody,
  DialogFooter,
  Button,
  Input,
} from './ui/Dialog';

interface CameraSidebarNewProps {
  onCameraDoubleClick?: (camera: Camera) => void;
  onCameraDragStart?: (camera: Camera, folderId: string | null) => void;
  selectedCameraId?: string | null;
}

export function CameraSidebarNew({
  onCameraDoubleClick,
  onCameraDragStart,
  selectedCameraId,
}: CameraSidebarNewProps) {
  const { cameras, loading, fetchCameras } = useCameraStore();
  const {
    buildFolderTree,
    createFolder,
    initializeDefaultFolders,
    expandAllFolders,
    collapseAllFolders,
    folders,
  } = useFolderStore();

  const [searchQuery, setSearchQuery] = useState('');
  const [sourceFilter, setSourceFilter] = useState<CameraSource | 'ALL'>('ALL');
  const [statusFilter, setStatusFilter] = useState<CameraStatus | 'ALL'>('ALL');
  const [showFilters, setShowFilters] = useState(false);
  const [viewMode, setViewMode] = useState<'tree' | 'list'>('tree');
  const [isCollapsed, setIsCollapsed] = useState(false);

  // Dialog states
  const [showCreateFolderDialog, setShowCreateFolderDialog] = useState(false);
  const [folderName, setFolderName] = useState('');
  const [folderNameAr, setFolderNameAr] = useState('');

  useEffect(() => {
    fetchCameras();
    if (folders.length === 0) {
      initializeDefaultFolders();
    }
  }, [fetchCameras, initializeDefaultFolders, folders.length]);

  // Filter cameras
  const filteredCameras = cameras.filter((camera) => {
    const matchesSearch =
      camera.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      camera.name_ar?.includes(searchQuery) ||
      camera.id.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesSource =
      sourceFilter === 'ALL' || camera.source === sourceFilter;
    const matchesStatus =
      statusFilter === 'ALL' || camera.status === statusFilter;

    return matchesSearch && matchesSource && matchesStatus;
  });

  // Build folder tree with filtered cameras
  const folderTrees = buildFolderTree(filteredCameras);

  // Get cameras not in any folder (unorganized)
  const allFolderCameraIds = new Set(
    folders.flatMap((folder) => folder.camera_ids)
  );
  const unorganizedCameras = filteredCameras.filter(
    (camera) => !allFolderCameraIds.has(camera.id)
  );

  const handleCreateRootFolder = () => {
    setShowCreateFolderDialog(true);
  };

  const handleCreateFolder = () => {
    if (folderName.trim()) {
      createFolder(folderName.trim(), folderNameAr.trim() || undefined, null);
      setFolderName('');
      setFolderNameAr('');
      setShowCreateFolderDialog(false);
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

  if (isCollapsed) {
    return (
      <div className="w-12 h-full bg-white border-r border-gray-200 flex flex-col items-center py-4 gap-4">
        <button
          onClick={() => setIsCollapsed(false)}
          className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
          title="Expand sidebar"
        >
          <Maximize2 className="w-5 h-5 text-gray-600" />
        </button>
        <div className="flex-1" />
        <div className="text-xs text-gray-500 writing-mode-vertical transform rotate-180">
          Cameras
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-full bg-white border-r border-gray-200 w-80">
      {/* Header */}
      <div className="p-4 border-b border-gray-200 space-y-3">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold text-gray-900">Cameras</h2>
          <div className="flex items-center gap-2">
            {/* View mode toggle */}
            <div className="flex items-center bg-gray-100 rounded-lg p-0.5">
              <button
                onClick={() => setViewMode('tree')}
                className={cn(
                  'p-1.5 rounded transition-colors',
                  viewMode === 'tree'
                    ? 'bg-white shadow-sm'
                    : 'hover:bg-gray-200'
                )}
                title="Tree view"
              >
                <LayoutGrid className="w-4 h-4" />
              </button>
              <button
                onClick={() => setViewMode('list')}
                className={cn(
                  'p-1.5 rounded transition-colors',
                  viewMode === 'list'
                    ? 'bg-white shadow-sm'
                    : 'hover:bg-gray-200'
                )}
                title="List view"
              >
                <List className="w-4 h-4" />
              </button>
            </div>

            <button
              onClick={() => setIsCollapsed(true)}
              className="p-1.5 hover:bg-gray-100 rounded-lg transition-colors"
              title="Collapse sidebar"
            >
              <Minimize2 className="w-4 h-4 text-gray-600" />
            </button>
          </div>
        </div>

        {/* Search */}
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
          <input
            type="text"
            placeholder="Search cameras or folders..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full pl-9 pr-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          />
        </div>

        {/* Action buttons */}
        <div className="flex items-center gap-2">
          <button
            onClick={() => setShowFilters(!showFilters)}
            className={cn(
              'flex items-center gap-2 px-3 py-1.5 text-sm rounded-lg transition-colors flex-1',
              showFilters
                ? 'bg-blue-50 text-blue-700 border border-blue-200'
                : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
            )}
          >
            <Filter className="w-4 h-4" />
            <span>Filters</span>
            <ChevronDown
              className={cn(
                'w-4 h-4 ml-auto transition-transform',
                showFilters && 'rotate-180'
              )}
            />
          </button>
        </div>

        {/* Filters */}
        {showFilters && (
          <div className="space-y-3 pt-3 border-t border-gray-200">
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1.5">
                Source
              </label>
              <select
                value={sourceFilter}
                onChange={(e) =>
                  setSourceFilter(e.target.value as CameraSource | 'ALL')
                }
                className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                {sources.map((source) => (
                  <option key={source} value={source}>
                    {source.replace('_', ' ')}
                  </option>
                ))}
              </select>
            </div>

            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1.5">
                Status
              </label>
              <select
                value={statusFilter}
                onChange={(e) =>
                  setStatusFilter(e.target.value as CameraStatus | 'ALL')
                }
                className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
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

        {/* Tree actions - Always visible when in tree mode */}
        {viewMode === 'tree' && (
          <div className="flex items-center gap-2 text-xs">
            <button
              onClick={expandAllFolders}
              className="px-2 py-1 text-blue-600 hover:bg-blue-50 rounded transition-colors"
            >
              Expand All
            </button>
            <button
              onClick={collapseAllFolders}
              className="px-2 py-1 text-blue-600 hover:bg-blue-50 rounded transition-colors"
            >
              Collapse All
            </button>
            <div className="flex-1" />
            <button
              onClick={handleCreateRootFolder}
              className="flex items-center gap-1.5 px-2 py-1 bg-blue-600 text-white hover:bg-blue-700 rounded transition-colors text-xs font-medium"
            >
              <FolderPlus className="w-3.5 h-3.5" />
              <span>Add Folder</span>
            </button>
          </div>
        )}
      </div>

      {/* Content */}
      <div className="flex-1 overflow-y-auto p-3">
        {loading ? (
          <div className="flex items-center justify-center py-12 text-gray-500">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600" />
          </div>
        ) : viewMode === 'tree' ? (
          <CameraTreeView
            folderTrees={folderTrees}
            unorganizedCameras={unorganizedCameras}
            allCameras={cameras}
            onCameraSelect={onCameraDoubleClick}
            onCameraDragStart={onCameraDragStart}
            selectedCameraId={selectedCameraId}
            searchQuery={searchQuery}
          />
        ) : (
          // List view with folder structure
          <div className="space-y-2">
            {folderTrees.map((folder) => (
              <div key={folder.id} className="space-y-1">
                {/* Folder header */}
                <div className="px-2 py-1.5 bg-gray-100 rounded-md">
                  <div className="flex items-center gap-2">
                    <Folder className="w-4 h-4 text-gray-600" />
                    <span className="text-sm font-semibold text-gray-900">
                      {folder.name}
                    </span>
                    {folder.name_ar && (
                      <span className="text-xs text-gray-500">({folder.name_ar})</span>
                    )}
                    <span className="ml-auto text-xs text-gray-500">
                      {folder.cameras.length}
                    </span>
                  </div>
                </div>

                {/* Cameras in folder */}
                {folder.cameras.map((camera) => {
                  const isOnline = camera.status === 'ONLINE';
                  const isSelected = selectedCameraId === camera.id;

                  return (
                    <div
                      key={camera.id}
                      draggable
                      onDragStart={(e) => onCameraDragStart?.(camera, folder.id)}
                      onDoubleClick={() => onCameraDoubleClick?.(camera)}
                      className={cn(
                        'ml-6 p-2 rounded-lg cursor-pointer transition-colors',
                        isSelected
                          ? 'bg-blue-100 border border-blue-500'
                          : 'hover:bg-gray-100 border border-transparent'
                      )}
                    >
                      <div className="flex items-start gap-2">
                        <Video className="w-4 h-4 text-gray-400 mt-0.5" />
                        <div className="flex-1 min-w-0">
                          <p className="font-medium text-sm text-gray-900 truncate">
                            {camera.name}
                          </p>
                          {camera.name_ar && (
                            <p className="text-xs text-gray-500 truncate">
                              {camera.name_ar}
                            </p>
                          )}
                          <div className="flex items-center gap-2 mt-1">
                            <span
                              className={cn(
                                'inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium',
                                isOnline
                                  ? 'bg-green-100 text-green-700'
                                  : 'bg-gray-100 text-gray-700'
                              )}
                            >
                              <span
                                className={cn(
                                  'w-1.5 h-1.5 rounded-full',
                                  isOnline ? 'bg-green-500' : 'bg-gray-400'
                                )}
                              />
                              {camera.status}
                            </span>
                          </div>
                        </div>
                      </div>
                    </div>
                  );
                })}
              </div>
            ))}

            {/* Unorganized cameras */}
            {unorganizedCameras.length > 0 && (
              <div className="space-y-1">
                <div className="px-2 py-1.5 bg-gray-100 rounded-md">
                  <div className="flex items-center gap-2">
                    <Folder className="w-4 h-4 text-gray-600" />
                    <span className="text-sm font-semibold text-gray-900">
                      Unorganized
                    </span>
                    <span className="ml-auto text-xs text-gray-500">
                      {unorganizedCameras.length}
                    </span>
                  </div>
                </div>

                {unorganizedCameras.map((camera) => {
                  const isOnline = camera.status === 'ONLINE';
                  const isSelected = selectedCameraId === camera.id;

                  return (
                    <div
                      key={camera.id}
                      draggable
                      onDragStart={(e) => onCameraDragStart?.(camera, null)}
                      onDoubleClick={() => onCameraDoubleClick?.(camera)}
                      className={cn(
                        'ml-6 p-2 rounded-lg cursor-pointer transition-colors',
                        isSelected
                          ? 'bg-blue-100 border border-blue-500'
                          : 'hover:bg-gray-100 border border-transparent'
                      )}
                    >
                      <div className="flex items-start gap-2">
                        <Video className="w-4 h-4 text-gray-400 mt-0.5" />
                        <div className="flex-1 min-w-0">
                          <p className="font-medium text-sm text-gray-900 truncate">
                            {camera.name}
                          </p>
                          {camera.name_ar && (
                            <p className="text-xs text-gray-500 truncate">
                              {camera.name_ar}
                            </p>
                          )}
                          <div className="flex items-center gap-2 mt-1">
                            <span
                              className={cn(
                                'inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium',
                                isOnline
                                  ? 'bg-green-100 text-green-700'
                                  : 'bg-gray-100 text-gray-700'
                              )}
                            >
                              <span
                                className={cn(
                                  'w-1.5 h-1.5 rounded-full',
                                  isOnline ? 'bg-green-500' : 'bg-gray-400'
                                )}
                              />
                              {camera.status}
                            </span>
                          </div>
                        </div>
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </div>
        )}
      </div>

      {/* Footer */}
      <div className="p-4 border-t border-gray-200 bg-gray-50">
        <div className="flex items-center justify-between text-sm">
          <p className="text-gray-600">
            <span className="font-semibold text-gray-900">
              {filteredCameras.length}
            </span>{' '}
            camera{filteredCameras.length !== 1 ? 's' : ''}
          </p>
          {viewMode === 'tree' && (
            <p className="text-gray-600">
              <span className="font-semibold text-gray-900">
                {folders.length}
              </span>{' '}
              folder{folders.length !== 1 ? 's' : ''}
            </p>
          )}
        </div>
      </div>

      {/* Create Folder Dialog */}
      <Dialog open={showCreateFolderDialog} onOpenChange={setShowCreateFolderDialog}>
        <DialogContent onClose={() => setShowCreateFolderDialog(false)}>
          <DialogHeader>
            <DialogTitle>Create New Folder</DialogTitle>
            <DialogDescription>
              Add a new folder to organize your cameras. You can provide both English and Arabic names.
            </DialogDescription>
          </DialogHeader>

          <DialogBody>
            <div className="space-y-4">
              <Input
                label="Folder Name (English)"
                placeholder="e.g., Traffic Cameras"
                value={folderName}
                onChange={(e) => setFolderName(e.target.value)}
                autoFocus
                onKeyDown={(e) => {
                  if (e.key === 'Enter' && folderName.trim()) {
                    handleCreateFolder();
                  }
                }}
              />

              <Input
                label="Folder Name (Arabic) - Optional"
                placeholder="مثال: كاميرات المرور"
                value={folderNameAr}
                onChange={(e) => setFolderNameAr(e.target.value)}
                dir="rtl"
                onKeyDown={(e) => {
                  if (e.key === 'Enter' && folderName.trim()) {
                    handleCreateFolder();
                  }
                }}
              />
            </div>
          </DialogBody>

          <DialogFooter>
            <Button
              variant="secondary"
              onClick={() => {
                setShowCreateFolderDialog(false);
                setFolderName('');
                setFolderNameAr('');
              }}
            >
              Cancel
            </Button>
            <Button
              variant="primary"
              onClick={handleCreateFolder}
              disabled={!folderName.trim()}
            >
              Create Folder
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
