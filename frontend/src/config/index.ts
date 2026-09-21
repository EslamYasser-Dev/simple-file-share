// API Configuration.
// When unset (no VITE_API_URL), the app talks to the same origin that served it
// (e.g. the Go server's /api routes), so the base is '' and api.ts appends '/api'.
export const API_BASE_URL = import.meta.env.VITE_API_URL || '';

// Upload size limit in bytes. 0 = unlimited (bounded only by disk space).
// Set VITE_MAX_UPLOAD_MB to enforce a limit at build time.
export const MAX_FILE_SIZE = (Number(import.meta.env.VITE_MAX_UPLOAD_MB) || 0) * 1024 * 1024;
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
