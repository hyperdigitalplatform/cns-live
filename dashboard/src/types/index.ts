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
  milestone_device_id?: string; // Milestone XProtect device ID
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
  | 'PARKING'
  | 'OTHER';

export type CameraStatus = 'ONLINE' | 'OFFLINE' | 'MAINTENANCE' | 'ERROR';

// Discovery types
export interface DiscoveredCamera {
  milestoneId: string;
  name: string;
  displayName: string;
  enabled: boolean;
  status: string;
  ptzEnabled: boolean;
  recordingEnabled: boolean;
  shortName: string;
  device: DeviceInfo;
  onvifEndpoint: string;
  onvifUsername: string;
  streams: StreamProfile[];
  ptzCapabilities: PtzCapabilities;
}

export interface DeviceInfo {
  ip: string;
  port: number;
  manufacturer: string;
  model: string;
  firmwareVersion: string;
  serialNumber: string;
  hardwareId: string;
}

export interface StreamProfile {
  profileToken: string;
  name: string;
  encoding: string;
  resolution: string;
  width: number;
  height: number;
  frameRate: number;
  bitrate: number;
  rtspUrl: string;
}

export interface PtzCapabilities {
  pan: boolean;
  tilt: boolean;
  zoom: boolean;
}

export interface CameraDiscoveryResponse {
  cameras: DiscoveredCamera[];
  total: number;
}

export interface ImportCameraRequest {
  milestoneId: string;
  name: string;
  source: CameraSource;
  status: CameraStatus;
  ptzEnabled: boolean;
  device: DeviceInfo;
  onvifEndpoint: string;
  onvifUsername: string;
  streams: StreamProfile[];
  ptzCapabilities: PtzCapabilities;
}

export interface ImportCamerasRequest {
  cameras: ImportCameraRequest[];
}

export interface ImportCamerasResponse {
  imported: number;
  failed: number;
  errors?: string[];
}

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

// Milestone Recording types
export interface MilestoneRecordingRequest {
  cameraId: string;
  durationMinutes?: number; // Default 15 if not specified
}

export interface MilestoneRecordingStatusResponse {
  cameraId: string;
  isRecording: boolean;
}

export interface MilestoneSequenceType {
  id: string;
  name: string;
}

export interface MilestoneSequenceTypesResponse {
  cameraId: string;
  types: MilestoneSequenceType[];
}

export interface MilestoneSequenceEntry {
  timeBegin: string;
  timeTrigged: string;
  timeEnd: string;
}

export interface MilestoneSequencesRequest {
  cameraId: string;
  startTime: string; // ISO 8601 format
  endTime: string;   // ISO 8601 format
  sequenceTypes?: string[];
}

export interface MilestoneSequencesResponse {
  cameraId: string;
  sequences: MilestoneSequenceEntry[];
}

export interface MilestoneTimelineRequest {
  cameraId: string;
  startTime: string; // ISO 8601 format
  endTime: string;   // ISO 8601 format
  sequenceType?: string;
}

export interface MilestoneTimelineResponse {
  cameraId: string;
  timeline: {
    count: number;
    data: string; // Base64 encoded bitmap
  };
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

// Layout Preference types
export type LayoutType = 'standard' | 'hotspot';
export type LayoutScope = 'global' | 'local';

export interface LayoutCameraAssignment {
  id?: string;
  layout_id?: string;
  camera_id: string;
  position_index: number;
  cell_size?: string;
  created_at?: string;
}

export interface LayoutPreference {
  id: string;
  name: string;
  description?: string;
  layout_type: LayoutType;
  grid_layout: string; // "2x2", "3x3", "9-way-1-hotspot", etc.
  scope: LayoutScope;
  created_by: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
  cameras?: LayoutCameraAssignment[];
}

export interface LayoutPreferenceSummary {
  id: string;
  name: string;
  description?: string;
  layout_type: LayoutType;
  grid_layout: string; // "2x2", "3x3", "9-way-1-hotspot", etc.
  scope: LayoutScope;
  created_by: string;
  camera_count: number;
  created_at: string;
  updated_at: string;
}

export interface CreateLayoutRequest {
  name: string;
  description?: string;
  layout_type: LayoutType;
  grid_layout: string; // "2x2", "3x3", "9-way-1-hotspot", etc.
  scope: LayoutScope;
  created_by: string;
  cameras: LayoutCameraAssignment[];
}

export interface UpdateLayoutRequest {
  name: string;
  description?: string;
  cameras: LayoutCameraAssignment[];
}

export interface LayoutListResponse {
  layouts: LayoutPreferenceSummary[];
  total: number;
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
