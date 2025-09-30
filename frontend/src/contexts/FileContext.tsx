import { createContext, useContext, useState, useCallback } from 'react';
import type { ReactNode } from 'react';
import { api, type ApiResponse } from '../services/api';
import { MAX_FILE_SIZE, ALLOWED_FILE_TYPES } from '../config';

export interface FileItem {
  name: string;
  path: string;
  size: number;
  isDir: boolean;
  modified: string;
  mimeType?: string;
}

interface FileContextType {
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

const FileContext = createContext<FileContextType | undefined>(undefined);

export const FileProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [files, setFiles] = useState<FileItem[]>([]);
  const [currentPath, setCurrentPath] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [uploadProgress, setUploadProgress] = useState(0);

  const fetchFiles = useCallback(async (path = '') => {
    setIsLoading(true);
    try {
      const { data, error } = await api.listFiles(path);
      if (error) throw new Error(error);
      setFiles(data || []);
      setCurrentPath(path);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch files');
    } finally {
      setIsLoading(false);
    }
  }, []);

  const uploadFile = useCallback(async (file: File, path = '') => {
    if (file.size > MAX_FILE_SIZE) {
      const error = `File size exceeds the limit of ${MAX_FILE_SIZE / (1024 * 1024)}MB`;
      setError(error);
      return { error };
    }

    if (!ALLOWED_FILE_TYPES.includes(file.type)) {
      const error = 'File type not allowed';
      setError(error);
      return { error };
    }

    setIsLoading(true);
    setUploadProgress(0);

    try {
      const result = await api.uploadFile(file, path);
      if (result.error) throw new Error(result.error);
      
      // Refresh the current directory
      await fetchFiles(currentPath);
      return result;
    } catch (err) {
      const error = err instanceof Error ? err.message : 'Upload failed';
      setError(error);
      return { error };
    } finally {
      setIsLoading(false);
      setUploadProgress(0);
    }
  }, [currentPath, fetchFiles]);

  const createDirectory = useCallback(async (name: string) => {
    const path = currentPath ? `${currentPath}/${name}` : name;
    const result = await api.createDirectory(path);
    if (!result.error) {
      await fetchFiles(currentPath);
    }
    return result;
  }, [currentPath, fetchFiles]);

  const deleteItem = useCallback(async (path: string) => {
    const result = await api.deletePath(path);
    if (!result.error) {
      await fetchFiles(currentPath);
    }
    return result;
  }, [currentPath, fetchFiles]);

  const navigateToPath = useCallback((path: string) => {
    fetchFiles(path);
  }, [fetchFiles]);

  const goUp = useCallback(() => {
    if (!currentPath) return;
    const parentPath = currentPath.split('/').slice(0, -1).join('/');
    fetchFiles(parentPath);
  }, [currentPath, fetchFiles]);

  const clearError = useCallback(() => {
    setError(null);
  }, []);

  return (
    <FileContext.Provider
      value={{
        files,
        currentPath,
        isLoading,
        error,
        uploadProgress,
        uploadFile,
        fetchFiles,
        createDirectory,
        deleteItem,
        navigateToPath,
        goUp,
        clearError,
      }}
    >
      {children}
    </FileContext.Provider>
  );
};

export const useFiles = (): FileContextType => {
  const context = useContext(FileContext);
  if (context === undefined) {
    throw new Error('useFiles must be used within a FileProvider');
  }
  return context;
};
