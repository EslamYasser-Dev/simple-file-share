import { useShallow } from 'zustand/react/shallow';
import { useToastStore } from '../store/toastStore';

export interface ToastApi {
  success: (message: string) => void;
  error: (message: string) => void;
  info: (message: string) => void;
  showToast: (type: 'success' | 'error' | 'info', message: string) => void;
}

/** Returns stable references to the global toast actions (backed by Zustand). */
export function useToast(): ToastApi {
  return useToastStore(
    useShallow((s) => ({
      success: s.success,
      error: s.error,
      info: s.info,
      showToast: s.showToast,
    })),
  );
}