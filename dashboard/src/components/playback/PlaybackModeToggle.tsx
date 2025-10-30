import React from 'react';
import { Play, Circle } from 'lucide-react';
import { cn } from '@/utils/cn';

interface PlaybackModeToggleProps {
  mode: 'live' | 'playback';
  onChange: (mode: 'live' | 'playback') => void;
  className?: string;
}

export function PlaybackModeToggle({
  mode,
  onChange,
  className,
}: PlaybackModeToggleProps) {
  return (
    <div className={cn('inline-flex items-center gap-1 bg-gray-100 rounded-lg p-1', className)}>
      {/* Live Mode Button */}
      <button
        onClick={() => onChange('live')}
        className={cn(
          'inline-flex items-center gap-1.5 px-3 py-1.5 rounded-md text-sm font-medium transition-all',
          mode === 'live'
            ? 'bg-red-600 text-white shadow-sm'
            : 'bg-transparent text-gray-700 hover:bg-gray-200'
        )}
        title="Switch to live stream"
      >
        <Circle className={cn('w-3 h-3', mode === 'live' && 'fill-current animate-pulse')} />
        <span>LIVE</span>
      </button>

      {/* Playback Mode Button */}
      <button
        onClick={() => onChange('playback')}
        className={cn(
          'inline-flex items-center gap-1.5 px-3 py-1.5 rounded-md text-sm font-medium transition-all',
          mode === 'playback'
            ? 'bg-blue-600 text-white shadow-sm'
            : 'bg-transparent text-gray-700 hover:bg-gray-200'
        )}
        title="Switch to playback mode"
      >
        <Play className="w-3 h-3" />
        <span>PLAYBACK</span>
        {mode === 'playback' && (
          <span className="text-xs">✓</span>
        )}
      </button>
    </div>
  );
}
