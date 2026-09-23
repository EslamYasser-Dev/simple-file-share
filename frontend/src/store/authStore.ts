import { create } from 'zustand';
import type { AuthUser } from '../services/api';

interface AuthState {
  /** Authenticated account, or null before login / when auth is disabled. */
  user: AuthUser | null;
  /** Whether the server allows self-service registration. */
  signupEnabled: boolean;
  /** Configured OAuth providers (e.g. github, google). */
  oauthProviders: string[];
  setUser: (user: AuthUser) => void;
  setSignupEnabled: (enabled: boolean) => void;
  setOAuthProviders: (providers: string[]) => void;
  clear: () => void;
}

export const useAuthStore = create<AuthState>()((set) => ({
  user: null,
  signupEnabled: false,
  oauthProviders: [],
  setUser: (user) => set({ user }),
  setSignupEnabled: (signupEnabled) => set({ signupEnabled }),
  setOAuthProviders: (oauthProviders) => set({ oauthProviders }),
  clear: () => set({ user: null }),
}));
