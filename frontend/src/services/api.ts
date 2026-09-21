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

function authHeader(): Record<string, string> {
  const cred = loadCredentials();
  return cred ? { Authorization: `Basic ${cred}` } : {};
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
}

export interface AdminUser extends AuthUser {
  files: number;
  size: number;
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
function buildUrl(endpoint: string, params?: Record<string, string | number>): string {
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

export const api = {
  // File operations
  uploadFile: async (file: File, path: string = ''): Promise<ApiResponse<{ path: string; size: number }>> => {
    const formData = new FormData();
    formData.append('file', file);
    if (path) {
      formData.append('path', path);
    }

    return request(buildUrl('/api/upload'), {
      method: 'POST',
      headers: { ...authHeader() },
      body: formData,
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
      throw new Error('Not authorized');
    }
    throw new Error(`Failed to load (${response.status})`);
  }
  return response.blob();
}

export default api;
