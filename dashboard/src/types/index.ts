// Camera types
export interface Camera {
  id: string;
  name: string;
  name_ar: string;
  source: CameraSource;
  rtsp_url: string;
  status: CameraStatus;
  ptz_enabled: boolean;
  recording_server?: string;
  metadata?: Record<string, unknown>;
  location?: CameraLocation;
  created_at: string;
  updated_at: string;
}

export type CameraSource =
  | 'DUBAI_POLICE'
  | 'SHARJAH_POLICE'
  | 'ABU_DHABI_POLICE'
  | 'METRO'
  | 'TAXI'
  | 'PARKING';

export type CameraStatus = 'ONLINE' | 'OFFLINE' | 'MAINTENANCE' | 'ERROR';

export interface CameraLocation {
  latitude: number;
  longitude: number;
  address: string;
  address_ar?: string;
}

// Stream types
export interface StreamReservation {
  reservation_id: string;
  camera_id: string;
  camera_name: string;
  room_name: string;
  token: string;
  livekit_url: string;
  expires_at: string;
  quality: StreamQuality;
}

export type StreamQuality = 'high' | 'medium' | 'low';

export interface StreamStats {
  active_streams: number;
  total_viewers: number;
  source_stats: Record<string, SourceStats>;
  camera_stats: CameraStreamStats[];
  timestamp: string;
}

export interface SourceStats {
  source: string;
  current: number;
  limit: number;
  usage_percent: number;
  active_cameras: number;
}

export interface CameraStreamStats {
  camera_id: string;
  camera_name: string;
  viewer_count: number;
  source: string;
  active_since: string;
}

// Playback types
export interface PlaybackRequest {
  camera_id: string;
  start_time: string;
  end_time: string;
  format: 'hls' | 'rtsp';
  user_id?: string;
}

export interface PlaybackResponse {
  session_id: string;
  camera_id: string;
  start_time: string;
  end_time: string;
  format: string;
  url: string;
  expires_at: string;
  segment_ids: string[];
}

export interface ExportRequest {
  camera_id: string;
  start_time: string;
  end_time: string;
  format: 'mp4' | 'avi' | 'mkv';
  user_id: string;
  title?: string;
  description?: string;
}

export interface ExportResponse {
  export_id: string;
  camera_id: string;
  start_time: string;
  end_time: string;
  format: string;
  status: 'pending' | 'processing' | 'ready' | 'failed';
  download_url?: string;
  created_at: string;
}

// Grid layout types
export interface GridLayout {
  id: string;
  name: string;
  columns: number;
  rows: number;
  cells: GridCell[];
}

export interface GridCell {
  index: number;
  camera_id?: string;
  span_columns?: number;
  span_rows?: number;
}

// Alert types
export interface Alert {
  id: string;
  type: AlertType;
  severity: AlertSeverity;
  camera_id?: string;
  camera_name?: string;
  message: string;
  message_ar?: string;
  timestamp: string;
  acknowledged: boolean;
}

export type AlertType =
  | 'CAMERA_OFFLINE'
  | 'CAMERA_ERROR'
  | 'STREAM_LIMIT'
  | 'RECORDING_ERROR'
  | 'SYSTEM_ERROR';

export type AlertSeverity = 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL';

// User types
export interface User {
  id: string;
  username: string;
  email: string;
  agency: string;
  role: UserRole;
}

export type UserRole = 'ADMIN' | 'OPERATOR' | 'VIEWER';

// API Error types
export interface APIError {
  code: string;
  message_en: string;
  message_ar?: string;
  details?: Record<string, unknown>;
}

// Camera Folder types (Tree structure for organizing cameras)
export interface CameraFolder {
  id: string;
  name: string;
  name_ar?: string;
  parent_id: string | null; // null for root folders
  camera_ids: string[];     // Cameras directly in this folder
  order: number;            // Display order
  expanded?: boolean;       // UI state for tree collapse/expand
  created_at: string;
  updated_at: string;
  created_by?: string;
  metadata?: Record<string, unknown>;
}

export interface CameraFolderTree extends CameraFolder {
  children: CameraFolderTree[];
  cameras: Camera[];
  depth: number;
}

// Drag and drop types
export interface DragItem {
  type: 'camera' | 'folder' | 'grid-cell';
  id?: string;
  sourceFolder?: string | null;
  data: Camera | CameraFolder | number;
}

export interface DropTarget {
  type: 'folder' | 'grid-cell' | 'trash';
  id: string;
  index?: number; // For grid cell drops
}
