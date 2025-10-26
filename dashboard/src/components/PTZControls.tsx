import React, { useState } from 'react';
import {
  ChevronUp,
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  ZoomIn,
  ZoomOut,
  Home,
  Pin,
  PinOff,
} from 'lucide-react';
import { api } from '@/services/api';
import { cn } from '@/utils/cn';
import type { Camera } from '@/types';

interface PTZControlsProps {
  camera: Camera;
  onTogglePin: () => void;
  isPinned: boolean;
  cellSize?: 'hotspot' | 'large' | 'medium' | 'small';
}

export function PTZControls({
  camera,
  onTogglePin,
  isPinned,
  cellSize = 'medium',
}: PTZControlsProps) {
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
    return null;
  }

  // Responsive sizing based on cell size
  const sizeClasses = {
    hotspot: {
      container: 'w-16',
      button: 'h-11 w-11',
      icon: 'w-5 h-5',
      gap: 'gap-1.5',
      padding: 'p-2',
    },
    large: {
      container: 'w-14',
      button: 'h-9 w-9',
      icon: 'w-4 h-4',
      gap: 'gap-1',
      padding: 'p-1.5',
    },
    medium: {
      container: 'w-12',
      button: 'h-8 w-8',
      icon: 'w-3.5 h-3.5',
      gap: 'gap-1',
      padding: 'p-1',
    },
    small: {
      container: 'w-10',
      button: 'h-6 w-6',
      icon: 'w-3 h-3',
      gap: 'gap-0.5',
      padding: 'p-1',
    },
  };

  const size = sizeClasses[cellSize];

  const buttonClass = cn(
    'bg-white/10 hover:bg-white/20 rounded transition-colors active:bg-white/30 flex items-center justify-center',
    size.button
  );

  const activeButtonClass = cn(buttonClass, 'bg-white/30 ring-1 ring-white/50');

  return (
    <div
      className={cn(
        'absolute left-0 top-0 bottom-0 flex flex-col bg-black/80 backdrop-blur-md border-r border-white/10 pointer-events-auto z-10',
        size.container,
        size.padding
      )}
      onClick={(e) => e.stopPropagation()} // Prevent click from bubbling to parent
    >
      {/* Directional controls - Single column */}
      <div className={cn('flex flex-col', size.gap)}>
        {/* Up */}
        <button
          onMouseDown={() => handleMouseDown('tilt_up')}
          onMouseUp={handleMouseUp}
          onMouseLeave={handleMouseUp}
          className={
            activeButton === 'tilt_up' ? activeButtonClass : buttonClass
          }
          title="Tilt Up"
        >
          <ChevronUp className={cn('text-white', size.icon)} />
        </button>

        {/* Left */}
        <button
          onMouseDown={() => handleMouseDown('pan_left')}
          onMouseUp={handleMouseUp}
          onMouseLeave={handleMouseUp}
          className={
            activeButton === 'pan_left' ? activeButtonClass : buttonClass
          }
          title="Pan Left"
        >
          <ChevronLeft className={cn('text-white', size.icon)} />
        </button>

        {/* Home */}
        <button
          onClick={() => handlePTZCommand('home')}
          className={cn(
            'bg-primary-600/80 hover:bg-primary-600 rounded transition-colors flex items-center justify-center',
            size.button
          )}
          title="Home Position"
        >
          <Home className={cn('text-white', size.icon)} />
        </button>

        {/* Right */}
        <button
          onMouseDown={() => handleMouseDown('pan_right')}
          onMouseUp={handleMouseUp}
          onMouseLeave={handleMouseUp}
          className={
            activeButton === 'pan_right' ? activeButtonClass : buttonClass
          }
          title="Pan Right"
        >
          <ChevronRight className={cn('text-white', size.icon)} />
        </button>

        {/* Down */}
        <button
          onMouseDown={() => handleMouseDown('tilt_down')}
          onMouseUp={handleMouseUp}
          onMouseLeave={handleMouseUp}
          className={
            activeButton === 'tilt_down' ? activeButtonClass : buttonClass
          }
          title="Tilt Down"
        >
          <ChevronDown className={cn('text-white', size.icon)} />
        </button>
      </div>

      {/* Divider */}
      <div className="h-px bg-white/10 my-1.5" />

      {/* Zoom controls */}
      <div className={cn('flex flex-col', size.gap)}>
        <button
          onMouseDown={() => handleMouseDown('zoom_in')}
          onMouseUp={handleMouseUp}
          onMouseLeave={handleMouseUp}
          className={
            activeButton === 'zoom_in' ? activeButtonClass : buttonClass
          }
          title="Zoom In"
        >
          <ZoomIn className={cn('text-white', size.icon)} />
        </button>
        <button
          onMouseDown={() => handleMouseDown('zoom_out')}
          onMouseUp={handleMouseUp}
          onMouseLeave={handleMouseUp}
          className={
            activeButton === 'zoom_out' ? activeButtonClass : buttonClass
          }
          title="Zoom Out"
        >
          <ZoomOut className={cn('text-white', size.icon)} />
        </button>
      </div>

      {/* Spacer */}
      <div className="flex-1" />

      {/* Pin/Unpin button at bottom */}
      <button
        onClick={onTogglePin}
        className={cn(
          isPinned
            ? 'bg-primary-600/80 hover:bg-primary-600'
            : 'bg-white/10 hover:bg-white/20',
          'rounded transition-colors flex items-center justify-center',
          size.button
        )}
        title={isPinned ? 'Unpin Controls' : 'Pin Controls'}
      >
        {isPinned ? (
          <Pin className={cn('text-white', size.icon)} />
        ) : (
          <PinOff className={cn('text-white', size.icon)} />
        )}
      </button>
    </div>
  );
}
