import React, { useState } from 'react';
import { LiveStreamPlayer } from './LiveStreamPlayer';
import type { Camera } from '@/types';
import { Grid, Maximize2, X } from 'lucide-react';
import { cn } from '@/utils/cn';

interface StreamGridProps {
  cameras: Camera[];
  defaultLayout?: GridLayoutType;
}

type GridLayoutType = '1x1' | '2x2' | '3x3' | '4x4' | '2x3' | '3x4';

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
};

export function StreamGrid({
  cameras,
  defaultLayout = '2x2',
}: StreamGridProps) {
  const [layout, setLayout] = useState<GridLayoutType>(defaultLayout);
  const [fullscreenIndex, setFullscreenIndex] = useState<number | null>(null);

  const { cols, rows } = GRID_LAYOUTS[layout];
  const maxCells = cols * rows;
  const visibleCameras = cameras.slice(0, maxCells);

  if (fullscreenIndex !== null && cameras[fullscreenIndex]) {
    return (
      <div className="fixed inset-0 z-50 bg-black">
        <div className="relative h-full">
          <LiveStreamPlayer camera={cameras[fullscreenIndex]} />
          <button
            onClick={() => setFullscreenIndex(null)}
            className="absolute top-4 right-4 p-2 bg-black/70 hover:bg-black/90 text-white rounded-lg transition-colors"
            title="Exit fullscreen"
          >
            <X className="w-6 h-6" />
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-full bg-gray-100 dark:bg-dark-base">
      {/* Toolbar */}
      <div className="bg-white dark:bg-dark-secondary border-b border-gray-200 dark:border-dark-border px-4 py-3 flex items-center gap-4">
        <div className="flex items-center gap-2">
          <Grid className="w-5 h-5 text-gray-600 dark:text-text-secondary" />
          <span className="font-medium text-gray-700 dark:text-text-primary">Layout:</span>
        </div>
        <div className="flex gap-2">
          {(Object.keys(GRID_LAYOUTS) as GridLayoutType[]).map((key) => (
            <button
              key={key}
              onClick={() => setLayout(key)}
              className={cn(
                'px-3 py-1 rounded text-sm font-medium transition-colors',
                layout === key
                  ? 'bg-primary-600 text-white'
                  : 'bg-gray-100 dark:bg-dark-surface text-gray-700 dark:text-text-secondary hover:bg-gray-200 dark:hover:bg-dark-elevated'
              )}
            >
              {GRID_LAYOUTS[key].label}
            </button>
          ))}
        </div>
        <div className="ml-auto text-sm text-gray-600 dark:text-text-secondary">
          {visibleCameras.length} of {cameras.length} cameras
        </div>
      </div>

      {/* Grid */}
      <div className="flex-1 p-4 overflow-auto">
        <div
          className="grid gap-4 h-full"
          style={{
            gridTemplateColumns: `repeat(${cols}, minmax(0, 1fr))`,
            gridTemplateRows: `repeat(${rows}, minmax(0, 1fr))`,
          }}
        >
          {visibleCameras.map((camera, index) => (
            <div
              key={camera.id}
              className="relative bg-gray-900 rounded-lg overflow-hidden shadow-lg group"
            >
              <LiveStreamPlayer camera={camera} quality="medium" />

              {/* Fullscreen button */}
              <button
                onClick={() => setFullscreenIndex(index)}
                className="absolute top-2 right-2 p-2 bg-black/70 hover:bg-black/90 text-white rounded-lg opacity-0 group-hover:opacity-100 transition-opacity"
                title="Fullscreen"
              >
                <Maximize2 className="w-4 h-4" />
              </button>
            </div>
          ))}

          {/* Empty cells */}
          {Array.from({ length: maxCells - visibleCameras.length }).map(
            (_, index) => (
              <div
                key={`empty-${index}`}
                className="bg-gray-200 dark:bg-dark-elevated rounded-lg flex items-center justify-center text-gray-400 dark:text-text-muted"
              >
                <div className="text-center">
                  <Grid className="w-12 h-12 mx-auto mb-2 opacity-30" />
                  <p className="text-sm">Empty slot</p>
                </div>
              </div>
            )
          )}
        </div>
      </div>
    </div>
  );
}
