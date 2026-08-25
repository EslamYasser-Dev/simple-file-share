import { useActionState, useEffect, useState, useTransition } from 'react';
import { Layout } from './components/Layout';
import { Modal } from './components/Modal';
import { ToastProvider } from './components/Toast';
import { useToast } from './hooks/useToast';
import { Home } from './pages/home';
import { Summary } from './pages/summary';
import { Chat } from './pages/chat';
import { Login } from './pages/login';
import { api, clearCredentials, setCredentials } from './services/api';
import { useFileStore } from './store/fileStore';
import { Loader2 } from 'lucide-react';

type Page = 'files' | 'summary' | 'chat';

function AppShell() {
  const [page, setPage] = useState<Page>('files');
  const [showNewFolder, setShowNewFolder] = useState(false);
  const [, startTransition] = useTransition();

  const navigate = (next: Page) => {
    startTransition(() => setPage(next));
  };

  return (
    <Layout
      active={page}
      onNavigate={navigate}
      onUpload={() => navigate('files')}
      onNewFolder={() => setShowNewFolder(true)}
    >
      {page === 'files' && <Home />}
      {page === 'summary' && <Summary />}
      {page === 'chat' && <Chat />}

      {/* New folder modal (mounted only while open, so its action state resets) */}
      {showNewFolder && (
        <Modal open onClose={() => setShowNewFolder(false)} title="New Folder">
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

  const [state, submit, isPending] = useActionState(
    async (_prev: FolderFormState, formData: FormData): Promise<FolderFormState> => {
      const name = String(formData.get('name') ?? '').trim();
      if (!name) return { error: 'Folder name is required' };
      const result = await createDirectory(name);
      if (result.error) return { error: result.error };
      success('Folder created');
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
        placeholder="Folder name"
        className="w-full rounded-lg border border-slate-700 bg-slate-800 px-3 py-2 text-sm text-slate-200 placeholder-slate-500 outline-none transition-colors focus:border-blue-500/50 focus:ring-1 focus:ring-blue-500/30"
      />
      {state.error && <p className="text-xs text-red-400">{state.error}</p>}
      <div className="flex justify-end gap-2">
        <button
          type="button"
          onClick={onClose}
          className="rounded-lg px-4 py-2 text-sm font-medium text-slate-400 transition-colors hover:bg-slate-800 hover:text-slate-200"
        >
          Cancel
        </button>
        <button
          type="submit"
          disabled={isPending}
          className="inline-flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-semibold text-white transition-colors hover:bg-blue-500 disabled:cursor-not-allowed disabled:opacity-50"
        >
          {isPending && <Loader2 className="h-4 w-4 animate-spin" />}
          Create
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
  const [status, setStatus] = useState<'checking' | 'login' | 'ready'>('checking');

  useEffect(() => {
    let cancelled = false;
    (async () => {
      const result = await api.listFiles('');
      if (cancelled) return;
      setStatus(result.unauthorized ? 'login' : 'ready');
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  const handleLogin = async (username: string, password: string): Promise<string | null> => {
    setCredentials(username, password);
    const probe = await api.listFiles('');
    if (probe.unauthorized) {
      clearCredentials();
      return 'Invalid username or password';
    }
    setStatus('ready');
    return null;
  };

  if (status === 'checking') {
    return (
      <div className="flex min-h-screen flex-col items-center justify-center gap-3 bg-slate-950 text-slate-500">
        <Loader2 className="h-8 w-8 animate-spin text-blue-500" />
        <p className="text-sm">Connecting...</p>
      </div>
    );
  }

  if (status === 'login') {
    return <Login onLogin={handleLogin} />;
  }

  return <AppShell />;
}