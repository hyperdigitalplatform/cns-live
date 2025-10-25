import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { CameraFolder, CameraFolderTree, Camera } from '@/types';

interface FolderState {
  folders: CameraFolder[];
  expandedFolders: Set<string>;
  selectedFolderId: string | null;

  // Actions
  createFolder: (
    name: string,
    nameAr?: string,
    parentId?: string | null
  ) => CameraFolder;
  updateFolder: (
    id: string,
    updates: Partial<Omit<CameraFolder, 'id' | 'created_at' | 'updated_at'>>
  ) => void;
  deleteFolder: (id: string) => void;
  moveFolder: (folderId: string, newParentId: string | null) => void;

  addCameraToFolder: (cameraId: string, folderId: string) => void;
  removeCameraFromFolder: (cameraId: string, folderId: string) => void;
  moveCameraBetweenFolders: (
    cameraId: string,
    sourceFolderId: string | null,
    targetFolderId: string | null
  ) => void;

  toggleFolderExpanded: (folderId: string) => void;
  expandAllFolders: () => void;
  collapseAllFolders: () => void;
  setSelectedFolder: (folderId: string | null) => void;

  buildFolderTree: (cameras: Camera[]) => CameraFolderTree[];
  getFolderPath: (folderId: string) => CameraFolder[];
  getCamerasInFolder: (folderId: string, includeSubfolders?: boolean) => string[];

  // Initialize with default folders
  initializeDefaultFolders: () => void;
}

