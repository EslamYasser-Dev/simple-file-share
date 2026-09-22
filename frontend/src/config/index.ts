// API Configuration.
// When unset (no VITE_API_URL), the app talks to the same origin that served it
// (e.g. the Go server's /api routes), so the base is '' and api.ts appends '/api'.
export const API_BASE_URL = import.meta.env.VITE_API_URL || '';

// Upload size limit in bytes. 0 = unlimited (bounded only by disk space).
// Set VITE_MAX_UPLOAD_MB to enforce a limit at build time.
export const MAX_FILE_SIZE = (Number(import.meta.env.VITE_MAX_UPLOAD_MB) || 0) * 1024 * 1024;

// Bounded parallel transfers keep the UI responsive without overwhelming the
// server or the browser connection pool. Uploads are batched client-side;
// downloads are one file at a time, so only the upload limit is tunable.
export const MAX_CONCURRENT_UPLOADS = parseConcurrency(import.meta.env.VITE_MAX_CONCURRENT_UPLOADS, 4);

function parseConcurrency(value: unknown, fallback: number): number {
  const parsed = typeof value === 'string' ? Number.parseInt(value, 10) : Number.NaN;
  if (!Number.isFinite(parsed)) return fallback;
  return Math.min(Math.max(Math.trunc(parsed), 1), 8);
}
export const ALLOWED_FILE_TYPES = [
  // Documents
  'application/pdf',
  'application/msword',
  'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
  'application/vnd.ms-excel',
  'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
  'text/plain',
  'text/csv',
  'text/markdown',
  'application/json',

  // Images
  'image/jpeg',
  'image/png',
  'image/gif',
  'image/webp',
  'image/svg+xml',
  'image/bmp',

  // Archives & disk images
  'application/zip',
  'application/x-rar-compressed',
  'application/x-7z-compressed',
  'application/x-tar',
  'application/gzip',
  'application/x-gzip',
  'application/x-iso9660-image',
  'application/x-iso',
  'application/iso-image',

  // Video
  'video/mp4',
  'video/webm',
  'video/ogg',
  'video/quicktime',
  'video/x-msvideo',
  'video/x-matroska',

  // Audio
  'audio/mpeg',
  'audio/ogg',
  'audio/wav',
  'audio/mp4',
  'audio/webm',
  'audio/x-m4a',
];

// Extensions accepted when the browser reports an empty or non-standard MIME
// type — common for disk images like `.iso`, which often arrive as "".
export const ALLOWED_FILE_EXTENSIONS = [
  // Documents
  'pdf', 'doc', 'docx', 'xls', 'xlsx', 'txt', 'md', 'csv', 'json',
  // Images
  'jpg', 'jpeg', 'png', 'gif', 'webp', 'svg', 'bmp',
  // Archives & disk images
  'zip', 'rar', '7z', 'tar', 'gz', 'iso', 'img',
  // Video
  'mp4', 'webm', 'ogv', 'mov', 'avi', 'mkv', 'm4v',
  // Audio
  'mp3', 'ogg', 'wav', 'm4a', 'weba',
];
