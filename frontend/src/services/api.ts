import { API_BASE_URL } from '@/config';

// ---------------------------------------------------------------------------
// Basic-auth credential store (kept in sessionStorage for the tab lifetime)
// ---------------------------------------------------------------------------
const CRED_KEY = 'fs_credentials';

function loadCredentials(): string | null {
  try {
    return sessionStorage.getItem(CRED_KEY);
  } catch {
    return null;
  }
}

export function setCredentials(username: string, password: string): void {
  try {
    sessionStorage.setItem(CRED_KEY, btoa(`${username}:${password}`));
  } catch {
    /* storage unavailable - credentials stay in memory only */
  }
}

export function clearCredentials(): void {
  try {
    sessionStorage.removeItem(CRED_KEY);
  } catch {
    /* ignore */
  }
}

export function authHeader(): Record<string, string> {
  const cred = loadCredentials();
  return cred ? { Authorization: `Basic ${cred}` } : {};
}

/** Tell the auth gate that the session is no longer accepted by the server. */
function notifySessionExpired(): void {
  try {
    window.dispatchEvent(new CustomEvent('fs:unauthorized'));
  } catch {
    /* non-browser environment */
  }
}

export interface FileItem {
  name: string;
  path: string;
  size: number;
  isDir: boolean;
  modified: string;
  mimeType?: string;
}

export interface AuthUser {
  username: string;
  isAdmin: boolean;
  createdAt?: string;
  /** Account storage quota in bytes; 0 or undefined means unlimited. */
  quotaBytes?: number;
  /** Bytes currently used by the account (from /me). */
  size?: number;
  files?: number;
}

export interface AdminUser extends AuthUser {
  files: number;
  size: number;
  quotaBytes: number;
}

/** A public share link, as returned by the shares API. */
export interface ShareItem {
  token: string;
  path: string;
  name: string;
  owner: string;
  createdAt: string;
  expiresAt: string;
}

export interface ApiResponse<T = unknown> {
  data?: T;
  error?: string;
  /** True when the server rejected the request due to missing/bad credentials. */
  unauthorized?: boolean;
}

/**
 * Resolve an API endpoint against the configured base URL. When `VITE_API_URL`
 * is unset the app runs on the same origin as the API, so relative paths must
 * be resolved against `window.location.origin` (a bare `new URL('/api')` throws).
 */
export function buildUrl(endpoint: string, params?: Record<string, string | number>): string {
  const base = (API_BASE_URL || window.location.origin).replace(/\/+$/, '');
  const url = new URL(`${base}${endpoint}`);
  if (params) {
    for (const [key, value] of Object.entries(params)) {
      url.searchParams.set(key, String(value));
    }
  }
  return url.toString();
}

async function handleResponse<T>(response: Response): Promise<ApiResponse<T>> {
  if (!response.ok) {
    if (response.status === 401) {
      clearCredentials();
      notifySessionExpired();
      return { error: 'Invalid username or password', unauthorized: true };
    }
    const body = await response.text().catch(() => '');
    let message = `Request failed (${response.status})`;
    if (body) {
      try {
        const parsed = JSON.parse(body) as { error?: string; message?: string };
        message = parsed.error || parsed.message || body;
      } catch {
        message = body;
      }
    }
    return { error: message };
  }

  const text = await response.text();
  if (!text) {
    return {};
  }
  try {
    return { data: JSON.parse(text) as T };
  } catch {
    return { error: 'Failed to parse response' };
  }
}

/**
 * fetch + error normalisation. Network failures resolve to `{ error }` instead
 * of rejecting, so callers can never hang on an unhandled promise rejection.
 */
async function request<T>(url: string, init?: RequestInit): Promise<ApiResponse<T>> {
  let response: Response;
  try {
    response = await fetch(url, init);
  } catch (e) {
    return { error: e instanceof Error ? e.message : 'Network request failed' };
  }
  return handleResponse<T>(response);
}

/** Normalise an XHR response into an `ApiResponse`, mirroring `handleResponse`. */
function parseXhrResponse<T>(xhr: XMLHttpRequest): ApiResponse<T> {
  if (xhr.status === 401) {
    clearCredentials();
    notifySessionExpired();
    return { error: 'Invalid username or password', unauthorized: true };
  }
  if (xhr.status < 200 || xhr.status >= 300) {
    const body = xhr.responseText || '';
    let message = `Request failed (${xhr.status})`;
    if (body) {
      try {
        const parsed = JSON.parse(body) as { error?: string; message?: string };
        message = parsed.error || parsed.message || body;
      } catch {
        message = body;
      }
    }
    return { error: message };
  }
  if (!xhr.responseText) {
    return {};
  }
  try {
    return { data: JSON.parse(xhr.responseText) as T };
  } catch {
    return { error: 'Failed to parse response' };
  }
}

