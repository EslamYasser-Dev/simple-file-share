import { create } from 'zustand';
import { MAX_CONCURRENT_UPLOADS } from '../config';

export type ThemePreference = 'light' | 'dark' | 'system';

const THEME_STORAGE_KEY = 'fs_theme';
const LAN_SHARE_KEY = 'fs_lan_share';

function loadTheme(): ThemePreference {
  try {
    const stored = localStorage.getItem(THEME_STORAGE_KEY);
    if (stored === 'light' || stored === 'dark' || stored === 'system') return stored;
  } catch {
    /* storage unavailable - fall back to system preference */
  }
  return 'system';
}

function loadLanShare(): boolean {
  try {
    const stored = localStorage.getItem(LAN_SHARE_KEY);
    if (stored === '0' || stored === 'false') return false;
    if (stored === '1' || stored === 'true') return true;
  } catch {
    /* storage unavailable */
  }
  return true;
}

interface ConfigState {
  /** Current theme: 'light' | 'dark' | 'system' */
  theme: ThemePreference;
  /** Maximum concurrent uploads */
  maxConcurrentUploads: number;
  /** When false, LAN Share is hidden and the P2P signaling stream stays closed. */
  lanShareEnabled: boolean;
  setTheme: (theme: ThemePreference) => void;
  setMaxConcurrentUploads: (n: number) => void;
  setLanShareEnabled: (enabled: boolean) => void;
}

export const useConfigStore = create<ConfigState>()((set) => ({
  theme: loadTheme(),
  maxConcurrentUploads: MAX_CONCURRENT_UPLOADS,
  lanShareEnabled: loadLanShare(),
  setTheme: (theme) => {
    try {
      localStorage.setItem(THEME_STORAGE_KEY, theme);
    } catch {
      /* ignore storage errors */
    }
    set({ theme });
  },
  setMaxConcurrentUploads: (n) => set({ maxConcurrentUploads: n }),
  setLanShareEnabled: (lanShareEnabled) => {
    try {
      localStorage.setItem(LAN_SHARE_KEY, lanShareEnabled ? '1' : '0');
    } catch {
      /* ignore storage errors */
    }
    set({ lanShareEnabled });
  },
}));
