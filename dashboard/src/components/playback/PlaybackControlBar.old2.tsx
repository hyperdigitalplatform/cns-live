import React, { useState, useRef, useEffect, useCallback, useMemo } from 'react';
import { Play, Pause, ChevronUp, ChevronDown } from 'lucide-react';
import { cn } from '@/utils/cn';

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
  onSpeedChange?: (speed: number) => void;
  className?: string;
}

// Zoom levels: 1 min to 1 week (from reference implementation)
const ZOOM_LEVELS = [
  { hours: 1/60,  label: '1 min',  majorTick: 10 * 1000,        minorTick: 2 * 1000 },
  { hours: 5/60,  label: '5 min',  majorTick: 60 * 1000,        minorTick: 10 * 1000 },
  { hours: 10/60, label: '10 min', majorTick: 1 * 60 * 1000,    minorTick: 15 * 1000 },
  { hours: 0.5,   label: '30 min', majorTick: 5 * 60 * 1000,    minorTick: 1 * 60 * 1000 },
  { hours: 1,     label: '1 hr',   majorTick: 5 * 60 * 1000,    minorTick: 1 * 60 * 1000 },
  { hours: 2,     label: '2 hr',   majorTick: 10 * 60 * 1000,   minorTick: 2 * 60 * 1000 },
  { hours: 8,     label: '8 hr',   majorTick: 60 * 60 * 1000,   minorTick: 15 * 60 * 1000 },
  { hours: 16,    label: '16 hr',  majorTick: 2 * 60 * 60 * 1000, minorTick: 30 * 60 * 1000 },
  { hours: 24,    label: '1 d',    majorTick: 4 * 60 * 60 * 1000, minorTick: 1 * 60 * 60 * 1000 },
  { hours: 168,   label: '1 wk',   majorTick: 24 * 60 * 60 * 1000, minorTick: 6 * 60 * 60 * 1000 }
];

// Speed options (from reference implementation)
const SPEED_OPTIONS = [0.25, 0.5, 1, 2, 4, 8, 16];

