import React, { useState } from 'react';
import { PlaybackPlayer } from '@/components/PlaybackPlayer';
import { useCameraStore } from '@/stores/cameraStore';
import type { Camera } from '@/types';
import { Calendar, Clock } from 'lucide-react';
import { format, subHours } from 'date-fns';

export function PlaybackView() {
  const cameras = useCameraStore((state) => state.cameras);
  const [selectedCamera, setSelectedCamera] = useState<Camera | null>(null);
  const [startTime, setStartTime] = useState(
    format(subHours(new Date(), 1), "yyyy-MM-dd'T'HH:mm")
  );
  const [endTime, setEndTime] = useState(
    format(new Date(), "yyyy-MM-dd'T'HH:mm")
  );
  const [showPlayer, setShowPlayer] = useState(false);

  const handlePlayback = () => {
    if (selectedCamera) {
      setShowPlayer(true);
    }
  };

  if (showPlayer && selectedCamera) {
    return (
      <PlaybackPlayer
        camera={selectedCamera}
        startTime={new Date(startTime)}
        endTime={new Date(endTime)}
        onClose={() => setShowPlayer(false)}
      />
    );
  }

  return (
    <div className="flex flex-col h-full bg-gray-100 dark:bg-dark-base">
      <div className="bg-white dark:bg-dark-secondary border-b border-gray-200 dark:border-dark-border px-6 py-4">
        <h1 className="text-2xl font-bold text-gray-900 dark:text-text-primary">Playback</h1>
        <p className="text-sm text-gray-600 dark:text-text-secondary mt-1">
          View recorded footage from cameras
        </p>
      </div>

      <div className="flex-1 flex items-center justify-center p-6">
        <div className="bg-white dark:bg-dark-secondary rounded-lg shadow-lg p-8 w-full max-w-2xl">
          <h2 className="text-xl font-semibold text-gray-900 dark:text-text-primary mb-6">
            Select Playback Parameters
          </h2>

          <div className="space-y-6">
            {/* Camera Selection */}
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-text-secondary mb-2">
                Camera
              </label>
              <select
                value={selectedCamera?.id || ''}
                onChange={(e) => {
                  const camera = cameras.find((c) => c.id === e.target.value);
                  setSelectedCamera(camera || null);
                }}
                className="w-full px-4 py-2 border border-gray-300 dark:border-dark-border rounded-lg bg-white dark:bg-dark-surface text-gray-900 dark:text-text-primary focus:outline-none focus:ring-2 focus:ring-primary-500"
              >
                <option value="">Select a camera...</option>
                {cameras.map((camera) => (
                  <option key={camera.id} value={camera.id}>
                    {camera.name} ({camera.source})
                  </option>
                ))}
              </select>
            </div>

            {/* Time Range */}
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-text-secondary mb-2">
                  <Calendar className="inline w-4 h-4 mr-1" />
                  Start Time
                </label>
                <input
                  type="datetime-local"
                  value={startTime}
                  onChange={(e) => setStartTime(e.target.value)}
                  className="w-full px-4 py-2 border border-gray-300 dark:border-dark-border rounded-lg bg-white dark:bg-dark-surface text-gray-900 dark:text-text-primary focus:outline-none focus:ring-2 focus:ring-primary-500"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-text-secondary mb-2">
                  <Clock className="inline w-4 h-4 mr-1" />
                  End Time
                </label>
                <input
                  type="datetime-local"
                  value={endTime}
                  onChange={(e) => setEndTime(e.target.value)}
                  className="w-full px-4 py-2 border border-gray-300 dark:border-dark-border rounded-lg bg-white dark:bg-dark-surface text-gray-900 dark:text-text-primary focus:outline-none focus:ring-2 focus:ring-primary-500"
                />
              </div>
            </div>

            {/* Quick Time Ranges */}
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-text-secondary mb-2">
                Quick Select
              </label>
              <div className="flex gap-2">
                {[
                  { label: 'Last Hour', hours: 1 },
                  { label: 'Last 6 Hours', hours: 6 },
                  { label: 'Last 24 Hours', hours: 24 },
                ].map(({ label, hours }) => (
                  <button
                    key={label}
                    onClick={() => {
                      const now = new Date();
                      setEndTime(format(now, "yyyy-MM-dd'T'HH:mm"));
                      setStartTime(
                        format(subHours(now, hours), "yyyy-MM-dd'T'HH:mm")
                      );
                    }}
                    className="px-4 py-2 bg-gray-100 dark:bg-dark-surface hover:bg-gray-200 dark:hover:bg-dark-elevated text-gray-700 dark:text-text-secondary rounded-lg text-sm font-medium transition-colors"
                  >
                    {label}
                  </button>
                ))}
              </div>
            </div>

            {/* Submit */}
            <button
              onClick={handlePlayback}
              disabled={!selectedCamera}
              className="w-full px-6 py-3 bg-primary-600 hover:bg-primary-700 disabled:bg-gray-300 dark:disabled:bg-dark-surface text-white dark:disabled:text-text-muted font-medium rounded-lg transition-colors"
            >
              Start Playback
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
