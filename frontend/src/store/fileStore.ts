import { create } from 'zustand';
import { api, type ApiResponse, type FileItem } from '../services/api';
import { ALLOWED_FILE_TYPES, MAX_FILE_SIZE } from '../config/index';

export interface UploadResult {
  uploaded: number;
  error?: string;
}

/** Which virtual root the file browser is currently operating in. */
export type FileScope = 'files' | 'shared';

/** Translate a browser-relative path into the virtual path the API expects. */
function scopePath(scope: FileScope, path: string): string {
  if (scope !== 'shared') return path;
  return path ? `shared/${path}` : 'shared';
}

interface FileState {
  /** Listing of the currently viewed directory. */
  files: FileItem[];
  /** Path of the currently viewed directory, relative to the active scope. */
  currentPath: string;
  /** Active virtual root (private home or the global shared folder). */
  scope: FileScope;
  /** Root-level listing used by the summary page (independent of navigation). */
  rootFiles: FileItem[];
  isLoading: boolean;
  error: string | null;
  uploadProgress: number; // 0..100
  isUploading: boolean;

  fetchFiles: (path?: string) => Promise<void>;
  refreshRoot: () => Promise<{ ok: boolean }>;
  /** Switch the active scope, resetting navigation to its root. */
  setScope: (scope: FileScope) => void;
  /** Navigate to a specific directory. */
  navigateTo: (path: string) => void;
  /** Navigate to the parent directory of the current path. */
  goUp: () => void;

  uploadFiles: (fileList: FileList, path?: string) => Promise<UploadResult>;
  createDirectory: (name: string) => Promise<ApiResponse>;
  deleteItem: (path: string) => Promise<ApiResponse>;

  clearError: () => void;
}

/** Validate a single file against the configured limits. Size is only enforced
 * when a positive limit is configured (0 = unlimited). */
function validateFile(file: File): { valid: boolean; error?: string } {
  if (MAX_FILE_SIZE > 0 && file.size > MAX_FILE_SIZE) {
    return { valid: false, error: `File size exceeds the limit of ${MAX_FILE_SIZE / (1024 * 1024)}MB` };
  }
  if (!ALLOWED_FILE_TYPES.includes(file.type)) {
    return { valid: false, error: `File type "${file.type || 'unknown'}" is not allowed` };
  }
  return { valid: true };
}

export const useFileStore = create<FileState>()((set, get) => ({
  files: [],
  currentPath: '',
  scope: 'files',
  rootFiles: [],
  isLoading: false,
  error: null,
  uploadProgress: 0,
  isUploading: false,

  fetchFiles: async (path = '') => {
    const { scope } = get();
    set({ isLoading: true, error: null });
    try {
      const { data, error } = await api.listFiles(scopePath(scope, path));
      if (error) throw new Error(error);
      set({ files: data || [], currentPath: path });
    } catch (e) {
      set({ error: e instanceof Error ? e.message : 'Failed to load files' });
    } finally {
      set({ isLoading: false });
    }
  },

  refreshRoot: async () => {
    try {
      const { data, error } = await api.listFiles('');
      if (error) throw new Error(error);
      set({ rootFiles: data || [], error: null });
      return { ok: true };
    } catch (e) {
      const message = e instanceof Error ? e.message : 'Failed to load summary';
      set({ error: message });
      return { ok: false };
    }
  },

  setScope: (scope) => {
    if (get().scope === scope) return;
    set({ scope, currentPath: '', files: [], error: null });
  },

  navigateTo: (path) => {
    get().fetchFiles(path);
  },

  goUp: () => {
    const current = get().currentPath;
    if (!current) return;
    const parent = current.split('/').slice(0, -1).join('/');
    get().fetchFiles(parent);
  },

  uploadFiles: async (fileList, path) => {
    if (!fileList || fileList.length === 0) return { uploaded: 0 };

    const files = Array.from(fileList);
    const invalid = files.map(validateFile).find((r) => !r.valid);
    if (invalid?.error) {
      set({ error: invalid.error });
      return { uploaded: 0, error: invalid.error };
    }

    const target = path ?? get().currentPath;
    const { scope } = get();
    set({ isUploading: true, uploadProgress: 0, error: null });

    try {
      for (let i = 0; i < files.length; i++) {
        const result = await api.uploadFile(files[i], scopePath(scope, target));
        if (result.error) throw new Error(result.error);
        set({ uploadProgress: Math.round(((i + 1) / files.length) * 100) });
      }
      await get().fetchFiles(target);
      return { uploaded: files.length };
    } catch (e) {
      const message = e instanceof Error ? e.message : 'Upload failed';
      set({ error: message });
      return { uploaded: 0, error: message };
    } finally {
      set({ uploadProgress: 0, isUploading: false });
    }
  },

  createDirectory: async (name) => {
    const trimmed = name.trim();
    if (!trimmed) return { error: 'Folder name is required' };
    const currentPath = get().currentPath;
    const fullPath = currentPath ? `${currentPath}/${trimmed}` : trimmed;
    set({ error: null });
    const result = await api.createDirectory(scopePath(get().scope, fullPath));
    if (!result.error) {
      await get().fetchFiles(currentPath);
    }
    return result;
  },

  deleteItem: async (path) => {
    set({ error: null });
    const result = await api.deletePath(path);
    if (!result.error) {
      await get().fetchFiles(get().currentPath);
    }
    return result;
  },

  clearError: () => set({ error: null }),
}));