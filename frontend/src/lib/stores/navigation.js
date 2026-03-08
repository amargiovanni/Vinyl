import { writable } from 'svelte/store';

export const currentView = writable('nowplaying');
export const selectedDate = writable(null);

export function navigateTo(view) {
  currentView.set(view);
}

export function navigateToDate(date) {
  selectedDate.set(date);
  currentView.set('diary');
}
