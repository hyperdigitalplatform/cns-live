import { create } from 'zustand';
import type { StreamReservation, StreamStats } from '@/types';
import { api } from '@/services/api';

interface StreamState {
  reservations: Map<string, StreamReservation>;
  stats: StreamStats | null;
  loading: boolean;
  error: string | null;

  // Actions
  reserveStream: (
    cameraId: string,
    quality?: 'high' | 'medium' | 'low'
  ) => Promise<StreamReservation>;
  releaseStream: (reservationId: string) => Promise<void>;
  sendHeartbeat: (reservationId: string) => Promise<void>;
  fetchStats: () => Promise<void>;
  startHeartbeat: (reservationId: string) => void;
  stopHeartbeat: (reservationId: string) => void;
}

// Heartbeat intervals for each reservation
const heartbeatIntervals = new Map<string, NodeJS.Timeout>();

export const useStreamStore = create<StreamState>((set, get) => ({
  reservations: new Map(),
  stats: null,
  loading: false,
  error: null,

  reserveStream: async (cameraId, quality = 'medium') => {
    set({ loading: true, error: null });
    try {
      const reservation = await api.reserveStream(cameraId, quality);
      const reservations = new Map(get().reservations);
      reservations.set(reservation.reservation_id, reservation);
      set({ reservations, loading: false });

      // Start heartbeat for this reservation
      get().startHeartbeat(reservation.reservation_id);

      return reservation;
    } catch (error) {
      set({
        error:
          error instanceof Error ? error.message : 'Failed to reserve stream',
        loading: false,
      });
      throw error;
    }
  },

  releaseStream: async (reservationId) => {
    try {
      await api.releaseStream(reservationId);
      const reservations = new Map(get().reservations);
      reservations.delete(reservationId);
      set({ reservations });

      // Stop heartbeat
      get().stopHeartbeat(reservationId);
    } catch (error) {
      console.error('Failed to release stream:', error);
    }
  },

  sendHeartbeat: async (reservationId) => {
    try {
      await api.sendHeartbeat(reservationId);
    } catch (error) {
      console.error('Heartbeat failed:', error);
      // If heartbeat fails, the reservation might have expired
      const reservations = new Map(get().reservations);
      reservations.delete(reservationId);
      set({ reservations });
      get().stopHeartbeat(reservationId);
    }
  },

  fetchStats: async () => {
    try {
      const stats = await api.getStreamStats();
      set({ stats });
    } catch (error) {
      console.error('Failed to fetch stats:', error);
    }
  },

  startHeartbeat: (reservationId) => {
    // Clear existing interval if any
    if (heartbeatIntervals.has(reservationId)) {
      clearInterval(heartbeatIntervals.get(reservationId)!);
    }

    // Send heartbeat every 25 seconds (server expects 30s)
    const interval = setInterval(() => {
      get().sendHeartbeat(reservationId);
    }, 25000);

    heartbeatIntervals.set(reservationId, interval);
  },

  stopHeartbeat: (reservationId) => {
    const interval = heartbeatIntervals.get(reservationId);
    if (interval) {
      clearInterval(interval);
      heartbeatIntervals.delete(reservationId);
    }
  },
}));
