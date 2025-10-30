import React, { useState, forwardRef, useImperativeHandle, useRef, useCallback } from 'react';
import { LiveStreamPlayer } from './LiveStreamPlayer';
import { RecordingPlayer } from './RecordingPlayer';
import { PlaybackModeToggle, PlaybackControlBar } from './playback';
import { SaveLayoutDialog } from './SaveLayoutDialog';
import { LoadLayoutDropdown } from './LoadLayoutDropdown';
import { LayoutManagerDialog } from './LayoutManagerDialog';
import { ToastContainer } from './Toast';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogBody,
  DialogFooter,
  Button,
} from './ui/Dialog';
import type { Camera, DragItem, LayoutType, LayoutPreferenceSummary } from '@/types';
import type { PlaybackState, TimelineData } from '@/types/playback';
import { Grid, Maximize2, X, Plus, Trash2, Save, Settings, AlertTriangle } from 'lucide-react';
import { cn } from '@/utils/cn';
import { api } from '@/services/api';
import { useToast } from '@/hooks/useToast';

export interface StreamGridEnhancedRef {
  addCameraToNextAvailableCell: (camera: Camera) => boolean;
}

interface StreamGridEnhancedProps {
  defaultLayout?: GridLayoutType;
  onLayoutChange?: (layout: GridLayoutType) => void;
}

type GridLayoutType =
  | '2x2' | '2x3' | '3x3' | '3x4'
  | '9-way-1-hotspot' | '12-way-1-hotspot' | '16-way-1-hotspot'
  | '25-way-1-hotspot' | '64-way-1-hotspot';

interface GridLayout {
  cols: number;
  rows: number;
  label: string;
  isHotspot?: boolean;
  hotspotCols?: number;
  hotspotRows?: number;
  totalCameras?: number;
}

const GRID_LAYOUTS: Record<GridLayoutType, GridLayout> = {
  // Standard layouts
  '2x2': { cols: 2, rows: 2, label: '2×2', totalCameras: 4 },
  '2x3': { cols: 2, rows: 3, label: '2×3', totalCameras: 6 },
  '3x3': { cols: 3, rows: 3, label: '3×3', totalCameras: 9 },
  '3x4': { cols: 3, rows: 4, label: '3×4', totalCameras: 12 },

  // Hotspot layouts
  '9-way-1-hotspot': {
    cols: 3, rows: 3, label: '9-Way-1-Hotspot',
    isHotspot: true, hotspotCols: 2, hotspotRows: 2, totalCameras: 6
  },
  '12-way-1-hotspot': {
    cols: 3, rows: 4, label: '12-Way-1-Hotspot',
    isHotspot: true, hotspotCols: 2, hotspotRows: 3, totalCameras: 7
  },
  '16-way-1-hotspot': {
    cols: 4, rows: 4, label: '16-Way-1-Hotspot',
    isHotspot: true, hotspotCols: 3, hotspotRows: 3, totalCameras: 8
  },
  '25-way-1-hotspot': {
    cols: 5, rows: 5, label: '25-Way-1-Hotspot',
    isHotspot: true, hotspotCols: 4, hotspotRows: 4, totalCameras: 10
  },
  '64-way-1-hotspot': {
    cols: 8, rows: 8, label: '64-Way-1-Hotspot',
    isHotspot: true, hotspotCols: 7, hotspotRows: 7, totalCameras: 16
  },
};

interface GridCell {
  camera: Camera | null;
  loading: boolean;
  isHotspot?: boolean;
  gridArea?: string;
  playbackState?: {
    mode: 'live' | 'playback';
    isPlaying: boolean;
    currentTime: Date;
    startTime: Date;
    endTime: Date;
    speed: number;
    zoomLevel: number;
    timelineData: TimelineData | null;
  };
}

