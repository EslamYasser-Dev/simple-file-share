import { useEffect, useState } from 'react';
import type { ReactNode } from 'react';
import type { LucideIcon } from 'lucide-react';
import {
  FolderOpen,
  LayoutDashboard,
  MessageSquare,
  HardDrive,
  FolderPlus,
  Upload,
  Languages,
  Share2,
  Users,
  LogOut,
  Settings,
  Radio,
} from 'lucide-react';
import { useI18n } from '../i18n';
import { useAuthStore } from '../store/authStore';
import { useConfigStore } from '../store/configStore';
import { ConfigPanel } from './ConfigPanel';

export type Page = 'files' | 'shared' | 'summary' | 'chat' | 'admin' | 'p2p';

interface LayoutProps {
  active: Page;
  onNavigate: (page: Page) => void;
  onUpload: () => void;
  onNewFolder: () => void;
  onSignOut: () => void;
  children: ReactNode;
}

/** Ambient animated background shared by every screen. */
export function AppBackground() {
  return (
    <div aria-hidden className="pointer-events-none fixed inset-0 -z-10 overflow-hidden">
      <div className="absolute inset-x-0 top-0 h-40 bg-gradient-to-b from-cyan-400/[0.06] to-transparent" />
      <div className="grid-overlay absolute -top-40 left-1/2 h-[70rem] w-[70rem] -translate-x-1/2" />
      <div className="absolute -left-40 -top-48 h-[40rem] w-[40rem] rounded-full bg-cyan-500/15 blur-[160px] animate-aurora" />
      <div className="absolute -right-48 top-1/4 h-[38rem] w-[38rem] rounded-full bg-violet-600/15 blur-[170px] animate-aurora [animation-delay:-6s]" />
      <div className="absolute bottom-[-16rem] left-1/3 h-[36rem] w-[36rem] rounded-full bg-fuchsia-500/10 blur-[180px] animate-aurora [animation-delay:-12s]" />
      <div className="absolute inset-x-0 top-0 h-px bg-gradient-to-r from-transparent via-cyan-400/40 to-transparent" />
    </div>
  );
}

