import React, { useEffect, useRef, useState } from 'react';
import { ZoomIn, ZoomOut, Maximize2 } from 'lucide-react';
import { cn } from '@/utils/cn';
import { Button } from './ui/Dialog';

interface RecordingSequence {
  sequenceId: string;
  startTime: string;
  endTime: string;
  durationSeconds: number;
  available: boolean;
  sizeBytes: number;
}

interface RecordingGap {
  startTime: string;
  endTime: string;
  durationSeconds: number;
}

interface TimelineData {
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

interface RecordingTimelineProps {
  cameraId: string;
  startTime: Date;
  endTime: Date;
  timelineData?: TimelineData;
  currentPlaybackTime?: Date;
  onSeek?: (timestamp: Date) => void;
  className?: string;
}

export function RecordingTimeline({
  cameraId,
  startTime,
  endTime,
  timelineData,
  currentPlaybackTime,
  onSeek,
  className,
}: RecordingTimelineProps) {
  const [zoomLevel, setZoomLevel] = useState(1);
  const [hoveredTime, setHoveredTime] = useState<Date | null>(null);
  const [mousePosition, setMousePosition] = useState({ x: 0, y: 0 });
  const timelineRef = useRef<HTMLDivElement>(null);
  const canvasRef = useRef<HTMLCanvasElement>(null);

  const totalDuration = endTime.getTime() - startTime.getTime();

  // Calculate time at mouse position
  const getTimeAtPosition = (x: number): Date => {
    if (!timelineRef.current) return startTime;

    const rect = timelineRef.current.getBoundingClientRect();
    const relativeX = x - rect.left;
    const percentage = relativeX / rect.width;
    const timestamp = startTime.getTime() + (totalDuration * percentage);

    return new Date(Math.max(startTime.getTime(), Math.min(endTime.getTime(), timestamp)));
  };

  // Handle mouse move for hover
  const handleMouseMove = (e: React.MouseEvent<HTMLDivElement>) => {
    const time = getTimeAtPosition(e.clientX);
    setHoveredTime(time);
    setMousePosition({ x: e.clientX, y: e.clientY });
  };

  // Handle click for seek
  const handleClick = (e: React.MouseEvent<HTMLDivElement>) => {
    if (!onSeek) return;
    const time = getTimeAtPosition(e.clientX);
    onSeek(time);
  };

  // Calculate position percentage for a timestamp
  const getPositionPercent = (timestamp: Date): number => {
    const time = timestamp.getTime();
    return ((time - startTime.getTime()) / totalDuration) * 100;
  };

  // Draw timeline on canvas
  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas || !timelineData) return;

    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    const width = canvas.width;
    const height = canvas.height;

    // Clear canvas
    ctx.clearRect(0, 0, width, height);

    // Draw background (no recording)
    ctx.fillStyle = '#E5E7EB'; // gray-200
    ctx.fillRect(0, 0, width, height);

    // Draw recording sequences
    timelineData.sequences.forEach((seq) => {
      const seqStart = new Date(seq.startTime);
      const seqEnd = new Date(seq.endTime);

      const startPercent = getPositionPercent(seqStart);
      const endPercent = getPositionPercent(seqEnd);

      const x = (startPercent / 100) * width;
      const segmentWidth = ((endPercent - startPercent) / 100) * width;

      // Draw recording segment
      ctx.fillStyle = seq.available ? '#10B981' : '#FCA5A5'; // green-500 or red-300
      ctx.fillRect(x, 0, segmentWidth, height);
    });

    // Draw gaps (optional, for emphasis)
    timelineData.gaps.forEach((gap) => {
      const gapStart = new Date(gap.startTime);
      const gapEnd = new Date(gap.endTime);

      const startPercent = getPositionPercent(gapStart);
      const endPercent = getPositionPercent(gapEnd);

      const x = (startPercent / 100) * width;
      const gapWidth = ((endPercent - startPercent) / 100) * width;

      // Draw gap indicator
      ctx.fillStyle = '#FEE2E2'; // red-100
      ctx.fillRect(x, 0, gapWidth, height);

      // Draw border
      ctx.strokeStyle = '#EF4444'; // red-500
      ctx.lineWidth = 1;
      ctx.strokeRect(x, 0, gapWidth, height);
    });

