import type {
  Camera,
  StreamReservation,
  StreamStats,
  PlaybackRequest,
  PlaybackResponse,
  ExportRequest,
  ExportResponse,
  LayoutPreference,
  LayoutPreferenceSummary,
  LayoutListResponse,
  CreateLayoutRequest,
  UpdateLayoutRequest,
  LayoutType,
  LayoutScope,
  MilestoneRecordingRequest,
  MilestoneRecordingStatusResponse,
  MilestoneSequenceTypesResponse,
  MilestoneSequencesRequest,
  MilestoneSequencesResponse,
  MilestoneTimelineRequest,
  MilestoneTimelineResponse,
  CameraDiscoveryResponse,
  ImportCamerasRequest,
  ImportCamerasResponse,
} from '@/types';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8000';

class APIClient {
  private baseURL: string;

  constructor(baseURL: string) {
    this.baseURL = baseURL;
  }

  private async request<T>(
    endpoint: string,
    options?: RequestInit
  ): Promise<T> {
    const response = await fetch(`${this.baseURL}${endpoint}`, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...options?.headers,
      },
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({
        message: response.statusText,
      }));
      throw new Error(error.message || 'API request failed');
    }

    // Handle 204 No Content responses (e.g., DELETE)
    if (response.status === 204) {
      return undefined as T;
    }

    return response.json();
  }

  // Camera endpoints
  async getCameras(params?: {
    source?: string;
    status?: string;
    limit?: number;
    offset?: number;
  }): Promise<{ cameras: Camera[]; count: number }> {
    const queryParams = new URLSearchParams();
    if (params?.source) queryParams.set('source', params.source);
    if (params?.status) queryParams.set('status', params.status);
    if (params?.limit) queryParams.set('limit', params.limit.toString());
    if (params?.offset) queryParams.set('offset', params.offset.toString());

    const query = queryParams.toString();
    return this.request(`/api/v1/cameras${query ? `?${query}` : ''}`);
  }

  async getCamera(cameraId: string): Promise<Camera> {
    return this.request(`/api/v1/cameras/${cameraId}`);
  }

  async deleteCamera(cameraId: string): Promise<void> {
    return this.request(`/api/v1/cameras/${cameraId}`, {
      method: 'DELETE',
    });
  }

  async controlPTZ(
    cameraId: string,
    command: string,
    params?: { speed?: number; preset_id?: number }
  ): Promise<{ status: string; message: string }> {
    return this.request(`/api/v1/cameras/${cameraId}/ptz`, {
      method: 'POST',
      body: JSON.stringify({
        command,
        ...params,
        user_id: 'dashboard-user', // TODO: Get from auth
      }),
    });
  }

  // Stream endpoints
  async reserveStream(
    cameraId: string,
    quality: 'high' | 'medium' | 'low' = 'medium'
  ): Promise<StreamReservation> {
    return this.request('/api/v1/stream/reserve', {
      method: 'POST',
      body: JSON.stringify({
        camera_id: cameraId,
        user_id: 'dashboard-user', // TODO: Get from auth
        quality,
      }),
    });
  }

  async releaseStream(reservationId: string): Promise<void> {
    await this.request(`/api/v1/stream/release/${reservationId}`, {
      method: 'DELETE',
    });
  }

  async sendHeartbeat(reservationId: string): Promise<void> {
    await this.request(`/api/v1/stream/heartbeat/${reservationId}`, {
      method: 'POST',
    });
  }

  async getStreamStats(): Promise<StreamStats> {
    return this.request('/api/v1/stream/stats');
  }

  // Playback endpoints
  async requestPlayback(
    request: PlaybackRequest
  ): Promise<PlaybackResponse> {
    return this.request('/api/v1/playback/request', {
      method: 'POST',
      body: JSON.stringify(request),
    });
  }

  async createExport(request: ExportRequest): Promise<ExportResponse> {
    return this.request('/api/v1/playback/export', {
      method: 'POST',
      body: JSON.stringify(request),
    });
  }

  // Layout preference endpoints
  async createLayout(request: CreateLayoutRequest): Promise<LayoutPreference> {
    return this.request('/api/v1/layouts', {
      method: 'POST',
      body: JSON.stringify(request),
    });
  }

  async getLayouts(params?: {
    layout_type?: LayoutType;
    scope?: LayoutScope;
    created_by?: string;
  }): Promise<LayoutListResponse> {
    const queryParams = new URLSearchParams();
    if (params?.layout_type) queryParams.set('layout_type', params.layout_type);
    if (params?.scope) queryParams.set('scope', params.scope);
    if (params?.created_by) queryParams.set('created_by', params.created_by);

    const query = queryParams.toString();
    return this.request(`/api/v1/layouts${query ? `?${query}` : ''}`);
  }

  async getLayout(layoutId: string): Promise<LayoutPreference> {
    return this.request(`/api/v1/layouts/${layoutId}`);
  }

  async updateLayout(
    layoutId: string,
    request: UpdateLayoutRequest
  ): Promise<LayoutPreference> {
    return this.request(`/api/v1/layouts/${layoutId}`, {
      method: 'PUT',
      body: JSON.stringify(request),
    });
  }

  async deleteLayout(layoutId: string): Promise<void> {
    await this.request(`/api/v1/layouts/${layoutId}`, {
      method: 'DELETE',
    });
  }

  // Milestone Recording Control endpoints
  async startMilestoneRecording(
    request: MilestoneRecordingRequest
  ): Promise<{ status: string; message: string }> {
    return this.request('/api/v1/milestone/recordings/start', {
      method: 'POST',
      body: JSON.stringify(request),
    });
  }

  async stopMilestoneRecording(
    cameraId: string
  ): Promise<{ status: string; message: string }> {
    return this.request('/api/v1/milestone/recordings/stop', {
      method: 'POST',
      body: JSON.stringify({ cameraId }),
    });
  }

  async getMilestoneRecordingStatus(
    cameraId: string
  ): Promise<MilestoneRecordingStatusResponse> {
    return this.request(`/api/v1/milestone/recordings/status/${cameraId}`);
  }

  async getMilestoneSequenceTypes(
    cameraId: string
  ): Promise<MilestoneSequenceTypesResponse> {
    return this.request(`/api/v1/milestone/sequences/types/${cameraId}`);
  }

  async getMilestoneSequences(
    request: MilestoneSequencesRequest
  ): Promise<MilestoneSequencesResponse> {
    return this.request('/api/v1/milestone/sequences', {
      method: 'POST',
      body: JSON.stringify(request),
    });
  }

  async getMilestoneTimeline(
    request: MilestoneTimelineRequest
  ): Promise<MilestoneTimelineResponse> {
    return this.request('/api/v1/milestone/timeline', {
      method: 'POST',
      body: JSON.stringify(request),
    });
  }

  // Camera Discovery endpoints
  async discoverCameras(): Promise<CameraDiscoveryResponse> {
    return this.request('/api/v1/milestone/cameras/discover');
  }

  async importCameras(
    request: ImportCamerasRequest
  ): Promise<ImportCamerasResponse> {
    return this.request('/api/v1/cameras/import', {
      method: 'POST',
      body: JSON.stringify(request),
    });
  }
}

export const api = new APIClient(API_BASE_URL);
