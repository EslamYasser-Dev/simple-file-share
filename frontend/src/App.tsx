import { useActionState, useCallback, useEffect, useState, useTransition } from 'react';
import { Layout } from './components/Layout';
import type { Page } from './components/Layout';
import { Modal } from './components/Modal';
import { ToastProvider } from './components/Toast';
import { useToast } from './hooks/useToast';
import { Home } from './pages/home';
import { Summary } from './pages/summary';
import { Chat } from './pages/chat';
import { Login } from './pages/login';
import { Register } from './pages/register';
import { Admin } from './pages/admin';
import { api, clearCredentials, setCredentials } from './services/api';
import { useFileStore } from './store/fileStore';
import { useAuthStore } from './store/authStore';
import { useI18n } from './i18n';
import { AlertCircle, Loader2, RefreshCw } from 'lucide-react';

function AppShell({ onSignOut }: { onSignOut: () => void }) {
  const { t } = useI18n();
  const user = useAuthStore((s) => s.user);
  const isAdmin = user?.isAdmin ?? false;
  const setScope = useFileStore((s) => s.setScope);
  const [page, setPage] = useState<Page>('files');
  const [showNewFolder, setShowNewFolder] = useState(false);
  const [, startTransition] = useTransition();

  const navigate = (next: Page) => {
    if (next === 'files' || next === 'shared') {
      setScope(next === 'shared' ? 'shared' : 'files');
    }
    startTransition(() => setPage(next));
  };

  return (
    <Layout
      active={page}
      onNavigate={navigate}
      onUpload={() => navigate('files')}
      onNewFolder={() => setShowNewFolder(true)}
      onSignOut={onSignOut}
    >
      {page === 'files' && <Home />}
      {page === 'shared' && <Home />}
      {page === 'summary' && <Summary />}
      {page === 'chat' && <Chat />}
      {page === 'admin' && isAdmin && <Admin />}

      {/* New folder modal (mounted only while open, so its action state resets) */}
      {showNewFolder && (
        <Modal open onClose={() => setShowNewFolder(false)} title={t('nav.newFolder')}>
          <NewFolderForm onClose={() => setShowNewFolder(false)} />
        </Modal>
      )}
    </Layout>
  );
}

interface FolderFormState {
  message?: string;
  error?: string;
}

/** New-folder form driven by React 19's useActionState + form actions. */
function NewFolderForm({ onClose }: { onClose: () => void }) {
  const createDirectory = useFileStore((s) => s.createDirectory);
  const { success, error } = useToast();
  const { t } = useI18n();

  const [state, submit, isPending] = useActionState(
    async (_prev: FolderFormState, formData: FormData): Promise<FolderFormState> => {
      const name = String(formData.get('name') ?? '').trim();
      if (!name) return { error: t('folder.nameRequired') };
      const result = await createDirectory(name);
      if (result.error) return { error: result.error };
      success(t('folder.created'));
      return { message: 'created' };
    },
    {},
  );

  useEffect(() => {
    if (state.error) {
      error(state.error);
    }
  }, [state.error, error]);

  useEffect(() => {
    if (state.message) {
      onClose();
    }
  }, [state.message, onClose]);

  return (
    <form action={submit} className="space-y-4">
      <input
        name="name"
        autoFocus
        type="text"
        placeholder={t('folder.namePlaceholder')}
        className="w-full rounded-xl border border-white/10 bg-white/5 px-3 py-2.5 text-sm text-slate-100 placeholder-slate-500 outline-none backdrop-blur transition-all focus:border-cyan-400/50 focus:ring-1 focus:ring-cyan-400/30"
      />
      {state.error && <p className="text-xs text-red-400">{state.error}</p>}
      <div className="flex justify-end gap-2">
        <button
          type="button"
          onClick={onClose}
          className="rounded-lg px-4 py-2 text-sm font-medium text-slate-400 transition-colors hover:bg-white/10 hover:text-white"
        >
          {t('common.cancel')}
        </button>
        <button
          type="submit"
          disabled={isPending}
          className="inline-flex items-center gap-2 rounded-xl bg-gradient-to-r from-cyan-500 to-violet-600 px-4 py-2 text-sm font-semibold text-white shadow-lg shadow-cyan-500/30 transition-all duration-300 hover:brightness-110 disabled:cursor-not-allowed disabled:opacity-50 disabled:shadow-none"
        >
          {isPending && <Loader2 className="h-4 w-4 animate-spin" />}
          {t('common.create')}
        </button>
      </div>
    </form>
  );
}

