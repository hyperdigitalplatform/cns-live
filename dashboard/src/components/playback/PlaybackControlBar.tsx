import React, { useState, useRef, useEffect, useCallback, useMemo } from 'react';
import { Play, Pause, ChevronUp, ChevronDown, SkipBack, SkipForward } from 'lucide-react';

interface RecordingSequence {
  sequenceId: string;
  startTime: string;
  endTime: string;
  durationSeconds: number;
}

interface PlaybackControlBarProps {
  startTime?: Date;
  endTime?: Date;
  currentTime?: Date;
  sequences?: RecordingSequence[];
  isPlaying?: boolean;
  zoomLevel?: number;
  mode?: 'live' | 'playback'; // Current mode
  onPlayPause?: () => void;
  onSeek?: (time: Date) => void;
  onScrollTimeline?: (direction: 'left' | 'right') => void;
  onZoomChange?: (zoom: number) => void;
  onModeChange?: (mode: 'live' | 'playback') => void; // Mode switcher callback
  hasRecording?: boolean;
  onSpeedChange?: (speed: number) => void;
  className?: string;
}

// Zoom levels matching test-webrtc-playback.html exactly
const ZOOM_LEVELS = [
  { hours: 1/60,  label: '1 min',  interval: 10 * 1000,        minorInterval: 2 * 1000 },
  { hours: 5/60,  label: '5 min',  interval: 60 * 1000,        minorInterval: 10 * 1000 },
  { hours: 10/60, label: '10 min', interval: 1 * 60 * 1000,    minorInterval: 15 * 1000 },
  { hours: 0.5,   label: '30 min', interval: 5 * 60 * 1000,    minorInterval: 1 * 60 * 1000 },
  { hours: 1,     label: '1 hr',   interval: 5 * 60 * 1000,    minorInterval: 1 * 60 * 1000 },
  { hours: 2,     label: '2 hr',   interval: 10 * 60 * 1000,   minorInterval: 2 * 60 * 1000 },
  { hours: 8,     label: '8 hr',   interval: 60 * 60 * 1000,   minorInterval: 15 * 60 * 1000 },
  { hours: 16,    label: '16 hr',  interval: 2 * 60 * 60 * 1000, minorInterval: 30 * 60 * 1000 },
  { hours: 24,    label: '1 d',    interval: 4 * 60 * 60 * 1000, minorInterval: 1 * 60 * 60 * 1000 },
  { hours: 168,   label: '1 wk',   interval: 24 * 60 * 60 * 1000, minorInterval: 6 * 60 * 60 * 1000 }
];

const SPEED_OPTIONS = [0.25, 0.5, 1, 2, 4, 8, 16];

