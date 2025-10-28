import React, { useEffect, useState } from 'react';
import { Circle, Square, Clock, Play, Film } from 'lucide-react';
import { cn } from '@/utils/cn';
import { Button } from './ui/Dialog';
import { api } from '@/services/api';

interface RecordingStatus {
  isRecording: boolean;
  currentRecording?: {
    recordingId: string;
    startTime: string;
    estimatedEndTime: string;
    durationSeconds: number;
  };
  lastRecording?: {
    recordingId: string;
    startTime: string;
    endTime?: string;
  };
}

interface RecordingControlProps {
  cameraId: string;
  className?: string;
  onViewRecordings?: () => void;
}

const DURATION_OPTIONS = [
  { label: '1 minute', value: 1 },
  { label: '5 minutes', value: 5 },
  { label: '15 minutes', value: 15 },
  { label: '30 minutes', value: 30 },
  { label: '1 hour', value: 60 },
  { label: '2 hours', value: 120 },
];

export function RecordingControl({
  cameraId,
  className,
  onViewRecordings,
}: RecordingControlProps) {
  const [isRecording, setIsRecording] = useState(false);
  const [duration, setDuration] = useState(15); // 15 min default (in minutes)
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Fetch recording status
  const fetchStatus = async () => {
    try {
      const data = await api.getMilestoneRecordingStatus(cameraId);
      setIsRecording(data.isRecording);
    } catch (err) {
      console.error('Failed to fetch recording status:', err);
    }
  };

  // Start recording
  const handleStartRecording = async () => {
    setLoading(true);
    setError(null);

    try {
      await api.startMilestoneRecording({
        cameraId,
        durationMinutes: duration,
      });

      await fetchStatus();
    } catch (err: any) {
      console.error('Failed to start recording:', err);
      setError(err.message || 'Failed to start recording');
    } finally {
      setLoading(false);
    }
  };

  // Stop recording
  const handleStopRecording = async () => {
    setLoading(true);
    setError(null);

    try {
      await api.stopMilestoneRecording(cameraId);
      await fetchStatus();
    } catch (err: any) {
      console.error('Failed to stop recording:', err);
      setError(err.message || 'Failed to stop recording');
    } finally {
      setLoading(false);
    }
  };

  // Poll status when component mounts and when recording
  useEffect(() => {
    fetchStatus();

    const interval = setInterval(() => {
      fetchStatus();
    }, 5000); // Poll every 5 seconds

    return () => clearInterval(interval);
  }, [cameraId]);

  return (
    <div className={cn("space-y-3", className)}>
      {/* Header */}
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-semibold text-gray-900">Recording Control</h3>
        {isRecording && (
          <span className="flex items-center gap-1 text-xs font-medium text-red-600">
            <Circle className="w-2 h-2 fill-current animate-pulse" />
            Recording
          </span>
        )}
      </div>

      {error && (
        <div className="p-2 bg-red-50 border border-red-200 rounded text-xs text-red-700">
          {error}
        </div>
      )}

      {isRecording ? (
        /* Recording Active UI */
        <div className="space-y-3">
          {/* Status Display */}
          <div className="p-3 bg-red-50 border border-red-200 rounded-lg text-center">
            <div className="flex items-center justify-center gap-2 mb-2">
              <Circle className="w-3 h-3 fill-red-600 text-red-600 animate-pulse" />
              <span className="text-sm font-semibold text-red-600">
                Recording in Progress
              </span>
            </div>
            <p className="text-xs text-gray-600">
              The camera is currently recording to Milestone
            </p>
          </div>

          {/* Stop Button */}
          <Button
            onClick={handleStopRecording}
            disabled={loading}
            className="w-full bg-red-600 hover:bg-red-700 text-white flex items-center justify-center gap-2"
          >
            <Square className="w-4 h-4" />
            {loading ? 'Stopping...' : 'Stop Recording'}
          </Button>
        </div>
      ) : (
        /* Recording Idle UI */
        <div className="space-y-3">
          {/* Duration Selector */}
          <div>
            <label className="block text-xs font-medium text-gray-700 mb-1.5">
              Duration
            </label>
            <select
              value={duration}
              onChange={(e) => setDuration(Number(e.target.value))}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              {DURATION_OPTIONS.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
          </div>

          {/* Start Button */}
          <Button
            onClick={handleStartRecording}
            disabled={loading}
            className="w-full bg-red-600 hover:bg-red-700 text-white flex items-center justify-center gap-2"
          >
            <Circle className="w-4 h-4" />
            {loading ? 'Starting...' : 'Start Recording'}
          </Button>
        </div>
      )}

      {/* View Recordings Button */}
      {onViewRecordings && (
        <Button
          onClick={onViewRecordings}
          variant="secondary"
          className="w-full flex items-center justify-center gap-2"
        >
          <Film className="w-4 h-4" />
          View Recordings
        </Button>
      )}
    </div>
  );
}