// Helper function to build grid cell structure for hotspot layouts
function buildHotspotCells(layoutConfig: GridLayout): GridCell[] {
  const { cols, rows, isHotspot, hotspotCols, hotspotRows, totalCameras } = layoutConfig;

  if (!isHotspot || !hotspotCols || !hotspotRows || !totalCameras) {
    // Standard layout - simple grid
    return Array.from({ length: cols * rows }, () => ({
      camera: null,
      loading: false
    }));
  }

  // Hotspot layout
  const cells: GridCell[] = [];

  // Cell 0: Hotspot (spans hotspotCols × hotspotRows)
  cells.push({
    camera: null,
    loading: false,
    isHotspot: true,
    gridArea: `1 / 1 / ${hotspotRows + 1} / ${hotspotCols + 1}`
  });

  // Right column cells (rows 1 to hotspotRows)
  for (let row = 1; row <= hotspotRows; row++) {
    cells.push({
      camera: null,
      loading: false,
      gridArea: `${row} / ${cols} / ${row + 1} / ${cols + 1}`
    });
  }

  // Bottom row cells (all columns)
  for (let col = 1; col <= cols; col++) {
    cells.push({
      camera: null,
      loading: false,
      gridArea: `${rows} / ${col} / ${rows + 1} / ${col + 1}`
    });
  }

  return cells.slice(0, totalCameras);
}

// Initialize playback state for a cell
function initializePlaybackState() {
  return {
    mode: 'live' as const,
    isPlaying: false,
    currentTime: new Date(),
    startTime: new Date(Date.now() - 24 * 60 * 60 * 1000), // 24 hours ago
    endTime: new Date(),
    speed: 1.0,
    zoomLevel: 12, // 12 hours
    timelineData: null,
  };
}

