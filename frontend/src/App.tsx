import { useState } from 'react';
import { Layout } from './components/Layout';
import { Modal } from './components/Modal';
import { ToastProvider, useToast } from './components/Toast';
import { Home } from './pages/home';
import { Summary } from './pages/summary';
import { Chat } from './pages/chat';
import { api } from './services/api';
import { Loader2 } from 'lucide-react';

type Page = 'files' | 'summary' | 'chat';

function AppShell() {
  const [page, setPage] = useState<Page>('files');
  const [showNewFolder, setShowNewFolder] = useState(false);
  const [newFolderName, setNewFolderName] = useState('');
  const [isCreatingFolder, setIsCreatingFolder] = useState(false);
  const { success, error } = useToast();

  const handleCreateFolder = async () => {
    if (!newFolderName.trim()) return;
    setIsCreatingFolder(true);
    try {
      const result = await api.createDirectory(newFolderName.trim());
      if (result.error) throw new Error(result.error);
      success('Folder created');
      setNewFolderName('');
      setShowNewFolder(false);
    } catch (e) {
      error(e instanceof Error ? e.message : 'Failed to create folder');
    } finally {
      setIsCreatingFolder(false);
    }
  };

  return (
    <Layout
      active={page}
      onNavigate={setPage}
      onUpload={() => setPage('files')}
      onNewFolder={() => setShowNewFolder(true)}
    >
      {page === 'files' && <Home />}
      {page === 'summary' && <Summary />}
      {page === 'chat' && <Chat />}

      {/* New folder modal */}
      <Modal open={showNewFolder} onClose={() => setShowNewFolder(false)} title="New Folder">
        <form
          onSubmit={(e) => {
            e.preventDefault();
            handleCreateFolder();
          }}
          className="space-y-4"
        >
          <input
            autoFocus
            type="text"
            value={newFolderName}
            onChange={(e) => setNewFolderName(e.target.value)}
            placeholder="Folder name"
            className="w-full rounded-lg border border-slate-700 bg-slate-800 px-3 py-2 text-sm text-slate-200 placeholder-slate-500 outline-none transition-colors focus:border-blue-500/50 focus:ring-1 focus:ring-blue-500/30"
          />
          <div className="flex justify-end gap-2">
            <button
              type="button"
              onClick={() => setShowNewFolder(false)}
              className="rounded-lg px-4 py-2 text-sm font-medium text-slate-400 transition-colors hover:bg-slate-800 hover:text-slate-200"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={!newFolderName.trim() || isCreatingFolder}
              className="inline-flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-semibold text-white transition-colors hover:bg-blue-500 disabled:cursor-not-allowed disabled:opacity-50"
            >
              {isCreatingFolder && <Loader2 className="h-4 w-4 animate-spin" />}
              Create
            </button>
          </div>
        </form>
      </Modal>
    </Layout>
  );
}

export default function App() {
  return (
    <ToastProvider>
      <AppShell />
    </ToastProvider>
  );
}