import React, { useEffect, useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import { LiveStreamPlayer } from '@/components/LiveStreamPlayer';
import { RecordingPlayer } from '@/components/RecordingPlayer';
import { PlaybackControlBar } from '@/components/playback/PlaybackControlBar';
import { useTheme } from '@/contexts/ThemeContext';
import { cn } from '@/utils/cn';
import { Plus } from 'lucide-react';
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

interface GridCell {
  camera: Camera | null;
  playbackState: PlaybackState;
}

export function GridView() {
  const [searchParams] = useSearchParams();
  const { theme, setTheme } = useTheme();

  const [cells, setCells] = useState<GridCell[]>(
    Array.from({ length: 9 }, () => ({
      camera: null,
      playbackState: {
        mode: 'live' as const,
        startTime: new Date(Date.now() - 24 * 60 * 60 * 1000), // 24 hours ago
        endTime: new Date(),
        currentTime: new Date(),
        isPlaying: false,
        zoomLevel: 1,
      },
    }))
  );
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Handle theme from URL parameter
  useEffect(() => {
    const themeParam = searchParams.get('theme');
    if (themeParam === 'light' || themeParam === 'dark') {
      setTheme(themeParam);
    }
  }, [searchParams, setTheme]);

  // Fetch cameras data
  useEffect(() => {
    const camerasParam = searchParams.get('cameras');
    if (!camerasParam) {
      setLoading(false);
      return;
    }

    const cameraIds = camerasParam.split(',').filter(id => id.trim()).slice(0, 9); // Max 9 cameras

    const fetchCameras = async () => {
      try {
        setLoading(true);
        const fetchPromises = cameraIds.map(async (id) => {
          try {
            const response = await fetch(`http://localhost:8000/api/v1/cameras/${id.trim()}`);
            if (!response.ok) {
              console.error(`Failed to fetch camera ${id}: ${response.statusText}`);
              return null;
            }
            return await response.json();
          } catch (err) {
            console.error(`Error fetching camera ${id}:`, err);
            return null;
          }
        });

        const camerasData = await Promise.all(fetchPromises);

        setCells(prev => {
          const newCells = [...prev];
          camerasData.forEach((camera, index) => {
            if (camera && index < 9) {
              newCells[index] = {
                ...newCells[index],
                camera,
              };
            }
          });
          return newCells;
        });

        setError(null);
      } catch (err) {
        console.error('Error fetching cameras:', err);
        setError(err instanceof Error ? err.message : 'Failed to load cameras');
      } finally {
        setLoading(false);
      }
    };

    fetchCameras();
  }, [searchParams]);

  const handleModeChange = (index: number, mode: 'live' | 'playback') => {
    setCells(prev => {
      const newCells = [...prev];
      newCells[index] = {
        ...newCells[index],
        playbackState: {
          ...newCells[index].playbackState,
          mode,
        },
      };
      return newCells;
    });
  };

  const handlePlayPause = (index: number) => {
    setCells(prev => {
      const newCells = [...prev];
      newCells[index] = {
        ...newCells[index],
        playbackState: {
          ...newCells[index].playbackState,
          isPlaying: !newCells[index].playbackState.isPlaying,
        },
      };
      return newCells;
    });
  };

  const handleSeek = (index: number, time: Date) => {
    setCells(prev => {
      const newCells = [...prev];
      newCells[index] = {
        ...newCells[index],
        playbackState: {
          ...newCells[index].playbackState,
          currentTime: time,
        },
      };
      return newCells;
    });
  };

  const handleScrollTimeline = (index: number, direction: 'left' | 'right') => {
    const scrollAmount = 60 * 60 * 1000; // 1 hour in milliseconds
    setCells(prev => {
      const newCells = [...prev];
      const state = newCells[index].playbackState;
      newCells[index] = {
        ...newCells[index],
        playbackState: {
          ...state,
          startTime: new Date(state.startTime.getTime() + (direction === 'right' ? scrollAmount : -scrollAmount)),
          endTime: new Date(state.endTime.getTime() + (direction === 'right' ? scrollAmount : -scrollAmount)),
        },
      };
      return newCells;
    });
  };

  const handleZoomChange = (index: number, zoomLevel: number) => {
    setCells(prev => {
      const newCells = [...prev];
      newCells[index] = {
        ...newCells[index],
        playbackState: {
          ...newCells[index].playbackState,
          zoomLevel,
        },
      };
      return newCells;
    });
  };

  const handleSpeedChange = (index: number, speed: number) => {
    console.log(`Cell ${index} playback speed changed:`, speed);
  };

  if (loading) {
    return (
      <div className="h-screen w-screen bg-white dark:bg-dark-base flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 dark:border-blue-500 mx-auto mb-4"></div>
          <p className="text-gray-600 dark:text-text-secondary">Loading cameras...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="h-screen w-screen bg-gray-100 dark:bg-dark-base overflow-hidden">
      {/* Hidden div to force Tailwind to generate dark theme classes */}
      <div className="hidden bg-dark-base bg-dark-secondary bg-dark-sidebar bg-dark-surface bg-dark-border bg-dark-elevated text-text-primary text-text-secondary text-text-muted border-dark-border" />

      {/* 3x3 Grid */}
      <div className="h-full w-full p-2">
        <div className="grid grid-cols-3 grid-rows-3 h-full w-full gap-2">
          {cells.map((cell, index) => (
            <div
              key={index}
              className={cn(
                'relative rounded-lg shadow-md group transition-all',
                cell.camera
                  ? 'bg-gray-900 dark:bg-dark-elevated overflow-visible'
                  : 'bg-gray-200 dark:bg-dark-surface border-2 border-dashed border-gray-300 dark:border-dark-border overflow-hidden'
              )}
            >
              {cell.camera ? (
                <>
                  {/* Video Player - Live or Playback */}
                  {cell.playbackState.mode === 'playback' && cell.camera.milestone_device_id ? (
                    <RecordingPlayer
                      key={cell.camera.id}
                      cameraId={cell.camera.id}
                      startTime={cell.playbackState.startTime}
                      endTime={cell.playbackState.endTime}
                      initialPlaybackTime={cell.playbackState.currentTime}
                      externalIsPlaying={cell.playbackState.isPlaying}
                      externalCurrentTime={cell.playbackState.currentTime}
                      onPlaybackTimeChange={(time) => handleSeek(index, time)}
                      onPlaybackStateChange={(state) => {
                        if (state === 'playing' && !cell.playbackState.isPlaying) {
                          handlePlayPause(index);
                        } else if (state === 'paused' && cell.playbackState.isPlaying) {
                          handlePlayPause(index);
                        }
                      }}
                      showControls={false}
                      className="absolute inset-0"
                    />
                  ) : (
                    <div className="absolute inset-0">
                      <LiveStreamPlayer
                        key={cell.camera.id}
                        camera={cell.camera}
                        quality="medium"
                      />
                    </div>
                  )}

                  {/* Camera info overlay */}
                  <div className="absolute top-0 left-0 right-0 bg-gradient-to-b from-black/70 to-transparent p-2 opacity-0 group-hover:opacity-100 transition-opacity z-10">
                    <div className="flex items-center justify-between gap-2">
                      <div className="flex-1 min-w-0">
                        <p className="text-white text-xs font-medium truncate">
                          {cell.camera.name}
                        </p>
                        {cell.camera.name_ar && (
                          <p className="text-white/80 text-[10px] truncate">
                            {cell.camera.name_ar}
                          </p>
                        )}
                      </div>
                    </div>
                  </div>

                  {/* Playback Controls */}
                  {cell.camera.milestone_device_id && (
                    <div className="absolute bottom-0 left-0 right-0 z-30">
                      <PlaybackControlBar
                        startTime={cell.playbackState.startTime}
                        endTime={cell.playbackState.endTime}
                        currentTime={cell.playbackState.currentTime}
                        sequences={cell.playbackState.timelineData?.sequences || []}
                        isPlaying={cell.playbackState.isPlaying}
                        zoomLevel={cell.playbackState.zoomLevel}
                        onPlayPause={() => handlePlayPause(index)}
                        onSeek={(time) => handleSeek(index, time)}
                        onScrollTimeline={(direction) => handleScrollTimeline(index, direction)}
                        onZoomChange={(zoom) => handleZoomChange(index, zoom)}
                        onSpeedChange={(speed) => handleSpeedChange(index, speed)}
                        hasRecording={true}
                        mode={cell.playbackState.mode}
                        onModeChange={(mode) => handleModeChange(index, mode)}
                      />
                    </div>
                  )}

                  {/* Cell number badge */}
                  {cell.playbackState.mode !== 'playback' && (
                    <div className="absolute bottom-2 left-2 bg-black/70 text-white text-xs px-2 py-1 rounded opacity-0 group-hover:opacity-100 transition-opacity">
                      Cell {index + 1}
                    </div>
                  )}
                </>
              ) : (
                /* Empty cell placeholder - Same as StreamGridEnhanced */
                <div className="absolute inset-0 flex flex-col items-center justify-center text-gray-400 dark:text-text-muted p-4">
                  <div className="text-center">
                    <Plus className="w-10 h-10 mx-auto mb-2 opacity-40" />
                    <p className="text-sm font-medium">Drop camera here</p>
                    <p className="text-xs mt-1 opacity-70">or double-click in sidebar</p>
                  </div>
                  <div className="mt-4 text-xs bg-white/50 dark:bg-dark-elevated/50 px-3 py-1 rounded-full">
                    Cell {index + 1}
                  </div>
                </div>
              )}
            </div>
          ))}
        </div>
      </div>

      {/* Error Toast */}
      {error && (
        <div className="fixed top-4 right-4 bg-red-600 text-white px-4 py-3 rounded-lg shadow-lg z-50">
          <p className="text-sm font-medium">{error}</p>
        </div>
      )}
    </div>
  );
}
