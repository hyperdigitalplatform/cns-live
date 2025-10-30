export interface RecordingSequence {
  sequenceId: string;
  startTime: string;
  endTime: string;
  durationSeconds: number;
  available: boolean;
  sizeBytes?: number;
}

export interface RecordingGap {
  startTime: string;
  endTime: string;
  durationSeconds: number;
  reason?: string;
}

export interface TimelineData {
  cameraId: string;
  queryRange: {
    start: string;
    end: string;
  };
  sequences: RecordingSequence[];
  gaps: RecordingGap[];
  totalRecordingSeconds: number;
  totalGapSeconds: number;
  coverage: number;
}

export interface PlaybackState {
  mode: 'live' | 'playback';
  isPlaying: boolean;
  currentTime: Date;
  startTime: Date;
  endTime: Date;
  speed: number;
  zoomLevel: number;
  timelineData: TimelineData | null;
}

export type PlaybackMode = 'live' | 'playback';
