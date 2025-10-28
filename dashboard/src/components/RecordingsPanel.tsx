import React, { useEffect, useState } from 'react';
import { X, Calendar, Clock, Video, Loader2, AlertCircle } from 'lucide-react';
import { api } from '@/services/api';
import type { Camera, MilestoneSequenceEntry } from '@/types';
import { RecordingPlayer } from './RecordingPlayer';

interface RecordingsPanelProps {
  camera: Camera;
  onClose: () => void;
}

export function RecordingsPanel({ camera, onClose }: RecordingsPanelProps) {
  const [recordings, setRecordings] = useState<MilestoneSequenceEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedRecording, setSelectedRecording] = useState<MilestoneSequenceEntry | null>(null);
  const [timeRange, setTimeRange] = useState({
    startTime: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString(), // Last 7 days
    endTime: new Date().toISOString(),
  });

  useEffect(() => {
    fetchRecordings();
  }, [camera.id, timeRange]);

  const fetchRecordings = async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await api.getMilestoneSequences({
        cameraId: camera.id,
        startTime: timeRange.startTime,
        endTime: timeRange.endTime,
      });
      setRecordings(response.sequences || []);
    } catch (err) {
      console.error('Failed to fetch recordings:', err);
      setError(err instanceof Error ? err.message : 'Failed to load recordings');
    } finally {
      setLoading(false);
    }
  };

  const formatDateTime = (isoString: string): string => {
    const date = new Date(isoString);
    return date.toLocaleString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
      hour12: false,
    });
  };

  const formatDuration = (start: string, end: string): string => {
    const startDate = new Date(start);
    const endDate = new Date(end);
    const durationMs = endDate.getTime() - startDate.getTime();
    const durationSeconds = Math.floor(durationMs / 1000);
    const hours = Math.floor(durationSeconds / 3600);
    const minutes = Math.floor((durationSeconds % 3600) / 60);
    const seconds = durationSeconds % 60;

    if (hours > 0) {
      return `${hours}h ${minutes}m ${seconds}s`;
    } else if (minutes > 0) {
      return `${minutes}m ${seconds}s`;
    } else {
      return `${seconds}s`;
    }
  };

  return (
    <div className="fixed inset-0 bg-black/80 flex items-center justify-center z-50 p-4">
      <div className="bg-gray-900 rounded-lg shadow-2xl w-full max-w-6xl h-[90vh] flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b border-gray-700">
          <div className="flex items-center gap-3">
            <Video className="w-5 h-5 text-blue-500" />
            <div>
              <h2 className="text-lg font-semibold text-white">{camera.name} - Recordings</h2>
              <p className="text-sm text-gray-400">
                {formatDateTime(timeRange.startTime)} - {formatDateTime(timeRange.endTime)}
              </p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="p-2 hover:bg-gray-800 rounded-lg transition-colors text-gray-400 hover:text-white"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Content */}
        <div className="flex-1 flex overflow-hidden">
          {/* Recordings List */}
          <div className="w-80 border-r border-gray-700 flex flex-col">
            <div className="p-4 border-b border-gray-700">
              <h3 className="text-sm font-medium text-gray-300 mb-2">Time Range</h3>
              <div className="flex flex-col gap-2">
                <input
                  type="datetime-local"
                  value={timeRange.startTime.slice(0, 16)}
                  onChange={(e) => setTimeRange(prev => ({ ...prev, startTime: new Date(e.target.value).toISOString() }))}
                  className="bg-gray-800 text-white text-sm px-3 py-2 rounded border border-gray-600 focus:border-blue-500 focus:outline-none"
                />
                <input
                  type="datetime-local"
                  value={timeRange.endTime.slice(0, 16)}
                  onChange={(e) => setTimeRange(prev => ({ ...prev, endTime: new Date(e.target.value).toISOString() }))}
                  className="bg-gray-800 text-white text-sm px-3 py-2 rounded border border-gray-600 focus:border-blue-500 focus:outline-none"
                />
                <button
                  onClick={fetchRecordings}
                  className="bg-blue-600 hover:bg-blue-700 text-white text-sm px-3 py-2 rounded transition-colors"
                >
                  Refresh
                </button>
              </div>
            </div>

            {/* List */}
            <div className="flex-1 overflow-y-auto">
              {loading ? (
                <div className="flex items-center justify-center h-full">
                  <Loader2 className="w-6 h-6 text-gray-400 animate-spin" />
                </div>
              ) : error ? (
                <div className="flex flex-col items-center justify-center h-full p-4 text-center">
                  <AlertCircle className="w-8 h-8 text-red-500 mb-2" />
                  <p className="text-sm text-gray-400">{error}</p>
                </div>
              ) : recordings.length === 0 ? (
                <div className="flex flex-col items-center justify-center h-full p-4 text-center">
                  <Video className="w-8 h-8 text-gray-600 mb-2" />
                  <p className="text-sm text-gray-400">No recordings found</p>
                  <p className="text-xs text-gray-500 mt-1">Try adjusting the time range</p>
                </div>
              ) : (
                <div className="divide-y divide-gray-700">
                  {recordings.map((recording, index) => (
                    <button
                      key={index}
                      onClick={() => setSelectedRecording(recording)}
                      className={`w-full text-left p-3 hover:bg-gray-800 transition-colors ${
                        selectedRecording === recording ? 'bg-gray-800 border-l-2 border-blue-500' : ''
                      }`}
                    >
                      <div className="flex items-start gap-2">
                        <Calendar className="w-4 h-4 text-gray-400 mt-0.5 flex-shrink-0" />
                        <div className="flex-1 min-w-0">
                          <p className="text-sm text-white font-medium truncate">
                            {formatDateTime(recording.timeTrigged)}
                          </p>
                          <div className="flex items-center gap-1 mt-1">
                            <Clock className="w-3 h-3 text-gray-500" />
                            <p className="text-xs text-gray-400">
                              {formatDuration(recording.timeBegin, recording.timeEnd)}
                            </p>
                          </div>
                          <p className="text-xs text-gray-500 mt-1">
                            {formatDateTime(recording.timeBegin)}
                          </p>
                        </div>
                      </div>
                    </button>
                  ))}
                </div>
              )}
            </div>
          </div>

          {/* Player */}
          <div className="flex-1 flex items-center justify-center bg-black">
            {selectedRecording ? (
              <RecordingPlayer
                cameraId={camera.id}
                startTime={new Date(selectedRecording.timeBegin)}
                endTime={new Date(selectedRecording.timeEnd)}
                initialPlaybackTime={new Date(selectedRecording.timeTrigged)}
                className="w-full h-full"
              />
            ) : (
              <div className="text-center text-gray-500">
                <Video className="w-16 h-16 mx-auto mb-4 opacity-50" />
                <p className="text-lg">Select a recording to play</p>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