export default function App() {
  return (
    <ToastProvider>
      <AuthGate />
    </ToastProvider>
  );
}

/** Probes the API on mount; shows the login screen when credentials are required. */
function AuthGate() {
  const { t } = useI18n();
  const setUser = useAuthStore((s) => s.setUser);
  const setSignupEnabled = useAuthStore((s) => s.setSignupEnabled);
  const signupEnabled = useAuthStore((s) => s.signupEnabled);
  const [status, setStatus] = useState<'checking' | 'login' | 'register' | 'ready' | 'error'>('checking');
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const loadAuthInfo = useCallback(async () => {
    const result = await api.authInfo();
    if (!result.error) {
      setSignupEnabled(Boolean(result.data?.signupEnabled));
    }
  }, [setSignupEnabled]);

  const probe = useCallback(async () => {
    setStatus('checking');
    setErrorMessage(null);
    await loadAuthInfo();
    const result = await api.me();
    if (result.unauthorized) {
      setStatus('login');
      return;
    }
    if (result.error) {
      setErrorMessage(result.error);
      setStatus('error');
      return;
    }
    if (result.data) setUser(result.data);
    setStatus('ready');
  }, [loadAuthInfo, setUser]);

  useEffect(() => {
    void probe();
  }, [probe]);

  const handleLogin = useCallback(
    async (username: string, password: string): Promise<string | null> => {
      setCredentials(username, password);
      const result = await api.me();
      if (result.unauthorized) {
        clearCredentials();
        return t('auth.invalidCredentials');
      }
      if (result.error) {
        return result.error;
      }
      if (result.data) setUser(result.data);
      setStatus('ready');
      return null;
    },
    [setUser, t],
  );

  const handleSignOut = useCallback(() => {
    clearCredentials();
    useAuthStore.getState().clear();
    useFileStore.getState().setScope('files');
    setStatus('login');
  }, []);

  if (status === 'checking') {
    return (
      <div className="flex min-h-screen flex-col items-center justify-center gap-3 bg-slate-950 text-slate-500">
        <Loader2 className="h-8 w-8 animate-spin text-blue-500" />
        <p className="text-sm">{t('auth.connecting')}</p>
      </div>
    );
  }

  if (status === 'error') {
    return (
      <div className="flex min-h-screen flex-col items-center justify-center gap-4 bg-slate-950 px-4 text-center">
        <AlertCircle className="h-10 w-10 text-red-400" />
        <div>
          <p className="text-sm font-medium text-slate-300">{t('auth.unableToReach')}</p>
          <p className="mt-1 text-xs text-slate-500">{errorMessage}</p>
        </div>
        <button
          onClick={() => void probe()}
          className="inline-flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-semibold text-white transition-colors hover:bg-blue-500"
        >
          <RefreshCw className="h-4 w-4" />
          {t('common.retry')}
        </button>
      </div>
    );
  }

  if (status === 'register') {
    return (
      <Register
        onRegistered={async (username, password) => {
          const err = await handleLogin(username, password);
          return err;
        }}
        onBack={() => setStatus('login')}
      />
    );
  }

  if (status === 'login') {
    return (
      <Login
        onLogin={handleLogin}
        signupEnabled={signupEnabled}
        onShowRegister={() => setStatus('register')}
      />
    );
  }

  return <AppShell onSignOut={handleSignOut} />;
}