import React, { useState, useRef, useEffect } from 'react';
import { createPortal } from 'react-dom';
import { Play, Pause, ChevronLeft, ChevronRight, ZoomIn, ChevronUp, ChevronDown, Calendar, Clock } from 'lucide-react';
import { cn } from '@/utils/cn';
import { NavigationSlider } from './NavigationSlider';
import { TimePickerDialog } from './TimePickerDialog';

interface RecordingSequence {
  sequenceId: string;
  startTime: string;
  endTime: string;
  durationSeconds: number;
}

interface PlaybackControlBarProps {
  startTime: Date;
  endTime: Date;
  currentTime: Date;
  sequences: RecordingSequence[];
  isPlaying: boolean;
  zoomLevel: number;
  onPlayPause: () => void;
  onSeek: (time: Date) => void;
  onScrollTimeline: (direction: 'left' | 'right') => void;
  onZoomChange: (zoom: number) => void;
  hasRecording: boolean;
  className?: string;
}

const ZOOM_LEVELS = [
  { label: '1 hour', value: 1 },
  { label: '4 hours', value: 4 },
  { label: '12 hours', value: 12 },
  { label: '24 hours', value: 24 },
  { label: '7 days', value: 168 },
  { label: '30 days', value: 720 },
];

export function PlaybackControlBar({
  startTime,
  endTime,
  currentTime,
  sequences,
  isPlaying,
  zoomLevel,
  onPlayPause,
  onSeek,
  onScrollTimeline,
  onZoomChange,
  hasRecording,
  className,
}: PlaybackControlBarProps) {
  const [isExpanded, setIsExpanded] = useState(false);
  const [showZoomMenu, setShowZoomMenu] = useState(false);
  const [showTimePicker, setShowTimePicker] = useState(false);
  const timelineRef = useRef<HTMLDivElement>(null);
  const zoomButtonRef = useRef<HTMLButtonElement>(null);
  const [hoveredTime, setHoveredTime] = useState<Date | null>(null);
  const [mousePosition, setMousePosition] = useState({ x: 0, y: 0 });
  const [isDragging, setIsDragging] = useState(false);
  const [zoomMenuPosition, setZoomMenuPosition] = useState({ top: 0, left: 0, width: 0 });

  const totalDuration = endTime.getTime() - startTime.getTime();

  // Calculate position percentage
  const getPositionPercent = (timestamp: Date | string): number => {
    const time = typeof timestamp === 'string' ? new Date(timestamp).getTime() : timestamp.getTime();
    return ((time - startTime.getTime()) / totalDuration) * 100;
  };

  // Calculate time at mouse position
  const getTimeAtPosition = (x: number): Date => {
    if (!timelineRef.current) return startTime;

    const rect = timelineRef.current.getBoundingClientRect();
    const relativeX = x - rect.left;
    const percentage = Math.max(0, Math.min(1, relativeX / rect.width));
    const timestamp = startTime.getTime() + (totalDuration * percentage);

    return new Date(timestamp);
  };

  // Handle timeline click
  const handleTimelineClick = (e: React.MouseEvent) => {
    if (!isExpanded || isDragging) return;
    const time = getTimeAtPosition(e.clientX);
    onSeek(time);
  };

  // Handle timeline hover
  const handleTimelineHover = (e: React.MouseEvent) => {
    if (!isExpanded) return;
    const time = getTimeAtPosition(e.clientX);
    setHoveredTime(time);
    setMousePosition({ x: e.clientX, y: e.clientY });

    // If dragging, seek to new position
    if (isDragging) {
      onSeek(time);
    }
  };

  // Handle mouse down on timeline - start dragging
  const handleTimelineMouseDown = (e: React.MouseEvent) => {
    if (!isExpanded) return;
    setIsDragging(true);
    const time = getTimeAtPosition(e.clientX);
    onSeek(time);
  };

  // Handle mouse up - stop dragging
  const handleMouseUp = () => {
    setIsDragging(false);
  };

  // Add global mouse up listener
  useEffect(() => {
    if (isDragging) {
      window.addEventListener('mouseup', handleMouseUp);
      return () => window.removeEventListener('mouseup', handleMouseUp);
    }
  }, [isDragging]);

  // Format time display
  const formatTime = (date: Date): string => {
    return date.toLocaleString('en-US', {
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
      hour12: false,
    });
  };

  // Format time marker
  const formatTimeMarker = (date: Date): string => {
    return date.toLocaleTimeString('en-US', {
      hour: '2-digit',
      minute: '2-digit',
      hour12: false,
    });
  };

  // Generate time markers
  const generateTimeMarkers = () => {
    const markers: Date[] = [];
    const duration = totalDuration / 1000; // seconds

    let intervalSeconds: number;
    if (duration <= 3600) {
      intervalSeconds = 300; // 5 min
    } else if (duration <= 21600) {
      intervalSeconds = 1800; // 30 min
    } else if (duration <= 86400) {
      intervalSeconds = 3600; // 1 hour
    } else {
      intervalSeconds = 10800; // 3 hours
    }

    let currentTime = new Date(Math.ceil(startTime.getTime() / (intervalSeconds * 1000)) * intervalSeconds * 1000);

    while (currentTime <= endTime) {
      markers.push(new Date(currentTime));
      currentTime = new Date(currentTime.getTime() + intervalSeconds * 1000);
    }

    return markers;
  };

  const timeMarkers = generateTimeMarkers();
  const currentZoomLabel = ZOOM_LEVELS.find(z => z.value === zoomLevel)?.label || `${zoomLevel}h`;

  // Auto-collapse after inactivity
  useEffect(() => {
    if (!isExpanded || !isPlaying) return;

    const timer = setTimeout(() => {
      setIsExpanded(false);
    }, 5000);

    return () => clearTimeout(timer);
  }, [isExpanded, isPlaying]);

  // Update zoom menu position when it opens
  useEffect(() => {
    if (showZoomMenu && zoomButtonRef.current) {
      const rect = zoomButtonRef.current.getBoundingClientRect();
      setZoomMenuPosition({
        top: rect.top,
        left: rect.left,
        width: rect.width,
      });
    }
  }, [showZoomMenu]);

  // Close zoom menu when clicking outside
  useEffect(() => {
    if (!showZoomMenu) return;

    const handleClickOutside = (e: MouseEvent) => {
      if (zoomButtonRef.current && !zoomButtonRef.current.contains(e.target as Node)) {
        setShowZoomMenu(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, [showZoomMenu]);

  return (
    <>
      <div className={cn('bg-gray-900/95 backdrop-blur-sm rounded-lg overflow-hidden', className)}>
        {isExpanded ? (
          /* EXPANDED STATE */
          <div className="p-4 space-y-3">
            {/* Control Buttons */}
            <div className="flex items-center gap-2 flex-wrap">
              {/* Date Picker */}
              <button
                onClick={() => setShowTimePicker(true)}
                className="inline-flex items-center gap-1.5 px-3 py-1.5 bg-white/10 hover:bg-white/20 text-white rounded-lg text-sm font-medium transition-colors"
                title="Select date"
              >
                <Calendar className="w-4 h-4" />
                <span>{currentTime.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })}</span>
              </button>

              {/* Time Picker */}
              <button
                onClick={() => setShowTimePicker(true)}
                className="inline-flex items-center gap-1.5 px-3 py-1.5 bg-white/10 hover:bg-white/20 text-white rounded-lg text-sm font-mono transition-colors"
                title="Select time"
              >
                <Clock className="w-4 h-4" />
                <span>{formatTimeMarker(currentTime)}:{currentTime.getSeconds().toString().padStart(2, '0')}</span>
              </button>

              {/* Play/Pause */}
              <button
                onClick={onPlayPause}
                className="inline-flex items-center gap-1.5 px-4 py-1.5 bg-blue-600 hover:bg-blue-700 text-white rounded-lg text-sm font-medium transition-colors"
                title={isPlaying ? 'Pause' : 'Play'}
              >
                {isPlaying ? <Pause className="w-4 h-4" /> : <Play className="w-4 h-4" />}
                <span>{isPlaying ? 'Pause' : 'Play'}</span>
              </button>

              {/* Scroll Left */}
              <button
                onClick={() => onScrollTimeline('left')}
                className="inline-flex items-center gap-1 px-3 py-1.5 bg-white/10 hover:bg-white/20 text-white rounded-lg text-sm font-medium transition-colors"
                title="Scroll timeline backward"
              >
                <ChevronLeft className="w-4 h-4" />
                <span>1hr</span>
              </button>

              {/* Scroll Right */}
              <button
                onClick={() => onScrollTimeline('right')}
                className="inline-flex items-center gap-1 px-3 py-1.5 bg-white/10 hover:bg-white/20 text-white rounded-lg text-sm font-medium transition-colors"
                title="Scroll timeline forward"
              >
                <span>1hr</span>
                <ChevronRight className="w-4 h-4" />
              </button>

              {/* Zoom Control */}
              <div className="relative">
                <button
                  ref={zoomButtonRef}
                  onClick={() => setShowZoomMenu(!showZoomMenu)}
                  className="inline-flex items-center gap-1.5 px-3 py-1.5 bg-white/10 hover:bg-white/20 text-white rounded-lg text-sm font-medium transition-colors"
                  title="Zoom level"
                >
                  <ZoomIn className="w-4 h-4" />
                  <span>{currentZoomLabel}</span>
                </button>
              </div>
            </div>

            {/* Detailed Timeline */}
            <div
              ref={timelineRef}
              className={cn(
                "relative h-16 bg-gray-800 rounded-lg overflow-hidden",
                isDragging ? "cursor-grabbing" : "cursor-pointer"
              )}
              onClick={handleTimelineClick}
              onMouseDown={handleTimelineMouseDown}
              onMouseMove={handleTimelineHover}
              onMouseLeave={() => {
                setHoveredTime(null);
                setIsDragging(false);
              }}
            >
              {/* Time markers */}
              <div className="absolute inset-x-0 bottom-0 flex justify-between px-2 pb-1">
                {timeMarkers.map((marker, index) => (
                  <div key={index} className="text-xs text-gray-400 font-mono">
                    {formatTimeMarker(marker)}
                  </div>
                ))}
              </div>

              {/* Recording bars */}
              <div className="absolute inset-0 flex items-center px-2">
                <div className="relative w-full h-8 bg-gray-700 rounded">
                  {sequences.map((seq) => {
                    const startPercent = getPositionPercent(seq.startTime);
                    const endPercent = getPositionPercent(seq.endTime);
                    const width = endPercent - startPercent;

                    return (
                      <div
                        key={seq.sequenceId}
                        className="absolute h-full bg-green-500 rounded"
                        style={{
                          left: `${startPercent}%`,
                          width: `${width}%`,
                        }}
                      />
                    );
                  })}

                  {/* Playhead */}
                  <div
                    className="absolute top-0 bottom-0 w-0.5 bg-blue-500"
                    style={{ left: `${getPositionPercent(currentTime)}%` }}
                  >
                    <div className="absolute -top-2 left-1/2 -translate-x-1/2 w-3 h-3 bg-blue-500 rounded-full border-2 border-white" />
                  </div>

                  {/* Sequence markers */}
                  {sequences.map((seq, index) => {
                    const startPercent = getPositionPercent(seq.startTime);
                    return (
                      <div
                        key={`marker-${seq.sequenceId}`}
                        className="absolute top-0 bottom-0 w-px bg-blue-300"
                        style={{ left: `${startPercent}%` }}
                        title={`Sequence ${index + 1}`}
                      >
                        <div className="absolute -top-1 left-1/2 -translate-x-1/2 text-xs text-blue-300">
                          ▲
                        </div>
                      </div>
                    );
                  })}
                </div>
              </div>

              {/* Hover tooltip */}
              {hoveredTime && (
                <div
                  className="fixed z-50 px-2 py-1 bg-gray-900 text-white text-xs rounded shadow-lg pointer-events-none"
                  style={{
                    left: mousePosition.x + 10,
                    top: mousePosition.y - 40,
                  }}
                >
                  {formatTime(hoveredTime)}
                </div>
              )}
            </div>

            {/* Navigation Slider */}
            <NavigationSlider
              startTime={startTime}
              endTime={endTime}
              currentTime={currentTime}
              sequences={sequences}
              onSeek={onSeek}
            />

            {/* Hide Button */}
            <div className="flex justify-center">
              <button
                onClick={() => setIsExpanded(false)}
                className="inline-flex items-center gap-1 px-4 py-1.5 bg-white/10 hover:bg-white/20 text-white rounded-lg text-sm font-medium transition-colors"
              >
                <ChevronDown className="w-4 h-4" />
                <span>Hide</span>
              </button>
            </div>
          </div>
        ) : (
          /* COLLAPSED STATE */
          <div className="p-3 space-y-2">
            {/* Status Line */}
            <div className="flex items-center gap-2 text-sm">
              <span className="px-2 py-0.5 bg-blue-600 text-white rounded text-xs font-medium">
                PLAYBACK
              </span>
              <button
                onClick={() => setShowTimePicker(true)}
                className="text-white font-mono hover:text-blue-300 transition-colors"
              >
                {formatTime(currentTime)}
              </button>
              <span className="text-gray-400">•</span>
              <span className={cn('text-xs', hasRecording ? 'text-green-400' : 'text-red-400')}>
                {hasRecording ? '✓ Recording' : '✗ No Recording'}
              </span>
            </div>

            {/* Thin Timeline with Time Markers */}
            <div className="relative">
              {/* Time markers */}
              <div className="flex justify-between text-[10px] text-gray-500 font-mono mb-1">
                {timeMarkers.slice(0, 5).map((marker, index) => (
                  <span key={index}>{formatTimeMarker(marker)}</span>
                ))}
              </div>

              {/* Thin timeline */}
              <div className="relative h-2 bg-gray-700 rounded-full overflow-hidden">
                {sequences.map((seq) => {
                  const startPercent = getPositionPercent(seq.startTime);
                  const endPercent = getPositionPercent(seq.endTime);
                  const width = endPercent - startPercent;

                  return (
                    <div
                      key={seq.sequenceId}
                      className="absolute h-full bg-green-500/60"
                      style={{
                        left: `${startPercent}%`,
                        width: `${width}%`,
                      }}
                    />
                  );
                })}
              </div>
            </div>

            {/* Navigation Slider */}
            <NavigationSlider
              startTime={startTime}
              endTime={endTime}
              currentTime={currentTime}
              sequences={sequences}
              onSeek={onSeek}
            />

            {/* Show Button */}
            <div className="flex justify-center pt-1">
              <button
                onClick={() => setIsExpanded(true)}
                className="inline-flex items-center gap-1 px-4 py-1 bg-white/10 hover:bg-white/20 text-white rounded-lg text-xs font-medium transition-colors"
              >
                <ChevronUp className="w-3 h-3" />
                <span>Show</span>
              </button>
            </div>
          </div>
        )}
      </div>

      {/* Time Picker Dialog */}
      <TimePickerDialog
        open={showTimePicker}
        onOpenChange={setShowTimePicker}
        currentTime={currentTime}
        onTimeSelect={onSeek}
      />

      {/* Zoom Menu Portal - renders outside component hierarchy to avoid overflow clipping */}
      {showZoomMenu && createPortal(
        <div
          className="fixed bg-gray-800 rounded-lg shadow-xl overflow-hidden z-[9999] min-w-[140px]"
          style={{
            top: `${zoomMenuPosition.top - 8}px`,
            left: `${zoomMenuPosition.left}px`,
            transform: 'translateY(-100%)',
          }}
        >
          {ZOOM_LEVELS.map((zoom) => (
            <button
              key={zoom.value}
              onClick={() => {
                onZoomChange(zoom.value);
                setShowZoomMenu(false);
              }}
              className={cn(
                'w-full px-4 py-2 text-left text-sm text-white hover:bg-white/10 transition-colors',
                zoomLevel === zoom.value && 'bg-blue-600'
              )}
            >
              {zoom.label}
            </button>
          ))}
        </div>,
        document.body
      )}
    </>
  );
}
