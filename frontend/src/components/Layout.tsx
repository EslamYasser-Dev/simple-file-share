import type { ReactNode } from 'react';
import { FolderOpen, LayoutDashboard, MessageSquare, HardDrive, FolderPlus, Upload } from 'lucide-react';

interface LayoutProps {
  active: 'files' | 'summary' | 'chat';
  onNavigate: (page: 'files' | 'summary' | 'chat') => void;
  onUpload: () => void;
  onNewFolder: () => void;
  children: ReactNode;
}

const NAV_ITEMS = [
  { key: 'files', label: 'My Files', icon: FolderOpen },
  { key: 'summary', label: 'Summary', icon: LayoutDashboard },
  { key: 'chat', label: 'Activity', icon: MessageSquare },
] as const;

export function Layout({ active, onNavigate, onUpload, onNewFolder, children }: LayoutProps) {
  return (
    <div className="flex min-h-screen bg-slate-950 text-slate-100">
      {/* Sidebar */}
      <aside className="fixed inset-y-0 left-0 z-30 flex w-16 flex-col border-r border-slate-800 bg-slate-900/50 backdrop-blur-xl md:w-64">
        {/* Logo */}
        <div className="flex items-center gap-3 border-b border-slate-800 px-4 py-5">
          <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-gradient-to-br from-blue-500 to-purple-600">
            <HardDrive className="h-5 w-5 text-white" />
          </div>
          <div className="hidden md:block">
            <h1 className="text-sm font-bold tracking-tight">FileShare</h1>
            <p className="text-[10px] uppercase tracking-widest text-slate-500">Cloud Storage</p>
          </div>
        </div>

        {/* Nav */}
        <nav className="flex-1 space-y-1 px-2 py-4">
          {NAV_ITEMS.map(({ key, label, icon: Icon }) => (
            <button
              key={key}
              onClick={() => onNavigate(key)}
              className={`flex w-full items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition-colors ${
                active === key
                  ? 'bg-blue-500/10 text-blue-400'
                  : 'text-slate-400 hover:bg-slate-800/60 hover:text-slate-200'
              }`}
            >
              <Icon className="h-5 w-5 shrink-0" />
              <span className="hidden md:inline">{label}</span>
            </button>
          ))}
        </nav>

        {/* Actions */}
        <div className="space-y-2 border-t border-slate-800 p-3">
          <button
            onClick={onUpload}
            className="flex w-full items-center justify-center gap-2 rounded-lg bg-blue-600 px-3 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-blue-500"
          >
            <Upload className="h-4 w-4" />
            <span className="hidden md:inline">Upload</span>
          </button>
          <button
            onClick={onNewFolder}
            className="flex w-full items-center justify-center gap-2 rounded-lg border border-slate-700 px-3 py-2.5 text-sm font-medium text-slate-300 transition-colors hover:bg-slate-800"
          >
            <FolderPlus className="h-4 w-4" />
            <span className="hidden md:inline">New Folder</span>
          </button>
        </div>
      </aside>

      {/* Main content */}
      <main className="ml-16 flex-1 md:ml-64">
        <div className="mx-auto max-w-6xl px-4 py-6 md:px-8">{children}</div>
      </main>
    </div>
  );
}