// Buffer multiplier for smooth scrolling
const BUFFER_MULTIPLIER = 3;

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
  onSpeedChange,
  className,
}: PlaybackControlBarProps) {
  // UI State
  const [isExpanded, setIsExpanded] = useState(false);
  const [showZoomMenu, setShowZoomMenu] = useState(false);
  const [showSpeedMenu, setShowSpeedMenu] = useState(false);
  const [showDatePicker, setShowDatePicker] = useState(false);
  const [currentSpeed, setCurrentSpeed] = useState(1);
  const [selectedDate, setSelectedDate] = useState(new Date());
  const [calendarDate, setCalendarDate] = useState(new Date());

  // Refs
  const contentWrapperRef = useRef<HTMLDivElement>(null);
  const timelineRef = useRef<HTMLDivElement>(null);

  // Timeline scroll offset
  const [timelineOffset, setTimelineOffset] = useState(0);

  // Find current zoom configuration
  const zoomConfig = useMemo(() => {
    const config = ZOOM_LEVELS.find(z => Math.abs(z.hours * 60 - zoomLevel) < 0.1);
    return config || ZOOM_LEVELS[4]; // Default to 1hr
  }, [zoomLevel]);

  // Calculate buffer times for smooth scrolling (3x buffer technique)
  const { bufferStart, bufferEnd, totalDuration } = useMemo(() => {
    const visible = endTime.getTime() - startTime.getTime();
    const halfBuffer = (visible * (BUFFER_MULTIPLIER - 1)) / 2;

    return {
      bufferStart: new Date(startTime.getTime() - halfBuffer),
      bufferEnd: new Date(endTime.getTime() + halfBuffer),
      totalDuration: visible * BUFFER_MULTIPLIER
    };
  }, [startTime, endTime]);

  // Smooth scrolling algorithm (from reference lines 1356-1405)
  const updateTimelineScroll = useCallback((targetTime: Date) => {
    if (!contentWrapperRef.current) return;

    const containerWidth = contentWrapperRef.current.offsetWidth;
    const elapsed = targetTime.getTime() - bufferStart.getTime();
    const percentComplete = elapsed / totalDuration;

    // Calculate offset: target time appears at center (50%)
    const targetPixelPosition = percentComplete * containerWidth;
    const centerPixelPosition = containerWidth / 2;
    const scrollOffset = centerPixelPosition - targetPixelPosition;

    setTimelineOffset(scrollOffset);
  }, [bufferStart, totalDuration]);

  // Update scroll when playing
  useEffect(() => {
    if (isPlaying) {
      updateTimelineScroll(currentTime);
    }
  }, [currentTime, isPlaying, updateTimelineScroll]);

  // Calculate position percentage
  const getPositionPercent = (timestamp: Date | string): number => {
    const time = typeof timestamp === 'string' ? new Date(timestamp).getTime() : timestamp.getTime();
    return ((time - bufferStart.getTime()) / totalDuration) * 100;
  };

  // Format time displays
  const formatTime = (date: Date): string => {
    const hours = date.getHours();
    const minutes = date.getMinutes();
    const seconds = date.getSeconds();
    const ampm = hours >= 12 ? 'PM' : 'AM';
    const displayHours = hours % 12 || 12;
    return `${displayHours}:${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')} ${ampm}`;
  };

  const formatDate = (date: Date): string => {
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day = String(date.getDate()).padStart(2, '0');
    return `${year}-${month}-${day}`;
  };

  const formatTimeLabel = (date: Date): string => {
    const hours = date.getHours();
    const minutes = date.getMinutes();
    const ampm = hours >= 12 ? 'PM' : 'AM';
    const displayHours = hours % 12 || 12;

    if (zoomConfig.hours >= 168) {
      // 1 week - show date
      const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
      return `${months[date.getMonth()]} ${date.getDate()}`;
    } else if (zoomConfig.hours < 0.1) {
      // < 10 min - show seconds
      const seconds = date.getSeconds();
      return `${displayHours}:${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`;
    } else {
      // Default - show hours:minutes
      return `${displayHours}:${minutes.toString().padStart(2, '0')} ${ampm}`;
    }
  };

  // Generate time labels (for visible 1x range, positioned in 3x buffer coordinates)
  const timeLabels = useMemo(() => {
    const labels: Date[] = [];
    const interval = zoomConfig.majorTick;
    const firstTime = new Date(Math.ceil(startTime.getTime() / interval) * interval);

    let current = firstTime;
    while (current <= endTime) {
      labels.push(new Date(current));
      current = new Date(current.getTime() + interval);
    }

    return labels;
  }, [startTime, endTime, zoomConfig]);

  // Generate tick marks
  const renderTicks = () => {
    const ticks: JSX.Element[] = [];

    // Minor ticks
    const minorInterval = zoomConfig.minorTick;
    const firstMinor = new Date(Math.ceil(bufferStart.getTime() / minorInterval) * minorInterval);
    let current = firstMinor;

    while (current <= bufferEnd) {
      const position = getPositionPercent(current);
      if (position >= 0 && position <= 100) {
        ticks.push(
          <div
            key={`minor-${current.getTime()}`}
            className="absolute bottom-0 w-px bg-white/20"
            style={{ left: `${position}%`, height: '8px' }}
          />
        );
      }
      current = new Date(current.getTime() + minorInterval);
    }

    // Major ticks
    const majorInterval = zoomConfig.majorTick;
    const firstMajor = new Date(Math.ceil(bufferStart.getTime() / majorInterval) * majorInterval);
    current = firstMajor;

    while (current <= bufferEnd) {
      const position = getPositionPercent(current);
      if (position >= 0 && position <= 100) {
        ticks.push(
          <div
            key={`major-${current.getTime()}`}
            className="absolute bottom-0 w-px bg-white/50"
            style={{ left: `${position}%`, height: '15px' }}
          />
        );
      }
      current = new Date(current.getTime() + majorInterval);
    }

    return ticks;
  };

  // Render future zone
  const renderFutureZone = () => {
    const now = new Date();
    if (bufferEnd <= now) return null;

    const futureStart = now > bufferStart ? now : bufferStart;
    const startPercent = getPositionPercent(futureStart);
    const width = 100 - startPercent;

    return (
      <div
        className="absolute top-0 bottom-0 bg-green-500/15 pointer-events-none z-5"
        style={{ left: `${startPercent}%`, width: `${width}%` }}
      />
    );
  };

  // Handle speed change
  const handleSpeedChange = (speed: number) => {
    setCurrentSpeed(speed);
    setShowSpeedMenu(false);
    onSpeedChange?.(speed);
  };

  // Handle timeline click
  const handleTimelineClick = (e: React.MouseEvent) => {
    if (!timelineRef.current) return;

    const rect = timelineRef.current.getBoundingClientRect();
    const x = e.clientX - rect.left;
    const percent = x / rect.width;
    const time = new Date(bufferStart.getTime() + totalDuration * percent);

    onSeek(time);
  };

  // Calendar rendering
  const renderCalendar = () => {
    const year = calendarDate.getFullYear();
    const month = calendarDate.getMonth();
    const monthNames = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];

    const firstDay = new Date(year, month, 1).getDay();
    const daysInMonth = new Date(year, month + 1, 0).getDate();
    const prevMonthDays = new Date(year, month, 0).getDate();

    const days: JSX.Element[] = [];

    // Previous month days
    for (let i = firstDay - 1; i >= 0; i--) {
      const day = prevMonthDays - i;
      days.push(
        <div key={`prev-${day}`} className="text-center py-2 text-gray-600 cursor-pointer hover:bg-gray-700 rounded">
          {day}
        </div>
      );
    }

    // Current month days
    const today = new Date();
    for (let day = 1; day <= daysInMonth; day++) {
      const date = new Date(year, month, day);
      const isToday = date.toDateString() === today.toDateString();
      const isSelected = date.toDateString() === selectedDate.toDateString();

      days.push(
        <div
          key={`current-${day}`}
          onClick={() => {
            const newDate = new Date(selectedDate);
            newDate.setFullYear(year);
            newDate.setMonth(month);
            newDate.setDate(day);
            setSelectedDate(newDate);
          }}
          className={cn(
            "text-center py-2 cursor-pointer rounded transition-colors",
            isSelected && "bg-blue-600 text-white",
            !isSelected && isToday && "border border-blue-500",
            !isSelected && !isToday && "hover:bg-gray-700"
          )}
        >
          {day}
        </div>
      );
    }

    // Next month days
    const remainingCells = 42 - days.length;
    for (let day = 1; day <= remainingCells; day++) {
      days.push(
        <div key={`next-${day}`} className="text-center py-2 text-gray-600 cursor-pointer hover:bg-gray-700 rounded">
          {day}
        </div>
      );
    }

    return (
      <div className="space-y-4">
        {/* Month/Year Navigation */}
        <div className="flex justify-between items-center">
          <div className="text-sm font-medium">
            {monthNames[month]} {year}
          </div>
          <div className="flex gap-2">
            <button
              onClick={() => setCalendarDate(new Date(year, month - 1, 1))}
              className="px-3 py-1 bg-gray-700 hover:bg-gray-600 rounded text-sm"
            >
              &lt;
            </button>
            <button
              onClick={() => {
                const now = new Date();
                setCalendarDate(now);
                setSelectedDate(now);
              }}
              className="px-3 py-1 bg-gray-700 hover:bg-gray-600 rounded text-sm"
            >
              Today
            </button>
            <button
              onClick={() => setCalendarDate(new Date(year, month + 1, 1))}
              className="px-3 py-1 bg-gray-700 hover:bg-gray-600 rounded text-sm"
            >
              &gt;
            </button>
          </div>
        </div>

        {/* Calendar Grid */}
        <div className="grid grid-cols-7 gap-1">
          {['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'].map(day => (
            <div key={day} className="text-center text-xs text-gray-400 py-1">
              {day}
            </div>
          ))}
          {days}
        </div>

        {/* Time Input */}
        <div className="flex gap-2 items-center">
          <span className="text-sm text-gray-400">Time:</span>
          <input
            type="time"
            step="1"
            value={`${selectedDate.getHours().toString().padStart(2, '0')}:${selectedDate.getMinutes().toString().padStart(2, '0')}:${selectedDate.getSeconds().toString().padStart(2, '0')}`}
            onChange={(e) => {
              const [hours, minutes, seconds] = e.target.value.split(':').map(Number);
              const newDate = new Date(selectedDate);
              newDate.setHours(hours || 0);
              newDate.setMinutes(minutes || 0);
              newDate.setSeconds(seconds || 0);
              setSelectedDate(newDate);
            }}
            className="flex-1 px-3 py-2 bg-gray-700 border border-gray-600 rounded text-sm"
          />
        </div>

        {/* Actions */}
        <div className="flex gap-2 justify-end">
          <button
            onClick={() => setShowDatePicker(false)}
            className="px-4 py-2 bg-gray-700 hover:bg-gray-600 rounded text-sm"
          >
            Cancel
          </button>
          <button
            onClick={() => {
              onSeek(selectedDate);
              setShowDatePicker(false);
            }}
            className="px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded text-sm"
          >
            Go to time
          </button>
        </div>
      </div>
    );
  };

  return (
    <>
      <div className={cn('bg-black/95 backdrop-blur-sm rounded-lg overflow-visible', className)}>
        {isExpanded ? (
          /* EXPANDED STATE */
          <div className="p-4 space-y-3">
            {/* Top Controls */}
            <div className="flex items-center justify-center gap-3">
              {/* Play/Pause */}
              <button
                onClick={onPlayPause}
                className="p-3 hover:bg-white/10 rounded-lg transition-colors text-white"
                title={isPlaying ? 'Pause' : 'Play'}
              >
                {isPlaying ? <Pause className="w-5 h-5" /> : <Play className="w-5 h-5" />}
              </button>

              {/* Scroll Buttons */}
              <button
                onClick={() => onScrollTimeline('left')}
                className="px-3 py-2 bg-white/10 hover:bg-white/20 rounded text-sm text-white transition-colors"
              >
                ◀
              </button>
              <button
                onClick={() => onScrollTimeline('right')}
                className="px-3 py-2 bg-white/10 hover:bg-white/20 rounded text-sm text-white transition-colors"
              >
                ▶
              </button>
            </div>

            {/* Scrolling Timeline */}
            <div className="relative">
              {/* Fixed center time indicator */}
              <div className="absolute left-1/2 -translate-x-1/2 top-0 z-30 pointer-events-none">
                <div className="bg-white text-black px-3 py-1.5 rounded shadow-lg text-sm font-mono flex items-center gap-2">
                  <span className="w-4 h-4 bg-blue-600 rounded flex items-center justify-center text-white text-xs">▶</span>
                  <span>{formatTime(currentTime)}, {formatDate(currentTime)}</span>
                </div>
              </div>

              {/* Timeline Container */}
              <div
                ref={timelineRef}
                className="relative mt-14 h-20 bg-gray-900 rounded-lg overflow-hidden cursor-pointer"
                onClick={handleTimelineClick}
              >
                {/* Fixed center line */}
                <div className="absolute left-1/2 top-0 bottom-0 w-0.5 bg-white z-20 pointer-events-none" />

                {/* Scrollable content */}
                <div
                  ref={contentWrapperRef}
                  className="absolute inset-0"
                  style={{
                    transform: `translateX(${timelineOffset}px)`,
                    transition: 'transform 0.05s linear'
                  }}
                >
                  {/* Time labels top */}
                  <div className="absolute inset-x-0 top-1">
                    {timeLabels.map((label, idx) => {
                      const pos = getPositionPercent(label);
                      return (
                        <div
                          key={idx}
                          className="absolute text-xs text-gray-400 -translate-x-1/2"
                          style={{ left: `${pos}%` }}
                        >
                          {formatTimeLabel(label)}
                        </div>
                      );
                    })}
                  </div>

                  {/* Tick marks */}
                  <div className="absolute inset-x-0 top-6 h-4">
                    {renderTicks()}
                  </div>

                  {/* Recording track */}
                  <div className="absolute inset-x-0 bottom-6 h-8 px-2">
                    <div className="relative w-full h-full bg-gray-800 rounded">
                      {/* Future zone */}
                      {renderFutureZone()}

                      {/* Recording sequences */}
                      {sequences.map((seq) => {
                        const startPercent = getPositionPercent(seq.startTime);
                        const endPercent = getPositionPercent(seq.endTime);
                        const width = endPercent - startPercent;

                        if (endPercent < 0 || startPercent > 100) return null;

                        return (
                          <div
                            key={seq.sequenceId}
                            className="absolute h-full bg-orange-600 rounded"
                            style={{
                              left: `${Math.max(0, startPercent)}%`,
                              width: `${Math.min(100 - startPercent, width)}%`,
                            }}
                          />
                        );
                      })}
                    </div>
                  </div>

                  {/* Time labels bottom */}
                  <div className="absolute inset-x-0 bottom-0">
                    {timeLabels.map((label, idx) => {
                      const pos = getPositionPercent(label);
                      return (
                        <div
                          key={`bottom-${idx}`}
                          className="absolute text-[10px] text-gray-500 -translate-x-1/2"
                          style={{ left: `${pos}%` }}
                        >
                          {formatTimeLabel(label)}
                        </div>
                      );
                    })}
                  </div>
                </div>
              </div>
            </div>

            {/* Bottom Controls */}
            <div className="flex items-center justify-between border-t border-gray-800 pt-3">
              {/* Left: Date/Time Picker */}
              <button
                onClick={() => {
                  setSelectedDate(currentTime);
                  setCalendarDate(currentTime);
                  setShowDatePicker(true);
                }}
                className="text-sm text-gray-400 hover:text-white transition-colors"
              >
                📅 {formatTime(currentTime)}, {formatDate(currentTime)}
              </button>

              {/* Center: Mode Toggle */}
              <div className="flex items-center gap-2">
                <button className="px-4 py-1.5 bg-blue-600 rounded-full text-xs font-medium text-white">
                  PLAYBACK
                </button>
                <button
                  onClick={() => setIsExpanded(false)}
                  className="p-1.5 hover:bg-white/10 rounded text-gray-400"
                >
                  <ChevronDown className="w-4 h-4" />
                </button>
              </div>

              {/* Right: Zoom and Speed */}
              <div className="flex items-center gap-2">
                {/* Speed */}
                <div className="relative">
                  <button
                    onClick={() => setShowSpeedMenu(!showSpeedMenu)}
                    className="px-3 py-1.5 bg-white/10 hover:bg-white/20 rounded text-sm text-white transition-colors"
                  >
                    {currentSpeed}x ▾
                  </button>
                  {showSpeedMenu && (
                    <div className="absolute bottom-full mb-1 right-0 bg-gray-800 rounded shadow-xl overflow-hidden min-w-[80px]">
                      {SPEED_OPTIONS.map(speed => (
                        <button
                          key={speed}
                          onClick={() => handleSpeedChange(speed)}
                          className={cn(
                            "w-full px-3 py-2 text-sm text-left hover:bg-white/10 transition-colors",
                            currentSpeed === speed ? "bg-blue-600 text-white" : "text-gray-300"
                          )}
                        >
                          {speed}x
                        </button>
                      ))}
                    </div>
                  )}
                </div>

                {/* Zoom */}
                <div className="relative">
                  <button
                    onClick={() => setShowZoomMenu(!showZoomMenu)}
                    className="px-3 py-1.5 bg-white/10 hover:bg-white/20 rounded text-sm text-white transition-colors"
                  >
                    {zoomConfig.label} ▾
                  </button>
                  {showZoomMenu && (
                    <div className="absolute bottom-full mb-1 right-0 bg-gray-800 rounded shadow-xl overflow-hidden min-w-[100px]">
                      {ZOOM_LEVELS.map(zoom => (
                        <button
                          key={zoom.hours}
                          onClick={() => {
                            onZoomChange(zoom.hours * 60);
                            setShowZoomMenu(false);
                          }}
                          className={cn(
                            "w-full px-3 py-2 text-sm text-left hover:bg-white/10 transition-colors",
                            Math.abs(zoomLevel - zoom.hours * 60) < 0.1 ? "bg-blue-600 text-white" : "text-gray-300"
                          )}
                        >
                          {zoom.label}
                        </button>
                      ))}
                    </div>
                  )}
                </div>
              </div>
            </div>
          </div>
        ) : (
          /* COLLAPSED STATE */
          <div className="px-4 py-2 flex items-center justify-between">
            {/* Left: Status */}
            <div className="flex items-center gap-2 text-sm">
              <span className="px-2 py-0.5 bg-blue-600 text-white rounded text-xs font-medium">
                PLAYBACK
              </span>
              <span className="text-white font-mono text-sm">
                {formatTime(currentTime)}
              </span>
              <span className="text-gray-400">•</span>
              <span className={cn('text-xs', hasRecording ? 'text-green-400' : 'text-red-400')}>
                {hasRecording ? '✓ Recording' : '✗ No Recording'}
              </span>
            </div>

            {/* Right: Expand Button */}
            <button
              onClick={() => setIsExpanded(true)}
              className="p-1.5 hover:bg-white/10 rounded text-gray-400"
            >
              <ChevronUp className="w-4 h-4" />
            </button>
          </div>
        )}
      </div>

      {/* Date/Time Picker Modal */}
      {showDatePicker && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-gray-900 border border-gray-700 rounded-lg p-6 min-w-[320px] max-w-md">
            <h3 className="text-lg font-medium mb-4 text-white">Select date and time</h3>
            {renderCalendar()}
          </div>
        </div>
      )}

      {/* Click outside to close menus */}
      {(showZoomMenu || showSpeedMenu || showDatePicker) && (
        <div
          className="fixed inset-0 z-40"
          onClick={() => {
            setShowZoomMenu(false);
            setShowSpeedMenu(false);
            setShowDatePicker(false);
          }}
        />
      )}
    </>
  );
}
