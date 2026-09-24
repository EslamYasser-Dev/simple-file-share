import { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react';
import type { ReactNode } from 'react';
import { api, onUnauthorized, type AuthUser } from './api';
import { getToken } from './token';
import { startEventStream, stopEventStream } from './events';

interface AuthState {
  status: 'loading' | 'signedOut' | 'signedIn';
  user: AuthUser | null;
  signIn: (username: string, password: string) => Promise<string | null>;
  signOut: () => Promise<void>;
  refreshUser: () => Promise<void>;
}

const AuthContext = createContext<AuthState | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [status, setStatus] = useState<AuthState['status']>('loading');
  const [user, setUser] = useState<AuthUser | null>(null);

  const refreshUser = useCallback(async () => {
    const res = await api.me();
    if (res.data) {
      setUser(res.data);
      setStatus('signedIn');
    } else if (res.unauthorized) {
      setUser(null);
      setStatus('signedOut');
      stopEventStream();
    }
  }, []);

  useEffect(() => {
    let cancelled = false;
    void (async () => {
      const token = await getToken();
      if (cancelled) return;
      if (!token) {
        setStatus('signedOut');
        return;
      }
      const res = await api.me();
      if (cancelled) return;
      if (res.data) {
        setUser(res.data);
        setStatus('signedIn');
        startEventStream();
      } else {
        setUser(null);
        setStatus('signedOut');
      }
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  useEffect(() => {
    onUnauthorized(() => {
      setUser(null);
      setStatus('signedOut');
      stopEventStream();
    });
    return () => onUnauthorized(null);
  }, []);

  const signIn = useCallback(async (username: string, password: string) => {
    const res = await api.login(username, password);
    if (res.error || !res.data) return res.error || 'Sign-in failed';
    const me = await api.me();
    if (me.error || !me.data) return me.error || 'Unable to load account';
    setUser(me.data);
    setStatus('signedIn');
    startEventStream();
    return null;
  }, []);

  const signOut = useCallback(async () => {
    stopEventStream();
    await api.revoke();
    setUser(null);
    setStatus('signedOut');
  }, []);

  const value = useMemo(
    () => ({ status, user, signIn, signOut, refreshUser }),
    [status, user, signIn, signOut, refreshUser],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthState {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
}
