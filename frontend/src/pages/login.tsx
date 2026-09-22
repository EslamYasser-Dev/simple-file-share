import { useState } from 'react';
import { HardDrive, Loader2, Lock } from 'lucide-react';
import { AppBackground } from '../components/Layout';
import { useI18n } from '../i18n';

interface LoginProps {
  onLogin: (username: string, password: string) => Promise<string | null>;
  signupEnabled?: boolean;
  onShowRegister?: () => void;
}

export function Login({ onLogin, signupEnabled, onShowRegister }: LoginProps) {
  const { t } = useI18n();
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!username.trim() || !password || isSubmitting) return;
    setIsSubmitting(true);
    setError(null);
    const err = await onLogin(username.trim(), password);
    if (err) setError(err);
    setIsSubmitting(false);
  };

  return (
    <div className="relative flex min-h-screen items-center justify-center px-4 text-slate-100">
      <AppBackground />

      <div className="w-full max-w-sm animate-rise">
        <div className="mb-8 flex flex-col items-center gap-3 text-center">
          <div className="relative flex h-14 w-14 items-center justify-center rounded-2xl bg-gradient-to-br from-cyan-400 via-blue-500 to-violet-600 shadow-lg shadow-cyan-500/40">
            <HardDrive className="h-7 w-7 text-white" />
            <span className="absolute inset-0 rounded-2xl bg-gradient-to-br from-cyan-400/60 to-violet-600/60 blur-lg animate-glow" />
          </div>
          <div>
            <h1 className="text-2xl font-bold tracking-tight">
              File<span className="text-gradient">Share</span>
            </h1>
            <p className="mt-1 text-sm text-slate-500">{t('auth.signInToAccess')}</p>
          </div>
        </div>

        <form
          onSubmit={handleSubmit}
          className="glass-panel space-y-4 p-6"
        >
          <div className="space-y-1.5">
            <label htmlFor="username" className="text-[11px] font-semibold uppercase tracking-widest text-slate-500">
              {t('auth.username')}
            </label>
            <input
              id="username"
              autoFocus
              autoComplete="username"
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              placeholder={t('auth.username')}
              className="w-full rounded-xl border border-white/10 bg-white/5 px-3 py-2.5 text-sm text-slate-100 outline-none backdrop-blur transition-all focus:border-cyan-400/50 focus:ring-1 focus:ring-cyan-400/30"
            />
          </div>

          <div className="space-y-1.5">
            <label htmlFor="password" className="text-[11px] font-semibold uppercase tracking-widest text-slate-500">
              {t('auth.password')}
            </label>
            <input
              id="password"
              autoComplete="current-password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="••••••••"
              className="w-full rounded-xl border border-white/10 bg-white/5 px-3 py-2.5 text-sm text-slate-100 outline-none backdrop-blur transition-all focus:border-cyan-400/50 focus:ring-1 focus:ring-cyan-400/30"
            />
          </div>

          {error && (
            <p className="flex items-center gap-2 rounded-xl border border-red-500/30 bg-red-950/40 px-3 py-2 text-xs text-red-300">
              <Lock className="h-3.5 w-3.5 shrink-0" />
              {error}
            </p>
          )}

          <button
            type="submit"
            disabled={!username.trim() || !password || isSubmitting}
            className="inline-flex w-full items-center justify-center gap-2 rounded-xl bg-gradient-to-r from-cyan-500 to-violet-600 px-4 py-2.5 text-sm font-semibold text-white shadow-lg shadow-cyan-500/30 transition-all duration-300 hover:shadow-cyan-400/50 hover:brightness-110 disabled:cursor-not-allowed disabled:opacity-50 disabled:shadow-none"
          >
            {isSubmitting && <Loader2 className="h-4 w-4 animate-spin" />}
            {t('auth.signIn')}
          </button>
        </form>

        {signupEnabled && onShowRegister && (
          <p className="mt-5 text-center text-xs text-slate-500">
            {t('auth.noAccount')}{' '}
            <button
              type="button"
              onClick={onShowRegister}
              className="font-semibold text-cyan-300 transition-colors hover:text-cyan-200"
            >
              {t('auth.createAccount')}
            </button>
          </p>
        )}
      </div>
    </div>
  );
}