export function PlaybackControlBar({
  startTime = new Date(),
  endTime = new Date(),
  currentTime = new Date(),
  sequences = [],
  isPlaying = false,
  zoomLevel = 6,
  mode,
  onPlayPause,
  onSeek,
  onScrollTimeline,
  onZoomChange,
  onModeChange,
  hasRecording = false,
  onSpeedChange,
  className = ''
}: PlaybackControlBarProps) {
  const [isExpanded, setIsExpanded] = useState(true);
  const [showSpeedMenu, setShowSpeedMenu] = useState(false);
  const [currentSpeed, setCurrentSpeed] = useState(1);
  const [showDatePicker, setShowDatePicker] = useState(false);
  const [selectedDate, setSelectedDate] = useState(currentTime);
  const [calendarDate, setCalendarDate] = useState(currentTime);

  const timelineViewportRef = useRef<HTMLDivElement>(null); // Viewport (1x zoom width)
  const timelineContentRef = useRef<HTMLDivElement>(null);   // Content (3x zoom width)
  const timelineTrackRef = useRef<HTMLDivElement>(null);

  // Map zoomLevel (hours * 60) to zoom index
  const zoomIndex = useMemo(() => {
    const hours = zoomLevel / 60;
    const idx = ZOOM_LEVELS.findIndex(z => Math.abs(z.hours - hours) < 0.01);
    return idx >= 0 ? idx : 3; // default to 30 min
  }, [zoomLevel]);

  const currentZoom = ZOOM_LEVELS[zoomIndex];
  const zoomMs = currentZoom.hours * 60 * 60 * 1000;

  // === DOUBLE-BUFFER SYSTEM FOR SEAMLESS TIMELINE ===
  // Maintains two buffers (A and B) to eliminate visible gaps during scrubbing
  // When currentTime approaches edge, we pre-load next buffer then swap instantly

  const [centerTime, setCenterTime] = useState(currentTime);
  const [nextCenterTime, setNextCenterTime] = useState<Date | null>(null);
  const [activeBuffer, setActiveBuffer] = useState<'A' | 'B'>('A');
  const [isPreloading, setIsPreloading] = useState(false);

  // CRITICAL FIX: Use ref to track effective center time synchronously
  // This solves React state batching issue where setCenterTime doesn't update immediately
  const effectiveCenterTimeRef = useRef(centerTime);

  // Sync centerTime when currentTime changes dramatically (e.g., switching to playback mode)
  // This prevents the 2-5 second delay showing old timeline position
  useEffect(() => {
    const timeDiff = Math.abs(currentTime.getTime() - centerTime.getTime());
    if (timeDiff > 5000) {
      // Large jump detected - immediately update centerTime to match
      console.log('🔄 Large time jump detected - syncing centerTime:', currentTime.toISOString());
      setCenterTime(currentTime);
    }
  }, [currentTime, centerTime]);

  // Check if currentTime has jumped significantly (manual navigation)
  const timeDiff = Math.abs(currentTime.getTime() - centerTime.getTime());
  const isNavigating = timeDiff > 5000; // More than 5 seconds = manual navigation

  // During navigation: use currentTime immediately (synchronous, no batching lag)
  // During playback: use centerTime (follows smooth buffer updates)
  const effectiveCenterTime = isNavigating ? currentTime : centerTime;

  // Always keep ref in sync with effectiveCenterTime for use in effects
  effectiveCenterTimeRef.current = effectiveCenterTime;

  // Buffer A boundaries - calculate directly without memoization to ensure fresh values
  // This fixes the 30-minute time misalignment issue
  const bufferAStart = new Date(effectiveCenterTime.getTime() - zoomMs * 1.5);
  const bufferAEnd = new Date(effectiveCenterTime.getTime() + zoomMs * 1.5);

  // Buffer B boundaries (only calculated when pre-loading)
  const bufferBStart = useMemo(() => {
    if (!nextCenterTime) return bufferAStart;
    return new Date(nextCenterTime.getTime() - zoomMs * 1.5);
  }, [nextCenterTime, zoomMs, bufferAStart]);

  const bufferBEnd = useMemo(() => {
    if (!nextCenterTime) return bufferAEnd;
    return new Date(nextCenterTime.getTime() + zoomMs * 1.5);
  }, [nextCenterTime, zoomMs, bufferAEnd]);

  // Active buffer (what's currently displayed)
  const bufferStart = activeBuffer === 'A' ? bufferAStart : bufferBStart;
  const bufferEnd = activeBuffer === 'A' ? bufferAEnd : bufferBEnd;
  const totalDuration = bufferEnd.getTime() - bufferStart.getTime();

  // Double-buffer management: Pre-load and swap seamlessly
  useEffect(() => {
    // Calculate time difference between currentTime and effectiveCenterTime (from ref)
    const timeDiff = Math.abs(currentTime.getTime() - effectiveCenterTimeRef.current.getTime());
    const largeJump = timeDiff > 5000; // More than 5 seconds = manual navigation

    if (largeJump) {
      // Manual navigation (backward/forward buttons) - reload immediately
      console.log('⚡ Manual navigation detected - reloading buffer centered on:', currentTime.toISOString());
      effectiveCenterTimeRef.current = currentTime; // Update ref synchronously
      setCenterTime(currentTime); // Update state (for next render)
      setNextCenterTime(null);
      setIsPreloading(false);
      return;
    }

    // Start pre-loading when 30% from edge (for smooth playback)
    const preloadThreshold = zoomMs * 0.3;
    const marginStart = new Date(bufferStart.getTime() + preloadThreshold);
    const marginEnd = new Date(bufferEnd.getTime() - preloadThreshold);

    const approachingLeft = currentTime < marginStart;
    const approachingRight = currentTime > marginEnd;

    if ((approachingLeft || approachingRight) && !isPreloading) {
      // Start pre-loading next buffer centered on current time
      console.log('🔄 Pre-loading next buffer:', {
        activeBuffer,
        direction: approachingLeft ? 'LEFT' : 'RIGHT',
        currentTime: currentTime.toISOString()
      });

      setIsPreloading(true);
      setNextCenterTime(currentTime);

      // Allow React to calculate next buffer, then swap
      requestAnimationFrame(() => {
        setTimeout(() => {
          console.log('✅ Buffer swap complete - seamless transition');
          const nextBuffer = activeBuffer === 'A' ? 'B' : 'A';
          setActiveBuffer(nextBuffer);
          effectiveCenterTimeRef.current = currentTime; // Update ref synchronously
          setCenterTime(currentTime); // Update state
          setNextCenterTime(null);
          setIsPreloading(false);
        }, 16); // One frame (16ms at 60fps)
      });
    }
  }, [currentTime, bufferStart, bufferEnd, centerTime, zoomMs, activeBuffer, isPreloading]);

  // Cleanup effect: Clear buffers on unmount to prevent memory leaks
  useEffect(() => {
    return () => {
      // Clear any pending state updates
      setNextCenterTime(null);
      setIsPreloading(false);
      console.log('🧹 Timeline buffers cleaned up');
    };
  }, []); // Run only on unmount

  // Calculate visible 1x range (effectiveCenterTime ± 0.5x zoom)
  const visibleStart = new Date(effectiveCenterTime.getTime() - zoomMs / 2);
  const visibleEnd = new Date(effectiveCenterTime.getTime() + zoomMs / 2);

  // Generate time labels for FULL 3x buffer (absolute positioned, like ticks)
  // Calculate directly without useMemo to ensure fresh values during navigation
  const timeLabels = (() => {
    const labels: { time: Date; position: number }[] = [];
    const interval = currentZoom.interval;
    const firstLabelTime = new Date(Math.ceil(bufferStart.getTime() / interval) * interval);

    let time = firstLabelTime;
    while (time <= bufferEnd) {
      if (time >= bufferStart) {
        // Calculate position as percentage across the full 3x buffer
        const position = ((time.getTime() - bufferStart.getTime()) / totalDuration) * 100;
        labels.push({
          time: new Date(time),
          position
        });
      }
      time = new Date(time.getTime() + interval);
    }

    return labels;
  })();

  // Generate tick marks for FULL 3x buffer (absolute positioned)
  // Calculate directly without useMemo to ensure fresh values during navigation
  const ticks = (() => {
    const result: { position: number; type: 'major' | 'minor' }[] = [];

    // Minor ticks across full 3x buffer
    const minorInterval = currentZoom.minorInterval;
    let time = new Date(Math.ceil(bufferStart.getTime() / minorInterval) * minorInterval);
    while (time <= bufferEnd) {
      const position = ((time.getTime() - bufferStart.getTime()) / totalDuration) * 100;
      if (position >= 0 && position <= 100) {
        result.push({ position, type: 'minor' });
      }
      time = new Date(time.getTime() + minorInterval);
    }

    // Major ticks across full 3x buffer
    const majorInterval = currentZoom.interval;
    time = new Date(Math.ceil(bufferStart.getTime() / majorInterval) * majorInterval);
    while (time <= bufferEnd) {
      const position = ((time.getTime() - bufferStart.getTime()) / totalDuration) * 100;
      if (position >= 0 && position <= 100) {
        result.push({ position, type: 'major' });
      }
      time = new Date(time.getTime() + majorInterval);
    }

    return result;
  })();

  // Format time label based on zoom level
  const formatTimeLabel = (date: Date): string => {
    const hours = date.getHours();
    const minutes = date.getMinutes();
    const seconds = date.getSeconds();
    const ampm = hours >= 12 ? 'PM' : 'AM';
    const displayHours = hours % 12 || 12;

    if (currentZoom.hours < 0.1) {
      // Very short zoom - show seconds
      return `${displayHours}:${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`;
    } else if (currentZoom.hours >= 168) {
      // 1 week - show month and day
      const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
      return `${months[date.getMonth()]} ${date.getDate()}`;
    } else {
      // Normal - show hours:minutes
      return `${displayHours}:${minutes.toString().padStart(2, '0')} ${ampm}`;
    }
  };

  const formatCurrentTime = (date: Date): string => {
    const hours = date.getHours();
    const minutes = date.getMinutes();
    const seconds = date.getSeconds();
    const ampm = hours >= 12 ? 'PM' : 'AM';
    const displayHours = hours % 12 || 12;
    const ms = date.getMilliseconds();
    return `${displayHours}:${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}.${ms.toString().padStart(3, '0')} ${ampm}`;
  };

  const formatDate = (date: Date): string => {
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day = String(date.getDate()).padStart(2, '0');
    return `${year}-${month}-${day}`;
  };

  // Calculate scroll offset for CSS transform
  // Don't use useMemo - let it recalculate on every render to ensure fresh dimensions
  const scrollOffset = (() => {
    if (!timelineViewportRef.current || !timelineContentRef.current) return 0;

    // CRITICAL: Use viewport width (1x zoom), not content width (3x zoom)!
    const viewportWidth = timelineViewportRef.current.offsetWidth;
    const contentWidth = timelineContentRef.current.offsetWidth; // This is 3x viewportWidth

    // Calculate where currentTime is positioned on the timeline content (3x wide)
    const elapsed = currentTime.getTime() - bufferStart.getTime();
    const percentComplete = elapsed / totalDuration; // 0 to 1 across the 3x buffer
    const targetPixelPosition = percentComplete * contentWidth; // Position on the 3x wide content

    // We want to center currentTime on the viewport (1x wide)
    const centerOfViewport = viewportWidth / 2;

    // Calculate offset: how much to shift the content left/right
    // Negative offset = scroll content left (shows later times)
    // Positive offset = scroll content right (shows earlier times)
    const offset = centerOfViewport - targetPixelPosition;

    // CLAMP: Prevent scrolling beyond buffer edges
    // When percentComplete = 0% (at bufferStart), offset should be max positive
    // When percentComplete = 100% (at bufferEnd), offset should be max negative
    const maxOffset = centerOfViewport; // Can scroll right up to center of viewport
    const minOffset = centerOfViewport - contentWidth; // Can scroll left until end reaches center
    const clampedOffset = Math.max(minOffset, Math.min(maxOffset, offset));

    // Debug: log scroll calculation details
    console.log('📏 SCROLL CALCULATION:', {
      currentTime: currentTime.toISOString(),
      effectiveCenterTime: effectiveCenterTime.toISOString(),
      centerTimeState: centerTime.toISOString(),
      bufferStart: bufferStart.toISOString(),
      bufferEnd: bufferEnd.toISOString(),
      elapsed: (elapsed / 1000).toFixed(1) + 's',
      totalDuration: (totalDuration / 1000).toFixed(1) + 's',
      percentComplete: (percentComplete * 100).toFixed(1) + '%',
      viewportWidth: viewportWidth + 'px',
      contentWidth: contentWidth + 'px',
      targetPixelPosition: targetPixelPosition.toFixed(2) + 'px',
      centerOfViewport: centerOfViewport.toFixed(2) + 'px',
      offset: offset.toFixed(2) + 'px',
      clampedOffset: clampedOffset.toFixed(2) + 'px'
    });

    // Debug: log time alignment (use effectiveCenterTime from ref for accurate values)
    const timeDiff = Math.abs(currentTime.getTime() - effectiveCenterTime.getTime());
    if (timeDiff > 1000) {
      console.log('⚠️ TIME MISALIGNMENT:', {
        currentTime: currentTime.toISOString(),
        effectiveCenterTime: effectiveCenterTime.toISOString(),
        centerTimeState: centerTime.toISOString(),
        bufferStart: bufferStart.toISOString(),
        bufferEnd: bufferEnd.toISOString(),
        timeDiff: (timeDiff / 1000).toFixed(1) + 's'
      });
    }

    // Debug: log if clamping occurred
    if (clampedOffset !== offset) {
      console.log('⚠️ CLAMPED scroll offset:', {
        calculated: offset.toFixed(2),
        clamped: clampedOffset.toFixed(2),
        percentComplete: (percentComplete * 100).toFixed(1) + '%',
        reason: offset > maxOffset ? 'TOO FAR RIGHT' : 'TOO FAR LEFT'
      });
    }

    return clampedOffset;
  })();

  // Apply scroll transform when currentTime changes
  useEffect(() => {
    if (timelineContentRef.current) {
      // Use faster transition during playback, smoother for manual seeks
      const transition = isPlaying ? 'transform 0.05s linear' : 'transform 0.3s cubic-bezier(0.4, 0, 0.2, 1)';
      timelineContentRef.current.style.transition = transition;
      timelineContentRef.current.style.transform = `translateX(${scrollOffset}px)`;

      // Debug log during playback
      if (isPlaying) {
        console.log('📊 Transform applied:', scrollOffset.toFixed(2), 'px | Time:', currentTime.toLocaleTimeString());
      }
    }
  }, [scrollOffset, isPlaying, currentTime]);

  // Draw sequence bars
  // Calculate directly without useMemo to ensure fresh values during navigation
  const sequenceBars = sequences.map((seq) => {
    const seqStart = new Date(seq.startTime).getTime();
    const seqEnd = new Date(seq.endTime).getTime();

    // Check if in buffer range
    if (seqEnd < bufferStart.getTime() || seqStart > bufferEnd.getTime()) {
      return null;
    }

    const startPercent = Math.max(0, ((seqStart - bufferStart.getTime()) / totalDuration) * 100);
    const endPercent = Math.min(100, ((seqEnd - bufferStart.getTime()) / totalDuration) * 100);

    return {
      id: seq.sequenceId,
      left: startPercent,
      width: endPercent - startPercent
    };
  }).filter(Boolean);

  // Draw future zone (from current system time to end of buffer)
  // This shows the area representing recordings that don't exist yet (future time)
  // Calculate directly without memoization to match buffer calculation changes
  const futureZone = (() => {
    const now = new Date(); // Current system time (not playback time!)

    // If current system time is before the buffer, no future zone
    if (now < bufferStart) return null;

    // If current system time is after the buffer, entire buffer is past (no future)
    if (now > bufferEnd) return null;

    // Calculate where "now" appears on the timeline (as percentage)
    const futureStart = Math.max(0, ((now.getTime() - bufferStart.getTime()) / totalDuration) * 100);
    const futureEnd = 100; // Extends to end of visible buffer

    return {
      left: futureStart,
      width: futureEnd - futureStart
    };
  })();

  const handleTimelineClick = (e: React.MouseEvent<HTMLDivElement>) => {
    if (!timelineTrackRef.current) return;

    const rect = timelineTrackRef.current.getBoundingClientRect();
    const clickX = e.clientX - rect.left;
    const percent = clickX / rect.width;
    const clickedTime = new Date(bufferStart.getTime() + percent * totalDuration);

    // Update centerTime immediately for timeline clicks
    setCenterTime(clickedTime);
    onSeek?.(clickedTime);
  };

  const handleSpeedChange = (speed: number) => {
    setCurrentSpeed(speed);
    setShowSpeedMenu(false);
    if (onSpeedChange) {
      onSpeedChange(speed);
    }
  };

  const handleZoomChange = (newIndex: number) => {
    if (newIndex >= 0 && newIndex < ZOOM_LEVELS.length) {
      const newZoom = ZOOM_LEVELS[newIndex];
      onZoomChange?.(newZoom.hours * 60); // Convert hours to minutes
    }
  };

  // Calendar rendering
  const renderCalendar = () => {
    const year = calendarDate.getFullYear();
    const month = calendarDate.getMonth();
    const firstDay = new Date(year, month, 1).getDay();
    const daysInMonth = new Date(year, month + 1, 0).getDate();
    const today = new Date();

    const days: JSX.Element[] = [];

    // Previous month days
    const prevMonthDays = new Date(year, month, 0).getDate();
    for (let i = firstDay - 1; i >= 0; i--) {
      days.push(
        <div key={`prev-${i}`} className="text-center py-2 text-gray-600 text-sm">
          {prevMonthDays - i}
        </div>
      );
    }

    // Current month days
    for (let day = 1; day <= daysInMonth; day++) {
      const date = new Date(year, month, day);
      const isSelected = selectedDate.getDate() === day &&
                        selectedDate.getMonth() === month &&
                        selectedDate.getFullYear() === year;
      const isToday = today.getDate() === day &&
                     today.getMonth() === month &&
                     today.getFullYear() === year;

      days.push(
        <div
          key={`current-${day}`}
          onClick={() => setSelectedDate(new Date(year, month, day, selectedDate.getHours(), selectedDate.getMinutes(), selectedDate.getSeconds()))}
          className={`text-center py-2 text-sm cursor-pointer rounded ${
            isSelected ? 'bg-blue-600 text-white' : isToday ? 'border border-blue-600' : 'hover:bg-gray-700'
          }`}
        >
          {day}
        </div>
      );
    }

    // Next month days
    const remainingDays = 42 - days.length;
    for (let day = 1; day <= remainingDays; day++) {
      days.push(
        <div key={`next-${day}`} className="text-center py-2 text-gray-600 text-sm">
          {day}
        </div>
      );
    }

    return (
      <div>
        <div className="flex items-center justify-between mb-3">
          <div className="text-white font-medium">
            {calendarDate.toLocaleDateString('en-US', { month: 'long', year: 'numeric' })}
          </div>
          <div className="flex gap-2">
            <button
              onClick={() => setCalendarDate(new Date(calendarDate.getFullYear(), calendarDate.getMonth() - 1))}
              className="px-2 py-1 bg-gray-700 hover:bg-gray-600 rounded text-sm"
            >
              &lt;
            </button>
            <button
              onClick={() => {
                const today = new Date();
                setCalendarDate(today);
                setSelectedDate(today);
              }}
              className="px-2 py-1 bg-gray-700 hover:bg-gray-600 rounded text-sm"
            >
              Today
            </button>
            <button
              onClick={() => setCalendarDate(new Date(calendarDate.getFullYear(), calendarDate.getMonth() + 1))}
              className="px-2 py-1 bg-gray-700 hover:bg-gray-600 rounded text-sm"
            >
              &gt;
            </button>
          </div>
        </div>

        <div className="grid grid-cols-7 gap-1 mb-2">
          {['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'].map(day => (
            <div key={day} className="text-center text-xs text-gray-400 font-medium py-1">
              {day}
            </div>
          ))}
        </div>

        <div className="grid grid-cols-7 gap-1 mb-4">
          {days}
        </div>

        <div className="mb-4">
          <label className="block text-xs text-gray-400 mb-1">Time:</label>
          <input
            type="time"
            step="1"
            value={`${selectedDate.getHours().toString().padStart(2, '0')}:${selectedDate.getMinutes().toString().padStart(2, '0')}:${selectedDate.getSeconds().toString().padStart(2, '0')}`}
            onChange={(e) => {
              const [hours, minutes, seconds] = e.target.value.split(':').map(Number);
              setSelectedDate(new Date(selectedDate.getFullYear(), selectedDate.getMonth(), selectedDate.getDate(), hours, minutes, seconds || 0));
            }}
            className="w-full bg-gray-800 border border-gray-700 rounded px-3 py-2 text-white"
          />
        </div>

        <div className="flex justify-end gap-2">
          <button
            onClick={() => setShowDatePicker(false)}
            className="px-4 py-2 bg-gray-700 hover:bg-gray-600 rounded text-sm"
          >
            Cancel
          </button>
          <button
            onClick={() => {
              onSeek?.(selectedDate);
              setShowDatePicker(false);
            }}
            className="px-4 py-2 bg-blue-600 hover:bg-blue-500 rounded text-sm"
          >
            Go to time
          </button>
        </div>
      </div>
    );
  };

  // Live mode - show only floating toggle (no bar background)
  if (mode === 'live') {
    return (
      <div className="absolute bottom-4 left-1/2 transform -translate-x-1/2 z-30">
        {/* Live/Playback Mode Toggle - Floating */}
        <div className="flex items-center gap-0.5 bg-gray-900/90 backdrop-blur-sm rounded-lg px-1.5 py-1 shadow-lg border border-gray-700/50">
          <button
            onClick={() => onModeChange?.('live')}
            className="px-3 py-1.5 text-xs font-medium rounded transition-colors bg-blue-600 text-white"
          >
            LIVE
          </button>
          <button
            onClick={() => onModeChange?.('playback')}
            className="px-3 py-1.5 text-xs font-medium rounded transition-colors text-gray-300 hover:text-white hover:bg-gray-700"
          >
            PLAYBACK
          </button>
        </div>
      </div>
    );
  }

  if (!isExpanded) {
    // Collapsed state - time on left, toggle centered
    return (
      <div className="bg-gradient-to-t from-black/95 to-black/80 px-5 py-2">
        <div className="flex items-center justify-between gap-4">
          {/* LEFT: Current time display */}
          <div
            className="flex items-center gap-1.5 cursor-pointer text-xs font-mono text-white hover:text-blue-400 transition-colors"
            onClick={() => setShowDatePicker(true)}
            title="Click to select date/time"
          >
            <span className="w-3 h-3 bg-blue-600 rounded flex items-center justify-center text-[8px]">◘</span>
            <span>{formatCurrentTime(currentTime)}, {formatDate(currentTime)}</span>
          </div>

          {/* CENTER: Mode Toggle + Expand Button */}
          <div className="flex items-center gap-2">
            <div className="flex items-center gap-0.5 bg-gray-800/50 rounded px-1 py-0.5">
              <button
                onClick={() => onModeChange?.('live')}
                className="px-3 py-1 text-xs font-medium rounded transition-colors text-gray-400 hover:text-white hover:bg-gray-700"
              >
                LIVE
              </button>
              <button
                onClick={() => onModeChange?.('playback')}
                className="px-3 py-1 text-xs font-medium rounded transition-colors bg-blue-600 text-white"
              >
                PLAYBACK
              </button>
            </div>

            {/* Expand button next to toggle */}
            <button
              onClick={() => setIsExpanded(true)}
              className="text-white/90 hover:text-white hover:bg-white/10 rounded transition-all w-7 h-7 flex items-center justify-center"
              title="Expand timeline"
            >
              <ChevronUp className="w-4 h-4" />
            </button>
          </div>

          {/* RIGHT: Empty spacer for balance */}
          <div className="w-[200px]"></div>
        </div>

        {/* Date/Time Picker Modal */}
        {showDatePicker && (
          <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50" onClick={() => setShowDatePicker(false)}>
            <div className="bg-gray-900 border border-gray-700 rounded-lg p-6 min-w-[320px] max-w-md" onClick={(e) => e.stopPropagation()}>
              <h3 className="text-lg font-medium mb-4 text-white">Select date and time</h3>
              {renderCalendar()}
            </div>
          </div>
        )}
      </div>
    );
  }

  // Expanded state
  return (
    <div className={`bg-gradient-to-t from-black/95 to-black/80 ${className}`}>
      {/* Playback Controls */}
      <div className="flex items-center justify-center gap-2 px-5 py-1 relative">
        {/* Playback buttons - CENTERED */}
        <div className="flex items-center gap-2">
          <button
            onClick={() => onScrollTimeline?.('left')}
            className="text-white/90 hover:text-white hover:bg-white/10 rounded transition-all w-8 h-8 flex items-center justify-center"
            title="Backward"
          >
            <SkipBack className="w-5 h-5" />
          </button>
          <button
            onClick={onPlayPause}
            className="text-white/90 hover:text-white hover:bg-white/10 rounded transition-all w-9 h-9 flex items-center justify-center"
            title="Play/Pause"
          >
            {isPlaying ? <Pause className="w-6 h-6" /> : <Play className="w-6 h-6 ml-0.5" />}
          </button>
          <button
            onClick={() => onScrollTimeline?.('right')}
            className="text-white/90 hover:text-white hover:bg-white/10 rounded transition-all w-8 h-8 flex items-center justify-center"
            title="Forward"
          >
            <SkipForward className="w-5 h-5" />
          </button>
        </div>

        {/* Speed selector - ABSOLUTE RIGHT */}
        <div className="absolute right-5">
          <button
            onClick={() => setShowSpeedMenu(!showSpeedMenu)}
            className="px-2 py-1 bg-gray-800/95 border border-gray-700 rounded text-[10px] text-white hover:bg-gray-700 flex items-center gap-1"
          >
            <span>{currentSpeed}x</span>
            <span className="text-[8px]">▾</span>
          </button>

          {showSpeedMenu && (
            <div className="absolute right-0 top-full mt-0.5 bg-gray-800/95 border border-gray-700 rounded overflow-hidden z-40">
              {SPEED_OPTIONS.map(speed => (
                <div
                  key={speed}
                  onClick={() => handleSpeedChange(speed)}
                  className={`px-3 py-1 cursor-pointer text-white text-[10px] hover:bg-white/10 ${
                    speed === currentSpeed ? 'bg-blue-600' : ''
                  }`}
                >
                  {speed}x
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Timeline Container */}
      <div className="px-5 pb-2 relative">
        {/* Fixed white line at center - starts below time indicator, ends at timeline track bottom */}
        <div className="absolute left-1/2 top-8 w-0.5 bg-white z-20 pointer-events-none" style={{ transform: 'translateX(-0.5px)', height: 'calc(100% - 2rem - 0.25rem)' }} />

        {/* Fixed time display at center */}
        <div
          className="absolute left-1/2 top-2 -translate-x-1/2 z-30 cursor-pointer pointer-events-auto"
          onClick={() => setShowDatePicker(true)}
          title="Click to select date/time"
        >
          <div className="bg-white text-black px-2 py-1 rounded shadow-lg text-xs font-mono flex items-center gap-1.5 border border-white/30">
            <span className="w-3 h-3 bg-blue-600 rounded flex items-center justify-center text-white text-[8px]">◘</span>
            <span>{formatCurrentTime(currentTime)}, {formatDate(currentTime)}</span>
          </div>
        </div>

        {/* Viewport wrapper - clips to visible area */}
        <div ref={timelineViewportRef} className="relative w-full overflow-hidden">
          {/* Scrollable timeline content - 3x wide! */}
          <div
            ref={timelineContentRef}
            className="relative"
            style={{
              width: '300%', // 3x buffer width
              willChange: 'transform'
            }}
          >
            {/* Timeline header with labels (absolute positioned across full 3x buffer) */}
            <div className="relative mb-0 mt-8 h-4">
              {timeLabels.map((label) => (
                <span
                  key={label.time.getTime()}
                  className="absolute whitespace-nowrap text-[10px] font-medium text-gray-400 -translate-x-1/2"
                  style={{ left: `${label.position}%` }}
                >
                  {formatTimeLabel(label.time)}
                </span>
              ))}
            </div>

            {/* Tick marks */}
            <div className="relative w-full h-3 mb-1 -mt-0.5">
            {ticks.map((tick, idx) => (
              <div
                key={idx}
                className={`absolute bottom-0 w-px transition-all duration-800 ${
                  tick.type === 'major'
                    ? 'h-3 bg-white/50'
                    : 'h-1.5 bg-white/20'
                }`}
                style={{ left: `${tick.position}%` }}
              />
            ))}
          </div>

          {/* Timeline track */}
          <div
            ref={timelineTrackRef}
            className="relative h-4 bg-white/5 rounded cursor-pointer mb-1"
            onClick={handleTimelineClick}
          >
            {/* Sequence bars */}
            {sequenceBars.map((bar: any) => {
              // Find the original sequence to get its start time for debugging
              const sequence = sequences.find(s => s.sequenceId === bar.id);
              const startTimeHint = sequence
                ? new Date(sequence.startTime).toLocaleString('en-US', {
                    month: 'short',
                    day: 'numeric',
                    hour: '2-digit',
                    minute: '2-digit',
                    second: '2-digit',
                    hour12: false,
                  })
                : '';

              return (
                <div
                  key={bar.id}
                  className="absolute top-0 h-full bg-orange-500 rounded hover:bg-orange-400 transition-colors cursor-help"
                  style={{ left: `${bar.left}%`, width: `${bar.width}%` }}
                  title={`Start: ${startTimeHint}`}
                />
              );
            })}

            {/* Future zone */}
            {futureZone && (
              <div
                className="absolute top-0 h-full bg-green-500/15 rounded-r"
                style={{ left: `${futureZone.left}%`, width: `${futureZone.width}%` }}
              />
            )}
          </div>
        </div>
        </div>
      </div>

      {/* Bottom Controls */}
      <div className="flex items-center justify-between px-5 pb-2 border-t border-gray-800 pt-2 relative">
        {/* LEFT: Empty spacer for balance */}
        <div className="w-[200px]"></div>

        {/* CENTER: Mode Toggle + Collapse Button */}
        <div className="absolute left-1/2 transform -translate-x-1/2 flex items-center gap-2">
          <div className="flex items-center gap-0.5 bg-gray-800/50 rounded px-1 py-0.5">
            <button
              onClick={() => onModeChange?.('live')}
              className="px-3 py-1 text-xs font-medium rounded transition-colors text-gray-400 hover:text-white hover:bg-gray-700"
            >
              LIVE
            </button>
            <button
              onClick={() => onModeChange?.('playback')}
              className="px-3 py-1 text-xs font-medium rounded transition-colors bg-blue-600 text-white"
            >
              PLAYBACK
            </button>
          </div>

          {/* Collapse button next to toggle */}
          <button
            onClick={() => setIsExpanded(false)}
            className="text-white/90 hover:text-white hover:bg-white/10 rounded transition-all w-7 h-7 flex items-center justify-center"
            title="Collapse timeline"
          >
            <ChevronDown className="w-4 h-4" />
          </button>
        </div>

        {/* RIGHT: Zoom Controls */}
        <div className="flex items-center gap-2">
          <span className="text-xs text-gray-400">{currentZoom.label}</span>
          <button
            onClick={() => handleZoomChange(zoomIndex - 1)}
            className="px-2 py-1 bg-gray-800 hover:bg-gray-700 rounded text-white text-xs disabled:opacity-50 disabled:cursor-not-allowed"
            disabled={zoomIndex === 0}
            title="Zoom out"
          >
            −
          </button>
          <input
            type="range"
            min="0"
            max={ZOOM_LEVELS.length - 1}
            value={zoomIndex}
            onChange={(e) => handleZoomChange(parseInt(e.target.value))}
            className="w-24 h-1"
            title={`Zoom: ${currentZoom.label}`}
          />
          <button
            onClick={() => handleZoomChange(zoomIndex + 1)}
            className="px-2 py-1 bg-gray-800 hover:bg-gray-700 rounded text-white text-xs disabled:opacity-50 disabled:cursor-not-allowed"
            disabled={zoomIndex === ZOOM_LEVELS.length - 1}
            title="Zoom in"
          >
            +
          </button>
        </div>
      </div>

      {/* Date/Time Picker Modal */}
      {showDatePicker && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50" onClick={() => setShowDatePicker(false)}>
          <div className="bg-gray-900 border border-gray-700 rounded-lg p-6 min-w-[320px] max-w-md" onClick={(e) => e.stopPropagation()}>
            <h3 className="text-lg font-medium mb-4 text-white">Select date and time</h3>
            {renderCalendar()}
          </div>
        </div>
      )}

      {/* Click outside to close menus */}
      {showSpeedMenu && (
        <div
          className="fixed inset-0 z-30"
          onClick={() => setShowSpeedMenu(false)}
        />
      )}
    </div>
  );
}
