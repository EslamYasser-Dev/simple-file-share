import { useState, useCallback } from 'react';
import { api } from '../services/api';
import type { ApiResponse } from '../services/api';
import { MAX_FILE_SIZE, ALLOWED_FILE_TYPES } from '../config';

export interface FileItem {
  name: string;
  path: string;
  size: number;
  isDir: boolean;
  modified: string;
  mimeType?: string;
}

export const useFileOperations = () => {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [uploadProgress, setUploadProgress] = useState(0);

  const validateFile = useCallback((file: File): { valid: boolean; error?: string } => {
    if (file.size > MAX_FILE_SIZE) {
      return {
        valid: false,
        error: `File size exceeds the limit of ${MAX_FILE_SIZE / (1024 * 1024)}MB`
      };
    }

    if (!ALLOWED_FILE_TYPES.includes(file.type)) {
      return {
        valid: false,
        error: 'File type not allowed'
      };
    }

    return { valid: true };
  }, []);

  const uploadFile = useCallback(async (file: File, path = ''): Promise<ApiResponse> => {
    const { valid, error: validationError } = validateFile(file);
    if (!valid && validationError) {
      setError(validationError);
      return { error: validationError };
    }

    setIsLoading(true);
    setUploadProgress(0);

    try {
      const result = await api.uploadFile(file, path);
      if (result.error) throw new Error(result.error);
      return result;
    } catch (err) {
      const error = err instanceof Error ? err.message : 'Upload failed';
      setError(error);
      return { error };
    } finally {
      setIsLoading(false);
      setUploadProgress(0);
    }
  }, [validateFile]);

  const fetchFiles = useCallback(async (path = ''): Promise<FileItem[] | null> => {
    setIsLoading(true);
    try {
      const { data, error } = await api.listFiles(path);
      if (error) throw new Error(error);
      return data || [];
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch files');
      return null;
    } finally {
      setIsLoading(false);
    }
  }, []);

  const createDirectory = useCallback(async (path: string): Promise<ApiResponse> => {
    setIsLoading(true);
    try {
      return await api.createDirectory(path);
    } catch (err) {
      const error = err instanceof Error ? err.message : 'Failed to create directory';
      setError(error);
      return { error };
    } finally {
      setIsLoading(false);
    }
  }, []);

  const deleteItem = useCallback(async (path: string): Promise<ApiResponse> => {
    setIsLoading(true);
    try {
      return await api.deletePath(path);
    } catch (err) {
      const error = err instanceof Error ? err.message : 'Failed to delete item';
      setError(error);
      return { error };
    } finally {
      setIsLoading(false);
    }
  }, []);

  const clearError = useCallback(() => {
    setError(null);
  }, []);

  return {
    isLoading,
    error,
    uploadProgress,
    uploadFile,
    fetchFiles,
    createDirectory,
    deleteItem,
    clearError,
  };
};

export default useFileOperations;
