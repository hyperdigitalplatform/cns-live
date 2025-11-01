import React, { useMemo } from 'react';
import { cn } from '@/utils/cn';

interface TimelineTicksProps {
  startTime: Date;
  endTime: Date;
  majorTickMs: number;
  minorTickMs: number;
  isAnimating?: boolean;
  zoomDirection?: 'in' | 'out' | 'none';
}

export function TimelineTicks({
  startTime,
  endTime,
  majorTickMs,
  minorTickMs,
  isAnimating = false,
  zoomDirection = 'none',
}: TimelineTicksProps) {
  const totalDuration = endTime.getTime() - startTime.getTime();

  // Generate tick marks based on intervals
  const ticks = useMemo(() => {
    const minorTicks: Array<{ time: Date; type: 'minor'; position: number }> = [];
    const majorTicks: Array<{ time: Date; type: 'major'; position: number }> = [];

    // Generate minor ticks
    const firstMinorTime = new Date(Math.ceil(startTime.getTime() / minorTickMs) * minorTickMs);
    let currentMinor = firstMinorTime;

    while (currentMinor.getTime() <= endTime.getTime()) {
      const position = ((currentMinor.getTime() - startTime.getTime()) / totalDuration) * 100;
      minorTicks.push({
        time: new Date(currentMinor),
        type: 'minor',
        position
      });
      currentMinor = new Date(currentMinor.getTime() + minorTickMs);
    }

    // Generate major ticks
    const firstMajorTime = new Date(Math.ceil(startTime.getTime() / majorTickMs) * majorTickMs);
    let currentMajor = firstMajorTime;

    while (currentMajor.getTime() <= endTime.getTime()) {
      const position = ((currentMajor.getTime() - startTime.getTime()) / totalDuration) * 100;
      majorTicks.push({
        time: new Date(currentMajor),
        type: 'major',
        position
      });
      currentMajor = new Date(currentMajor.getTime() + majorTickMs);
    }

    return { minorTicks, majorTicks };
  }, [startTime, endTime, majorTickMs, minorTickMs, totalDuration]);

  // Calculate animation transform based on zoom direction
  const getTickAnimation = (position: number, index: number) => {
    if (!isAnimating) {
      return {
        opacity: 1,
        transform: 'scaleY(1)',
        transitionDelay: '0ms'
      };
    }

    // Distance from center (50%)
    const distanceFromCenter = position - 50;

    // Zoom IN: ticks move OUTWARD (away from center)
    // Zoom OUT: ticks move INWARD (toward center)
    const moveDistance = zoomDirection === 'in'
      ? distanceFromCenter * 0.5
      : zoomDirection === 'out'
      ? -distanceFromCenter * 0.3
      : 0;

    const scale = zoomDirection === 'out' ? 0.5 : 1.3;

    return {
      opacity: 0,
      transform: `translateX(${moveDistance}px) scaleY(${scale})`,
      transitionDelay: `${index * 2}ms` // Staggered animation
    };
  };

  return (
    <div className="relative w-full h-full">
      {/* Minor ticks */}
      {ticks.minorTicks.map((tick, index) => {
        const animation = isAnimating
          ? getTickAnimation(tick.position, index)
          : { opacity: 1, transform: 'scaleY(1)', transitionDelay: '0ms' };

        return (
          <div
            key={`minor-${tick.time.getTime()}`}
            className={cn(
              "absolute bottom-0 w-px bg-white/20",
              "transition-all duration-800 ease-in-out"
            )}
            style={{
              left: `${tick.position}%`,
              height: '8px',
              opacity: animation.opacity,
              transform: animation.transform,
              transitionDelay: animation.transitionDelay
            }}
          />
        );
      })}

      {/* Major ticks */}
      {ticks.majorTicks.map((tick, index) => {
        const animation = isAnimating
          ? getTickAnimation(tick.position, index)
          : { opacity: 1, transform: 'scaleY(1)', transitionDelay: '0ms' };

        return (
          <div
            key={`major-${tick.time.getTime()}`}
            className={cn(
              "absolute bottom-0 w-px bg-white/50",
              "transition-all duration-800 ease-in-out"
            )}
            style={{
              left: `${tick.position}%`,
              height: '15px',
              opacity: animation.opacity,
              transform: animation.transform,
              transitionDelay: animation.transitionDelay
            }}
          />
        );
      })}
    </div>
  );
}
