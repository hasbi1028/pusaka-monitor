// Theme store — light/dark, persisted, follows existing writable convention
import { writable } from 'svelte/store';

const KEY = 'pusaka-theme';

function initial() {
  if (typeof localStorage !== 'undefined') {
    const saved = localStorage.getItem(KEY);
    if (saved === 'light' || saved === 'dark') return saved;
  }
  if (typeof window !== 'undefined' && window.matchMedia?.('(prefers-color-scheme: dark)').matches) {
    return 'dark';
  }
  return 'light';
}

function apply(mode) {
  if (typeof document !== 'undefined') {
    document.documentElement.classList.toggle('dark', mode === 'dark');
  }
  if (typeof localStorage !== 'undefined') {
    localStorage.setItem(KEY, mode);
  }
}

function createThemeStore() {
  const { subscribe, set } = writable('light');

  return {
    subscribe,
    init() {
      const mode = initial();
      apply(mode);
      set(mode);
    },
    toggle() {
      let next = 'light';
      const unsub = subscribe(current => {
        next = current === 'dark' ? 'light' : 'dark';
      });
      unsub();
      apply(next);
      set(next);
    }
  };
}

export const theme = createThemeStore();
