import React, { useEffect, useRef, useState } from 'react';
import {
  Play,
  Pause,
  SkipBack,
  SkipForward,
  Maximize,
  Minimize,
} from 'lucide-react';
import { cn } from '@/utils/cn';
import { Button } from './ui/Dialog';
import { useWebRTCPlayback } from '@/hooks/useWebRTCPlayback';

interface RecordingPlayerProps {
  cameraId: string;
  startTime: Date;
  endTime: Date;
  initialPlaybackTime?: Date;
  onPlaybackTimeChange?: (time: Date) => void;
  onPlaybackStateChange?: (state: 'playing' | 'paused' | 'loading' | 'error') => void;
  className?: string;
  showControls?: boolean; // Option to hide built-in controls (when external controls are used)
  externalIsPlaying?: boolean; // External control for play/pause state
  externalCurrentTime?: Date; // External control for current playback time
}

export function RecordingPlayer({
  cameraId,
  startTime,
  endTime,
  initialPlaybackTime,
  onPlaybackTimeChange,
  onPlaybackStateChange,
  className,
  showControls: showBuiltInControls = true, // Default to showing controls
  externalIsPlaying,
  externalCurrentTime,
}: RecordingPlayerProps) {
  const containerRef = useRef<HTMLDivElement>(null);

  const [isPlaying, setIsPlaying] = useState(false);
  const [currentTime, setCurrentTime] = useState(initialPlaybackTime || startTime);
  const [isFullscreen, setIsFullscreen] = useState(false);
  const [showControls, setShowControls] = useState(true);

  // Track frame start time for RTP timestamp calculation
  const frameStartTimeRef = useRef<number | null>(null);
  // WebRTC playback time - only changes on manual seeks, NOT on frame updates
  const [webrtcPlaybackTime, setWebrtcPlaybackTime] = useState(initialPlaybackTime || startTime);

  // Use ref to track latest externalIsPlaying value in callbacks
  const externalIsPlayingRef = useRef(externalIsPlaying);

  // Track if this is the first connection (show spinner) or navigation (don't show spinner)
  const hasConnectedOnceRef = useRef(false);

  useEffect(() => {
    externalIsPlayingRef.current = externalIsPlaying;
  }, [externalIsPlaying]);

  // Use WebRTC playback hook
  const { videoRef, state, error, stop, isConnecting, isConnected } = useWebRTCPlayback({
    cameraId,
    playbackTime: webrtcPlaybackTime, // Use separate state for WebRTC time
    skipGaps: true,
    speed: 1.0,
    onStateChange: (webrtcState) => {
      if (webrtcState === 'connected') {
        // Mark that we've connected at least once (hide spinner on reconnections)
        hasConnectedOnceRef.current = true;

        // Set frame start time when connection is established
        frameStartTimeRef.current = currentTime.getTime();

        // Respect external play/pause state (from parent component)
        // Use ref to get LATEST value, not stale closure value
        // This prevents auto-play when user has paused (e.g., after forward/backward navigation)
        const shouldPlay = externalIsPlayingRef.current ?? false;

        console.log('🎬 WebRTC connected - shouldPlay:', shouldPlay, 'externalIsPlaying:', externalIsPlayingRef.current);

        if (shouldPlay) {
          const video = videoRef.current;
          if (video) {
            video.play().catch(err => console.error('Play failed:', err));
          }
          setIsPlaying(true);
          onPlaybackStateChange?.('playing');
        } else {
          // Connected but paused - show first frame without playing
          setIsPlaying(false);
          onPlaybackStateChange?.('paused');
        }
      } else if (webrtcState === 'failed') {
        onPlaybackStateChange?.('error');
        setIsPlaying(false);
        frameStartTimeRef.current = null;
      } else if (webrtcState === 'connecting') {
        // Only show loading on first connection, not on navigation
        if (!hasConnectedOnceRef.current) {
          onPlaybackStateChange?.('loading');
        }
      } else if (webrtcState === 'disconnected') {
        setIsPlaying(false);
        frameStartTimeRef.current = null;
      }
    },
  });

  // Play/Pause
  const togglePlayPause = async () => {
    const video = videoRef.current;
    if (!video) return;

    try {
      if (isPlaying) {
        video.pause();
        setIsPlaying(false);
        onPlaybackStateChange?.('paused');
      } else {
        await video.play();
        setIsPlaying(true);
        onPlaybackStateChange?.('playing');
      }
    } catch (err) {
      console.error('Playback error:', err);
    }
  };

  // Skip forward/backward
  // This triggers a new WebRTC connection at the new time
  const skip = (seconds: number) => {
    const newTime = new Date(currentTime.getTime() + seconds * 1000);

    // Clamp to valid range
    if (newTime < startTime) {
      setCurrentTime(startTime);
      onPlaybackTimeChange?.(startTime);
    } else if (newTime > endTime) {
      setCurrentTime(endTime);
      onPlaybackTimeChange?.(endTime);
    } else {
      setCurrentTime(newTime);
      onPlaybackTimeChange?.(newTime);
    }
  };

  // Fullscreen toggle
  const toggleFullscreen = () => {
    const container = containerRef.current;
    if (!container) return;

    if (!isFullscreen) {
      if (container.requestFullscreen) {
        container.requestFullscreen();
      }
      setIsFullscreen(true);
    } else {
      if (document.exitFullscreen) {
        document.exitFullscreen();
      }
      setIsFullscreen(false);
    }
  };

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

  // Sync external play/pause control
  useEffect(() => {
    if (externalIsPlaying === undefined) return;

    const video = videoRef.current;
    if (!video) return;

    const syncPlayPause = async () => {
      try {
        if (externalIsPlaying && !isPlaying) {
          await video.play();
          setIsPlaying(true);
        } else if (!externalIsPlaying && isPlaying) {
          video.pause();
          setIsPlaying(false);
        }
      } catch (err) {
        console.error('External playback control error:', err);
      }
    };

    syncPlayPause();
  }, [externalIsPlaying, isPlaying, videoRef]);

  // Sync external time control (only for MAJOR time jumps like button clicks)
  // Like test-webrtc-playback.html, frame updates don't cause reconnection
  useEffect(() => {
    if (externalCurrentTime === undefined) return;

    // Only trigger reconnection if there's a SIGNIFICANT time jump (> 5 seconds)
    // This means user clicked forward/backward button or timeline
    // Small differences are just frame updates echoed back - IGNORE THEM
    const timeDiff = Math.abs(externalCurrentTime.getTime() - currentTime.getTime());
    if (timeDiff < 5000) return; // Less than 5 seconds = ignore (frame updates)

    // This is a real external seek (forward/backward button or timeline jump)
    console.log('🔄 External time control: jumping to', externalCurrentTime.toISOString());
    setCurrentTime(externalCurrentTime);
    setWebrtcPlaybackTime(externalCurrentTime); // Update WebRTC time to trigger reconnection
    frameStartTimeRef.current = externalCurrentTime.getTime();
  }, [externalCurrentTime, currentTime]);

  // Track video frames and update playback time (like test-webrtc-playback.html)
  useEffect(() => {
    const video = videoRef.current;
    if (!video || !isConnected || !isPlaying) return;

    let rafId: number | null = null;

    const onFrameReceived = (now: number, metadata: any) => {
      if (!frameStartTimeRef.current) return;

      // Calculate current playback time from RTP timestamp
      // metadata.rtpTimestamp is in milliseconds offset from stream start
      const frameDate = new Date(frameStartTimeRef.current + metadata.rtpTimestamp);

      // Update internal currentTime (for timeline display in parent)
      // Like test-webrtc-playback.html, this only updates UI, doesn't trigger reconnection
      setCurrentTime(frameDate);

      // Notify parent ONLY for timeline UI update (not for triggering seeks)
      // Parent should update timeline position but NOT update externalCurrentTime prop
      onPlaybackTimeChange?.(frameDate);

      // Request next frame
      if (video && video.readyState >= 2) {
        rafId = video.requestVideoFrameCallback(onFrameReceived);
      }
    };

    // Start frame tracking
    if (video.requestVideoFrameCallback) {
      rafId = video.requestVideoFrameCallback(onFrameReceived);
    }

    return () => {
      if (rafId !== null && video.cancelVideoFrameCallback) {
        video.cancelVideoFrameCallback(rafId);
      }
    };
  }, [videoRef, isConnected, isPlaying, onPlaybackTimeChange]);

  // Auto-hide controls
  useEffect(() => {
    if (!isPlaying) return;

    const timeout = setTimeout(() => {
      setShowControls(false);
    }, 3000);

    return () => clearTimeout(timeout);
  }, [isPlaying, showControls]);

  return (
    <div
      ref={containerRef}
      className={cn(
        'relative bg-black rounded-lg overflow-hidden group',
        className
      )}
      onMouseMove={() => setShowControls(true)}
      onMouseLeave={() => isPlaying && setShowControls(false)}
    >
      {/* Video Element */}
      <video
        ref={videoRef}
        className="w-full h-full object-contain"
        onClick={togglePlayPause}
        playsInline
      />

      {/* Loading Overlay - only on first connection, not during navigation */}
      {isConnecting && !hasConnectedOnceRef.current && (
        <div className="absolute inset-0 flex items-center justify-center bg-black/50">
          <div className="text-center text-white">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-white mx-auto mb-2" />
            <p className="text-sm">Connecting to playback...</p>
          </div>
        </div>
      )}

      {/* Error Overlay */}
      {error && (
        <div className="absolute inset-0 flex items-center justify-center bg-black/80">
          <div className="text-center text-white">
            <p className="text-lg mb-2">⚠️ Playback Error</p>
            <p className="text-sm text-gray-300">{error}</p>
            <Button
              onClick={() => {
                stop();
                setCurrentTime(initialPlaybackTime || startTime);
              }}
              className="mt-4"
            >
              Retry
            </Button>
          </div>
        </div>
      )}

      {/* Controls Overlay */}
      {showBuiltInControls && (
        <div
          className={cn(
            'absolute inset-x-0 bottom-0 bg-gradient-to-t from-black/80 to-transparent p-4 transition-opacity duration-300 z-20',
            showControls ? 'opacity-100' : 'opacity-0'
          )}
        >
        <div className="flex items-center gap-3">
          {/* Play/Pause */}
          <Button
            variant="secondary"
            size="sm"
            onClick={togglePlayPause}
            disabled={!isConnected}
            className="bg-white/10 hover:bg-white/20 text-white border-0 p-2 disabled:opacity-50"
          >
            {isPlaying ? <Pause className="w-5 h-5" /> : <Play className="w-5 h-5" />}
          </Button>

          {/* Skip Backward */}
          <Button
            variant="secondary"
            size="sm"
            onClick={() => skip(-10)}
            disabled={isConnecting}
            className="bg-white/10 hover:bg-white/20 text-white border-0 p-2 disabled:opacity-50"
          >
            <SkipBack className="w-4 h-4" />
          </Button>

          {/* Skip Forward */}
          <Button
            variant="secondary"
            size="sm"
            onClick={() => skip(10)}
            disabled={isConnecting}
            className="bg-white/10 hover:bg-white/20 text-white border-0 p-2 disabled:opacity-50"
          >
            <SkipForward className="w-4 h-4" />
          </Button>

          {/* Time Display */}
          <div className="flex-1 text-center">
            <div className="text-sm text-white font-mono">
              {formatTime(currentTime)}
            </div>
            <div className="text-xs text-gray-400 mt-1">
              {formatTime(startTime)} - {formatTime(endTime)}
            </div>
          </div>

          {/* Connection Status */}
          <div className="text-xs text-gray-300">
            {isConnecting && '🔄 Connecting...'}
            {isConnected && '🟢 Connected'}
            {state === 'failed' && '🔴 Failed'}
          </div>

          {/* Fullscreen */}
          <Button
            variant="secondary"
            size="sm"
            onClick={toggleFullscreen}
            className="bg-white/10 hover:bg-white/20 text-white border-0 p-2"
          >
            {isFullscreen ? <Minimize className="w-4 h-4" /> : <Maximize className="w-4 h-4" />}
          </Button>
        </div>
      </div>
      )}

      {/* Center Play Button (when paused and connected) */}
      {!isPlaying && isConnected && (
        <div
          className="absolute inset-0 flex items-center justify-center cursor-pointer"
          onClick={togglePlayPause}
        >
          <div className="w-20 h-20 bg-white/20 hover:bg-white/30 rounded-full flex items-center justify-center backdrop-blur-sm transition-all">
            <Play className="w-10 h-10 text-white ml-1" />
          </div>
        </div>
      )}
    </div>
  );
}
