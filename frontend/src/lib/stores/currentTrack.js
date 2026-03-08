import { writable } from 'svelte/store';

export const currentTrack = writable(null);
export const currentSession = writable(null);
export const isPlaying = writable(false);
export const currentWeather = writable(null);

export function initTrackEvents() {
  if (typeof window !== 'undefined' && window.runtime) {
    window.runtime.EventsOn('track:changed', (track) => {
      currentTrack.set(track);
      isPlaying.set(true);
    });
    window.runtime.EventsOn('track:updated', (track) => {
      currentTrack.set(track);
    });
    window.runtime.EventsOn('playback:started', (track) => {
      if (track) currentTrack.set(track);
      isPlaying.set(true);
    });
    window.runtime.EventsOn('playback:stopped', () => {
      isPlaying.set(false);
    });
    window.runtime.EventsOn('session:started', (session) => {
      currentSession.set(session);
    });
    window.runtime.EventsOn('session:ended', () => {
      currentSession.set(null);
      currentTrack.set(null);
      isPlaying.set(false);
    });
  }
}
