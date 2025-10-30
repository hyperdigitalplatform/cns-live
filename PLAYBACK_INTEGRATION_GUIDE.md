# Playback Mode Integration Guide

This guide explains how to integrate the playback mode components into your existing grid cells.

## Quick Start

### 1. Import Required Components

```typescript
import { PlaybackModeToggle, PlaybackControlBar } from '@/components/playback';
import { PlaybackState, TimelineData } from '@/types/playback';
import { LiveStreamPlayer } from '@/components/LiveStreamPlayer';
import { RecordingPlayer } from '@/components/RecordingPlayer';
```

### 2. Add Playback State to Grid Cell

```typescript
interface GridCellPlaybackState {
  mode: 'live' | 'playback';
  isPlaying: boolean;
  currentTime: Date;
  startTime: Date;
  endTime: Date;
  speed: number;
  zoomLevel: number;
  timelineData: TimelineData | null;
  streamUrl: string | null;
}

// Initialize state
const [playbackState, setPlaybackState] = useState<GridCellPlaybackState>({
  mode: 'live',
  isPlaying: false,
  currentTime: new Date(),
  startTime: new Date(Date.now() - 24 * 60 * 60 * 1000), // 24 hours ago
  endTime: new Date(),
  speed: 1.0,
  zoomLevel: 12, // 12 hours
  timelineData: null,
  streamUrl: null,
});
```

### 3. Handle Mode Toggle

```typescript
const handleModeChange = async (newMode: 'live' | 'playback') => {
  if (newMode === 'playback' && !playbackState.timelineData) {
    // Query recordings when switching to playback
    try {
      const data = await api.getMilestoneSequences({
        cameraId: camera.id,
        startTime: playbackState.startTime.toISOString(),
        endTime: playbackState.endTime.toISOString(),
      });

      setPlaybackState(prev => ({
        ...prev,
        mode: 'playback',
        timelineData: data,
      }));
    } catch (error) {
      console.error('Failed to query recordings:', error);
    }
  } else {
    setPlaybackState(prev => ({ ...prev, mode: newMode }));
  }
};
```

### 4. Render Grid Cell with Playback Support

```typescript
<div className="grid-cell">
  {/* Header with mode toggle */}
  <div className="cell-header">
    <span>{camera.name}</span>
    <PlaybackModeToggle
      mode={playbackState.mode}
      onChange={handleModeChange}
    />
    <button onClick={handleFullscreen}>Fullscreen</button>
  </div>

  {/* Video Area */}
  <div className="video-area">
    {playbackState.mode === 'live' ? (
      <LiveStreamPlayer camera={camera} />
    ) : (
      <RecordingPlayer
        cameraId={camera.id}
        startTime={playbackState.startTime}
        endTime={playbackState.endTime}
        initialPlaybackTime={playbackState.currentTime}
        streamUrl={playbackState.streamUrl}
        onPlaybackTimeChange={(time) => {
          setPlaybackState(prev => ({ ...prev, currentTime: time }));
        }}
        onPlaybackStateChange={(state) => {
          setPlaybackState(prev => ({
            ...prev,
            isPlaying: state === 'playing',
          }));
        }}
      />
    )}
  </div>

  {/* Playback Controls (only in playback mode) */}
  {playbackState.mode === 'playback' && playbackState.timelineData && (
    <PlaybackControlBar
      startTime={playbackState.startTime}
      endTime={playbackState.endTime}
      currentTime={playbackState.currentTime}
      sequences={playbackState.timelineData.sequences}
      isPlaying={playbackState.isPlaying}
      zoomLevel={playbackState.zoomLevel}
      onPlayPause={handlePlayPause}
      onSeek={handleSeek}
      onScrollTimeline={handleScrollTimeline}
      onZoomChange={handleZoomChange}
      hasRecording={hasRecordingAtCurrentTime()}
    />
  )}
</div>
```

## Handler Functions

### Play/Pause Handler

```typescript
const handlePlayPause = () => {
  setPlaybackState(prev => ({
    ...prev,
    isPlaying: !prev.isPlaying,
  }));

  // Trigger video player play/pause
  // This is handled by RecordingPlayer through onPlaybackStateChange
};
```

### Seek Handler

```typescript
const handleSeek = async (newTime: Date) => {
  // Update state
  setPlaybackState(prev => ({
    ...prev,
    currentTime: newTime,
  }));

  // Optionally: Request new playback stream from backend
  try {
    const response = await api.startMilestonePlayback({
      cameraId: camera.id,
      timestamp: newTime.toISOString(),
      speed: playbackState.speed,
      format: 'hls',
    });

    setPlaybackState(prev => ({
      ...prev,
      streamUrl: response.streamUrl,
    }));
  } catch (error) {
    console.error('Failed to seek:', error);
  }
};
```

### Scroll Timeline Handler

```typescript
const handleScrollTimeline = (direction: 'left' | 'right') => {
  const scrollAmount = playbackState.zoomLevel * 60 * 60 * 1000; // zoom level in ms

  setPlaybackState(prev => {
    const newStartTime = new Date(
      direction === 'left'
        ? prev.startTime.getTime() - scrollAmount
        : prev.startTime.getTime() + scrollAmount
    );
    const newEndTime = new Date(
      direction === 'left'
        ? prev.endTime.getTime() - scrollAmount
        : prev.endTime.getTime() + scrollAmount
    );

    return {
      ...prev,
      startTime: newStartTime,
      endTime: newEndTime,
    };
  });

  // Optionally: Re-query recordings for new time range
  // queryRecordings(newStartTime, newEndTime);
};
```

