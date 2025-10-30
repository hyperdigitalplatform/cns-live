import { useEffect, useRef, useState, useCallback } from 'react';

interface WebRTCPlaybackOptions {
  cameraId: string;
  playbackTime: Date;
  skipGaps?: boolean;
  speed?: number;
  onStateChange?: (state: 'connecting' | 'connected' | 'failed' | 'disconnected') => void;
  onError?: (error: string) => void;
}

type ConnectionState = 'idle' | 'connecting' | 'connected' | 'failed' | 'disconnected';

/**
 * Custom hook for WebRTC playback from Milestone VMS
 * Manages WebRTC peer connection, signaling, and ICE candidate exchange
 *
 * Based on Milestone WebRTC API:
 * - POST /api/v1/cameras/{id}/playback/start - Create session
 * - PUT /api/v1/playback/webrtc/answer - Send answer SDP
 * - POST/GET /api/v1/playback/webrtc/ice - Exchange ICE candidates
 *
 * Performance & Memory Profiling:
 * To check for memory leaks using Chrome DevTools:
 * 1. Open DevTools → Memory tab
 * 2. Take heap snapshot before playback
 * 3. Start playback, let it run for 5-10 minutes
 * 4. Stop playback and take another snapshot
 * 5. Compare snapshots - look for:
 *    - RTCPeerConnection objects that weren't freed
 *    - MediaStream objects accumulating
 *    - Interval/Timeout handles not cleared
 *
 * Expected behavior:
 * - Each connection should fully clean up on unmount
 * - Switching cameras should release previous connections
 * - No objects should accumulate over multiple play/stop cycles
 */
