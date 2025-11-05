import React, { useEffect, useState } from 'react';
import { useParams, useSearchParams } from 'react-router-dom';
import { LiveStreamPlayer } from '@/components/LiveStreamPlayer';
import { RecordingPlayer } from '@/components/RecordingPlayer';
import { PlaybackControlBar } from '@/components/playback/PlaybackControlBar';
import { useTheme } from '@/contexts/ThemeContext';
import { cn } from '@/utils/cn';
import { Camera } from '@/types';

interface PlaybackState {
  mode: 'live' | 'playback';
  startTime: Date;
  endTime: Date;
  currentTime: Date;
  isPlaying: boolean;
  zoomLevel: number;
  timelineData?: any;
}

export function SingleCameraView() {
  const { cameraId } = useParams<{ cameraId: string }>();
  const [searchParams] = useSearchParams();
  const { theme, setTheme } = useTheme();

  const [camera, setCamera] = useState<Camera | null>(null);
  const [playbackState, setPlaybackState] = useState<PlaybackState>({
    mode: 'live',
    startTime: new Date(Date.now() - 24 * 60 * 60 * 1000), // 24 hours ago
    endTime: new Date(),
    currentTime: new Date(),
    isPlaying: false,
    zoomLevel: 1,
  });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Handle theme from URL parameter
  useEffect(() => {
    const themeParam = searchParams.get('theme');
    if (themeParam === 'light' || themeParam === 'dark') {
      setTheme(themeParam);
    }
  }, [searchParams, setTheme]);

  // Fetch camera data
  useEffect(() => {
    if (!cameraId) return;

    const fetchCamera = async () => {
      try {
        setLoading(true);
        const response = await fetch(`http://localhost:8000/api/v1/cameras/${cameraId}`);

        if (!response.ok) {
          throw new Error(`Failed to fetch camera: ${response.statusText}`);
        }

        const data = await response.json();
        setCamera(data);
        setError(null);
      } catch (err) {
        console.error('Error fetching camera:', err);
        setError(err instanceof Error ? err.message : 'Failed to load camera');
      } finally {
        setLoading(false);
      }
    };

    fetchCamera();
  }, [cameraId]);

  const handleModeChange = (mode: 'live' | 'playback') => {
    setPlaybackState(prev => ({ ...prev, mode }));
  };

  const handlePlayPause = () => {
    setPlaybackState(prev => ({ ...prev, isPlaying: !prev.isPlaying }));
  };

  const handleSeek = (time: Date) => {
    setPlaybackState(prev => ({ ...prev, currentTime: time }));
  };

  const handleScrollTimeline = (direction: 'left' | 'right') => {
    const scrollAmount = 60 * 60 * 1000; // 1 hour in milliseconds
    setPlaybackState(prev => ({
      ...prev,
      startTime: new Date(prev.startTime.getTime() + (direction === 'right' ? scrollAmount : -scrollAmount)),
      endTime: new Date(prev.endTime.getTime() + (direction === 'right' ? scrollAmount : -scrollAmount)),
    }));
  };

  const handleZoomChange = (zoomLevel: number) => {
    setPlaybackState(prev => ({ ...prev, zoomLevel }));
  };

  const handleSpeedChange = (speed: number) => {
    console.log('Playback speed changed:', speed);
  };

  if (loading) {
    return (
      <div className="h-screen w-screen bg-white dark:bg-dark-base flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 dark:border-blue-500 mx-auto mb-4"></div>
          <p className="text-gray-600 dark:text-text-secondary">Loading camera...</p>
        </div>
      </div>
    );
  }

  if (error || !camera) {
    return (
      <div className="h-screen w-screen bg-white dark:bg-dark-base flex items-center justify-center">
        <div className="text-center">
          <div className="text-red-600 dark:text-red-400 text-6xl mb-4">⚠</div>
          <h2 className="text-xl font-semibold text-gray-900 dark:text-text-primary mb-2">
            Camera Not Found
          </h2>
          <p className="text-gray-600 dark:text-text-secondary">
            {error || `Camera with ID "${cameraId}" could not be loaded.`}
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="h-screen w-screen bg-black dark:bg-dark-base flex flex-col overflow-hidden">
      {/* Hidden div to force Tailwind to generate dark theme classes */}
      <div className="hidden bg-dark-base bg-dark-secondary bg-dark-sidebar bg-dark-surface bg-dark-border bg-dark-elevated text-text-primary text-text-secondary text-text-muted border-dark-border" />

      {/* Video Player */}
      <div className="relative flex-1 bg-gray-900 dark:bg-dark-elevated">
        {playbackState.mode === 'playback' && camera.milestone_device_id ? (
          <RecordingPlayer
            key={camera.id}
            cameraId={camera.id}
            startTime={playbackState.startTime}
            endTime={playbackState.endTime}
            initialPlaybackTime={playbackState.currentTime}
            externalIsPlaying={playbackState.isPlaying}
            externalCurrentTime={playbackState.currentTime}
            onPlaybackTimeChange={handleSeek}
            onPlaybackStateChange={(state) => {
              if (state === 'playing' && !playbackState.isPlaying) {
                handlePlayPause();
              } else if (state === 'paused' && playbackState.isPlaying) {
                handlePlayPause();
              }
            }}
            showControls={false}
            className="absolute inset-0"
          />
        ) : (
          <div className="absolute inset-0">
            <LiveStreamPlayer
              key={camera.id}
              camera={camera}
              quality="high"
            />
          </div>
        )}

        {/* Camera Info Overlay */}
        <div className="absolute top-0 left-0 right-0 bg-gradient-to-b from-black/70 to-transparent p-4 z-10">
          <div className="flex items-center justify-between">
            <div className="flex-1 min-w-0">
              <h1 className="text-white text-xl font-semibold truncate">
                {camera.name}
              </h1>
              {camera.name_ar && (
                <p className="text-white/80 text-sm truncate">
                  {camera.name_ar}
                </p>
              )}
            </div>
          </div>
        </div>

        {/* Playback Controls - Only show if camera has Milestone recording */}
        {camera.milestone_device_id && (
          <div className="absolute bottom-0 left-0 right-0 z-30">
            <PlaybackControlBar
              startTime={playbackState.startTime}
              endTime={playbackState.endTime}
              currentTime={playbackState.currentTime}
              sequences={playbackState.timelineData?.sequences || []}
              isPlaying={playbackState.isPlaying}
              zoomLevel={playbackState.zoomLevel}
              onPlayPause={handlePlayPause}
              onSeek={handleSeek}
              onScrollTimeline={handleScrollTimeline}
              onZoomChange={handleZoomChange}
              onSpeedChange={handleSpeedChange}
              hasRecording={true}
              mode={playbackState.mode}
              onModeChange={handleModeChange}
            />
          </div>
        )}
      </div>
    </div>
  );
}
