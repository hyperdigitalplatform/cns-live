import React, { useState, useRef } from 'react';
import {
  ChevronRight,
  ChevronDown,
  Folder,
  FolderOpen,
  Video,
  VideoOff,
  Plus,
  Edit2,
  Trash2,
  MoreVertical,
  FolderPlus,
  Search,
  AlertTriangle,
} from 'lucide-react';
import { cn } from '@/utils/cn';
import type { CameraFolderTree, Camera, DragItem } from '@/types';
import { useFolderStore } from '@/stores/folderStore';
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

interface CameraTreeViewProps {
  folderTrees: CameraFolderTree[];
  unorganizedCameras: Camera[];
  allCameras?: Camera[];
  onCameraSelect?: (camera: Camera) => void;
  onCameraDragStart?: (camera: Camera, folderId: string | null) => void;
  selectedCameraId?: string | null;
  searchQuery?: string;
  viewMode?: 'tree' | 'list';
}

export function CameraTreeView({
  folderTrees,
  unorganizedCameras,
  allCameras = [],
  onCameraSelect,
  onCameraDragStart,
  selectedCameraId,
  searchQuery = '',
  viewMode = 'tree',
}: CameraTreeViewProps) {
  const {
    toggleFolderExpanded,
    setSelectedFolder,
    selectedFolderId,
    createFolder,
    updateFolder,
    deleteFolder,
    removeCameraFromFolder,
    moveCameraBetweenFolders,
    moveFolder,
    addCameraToFolder,
  } = useFolderStore();

  const [dragOver, setDragOver] = useState<string | null>(null);
  const [contextMenu, setContextMenu] = useState<{
    x: number;
    y: number;
    type: 'folder' | 'camera';
    id: string;
    folderId?: string;
  } | null>(null);
  const [editingFolderId, setEditingFolderId] = useState<string | null>(null);
  const [editingName, setEditingName] = useState('');

  // Dialog states
  const [showCreateSubfolderDialog, setShowCreateSubfolderDialog] = useState(false);
  const [showRenameDialog, setShowRenameDialog] = useState(false);
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);
  const [showAddCameraDialog, setShowAddCameraDialog] = useState(false);
  const [currentFolderId, setCurrentFolderId] = useState<string | null>(null);
  const [folderName, setFolderName] = useState('');
  const [folderNameAr, setFolderNameAr] = useState('');
  const [deleteConfirmName, setDeleteConfirmName] = useState('');
  const [searchCameraQuery, setSearchCameraQuery] = useState('');

  const handleFolderClick = (folderId: string) => {
    toggleFolderExpanded(folderId);
    setSelectedFolder(folderId);
  };

  const handleCameraDoubleClick = (camera: Camera, event: React.MouseEvent) => {
    event.stopPropagation();
    onCameraSelect?.(camera);
  };

  const handleCameraDragStart = (
    event: React.DragEvent,
    camera: Camera,
    folderId: string | null
  ) => {
    const dragItem: DragItem = {
      type: 'camera',
      id: camera.id,
      sourceFolder: folderId,
      data: camera,
    };
    event.dataTransfer.setData('application/json', JSON.stringify(dragItem));
    event.dataTransfer.effectAllowed = 'move';
    onCameraDragStart?.(camera, folderId);
  };

  const handleFolderDragStart = (
    event: React.DragEvent,
    folder: CameraFolderTree
  ) => {
    const dragItem: DragItem = {
      type: 'folder',
      id: folder.id,
      sourceFolder: folder.parent_id,
      data: folder,
    };
    event.dataTransfer.setData('application/json', JSON.stringify(dragItem));
    event.dataTransfer.effectAllowed = 'move';
  };

  const handleDragOver = (event: React.DragEvent, folderId: string) => {
    event.preventDefault();
    event.stopPropagation();
    setDragOver(folderId);
  };

  const handleDragLeave = (event: React.DragEvent) => {
    event.preventDefault();
    event.stopPropagation();
    setDragOver(null);
  };

  const handleDrop = (
    event: React.DragEvent,
    targetFolderId: string | null
  ) => {
    event.preventDefault();
    event.stopPropagation();
    setDragOver(null);

    try {
      const data = event.dataTransfer.getData('application/json');
      if (!data) return;

      const dragItem: DragItem = JSON.parse(data);

      if (dragItem.type === 'camera' && dragItem.id) {
        moveCameraBetweenFolders(
          dragItem.id,
          dragItem.sourceFolder || null,
          targetFolderId
        );
      } else if (dragItem.type === 'folder' && dragItem.id) {
        if (dragItem.id !== targetFolderId) {
          moveFolder(dragItem.id, targetFolderId);
        }
      }
    } catch (error) {
      console.error('Drop failed:', error);
    }
  };

  const handleContextMenu = (
    event: React.MouseEvent,
    type: 'folder' | 'camera',
    id: string,
    folderId?: string
  ) => {
    event.preventDefault();
    event.stopPropagation();
    setContextMenu({
      x: event.clientX,
      y: event.clientY,
      type,
      id,
      folderId,
    });
  };

  const handleCreateSubfolder = (parentId: string) => {
    setCurrentFolderId(parentId);
    setFolderName('');
    setFolderNameAr('');
    setShowCreateSubfolderDialog(true);
    setContextMenu(null);
  };

  const handleConfirmCreateSubfolder = () => {
    if (folderName.trim() && currentFolderId) {
      createFolder(folderName.trim(), undefined, currentFolderId);
      setFolderName('');
      setShowCreateSubfolderDialog(false);
      setCurrentFolderId(null);
    }
  };

  const handleRenameFolder = (folderId: string, name: string, nameAr?: string) => {
    const folder = [...folderTrees].find(f => findFolderById(f, folderId));
    setCurrentFolderId(folderId);
    setFolderName(name);
    setFolderNameAr(nameAr || '');
    setShowRenameDialog(true);
    setContextMenu(null);
  };

  const handleConfirmRename = () => {
    if (folderName.trim() && currentFolderId) {
      updateFolder(currentFolderId, { name: folderName.trim() });
      setFolderName('');
      setShowRenameDialog(false);
      setCurrentFolderId(null);
    }
  };

  const handleDeleteFolder = (folderId: string, folderDisplayName: string) => {
    setCurrentFolderId(folderId);
    setFolderName(folderDisplayName);
    setDeleteConfirmName('');
    setShowDeleteDialog(true);
    setContextMenu(null);
  };

  const handleConfirmDelete = () => {
    if (deleteConfirmName === folderName && currentFolderId) {
      deleteFolder(currentFolderId);
      setDeleteConfirmName('');
      setShowDeleteDialog(false);
      setCurrentFolderId(null);
      setFolderName('');
    }
  };

  const handleAddCamera = (folderId: string) => {
    setCurrentFolderId(folderId);
    setSearchCameraQuery('');
    setShowAddCameraDialog(true);
    setContextMenu(null);
  };

  const handleConfirmAddCamera = (cameraIds: string[]) => {
    if (currentFolderId && cameraIds.length > 0) {
      cameraIds.forEach(cameraId => {
        addCameraToFolder(cameraId, currentFolderId);
      });
      setShowAddCameraDialog(false);
      setCurrentFolderId(null);
      setSearchCameraQuery('');
    }
  };

  // Helper function to find folder by ID in tree
  const findFolderById = (folder: CameraFolderTree, id: string): CameraFolderTree | null => {
    if (folder.id === id) return folder;
    for (const child of folder.children) {
      const found = findFolderById(child, id);
      if (found) return found;
    }
    return null;
  };

  const handleRemoveCamera = (cameraId: string, folderId: string) => {
    if (confirm('Remove camera from this folder?')) {
      removeCameraFromFolder(cameraId, folderId);
    }
    setContextMenu(null);
  };

  const handleSaveEdit = (folderId: string) => {
    if (editingName.trim()) {
      updateFolder(folderId, { name: editingName.trim() });
    }
    setEditingFolderId(null);
    setEditingName('');
  };

  const handleCancelEdit = () => {
    setEditingFolderId(null);
    setEditingName('');
  };

  const filterTree = (tree: CameraFolderTree): CameraFolderTree | null => {
    if (!searchQuery) return tree;

    const query = searchQuery.toLowerCase();
    const matchesFolder = tree.name.toLowerCase().includes(query) ||
                         tree.name_ar?.includes(query);
    const matchingCameras = tree.cameras.filter(
      (cam) =>
        cam.name.toLowerCase().includes(query) ||
        cam.name_ar?.includes(query) ||
        cam.id.toLowerCase().includes(query)
    );
    const matchingChildren = tree.children
      .map((child) => filterTree(child))
      .filter((c): c is CameraFolderTree => c !== null);

    if (matchesFolder || matchingCameras.length > 0 || matchingChildren.length > 0) {
      return {
        ...tree,
        cameras: matchingCameras,
        children: matchingChildren,
        expanded: searchQuery ? true : tree.expanded,
      };
    }

    return null;
  };

  const renderCamera = (camera: Camera, folderId: string | null, depth: number = 0) => {
    const isOnline = camera.status === 'ONLINE';
    const isSelected = selectedCameraId === camera.id;

    return (
      <div
        key={camera.id}
        draggable
        onDragStart={(e) => handleCameraDragStart(e, camera, folderId)}
        onDoubleClick={(e) => handleCameraDoubleClick(camera, e)}
        onContextMenu={(e) => handleContextMenu(e, 'camera', camera.id, folderId || undefined)}
        className={cn(
          'flex items-center gap-2 px-3 py-2 rounded-md cursor-pointer transition-colors group relative',
          isSelected
            ? 'bg-blue-100 border border-blue-500'
            : 'hover:bg-gray-100'
        )}
        style={{ paddingLeft: `${depth * 20 + 28}px` }}
      >
        {/* Tree connector lines for cameras */}
        {depth > 0 && (
          <>
            <div
              className="absolute left-0 top-0 bottom-0 w-px bg-gray-300"
              style={{ left: `${(depth - 1) * 20 + 18}px` }}
            />
            <div
              className="absolute top-1/2 w-3 h-px bg-gray-300"
              style={{ left: `${(depth - 1) * 20 + 18}px` }}
            />
          </>
        )}
        {isOnline ? (
          <Video className="w-4 h-4 text-green-600 flex-shrink-0" />
        ) : (
          <VideoOff className="w-4 h-4 text-gray-400 flex-shrink-0" />
        )}
        <span className="flex-1 text-sm font-medium text-gray-900 truncate text-left">
          {camera.name}
        </span>
        <span
          className={cn(
            'w-2 h-2 rounded-full flex-shrink-0',
            isOnline ? 'bg-green-500' : 'bg-gray-400'
          )}
        />
      </div>
    );
  };

  const renderFolder = (folder: CameraFolderTree) => {
    const isExpanded = folder.expanded;
    const isSelected = selectedFolderId === folder.id;
    const isDraggedOver = dragOver === folder.id;
    const isEditing = editingFolderId === folder.id;

    const filteredFolder = filterTree(folder);
    if (!filteredFolder) return null;

    return (
      <div key={folder.id} className="select-none">
        <div
          draggable={!isEditing}
          onDragStart={(e) => !isEditing && handleFolderDragStart(e, folder)}
          onDragOver={(e) => handleDragOver(e, folder.id)}
          onDragLeave={handleDragLeave}
          onDrop={(e) => handleDrop(e, folder.id)}
          onContextMenu={(e) => handleContextMenu(e, 'folder', folder.id)}
          className={cn(
            'flex items-center gap-2 px-2 py-2 rounded-md cursor-pointer transition-colors group relative',
            isSelected && 'bg-blue-50',
            isDraggedOver && 'bg-blue-100 ring-2 ring-blue-500',
            'hover:bg-gray-100'
          )}
          style={{ paddingLeft: `${folder.depth * 20 + 8}px` }}
        >
          {/* Tree connector lines */}
          {folder.depth > 0 && (
            <>
              <div
                className="absolute left-0 top-0 bottom-0 w-px bg-gray-300"
                style={{ left: `${(folder.depth - 1) * 20 + 18}px` }}
              />
              <div
                className="absolute top-1/2 w-3 h-px bg-gray-300"
                style={{ left: `${(folder.depth - 1) * 20 + 18}px` }}
              />
            </>
          )}
          <button
            onClick={() => handleFolderClick(folder.id)}
            className="flex items-center gap-1 flex-1 min-w-0"
          >
            {isExpanded ? (
              <ChevronDown className="w-4 h-4 text-gray-500 flex-shrink-0" />
            ) : (
              <ChevronRight className="w-4 h-4 text-gray-500 flex-shrink-0" />
            )}
            {isExpanded ? (
              <FolderOpen className="w-4 h-4 text-blue-500 flex-shrink-0" />
            ) : (
              <Folder className="w-4 h-4 text-blue-500 flex-shrink-0" />
            )}
            {isEditing ? (
              <input
                type="text"
                value={editingName}
                onChange={(e) => setEditingName(e.target.value)}
                onBlur={() => handleSaveEdit(folder.id)}
                onKeyDown={(e) => {
                  if (e.key === 'Enter') handleSaveEdit(folder.id);
                  if (e.key === 'Escape') handleCancelEdit();
                }}
                onClick={(e) => e.stopPropagation()}
                autoFocus
                className="flex-1 px-2 py-1 text-sm border border-blue-500 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            ) : (
              <span className="text-sm font-medium text-gray-900 truncate text-left">
                {filteredFolder.name}
              </span>
            )}
          </button>
          <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
            <span className="text-xs text-gray-500 bg-gray-200 px-2 py-0.5 rounded-full">
              {filteredFolder.cameras.length}
            </span>
            <button
              onClick={(e) => {
                e.stopPropagation();
                handleCreateSubfolder(folder.id);
              }}
              className="p-1 hover:bg-gray-200 rounded"
              title="Add subfolder"
            >
              <FolderPlus className="w-3.5 h-3.5 text-gray-600" />
            </button>
          </div>
        </div>

        {isExpanded && (
          <div className="mt-1 space-y-0.5">
            {filteredFolder.cameras.map((camera) =>
              renderCamera(camera, folder.id, folder.depth + 1)
            )}
            {filteredFolder.children.map((child) => renderFolder(child))}
          </div>
        )}
      </div>
    );
  };

  // Close context menu on outside click
  React.useEffect(() => {
    const handleClick = () => setContextMenu(null);
    if (contextMenu) {
      document.addEventListener('click', handleClick);
      return () => document.removeEventListener('click', handleClick);
    }
  }, [contextMenu]);

  return (
    <div className="space-y-1">
      {folderTrees.map((tree) => renderFolder(tree))}

      {/* Unorganized cameras section - Now part of tree */}
      {unorganizedCameras.length > 0 && (
        <div
          className={cn(
            dragOver === 'unorganized' && 'bg-blue-50 ring-2 ring-blue-500'
          )}
          onDragOver={(e) => handleDragOver(e, 'unorganized')}
          onDragLeave={handleDragLeave}
          onDrop={(e) => handleDrop(e, null)}
        >
          <div className="flex items-center gap-2 px-2 py-2 rounded-md hover:bg-gray-100 cursor-pointer transition-colors group"
            style={{ paddingLeft: '8px' }}
          >
            <ChevronRight className="w-4 h-4 text-gray-400 flex-shrink-0 invisible" />
            <Folder className="w-4 h-4 text-gray-400 flex-shrink-0" />
            <span className="text-sm font-medium text-gray-600 truncate text-left">
              Unorganized
            </span>
            <div className="flex items-center gap-1 opacity-100 transition-opacity ml-auto">
              <span className="text-xs text-gray-500 bg-gray-200 px-2 py-0.5 rounded-full">
                {unorganizedCameras.length}
              </span>
            </div>
          </div>
          <div className="space-y-0.5">
            {unorganizedCameras.map((camera) => renderCamera(camera, null, 1))}
          </div>
        </div>
      )}

      {/* Context menu */}
      {contextMenu && (
        <div
          className="fixed bg-white rounded-lg shadow-lg border border-gray-200 py-1 z-50 min-w-[160px]"
          style={{
            left: contextMenu.x,
            top: contextMenu.y,
          }}
          onClick={(e) => e.stopPropagation()}
        >
          {contextMenu.type === 'folder' ? (
            <>
              <button
                onClick={() => handleAddCamera(contextMenu.id)}
                className="w-full px-4 py-2 text-left text-sm hover:bg-gray-100 flex items-center gap-2"
              >
                <Plus className="w-4 h-4" />
                Add Camera
              </button>
              <button
                onClick={() => handleCreateSubfolder(contextMenu.id)}
                className="w-full px-4 py-2 text-left text-sm hover:bg-gray-100 flex items-center gap-2"
              >
                <FolderPlus className="w-4 h-4" />
                Add Subfolder
              </button>
              <button
                onClick={() => {
                  const findFolder = (folders: CameraFolderTree[]): CameraFolderTree | null => {
                    for (const folder of folders) {
                      if (folder.id === contextMenu.id) return folder;
                      const found = findFolder(folder.children);
                      if (found) return found;
                    }
                    return null;
                  };
                  const folder = findFolder(folderTrees);
                  if (folder) handleRenameFolder(folder.id, folder.name, folder.name_ar);
                }}
                className="w-full px-4 py-2 text-left text-sm hover:bg-gray-100 flex items-center gap-2"
              >
                <Edit2 className="w-4 h-4" />
                Rename
              </button>
              <button
                onClick={() => {
                  const findFolder = (folders: CameraFolderTree[]): CameraFolderTree | null => {
                    for (const folder of folders) {
                      if (folder.id === contextMenu.id) return folder;
                      const found = findFolder(folder.children);
                      if (found) return found;
                    }
                    return null;
                  };
                  const folder = findFolder(folderTrees);
                  if (folder) handleDeleteFolder(folder.id, folder.name);
                }}
                className="w-full px-4 py-2 text-left text-sm hover:bg-gray-100 flex items-center gap-2 text-red-600"
              >
                <Trash2 className="w-4 h-4" />
                Delete
              </button>
            </>
          ) : (
            <>
              {contextMenu.folderId && (
                <button
                  onClick={() =>
                    handleRemoveCamera(contextMenu.id, contextMenu.folderId!)
                  }
                  className="w-full px-4 py-2 text-left text-sm hover:bg-gray-100 flex items-center gap-2 text-red-600"
                >
                  <Trash2 className="w-4 h-4" />
                  Remove from Folder
                </button>
              )}
            </>
          )}
        </div>
      )}

      {/* Create Subfolder Dialog */}
      <Dialog open={showCreateSubfolderDialog} onOpenChange={setShowCreateSubfolderDialog}>
        <DialogContent onClose={() => setShowCreateSubfolderDialog(false)}>
          <DialogHeader>
            <DialogTitle>Create Subfolder</DialogTitle>
            <DialogDescription>
              Add a subfolder to organize cameras within this folder.
            </DialogDescription>
          </DialogHeader>

          <DialogBody>
            <div className="space-y-4">
              <Input
                label="Subfolder Name"
                placeholder="e.g., Highway Cameras"
                value={folderName}
                onChange={(e) => setFolderName(e.target.value)}
                autoFocus
                onKeyDown={(e) => {
                  if (e.key === 'Enter' && folderName.trim()) {
                    handleConfirmCreateSubfolder();
                  }
                }}
              />
            </div>
          </DialogBody>

          <DialogFooter>
            <Button
              variant="secondary"
              onClick={() => {
                setShowCreateSubfolderDialog(false);
                setFolderName('');
              }}
            >
              Cancel
            </Button>
            <Button
              variant="primary"
              onClick={handleConfirmCreateSubfolder}
              disabled={!folderName.trim()}
            >
              Create Subfolder
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Rename Folder Dialog */}
      <Dialog open={showRenameDialog} onOpenChange={setShowRenameDialog}>
        <DialogContent onClose={() => setShowRenameDialog(false)}>
          <DialogHeader>
            <DialogTitle>Rename Folder</DialogTitle>
            <DialogDescription>
              Update the folder name. Changes will be saved immediately.
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
                    handleConfirmRename();
                  }
                }}
              />
            </div>
          </DialogBody>

          <DialogFooter>
            <Button
              variant="secondary"
              onClick={() => {
                setShowRenameDialog(false);
                setFolderName('');
              }}
            >
              Cancel
            </Button>
            <Button
              variant="primary"
              onClick={handleConfirmRename}
              disabled={!folderName.trim()}
            >
              Save Changes
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Folder Dialog */}
      <Dialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
        <DialogContent onClose={() => setShowDeleteDialog(false)}>
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2 text-red-600">
              <AlertTriangle className="w-5 h-5" />
              Delete Folder
            </DialogTitle>
            <DialogDescription>
              This action cannot be undone. Cameras in this folder will be moved to the parent folder.
            </DialogDescription>
          </DialogHeader>

          <DialogBody>
            <div className="space-y-4">
              <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
                <p className="text-sm text-yellow-800">
                  <strong>Warning:</strong> You are about to delete the folder <strong>"{folderName}"</strong>.
                </p>
              </div>

              <Input
                label={`To confirm, type the folder name: ${folderName}`}
                placeholder={folderName}
                value={deleteConfirmName}
                onChange={(e) => setDeleteConfirmName(e.target.value)}
                autoFocus
                error={deleteConfirmName && deleteConfirmName !== folderName ? 'Folder name does not match' : undefined}
              />
            </div>
          </DialogBody>

          <DialogFooter>
            <Button
              variant="secondary"
              onClick={() => {
                setShowDeleteDialog(false);
                setDeleteConfirmName('');
                setFolderName('');
              }}
            >
              Cancel
            </Button>
            <Button
              variant="danger"
              onClick={handleConfirmDelete}
              disabled={deleteConfirmName !== folderName}
            >
              Delete Folder
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Add Camera Dialog */}
      <Dialog open={showAddCameraDialog} onOpenChange={setShowAddCameraDialog}>
        <DialogContent onClose={() => setShowAddCameraDialog(false)}>
          <DialogHeader>
            <DialogTitle>Add Camera to Folder</DialogTitle>
            <DialogDescription>
              Select cameras from the list below to add to this folder.
            </DialogDescription>
          </DialogHeader>

          <DialogBody>
            <div className="space-y-4">
              <Input
                label="Search Cameras"
                placeholder="Search by name or ID..."
                value={searchCameraQuery}
                onChange={(e) => setSearchCameraQuery(e.target.value)}
                autoFocus
              />

              <div className="max-h-80 overflow-y-auto border border-gray-200 rounded-lg">
                {allCameras
                  .filter(camera =>
                    !currentFolderId ||
                    !folderTrees.some(ft =>
                      findFolderById(ft, currentFolderId)?.camera_ids.includes(camera.id)
                    )
                  )
                  .filter(camera =>
                    searchCameraQuery === '' ||
                    camera.name.toLowerCase().includes(searchCameraQuery.toLowerCase()) ||
                    camera.name_ar?.includes(searchCameraQuery) ||
                    camera.id.toLowerCase().includes(searchCameraQuery.toLowerCase())
                  )
                  .map(camera => {
                    const [selectedCameras, setSelectedCameras] = React.useState<string[]>([]);

                    return (
                      <label
                        key={camera.id}
                        className="flex items-center gap-3 px-4 py-3 hover:bg-gray-50 cursor-pointer border-b border-gray-100 last:border-0"
                      >
                        <input
                          type="checkbox"
                          className="w-4 h-4 text-blue-600 border-gray-300 rounded focus:ring-blue-500"
                          onChange={(e) => {
                            if (e.target.checked) {
                              addCameraToFolder(camera.id, currentFolderId!);
                            } else {
                              removeCameraFromFolder(camera.id, currentFolderId!);
                            }
                          }}
                        />
                        <div className="flex-1 min-w-0">
                          <p className="text-sm font-medium text-gray-900 truncate">
                            {camera.name}
                          </p>
                          {camera.name_ar && (
                            <p className="text-xs text-gray-500 truncate">
                              {camera.name_ar}
                            </p>
                          )}
                        </div>
                        <span className={cn(
                          'text-xs px-2 py-0.5 rounded-full',
                          camera.status === 'ONLINE'
                            ? 'bg-green-100 text-green-700'
                            : 'bg-gray-100 text-gray-700'
                        )}>
                          {camera.status}
                        </span>
                      </label>
                    );
                  })}
                {allCameras.filter(camera =>
                  searchCameraQuery === '' ||
                  camera.name.toLowerCase().includes(searchCameraQuery.toLowerCase()) ||
                  camera.name_ar?.includes(searchCameraQuery) ||
                  camera.id.toLowerCase().includes(searchCameraQuery.toLowerCase())
                ).length === 0 && (
                  <div className="px-4 py-8 text-center text-gray-500 text-sm">
                    No cameras found
                  </div>
                )}
              </div>
            </div>
          </DialogBody>

          <DialogFooter>
            <Button
              variant="secondary"
              onClick={() => {
                setShowAddCameraDialog(false);
                setSearchCameraQuery('');
              }}
            >
              Done
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
