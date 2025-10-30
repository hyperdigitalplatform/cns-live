import React, { useRef, useState, useEffect } from 'react';
import { cn } from '@/utils/cn';

interface RecordingSequence {
  sequenceId: string;
  startTime: string;
  endTime: string;
  durationSeconds: number;
}

interface NavigationSliderProps {
  startTime: Date;
  endTime: Date;
  currentTime: Date;
  sequences: RecordingSequence[];
  onSeek: (time: Date) => void;
  className?: string;
}

export function NavigationSlider({
  startTime,
  endTime,
  currentTime,
  sequences,
  onSeek,
  className,
}: NavigationSliderProps) {
  const sliderRef = useRef<HTMLDivElement>(null);
  const [isDragging, setIsDragging] = useState(false);

  const totalDuration = endTime.getTime() - startTime.getTime();

  // Calculate position percentage for a timestamp
  const getPositionPercent = (timestamp: Date | string): number => {
    const time = typeof timestamp === 'string' ? new Date(timestamp).getTime() : timestamp.getTime();
    return ((time - startTime.getTime()) / totalDuration) * 100;
  };

  // Calculate timestamp from position
  const getTimeFromPosition = (clientX: number): Date => {
    if (!sliderRef.current) return startTime;

    const rect = sliderRef.current.getBoundingClientRect();
    const relativeX = clientX - rect.left;
    const percentage = Math.max(0, Math.min(1, relativeX / rect.width));
    const timestamp = startTime.getTime() + (totalDuration * percentage);

    return new Date(timestamp);
  };

  // Handle mouse/touch move
  const handleMove = (clientX: number) => {
    const newTime = getTimeFromPosition(clientX);
    onSeek(newTime);
  };

  // Handle mouse down
  const handleMouseDown = (e: React.MouseEvent) => {
    setIsDragging(true);
    handleMove(e.clientX);
  };

  // Handle mouse move
  useEffect(() => {
    if (!isDragging) return;

    const handleMouseMove = (e: MouseEvent) => {
      handleMove(e.clientX);
    };

    const handleMouseUp = () => {
      setIsDragging(false);
    };

    window.addEventListener('mousemove', handleMouseMove);
    window.addEventListener('mouseup', handleMouseUp);

    return () => {
      window.removeEventListener('mousemove', handleMouseMove);
      window.removeEventListener('mouseup', handleMouseUp);
    };
  }, [isDragging]);

  // Handle click on track
  const handleTrackClick = (e: React.MouseEvent) => {
    if (e.target === sliderRef.current || (e.target as HTMLElement).classList.contains('slider-track-inner')) {
      handleMove(e.clientX);
    }
  };

  const currentPositionPercent = getPositionPercent(currentTime);

  // Format time for labels
  const formatTime = (date: Date): string => {
    return date.toLocaleTimeString('en-US', {
      hour: '2-digit',
      minute: '2-digit',
      hour12: false,
    });
  };

  return (
    <div className={cn('flex items-center gap-2', className)}>
      {/* Start time label */}
      <span className="text-xs text-gray-600 font-mono min-w-[40px]">
        {formatTime(startTime)}
      </span>

      {/* Slider track */}
      <div
        ref={sliderRef}
        className="relative flex-1 h-6 cursor-pointer select-none"
        onMouseDown={handleMouseDown}
        onClick={handleTrackClick}
      >
        {/* Track background */}
        <div className="absolute inset-x-0 top-1/2 -translate-y-1/2 h-2 bg-gray-200 rounded-full slider-track-inner">
          {/* Recording sequence overlays */}
          {sequences.map((seq) => {
            const startPercent = getPositionPercent(seq.startTime);
            const endPercent = getPositionPercent(seq.endTime);
            const width = endPercent - startPercent;

            return (
              <div
                key={seq.sequenceId}
                className="absolute h-full bg-green-500/40 rounded-full"
                style={{
                  left: `${startPercent}%`,
                  width: `${width}%`,
                }}
              />
            );
          })}
        </div>

        {/* Draggable handle */}
        <div
          className={cn(
            'absolute top-1/2 -translate-y-1/2 -translate-x-1/2 w-4 h-4 bg-blue-600 border-2 border-white rounded-full shadow-lg transition-transform',
            isDragging && 'scale-125'
          )}
          style={{ left: `${currentPositionPercent}%` }}
        >
          {/* Handle tooltip */}
          <div className="absolute bottom-full mb-2 left-1/2 -translate-x-1/2 px-2 py-1 bg-gray-900 text-white text-xs rounded whitespace-nowrap pointer-events-none opacity-0 group-hover:opacity-100 transition-opacity">
            {currentTime.toLocaleTimeString('en-US', {
              hour: '2-digit',
              minute: '2-digit',
              second: '2-digit',
              hour12: false,
            })}
          </div>
        </div>

        {/* Sequence markers */}
        {sequences.map((seq, index) => {
          const startPercent = getPositionPercent(seq.startTime);

          return (
            <div
              key={`marker-${seq.sequenceId}`}
              className="absolute top-0 w-px h-full bg-blue-400"
              style={{ left: `${startPercent}%` }}
              title={`Sequence ${index + 1}: ${new Date(seq.startTime).toLocaleString()}`}
            >
              {/* Sequence number badge */}
              <div className="absolute -top-1 left-1/2 -translate-x-1/2 w-4 h-4 bg-blue-500 text-white text-[8px] flex items-center justify-center rounded-full font-bold">
                {index + 1}
              </div>
            </div>
          );
        })}
      </div>

      {/* End time label */}
      <span className="text-xs text-gray-600 font-mono min-w-[40px]">
        {formatTime(endTime)}
      </span>
    </div>
  );
}
