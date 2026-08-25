import { create } from 'zustand';
import type { ToastType } from '../components/toastContext';

interface ToastItem {
  id: number;
  type: ToastType;
  message: string;
}

export interface ToastState {
  toasts: ToastItem[];
  removeToast: (id: number) => void;
  showToast: (type: ToastType, message: string) => void;
  success: (message: string) => void;
  error: (message: string) => void;
  info: (message: string) => void;
}

let toastId = 0;
const AUTO_DISMISS_MS = 4000;

export const useToastStore = create<ToastState>()((set, get) => ({
  toasts: [],

  removeToast: (id) => {
    set((state) => ({ toasts: state.toasts.filter((t) => t.id !== id) }));
  },

  showToast: (type, message) => {
    const id = ++toastId;
    set((state) => ({ toasts: [...state.toasts, { id, type, message }] }));
    setTimeout(() => get().removeToast(id), AUTO_DISMISS_MS);
  },

  success: (message) => get().showToast('success', message),
  error: (message) => get().showToast('error', message),
  info: (message) => get().showToast('info', message),
}));