export function useWebRTCPlayback(options: WebRTCPlaybackOptions) {
  const [state, setState] = useState<ConnectionState>('idle');
  const [error, setError] = useState<string | null>(null);

  const videoRef = useRef<HTMLVideoElement>(null);
  const pcRef = useRef<RTCPeerConnection | null>(null);
  const sessionIdRef = useRef<string | null>(null);
  const candidateIntervalRef = useRef<NodeJS.Timeout | null>(null);
  const connectionTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const statsIntervalRef = useRef<NodeJS.Timeout | null>(null);
  const mountedRef = useRef(true);

  // Use refs for callbacks to prevent dependency changes
  const onStateChangeRef = useRef(options.onStateChange);
  const onErrorRef = useRef(options.onError);

  // Update refs when options change
  useEffect(() => {
    onStateChangeRef.current = options.onStateChange;
    onErrorRef.current = options.onError;
  }, [options.onStateChange, options.onError]);

  /**
   * Update state and notify parent
   */
  const updateState = useCallback((newState: ConnectionState) => {
    if (!mountedRef.current) return;
    setState(newState);
    if (newState === 'connecting' || newState === 'connected' || newState === 'failed' || newState === 'disconnected') {
      onStateChangeRef.current?.(newState);
    }
  }, []); // No dependencies - uses ref

  /**
   * Update error and notify parent
   */
  const updateError = useCallback((errorMsg: string) => {
    if (!mountedRef.current) return;
    setError(errorMsg);
    onErrorRef.current?.(errorMsg);
  }, []); // No dependencies - uses ref

  /**
   * Collect WebRTC statistics for monitoring
   */
  const collectStats = useCallback(async () => {
    if (!pcRef.current) return;

    try {
      const stats = await pcRef.current.getStats();
      let bytesReceived = 0;
      let packetsLost = 0;
      let packetsReceived = 0;
      let jitter = 0;

      stats.forEach((report) => {
        if (report.type === 'inbound-rtp' && report.kind === 'video') {
          bytesReceived = report.bytesReceived || 0;
          packetsLost = report.packetsLost || 0;
          packetsReceived = report.packetsReceived || 0;
          jitter = report.jitter || 0;
        }
      });

      // Calculate bandwidth (rough estimate based on bytes received)
      const bandwidthMbps = (bytesReceived * 8) / (1024 * 1024 * 10); // Over 10 second interval

      // Log statistics for debugging and monitoring
      console.log('📊 WebRTC Stats:', {
        bandwidth: `${bandwidthMbps.toFixed(2)} Mbps`,
        packetsLost,
        packetsReceived,
        packetLossRate: packetsReceived > 0 ? `${((packetsLost / (packetsLost + packetsReceived)) * 100).toFixed(2)}%` : '0%',
        jitter: `${jitter.toFixed(3)}s`,
        connectionState: pcRef.current?.connectionState,
        iceConnectionState: pcRef.current?.iceConnectionState,
      });
    } catch (err) {
      console.error('Failed to collect WebRTC stats:', err);
    }
  }, []);

  /**
   * Cleanup function to stop playback and close connections
   */
  const cleanup = useCallback(() => {
    // Clear ICE candidate polling
    if (candidateIntervalRef.current) {
      clearTimeout(candidateIntervalRef.current);
      candidateIntervalRef.current = null;
    }

    // Clear connection timeout
    if (connectionTimeoutRef.current) {
      clearTimeout(connectionTimeoutRef.current);
      connectionTimeoutRef.current = null;
    }

    // Clear stats collection
    if (statsIntervalRef.current) {
      clearInterval(statsIntervalRef.current);
      statsIntervalRef.current = null;
    }

    // Close peer connection
    if (pcRef.current) {
      pcRef.current.close();
      pcRef.current = null;
    }

    // Clear video source
    if (videoRef.current) {
      videoRef.current.srcObject = null;
    }

    sessionIdRef.current = null;
  }, []);

  /**
   * Main effect to start WebRTC playback
   */
  useEffect(() => {
    if (!options.cameraId || !options.playbackTime) {
      return;
    }

    let pc: RTCPeerConnection | null = null;
    mountedRef.current = true;

    async function startPlayback() {
      try {
        updateState('connecting');
        setError(null);

        // Step 1: Create RTCPeerConnection
        pc = new RTCPeerConnection({
          iceServers: [
            { urls: 'stun:stun.l.google.com:19302' },
            { urls: 'stun:stun1.l.google.com:19302' }
          ]
        });

        pcRef.current = pc;

        // CRITICAL: Set up ontrack handler FIRST, before any SDP negotiation
        // This ensures we can receive the video track as soon as it arrives
        pc.ontrack = (event) => {
          console.log('🎥 Video track received');
          if (videoRef.current && event.streams[0]) {
            videoRef.current.srcObject = event.streams[0];
          }
        };

        // Step 2: Request playback session from backend
        const response = await fetch(`/api/v1/cameras/${options.cameraId}/playback/start`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            playbackTime: options.playbackTime.toISOString(),
            skipGaps: options.skipGaps ?? true,
            speed: options.speed ?? 1.0,
          }),
        });

        if (!response.ok) {
          const errorData = await response.json().catch(() => ({ error: 'Unknown error' }));

          // Provide user-friendly error messages
          let errorMessage = errorData.message || errorData.error || `HTTP ${response.status}`;

          if (response.status === 404) {
            errorMessage = 'No recording available at this time';
          } else if (response.status === 401 || response.status === 403) {
            errorMessage = 'Authentication required. Please log in again.';
          } else if (response.status === 500) {
            errorMessage = 'Server error. Please try again later.';
          } else if (response.status >= 500) {
            errorMessage = 'Service temporarily unavailable';
          }

          throw new Error(errorMessage);
        }

        const session = await response.json();
        sessionIdRef.current = session.sessionId;

        console.log('📡 WebRTC session created:', session.sessionId);

        // Step 3: Set remote description (offer from Milestone)
        const offerSDP = JSON.parse(session.offerSDP);
        await pc.setRemoteDescription(new RTCSessionDescription(offerSDP));

        console.log('✅ Remote description set');

        // Step 3.5: Create data channel (REQUIRED by Milestone)
        // This must be done BEFORE creating the answer for proper SDP negotiation
        const dataChannel = pc.createDataChannel("commands", { protocol: "videoos-commands" });
        dataChannel.onopen = () => {
          console.log('✅ Data channel opened');
        };
        dataChannel.onerror = (err) => {
          console.error('❌ Data channel error:', err);
        };
        console.log('✅ Data channel created');

        // Step 4: Create answer
        const answer = await pc.createAnswer();
        await pc.setLocalDescription(answer);

        console.log('✅ Local description set');

        // Step 5: Send answer back to Milestone via backend
        const answerResponse = await fetch('/api/v1/playback/webrtc/answer', {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            sessionId: session.sessionId,
            answerSDP: JSON.stringify(pc.localDescription),
          }),
        });

        if (!answerResponse.ok) {
          throw new Error('Failed to send answer SDP');
        }

        console.log('✅ Answer SDP sent');

        // Step 6: Handle ICE candidates from local peer
        pc.onicecandidate = async (event) => {
          if (event.candidate && sessionIdRef.current) {
            try {
              await fetch('/api/v1/playback/webrtc/ice', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                  sessionId: sessionIdRef.current,
                  candidate: event.candidate.toJSON(),
                }),
              });
              console.log('📤 ICE candidate sent');
            } catch (err) {
              console.error('Failed to send ICE candidate:', err);
            }
          }
        };

        // Step 7: Poll for server ICE candidates with exponential backoff
        let pollInterval = 500; // Start at 500ms
        let pollAttempts = 0;
        const maxPollInterval = 5000; // Max 5 seconds

        const pollCandidates = async () => {
          if (!sessionIdRef.current || !pcRef.current) return;

          try {
            const resp = await fetch(`/api/v1/playback/webrtc/ice/${sessionIdRef.current}`);
            if (resp.ok) {
              const data = await resp.json();
              if (data.candidates && data.candidates.length > 0) {
                for (const candidate of data.candidates) {
                  if (pcRef.current && pcRef.current.remoteDescription) {
                    await pcRef.current.addIceCandidate(new RTCIceCandidate(candidate));
                    console.log('📥 ICE candidate received');
                  }
                }
                // Reset interval on successful candidate receipt
                pollInterval = 500;
                pollAttempts = 0;
              } else {
                // No candidates - use exponential backoff
                pollAttempts++;
                pollInterval = Math.min(pollInterval * 1.5, maxPollInterval);
              }
            }
          } catch (err) {
            console.error('Failed to poll ICE candidates:', err);
            // On error, slow down polling
            pollInterval = Math.min(pollInterval * 2, maxPollInterval);
          }

          // Schedule next poll with current interval
          if (sessionIdRef.current && pcRef.current && pc?.connectionState !== 'connected') {
            candidateIntervalRef.current = setTimeout(pollCandidates, pollInterval);
          }
        };

        // Start initial poll
        pollCandidates();

        // Step 8: Connection state monitoring
        pc.onconnectionstatechange = () => {
          const connectionState = pc?.connectionState;
          console.log('🔗 Connection state:', connectionState);

          if (connectionState === 'connected') {
            updateState('connected');

            // Clear connection timeout
            if (connectionTimeoutRef.current) {
              clearTimeout(connectionTimeoutRef.current);
              connectionTimeoutRef.current = null;
            }

            // Stop polling ICE candidates once connected
            if (candidateIntervalRef.current) {
              clearTimeout(candidateIntervalRef.current);
              candidateIntervalRef.current = null;
            }

            // Start collecting statistics every 10 seconds
            statsIntervalRef.current = setInterval(collectStats, 10000);
            // Collect initial stats immediately
            collectStats();
          } else if (connectionState === 'failed') {
            updateState('failed');
            updateError('Unable to establish playback connection. Please check your network and try again.');
            if (candidateIntervalRef.current) {
              clearTimeout(candidateIntervalRef.current);
              candidateIntervalRef.current = null;
            }
          } else if (connectionState === 'disconnected') {
            updateState('disconnected');
            if (candidateIntervalRef.current) {
              clearTimeout(candidateIntervalRef.current);
              candidateIntervalRef.current = null;
            }
          }
        };

        // Step 9: ICE connection state monitoring
        pc.oniceconnectionstatechange = () => {
          const iceState = pc?.iceConnectionState;
          console.log('🧊 ICE connection state:', iceState);

          if (iceState === 'failed') {
            console.error('ICE connection failed - possible network/firewall issue');
            updateError('Connection failed. Please check firewall settings and try again.');
          }
        };

        // Step 10: Connection timeout (30 seconds)
        connectionTimeoutRef.current = setTimeout(() => {
          if (pc?.connectionState !== 'connected') {
            console.error('Connection timeout after 30 seconds');
            updateState('failed');
            updateError('Connection timeout. The server took too long to respond.');
            cleanup();
          }
        }, 30000);

      } catch (err) {
        if (mountedRef.current) {
          let errorMsg = 'Unknown error occurred';

          if (err instanceof Error) {
            errorMsg = err.message;

            // Detect common network errors
            if (err.message.includes('Failed to fetch') || err.message.includes('NetworkError')) {
              errorMsg = 'Network error. Please check your internet connection.';
            } else if (err.message.includes('CORS')) {
              errorMsg = 'Security error. Please contact support.';
            }
          }

          console.error('❌ WebRTC playback error:', errorMsg, err);
          updateState('failed');
          updateError(errorMsg);
        }
      }
    }

    startPlayback();

    // Cleanup on unmount or dependency change
    return () => {
      mountedRef.current = false;
      cleanup();
    };
  }, [
    options.cameraId,
    options.playbackTime?.getTime(),
    options.skipGaps,
    options.speed,
    // NOTE: Do not include cleanup, updateState, updateError
    // They are stable via useCallback and including them causes infinite re-renders
  ]);

  /**
   * Manual stop function
   */
  const stop = useCallback(() => {
    updateState('disconnected');
    cleanup();
  }, [cleanup, updateState]);

  return {
    videoRef,
    state,
    error,
    stop,
    isConnecting: state === 'connecting',
    isConnected: state === 'connected',
    isFailed: state === 'failed',
  };
}