export const api = {
  // File operations
  //
  // Uploaded with XMLHttpRequest rather than fetch: fetch cannot report request
  // upload progress, so the UI progress bar would sit at 0% until completion.
  // `options.onProgress` receives bytes sent / total for the request body, and
  // `options.signal` aborts the upload (resolving with `{ error }`).
  uploadFile: (
    file: File,
    path: string = '',
    options: {
      onProgress?: (loaded: number, total: number) => void;
      signal?: AbortSignal;
    } = {},
  ): Promise<ApiResponse<Array<{ path: string; size: number }>>> => {
    const formData = new FormData();
    if (path) {
      formData.append('path', path);
    }
    formData.append('file', file);

    return new Promise((resolve) => {
      const xhr = new XMLHttpRequest();
      let settled = false;
      let removeAbortListener: (() => void) | undefined;

      const settle = (result: ApiResponse<Array<{ path: string; size: number }>>) => {
        if (settled) return;
        settled = true;
        removeAbortListener?.();
        resolve(result);
      };

      const abort = () => xhr.abort();

      xhr.open('POST', buildUrl('/api/upload'));
      for (const [key, value] of Object.entries(authHeader())) {
        xhr.setRequestHeader(key, value);
      }

      xhr.upload.onprogress = (event) => {
        if (options.onProgress && event.lengthComputable) {
          options.onProgress(event.loaded, event.total);
        }
      };
      xhr.onload = () => settle(parseXhrResponse<Array<{ path: string; size: number }>>(xhr));
      xhr.onerror = () => settle({ error: 'Network request failed' });
      xhr.onabort = () => settle({ error: 'Upload cancelled' });

      if (options.signal) {
        if (options.signal.aborted) {
          settle({ error: 'Upload cancelled' });
          return;
        }
        options.signal.addEventListener('abort', abort);
        removeAbortListener = () => options.signal?.removeEventListener('abort', abort);
      }

      xhr.send(formData);
    });
  },

  // File listing
  listFiles: async (path: string = ''): Promise<ApiResponse<FileItem[]>> =>
    request<FileItem[]>(buildUrl('/api/files', path ? { path } : undefined), {
      headers: authHeader(),
    }),

  searchFiles: async (query: string, limit = 50): Promise<ApiResponse<FileItem[]>> =>
    request<FileItem[]>(buildUrl('/api/files/search', { q: query, limit }), {
      headers: authHeader(),
    }),

  // Create directory
  createDirectory: async (path: string): Promise<ApiResponse> =>
    request(buildUrl('/api/directories'), {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', ...authHeader() },
      body: JSON.stringify({ path }),
    }),

  // Delete file or directory
  deletePath: async (path: string): Promise<ApiResponse> =>
    request(buildUrl('/api/files'), {
      method: 'DELETE',
      headers: { 'Content-Type': 'application/json', ...authHeader() },
      body: JSON.stringify({ path }),
    }),

  // Download file (or ZIP archive when the path points at a directory)
  downloadFile: async (path: string): Promise<Blob> => fetchBlob('/api/files/download', { path }),

  // Fetch a file's bytes rendered inline (used by the in-app viewer)
  viewFile: async (path: string): Promise<Blob> => fetchBlob('/api/files/view', { path }),

  // Overwrite a text file's contents (markdown editor)
  updateFileText: async (path: string, content: string): Promise<ApiResponse<{ size: number }>> =>
    request<{ size: number }>(buildUrl('/api/files/content'), {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json', ...authHeader() },
      body: JSON.stringify({ path, content }),
    }),

  // Get file info
  getFileInfo: async (path: string): Promise<ApiResponse<FileItem>> =>
    request<FileItem>(buildUrl('/api/files/info', { path }), {
      headers: authHeader(),
    }),

  // Authentication / accounts
  register: async (username: string, password: string): Promise<ApiResponse<AuthUser>> =>
    request<AuthUser>(buildUrl('/api/auth/register'), {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password }),
    }),

  authInfo: async (): Promise<ApiResponse<{ signupEnabled: boolean }>> =>
    request<{ signupEnabled: boolean }>(buildUrl('/api/auth/info')),

  me: async (): Promise<ApiResponse<AuthUser>> =>
    request<AuthUser>(buildUrl('/api/auth/me'), { headers: authHeader() }),

  adminUsers: async (): Promise<ApiResponse<AdminUser[]>> =>
    request<AdminUser[]>(buildUrl('/api/admin/users'), { headers: authHeader() }),

  /**
   * Set an account's storage quota. Accepts a byte count, a human size
   * ("2GB"), or "unlimited"; returns the updated account.
   */
  setQuota: async (username: string, quota: string): Promise<ApiResponse<AdminUser>> =>
    request<AdminUser>(buildUrl(`/api/admin/users/${encodeURIComponent(username)}/quota`), {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json', ...authHeader() },
      body: JSON.stringify({ quota }),
    }),

  // Public links
  //
  // Shared links are plain GET URLs anyone (even unauthenticated) can open, so
  // `shareUrl` is the public-shaped URL, unlike authenticated blob fetches.

  /** Create a share link. `expiresInSeconds` of 0 means no expiry. */
  createShare: async (path: string, expiresInSeconds: number): Promise<ApiResponse<ShareItem>> =>
    request<ShareItem>(buildUrl('/api/shares'), {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', ...authHeader() },
      body: JSON.stringify({ path, expiresInSeconds }),
    }),

  /** List every share link visible to the current user. */
  listShares: async (): Promise<ApiResponse<ShareItem[]>> =>
    request<ShareItem[]>(buildUrl('/api/shares'), { headers: authHeader() }),

  /** Revoke a share link by token. */
  revokeShare: async (token: string): Promise<ApiResponse> =>
    request(buildUrl('/api/shares'), {
      method: 'DELETE',
      headers: { 'Content-Type': 'application/json', ...authHeader() },
      body: JSON.stringify({ token }),
    }),

  /** Build the public URL for a share token. */
  shareUrl: (token: string): string => buildUrl(`/api/share/${token}`),
};

/** Fetch a raw binary response with auth, throwing descriptive errors. */
async function fetchBlob(endpoint: string, params: Record<string, string>): Promise<Blob> {
  let response: Response;
  try {
    response = await fetch(buildUrl(endpoint, params), {
      headers: authHeader(),
    });
  } catch (e) {
    throw new Error(e instanceof Error ? e.message : 'Network request failed');
  }
  if (!response.ok) {
    if (response.status === 401) {
      clearCredentials();
      notifySessionExpired();
      throw new Error('Not authorized');
    }
    throw new Error(`Failed to load (${response.status})`);
  }
  return response.blob();
}

export default api;
