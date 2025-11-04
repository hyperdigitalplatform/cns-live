import React, { useEffect, useState } from 'react';
import { useCameraStore } from '@/stores/cameraStore';
import { useFolderStore } from '@/stores/folderStore';
import { CameraTreeView } from './CameraTreeView';
import { MilestoneCameraDiscovery } from './MilestoneCameraDiscovery';
import type { Camera, CameraSource, CameraStatus } from '@/types';
import {
  Search,
  Filter,
  FolderPlus,
  Settings,
  ChevronDown,
  Maximize2,
  Minimize2,
  Folder,
  Video,
  Download,
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
  const [isCollapsed, setIsCollapsed] = useState(false);

  // Dialog states
  const [showCreateFolderDialog, setShowCreateFolderDialog] = useState(false);
  const [showMilestoneDiscovery, setShowMilestoneDiscovery] = useState(false);
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
      createFolder(folderName.trim(), undefined, null);
      setFolderName('');
      setShowCreateFolderDialog(false);
    }
  };

  const handleMilestoneImport = async () => {
    // Refresh camera list after import
    await fetchCameras();
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
      <div className="w-12 h-full bg-white dark:bg-dark-sidebar border-r border-gray-200 dark:border-dark-border flex flex-col items-center py-4 gap-4">
        <button
          onClick={() => setIsCollapsed(false)}
          className="p-2 hover:bg-gray-100 dark:hover:bg-dark-surface rounded-lg transition-colors"
          title="Expand sidebar"
        >
          <Maximize2 className="w-5 h-5 text-gray-600 dark:text-text-secondary" />
        </button>
        <div className="flex-1" />
        <div className="text-xs text-gray-500 dark:text-text-muted writing-mode-vertical transform rotate-180">
          Cameras
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-full bg-white dark:bg-dark-sidebar border-r border-gray-200 dark:border-dark-border w-80">
      {/* Header */}
      <div className="p-4 border-b border-gray-200 dark:border-dark-border space-y-3">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-text-primary">Cameras</h2>
          <button
            onClick={() => setIsCollapsed(true)}
            className="p-1.5 hover:bg-gray-100 dark:hover:bg-dark-surface rounded-lg transition-colors"
            title="Collapse sidebar"
          >
            <Minimize2 className="w-4 h-4 text-gray-600 dark:text-text-secondary" />
          </button>
        </div>

        {/* Search */}
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 dark:text-text-muted" />
          <input
            type="text"
            placeholder="Search cameras or folders..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full pl-9 pr-3 py-2 border border-gray-300 dark:border-dark-border rounded-lg text-sm bg-white dark:bg-dark-surface text-gray-900 dark:text-text-primary placeholder:text-gray-400 dark:placeholder:text-text-muted focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          />
        </div>

        {/* Action buttons */}
        <div className="flex items-center gap-2">
          <button
            onClick={() => setShowFilters(!showFilters)}
            className={cn(
              'flex items-center gap-2 px-3 py-1.5 text-sm rounded-lg transition-colors flex-1',
              showFilters
                ? 'bg-blue-50 dark:bg-blue-900/30 text-blue-700 dark:text-blue-400 border border-blue-200 dark:border-blue-700'
                : 'bg-gray-100 dark:bg-dark-surface text-gray-700 dark:text-text-secondary hover:bg-gray-200 dark:hover:bg-dark-elevated'
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
          <button
            onClick={() => setShowMilestoneDiscovery(true)}
            className="flex items-center gap-1.5 px-3 py-1.5 bg-purple-600 dark:bg-purple-700 text-white hover:bg-purple-700 dark:hover:bg-purple-600 rounded-lg transition-colors text-sm font-medium"
            title="Import cameras from Milestone XProtect"
          >
            <Download className="w-4 h-4" />
          </button>
        </div>

        {/* Filters */}
        {showFilters && (
          <div className="space-y-3 pt-3 border-t border-gray-200 dark:border-dark-border">
            <div>
              <label className="block text-xs font-medium text-gray-700 dark:text-text-secondary mb-1.5">
                Source
              </label>
              <select
                value={sourceFilter}
                onChange={(e) =>
                  setSourceFilter(e.target.value as CameraSource | 'ALL')
                }
                className="w-full px-3 py-2 border border-gray-300 dark:border-dark-border rounded-lg text-sm bg-white dark:bg-dark-surface text-gray-900 dark:text-text-primary focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                {sources.map((source) => (
                  <option key={source} value={source}>
                    {source.replace('_', ' ')}
                  </option>
                ))}
              </select>
            </div>

            <div>
              <label className="block text-xs font-medium text-gray-700 dark:text-text-secondary mb-1.5">
                Status
              </label>
              <select
                value={statusFilter}
                onChange={(e) =>
                  setStatusFilter(e.target.value as CameraStatus | 'ALL')
                }
                className="w-full px-3 py-2 border border-gray-300 dark:border-dark-border rounded-lg text-sm bg-white dark:bg-dark-surface text-gray-900 dark:text-text-primary focus:outline-none focus:ring-2 focus:ring-blue-500"
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

        {/* Tree/List actions - Always visible */}
        <div className="flex items-center gap-2 text-xs">
          <button
            onClick={expandAllFolders}
            className="px-2 py-1 text-blue-600 dark:text-blue-400 hover:bg-blue-50 dark:hover:bg-blue-900/30 rounded transition-colors"
          >
            Expand All
          </button>
          <button
            onClick={collapseAllFolders}
            className="px-2 py-1 text-blue-600 dark:text-blue-400 hover:bg-blue-50 dark:hover:bg-blue-900/30 rounded transition-colors"
          >
            Collapse All
          </button>
          <div className="flex-1" />
          <button
            onClick={handleCreateRootFolder}
            className="flex items-center gap-1.5 px-2 py-1 bg-blue-600 dark:bg-blue-700 text-white hover:bg-blue-700 dark:hover:bg-blue-600 rounded transition-colors text-xs font-medium"
          >
            <FolderPlus className="w-3.5 h-3.5" />
            <span>Add Folder</span>
          </button>
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-y-auto p-3">
        {loading ? (
          <div className="flex items-center justify-center py-12 text-gray-500 dark:text-text-muted">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600" />
          </div>
        ) : (
          <CameraTreeView
            folderTrees={folderTrees}
            unorganizedCameras={unorganizedCameras}
            allCameras={cameras}
            onCameraSelect={onCameraDoubleClick}
            onCameraDragStart={onCameraDragStart}
            selectedCameraId={selectedCameraId}
            searchQuery={searchQuery}
          />
        )}
      </div>

      {/* Footer */}
      <div className="p-4 border-t border-gray-200 dark:border-dark-border bg-gray-50 dark:bg-dark-secondary">
        <div className="flex items-center justify-between text-sm">
          <p className="text-gray-600 dark:text-text-secondary">
            <span className="font-semibold text-gray-900 dark:text-text-primary">
              {filteredCameras.length}
            </span>{' '}
            camera{filteredCameras.length !== 1 ? 's' : ''}
          </p>
          <p className="text-gray-600 dark:text-text-secondary">
            <span className="font-semibold text-gray-900 dark:text-text-primary">
              {folders.length}
            </span>{' '}
            folder{folders.length !== 1 ? 's' : ''}
          </p>
        </div>
      </div>

      {/* Create Folder Dialog */}
      <Dialog open={showCreateFolderDialog} onOpenChange={setShowCreateFolderDialog}>
        <DialogContent onClose={() => setShowCreateFolderDialog(false)}>
          <DialogHeader>
            <DialogTitle>Create New Folder</DialogTitle>
            <DialogDescription>
              Add a new folder to organize your cameras.
            </DialogDescription>
          </DialogHeader>

          <DialogBody>
            <div className="space-y-4">
              <Input
                label="Folder Name"
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
            </div>
          </DialogBody>

          <DialogFooter>
            <Button
              variant="secondary"
              onClick={() => {
                setShowCreateFolderDialog(false);
                setFolderName('');
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

      {/* Milestone Camera Discovery Dialog */}
      <MilestoneCameraDiscovery
        open={showMilestoneDiscovery}
        onClose={() => setShowMilestoneDiscovery(false)}
        onImport={handleMilestoneImport}
      />
    </div>
  );
}