### Zoom Change Handler

```typescript
const handleZoomChange = (newZoomLevel: number) => {
  setPlaybackState(prev => {
    const center = prev.currentTime.getTime();
    const halfDuration = (newZoomLevel * 60 * 60 * 1000) / 2;

    return {
      ...prev,
      zoomLevel: newZoomLevel,
      startTime: new Date(center - halfDuration),
      endTime: new Date(center + halfDuration),
    };
  });
};
```

### Check Recording Availability

```typescript
const hasRecordingAtCurrentTime = (): boolean => {
  if (!playbackState.timelineData) return false;

  const currentTimestamp = playbackState.currentTime.getTime();

  return playbackState.timelineData.sequences.some(seq => {
    const start = new Date(seq.startTime).getTime();
    const end = new Date(seq.endTime).getTime();
    return currentTimestamp >= start && currentTimestamp <= end;
  });
};
```

## API Integration

### Query Recordings

```typescript
export const api = {
  getMilestoneSequences: async (params: {
    cameraId: string;
    startTime: string;
    endTime: string;
  }) => {
    const response = await fetch(
      `/api/v1/playback/cameras/${params.cameraId}/sequences?` +
      new URLSearchParams({
        startTime: params.startTime,
        endTime: params.endTime,
      })
    );

    if (!response.ok) {
      throw new Error('Failed to query recordings');
    }

    const data = await response.json();

    // Transform to TimelineData format
    return {
      cameraId: params.cameraId,
      queryRange: {
        start: params.startTime,
        end: params.endTime,
      },
      sequences: data.sequences || [],
      gaps: data.gaps || [],
      totalRecordingSeconds: data.totalRecordingSeconds || 0,
      totalGapSeconds: data.totalGapSeconds || 0,
      coverage: data.coverage || 0,
    };
  },

  startMilestonePlayback: async (params: {
    cameraId: string;
    timestamp: string;
    speed: number;
    format: string;
  }) => {
    const response = await fetch(
      `/api/v1/playback/cameras/${params.cameraId}/start`,
      {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(params),
      }
    );

    if (!response.ok) {
      throw new Error('Failed to start playback');
    }

    return response.json();
  },
};
```

## StreamGridEnhanced Integration

To integrate into `StreamGridEnhanced.tsx`, modify the `GridCell` interface:

```typescript
interface GridCell {
  camera: Camera | null;
  loading: boolean;
  isHotspot?: boolean;
  gridArea?: string;
  // Add playback state
  playbackState?: GridCellPlaybackState;
}
```

Update cell rendering to support playback mode:

```typescript
{gridCells.map((cell, index) => (
  <div key={index} className="grid-cell">
    {cell.camera ? (
      <>
        {/* Cell header with mode toggle */}
        <div className="cell-header">
          <span>{cell.camera.name}</span>
          {cell.camera.milestone_device_id && (
            <PlaybackModeToggle
              mode={cell.playbackState?.mode || 'live'}
              onChange={(mode) => handleCellModeChange(index, mode)}
            />
          )}
          <button onClick={() => handleFullscreen(index)}>
            Fullscreen
          </button>
        </div>

        {/* Video player */}
        {cell.playbackState?.mode === 'playback' ? (
          <RecordingPlayer {...playbackProps} />
        ) : (
          <LiveStreamPlayer camera={cell.camera} />
        )}

        {/* Playback controls */}
        {cell.playbackState?.mode === 'playback' &&
         cell.playbackState.timelineData && (
          <PlaybackControlBar {...controlBarProps} />
        )}
      </>
    ) : (
      <EmptyCellPlaceholder />
    )}
  </div>
))}
```

## Styling

Add these styles to your CSS:

```css
.grid-cell {
  position: relative;
  display: flex;
  flex-direction: column;
  background: #000;
  border-radius: 8px;
  overflow: hidden;
}

.cell-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  background: rgba(0, 0, 0, 0.7);
  backdrop-filter: blur(10px);
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  z-index: 10;
}

.video-area {
  flex: 1;
  position: relative;
}
```

## Best Practices

1. **State Management**: Keep playback state per cell, not global
2. **API Caching**: Cache timeline data to avoid repeated queries
3. **Error Handling**: Show user-friendly messages when recordings unavailable
4. **Performance**: Limit number of cells in playback mode simultaneously
5. **Auto-collapse**: Let controls auto-hide during playback for cleaner UI
6. **Keyboard Shortcuts**: Consider adding Space for play/pause globally

## Troubleshooting

### Issue: Timeline doesn't show recordings
**Solution**: Verify API response format matches `TimelineData` interface

### Issue: Seek doesn't work
**Solution**: Check that `onSeek` handler updates both state and backend

### Issue: Playback mode button doesn't appear
**Solution**: Ensure camera has `milestone_device_id` property

### Issue: Controls don't collapse
**Solution**: Check `isPlaying` state is properly updated

## Example: Complete Grid Cell Component

See `PLAYBACK_MODE_DESIGN.md` for complete component structure and detailed specifications.

## Next Steps

1. Test with real Milestone backend
2. Add loading states during API calls
3. Implement error boundaries for playback failures
4. Add keyboard shortcuts
5. Consider adding export/download features
6. Implement bookmarks on timeline

---

**For more details, see:**
- `PLAYBACK_MODE_DESIGN.md` - Complete design specification
- `dashboard/src/components/playback/` - Component implementations
- `dashboard/src/types/playback.ts` - TypeScript interfaces
