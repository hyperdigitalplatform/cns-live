import React from 'react';
import { format } from 'date-fns';
import { cn } from '@/utils/cn';

interface PlaybackTimelineProps {
  startTime: Date;
  endTime: Date;
  segments: Array<{ start: Date; end: Date }>;
  currentTime: number; // seconds from startTime
  duration: number; // total duration in seconds
  onSeek: (time: number) => void;
}

export function PlaybackTimeline({
  startTime,
  endTime,
  segments,
  currentTime,
  duration,
  onSeek,
}: PlaybackTimelineProps) {
  const totalDuration = (endTime.getTime() - startTime.getTime()) / 1000; // in seconds

  // Calculate percentage positions for segments
  const segmentBars = segments.map((segment) => {
    const segmentStart =
      (segment.start.getTime() - startTime.getTime()) / 1000;
    const segmentEnd = (segment.end.getTime() - startTime.getTime()) / 1000;
    const startPercent = (segmentStart / totalDuration) * 100;
    const widthPercent = ((segmentEnd - segmentStart) / totalDuration) * 100;

    return {
      left: `${startPercent}%`,
      width: `${widthPercent}%`,
      segment,
    };
  });

  const currentPercent = (currentTime / totalDuration) * 100;

  const handleClick = (e: React.MouseEvent<HTMLDivElement>) => {
    const rect = e.currentTarget.getBoundingClientRect();
    const x = e.clientX - rect.left;
    const percent = x / rect.width;
    const time = percent * totalDuration;
    onSeek(time);
  };

  // Generate time markers (every hour or appropriate interval)
  const timeMarkers = [];
  const markerInterval = totalDuration > 7200 ? 3600 : 1800; // 1 hour or 30 min
  for (let i = 0; i <= totalDuration; i += markerInterval) {
    const time = new Date(startTime.getTime() + i * 1000);
    const percent = (i / totalDuration) * 100;
    timeMarkers.push({ time, percent });
  }

  return (
    <div className="space-y-2">
      {/* Time range labels */}
      <div className="flex justify-between text-xs text-gray-400">
        <span>{format(startTime, 'HH:mm:ss')}</span>
        <span className="text-white font-medium">
          Available Playback Content
        </span>
        <span>{format(endTime, 'HH:mm:ss')}</span>
      </div>

      {/* Timeline track */}
      <div
        onClick={handleClick}
        className="relative h-12 bg-gray-800 rounded-lg cursor-pointer group"
      >
        {/* Available segments (green bars) */}
        {segmentBars.map((bar, index) => (
          <div
            key={index}
            className="absolute top-0 h-full bg-green-600/40 group-hover:bg-green-600/60 transition-colors"
            style={{
              left: bar.left,
              width: bar.width,
            }}
            title={`${format(bar.segment.start, 'HH:mm:ss')} - ${format(
              bar.segment.end,
              'HH:mm:ss'
            )}`}
          >
            {/* Segment border for clarity */}
            <div className="h-full border-l border-r border-green-500/50" />
          </div>
        ))}

        {/* Time markers */}
        {timeMarkers.map((marker, index) => (
          <div
            key={index}
            className="absolute top-0 h-full border-l border-gray-600/50"
            style={{ left: `${marker.percent}%` }}
          >
            <div className="absolute top-full mt-1 -translate-x-1/2 text-xs text-gray-500">
              {format(marker.time, 'HH:mm')}
            </div>
          </div>
        ))}

        {/* Current position indicator */}
        <div
          className="absolute top-0 h-full w-0.5 bg-red-500 z-10 transition-all duration-100"
          style={{ left: `${currentPercent}%` }}
        >
          {/* Playhead */}
          <div className="absolute top-1/2 -translate-y-1/2 -translate-x-1/2">
            <div className="w-3 h-3 bg-red-500 rounded-full border-2 border-white shadow-lg" />
          </div>

          {/* Current time tooltip */}
          <div className="absolute -top-8 -translate-x-1/2 bg-black/90 px-2 py-1 rounded text-xs text-white whitespace-nowrap">
            {format(
              new Date(startTime.getTime() + currentTime * 1000),
              'HH:mm:ss'
            )}
          </div>
        </div>

        {/* Hover indicator */}
        <div className="absolute inset-0 pointer-events-none opacity-0 group-hover:opacity-100 transition-opacity">
          <div className="absolute inset-x-0 top-0 h-1 bg-white/10" />
        </div>
      </div>

      {/* Legend */}
      <div className="flex items-center gap-4 text-xs text-gray-400">
        <div className="flex items-center gap-1.5">
          <div className="w-4 h-2 bg-green-600/60 rounded" />
          <span>Available Video</span>
        </div>
        <div className="flex items-center gap-1.5">
          <div className="w-4 h-2 bg-gray-800 rounded" />
          <span>No Recording</span>
        </div>
        <div className="flex items-center gap-1.5">
          <div className="w-2 h-2 bg-red-500 rounded-full" />
          <span>Current Position</span>
        </div>
      </div>

      {/* Additional info */}
      {segments.length === 0 && (
        <div className="text-center text-yellow-500 text-sm mt-2">
          ⚠️ No recorded content available for selected time range
        </div>
      )}
    </div>
  );
}