    // Draw playhead
    if (currentPlaybackTime) {
      const playheadPercent = getPositionPercent(currentPlaybackTime);
      const playheadX = (playheadPercent / 100) * width;

      ctx.strokeStyle = '#3B82F6'; // blue-500
      ctx.lineWidth = 2;
      ctx.beginPath();
      ctx.moveTo(playheadX, 0);
      ctx.lineTo(playheadX, height);
      ctx.stroke();

      // Draw playhead indicator
      ctx.fillStyle = '#3B82F6';
      ctx.beginPath();
      ctx.moveTo(playheadX, 0);
      ctx.lineTo(playheadX - 5, -8);
      ctx.lineTo(playheadX + 5, -8);
      ctx.closePath();
      ctx.fill();
    }

    // Draw hover indicator
    if (hoveredTime) {
      const hoverPercent = getPositionPercent(hoveredTime);
      const hoverX = (hoverPercent / 100) * width;

      ctx.strokeStyle = '#6B7280'; // gray-500
      ctx.lineWidth = 1;
      ctx.setLineDash([5, 3]);
      ctx.beginPath();
      ctx.moveTo(hoverX, 0);
      ctx.lineTo(hoverX, height);
      ctx.stroke();
      ctx.setLineDash([]);
    }
  }, [timelineData, currentPlaybackTime, hoveredTime, startTime, endTime, totalDuration]);

  // Generate time markers
  const generateTimeMarkers = () => {
    const markers: { time: Date; label: string }[] = [];
    const duration = totalDuration / 1000; // seconds

    // Determine marker interval based on duration
    let intervalSeconds: number;
    if (duration <= 3600) {
      // <= 1 hour: 5 min intervals
      intervalSeconds = 300;
    } else if (duration <= 21600) {
      // <= 6 hours: 30 min intervals
      intervalSeconds = 1800;
    } else if (duration <= 86400) {
      // <= 1 day: 1 hour intervals
      intervalSeconds = 3600;
    } else {
      // > 1 day: 3 hour intervals
      intervalSeconds = 10800;
    }

    let currentTime = new Date(Math.ceil(startTime.getTime() / (intervalSeconds * 1000)) * intervalSeconds * 1000);

    while (currentTime <= endTime) {
      markers.push({
        time: currentTime,
        label: currentTime.toLocaleTimeString('en-US', {
          hour: '2-digit',
          minute: '2-digit',
          hour12: false,
        }),
      });
      currentTime = new Date(currentTime.getTime() + intervalSeconds * 1000);
    }

    return markers;
  };

  const timeMarkers = generateTimeMarkers();

  // Format time for tooltip
  const formatTooltipTime = (time: Date): string => {
    return time.toLocaleString('en-US', {
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
      hour12: false,
    });
  };

  // Check if time has recording
  const hasRecordingAtTime = (time: Date): boolean => {
    if (!timelineData) return false;
    const timestamp = time.getTime();

    return timelineData.sequences.some((seq) => {
      const seqStart = new Date(seq.startTime).getTime();
      const seqEnd = new Date(seq.endTime).getTime();
      return timestamp >= seqStart && timestamp <= seqEnd;
    });
  };

  // Zoom controls
  const handleZoomIn = () => {
    setZoomLevel((prev) => Math.min(prev * 2, 16));
  };

  const handleZoomOut = () => {
    setZoomLevel((prev) => Math.max(prev / 2, 1));
  };

  const handleZoomReset = () => {
    setZoomLevel(1);
  };

  return (
    <div className={cn('space-y-3', className)}>
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-sm font-semibold text-gray-900">Recording Timeline</h3>
          {timelineData && (
            <p className="text-xs text-gray-500">
              Coverage: {(timelineData.coverage * 100).toFixed(1)}% •{' '}
              {timelineData.sequences.length} segment{timelineData.sequences.length !== 1 ? 's' : ''} •{' '}
              {timelineData.gaps.length} gap{timelineData.gaps.length !== 1 ? 's' : ''}
            </p>
          )}
        </div>

        {/* Zoom Controls */}
        <div className="flex items-center gap-1">
          <Button
            variant="secondary"
            size="sm"
            onClick={handleZoomOut}
            disabled={zoomLevel <= 1}
            className="p-1.5"
          >
            <ZoomOut className="w-4 h-4" />
          </Button>
          <span className="text-xs text-gray-600 w-12 text-center">{zoomLevel}x</span>
          <Button
            variant="secondary"
            size="sm"
            onClick={handleZoomIn}
            disabled={zoomLevel >= 16}
            className="p-1.5"
          >
            <ZoomIn className="w-4 h-4" />
          </Button>
          <Button
            variant="secondary"
            size="sm"
            onClick={handleZoomReset}
            disabled={zoomLevel === 1}
            className="p-1.5"
          >
            <Maximize2 className="w-4 h-4" />
          </Button>
        </div>
      </div>

      {/* Timeline Container */}
      <div className="border border-gray-200 rounded-lg overflow-hidden bg-white">
        {/* Canvas Timeline */}
        <div
          ref={timelineRef}
          className={cn(
            'relative h-12 cursor-pointer',
            onSeek && 'hover:bg-blue-50/30'
          )}
          onMouseMove={handleMouseMove}
          onMouseLeave={() => setHoveredTime(null)}
          onClick={handleClick}
          style={{ width: `${100 * zoomLevel}%` }}
        >
          <canvas
            ref={canvasRef}
            width={800 * zoomLevel}
            height={48}
            className="w-full h-full"
          />
        </div>

        {/* Time Markers */}
        <div className="relative h-6 border-t border-gray-200 bg-gray-50">
          <div
            className="relative h-full"
            style={{ width: `${100 * zoomLevel}%` }}
          >
            {timeMarkers.map((marker, index) => {
              const position = getPositionPercent(marker.time);
              return (
                <div
                  key={index}
                  className="absolute top-0 h-full flex items-center"
                  style={{ left: `${position}%` }}
                >
                  <div className="flex flex-col items-center">
                    <div className="w-px h-2 bg-gray-400" />
                    <span className="text-xs text-gray-600 mt-0.5">{marker.label}</span>
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      </div>

      {/* Legend */}
      <div className="flex items-center gap-4 text-xs text-gray-600">
        <div className="flex items-center gap-1.5">
          <div className="w-4 h-3 bg-green-500 rounded" />
          <span>Recording Available</span>
        </div>
        <div className="flex items-center gap-1.5">
          <div className="w-4 h-3 bg-red-100 border border-red-500 rounded" />
          <span>Gap</span>
        </div>
        <div className="flex items-center gap-1.5">
          <div className="w-4 h-3 bg-gray-200 rounded" />
          <span>No Data</span>
        </div>
        {currentPlaybackTime && (
          <div className="flex items-center gap-1.5">
            <div className="w-0.5 h-3 bg-blue-500" />
            <span>Playhead</span>
          </div>
        )}
      </div>

      {/* Hover Tooltip */}
      {hoveredTime && (
        <div
          className="fixed z-50 px-2 py-1 bg-gray-900 text-white text-xs rounded shadow-lg pointer-events-none"
          style={{
            left: mousePosition.x + 10,
            top: mousePosition.y - 30,
          }}
        >
          <div>{formatTooltipTime(hoveredTime)}</div>
          <div className="text-gray-300">
            {hasRecordingAtTime(hoveredTime) ? 'Recording Available' : 'No Recording'}
          </div>
        </div>
      )}

      {/* Date Range Display */}
      <div className="flex items-center justify-between text-xs text-gray-500">
        <div>
          Start: {startTime.toLocaleString('en-US', {
            month: 'short',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit',
          })}
        </div>
        <div>
          End: {endTime.toLocaleString('en-US', {
            month: 'short',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit',
          })}
        </div>
      </div>
    </div>
  );
}
