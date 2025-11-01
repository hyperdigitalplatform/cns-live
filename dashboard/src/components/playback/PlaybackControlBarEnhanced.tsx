import React, { useState, useRef, useEffect, useCallback, useMemo } from 'react';
import { createPortal } from 'react-dom';
import { Play, Pause, ChevronLeft, ChevronRight, ZoomIn, ChevronUp, ChevronDown, Calendar, Clock } from 'lucide-react';
import { cn } from '@/utils/cn';
import { TimePickerDialog } from './TimePickerDialog';
import { TimelineTicks } from './TimelineTicks';

interface RecordingSequence {
  sequenceId: string;
  startTime: string;
  endTime: string;
  durationSeconds: number;
}

interface PlaybackControlBarEnhancedProps {
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

// Enhanced zoom levels with granular control (1 min to 1 week)
const ZOOM_LEVELS = [
  { hours: 1/60,  label: '1 min',  majorTickMs: 10 * 1000,      minorTickMs: 2 * 1000 },       // Major: 10s, Minor: 2s
  { hours: 5/60,  label: '5 min',  majorTickMs: 60 * 1000,      minorTickMs: 10 * 1000 },      // Major: 1m, Minor: 10s
  { hours: 10/60, label: '10 min', majorTickMs: 1 * 60 * 1000,  minorTickMs: 15 * 1000 },      // Major: 1m, Minor: 15s
  { hours: 0.5,   label: '30 min', majorTickMs: 5 * 60 * 1000,  minorTickMs: 1 * 60 * 1000 },  // Major: 5m, Minor: 1m
  { hours: 1,     label: '1 hr',   majorTickMs: 5 * 60 * 1000,  minorTickMs: 1 * 60 * 1000 },  // Major: 5m, Minor: 1m
  { hours: 2,     label: '2 hr',   majorTickMs: 10 * 60 * 1000, minorTickMs: 2 * 60 * 1000 },  // Major: 10m, Minor: 2m
  { hours: 8,     label: '8 hr',   majorTickMs: 60 * 60 * 1000, minorTickMs: 15 * 60 * 1000 }, // Major: 1h, Minor: 15m
  { hours: 16,    label: '16 hr',  majorTickMs: 2 * 60 * 60 * 1000, minorTickMs: 30 * 60 * 1000 }, // Major: 2h, Minor: 30m
  { hours: 24,    label: '1 d',    majorTickMs: 4 * 60 * 60 * 1000, minorTickMs: 1 * 60 * 60 * 1000 }, // Major: 4h, Minor: 1h
  { hours: 168,   label: '1 wk',   majorTickMs: 24 * 60 * 60 * 1000, minorTickMs: 6 * 60 * 60 * 1000 } // Major: 1d, Minor: 6h
];

// Buffer multiplier for smooth scrolling (3x = 1.5x on each side)
const BUFFER_MULTIPLIER = 3;

export function PlaybackControlBarEnhanced({
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
}: PlaybackControlBarEnhancedProps) {
  const [isExpanded, setIsExpanded] = useState(false);
  const [showZoomMenu, setShowZoomMenu] = useState(false);
  const [showTimePicker, setShowTimePicker] = useState(false);
  const timelineRef = useRef<HTMLDivElement>(null);
  const contentWrapperRef = useRef<HTMLDivElement>(null);
  const zoomButtonRef = useRef<HTMLButtonElement>(null);
  const [hoveredTime, setHoveredTime] = useState<Date | null>(null);
  const [mousePosition, setMousePosition] = useState({ x: 0, y: 0 });
  const [isDragging, setIsDragging] = useState(false);
  const [zoomMenuPosition, setZoomMenuPosition] = useState({ top: 0, left: 0, width: 0 });
  const [timelineOffset, setTimelineOffset] = useState(0);
  const [isAnimating, setIsAnimating] = useState(false);
  const previousZoomLevelRef = useRef(zoomLevel);

  // Find zoom configuration
  const currentZoomConfig = useMemo(() => {
    const zoomIndex = ZOOM_LEVELS.findIndex(z => z.hours * 60 === zoomLevel);
    return zoomIndex >= 0 ? ZOOM_LEVELS[zoomIndex] : ZOOM_LEVELS[4]; // Default to 1hr
  }, [zoomLevel]);

  // Calculate buffer times for smooth scrolling
  const { bufferStart, bufferEnd, visibleDuration } = useMemo(() => {
    const visible = endTime.getTime() - startTime.getTime();
    const halfBuffer = (visible * (BUFFER_MULTIPLIER - 1)) / 2;

    return {
      bufferStart: new Date(startTime.getTime() - halfBuffer),
      bufferEnd: new Date(endTime.getTime() + halfBuffer),
      visibleDuration: visible
    };
  }, [startTime, endTime]);

  const totalBufferDuration = bufferEnd.getTime() - bufferStart.getTime();

  // Smooth scrolling algorithm (from reference lines 1356-1405)
  const updateTimelineScroll = useCallback((targetTime: Date) => {
    if (!contentWrapperRef.current) return;

    // Clamp to current time (cannot go into future)
    const now = new Date();
    const clampedTime = targetTime > now ? now : targetTime;

    // Check if target is outside buffer (need to reload timeline)
    const bufferMargin = visibleDuration * 0.3; // 30% margin
    const isInBuffer = clampedTime >= new Date(bufferStart.getTime() + bufferMargin) &&
                      clampedTime <= new Date(bufferEnd.getTime() - bufferMargin);

    if (!isInBuffer) {
      // Target is outside buffer - parent should reload timeline
      // For now, just clamp to buffer boundaries
      console.log('Target outside buffer, would reload timeline here');
      return;
    }

    // Calculate smooth CSS transform scroll
    const containerWidth = contentWrapperRef.current.offsetWidth;
    const elapsed = clampedTime.getTime() - bufferStart.getTime();
    const percentComplete = elapsed / totalBufferDuration;

    // Calculate offset: target time should appear at center (50%) of viewport
    const targetPixelPosition = percentComplete * containerWidth;
    const centerPixelPosition = containerWidth / 2;
    const scrollOffset = centerPixelPosition - targetPixelPosition;

    // Apply smooth CSS transform
    setTimelineOffset(scrollOffset);
  }, [bufferStart, bufferEnd, totalBufferDuration, visibleDuration]);

  // Update scroll when currentTime changes
  useEffect(() => {
    if (isPlaying) {
      updateTimelineScroll(currentTime);
    }
  }, [currentTime, isPlaying, updateTimelineScroll]);

  // Calculate position percentage within buffer
  const getPositionPercent = useCallback((timestamp: Date | string): number => {
    const time = typeof timestamp === 'string' ? new Date(timestamp).getTime() : timestamp.getTime();
    return ((time - bufferStart.getTime()) / totalBufferDuration) * 100;
  }, [bufferStart, totalBufferDuration]);

  // Calculate time at mouse position
  const getTimeAtPosition = useCallback((x: number): Date => {
    if (!timelineRef.current) return startTime;

    const rect = timelineRef.current.getBoundingClientRect();
    const relativeX = x - rect.left;
    const percentage = Math.max(0, Math.min(1, relativeX / rect.width));

    // Account for scroll offset
    const containerWidth = rect.width;
    const scrolledPercentage = ((relativeX - timelineOffset) / containerWidth);
    const timestamp = bufferStart.getTime() + (totalBufferDuration * scrolledPercentage);

    return new Date(timestamp);
  }, [bufferStart, totalBufferDuration, timelineOffset, startTime]);

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

  // Zoom change animation
  const handleZoomChange = useCallback((newZoom: number) => {
    setIsAnimating(true);
    onZoomChange(newZoom);

    // Reset animation flag after transition
    setTimeout(() => {
      setIsAnimating(false);
    }, 800);
  }, [onZoomChange]);

  // Detect zoom direction for animations
  const zoomDirection = useMemo(() => {
    const prevIndex = ZOOM_LEVELS.findIndex(z => z.hours * 60 === previousZoomLevelRef.current);
    const currIndex = ZOOM_LEVELS.findIndex(z => z.hours * 60 === zoomLevel);
    previousZoomLevelRef.current = zoomLevel;

    if (prevIndex < currIndex) return 'out'; // Zooming out (increasing hours)
    if (prevIndex > currIndex) return 'in';  // Zooming in (decreasing hours)
    return 'none';
  }, [zoomLevel]);

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

  // Generate time labels based on zoom level
  const generateTimeLabels = useCallback(() => {
    const labels: Date[] = [];
    const interval = currentZoomConfig.majorTickMs;

    // Calculate visible range (not buffer)
    const visibleStart = startTime;
    const visibleEnd = endTime;

    // Find first label time within visible range
    const firstLabelTime = new Date(Math.ceil(visibleStart.getTime() / interval) * interval);

    let currentLabel = firstLabelTime;
    while (currentLabel <= visibleEnd) {
      labels.push(new Date(currentLabel));
      currentLabel = new Date(currentLabel.getTime() + interval);
    }

    return labels;
  }, [startTime, endTime, currentZoomConfig]);

  const timeLabels = generateTimeLabels();
  const currentZoomLabel = currentZoomConfig.label;

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

  // Render future zone (green overlay for time that hasn't happened)
  const renderFutureZone = () => {
    const now = new Date();
    if (bufferEnd <= now) return null; // No future zone

    const futureStartTime = now > bufferStart ? now : bufferStart;
    const futureStartPercent = getPositionPercent(futureStartTime);
    const futureWidth = 100 - futureStartPercent;

    return (
      <div
        className="absolute top-0 bottom-0 bg-green-500/15 pointer-events-none z-5"
        style={{
          left: `${futureStartPercent}%`,
          width: `${futureWidth}%`,
        }}
      />
    );
  };

  return (
    <>
      <div className={cn('bg-gray-900/95 backdrop-blur-sm rounded-lg overflow-visible', className)}>
        {isExpanded ? (
          /* EXPANDED STATE */
          <div className="p-4 space-y-3">
            {/* Control Buttons */}
            <div className="flex items-center gap-2 flex-wrap">
              {/* Date/Time Picker */}
              <button
                onClick={() => setShowTimePicker(true)}
                className="inline-flex items-center gap-1.5 px-3 py-1.5 bg-white/10 hover:bg-white/20 text-white rounded-lg text-sm font-medium transition-colors"
                title="Select date and time"
              >
                <Calendar className="w-4 h-4" />
                <Clock className="w-4 h-4" />
                <span className="font-mono">
                  {currentTime.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })}
                  {' '}
                  {formatTimeMarker(currentTime)}:{currentTime.getSeconds().toString().padStart(2, '0')}
                </span>
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
                <span>{currentZoomConfig.label}</span>
              </button>

              {/* Scroll Right */}
              <button
                onClick={() => onScrollTimeline('right')}
                className="inline-flex items-center gap-1 px-3 py-1.5 bg-white/10 hover:bg-white/20 text-white rounded-lg text-sm font-medium transition-colors"
                title="Scroll timeline forward"
              >
                <span>{currentZoomConfig.label}</span>
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

            {/* Scrolling Timeline */}
            <div className="relative">
              {/* Fixed center time indicator */}
              <div className="absolute left-1/2 -translate-x-1/2 top-0 z-20 pointer-events-none">
                <div className="bg-white text-black px-3 py-1 rounded-md text-sm font-mono shadow-lg border border-gray-300">
                  <div className="flex items-center gap-2">
                    <span className="w-4 h-4 bg-blue-600 rounded flex items-center justify-center text-white text-xs">
                      ▶
                    </span>
                    <span>{formatTime(currentTime)}</span>
                  </div>
                </div>
              </div>

              {/* Timeline container */}
              <div
                ref={timelineRef}
                className={cn(
                  "relative mt-12 h-24 bg-gray-800 rounded-lg overflow-hidden",
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
                {/* Fixed center line */}
                <div className="absolute left-1/2 top-0 bottom-0 w-0.5 bg-white z-50 pointer-events-none" />

                {/* Scrollable content wrapper */}
                <div
                  ref={contentWrapperRef}
                  className="absolute inset-0 will-change-transform"
                  style={{
                    transform: `translateX(${timelineOffset}px)`,
                    transition: isAnimating
                      ? 'transform 0.8s cubic-bezier(0.4, 0, 0.2, 1)'
                      : 'transform 0.05s linear',
                  }}
                >
                  {/* Time labels (animated on zoom) */}
                  <div className="absolute inset-x-0 top-2 flex justify-between px-4">
                    {timeLabels.map((label, index) => {
                      const position = getPositionPercent(label);
                      return (
                        <div
                          key={index}
                          className={cn(
                            "absolute text-xs text-gray-400 font-mono -translate-x-1/2 transition-all duration-800",
                            isAnimating && 'opacity-0'
                          )}
                          style={{
                            left: `${position}%`,
                            transitionDelay: `${index * 20}ms`
                          }}
                        >
                          {formatTimeMarker(label)}
                        </div>
                      );
                    })}
                  </div>

                  {/* Timeline ticks */}
                  <div className="absolute inset-x-0 top-8 h-4 px-2">
                    <TimelineTicks
                      startTime={bufferStart}
                      endTime={bufferEnd}
                      majorTickMs={currentZoomConfig.majorTickMs}
                      minorTickMs={currentZoomConfig.minorTickMs}
                      isAnimating={isAnimating}
                      zoomDirection={zoomDirection}
                    />
                  </div>

                  {/* Timeline track with recordings */}
                  <div className="absolute inset-x-0 bottom-8 h-8 px-2">
                    <div className="relative w-full h-full bg-gray-700/50 rounded">
                      {/* Future zone overlay */}
                      {renderFutureZone()}

                      {/* Recording sequences */}
                      {sequences.map((seq) => {
                        const startPercent = getPositionPercent(seq.startTime);
                        const endPercent = getPositionPercent(seq.endTime);
                        const width = endPercent - startPercent;

                        // Check if sequence overlaps with visible range
                        if (endPercent < 0 || startPercent > 100) return null;

                        return (
                          <div
                            key={seq.sequenceId}
                            className="absolute h-full bg-orange-600 rounded transition-all duration-300"
                            style={{
                              left: `${Math.max(0, startPercent)}%`,
                              width: `${Math.min(100 - startPercent, width)}%`,
                            }}
                            title={`Recording: ${new Date(seq.startTime).toLocaleString()} - ${new Date(seq.endTime).toLocaleString()}`}
                          />
                        );
                      })}
                    </div>
                  </div>

                  {/* Bottom time markers */}
                  <div className="absolute inset-x-0 bottom-1 flex justify-between px-4">
                    {timeLabels.map((label, index) => {
                      const position = getPositionPercent(label);
                      return (
                        <div
                          key={`bottom-${index}`}
                          className={cn(
                            "absolute text-[10px] text-gray-500 font-mono -translate-x-1/2 transition-all duration-800",
                            isAnimating && 'opacity-0'
                          )}
                          style={{
                            left: `${position}%`,
                            transitionDelay: `${index * 20}ms`
                          }}
                        >
                          {formatTimeMarker(label)}
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
            </div>

            {/* Hide Button */}
            <div className="flex justify-center">
              <button
                onClick={() => setIsExpanded(false)}
                className="inline-flex items-center gap-1 px-4 py-1.5 bg-white/10 hover:bg-white/20 text-white rounded-lg text-sm font-medium transition-colors"
              >
                <ChevronDown className="w-4 h-4" />
                <span>Hide Timeline</span>
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

            {/* Thin timeline */}
            <div className="relative h-2 bg-gray-700 rounded-full overflow-hidden">
              {sequences.map((seq) => {
                const startPercent = getPositionPercent(seq.startTime);
                const endPercent = getPositionPercent(seq.endTime);
                const width = endPercent - startPercent;

                return (
                  <div
                    key={seq.sequenceId}
                    className="absolute h-full bg-orange-500/60"
                    style={{
                      left: `${startPercent}%`,
                      width: `${width}%`,
                    }}
                  />
                );
              })}

              {/* Playhead */}
              <div
                className="absolute top-0 bottom-0 w-0.5 bg-white"
                style={{ left: `${getPositionPercent(currentTime)}%` }}
              />
            </div>

            {/* Show Button */}
            <div className="flex justify-center pt-1">
              <button
                onClick={() => setIsExpanded(true)}
                className="inline-flex items-center gap-1 px-4 py-1 bg-white/10 hover:bg-white/20 text-white rounded-lg text-xs font-medium transition-colors"
              >
                <ChevronUp className="w-3 h-3" />
                <span>Show Timeline</span>
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

      {/* Zoom Menu Portal */}
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
              key={zoom.hours}
              onClick={() => {
                handleZoomChange(zoom.hours * 60);
                setShowZoomMenu(false);
              }}
              className={cn(
                'w-full px-4 py-2 text-left text-sm text-white hover:bg-white/10 transition-colors',
                Math.abs(zoomLevel - zoom.hours * 60) < 0.1 && 'bg-blue-600'
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
