import type {
  Camera,
  StreamReservation,
  StreamStats,
  PlaybackRequest,
  PlaybackResponse,
  ExportRequest,
  ExportResponse,
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
}

export const api = new APIClient(API_BASE_URL);
