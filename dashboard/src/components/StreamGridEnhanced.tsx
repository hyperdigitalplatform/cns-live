import React, { useState } from 'react';
import { LiveStreamPlayer } from './LiveStreamPlayer';
import type { Camera, DragItem } from '@/types';
import { Grid, Maximize2, X, Plus, Trash2 } from 'lucide-react';
import { cn } from '@/utils/cn';

interface StreamGridEnhancedProps {
  defaultLayout?: GridLayoutType;
  onLayoutChange?: (layout: GridLayoutType) => void;
}

type GridLayoutType = '1x1' | '2x2' | '3x3' | '4x4' | '2x3' | '3x4' | '4x5' | '5x5' | '6x6';

const GRID_LAYOUTS: Record<
  GridLayoutType,
  { cols: number; rows: number; label: string }
> = {
  '1x1': { cols: 1, rows: 1, label: '1×1' },
  '2x2': { cols: 2, rows: 2, label: '2×2' },
  '3x3': { cols: 3, rows: 3, label: '3×3' },
  '4x4': { cols: 4, rows: 4, label: '4×4' },
  '2x3': { cols: 2, rows: 3, label: '2×3' },
  '3x4': { cols: 3, rows: 4, label: '3×4' },
  '4x5': { cols: 4, rows: 5, label: '4×5' },
  '5x5': { cols: 5, rows: 5, label: '5×5' },
  '6x6': { cols: 6, rows: 6, label: '6×6' },
};

interface GridCell {
  camera: Camera | null;
  loading: boolean;
}

export function StreamGridEnhanced({
  defaultLayout = '3x3',
  onLayoutChange,
}: StreamGridEnhancedProps) {
  const [layout, setLayout] = useState<GridLayoutType>(defaultLayout);
  const [fullscreenIndex, setFullscreenIndex] = useState<number | null>(null);
  const [dragOverIndex, setDragOverIndex] = useState<number | null>(null);

  const { cols, rows } = GRID_LAYOUTS[layout];
  const maxCells = cols * rows;

  // Initialize grid cells
  const [gridCells, setGridCells] = useState<GridCell[]>(
    Array.from({ length: maxCells }, () => ({ camera: null, loading: false }))
  );

  // Update grid cells when layout changes
  React.useEffect(() => {
    setGridCells((prev) => {
      const newCells = Array.from({ length: maxCells }, (_, index) =>
        prev[index] || { camera: null, loading: false }
      );
      return newCells;
    });
  }, [maxCells]);

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
      }
    } catch (error) {
      console.error('Drop failed:', error);
    }
  };

  const assignCameraToCell = (camera: Camera, index: number) => {
    setGridCells((prev) => {
      const newCells = [...prev];
      newCells[index] = { camera, loading: false };
      return newCells;
    });
  };

  const removeCameraFromCell = (index: number) => {
    setGridCells((prev) => {
      const newCells = [...prev];
      newCells[index] = { camera: null, loading: false };
      return newCells;
    });
  };

  const clearAllCells = () => {
    if (confirm('Clear all cameras from grid?')) {
      setGridCells(
        Array.from({ length: maxCells }, () => ({ camera: null, loading: false }))
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
        <div className="flex items-center gap-4 flex-wrap">
          <div className="flex items-center gap-2">
            <Grid className="w-5 h-5 text-gray-600" />
            <span className="font-medium text-gray-700">Layout:</span>
          </div>
          <div className="flex gap-2 flex-wrap">
            {(Object.keys(GRID_LAYOUTS) as GridLayoutType[]).map((key) => (
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
              onDragOver={(e) => handleDragOver(e, index)}
              onDragLeave={handleDragLeave}
              onDrop={(e) => handleDrop(e, index)}
              className={cn(
                'relative rounded-lg overflow-hidden shadow-md group transition-all',
                cell.camera
                  ? 'bg-gray-900'
                  : 'bg-gray-200 border-2 border-dashed border-gray-300',
                dragOverIndex === index && 'ring-4 ring-blue-500 ring-opacity-50 scale-105'
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
                    Cell {index + 1}
                  </div>
                </>
              ) : (
                /* Empty cell placeholder */
                <div className="absolute inset-0 flex flex-col items-center justify-center text-gray-400 p-4">
                  <div className="text-center">
                    <Plus className="w-10 h-10 mx-auto mb-2 opacity-40" />
                    <p className="text-sm font-medium">Drop camera here</p>
                    <p className="text-xs mt-1 opacity-70">or double-click in sidebar</p>
                  </div>
                  <div className="mt-4 text-xs bg-white/50 px-3 py-1 rounded-full">
                    Cell {index + 1}
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
}
