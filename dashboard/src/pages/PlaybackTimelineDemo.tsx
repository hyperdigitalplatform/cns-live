import React, { useState, useEffect } from 'react';
import { PlaybackControlBarEnhanced } from '../components/playback/PlaybackControlBarEnhanced';

/**
 * Demo page to test the enhanced playback timeline
 *
 * Usage:
 * 1. Add to your router: /playback-demo
 * 2. Navigate to the page
 * 3. Click "Start Demo" to begin playback simulation
 * 4. Test zoom, scrubbing, animations
 */
export function PlaybackTimelineDemo() {
  const [isPlaying, setIsPlaying] = useState(false);
  const [currentTime, setCurrentTime] = useState(new Date(Date.now() - 30 * 60 * 1000)); // 30 min ago
  const [zoomLevel, setZoomLevel] = useState(60); // 1 hour in minutes

  // Define timeline range (2 hours centered on current time)
  const startTime = new Date(currentTime.getTime() - 60 * 60 * 1000); // 1 hour before
  const endTime = new Date(currentTime.getTime() + 60 * 60 * 1000);   // 1 hour after

  // Mock recording sequences (simulate 3 recordings with gaps)
  const sequences = [
    {
      sequenceId: 'seq-1',
      startTime: new Date(Date.now() - 55 * 60 * 1000).toISOString(), // 55 min ago
      endTime: new Date(Date.now() - 45 * 60 * 1000).toISOString(),   // 45 min ago
      durationSeconds: 600
    },
    {
      sequenceId: 'seq-2',
      startTime: new Date(Date.now() - 35 * 60 * 1000).toISOString(), // 35 min ago
      endTime: new Date(Date.now() - 20 * 60 * 1000).toISOString(),   // 20 min ago
      durationSeconds: 900
    },
    {
      sequenceId: 'seq-3',
      startTime: new Date(Date.now() - 10 * 60 * 1000).toISOString(), // 10 min ago
      endTime: new Date(Date.now() - 2 * 60 * 1000).toISOString(),    // 2 min ago
      durationSeconds: 480
    }
  ];

  // Simulate playback progression
  useEffect(() => {
    if (!isPlaying) return;

    const interval = setInterval(() => {
      setCurrentTime(prev => {
        const next = new Date(prev.getTime() + 1000); // Advance 1 second

        // Stop at current time (don't go into future)
        if (next > new Date()) {
          setIsPlaying(false);
          return prev;
        }

        return next;
      });
    }, 1000);

    return () => clearInterval(interval);
  }, [isPlaying]);

  // Handle play/pause
  const handlePlayPause = () => {
    setIsPlaying(!isPlaying);
  };

  // Handle seek
  const handleSeek = (time: Date) => {
    setCurrentTime(time);
  };

  // Handle timeline scroll
  const handleScrollTimeline = (direction: 'left' | 'right') => {
    const scrollAmount = zoomLevel * 60 * 1000; // zoom level in ms

    if (direction === 'left') {
      setCurrentTime(prev => new Date(prev.getTime() - scrollAmount));
    } else {
      setCurrentTime(prev => {
        const next = new Date(prev.getTime() + scrollAmount);
        // Don't scroll beyond current time
        return next > new Date() ? new Date() : next;
      });
    }
  };

  // Handle zoom change
  const handleZoomChange = (newZoom: number) => {
    setZoomLevel(newZoom);
  };

  // Check if current time has recording
  const hasRecording = sequences.some(seq => {
    const start = new Date(seq.startTime).getTime();
    const end = new Date(seq.endTime).getTime();
    return currentTime.getTime() >= start && currentTime.getTime() <= end;
  });

  return (
    <div className="min-h-screen bg-gray-950 text-white p-8">
      <div className="max-w-7xl mx-auto space-y-8">
        {/* Header */}
        <div className="text-center space-y-4">
          <h1 className="text-4xl font-bold">Enhanced Playback Timeline Demo</h1>
          <p className="text-gray-400">
            Test the new scrolling timeline with smooth animations
          </p>
        </div>

        {/* Demo Controls */}
        <div className="bg-gray-900 rounded-lg p-6 space-y-4">
          <h2 className="text-xl font-semibold mb-4">Demo Controls</h2>

          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <button
              onClick={() => setIsPlaying(!isPlaying)}
              className="px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded-lg font-medium transition-colors"
            >
              {isPlaying ? '⏸ Pause Demo' : '▶ Start Demo'}
            </button>

            <button
              onClick={() => setCurrentTime(new Date(Date.now() - 30 * 60 * 1000))}
              className="px-4 py-2 bg-gray-700 hover:bg-gray-600 rounded-lg font-medium transition-colors"
            >
              🔄 Reset Time
            </button>

            <button
              onClick={() => handleZoomChange(1)} // 1 min zoom
              className="px-4 py-2 bg-purple-600 hover:bg-purple-700 rounded-lg font-medium transition-colors"
            >
              🔍 Zoom In (1 min)
            </button>

            <button
              onClick={() => handleZoomChange(168 * 60)} // 1 week zoom
              className="px-4 py-2 bg-purple-600 hover:bg-purple-700 rounded-lg font-medium transition-colors"
            >
              🔍 Zoom Out (1 wk)
            </button>
          </div>

          {/* Status Display */}
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 pt-4 border-t border-gray-700">
            <div>
              <div className="text-xs text-gray-400 mb-1">Current Time</div>
              <div className="font-mono text-sm">
                {currentTime.toLocaleTimeString()}
              </div>
            </div>
            <div>
              <div className="text-xs text-gray-400 mb-1">Playback Status</div>
              <div className="font-mono text-sm">
                {isPlaying ? '▶ Playing' : '⏸ Paused'}
              </div>
            </div>
            <div>
              <div className="text-xs text-gray-400 mb-1">Recording</div>
              <div className={`font-mono text-sm ${hasRecording ? 'text-green-400' : 'text-red-400'}`}>
                {hasRecording ? '✓ Available' : '✗ No Recording'}
              </div>
            </div>
            <div>
              <div className="text-xs text-gray-400 mb-1">Zoom Level</div>
              <div className="font-mono text-sm">
                {zoomLevel < 60 ? `${zoomLevel} min` : `${zoomLevel / 60} hr`}
              </div>
            </div>
          </div>
        </div>

        {/* Enhanced Timeline */}
        <div className="bg-gray-900 rounded-lg p-6 space-y-4">
          <h2 className="text-xl font-semibold mb-4">Enhanced Timeline Component</h2>

          <PlaybackControlBarEnhanced
            startTime={startTime}
            endTime={endTime}
            currentTime={currentTime}
            sequences={sequences}
            isPlaying={isPlaying}
            zoomLevel={zoomLevel}
            onPlayPause={handlePlayPause}
            onSeek={handleSeek}
            onScrollTimeline={handleScrollTimeline}
            onZoomChange={handleZoomChange}
            hasRecording={hasRecording}
          />
        </div>

        {/* Feature Showcase */}
        <div className="bg-gray-900 rounded-lg p-6 space-y-4">
          <h2 className="text-xl font-semibold mb-4">✨ Features to Test</h2>

          <div className="grid md:grid-cols-2 gap-4 text-sm">
            <div className="space-y-2">
              <h3 className="font-semibold text-blue-400">Scrolling</h3>
              <ul className="space-y-1 text-gray-300">
                <li>• Click "Start Demo" to see auto-scroll</li>
                <li>• Timeline scrolls, center stays fixed</li>
                <li>• Current time always centered</li>
                <li>• Smooth 60fps scrolling</li>
              </ul>
            </div>

            <div className="space-y-2">
              <h3 className="font-semibold text-purple-400">Zoom Animations</h3>
              <ul className="space-y-1 text-gray-300">
                <li>• Click zoom levels in dropdown</li>
                <li>• Labels fade out/in smoothly</li>
                <li>• Ticks animate on zoom change</li>
                <li>• 10 levels: 1 min to 1 week</li>
              </ul>
            </div>

            <div className="space-y-2">
              <h3 className="font-semibold text-green-400">Interactive</h3>
              <ul className="space-y-1 text-gray-300">
                <li>• Click timeline to seek</li>
                <li>• Drag to scrub through time</li>
                <li>• Hover for time tooltip</li>
                <li>• Use arrow buttons to scroll</li>
              </ul>
            </div>

            <div className="space-y-2">
              <h3 className="font-semibold text-orange-400">Visual Polish</h3>
              <ul className="space-y-1 text-gray-300">
                <li>• Tick marks (major + minor)</li>
                <li>• Green future zone overlay</li>
                <li>• Orange recording bars</li>
                <li>• Smooth transitions everywhere</li>
              </ul>
            </div>
          </div>
        </div>

        {/* Technical Info */}
        <div className="bg-gray-900 rounded-lg p-6 space-y-4">
          <h2 className="text-xl font-semibold mb-4">🔧 Technical Details</h2>

          <div className="space-y-3 text-sm">
            <div className="flex justify-between border-b border-gray-700 pb-2">
              <span className="text-gray-400">Buffer Technique:</span>
              <span className="font-mono">3x (1.5x on each side)</span>
            </div>
            <div className="flex justify-between border-b border-gray-700 pb-2">
              <span className="text-gray-400">Scrolling Method:</span>
              <span className="font-mono">CSS Transform (GPU)</span>
            </div>
            <div className="flex justify-between border-b border-gray-700 pb-2">
              <span className="text-gray-400">Animation Duration:</span>
              <span className="font-mono">800ms (cubic-bezier)</span>
            </div>
            <div className="flex justify-between border-b border-gray-700 pb-2">
              <span className="text-gray-400">Tick Marks:</span>
              <span className="font-mono">Dynamic (major + minor)</span>
            </div>
            <div className="flex justify-between">
              <span className="text-gray-400">Performance:</span>
              <span className="font-mono text-green-400">60fps target</span>
            </div>
          </div>
        </div>

        {/* Instructions */}
        <div className="bg-blue-900/20 border border-blue-500/30 rounded-lg p-6">
          <h3 className="font-semibold text-blue-400 mb-2">💡 Testing Tips</h3>
          <ul className="space-y-1 text-sm text-gray-300">
            <li>1. Open browser DevTools → Performance tab</li>
            <li>2. Start recording, then click "Start Demo"</li>
            <li>3. Change zoom levels multiple times</li>
            <li>4. Check for 60fps (green bars in timeline)</li>
            <li>5. Monitor memory usage for leaks</li>
          </ul>
        </div>
      </div>
    </div>
  );
}
