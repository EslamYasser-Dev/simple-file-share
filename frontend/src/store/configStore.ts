import { create } from 'zustand';
import { MAX_CONCURRENT_UPLOADS } from '../config';

export type ThemePreference = 'light' | 'dark' | 'system';

const THEME_STORAGE_KEY = 'fs_theme';

function loadTheme(): ThemePreference {
  try {
    const stored = localStorage.getItem(THEME_STORAGE_KEY);
    if (stored === 'light' || stored === 'dark' || stored === 'system') return stored;
  } catch {
    /* storage unavailable - fall back to system preference */
  }
  return 'system';
}

interface ConfigState {
  /** Current theme: 'light' | 'dark' | 'system' */
  theme: ThemePreference;
  /** Maximum concurrent uploads */
  maxConcurrentUploads: number;
  setTheme: (theme: ThemePreference) => void;
  setMaxConcurrentUploads: (n: number) => void;
}

export const useConfigStore = create<ConfigState>()((set) => ({
  theme: loadTheme(),
  maxConcurrentUploads: MAX_CONCURRENT_UPLOADS,
  setTheme: (theme) => {
    try {
      localStorage.setItem(THEME_STORAGE_KEY, theme);
    } catch {
      /* ignore storage errors */
    }
    set({ theme });
  },
  setMaxConcurrentUploads: (n) => set({ maxConcurrentUploads: n }),
}));