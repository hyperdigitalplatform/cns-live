import React, { useState } from 'react';
import {
  ChevronUp,
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  ZoomIn,
  ZoomOut,
  Home,
  X,
} from 'lucide-react';
import { api } from '@/services/api';
import { cn } from '@/utils/cn';
import type { Camera } from '@/types';

interface PTZControlsProps {
  camera: Camera;
  onClose: () => void;
  isPinned: boolean;
}

export function PTZControls({ camera, onClose, isPinned }: PTZControlsProps) {
  const [activeButton, setActiveButton] = useState<string | null>(null);

  const handlePTZCommand = async (
    command: string,
    params?: { speed?: number; preset_id?: number }
  ) => {
    try {
      await api.controlPTZ(camera.id, command, params);
    } catch (error) {
      console.error('PTZ command failed:', error);
    }
  };

  const handleMouseDown = (command: string) => {
    setActiveButton(command);
    handlePTZCommand(command, { speed: 0.5 });
  };

  const handleMouseUp = () => {
    setActiveButton(null);
    // Send stop command if needed
  };

  if (!camera.ptz_enabled) {
    return (
      <div className="absolute inset-0 flex items-center justify-center bg-black/60 backdrop-blur-sm">
        <div className="bg-white rounded-lg p-6 text-center">
          <p className="text-gray-700">PTZ not available for this camera</p>
          <button
            onClick={onClose}
            className="mt-4 px-4 py-2 bg-gray-200 hover:bg-gray-300 rounded-lg text-sm"
          >
            Close
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
      <div className="pointer-events-auto">
        {/* Close/Back button (top-left when pinned) */}
        {isPinned && (
          <button
            onClick={onClose}
            className="absolute top-4 left-4 p-2 bg-black/70 hover:bg-black/90 text-white rounded-lg transition-colors"
            title="Close PTZ Controls"
          >
            <X className="w-5 h-5" />
          </button>
        )}

        {/* PTZ Control Panel */}
        <div className="bg-black/80 backdrop-blur-md rounded-2xl p-6 shadow-2xl border border-white/20">
          <div className="flex flex-col gap-4">
            {/* Camera name */}
            <div className="text-center text-white text-sm font-medium mb-2">
              PTZ Controls - {camera.name}
            </div>

            {/* Directional Pad */}
            <div className="grid grid-cols-3 gap-2">
              {/* Top row */}
              <div />
              <button
                onMouseDown={() => handleMouseDown('tilt_up')}
                onMouseUp={handleMouseUp}
                onMouseLeave={handleMouseUp}
                className={cn(
                  'p-4 bg-white/10 hover:bg-white/20 rounded-lg transition-colors active:bg-white/30',
                  activeButton === 'tilt_up' && 'bg-white/30'
                )}
                title="Tilt Up"
              >
                <ChevronUp className="w-6 h-6 text-white" />
              </button>
              <div />

              {/* Middle row */}
              <button
                onMouseDown={() => handleMouseDown('pan_left')}
                onMouseUp={handleMouseUp}
                onMouseLeave={handleMouseUp}
                className={cn(
                  'p-4 bg-white/10 hover:bg-white/20 rounded-lg transition-colors active:bg-white/30',
                  activeButton === 'pan_left' && 'bg-white/30'
                )}
                title="Pan Left"
              >
                <ChevronLeft className="w-6 h-6 text-white" />
              </button>
              <button
                onClick={() => handlePTZCommand('home')}
                className="p-4 bg-primary-600/80 hover:bg-primary-600 rounded-lg transition-colors"
                title="Home Position"
              >
                <Home className="w-6 h-6 text-white" />
              </button>
              <button
                onMouseDown={() => handleMouseDown('pan_right')}
                onMouseUp={handleMouseUp}
                onMouseLeave={handleMouseUp}
                className={cn(
                  'p-4 bg-white/10 hover:bg-white/20 rounded-lg transition-colors active:bg-white/30',
                  activeButton === 'pan_right' && 'bg-white/30'
                )}
                title="Pan Right"
              >
                <ChevronRight className="w-6 h-6 text-white" />
              </button>

              {/* Bottom row */}
              <div />
              <button
                onMouseDown={() => handleMouseDown('tilt_down')}
                onMouseUp={handleMouseUp}
                onMouseLeave={handleMouseUp}
                className={cn(
                  'p-4 bg-white/10 hover:bg-white/20 rounded-lg transition-colors active:bg-white/30',
                  activeButton === 'tilt_down' && 'bg-white/30'
                )}
                title="Tilt Down"
              >
                <ChevronDown className="w-6 h-6 text-white" />
              </button>
              <div />
            </div>

            {/* Zoom Controls */}
            <div className="grid grid-cols-2 gap-2 mt-2">
              <button
                onMouseDown={() => handleMouseDown('zoom_in')}
                onMouseUp={handleMouseUp}
                onMouseLeave={handleMouseUp}
                className={cn(
                  'p-3 bg-white/10 hover:bg-white/20 rounded-lg transition-colors flex items-center justify-center gap-2 active:bg-white/30',
                  activeButton === 'zoom_in' && 'bg-white/30'
                )}
                title="Zoom In"
              >
                <ZoomIn className="w-5 h-5 text-white" />
                <span className="text-white text-sm">Zoom In</span>
              </button>
              <button
                onMouseDown={() => handleMouseDown('zoom_out')}
                onMouseUp={handleMouseUp}
                onMouseLeave={handleMouseUp}
                className={cn(
                  'p-3 bg-white/10 hover:bg-white/20 rounded-lg transition-colors flex items-center justify-center gap-2 active:bg-white/30',
                  activeButton === 'zoom_out' && 'bg-white/30'
                )}
                title="Zoom Out"
              >
                <ZoomOut className="w-5 h-5 text-white" />
                <span className="text-white text-sm">Zoom Out</span>
              </button>
            </div>

            {/* Presets */}
            <div className="border-t border-white/20 pt-3 mt-2">
              <div className="text-white/60 text-xs mb-2">Presets</div>
              <div className="grid grid-cols-4 gap-2">
                {[1, 2, 3, 4].map((preset) => (
                  <button
                    key={preset}
                    onClick={() => handlePTZCommand('preset', { preset_id: preset })}
                    className="p-2 bg-white/10 hover:bg-white/20 rounded text-white text-sm transition-colors"
                    title={`Preset ${preset}`}
                  >
                    {preset}
                  </button>
                ))}
              </div>
            </div>

            {/* Help text */}
            <div className="text-white/40 text-xs text-center mt-2">
              Hold directional buttons to move camera
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
