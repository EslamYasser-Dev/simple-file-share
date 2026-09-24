import { API_BASE_URL } from '../config';
import { clearToken, getToken, setToken } from './token';

export interface FileItem {
  name: string;
  path: string;
  size: number;
  isDir: boolean;
  modified: string;
  mimeType?: string;
  version?: number;
}

export interface AuthUser {
  username: string;
  isAdmin: boolean;
  role?: string;
  enabled?: boolean;
  permissions?: string[];
  createdAt?: string;
  quotaBytes?: number;
  size?: number;
  files?: number;
}

export interface ShareItem {
  token: string;
  path: string;
  name: string;
  owner: string;
  createdAt: string;
  expiresAt: string;
}

export interface TokenResponse {
  accessToken: string;
  tokenType: string;
  expiresIn: number;
}

export interface AuthInfo {
  signupEnabled: boolean;
  oauth?: string[];
}

export interface ApiResponse<T = unknown> {
  data?: T;
  error?: string;
  unauthorized?: boolean;
}

type UnauthorizedListener = () => void;

let unauthorizedListener: UnauthorizedListener | null = null;

export function onUnauthorized(listener: UnauthorizedListener | null): void {
  unauthorizedListener = listener;
}

export function buildUrl(endpoint: string, params?: Record<string, string | number>): string {
  const url = new URL(`${API_BASE_URL}${endpoint}`);
  if (params) {
    for (const [key, value] of Object.entries(params)) {
      url.searchParams.set(key, String(value));
    }
  }
  return url.toString();
}

async function authHeaders(extra?: Record<string, string>): Promise<Record<string, string>> {
  const headers: Record<string, string> = { ...extra };
  const token = await getToken();
  if (token) headers.Authorization = `Bearer ${token}`;
  return headers;
}

