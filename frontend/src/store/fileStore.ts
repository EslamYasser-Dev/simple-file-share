import { create } from 'zustand';
import { api, type ApiResponse, type FileItem } from '../services/api';
import { ALLOWED_FILE_EXTENSIONS, ALLOWED_FILE_TYPES, MAX_FILE_SIZE } from '../config/index';
import { useConfigStore } from './configStore';

export interface UploadResult {
  uploaded: number;
  error?: string;
  /** True when the user cancelled the upload (not an error). */
  cancelled?: boolean;
}

/** Module-level handle to the in-flight upload so `cancelUpload` can abort it. */
let activeUploadController: AbortController | null = null;

/** Live metadata for one file in the batch currently being uploaded. */
export interface ActiveUploadFile {
  name: string;
  size: number;
  loaded: number;
  status: 'queued' | 'active' | 'done';
}

/** Live metadata for the batch currently being uploaded. */
export interface ActiveUpload {
  totalFiles: number;
  completedFiles: number;
  totalBytes: number;
  loadedBytes: number;
  files: ActiveUploadFile[];
}

/** Which virtual root the file browser is currently operating in. */
export type FileScope = 'files' | 'shared';

/** Translate a browser-relative path into the virtual path the API expects. */
function scopePath(scope: FileScope, path: string): string {
  if (scope !== 'shared') return path;
  return path ? `shared/${path}` : 'shared';
}

function hasDuplicate(values: string[]): boolean {
  return new Set(values).size !== values.length;
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
  /** Smoothed upload rate in bytes per second (0 when idle). */
  uploadSpeed: number;
  isUploading: boolean;
  /** Details shown by the floating upload indicator; null when idle. */
  activeUpload: ActiveUpload | null;

  fetchFiles: (path?: string) => Promise<void>;
  refreshRoot: () => Promise<{ ok: boolean }>;
  /** Switch the active scope, resetting navigation to its root. */
  setScope: (scope: FileScope) => void;
  /** Navigate to a specific directory. */
  navigateTo: (path: string) => void;
  /** Navigate to the parent directory of the current path. */
  goUp: () => void;

  uploadFiles: (fileList: FileList, path?: string) => Promise<UploadResult>;
  /** Abort the in-flight upload, if any. */
  cancelUpload: () => void;
  createDirectory: (name: string) => Promise<ApiResponse>;
  deleteItem: (path: string) => Promise<ApiResponse>;

  clearError: () => void;
}

/** Lower-cased extension without the dot ('' when the name has none). */
function fileExtension(name: string): string {
  const dot = name.lastIndexOf('.');
  return dot >= 0 ? name.slice(dot + 1).toLowerCase() : '';
}

/** Validate a single file against the configured limits. Size is only enforced
 * when a positive limit is configured (0 = unlimited). A file is accepted when
 * either its MIME type or its extension is allowlisted, so formats the browser
 * cannot type (e.g. `.iso`, often reported as "") still upload. */
