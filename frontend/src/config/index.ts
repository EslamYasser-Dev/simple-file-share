// API Configuration
export const API_BASE_URL = import.meta.env.VITE_API_URL || '/api';

// File upload configuration
export const MAX_FILE_SIZE = 50 * 1024 * 1024; // 50MB
export const ALLOWED_FILE_TYPES = [
  // Documents
  'application/pdf',
  'application/msword',
  'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
  'application/vnd.ms-excel',
  'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
  'text/plain',
  'text/csv',
  
  // Images
  'image/jpeg',
  'image/png',
  'image/gif',
  'image/webp',
  'image/svg+xml',
  
  // Archives
  'application/zip',
  'application/x-rar-compressed',
  'application/x-7z-compressed',
];

// UI Configuration
export const UI_CONFIG = {
  maxFilesToShow: 100,
  defaultPageSize: 20,
  debounceTime: 300, // ms
  toastDuration: 5000, // ms
};
