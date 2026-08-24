import { createContext } from 'react';
import type { ApiResponse } from '../services/api';

export interface FileItem {
  name: string;
  path: string;
  size: number;
  isDir: boolean;
  modified: string;
  mimeType?: string;
}

export interface FileContextType {
  files: FileItem[];
  currentPath: string;
  isLoading: boolean;
  error: string | null;
  uploadProgress: number;
  uploadFile: (file: File, path?: string) => Promise<ApiResponse>;
  fetchFiles: (path?: string) => Promise<void>;
  createDirectory: (path: string) => Promise<ApiResponse>;
  deleteItem: (path: string) => Promise<ApiResponse>;
  navigateToPath: (path: string) => void;
  goUp: () => void;
  clearError: () => void;
}

export const FileContext = createContext<FileContextType | undefined>(undefined);