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
}: RecordingPlayerProps) {
  const containerRef = useRef<HTMLDivElement>(null);

  const [isPlaying, setIsPlaying] = useState(false);
  const [currentTime, setCurrentTime] = useState(initialPlaybackTime || startTime);
  const [isFullscreen, setIsFullscreen] = useState(false);
  const [showControls, setShowControls] = useState(true);

  // Use WebRTC playback hook
  const { videoRef, state, error, stop, isConnecting, isConnected } = useWebRTCPlayback({
    cameraId,
    playbackTime: currentTime,
    skipGaps: true,
    speed: 1.0,
    onStateChange: (webrtcState) => {
      if (webrtcState === 'connected') {
        onPlaybackStateChange?.('playing');
        setIsPlaying(true);
      } else if (webrtcState === 'failed') {
        onPlaybackStateChange?.('error');
        setIsPlaying(false);
      } else if (webrtcState === 'connecting') {
        onPlaybackStateChange?.('loading');
      } else if (webrtcState === 'disconnected') {
        setIsPlaying(false);
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
        autoPlay
        playsInline
      />

      {/* Loading Overlay */}
      {isConnecting && (
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
