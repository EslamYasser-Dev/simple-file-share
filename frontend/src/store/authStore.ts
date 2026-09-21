import { create } from 'zustand';
import type { AuthUser } from '../services/api';

interface AuthState {
  /** Authenticated account, or null before login / when auth is disabled. */
  user: AuthUser | null;
  /** Whether the server allows self-service registration. */
  signupEnabled: boolean;
  setUser: (user: AuthUser) => void;
  setSignupEnabled: (enabled: boolean) => void;
  clear: () => void;
}

export const useAuthStore = create<AuthState>()((set) => ({
  user: null,
  signupEnabled: false,
  setUser: (user) => set({ user }),
  setSignupEnabled: (signupEnabled) => set({ signupEnabled }),
  clear: () => set({ user: null }),
}));