export const StreamGridEnhanced = forwardRef<StreamGridEnhancedRef, StreamGridEnhancedProps>(
  function StreamGridEnhanced({ defaultLayout = '3x3', onLayoutChange }, ref) {
    const [layout, setLayout] = useState<GridLayoutType>(defaultLayout);
    const [fullscreenIndex, setFullscreenIndex] = useState<number | null>(null);
    const [dragOverIndex, setDragOverIndex] = useState<number | null>(null);
    const [showSaveDialog, setShowSaveDialog] = useState(false);
    const [showManageDialog, setShowManageDialog] = useState(false);
    const [showClearAllDialog, setShowClearAllDialog] = useState(false);
    const { toasts, removeToast, error: showError, warning: showWarning } = useToast();

    const layoutConfig = GRID_LAYOUTS[layout];
    const { cols, rows, isHotspot, totalCameras } = layoutConfig;
    const maxCells = totalCameras || cols * rows;

    // Initialize grid cells
    const [gridCells, setGridCells] = useState<GridCell[]>(
      buildHotspotCells(layoutConfig)
    );

    // Debounce timer for seek operations
    const seekDebounceTimers = useRef<Map<number, NodeJS.Timeout>>(new Map());

  // Update grid cells when layout changes
  React.useEffect(() => {
    setGridCells((prev) => {
      const newCells = buildHotspotCells(layoutConfig);
      // Preserve existing camera assignments where possible
      return newCells.map((cell, index) => ({
        ...cell,
        camera: prev[index]?.camera || null,
        loading: prev[index]?.loading || false,
      }));
    });
  }, [layout]);

  const handleLayoutChange = (newLayout: GridLayoutType) => {
    setLayout(newLayout);
    onLayoutChange?.(newLayout);
  };

  const handleDragOver = (event: React.DragEvent, index: number) => {
    event.preventDefault();
    event.stopPropagation();
    setDragOverIndex(index);
  };

  const handleDragLeave = (event: React.DragEvent) => {
    event.preventDefault();
    event.stopPropagation();
    setDragOverIndex(null);
  };

  const handleDrop = (event: React.DragEvent, index: number) => {
    event.preventDefault();
    event.stopPropagation();
    setDragOverIndex(null);

    try {
      const data = event.dataTransfer.getData('application/json');
      if (!data) return;

      const dragItem: DragItem = JSON.parse(data);

      if (dragItem.type === 'camera') {
        const camera = dragItem.data as Camera;
        assignCameraToCell(camera, index);
      } else if (dragItem.type === 'grid-cell') {
        // Handle drag between cells
        const sourceCellIndex = dragItem.data as number;
        swapCells(sourceCellIndex, index);
      }
    } catch (error) {
      console.error('Drop failed:', error);
    }
  };

  const assignCameraToCell = (camera: Camera, index: number) => {
    setGridCells((prev) => {
      const newCells = [...prev];
      // Preserve isHotspot and gridArea properties
      newCells[index] = {
        ...newCells[index],
        camera,
        loading: false
      };
      return newCells;
    });
  };

  const addCameraToNextAvailableCell = (camera: Camera): boolean => {
    const firstEmptyIndex = gridCells.findIndex((cell) => cell.camera === null);
    if (firstEmptyIndex === -1) {
      return false; // Grid is full
    }
    assignCameraToCell(camera, firstEmptyIndex);
    return true;
  };

  // Expose methods to parent component
  useImperativeHandle(ref, () => ({
    addCameraToNextAvailableCell,
  }));

  const removeCameraFromCell = (index: number) => {
    setGridCells((prev) => {
      const newCells = [...prev];
      // Preserve isHotspot and gridArea properties
      newCells[index] = {
        ...newCells[index],
        camera: null,
        loading: false
      };
      return newCells;
    });
  };

  const swapCells = (sourceIndex: number, targetIndex: number) => {
    if (sourceIndex === targetIndex) return;

    setGridCells((prev) => {
      const newCells = [...prev];
      const sourceCamera = newCells[sourceIndex].camera;
      const targetCamera = newCells[targetIndex].camera;

      // Swap cameras while preserving grid structure properties
      newCells[sourceIndex] = {
        ...newCells[sourceIndex],
        camera: targetCamera,
        loading: false
      };
      newCells[targetIndex] = {
        ...newCells[targetIndex],
        camera: sourceCamera,
        loading: false
      };

      return newCells;
    });
  };

  const handleCellDragStart = (event: React.DragEvent, index: number) => {
    const dragData: DragItem = {
      type: 'grid-cell',
      data: index
    };
    event.dataTransfer.setData('application/json', JSON.stringify(dragData));
    event.dataTransfer.effectAllowed = 'move';
  };

  // Playback mode handlers
  const handleModeChange = async (index: number, newMode: 'live' | 'playback') => {
    const cell = gridCells[index];
    if (!cell.camera) return;

    if (newMode === 'playback') {
      // Initialize playback state if not exists
      if (!cell.playbackState) {
        const initialState = initializePlaybackState();

        setGridCells((prev) => {
          const newCells = [...prev];
          newCells[index] = {
            ...newCells[index],
            playbackState: { ...initialState, mode: 'playback' },
          };
          return newCells;
        });

        // Query recordings
        try {
          const data = await api.getMilestoneSequences({
            cameraId: cell.camera.id,
            startTime: initialState.startTime.toISOString(),
            endTime: initialState.endTime.toISOString(),
          });

          // Transform Milestone sequences to timeline format
          const transformedData = {
            cameraId: cell.camera.id,
            queryRange: {
              start: initialState.startTime.toISOString(),
              end: initialState.endTime.toISOString(),
            },
            sequences: data.sequences.map((seq) => ({
              sequenceId: `${seq.timeBegin}-${seq.timeEnd}`,
              startTime: seq.timeBegin,
              endTime: seq.timeEnd,
              durationSeconds: (new Date(seq.timeEnd).getTime() - new Date(seq.timeBegin).getTime()) / 1000,
              available: true,
              sizeBytes: 0,
            })),
            gaps: [],
            totalRecordingSeconds: 0,
            totalGapSeconds: 0,
            coverage: 0,
          };

          setGridCells((prev) => {
            const newCells = [...prev];
            if (newCells[index].playbackState) {
              newCells[index].playbackState!.timelineData = transformedData;
            }
            return newCells;
          });
        } catch (error) {
          console.error('Failed to query recordings:', error);
          showError('Failed to load recordings');
        }
      } else {
        // Just switch mode
        setGridCells((prev) => {
          const newCells = [...prev];
          newCells[index].playbackState!.mode = 'playback';
          return newCells;
        });
      }
    } else {
      // Switch back to live
      setGridCells((prev) => {
        const newCells = [...prev];
        if (newCells[index].playbackState) {
          newCells[index].playbackState!.mode = 'live';
        }
        return newCells;
      });
    }
  };

  const handlePlayPause = (index: number) => {
    setGridCells((prev) => {
      const newCells = [...prev];
      if (newCells[index].playbackState) {
        newCells[index].playbackState!.isPlaying = !newCells[index].playbackState!.isPlaying;
      }
      return newCells;
    });
  };

  const handleSeek = useCallback((index: number, newTime: Date, immediate = false) => {
    const cell = gridCells[index];
    if (!cell.camera || !cell.playbackState) return;

    // Clear existing debounce timer for this cell
    const existingTimer = seekDebounceTimers.current.get(index);
    if (existingTimer) {
      clearTimeout(existingTimer);
    }

    // If immediate seek (e.g., from timeline click), update right away
    if (immediate) {
      setGridCells((prev) => {
        const newCells = [...prev];
        if (newCells[index].playbackState) {
          newCells[index].playbackState!.currentTime = newTime;
        }
        return newCells;
      });
      return;
    }

    // Otherwise, debounce the seek operation (500ms delay)
    const timer = setTimeout(() => {
      setGridCells((prev) => {
        const newCells = [...prev];
        if (newCells[index].playbackState) {
          newCells[index].playbackState!.currentTime = newTime;
        }
        return newCells;
      });
      seekDebounceTimers.current.delete(index);
    }, 500);

    seekDebounceTimers.current.set(index, timer);
  }, [gridCells]);

  const handleScrollTimeline = (index: number, direction: 'left' | 'right') => {
    const cell = gridCells[index];
    if (!cell.playbackState) return;

    const scrollAmount = cell.playbackState.zoomLevel * 60 * 60 * 1000; // zoom level in ms

    setGridCells((prev) => {
      const newCells = [...prev];
      if (newCells[index].playbackState) {
        const state = newCells[index].playbackState!;
        state.startTime = new Date(
          direction === 'left'
            ? state.startTime.getTime() - scrollAmount
            : state.startTime.getTime() + scrollAmount
        );
        state.endTime = new Date(
          direction === 'left'
            ? state.endTime.getTime() - scrollAmount
            : state.endTime.getTime() + scrollAmount
        );
      }
      return newCells;
    });
  };

  const handleZoomChange = (index: number, newZoomLevel: number) => {
    const cell = gridCells[index];
    if (!cell.playbackState) return;

    const center = cell.playbackState.currentTime.getTime();
    const halfDuration = (newZoomLevel * 60 * 60 * 1000) / 2;

    setGridCells((prev) => {
      const newCells = [...prev];
      if (newCells[index].playbackState) {
        newCells[index].playbackState!.zoomLevel = newZoomLevel;
        newCells[index].playbackState!.startTime = new Date(center - halfDuration);
        newCells[index].playbackState!.endTime = new Date(center + halfDuration);
      }
      return newCells;
    });
  };

  const hasRecordingAtCurrentTime = (index: number): boolean => {
    const cell = gridCells[index];
    if (!cell.playbackState || !cell.playbackState.timelineData) return false;

    const currentTimestamp = cell.playbackState.currentTime.getTime();

    return cell.playbackState.timelineData.sequences.some((seq) => {
      const start = new Date(seq.startTime).getTime();
      const end = new Date(seq.endTime).getTime();
      return currentTimestamp >= start && currentTimestamp <= end;
    });
  };

  const handleClearAllClick = () => {
    setShowClearAllDialog(true);
  };

  const handleConfirmClearAll = () => {
    setGridCells((prev) =>
      prev.map((cell) => ({
        ...cell,
        camera: null,
        loading: false
      }))
    );
    setShowClearAllDialog(false);
  };

  const activeCameras = gridCells.filter((cell) => cell.camera !== null).length;

  // Layout preference helpers
  const getCurrentLayoutType = (): LayoutType => {
    return isHotspot ? 'hotspot' : 'standard';
  };

  const getCurrentCameraAssignments = () => {
    return gridCells
      .map((cell, index) => ({
        camera_id: cell.camera?.id || '',
        position_index: index,
      }))
      .filter((assignment) => assignment.camera_id !== '');
  };

  const detectGridLayout = (layoutType: LayoutType, cameraCount: number, maxPositionIndex: number): GridLayoutType => {
    const totalCells = maxPositionIndex + 1;

    if (layoutType === 'hotspot') {
      // Hotspot layouts - match by camera count
      if (cameraCount <= 6 && totalCells <= 9) return '9-way-1-hotspot';
      if (cameraCount <= 7 && totalCells <= 12) return '12-way-1-hotspot';
      if (cameraCount <= 8 && totalCells <= 16) return '16-way-1-hotspot';
      if (cameraCount <= 10 && totalCells <= 25) return '25-way-1-hotspot';
      if (cameraCount <= 16 && totalCells <= 64) return '64-way-1-hotspot';
      return '9-way-1-hotspot'; // Default hotspot
    } else {
      // Standard layouts - match by total cells needed
      if (totalCells <= 4) return '2x2';
      if (totalCells <= 6) return '2x3';
      if (totalCells <= 9) return '3x3';
      if (totalCells <= 12) return '3x4';
      return '3x4'; // Default standard
    }
  };

  const handleLoadLayout = async (layoutSummary: LayoutPreferenceSummary) => {
    try {
      // Fetch full layout details
      const fullLayout = await api.getLayout(layoutSummary.id);

      if (!fullLayout.cameras || fullLayout.cameras.length === 0) {
        showWarning('This layout has no cameras saved.');
        return;
      }

      // Use the saved grid_layout directly
      const savedGridLayout = fullLayout.grid_layout as GridLayoutType;

      // Switch to the saved grid layout
      if (layout !== savedGridLayout) {
        handleLayoutChange(savedGridLayout);
        // Wait a bit for the layout to change
        await new Promise(resolve => setTimeout(resolve, 100));
      }

      // Clear current grid
      setGridCells((prev) =>
        prev.map((cell) => ({
          ...cell,
          camera: null,
          loading: false,
        }))
      );

      // Set loading state for cells that will receive cameras
      setGridCells((prev) => {
        const newCells = [...prev];
        fullLayout.cameras?.forEach((assignment) => {
          if (assignment.position_index < newCells.length) {
            newCells[assignment.position_index] = {
              ...newCells[assignment.position_index],
              loading: true,
            };
          }
        });
        return newCells;
      });

      // Fetch and assign each camera
      for (const assignment of fullLayout.cameras) {
        try {
          // Fetch full camera details by ID
          const camera = await api.getCamera(assignment.camera_id);

          // Assign camera to the correct position in the grid
          assignCameraToCell(camera, assignment.position_index);
        } catch (error) {
          console.error(`Failed to load camera ${assignment.camera_id}:`, error);
          // Clear loading state for this cell
          setGridCells((prev) => {
            const newCells = [...prev];
            if (assignment.position_index < newCells.length) {
              newCells[assignment.position_index] = {
                ...newCells[assignment.position_index],
                loading: false,
              };
            }
            return newCells;
          });
          // Continue loading other cameras even if one fails
        }
      }
    } catch (error) {
      console.error('Failed to load layout:', error);
      showError('Failed to load layout. Please try again.');
    }
  };

  // Fullscreen view
  if (fullscreenIndex !== null && gridCells[fullscreenIndex]?.camera) {
    return (
      <div className="fixed inset-0 z-50 bg-black">
        <div className="relative h-full">
          <LiveStreamPlayer
            key={gridCells[fullscreenIndex].camera!.id}
            camera={gridCells[fullscreenIndex].camera!}
          />
          <button
            onClick={() => setFullscreenIndex(null)}
            className="absolute top-4 right-4 p-2 bg-black/70 hover:bg-black/90 text-white rounded-lg transition-colors z-10"
            title="Exit fullscreen"
          >
            <X className="w-6 h-6" />
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-full bg-gray-100">
      {/* Toolbar */}
      <div className="bg-white border-b border-gray-200 px-4 py-3">
        <div className="flex items-start gap-4 flex-wrap">
          <div className="flex items-center gap-2 pt-6">
            <Grid className="w-5 h-5 text-gray-600" />
            <span className="font-medium text-gray-700">Layout:</span>
          </div>

          {/* Standard Layouts */}
          <div className="flex flex-col gap-1">
            <span className="text-xs text-gray-500 font-medium">Standard</span>
            <div className="flex gap-2 flex-wrap">
              {(Object.keys(GRID_LAYOUTS) as GridLayoutType[])
                .filter((key) => !GRID_LAYOUTS[key].isHotspot)
                .map((key) => (
                  <button
                    key={key}
                    onClick={() => handleLayoutChange(key)}
                    className={cn(
                      'px-3 py-1.5 rounded-lg text-sm font-medium transition-colors',
                      layout === key
                        ? 'bg-blue-600 text-white shadow-sm'
                        : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                    )}
                  >
                    {GRID_LAYOUTS[key].label}
                  </button>
                ))}
            </div>
          </div>

          {/* Hotspot Layouts */}
          <div className="flex flex-col gap-1">
            <span className="text-xs text-gray-500 font-medium">Hotspot</span>
            <div className="flex gap-2 flex-wrap">
              {(Object.keys(GRID_LAYOUTS) as GridLayoutType[])
                .filter((key) => GRID_LAYOUTS[key].isHotspot)
                .map((key) => (
                  <button
                    key={key}
                    onClick={() => handleLayoutChange(key)}
                    className={cn(
                      'px-3 py-1.5 rounded-lg text-sm font-medium transition-colors border',
                      layout === key
                        ? 'bg-amber-500 text-white shadow-sm border-amber-600'
                        : 'bg-amber-50 text-amber-800 hover:bg-amber-100 border-amber-200'
                    )}
                  >
                    {GRID_LAYOUTS[key].label}
                  </button>
                ))}
            </div>
          </div>

          <div className="ml-auto flex items-center gap-3">
            <div className="text-sm text-gray-600">
              <span className="font-semibold text-gray-900">{activeCameras}</span> /{' '}
              {maxCells} cells
            </div>

            {/* Layout Management Buttons */}
            <div className="flex items-center gap-2 border-l border-gray-300 pl-3">
              <LoadLayoutDropdown
                onLayoutSelect={handleLoadLayout}
                currentLayoutType={getCurrentLayoutType()}
              />

              <button
                onClick={() => setShowSaveDialog(true)}
                disabled={activeCameras === 0}
                className={cn(
                  'inline-flex items-center gap-2 px-3 h-10 rounded-md border text-sm font-medium transition-colors',
                  activeCameras > 0
                    ? 'border-blue-600 bg-blue-600 text-white hover:bg-blue-700'
                    : 'border-gray-300 bg-gray-100 text-gray-400 cursor-not-allowed'
                )}
                title={activeCameras === 0 ? 'Add cameras to save layout' : 'Save current layout'}
              >
                <Save className="w-4 h-4" />
                Save Layout
              </button>

              <button
                onClick={() => setShowManageDialog(true)}
                className="inline-flex items-center gap-2 px-3 h-10 rounded-md border border-gray-300 bg-white text-gray-700 text-sm font-medium hover:bg-gray-50 transition-colors"
                title="Manage saved layouts"
              >
                <Settings className="w-4 h-4" />
                Manage
              </button>
            </div>

            {activeCameras > 0 && (
              <button
                onClick={handleClearAllClick}
                className="flex items-center gap-2 px-3 py-1.5 text-sm bg-red-50 text-red-700 rounded-lg hover:bg-red-100 transition-colors"
              >
                <Trash2 className="w-4 h-4" />
                Clear All
              </button>
            )}
          </div>
        </div>
      </div>

      {/* Grid */}
      <div className="flex-1 p-4 overflow-auto">
        <div
          className="grid gap-3 h-full"
          style={{
            gridTemplateColumns: `repeat(${cols}, minmax(0, 1fr))`,
            gridTemplateRows: `repeat(${rows}, minmax(0, 1fr))`,
          }}
        >
          {gridCells.map((cell, index) => (
            <div
              key={cell.camera?.id || `empty-${index}`}
              draggable={!!cell.camera}
              onDragStart={(e) => cell.camera && handleCellDragStart(e, index)}
              onDragOver={(e) => handleDragOver(e, index)}
              onDragLeave={handleDragLeave}
              onDrop={(e) => handleDrop(e, index)}
              style={cell.gridArea ? { gridArea: cell.gridArea } : undefined}
              className={cn(
                'relative rounded-lg shadow-md group transition-all',
                // Use overflow-visible to allow dropdowns to show, video players handle their own overflow
                cell.camera ? 'overflow-visible' : 'overflow-hidden',
                cell.camera
                  ? 'bg-gray-900 cursor-move'
                  : 'bg-gray-200 border-2 border-dashed border-gray-300',
                dragOverIndex === index && 'ring-4 ring-blue-500 ring-opacity-50 scale-105',
                cell.isHotspot && 'ring-2 ring-amber-500'
              )}
            >
              {cell.camera ? (
                <>
                  {/* Video Player - Live or Playback */}
                  {cell.playbackState?.mode === 'playback' ? (
                    <RecordingPlayer
                      key={cell.camera.id}
                      cameraId={cell.camera.id}
                      startTime={cell.playbackState.startTime}
                      endTime={cell.playbackState.endTime}
                      initialPlaybackTime={cell.playbackState.currentTime}
                      onPlaybackTimeChange={(time) => handleSeek(index, time)}
                      onPlaybackStateChange={(state) => {
                        if (state === 'playing' && !cell.playbackState!.isPlaying) {
                          handlePlayPause(index);
                        } else if (state === 'paused' && cell.playbackState!.isPlaying) {
                          handlePlayPause(index);
                        }
                      }}
                      showControls={false}
                      className="absolute inset-0"
                    />
                  ) : (
                    <LiveStreamPlayer
                      key={cell.camera.id}
                      camera={cell.camera}
                      quality="medium"
                    />
                  )}

                  {/* Camera info overlay with mode toggle */}
                  <div className="absolute top-0 left-0 right-0 bg-gradient-to-b from-black/70 to-transparent p-3 opacity-0 group-hover:opacity-100 transition-opacity z-10">
                    <div className="flex items-center justify-between gap-2">
                      <div className="flex-1 min-w-0">
                        <p className="text-white text-sm font-medium truncate">
                          {cell.camera.name}
                        </p>
                        {cell.camera.name_ar && (
                          <p className="text-white/80 text-xs truncate">
                            {cell.camera.name_ar}
                          </p>
                        )}
                      </div>
                      {/* Mode Toggle - only show if camera has milestone_device_id */}
                      {cell.camera.milestone_device_id && (
                        <PlaybackModeToggle
                          mode={cell.playbackState?.mode || 'live'}
                          onChange={(mode) => handleModeChange(index, mode)}
                          className="flex-shrink-0"
                        />
                      )}
                    </div>
                  </div>

                  {/* Action buttons */}
                  <div className="absolute top-2 right-2 flex gap-2 opacity-0 group-hover:opacity-100 transition-opacity z-10">
                    <button
                      onClick={() => setFullscreenIndex(index)}
                      className="p-2 bg-black/70 hover:bg-black/90 text-white rounded-lg transition-colors"
                      title="Fullscreen"
                    >
                      <Maximize2 className="w-4 h-4" />
                    </button>
                    <button
                      onClick={() => removeCameraFromCell(index)}
                      className="p-2 bg-black/70 hover:bg-red-600 text-white rounded-lg transition-colors"
                      title="Remove camera"
                    >
                      <X className="w-4 h-4" />
                    </button>
                  </div>

                  {/* Playback Controls */}
                  {cell.playbackState?.mode === 'playback' &&
                   cell.playbackState.timelineData && (
                    <div className="absolute bottom-0 left-0 right-0 z-30">
                      <PlaybackControlBar
                        startTime={cell.playbackState.startTime}
                        endTime={cell.playbackState.endTime}
                        currentTime={cell.playbackState.currentTime}
                        sequences={cell.playbackState.timelineData.sequences}
                        isPlaying={cell.playbackState.isPlaying}
                        zoomLevel={cell.playbackState.zoomLevel}
                        onPlayPause={() => handlePlayPause(index)}
                        onSeek={(time) => handleSeek(index, time, true)}
                        onScrollTimeline={(direction) => handleScrollTimeline(index, direction)}
                        onZoomChange={(zoom) => handleZoomChange(index, zoom)}
                        hasRecording={hasRecordingAtCurrentTime(index)}
                      />
                    </div>
                  )}

                  {/* Cell number badge - only show in live mode */}
                  {cell.playbackState?.mode !== 'playback' && (
                    <div className="absolute bottom-2 left-2 bg-black/70 text-white text-xs px-2 py-1 rounded opacity-0 group-hover:opacity-100 transition-opacity">
                      {cell.isHotspot ? 'Hotspot' : `Cell ${index + 1}`}
                    </div>
                  )}

                  {/* Hotspot indicator */}
                  {cell.isHotspot && (
                    <div className="absolute top-2 left-2 bg-amber-500 text-white text-xs px-2 py-1 rounded-md font-semibold shadow-lg">
                      HOTSPOT
                    </div>
                  )}
                </>
              ) : (
                /* Empty cell placeholder */
                <div className="absolute inset-0 flex flex-col items-center justify-center text-gray-400 p-4">
                  {cell.isHotspot && (
                    <div className="absolute top-2 left-2 bg-amber-500 text-white text-xs px-2 py-1 rounded-md font-semibold shadow-lg">
                      HOTSPOT
                    </div>
                  )}
                  <div className="text-center">
                    <Plus className={cn(
                      "mx-auto mb-2 opacity-40",
                      cell.isHotspot ? "w-16 h-16" : "w-10 h-10"
                    )} />
                    <p className="text-sm font-medium">Drop camera here</p>
                    <p className="text-xs mt-1 opacity-70">or double-click in sidebar</p>
                  </div>
                  <div className="mt-4 text-xs bg-white/50 px-3 py-1 rounded-full">
                    {cell.isHotspot ? 'Hotspot' : `Cell ${index + 1}`}
                  </div>
                </div>
              )}
            </div>
          ))}
        </div>
      </div>

      {/* Toast Notifications */}
      <ToastContainer toasts={toasts} onRemove={removeToast} />

      {/* Clear All Confirmation Dialog */}
      <Dialog open={showClearAllDialog} onOpenChange={setShowClearAllDialog}>
        <DialogContent onClose={() => setShowClearAllDialog(false)}>
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2 text-red-600">
              <AlertTriangle className="w-5 h-5" />
              Clear All Cameras?
            </DialogTitle>
            <DialogDescription>
              This will remove all cameras from the grid. This action cannot be undone.
            </DialogDescription>
          </DialogHeader>

          <DialogBody>
            <div className="space-y-4">
              <div className="bg-red-50 border border-red-200 rounded-lg p-4">
                <p className="text-sm text-red-800">
                  <strong>Warning:</strong> You are about to clear <strong>{activeCameras} camera{activeCameras !== 1 ? 's' : ''}</strong> from the grid.
                </p>
              </div>
              <p className="text-sm text-gray-600">
                The cameras will remain in your camera list and can be added back to the grid at any time.
              </p>
            </div>
          </DialogBody>

          <DialogFooter>
            <Button
              variant="secondary"
              onClick={() => setShowClearAllDialog(false)}
            >
              Cancel
            </Button>
            <Button
              variant="danger"
              onClick={handleConfirmClearAll}
            >
              Clear All Cameras
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Layout Dialogs */}
      <SaveLayoutDialog
        open={showSaveDialog}
        onOpenChange={setShowSaveDialog}
        layoutType={getCurrentLayoutType()}
        gridLayout={layout}
        cameras={getCurrentCameraAssignments()}
        onSuccess={() => {
          // Optionally reload or show success message
          console.log('Layout saved successfully');
        }}
      />

      <LayoutManagerDialog
        open={showManageDialog}
        onOpenChange={setShowManageDialog}
        onLayoutUpdate={() => {
          // Optionally refresh something
          console.log('Layout updated');
        }}
      />
    </div>
  );
});
