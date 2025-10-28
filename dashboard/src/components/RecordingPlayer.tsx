import React, { useEffect, useRef, useState } from 'react';
import {
  Play,
  Pause,
  SkipBack,
  SkipForward,
  FastForward,
  Rewind,
  Volume2,
  VolumeX,
  Maximize,
  Minimize,
  Download,
} from 'lucide-react';
import { cn } from '@/utils/cn';
import { Button } from './ui/Dialog';
import Hls from 'hls.js';

interface RecordingPlayerProps {
  cameraId: string;
  startTime: Date;
  endTime: Date;
  initialPlaybackTime?: Date;
  streamUrl?: string;
  onPlaybackTimeChange?: (time: Date) => void;
  onPlaybackStateChange?: (state: 'playing' | 'paused' | 'loading' | 'error') => void;
  className?: string;
}

const SPEED_OPTIONS = [
  { label: '-8x', value: -8 },
  { label: '-4x', value: -4 },
  { label: '-2x', value: -2 },
  { label: '-1x', value: -1 },
  { label: '0.5x', value: 0.5 },
  { label: '1x', value: 1 },
  { label: '2x', value: 2 },
  { label: '4x', value: 4 },
  { label: '8x', value: 8 },
];

export function RecordingPlayer({
  cameraId,
  startTime,
  endTime,
  initialPlaybackTime,
  streamUrl,
  onPlaybackTimeChange,
  onPlaybackStateChange,
  className,
}: RecordingPlayerProps) {
  const videoRef = useRef<HTMLVideoElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const hlsRef = useRef<Hls | null>(null);

  const [isPlaying, setIsPlaying] = useState(false);
  const [currentTime, setCurrentTime] = useState(initialPlaybackTime || startTime);
  const [speed, setSpeed] = useState(1);
  const [volume, setVolume] = useState(1);
  const [isMuted, setIsMuted] = useState(false);
  const [isFullscreen, setIsFullscreen] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [showControls, setShowControls] = useState(true);
  const [buffered, setBuffered] = useState(0);

  // Initialize HLS player
  useEffect(() => {
    if (!videoRef.current || !streamUrl) return;

    const video = videoRef.current;

    // Check if HLS is supported natively
    if (video.canPlayType('application/vnd.apple.mpegurl')) {
      video.src = streamUrl;
    } else if (Hls.isSupported()) {
      const hls = new Hls({
        enableWorker: true,
        lowLatencyMode: true,
      });

      hls.loadSource(streamUrl);
      hls.attachMedia(video);

      hls.on(Hls.Events.MANIFEST_PARSED, () => {
        setLoading(false);
      });

      hls.on(Hls.Events.ERROR, (event, data) => {
        if (data.fatal) {
          console.error('HLS Error:', data);
          setError('Failed to load video stream');
          onPlaybackStateChange?.('error');
        }
      });

      hlsRef.current = hls;

      return () => {
        hls.destroy();
      };
    } else {
      setError('HLS is not supported in this browser');
      onPlaybackStateChange?.('error');
    }
  }, [streamUrl, onPlaybackStateChange]);

  // Update playback time
  useEffect(() => {
    const video = videoRef.current;
    if (!video) return;

    const handleTimeUpdate = () => {
      // Calculate actual timestamp based on video position
      const videoDuration = video.duration;
      const videoCurrentTime = video.currentTime;
      const totalDuration = endTime.getTime() - startTime.getTime();
      const progress = videoDuration > 0 ? videoCurrentTime / videoDuration : 0;
      const actualTimestamp = new Date(startTime.getTime() + totalDuration * progress);

      setCurrentTime(actualTimestamp);
      onPlaybackTimeChange?.(actualTimestamp);
    };

    const handleProgress = () => {
      if (video.buffered.length > 0) {
        const bufferedEnd = video.buffered.end(video.buffered.length - 1);
        const duration = video.duration;
        setBuffered((bufferedEnd / duration) * 100);
      }
    };

    video.addEventListener('timeupdate', handleTimeUpdate);
    video.addEventListener('progress', handleProgress);

    return () => {
      video.removeEventListener('timeupdate', handleTimeUpdate);
      video.removeEventListener('progress', handleProgress);
    };
  }, [startTime, endTime, onPlaybackTimeChange]);

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
      setError('Failed to play video');
    }
  };

  // Change playback speed
  const handleSpeedChange = (newSpeed: number) => {
    const video = videoRef.current;
    if (!video) return;

    // For negative speeds, we need to implement custom logic
    if (newSpeed < 0) {
      // TODO: Implement reverse playback
      console.warn('Reverse playback not yet implemented');
      return;
    }

    video.playbackRate = newSpeed;
    setSpeed(newSpeed);
  };

  // Skip forward/backward
  const skip = (seconds: number) => {
    const video = videoRef.current;
    if (!video) return;

    video.currentTime = Math.max(0, Math.min(video.duration, video.currentTime + seconds));
  };

  // Seek to timestamp
  const seekToTime = (timestamp: Date) => {
    const video = videoRef.current;
    if (!video || video.duration === 0) return;

    const totalDuration = endTime.getTime() - startTime.getTime();
    const targetTime = timestamp.getTime() - startTime.getTime();
    const progress = targetTime / totalDuration;
    const videoTime = video.duration * progress;

    video.currentTime = Math.max(0, Math.min(video.duration, videoTime));
  };

  // Volume control
  const handleVolumeChange = (newVolume: number) => {
    const video = videoRef.current;
    if (!video) return;

    video.volume = newVolume;
    setVolume(newVolume);
    setIsMuted(newVolume === 0);
  };

  // Mute toggle
  const toggleMute = () => {
    const video = videoRef.current;
    if (!video) return;

    if (isMuted) {
      video.volume = volume || 0.5;
      setIsMuted(false);
    } else {
      video.volume = 0;
      setIsMuted(true);
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

  // Download recording
  const handleDownload = () => {
    if (!streamUrl) return;
    window.open(streamUrl, '_blank');
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
      />

      {/* Loading Overlay */}
      {loading && (
        <div className="absolute inset-0 flex items-center justify-center bg-black/50">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-white" />
        </div>
      )}

      {/* Error Overlay */}
      {error && (
        <div className="absolute inset-0 flex items-center justify-center bg-black/80">
          <div className="text-center text-white">
            <p className="text-lg mb-2">⚠️ Playback Error</p>
            <p className="text-sm text-gray-300">{error}</p>
          </div>
        </div>
      )}

      {/* Controls Overlay */}
      <div
        className={cn(
          'absolute inset-x-0 bottom-0 bg-gradient-to-t from-black/80 to-transparent transition-opacity duration-300',
          showControls ? 'opacity-100' : 'opacity-0'
        )}
      >
        {/* Progress Bar */}
        <div className="px-4 pb-2">
          <div className="relative h-1 bg-white/30 rounded-full cursor-pointer group/progress">
            {/* Buffered */}
            <div
              className="absolute h-full bg-white/40 rounded-full"
              style={{ width: `${buffered}%` }}
            />
            {/* TODO: Add progress based on current time */}
            <div className="absolute h-full bg-blue-500 rounded-full" style={{ width: '0%' }} />
          </div>
        </div>

        {/* Controls */}
        <div className="px-4 pb-4 flex items-center gap-3">
          {/* Play/Pause */}
          <Button
            variant="secondary"
            size="sm"
            onClick={togglePlayPause}
            className="bg-white/10 hover:bg-white/20 text-white border-0 p-2"
          >
            {isPlaying ? <Pause className="w-5 h-5" /> : <Play className="w-5 h-5" />}
          </Button>

          {/* Skip Backward */}
          <Button
            variant="secondary"
            size="sm"
            onClick={() => skip(-10)}
            className="bg-white/10 hover:bg-white/20 text-white border-0 p-2"
          >
            <Rewind className="w-4 h-4" />
          </Button>

          {/* Skip Forward */}
          <Button
            variant="secondary"
            size="sm"
            onClick={() => skip(10)}
            className="bg-white/10 hover:bg-white/20 text-white border-0 p-2"
          >
            <FastForward className="w-4 h-4" />
          </Button>

          {/* Speed Control */}
          <select
            value={speed}
            onChange={(e) => handleSpeedChange(Number(e.target.value))}
            className="bg-white/10 hover:bg-white/20 text-white border-0 rounded px-2 py-1 text-sm"
          >
            {SPEED_OPTIONS.map((option) => (
              <option key={option.value} value={option.value} className="bg-gray-900">
                {option.label}
              </option>
            ))}
          </select>

          {/* Time Display */}
          <div className="flex-1 text-center">
            <div className="text-sm text-white font-mono">
              {formatTime(currentTime)}
            </div>
          </div>

          {/* Volume */}
          <div className="flex items-center gap-2">
            <Button
              variant="secondary"
              size="sm"
              onClick={toggleMute}
              className="bg-white/10 hover:bg-white/20 text-white border-0 p-2"
            >
              {isMuted ? <VolumeX className="w-4 h-4" /> : <Volume2 className="w-4 h-4" />}
            </Button>
            <input
              type="range"
              min="0"
              max="1"
              step="0.1"
              value={isMuted ? 0 : volume}
              onChange={(e) => handleVolumeChange(Number(e.target.value))}
              className="w-20"
            />
          </div>

          {/* Download */}
          <Button
            variant="secondary"
            size="sm"
            onClick={handleDownload}
            className="bg-white/10 hover:bg-white/20 text-white border-0 p-2"
          >
            <Download className="w-4 h-4" />
          </Button>

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

      {/* Center Play Button (when paused) */}
      {!isPlaying && !loading && !error && (
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
