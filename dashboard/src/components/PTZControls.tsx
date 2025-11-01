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
  Circle,
  Square,
  Video,
} from 'lucide-react';
import { api } from '@/services/api';
import { cn } from '@/utils/cn';
import type { Camera } from '@/types';

interface PTZControlsProps {
  camera: Camera;
  onTogglePin: () => void;
  isPinned: boolean;
  cellSize?: 'hotspot' | 'large' | 'medium' | 'small';
  isRecording?: boolean;
  onRecordingChange?: (isRecording: boolean) => void;
  onViewRecordings?: () => void;
}

export function PTZControls({
  camera,
  onTogglePin,
  isPinned,
  cellSize = 'medium',
  isRecording = false,
  onRecordingChange,
  onViewRecordings,
}: PTZControlsProps) {
  const [activeButton, setActiveButton] = useState<string | null>(null);
  const [isCommandPending, setIsCommandPending] = useState(false);
  const [recordingLoading, setRecordingLoading] = useState(false);

  const handlePTZCommand = async (
    command: string,
    params?: { speed?: number; preset_id?: number }
  ) => {
    setIsCommandPending(true);
    try {
      await api.controlPTZ(camera.id, command, params);
    } catch (error) {
      console.error('PTZ command failed:', error);
    } finally {
      setIsCommandPending(false);
    }
  };

  const handleMouseDown = (command: string) => {
    if (isCommandPending) return; // Prevent new commands while one is pending
    setActiveButton(command);
    handlePTZCommand(command, { speed: 0.5 });
  };

  const handleMouseUp = () => {
    // STOP commands should always be allowed, even when pending
    // Send STOP command if a button was active
    if (activeButton) {
      // Don't use handlePTZCommand for STOP to avoid setting isCommandPending again
      api.controlPTZ(camera.id, 'stop', { speed: 0 }).catch((error) => {
        console.error('PTZ stop command failed:', error);
      });
    }
    setActiveButton(null);
  };

  const handleToggleRecording = async () => {
    setRecordingLoading(true);
    try {
      if (isRecording) {
        await api.stopMilestoneRecording(camera.id);
        onRecordingChange?.(false);
      } else {
        await api.startMilestoneRecording({
          cameraId: camera.id,
          durationMinutes: 30
        });
        onRecordingChange?.(true);
      }
    } catch (error) {
      console.error('Recording control error:', error);
    } finally {
      setRecordingLoading(false);
    }
  };

  const handleViewRecordings = () => {
    onViewRecordings?.();
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
      padding: 'px-2 pb-2',
    },
    large: {
      container: 'w-14',
      button: 'h-9 w-9',
      icon: 'w-4 h-4',
      gap: 'gap-1',
      padding: 'px-1.5 pb-1.5',
    },
    medium: {
      container: 'w-12',
      button: 'h-8 w-8',
      icon: 'w-3.5 h-3.5',
      gap: 'gap-1',
      padding: 'px-1 pb-1',
    },
    small: {
      container: 'w-10',
      button: 'h-6 w-6',
      icon: 'w-3 h-3',
      gap: 'gap-0.5',
      padding: 'px-1 pb-1',
    },
  };

  const size = sizeClasses[cellSize];

  const buttonClass = cn(
    'bg-white/10 rounded transition-colors flex items-center justify-center',
    isCommandPending
      ? 'opacity-50 cursor-not-allowed'
      : 'hover:bg-white/20 active:bg-white/30 cursor-pointer',
    size.button
  );

  const activeButtonClass = cn(buttonClass, 'bg-white/30 ring-1 ring-white/50');

  return (
    <div
      className={cn(
        'absolute left-0 top-0 bottom-0 flex flex-col bg-black/80 backdrop-blur-md border-r border-white/10 pointer-events-auto z-10 pt-12',
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
          disabled={isCommandPending}
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
          disabled={isCommandPending}
          className={
            activeButton === 'pan_left' ? activeButtonClass : buttonClass
          }
          title="Pan Left"
        >
          <ChevronLeft className={cn('text-white', size.icon)} />
        </button>

        {/* Home */}
        <button
          onClick={() => !isCommandPending && handlePTZCommand('home')}
          disabled={isCommandPending}
          className={cn(
            'bg-primary-600/80 rounded transition-colors flex items-center justify-center',
            isCommandPending
              ? 'opacity-50 cursor-not-allowed'
              : 'hover:bg-primary-600 cursor-pointer',
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
          disabled={isCommandPending}
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
          disabled={isCommandPending}
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
          disabled={isCommandPending}
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
          disabled={isCommandPending}
          className={
            activeButton === 'zoom_out' ? activeButtonClass : buttonClass
          }
          title="Zoom Out"
        >
          <ZoomOut className={cn('text-white', size.icon)} />
        </button>
      </div>

      {/* Divider */}
      <div className="h-px bg-white/10 my-1.5" />

      {/* Recording controls */}
      <div className={cn('flex flex-col', size.gap)}>
        <button
          onClick={handleToggleRecording}
          disabled={recordingLoading}
          className={cn(
            'rounded transition-colors flex items-center justify-center',
            recordingLoading
              ? 'opacity-50 cursor-not-allowed'
              : isRecording
              ? 'bg-red-600/80 hover:bg-red-600 cursor-pointer'
              : 'bg-white/10 hover:bg-white/20 cursor-pointer',
            size.button
          )}
          title={isRecording ? 'Stop Recording' : 'Start Recording'}
        >
          {recordingLoading ? (
            <div className={cn('border-2 border-white border-t-transparent rounded-full animate-spin', size.icon)} />
          ) : isRecording ? (
            <Square className={cn('text-white fill-current', size.icon)} />
          ) : (
            <Circle className={cn('text-white fill-current', size.icon)} />
          )}
        </button>
        <button
          onClick={handleViewRecordings}
          className={cn(
            'bg-white/10 rounded transition-colors flex items-center justify-center hover:bg-white/20 cursor-pointer',
            size.button
          )}
          title="View Recordings"
        >
          <Video className={cn('text-white', size.icon)} />
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