async function handleResponse<T>(response: Response): Promise<ApiResponse<T>> {
  if (!response.ok) {
    if (response.status === 401) {
      await clearToken();
      unauthorizedListener?.();
      return { error: 'Not authorized', unauthorized: true };
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
  if (!text) return {};
  try {
    return { data: JSON.parse(text) as T };
  } catch {
    return { error: 'Failed to parse response' };
  }
}

async function request<T>(endpoint: string, init?: RequestInit): Promise<ApiResponse<T>> {
  let response: Response;
  try {
    response = await fetch(buildUrl(endpoint), {
      ...init,
      headers: await authHeaders(
        (init?.headers as Record<string, string> | undefined) ?? undefined,
      ),
    });
  } catch (e) {
    return { error: e instanceof Error ? e.message : 'Network request failed' };
  }
  return handleResponse<T>(response);
}

interface UploadSessionInfo {
  id: string;
  offset: number;
  size: number;
  chunkSize: number;
}

/** XHR upload with progress; resolves parse-style ApiResponse. */
function xhrUpload(
  method: string,
  url: string,
  options: {
    body?: BodyInit | null;
    headers?: Record<string, string>;
    onUploadProgress?: (loaded: number, total: number) => void;
  } = {},
): Promise<ApiResponse<unknown> & { status?: number; expectedOffset?: number }> {
  return new Promise((resolve) => {
    void (async () => {
      const token = await getToken();
      const xhr = new XMLHttpRequest();
      let settled = false;
      const settle = (result: ApiResponse<unknown> & { status?: number; expectedOffset?: number }) => {
        if (settled) return;
        settled = true;
        resolve(result);
      };

      xhr.open(method, url);
      if (token) xhr.setRequestHeader('Authorization', `Bearer ${token}`);
      if (options.headers) {
        for (const [k, v] of Object.entries(options.headers)) xhr.setRequestHeader(k, v);
      }
      if (options.onUploadProgress) {
        xhr.upload.onprogress = (event) => {
          if (event.lengthComputable) options.onUploadProgress!(event.loaded, event.total);
        };
      }
      xhr.onload = () => {
        if (xhr.status === 401) {
          void clearToken();
          unauthorizedListener?.();
          settle({ error: 'Not authorized', unauthorized: true, status: 401 });
          return;
        }
        const extra: { status: number; expectedOffset?: number } = { status: xhr.status };
        if (xhr.status === 409) {
          try {
            const body = JSON.parse(xhr.responseText || '{}') as { expected?: number };
            if (typeof body.expected === 'number') extra.expectedOffset = body.expected;
          } catch {
            /* ignore */
          }
        }
        if (xhr.status < 200 || xhr.status >= 300) {
          let message = `Request failed (${xhr.status})`;
          try {
            const parsed = JSON.parse(xhr.responseText || '') as { error?: string };
            message = parsed.error || message;
          } catch {
            /* keep default */
          }
          settle({ error: message, ...extra });
          return;
        }
        if (!xhr.responseText) {
          settle({ ...extra });
          return;
        }
        try {
          settle({ data: JSON.parse(xhr.responseText), ...extra });
        } catch {
          settle({ error: 'Failed to parse response', ...extra });
        }
      };
      xhr.onerror = () => settle({ error: 'Network request failed' });
      xhr.onabort = () => settle({ error: 'Upload cancelled' });
      xhr.send(options.body ?? null);
    })();
  });
}

async function uploadResumableBlob(
  blob: Blob,
  fileName: string,
  path: string,
  onProgress?: (loaded: number, total: number) => void,
): Promise<ApiResponse<{ path: string; size: number }[]>> {
  const fingerprint = `${path}|${fileName}|${blob.size}`;
  const created = await xhrUpload('POST', buildUrl('/api/uploads'), {
    body: JSON.stringify({ path, filename: fileName, size: blob.size, fingerprint }),
    headers: { 'Content-Type': 'application/json' },
  });
  if (created.error || !created.data) {
    return created as ApiResponse<{ path: string; size: number }[]>;
  }
  const session = created.data as UploadSessionInfo;
  const chunkSize = Math.max(1, session.chunkSize || 4 * 1024 * 1024);
  let offset = Math.min(Math.max(0, session.offset), blob.size);
  onProgress?.(offset, blob.size);

  while (offset < blob.size) {
    const end = Math.min(offset + chunkSize, blob.size);
    const base = offset;
    const chunk = await xhrUpload('PATCH', buildUrl(`/api/uploads/${session.id}`), {
      body: blob.slice(offset, end),
      headers: {
        'Content-Type': 'application/octet-stream',
        'Upload-Offset': String(offset),
      },
      onUploadProgress: (loaded) => onProgress?.(base + loaded, blob.size),
    });
    if (chunk.status === 409 && typeof chunk.expectedOffset === 'number') {
      offset = Math.min(Math.max(0, chunk.expectedOffset), blob.size);
      onProgress?.(offset, blob.size);
      continue;
    }
    if (chunk.error) return chunk as ApiResponse<{ path: string; size: number }[]>;
    if (chunk.data) {
      offset = Math.min((chunk.data as UploadSessionInfo).offset, blob.size);
      onProgress?.(offset, blob.size);
    } else {
      offset = end;
      onProgress?.(offset, blob.size);
    }
  }

  const done = await xhrUpload('POST', buildUrl(`/api/uploads/${session.id}/complete`));
  if (done.error) return done as ApiResponse<{ path: string; size: number }[]>;
  return { data: done.data ? [done.data as { path: string; size: number }] : [] };
}

async function uploadMultipartBlob(
  blob: Blob,
  fileName: string,
  mimeType: string,
  path: string,
  onProgress?: (loaded: number, total: number) => void,
): Promise<ApiResponse<{ path: string; size: number }[]>> {
  const form = new FormData();
  if (path) form.append('path', path);
  form.append('file', blob, fileName);
  void mimeType;
  const result = await xhrUpload('POST', buildUrl('/api/upload'), {
    body: form,
    onUploadProgress: onProgress,
  });
  return result as ApiResponse<{ path: string; size: number }[]>;
}

async function uploadMultipartUri(
  uri: string,
  fileName: string,
  mimeType: string,
  path: string,
  onProgress?: (loaded: number, total: number) => void,
): Promise<ApiResponse<{ path: string; size: number }[]>> {
  const form = new FormData();
  if (path) form.append('path', path);
  form.append('file', { uri, name: fileName, type: mimeType } as unknown as Blob);
  const result = await xhrUpload('POST', buildUrl('/api/upload'), {
    body: form,
    onUploadProgress: onProgress,
  });
  return result as ApiResponse<{ path: string; size: number }[]>;
}

export const api = {
  login: async (username: string, password: string): Promise<ApiResponse<TokenResponse>> => {
    let response: Response;
    try {
      response = await fetch(buildUrl('/api/auth/token'), {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password }),
      });
    } catch (e) {
      return { error: e instanceof Error ? e.message : 'Network request failed' };
    }
    const result = await handleResponse<TokenResponse>(response);
    if (result.data?.accessToken) {
      await setToken(result.data.accessToken);
    }
    return result;
  },

  revoke: async (): Promise<ApiResponse> => {
    const res = await request('/api/auth/revoke', { method: 'POST' });
    await clearToken();
    return res;
  },

  logoutLocal: async (): Promise<void> => {
    await clearToken();
  },

  authInfo: (): Promise<ApiResponse<AuthInfo>> => request('/api/auth/info'),

  me: (): Promise<ApiResponse<AuthUser>> => request('/api/auth/me'),

  listFiles: async (path = ''): Promise<ApiResponse<FileItem[]>> => {
    let response: Response;
    try {
      response = await fetch(buildUrl('/api/files', path ? { path } : undefined), {
        headers: await authHeaders(),
      });
    } catch (e) {
      return { error: e instanceof Error ? e.message : 'Network request failed' };
    }
    return handleResponse<FileItem[]>(response);
  },

  createDirectory: (path: string): Promise<ApiResponse> =>
    request('/api/directories', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ path }),
    }),

  deletePath: (path: string): Promise<ApiResponse> =>
    request('/api/files', {
      method: 'DELETE',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ path }),
    }),

  createShare: (path: string, expiresInSeconds: number): Promise<ApiResponse<ShareItem>> =>
    request<ShareItem>('/api/shares', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ path, expiresInSeconds }),
    }),

  listShares: (): Promise<ApiResponse<ShareItem[]>> => request('/api/shares'),

  revokeShare: (token: string): Promise<ApiResponse> =>
    request('/api/shares', {
      method: 'DELETE',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ token }),
    }),

  shareUrl: (token: string): string => buildUrl(`/api/share/${token}`),

  uploadFile: async (
    uri: string,
    fileName: string,
    mimeType: string,
    path: string,
    onProgress?: (loaded: number, total: number) => void,
  ): Promise<ApiResponse<{ path: string; size: number }[]>> => {
    // Prefer the resumable session API when we can read the file as a Blob
    // (chunked PATCH + complete). Fall back to single-shot multipart otherwise.
    try {
      const response = await fetch(uri);
      if (response.ok) {
        const blob = await response.blob();
        if (blob.size > 4 * 1024 * 1024) {
          return await uploadResumableBlob(blob, fileName, path, onProgress);
        }
        return await uploadMultipartBlob(blob, fileName, mimeType, path, onProgress);
      }
    } catch {
      /* fall through to FormData path */
    }
    return uploadMultipartUri(uri, fileName, mimeType, path, onProgress);
  },

  downloadUrl: (path: string): string => buildUrl('/api/files/download', { path }),

  fetchDownload: async (path: string): Promise<ApiResponse<{ blob: Blob; name: string }>> => {
    let response: Response;
    try {
      response = await fetch(buildUrl('/api/files/download', { path }), {
        headers: await authHeaders(),
      });
    } catch (e) {
      return { error: e instanceof Error ? e.message : 'Network request failed' };
    }
    if (!response.ok) {
      if (response.status === 401) {
        await clearToken();
        unauthorizedListener?.();
        return { error: 'Not authorized', unauthorized: true };
      }
      return { error: `Download failed (${response.status})` };
    }
    const blob = await response.blob();
    const disposition = response.headers.get('content-disposition') || '';
    const match = /filename\*?=(?:UTF-8'')?"?([^";]+)"?/i.exec(disposition);
    const name = match ? decodeURIComponent(match[1]) : path.split('/').pop() || 'download';
    return { data: { blob, name } };
  },
};
