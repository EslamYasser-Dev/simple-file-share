// This file ensures TypeScript understands our path aliases
declare module '@/config' {
  export const API_BASE_URL: string;
  export const MAX_FILE_SIZE: number;
  export const ALLOWED_FILE_TYPES: string[];
  export const UI_CONFIG: {
    maxFilesToShow: number;
    defaultPageSize: number;
    debounceTime: number;
    toastDuration: number;
  };
}

declare module '*.module.css' {
  const classes: { [key: string]: string };
  export default classes;
}

declare module '*.module.scss' {
  const classes: { [key: string]: string };
  export default classes;
}
