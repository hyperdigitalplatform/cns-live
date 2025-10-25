import React, { useEffect, useRef, useState } from 'react';
import Hls from 'hls.js';
import { api } from '@/services/api';
import { PlaybackTimeline } from './PlaybackTimeline';
import type { Camera } from '@/types';
import {
  Play,
  Pause,
  Volume2,
  VolumeX,
  Maximize,
  Loader2,
  AlertCircle,
} from 'lucide-react';
import { format } from 'date-fns';

interface PlaybackPlayerProps {
  camera: Camera;
  startTime: Date;
  endTime: Date;
  onClose?: () => void;
}

export function PlaybackPlayer({
  camera,
  startTime,
  endTime,
  onClose,
}: PlaybackPlayerProps) {
  const videoRef = useRef<HTMLVideoElement>(null);
  const hlsRef = useRef<Hls | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [playing, setPlaying] = useState(false);
  const [muted, setMuted] = useState(false);
  const [currentTime, setCurrentTime] = useState(0);
  const [duration, setDuration] = useState(0);
  const [segments, setSegments] = useState<Array<{ start: Date; end: Date }>>(
    []
  );

  useEffect(() => {
    let mounted = true;

    const initPlayback = async () => {
      try {
        setLoading(true);
        setError(null);

        // Request playback session
        const response = await api.requestPlayback({
          camera_id: camera.id,
          start_time: startTime.toISOString(),
          end_time: endTime.toISOString(),
          format: 'hls',
          user_id: 'dashboard-user',
        });

        if (!mounted) return;

        // Parse segment_ids from response to create visual timeline
        if (response.segment_ids && response.segment_ids.length > 0) {
          // Convert segment IDs to approximate time ranges
          // Assuming each segment is ~60 seconds (this should come from API)
          const segmentDuration = 60; // seconds
          const segs = response.segment_ids.map((_, index) => {
            const segStart = new Date(
              startTime.getTime() + index * segmentDuration * 1000
            );
            const segEnd = new Date(
              startTime.getTime() + (index + 1) * segmentDuration * 1000
            );
            return { start: segStart, end: segEnd };
          });
          setSegments(segs);
        }

        const video = videoRef.current;
        if (!video) return;

        // Initialize HLS.js
        if (Hls.isSupported()) {
          const hls = new Hls({
            enableWorker: true,
            lowLatencyMode: false,
          });

          hls.loadSource(response.url);
          hls.attachMedia(video);

          hls.on(Hls.Events.MANIFEST_PARSED, () => {
            setLoading(false);
          });

          hls.on(Hls.Events.ERROR, (event, data) => {
            if (data.fatal) {
              setError(`Playback error: ${data.type}`);
              setLoading(false);
            }
          });

          hlsRef.current = hls;
        } else if (video.canPlayType('application/vnd.apple.mpegurl')) {
          // Native HLS support (Safari)
          video.src = response.url;
          video.addEventListener('loadedmetadata', () => {
            setLoading(false);
          });
        } else {
          setError('HLS not supported in this browser');
          setLoading(false);
        }
      } catch (err) {
        if (mounted) {
          setError(
            err instanceof Error ? err.message : 'Failed to load playback'
          );
          setLoading(false);
        }
      }
    };

    initPlayback();

    return () => {
      mounted = false;
      if (hlsRef.current) {
        hlsRef.current.destroy();
      }
    };
  }, [camera.id, startTime, endTime]);

  useEffect(() => {
    const video = videoRef.current;
    if (!video) return;

    const handleTimeUpdate = () => setCurrentTime(video.currentTime);
    const handleDurationChange = () => setDuration(video.duration);
    const handlePlay = () => setPlaying(true);
    const handlePause = () => setPlaying(false);

    video.addEventListener('timeupdate', handleTimeUpdate);
    video.addEventListener('durationchange', handleDurationChange);
    video.addEventListener('play', handlePlay);
    video.addEventListener('pause', handlePause);

    return () => {
      video.removeEventListener('timeupdate', handleTimeUpdate);
      video.removeEventListener('durationchange', handleDurationChange);
      video.removeEventListener('play', handlePlay);
      video.removeEventListener('pause', handlePause);
    };
  }, []);

  const togglePlay = () => {
    const video = videoRef.current;
    if (!video) return;

    if (playing) {
      video.pause();
    } else {
      video.play();
    }
  };

  const toggleMute = () => {
    const video = videoRef.current;
    if (!video) return;

    video.muted = !video.muted;
    setMuted(video.muted);
  };

  const handleSeek = (time: number) => {
    const video = videoRef.current;
    if (!video) return;

    // Convert time from timeline (seconds from startTime) to video currentTime
    video.currentTime = time;
  };

  const toggleFullscreen = () => {
    const video = videoRef.current;
    if (!video) return;

    if (document.fullscreenElement) {
      document.exitFullscreen();
    } else {
      video.requestFullscreen();
    }
  };

  const formatTime = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = Math.floor(seconds % 60);
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full bg-gray-900">
        <div className="text-center text-white">
          <Loader2 className="w-12 h-12 animate-spin mx-auto mb-3" />
          <p>Loading playback...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex flex-col items-center justify-center h-full bg-gray-900 text-white p-4">
        <AlertCircle className="w-16 h-16 text-red-500 mb-4" />
        <p className="text-lg font-medium mb-2">Playback Error</p>
        <p className="text-sm text-gray-400">{error}</p>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-full bg-black">
      {/* Video Header */}
      <div className="bg-gray-900 px-4 py-3 border-b border-gray-800">
        <div className="flex items-center justify-between">
          <div>
            <h3 className="text-white font-medium">{camera.name}</h3>
            <p className="text-sm text-gray-400">
              {format(startTime, 'PPpp')} - {format(endTime, 'PPpp')}
            </p>
          </div>
          {onClose && (
            <button
              onClick={onClose}
              className="text-gray-400 hover:text-white transition-colors"
            >
              ×
            </button>
          )}
        </div>
      </div>

      {/* Video */}
      <div className="flex-1 relative">
        <video
          ref={videoRef}
          className="w-full h-full object-contain"
          controls={false}
        />
      </div>

      {/* Controls */}
      <div className="bg-gray-900 px-4 py-4 space-y-3">
        {/* Enhanced Timeline */}
        <PlaybackTimeline
          startTime={startTime}
          endTime={endTime}
          segments={segments}
          currentTime={currentTime}
          duration={(endTime.getTime() - startTime.getTime()) / 1000}
          onSeek={handleSeek}
        />

        <div className="flex items-center justify-between">
          {/* Left controls */}
          <div className="flex items-center gap-3">
            <button
              onClick={togglePlay}
              className="text-white hover:text-primary-400 transition-colors"
            >
              {playing ? (
                <Pause className="w-6 h-6" />
              ) : (
                <Play className="w-6 h-6" />
              )}
            </button>

            <button
              onClick={toggleMute}
              className="text-white hover:text-primary-400 transition-colors"
            >
              {muted ? (
                <VolumeX className="w-6 h-6" />
              ) : (
                <Volume2 className="w-6 h-6" />
              )}
            </button>

            <span className="text-sm text-gray-400">
              {formatTime(currentTime)} / {formatTime(duration)}
            </span>
          </div>

          {/* Right controls */}
          <div>
            <button
              onClick={toggleFullscreen}
              className="text-white hover:text-primary-400 transition-colors"
            >
              <Maximize className="w-6 h-6" />
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
