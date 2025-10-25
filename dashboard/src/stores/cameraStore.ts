import { create } from 'zustand';
import type { Camera, CameraSource, CameraStatus } from '@/types';
import { api } from '@/services/api';

interface CameraState {
  cameras: Camera[];
  selectedCamera: Camera | null;
  loading: boolean;
  error: string | null;

  // Actions
  fetchCameras: (params?: {
    source?: CameraSource;
    status?: CameraStatus;
  }) => Promise<void>;
  selectCamera: (camera: Camera | null) => void;
  refreshCamera: (cameraId: string) => Promise<void>;
}

export const useCameraStore = create<CameraState>((set, get) => ({
  cameras: [],
  selectedCamera: null,
  loading: false,
  error: null,

  fetchCameras: async (params) => {
    set({ loading: true, error: null });
    try {
      const response = await api.getCameras(params);
      set({ cameras: response.cameras, loading: false });
    } catch (error) {
      set({
        error: error instanceof Error ? error.message : 'Failed to fetch cameras',
        loading: false,
      });
    }
  },

  selectCamera: (camera) => {
    set({ selectedCamera: camera });
  },

  refreshCamera: async (cameraId) => {
    try {
      const camera = await api.getCamera(cameraId);
      const cameras = get().cameras.map((c) =>
        c.id === cameraId ? camera : c
      );
      set({ cameras });

      if (get().selectedCamera?.id === cameraId) {
        set({ selectedCamera: camera });
      }
    } catch (error) {
      console.error('Failed to refresh camera:', error);
    }
  },
}));
