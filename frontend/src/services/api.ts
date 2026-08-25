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

export interface ApiResponse<T = unknown> {
  data?: T;
  error?: string;
  /** True when the server rejected the request due to missing/bad credentials. */
  unauthorized?: boolean;
}

async function handleResponse<T>(response: Response): Promise<ApiResponse<T>> {
  if (!response.ok) {
    if (response.status === 401) {
      clearCredentials();
      return { error: 'Invalid username or password', unauthorized: true };
    }
    const error = await response.text().catch(() => '');
    return { error: error || `Request failed (${response.status})` };
  }

  try {
    const data = await response.json();
    return { data };
  } catch {
    return { error: 'Failed to parse response' };
  }
}

export const api = {
  // File operations
  uploadFile: async (file: File, path: string = ''): Promise<ApiResponse<{ path: string; size: number }>> => {
    const formData = new FormData();
    formData.append('file', file);
    if (path) {
      formData.append('path', path);
    }

    const response = await fetch(`${API_BASE_URL}/api/upload`, {
      method: 'POST',
      headers: { ...authHeader() },
      body: formData,
    });

    return handleResponse(response);
  },

  // File listing
  listFiles: async (path: string = ''): Promise<ApiResponse<Array<{
    name: string;
    path: string;
    size: number;
    isDir: boolean;
    modified: string;
  }>>> => {
    const url = new URL(`${API_BASE_URL}/api/files`);
    if (path) {
      url.searchParams.append('path', path);
    }

    const response = await fetch(url.toString(), { headers: authHeader() });
    return handleResponse(response);
  },

  searchFiles: async (query: string, limit = 50): Promise<ApiResponse<FileItem[]>> => {
    const url = new URL(`${API_BASE_URL}/api/files/search`);
    url.searchParams.append('q', query);
    url.searchParams.append('limit', String(limit));

    const response = await fetch(url.toString(), { headers: authHeader() });
    return handleResponse(response);
  },

  // Create directory
  createDirectory: async (path: string): Promise<ApiResponse> => {
    const response = await fetch(`${API_BASE_URL}/api/directories`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', ...authHeader() },
      body: JSON.stringify({ path }),
    });

    return handleResponse(response);
  },

  // Delete file or directory
  deletePath: async (path: string): Promise<ApiResponse> => {
    const response = await fetch(`${API_BASE_URL}/api/files`, {
      method: 'DELETE',
      headers: { 'Content-Type': 'application/json', ...authHeader() },
      body: JSON.stringify({ path }),
    });

    return handleResponse(response);
  },

  // Download file (or ZIP archive when the path points at a directory)
  downloadFile: async (path: string): Promise<Blob> => {
    const response = await fetch(`${API_BASE_URL}/api/files/download?path=${encodeURIComponent(path)}`, {
      headers: authHeader(),
    });
    if (!response.ok) {
      throw new Error(response.status === 401 ? 'Not authorized' : 'Failed to download');
    }
    return response.blob();
  },

  // Get file info
  getFileInfo: async (path: string): Promise<ApiResponse<{
    name: string;
    path: string;
    size: number;
    isDir: boolean;
    modified: string;
    mimeType: string;
  }>> => {
    const response = await fetch(`${API_BASE_URL}/api/files/info?path=${encodeURIComponent(path)}`, {
      headers: authHeader(),
    });
    return handleResponse(response);
  },
};

export default api;