export const useFolderStore = create<FolderState>()(
  persist(
    (set, get) => ({
      folders: [],
      expandedFolders: new Set<string>(),
      selectedFolderId: null,

      createFolder: (name, nameAr, parentId = null) => {
        const newFolder: CameraFolder = {
          id: `folder_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
          name,
          name_ar: nameAr,
          parent_id: parentId,
          camera_ids: [],
          order: get().folders.filter((f) => f.parent_id === parentId).length,
          expanded: true,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        };

        set((state) => ({
          folders: [...state.folders, newFolder],
          expandedFolders: new Set([...state.expandedFolders, newFolder.id]),
        }));

        return newFolder;
      },

      updateFolder: (id, updates) => {
        set((state) => ({
          folders: state.folders.map((folder) =>
            folder.id === id
              ? { ...folder, ...updates, updated_at: new Date().toISOString() }
              : folder
          ),
        }));
      },

      deleteFolder: (id) => {
        const folder = get().folders.find((f) => f.id === id);
        if (!folder) return;

        // Move child folders to parent
        const childFolders = get().folders.filter((f) => f.parent_id === id);
        childFolders.forEach((child) => {
          get().moveFolder(child.id, folder.parent_id);
        });

        // Remove cameras from folder (they go to "Unorganized")
        set((state) => ({
          folders: state.folders.filter((f) => f.id !== id),
          expandedFolders: new Set(
            Array.from(state.expandedFolders).filter((fId) => fId !== id)
          ),
          selectedFolderId:
            state.selectedFolderId === id ? null : state.selectedFolderId,
        }));
      },

      moveFolder: (folderId, newParentId) => {
        // Prevent circular references
        if (newParentId === folderId) return;

        const wouldCreateCircle = (targetId: string | null): boolean => {
          if (targetId === null) return false;
          if (targetId === folderId) return true;
          const parent = get().folders.find((f) => f.id === targetId);
          return parent ? wouldCreateCircle(parent.parent_id) : false;
        };

        if (wouldCreateCircle(newParentId)) return;

        set((state) => ({
          folders: state.folders.map((folder) =>
            folder.id === folderId
              ? {
                  ...folder,
                  parent_id: newParentId,
                  updated_at: new Date().toISOString(),
                }
              : folder
          ),
        }));
      },

      addCameraToFolder: (cameraId, folderId) => {
        set((state) => ({
          folders: state.folders.map((folder) =>
            folder.id === folderId
              ? {
                  ...folder,
                  camera_ids: folder.camera_ids.includes(cameraId)
                    ? folder.camera_ids
                    : [...folder.camera_ids, cameraId],
                  updated_at: new Date().toISOString(),
                }
              : folder
          ),
        }));
      },

      removeCameraFromFolder: (cameraId, folderId) => {
        set((state) => ({
          folders: state.folders.map((folder) =>
            folder.id === folderId
              ? {
                  ...folder,
                  camera_ids: folder.camera_ids.filter((id) => id !== cameraId),
                  updated_at: new Date().toISOString(),
                }
              : folder
          ),
        }));
      },

      moveCameraBetweenFolders: (cameraId, sourceFolderId, targetFolderId) => {
        set((state) => {
          let updatedFolders = state.folders;

          // Remove from source folder
          if (sourceFolderId) {
            updatedFolders = updatedFolders.map((folder) =>
              folder.id === sourceFolderId
                ? {
                    ...folder,
                    camera_ids: folder.camera_ids.filter((id) => id !== cameraId),
                    updated_at: new Date().toISOString(),
                  }
                : folder
            );
          }

          // Add to target folder
          if (targetFolderId) {
            updatedFolders = updatedFolders.map((folder) =>
              folder.id === targetFolderId
                ? {
                    ...folder,
                    camera_ids: folder.camera_ids.includes(cameraId)
                      ? folder.camera_ids
                      : [...folder.camera_ids, cameraId],
                    updated_at: new Date().toISOString(),
                  }
                : folder
            );
          }

          return { folders: updatedFolders };
        });
      },

      toggleFolderExpanded: (folderId) => {
        set((state) => {
          const newExpanded = new Set(state.expandedFolders);
          if (newExpanded.has(folderId)) {
            newExpanded.delete(folderId);
          } else {
            newExpanded.add(folderId);
          }
          return { expandedFolders: newExpanded };
        });
      },

      expandAllFolders: () => {
        set((state) => ({
          expandedFolders: new Set(state.folders.map((f) => f.id)),
        }));
      },

      collapseAllFolders: () => {
        set({ expandedFolders: new Set<string>() });
      },

      setSelectedFolder: (folderId) => {
        set({ selectedFolderId: folderId });
      },

      buildFolderTree: (cameras) => {
        const { folders, expandedFolders } = get();
        const cameraMap = new Map(cameras.map((c) => [c.id, c]));

        const buildTree = (
          parentId: string | null,
          depth: number = 0
        ): CameraFolderTree[] => {
          return folders
            .filter((folder) => folder.parent_id === parentId)
            .sort((a, b) => a.order - b.order)
            .map((folder) => {
              const folderCameras = folder.camera_ids
                .map((id) => cameraMap.get(id))
                .filter((c): c is Camera => c !== undefined);

              return {
                ...folder,
                children: buildTree(folder.id, depth + 1),
                cameras: folderCameras,
                depth,
                expanded: expandedFolders.has(folder.id),
              };
            });
        };

        return buildTree(null);
      },

      getFolderPath: (folderId) => {
        const path: CameraFolder[] = [];
        let currentId: string | null = folderId;

        while (currentId) {
          const folder = get().folders.find((f) => f.id === currentId);
          if (!folder) break;
          path.unshift(folder);
          currentId = folder.parent_id;
        }

        return path;
      },

      getCamerasInFolder: (folderId, includeSubfolders = false) => {
        const folder = get().folders.find((f) => f.id === folderId);
        if (!folder) return [];

        if (!includeSubfolders) {
          return folder.camera_ids;
        }

        const getAllCameraIds = (fId: string): string[] => {
          const f = get().folders.find((folder) => folder.id === fId);
          if (!f) return [];

          const childFolders = get().folders.filter((cf) => cf.parent_id === fId);
          const childCameraIds = childFolders.flatMap((cf) =>
            getAllCameraIds(cf.id)
          );

          return [...f.camera_ids, ...childCameraIds];
        };

        return getAllCameraIds(folderId);
      },

      initializeDefaultFolders: () => {
        const { folders } = get();
        if (folders.length > 0) return; // Already initialized

        const timestamp = new Date().toISOString();

        const defaultFolders: CameraFolder[] = [
          {
            id: 'folder_dubai_police',
            name: 'Dubai Police',
            name_ar: 'شرطة دبي',
            parent_id: null,
            camera_ids: [],
            order: 0,
            expanded: true,
            created_at: timestamp,
            updated_at: timestamp,
          },
          {
            id: 'folder_sharjah_police',
            name: 'Sharjah Police',
            name_ar: 'شرطة الشارقة',
            parent_id: null,
            camera_ids: [],
            order: 1,
            expanded: true,
            created_at: timestamp,
            updated_at: timestamp,
          },
          {
            id: 'folder_metro',
            name: 'Metro',
            name_ar: 'المترو',
            parent_id: null,
            camera_ids: [],
            order: 2,
            expanded: true,
            created_at: timestamp,
            updated_at: timestamp,
          },
          {
            id: 'folder_taxi',
            name: 'Taxi',
            name_ar: 'التاكسي',
            parent_id: null,
            camera_ids: [],
            order: 3,
            expanded: true,
            created_at: timestamp,
            updated_at: timestamp,
          },
          {
            id: 'folder_parking',
            name: 'Parking',
            name_ar: 'مواقف السيارات',
            parent_id: null,
            camera_ids: [],
            order: 4,
            expanded: true,
            created_at: timestamp,
            updated_at: timestamp,
          },
          {
            id: 'folder_unorganized',
            name: 'Unorganized',
            name_ar: 'غير منظم',
            parent_id: null,
            camera_ids: [],
            order: 999,
            expanded: true,
            created_at: timestamp,
            updated_at: timestamp,
          },
        ];

        set({
          folders: defaultFolders,
          expandedFolders: new Set(defaultFolders.map((f) => f.id)),
        });
      },
    }),
    {
      name: 'folder-storage',
      partialize: (state) => ({
        folders: state.folders,
        expandedFolders: Array.from(state.expandedFolders),
      }),
      onRehydrateStorage: () => (state) => {
        if (state) {
          state.expandedFolders = new Set(state.expandedFolders);
        }
      },
    }
  )
);
