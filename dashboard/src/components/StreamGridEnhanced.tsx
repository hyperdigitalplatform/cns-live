import React, { useState, forwardRef, useImperativeHandle } from 'react';
import { LiveStreamPlayer } from './LiveStreamPlayer';
import type { Camera, DragItem } from '@/types';
import { Grid, Maximize2, X, Plus, Trash2 } from 'lucide-react';
import { cn } from '@/utils/cn';

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

export const StreamGridEnhanced = forwardRef<StreamGridEnhancedRef, StreamGridEnhancedProps>(
  function StreamGridEnhanced({ defaultLayout = '3x3', onLayoutChange }, ref) {
    const [layout, setLayout] = useState<GridLayoutType>(defaultLayout);
    const [fullscreenIndex, setFullscreenIndex] = useState<number | null>(null);
    const [dragOverIndex, setDragOverIndex] = useState<number | null>(null);

    const layoutConfig = GRID_LAYOUTS[layout];
    const { cols, rows, isHotspot, totalCameras } = layoutConfig;
    const maxCells = totalCameras || cols * rows;

    // Initialize grid cells
    const [gridCells, setGridCells] = useState<GridCell[]>(
      buildHotspotCells(layoutConfig)
    );

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

  const clearAllCells = () => {
    if (confirm('Clear all cameras from grid?')) {
      setGridCells((prev) =>
        prev.map((cell) => ({
          ...cell,
          camera: null,
          loading: false
        }))
      );
    }
  };

  const activeCameras = gridCells.filter((cell) => cell.camera !== null).length;

  // Fullscreen view
  if (fullscreenIndex !== null && gridCells[fullscreenIndex]?.camera) {
    return (
      <div className="fixed inset-0 z-50 bg-black">
        <div className="relative h-full">
          <LiveStreamPlayer camera={gridCells[fullscreenIndex].camera!} />
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

          <div className="ml-auto flex items-center gap-4">
            <div className="text-sm text-gray-600">
              <span className="font-semibold text-gray-900">{activeCameras}</span> /{' '}
              {maxCells} cells
            </div>
            {activeCameras > 0 && (
              <button
                onClick={clearAllCells}
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
              key={index}
              draggable={!!cell.camera}
              onDragStart={(e) => cell.camera && handleCellDragStart(e, index)}
              onDragOver={(e) => handleDragOver(e, index)}
              onDragLeave={handleDragLeave}
              onDrop={(e) => handleDrop(e, index)}
              style={cell.gridArea ? { gridArea: cell.gridArea } : undefined}
              className={cn(
                'relative rounded-lg overflow-hidden shadow-md group transition-all',
                cell.camera
                  ? 'bg-gray-900 cursor-move'
                  : 'bg-gray-200 border-2 border-dashed border-gray-300',
                dragOverIndex === index && 'ring-4 ring-blue-500 ring-opacity-50 scale-105',
                cell.isHotspot && 'ring-2 ring-amber-500'
              )}
            >
              {cell.camera ? (
                <>
                  {/* Camera stream */}
                  <LiveStreamPlayer camera={cell.camera} quality="medium" />

                  {/* Camera info overlay */}
                  <div className="absolute top-0 left-0 right-0 bg-gradient-to-b from-black/70 to-transparent p-3 opacity-0 group-hover:opacity-100 transition-opacity">
                    <p className="text-white text-sm font-medium truncate">
                      {cell.camera.name}
                    </p>
                    {cell.camera.name_ar && (
                      <p className="text-white/80 text-xs truncate">
                        {cell.camera.name_ar}
                      </p>
                    )}
                  </div>

                  {/* Action buttons */}
                  <div className="absolute top-2 right-2 flex gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
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

                  {/* Cell number badge */}
                  <div className="absolute bottom-2 left-2 bg-black/70 text-white text-xs px-2 py-1 rounded opacity-0 group-hover:opacity-100 transition-opacity">
                    {cell.isHotspot ? 'Hotspot' : `Cell ${index + 1}`}
                  </div>

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

      {/* Help text */}
      {activeCameras === 0 && (
        <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
          <div className="bg-white/95 rounded-xl shadow-xl p-8 max-w-md text-center">
            <Grid className="w-16 h-16 mx-auto mb-4 text-blue-600" />
            <h3 className="text-lg font-semibold text-gray-900 mb-2">
              No Cameras in Grid
            </h3>
            <p className="text-gray-600 text-sm mb-4">
              Drag cameras from the sidebar and drop them into grid cells, or
              double-click a camera in the sidebar to auto-assign it to the next
              available cell.
            </p>
            <div className="flex items-center justify-center gap-4 text-xs text-gray-500">
              <div className="flex items-center gap-2">
                <div className="w-3 h-3 bg-blue-500 rounded" />
                <span>Drag & Drop</span>
              </div>
              <div className="flex items-center gap-2">
                <div className="w-3 h-3 bg-green-500 rounded" />
                <span>Double-click</span>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
});
