import { useState } from 'react';
import { HardDrive, Loader2, Lock } from 'lucide-react';
import { AppBackground } from '../components/Layout';
import { useI18n } from '../i18n';
import { buildUrl } from '../services/api';

interface LoginProps {
  onLogin: (username: string, password: string) => Promise<string | null>;
  signupEnabled?: boolean;
  oauthProviders?: string[];
  onShowRegister?: () => void;
}

const OAUTH_ICONS: Record<string, string> = {
  github:
    'M12 2C6.477 2 2 6.477 2 12c0 4.42 2.87 8.17 6.84 9.5.5.08.66-.22.66-.48v-1.7c-2.78.6-3.37-1.34-3.37-1.34-.45-1.15-1.11-1.46-1.11-1.46-.91-.62.07-.6.07-.6 1 .07 1.53 1.03 1.53 1.03.89 1.53 2.34 1.09 2.91.83.09-.65.35-1.09.63-1.34-2.22-.25-4.56-1.11-4.56-4.95 0-1.09.39-1.98 1.03-2.68-.1-.25-.45-1.27.1-2.64 0 0 .84-.27 2.75 1.02.8-.22 1.65-.33 2.5-.33s1.7.11 2.5.33c1.91-1.29 2.75-1.02 2.75-1.02.55 1.37.2 2.39.1 2.64.64.7 1.03 1.59 1.03 2.68 0 3.85-2.34 4.7-4.57 4.94.36.31.68.92.68 1.86v2.75c0 .27.16.57.67.48A10 10 0 0 0 22 12c0-5.52-4.48-10-10-10z',
  google:
    'M21.35 11.1H12v2.9h5.35c-.5 2.5-2.6 4.3-5.35 4.3a5.9 5.9 0 1 1 0-11.8c1.5 0 2.85.55 3.9 1.45l2.15-2.15A8.9 8.9 0 1 0 12 20.9c5.15 0 8.7-3.65 8.7-8.8 0-.35-.05-.7-.1-1z',
};

export function Login({ onLogin, signupEnabled, oauthProviders = [], onShowRegister }: LoginProps) {
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

  const oauthLabel = (name: string) =>
    name === 'github' ? t('auth.continueWithGithub') : t('auth.continueWithGoogle');

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

        {oauthProviders.length > 0 && (
          <div className="glass-panel mt-4 space-y-3 p-6">
            <p className="text-center text-[11px] font-semibold uppercase tracking-widest text-slate-500">
              {t('auth.or')}
            </p>
            {oauthProviders.map((provider) => (
              <a
                key={provider}
                href={buildUrl(`/api/auth/oauth/${provider}/start`)}
                className="inline-flex w-full items-center justify-center gap-2 rounded-xl border border-white/10 bg-white/5 px-4 py-2.5 text-sm font-semibold text-slate-200 transition-colors hover:bg-white/10 hover:text-white"
              >
                <svg viewBox="0 0 24 24" aria-hidden="true" className="h-4 w-4 fill-current">
                  <path d={OAUTH_ICONS[provider] ?? OAUTH_ICONS.github} />
                </svg>
                {oauthLabel(provider)}
              </a>
            ))}
          </div>
        )}

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