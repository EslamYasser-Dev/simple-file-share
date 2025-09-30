import { API_BASE_URL } from '@/config';

export interface FileItem {
  name: string;
  path: string;
  size: number;
  isDir: boolean;
  modified: string;
  mimeType?: string;
}

export interface ApiResponse<T = any> {
  data?: T;
  error?: string;
  message?: string;
}

async function handleResponse<T>(response: Response): Promise<ApiResponse<T>> {
  if (!response.ok) {
    const error = await response.text();
    return { error: error || 'An error occurred' };
  }
  
  try {
    const data = await response.json();
    return { data };
  } catch (error) {
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

    const response = await fetch(url.toString());
    return handleResponse(response);
  },

  // Create directory
  createDirectory: async (path: string): Promise<ApiResponse> => {
    const response = await fetch(`${API_BASE_URL}/api/directories`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ path }),
    });

    return handleResponse(response);
  },

  // Delete file or directory
  deletePath: async (path: string): Promise<ApiResponse> => {
    const response = await fetch(`${API_BASE_URL}/api/files`, {
      method: 'DELETE',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ path }),
    });

    return handleResponse(response);
  },

  // Download file
  downloadFile: async (path: string): Promise<Blob> => {
    const response = await fetch(`${API_BASE_URL}/api/files/download?path=${encodeURIComponent(path)}`);
    if (!response.ok) {
      throw new Error('Failed to download file');
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
    const response = await fetch(`${API_BASE_URL}/api/files/info?path=${encodeURIComponent(path)}`);
    return handleResponse(response);
  },
};

export default api;