export function Layout({ active, onNavigate, onUpload, onNewFolder, onSignOut, children }: LayoutProps) {
  const { t, dir, lang, setLang } = useI18n();
  const isRtl = dir === 'rtl';
  const user = useAuthStore((s) => s.user);
  const isAdmin = user?.isAdmin ?? false;
  const theme = useConfigStore((s) => s.theme);
  const setTheme = useConfigStore((s) => s.setTheme);
  const lanShareEnabled = useConfigStore((s) => s.lanShareEnabled);
  const [showConfig, setShowConfig] = useState(false);

  // Resolve the effective theme (system follows the OS preference).
  const resolvedTheme: 'light' | 'dark' =
    theme === 'system'
      ? window.matchMedia('(prefers-color-scheme: light)').matches
        ? 'light'
        : 'dark'
      : theme;

  // Apply the active theme to the <html> element so CSS variables switch, and
  // keep the browser chrome (address bar) in sync with the resolved theme.
  useEffect(() => {
    document.documentElement.classList.toggle('light', resolvedTheme === 'light');
    const meta = document.querySelector('meta[name="theme-color"]');
    meta?.setAttribute('content', resolvedTheme === 'light' ? '#f8fafc' : '#04050c');
  }, [resolvedTheme]);

  const NAV_ITEMS: { key: Page; label: string; icon: LucideIcon }[] = [
    { key: 'files', label: t('nav.myFiles'), icon: FolderOpen },
    { key: 'shared', label: t('nav.shared'), icon: Share2 },
    { key: 'summary', label: t('nav.summary'), icon: LayoutDashboard },
    { key: 'chat', label: t('nav.activity'), icon: MessageSquare },
    ...(lanShareEnabled ? [{ key: 'p2p' as Page, label: t('nav.p2p'), icon: Radio }] : []),
    ...(isAdmin ? [{ key: 'admin' as Page, label: t('nav.admin'), icon: Users }] : []),
  ];

  // The global shared folder is read-only for everyone except admins.
  const readOnly = active === 'shared' && !isAdmin;

  return (
    <div className="relative flex min-h-screen text-slate-100">
      <AppBackground />

      {/* Sidebar */}
      <aside
        className={`glass fixed inset-y-0 z-30 flex w-16 flex-col border-r border-white/10 md:w-64 ${
          isRtl ? 'right-0 border-r-0 border-l' : 'left-0'
        }`}
      >
        {/* Logo */}
        <div className="flex items-center gap-3 border-b border-white/10 px-4 py-5">
          <div className="relative flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-gradient-to-br from-cyan-400 via-blue-500 to-violet-600 shadow-lg shadow-cyan-500/40">
            <HardDrive className="h-5 w-5 text-white" />
            <span className="absolute inset-0 rounded-xl bg-gradient-to-br from-cyan-400/60 to-violet-600/60 blur-md animate-glow" />
          </div>
          <div className="hidden md:block">
            <h1 className="text-sm font-bold tracking-tight">
              File<span className="text-gradient">Share</span>
            </h1>
            <p className="text-[10px] uppercase tracking-[0.25em] text-cyan-300/60" dir="ltr">
              {t('nav.cloudStorage')}
            </p>
          </div>
        </div>

        {/* Nav */}
        <nav className="flex-1 space-y-1 px-2 py-4">
          {NAV_ITEMS.map(({ key, label, icon: Icon }) => {
            const isActive = active === key;
            return (
              <button
                key={key}
                onClick={() => onNavigate(key)}
                aria-current={isActive ? 'page' : undefined}
                className={`group relative flex w-full items-center gap-3 overflow-hidden rounded-xl px-3 py-2.5 text-sm font-medium transition-all duration-300 ${
                  isActive ? 'app-nav-active' : 'text-slate-400 hover:bg-white/5 hover:text-slate-100'
                }`}
              >
                {isActive && (
                  <>
                    <span className="absolute inset-0 bg-gradient-to-r from-cyan-500/20 to-violet-500/10" />
                    <span
                      className={`absolute top-1/2 h-6 w-1 -translate-y-1/2 rounded-full bg-gradient-to-b from-cyan-400 to-violet-500 shadow-[0_0_12px_2px_rgba(34,211,238,0.7)] ${
                        isRtl ? 'right-0' : 'left-0'
                      }`}
                    />
                  </>
                )}
                <Icon
                  className={`relative h-5 w-5 shrink-0 transition-colors ${
                    isActive ? 'text-cyan-300' : 'group-hover:text-cyan-300'
                  }`}
                />
                <span className="relative hidden md:inline">{label}</span>
              </button>
            );
          })}
        </nav>

        {/* Actions */}
        <div className="space-y-2 border-t border-white/10 p-3">
          {!readOnly && (
            <>
              <button
                onClick={onUpload}
                className="flex w-full items-center justify-center gap-2 rounded-xl bg-gradient-to-r from-cyan-500 to-violet-600 px-3 py-2.5 text-sm font-semibold text-white shadow-lg shadow-cyan-500/30 transition-all duration-300 hover:shadow-cyan-400/50 hover:brightness-110"
              >
                <Upload className="h-4 w-4" />
                <span className="hidden md:inline">{t('nav.upload')}</span>
              </button>
              <button
                onClick={onNewFolder}
                className="flex w-full items-center justify-center gap-2 rounded-xl border border-white/10 bg-white/5 px-3 py-2.5 text-sm font-medium text-slate-200 backdrop-blur transition-all duration-300 hover:border-cyan-400/40 hover:bg-white/10 hover:text-white"
              >
                <FolderPlus className="h-4 w-4" />
                <span className="hidden md:inline">{t('nav.newFolder')}</span>
              </button>
            </>
          )}

          {/* User badge + sign out */}
          {user && (
            <div className="flex items-center gap-2 rounded-xl border border-white/10 bg-white/5 px-2 py-2">
              <div className="flex h-7 w-7 shrink-0 items-center justify-center rounded-lg bg-gradient-to-br from-cyan-400 to-violet-600 text-xs font-bold uppercase text-white">
                {(user.username || '?').charAt(0)}
              </div>
              <span className="hidden flex-1 truncate text-xs font-medium text-slate-300 md:inline">
                {user.username || t('admin.roleAdmin')}
              </span>
              <button
                onClick={onSignOut}
                className="rounded-lg p-1.5 text-slate-400 transition-colors hover:bg-red-500/10 hover:text-red-400"
                title={t('nav.signOut')}
                aria-label={t('nav.signOut')}
              >
                <LogOut className="h-4 w-4" />
              </button>
            </div>
          )}

          {/* Language toggle */}
          <div className="flex items-center gap-2 px-1 pt-2">
            <Languages className="h-4 w-4 shrink-0 text-slate-500" />
            <div className="hidden flex-1 md:flex">
              <button
                onClick={() => setLang(lang === 'ar' ? 'en' : 'ar')}
                className="w-full rounded-lg border border-white/10 bg-white/5 px-2 py-1.5 text-xs font-semibold text-slate-300 transition-colors hover:border-cyan-400/40 hover:text-white"
                aria-label={t('nav.language')}
              >
                {lang === 'ar' ? 'English' : 'العربية'}
              </button>
            </div>
          </div>

          {/* Configuration panel */}
          <button
            onClick={() => setShowConfig(true)}
            className="flex w-full items-center gap-3 rounded-xl border border-white/10 bg-white/5 px-3 py-2.5 text-sm font-medium text-slate-300 transition-all duration-300 hover:border-cyan-400/40 hover:bg-white/10 hover:text-white"
            title={t('nav.config')}
            aria-label={t('nav.config')}
          >
            <Settings className="h-4 w-4 shrink-0" />
            <span className="hidden md:inline">{t('nav.config')}</span>
          </button>

{/* Theme toggle */}
          <div className="flex items-center gap-2 px-1 pt-2">
            <div
              role="switch"
              aria-checked={resolvedTheme === 'light'}
              tabIndex={0}
              onClick={() => setTheme(resolvedTheme === 'light' ? 'dark' : 'light')}
              onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') setTheme(resolvedTheme === 'light' ? 'dark' : 'light'); }}
              className="relative h-4 w-8 cursor-pointer rounded-full border border-white/20 bg-white/10 transition-colors"
              aria-label={t('nav.theme')}
              title={t('nav.theme')}
            >
              <span
                className={`absolute top-1/2 h-3 w-3 -translate-y-1/2 rounded-full bg-cyan-400 shadow-[0_0_8px_rgba(34,211,238,0.8)] transition-all duration-300 ${
                  resolvedTheme === 'light' ? 'left-4' : 'left-0.5'
                }`}
              />
            </div>
            <span className="hidden text-xs text-slate-400 md:inline">{t('nav.theme')}</span>
          </div>
        </div>
      </aside>

      {/* Main content */}
      <main className={`flex-1 ${isRtl ? 'mr-16 md:mr-64' : 'ml-16 md:ml-64'}`}>
        <div className="mx-auto max-w-6xl px-4 py-6 md:px-8">
          <div className="animate-rise">{children}</div>
        </div>
      </main>

      <ConfigPanel open={showConfig} onClose={() => setShowConfig(false)} />
    </div>
  );
}