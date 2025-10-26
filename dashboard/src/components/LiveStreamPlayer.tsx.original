import React, { useEffect, useState } from 'react';
import {
  LiveKitRoom,
  VideoTrack,
  useRoomContext,
  useTracks,
} from '@livekit/components-react';
import { Track } from 'livekit-client';
import type { Camera, StreamReservation } from '@/types';
import { useStreamStore } from '@/stores/streamStore';
import { PTZControls } from './PTZControls';
import { Loader2, AlertCircle } from 'lucide-react';

// Module-level map to track reservations in progress (persists across React.StrictMode remounts)
const pendingReservations = new Map<string, Promise<StreamReservation>>();

interface LiveStreamPlayerProps {
  camera: Camera;
  quality?: 'high' | 'medium' | 'low';
  onError?: (error: Error) => void;
}

export function LiveStreamPlayer({
  camera,
  quality = 'medium',
  onError,
}: LiveStreamPlayerProps) {
  const [reservation, setReservation] = useState<StreamReservation | null>(
    null
  );
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showPTZ, setShowPTZ] = useState(false);
  const [ptzPinned, setPTZPinned] = useState(false);
  const reserveStream = useStreamStore((state) => state.reserveStream);
  const releaseStream = useStreamStore((state) => state.releaseStream);

  useEffect(() => {
    let mounted = true;
    let currentReservationId: string | null = null;

    const initStream = async () => {
      const cameraKey = `${camera.id}-${quality}`;

      // Check if there's already a pending reservation for this camera
      let pendingPromise = pendingReservations.get(cameraKey);

      if (!pendingPromise) {
        // Create new reservation promise
        pendingPromise = reserveStream(camera.id, quality);
        pendingReservations.set(cameraKey, pendingPromise);

        // Clean up the pending promise after it completes (success or failure)
        pendingPromise.finally(() => {
          pendingReservations.delete(cameraKey);
        });
      }

      try {
        setLoading(true);
        const res = await pendingPromise;
        if (mounted) {
          currentReservationId = res.reservation_id;
          setReservation(res);
          setLoading(false);
        }
      } catch (err) {
        if (mounted) {
          const errorMessage =
            err instanceof Error ? err.message : 'Failed to reserve stream';
          setError(errorMessage);
          setLoading(false);
          onError?.(
            err instanceof Error ? err : new Error('Failed to reserve stream')
          );
        }
      }
    };

    initStream();

    return () => {
      mounted = false;
      // Release stream on cleanup using the captured reservation ID
      if (currentReservationId) {
        releaseStream(currentReservationId);
      }
    };
      // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [camera.id, quality]);
  if (loading) {
    return (
      <div className="flex items-center justify-center h-full bg-gray-900">
        <Loader2 className="w-8 h-8 text-white animate-spin" />
      </div>
    );
  }

  if (error || !reservation) {
    return (
      <div className="flex flex-col items-center justify-center h-full bg-gray-900 text-white p-4">
        <AlertCircle className="w-12 h-12 text-red-500 mb-2" />
        <p className="text-sm">{error || 'Failed to load stream'}</p>
      </div>
    );
  }

  const handlePTZClick = () => {
    setPTZPinned(true);
    setShowPTZ(true);
  };

  const handlePTZClose = () => {
    setPTZPinned(false);
    setShowPTZ(false);
  };

  return (
    <LiveKitRoom
      serverUrl={reservation.livekit_url}
      token={reservation.token}
      connect={true}
      audio={false}
      video={false} // Don't publish - we only subscribe to camera streams
      className="h-full"
    >
      <LiveStreamView
        camera={camera}
        showPTZ={showPTZ}
        ptzPinned={ptzPinned}
        onShowPTZ={setShowPTZ}
        onPTZClick={handlePTZClick}
        onPTZClose={handlePTZClose}
      />
    </LiveKitRoom>
  );
}

function LiveStreamView({
  camera,
  showPTZ,
  ptzPinned,
  onShowPTZ,
  onPTZClick,
  onPTZClose,
}: {
  camera: Camera;
  showPTZ: boolean;
  ptzPinned: boolean;
  onShowPTZ: (show: boolean) => void;
  onPTZClick: () => void;
  onPTZClose: () => void;
}) {
  const room = useRoomContext();
  const tracks = useTracks([Track.Source.Camera]);

  if (tracks.length === 0) {
    return (
      <div className="flex items-center justify-center h-full bg-gray-900 text-white">
        <div className="text-center">
          <Loader2 className="w-8 h-8 animate-spin mx-auto mb-2" />
          <p className="text-sm">Connecting to {camera.name}...</p>
        </div>
      </div>
    );
  }

  const videoTrack = tracks[0];

  return (
    <div
      className="relative h-full bg-black group"
      onMouseEnter={() => !ptzPinned && camera.ptz_enabled && onShowPTZ(true)}
      onMouseLeave={() => !ptzPinned && onShowPTZ(false)}
      onClick={() => !ptzPinned && camera.ptz_enabled && onPTZClick()}
    >
      <VideoTrack
        trackRef={videoTrack}
        className="h-full w-full object-contain"
      />

      {/* Camera name overlay */}
      <div className="absolute top-2 left-2 bg-black/70 px-3 py-1 rounded">
        <p className="text-white text-sm font-medium">{camera.name}</p>
      </div>

      {/* Stream info overlay */}
      <div className="absolute bottom-2 right-2 bg-black/70 px-3 py-1 rounded text-xs text-white">
        <span className="inline-block w-2 h-2 bg-red-500 rounded-full mr-1 animate-pulse" />
        LIVE
      </div>

      {/* PTZ hint (only show on hover if PTZ enabled and not pinned) */}
      {camera.ptz_enabled && !ptzPinned && (
        <div className="absolute inset-0 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none">
          <div className="bg-black/50 px-4 py-2 rounded-lg text-white text-sm">
            Click for PTZ Controls
          </div>
        </div>
      )}

      {/* PTZ Controls Overlay */}
      {showPTZ && (
        <PTZControls
          camera={camera}
          onClose={onPTZClose}
          isPinned={ptzPinned}
        />
      )}
    </div>
  );
}
