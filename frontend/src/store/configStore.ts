import { create } from 'zustand';
import { MAX_CONCURRENT_UPLOADS } from '../config';

interface ConfigState {
  /** Current theme: 'light' | 'dark' | 'system' */
  theme: 'light' | 'dark' | 'system';
  /** Maximum concurrent uploads */
  maxConcurrentUploads: number;
  setTheme: (theme: 'light' | 'dark' | 'system') => void;
  setMaxConcurrentUploads: (n: number) => void;
}

export const useConfigStore = create<ConfigState>()((set) => ({
  theme: 'system',
  maxConcurrentUploads: MAX_CONCURRENT_UPLOADS,
  setTheme: (theme) => set({ theme }),
  setMaxConcurrentUploads: (n) => set({ maxConcurrentUploads: n }),
}));