function validateFile(file: File): { valid: boolean; error?: string } {
  if (MAX_FILE_SIZE > 0 && file.size > MAX_FILE_SIZE) {
    return { valid: false, error: `File size exceeds the limit of ${MAX_FILE_SIZE / (1024 * 1024)}MB` };
  }
  const ext = fileExtension(file.name);
  const allowed = ALLOWED_FILE_TYPES.includes(file.type) || ALLOWED_FILE_EXTENSIONS.includes(ext);
  if (!allowed) {
    return { valid: false, error: `File type "${file.type || ext || 'unknown'}" is not allowed` };
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
  uploadSpeed: 0,
  isUploading: false,
  activeUpload: null,

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

    const controller = new AbortController();
    activeUploadController = controller;

    // Upload distinct destinations concurrently, but keep duplicate target
    // paths sequential so concurrent writes cannot race on the same file.
    const destinations = files.map((file) =>
      scopePath(scope, target ? `${target}/${file.name}` : file.name),
    );
    const concurrency = hasDuplicate(destinations)
      ? 1
      : Math.min(useConfigStore.getState().maxConcurrentUploads, files.length);

    // Track progress across the whole batch by bytes, so a single large file
    // (and mixed-size batches) reports real progress instead of jumping 0→100.
    const totalBytes = files.reduce((sum, file) => sum + file.size, 0);
    const loadedBytes = files.map(() => 0);
    const statuses = files.map(() => 'queued' as ActiveUploadFile['status']);
    let completedCount = 0;
    let firstError: Error | null = null;
    let userCancelled = false;

    // Speed sampling: compare bytes/time between progress events, throttled and
    // exponentially smoothed so the displayed rate doesn't jitter.
    let lastTime = Date.now();
    let lastBytes = 0;
    let smoothedSpeed = 0;

    const publish = () => {
      const globalLoaded = files.reduce(
        (sum, file, index) => sum + Math.min(loadedBytes[index] ?? 0, file.size),
        0,
      );

      const now = Date.now();
      const elapsed = now - lastTime;
      if (elapsed >= 250) {
        const instant = ((globalLoaded - lastBytes) * 1000) / elapsed;
        smoothedSpeed = smoothedSpeed > 0 ? smoothedSpeed * 0.7 + instant * 0.3 : instant;
        lastTime = now;
        lastBytes = globalLoaded;
      }

      const percent =
        totalBytes > 0
          ? (globalLoaded / totalBytes) * 100
          : (completedCount / files.length) * 100;

      set({
        isUploading: true,
        uploadProgress: Math.min(100, Math.round(percent)),
        uploadSpeed: Math.max(0, smoothedSpeed),
        error: null,
        activeUpload: {
          totalFiles: files.length,
          completedFiles: completedCount,
          totalBytes,
          loadedBytes: Math.round(globalLoaded),
          files: files.map((file, index) => ({
            name: file.name,
            size: file.size,
            loaded: Math.round(Math.min(loadedBytes[index] ?? 0, file.size)),
            status: statuses[index] ?? 'queued',
          })),
        },
      });
    };

    const uploadOne = async (index: number): Promise<void> => {
      statuses[index] = 'active';
      publish();

      try {
        const result = await api.uploadFile(files[index], destinations[index], {
          signal: controller.signal,
          onProgress: (loaded) => {
            loadedBytes[index] = loaded;
            publish();
          },
        });
        if (result.error) throw new Error(result.error);
        loadedBytes[index] = files[index].size;
        statuses[index] = 'done';
        completedCount += 1;
        lastBytes = files.reduce(
          (sum, file, fileIndex) => sum + Math.min(loadedBytes[fileIndex] ?? 0, file.size),
          0,
        );
        publish();
      } catch (error) {
        const message = error instanceof Error ? error.message : 'Upload failed';
        if (message === 'Upload cancelled') {
          userCancelled = true;
        } else if (!firstError) {
          firstError = error instanceof Error ? error : new Error(message);
        }
        controller.abort();
      }
    };

    let pendingIndex = 0;
    const worker = async (): Promise<void> => {
      while (!firstError && !userCancelled) {
        const index = pendingIndex;
        pendingIndex += 1;
        if (index >= files.length) return;
        await uploadOne(index);
      }
    };

    publish();

    try {
      await Promise.all(
        Array.from({ length: Math.min(concurrency, files.length) }, () => worker()),
      );
      if (userCancelled) {
        return { uploaded: 0, cancelled: true };
      }
      if (firstError) {
        throw firstError;
      }
      await get().fetchFiles(target);
      return { uploaded: completedCount };
    } catch (e) {
      const message = e instanceof Error ? e.message : 'Upload failed';
      if (message === 'Upload cancelled') {
        return { uploaded: 0, cancelled: true };
      }
      set({ error: message });
      return { uploaded: 0, error: message };
    } finally {
      activeUploadController = null;
      set({ uploadProgress: 0, uploadSpeed: 0, isUploading: false, activeUpload: null });
    }
  },

  cancelUpload: () => {
    activeUploadController?.abort();
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