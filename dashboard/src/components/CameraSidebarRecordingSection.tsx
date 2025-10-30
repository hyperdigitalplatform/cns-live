import React, { useState } from 'react';
import { RecordingControl } from './RecordingControl';
import { RecordingTimeline } from './RecordingTimeline';
import { RecordingPlayer } from './RecordingPlayer';
import { Film, Clock } from 'lucide-react';
import { Button, Dialog, DialogContent, DialogHeader, DialogTitle, DialogBody } from './ui/Dialog';
import type { Camera } from '@/types';
import { api } from '@/services/api';

interface CameraSidebarRecordingSectionProps {
  selectedCamera: Camera | null;
}

export function CameraSidebarRecordingSection({
  selectedCamera,
}: CameraSidebarRecordingSectionProps) {
  const [showRecordings, setShowRecordings] = useState(false);
  const [queryStartTime, setQueryStartTime] = useState(
    new Date(Date.now() - 24 * 60 * 60 * 1000) // 24 hours ago
  );
  const [queryEndTime, setQueryEndTime] = useState(new Date());
  const [timelineData, setTimelineData] = useState<any>(null);
  const [playbackTime, setPlaybackTime] = useState<Date | undefined>(undefined);

  // Query recordings when dialog opens
  const handleOpenRecordings = async () => {
    setShowRecordings(true);
    if (!selectedCamera) return;

    try {
      const data = await api.getMilestoneSequences({
        cameraId: selectedCamera.id,
        startTime: queryStartTime.toISOString(),
        endTime: queryEndTime.toISOString(),
      });

      // Transform Milestone sequences to timeline format
      const transformedData = {
        cameraId: selectedCamera.id,
        queryRange: {
          start: queryStartTime.toISOString(),
          end: queryEndTime.toISOString(),
        },
        sequences: data.sequences.map((seq) => ({
          sequenceId: `${seq.timeBegin}-${seq.timeEnd}`,
          startTime: seq.timeBegin,
          endTime: seq.timeEnd,
          durationSeconds: (new Date(seq.timeEnd).getTime() - new Date(seq.timeBegin).getTime()) / 1000,
          available: true,
          sizeBytes: 0,
        })),
        gaps: [],
        totalRecordingSeconds: 0,
        totalGapSeconds: 0,
        coverage: 0,
      };

      setTimelineData(transformedData);
    } catch (error) {
      console.error('Failed to query recordings:', error);
    }
  };

  const handleSeek = (timestamp: Date) => {
    setPlaybackTime(timestamp);
  };


  if (!selectedCamera) {
    return (
      <div className="p-4 text-center text-gray-500">
        <p className="text-sm">Select a camera to access recording features</p>
      </div>
    );
  }

  return (
    <div className="border-t border-gray-200">
      {/* Recording Control */}
      {selectedCamera.milestone_device_id ? (
        <div className="p-3 border-b border-gray-200">
          <RecordingControl
            cameraId={selectedCamera.id}
            onViewRecordings={handleOpenRecordings}
          />
        </div>
      ) : (
        <div className="p-4 text-center text-gray-500 bg-yellow-50 border-b border-yellow-200">
          <p className="text-sm font-medium text-yellow-800 mb-1">
            Milestone Integration Required
          </p>
          <p className="text-xs text-yellow-700">
            This camera needs a <code className="bg-yellow-100 px-1 rounded">milestone_device_id</code> to use recording features.
          </p>
        </div>
      )}

      {/* Quick Actions */}
      {selectedCamera.milestone_device_id && (
        <div className="p-3 space-y-2">
          <Button
            onClick={handleOpenRecordings}
            variant="secondary"
            className="w-full flex items-center justify-center gap-2"
          >
            <Film className="w-4 h-4" />
            View Recordings
          </Button>
        </div>
      )}

      {/* Recordings Dialog */}
      <Dialog open={showRecordings} onOpenChange={setShowRecordings}>
        <DialogContent onClose={() => setShowRecordings(false)} className="max-w-6xl max-h-[90vh]">
          <DialogHeader>
            <DialogTitle>
              Recordings - {selectedCamera.name}
            </DialogTitle>
          </DialogHeader>

          <DialogBody className="space-y-4">
            {/* Date Range Selector */}
            <div className="flex items-center gap-3">
              <div className="flex-1">
                <label className="block text-xs font-medium text-gray-700 mb-1">
                  Start Time
                </label>
                <input
                  type="datetime-local"
                  value={queryStartTime.toISOString().slice(0, 16)}
                  onChange={(e) => setQueryStartTime(new Date(e.target.value))}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm"
                />
              </div>
              <div className="flex-1">
                <label className="block text-xs font-medium text-gray-700 mb-1">
                  End Time
                </label>
                <input
                  type="datetime-local"
                  value={queryEndTime.toISOString().slice(0, 16)}
                  onChange={(e) => setQueryEndTime(new Date(e.target.value))}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm"
                />
              </div>
              <div className="self-end">
                <Button
                  onClick={handleOpenRecordings}
                  variant="primary"
                  className="flex items-center gap-2"
                >
                  <Clock className="w-4 h-4" />
                  Query
                </Button>
              </div>
            </div>

            {/* Video Player */}
            {playbackTime && (
              <RecordingPlayer
                cameraId={selectedCamera.id}
                startTime={queryStartTime}
                endTime={queryEndTime}
                initialPlaybackTime={playbackTime}
                onPlaybackTimeChange={setPlaybackTime}
                className="aspect-video"
              />
            )}

            {/* Timeline */}
            <RecordingTimeline
              cameraId={selectedCamera.id}
              startTime={queryStartTime}
              endTime={queryEndTime}
              timelineData={timelineData}
              currentPlaybackTime={playbackTime}
              onSeek={handleSeek}
            />

            {/* Recording Segments List */}
            {timelineData && timelineData.sequences && (
              <div className="border border-gray-200 rounded-lg">
                <div className="p-3 bg-gray-50 border-b border-gray-200">
                  <h4 className="text-sm font-semibold text-gray-900">
                    Recording Segments ({timelineData.sequences.length})
                  </h4>
                </div>
                <div className="max-h-60 overflow-y-auto divide-y divide-gray-200">
                  {timelineData.sequences.map((segment: any, index: number) => (
                    <div
                      key={segment.sequenceId}
                      className="p-3 hover:bg-gray-50 cursor-pointer transition-colors"
                      onClick={() => handleSeek(new Date(segment.startTime))}
                    >
                      <div className="flex items-center justify-between">
                        <div className="flex-1">
                          <div className="text-sm font-medium text-gray-900">
                            Segment {index + 1}
                          </div>
                          <div className="text-xs text-gray-500">
                            {new Date(segment.startTime).toLocaleString()} -{' '}
                            {new Date(segment.endTime).toLocaleString()}
                          </div>
                        </div>
                        <div className="text-xs text-gray-600">
                          {Math.floor(segment.durationSeconds / 60)} min
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </DialogBody>
        </DialogContent>
      </Dialog>
    </div>
  );
}
