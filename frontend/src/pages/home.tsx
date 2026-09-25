import { useCallback, useEffect, useOptimistic, useRef, useState, useTransition } from 'react';
import { AlertTriangle, ArrowUp, ChevronRight, Download, Eye, FolderPlus, History, Link2, Loader2, RefreshCw, Search, Trash2, Upload, X } from 'lucide-react';
import { buildUrl, clearCredentials, api } from '../services/api';
import type { FileItem } from '../services/api';
import { FileIcon } from '../components/FileIcon';
import { FilePreview } from '../components/FilePreview';
import { Modal } from '../components/Modal';
import { PendingUploads } from '../components/PendingUploads';
import { ShareModal } from '../components/ShareModal';
import { VersionHistory } from '../components/VersionHistory';
import { useToast } from '../hooks/useToast';
import { useI18n } from '../i18n';
import { useFileStore } from '../store/fileStore';
import { useAuthStore } from '../store/authStore';
import { formatBytes, formatDate, getExtension, fileTypeColor, viewerKindFor } from '../lib/utils';

type SortKey = 'name' | 'size' | 'modified';
type SortDir = 'asc' | 'desc';

export function Home() {
  const { t } = useI18n();
  const files = useFileStore((s) => s.files);
  const currentPath = useFileStore((s) => s.currentPath);
  const scope = useFileStore((s) => s.scope);
  const isLoading = useFileStore((s) => s.isLoading);
  const loadError = useFileStore((s) => s.error);
  const fetchFiles = useFileStore((s) => s.fetchFiles);
  const navigateTo = useFileStore((s) => s.navigateTo);
  const uploadFiles = useFileStore((s) => s.uploadFiles);
  const deleteItem = useFileStore((s) => s.deleteItem);
  const isUploading = useFileStore((s) => s.isUploading);
  const uploadStatus = useFileStore((s) => s.uploadStatus);
  const isAdmin = useAuthStore((s) => s.user?.isAdmin ?? false);
  const account = useAuthStore((s) => s.user);
  const setUser = useAuthStore((s) => s.setUser);
  const [pendingKey, setPendingKey] = useState(0);

  const isShared = scope === 'shared';
  const readOnly = isShared && !isAdmin;

  const [searchQuery, setSearchQuery] = useState('');
  const [sortKey, setSortKey] = useState<SortKey>('name');
  const [sortDir, setSortDir] = useState<SortDir>('asc');
  const [dragOver, setDragOver] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState<FileItem | null>(null);
  const [shareItem, setShareItem] = useState<FileItem | null>(null);
  const [historyItem, setHistoryItem] = useState<FileItem | null>(null);
  const [previewItem, setPreviewItem] = useState<FileItem | null>(null);
  const closePreview = useCallback(() => setPreviewItem(null), []);
  const [, startTransition] = useTransition();
  const fileInputRef = useRef<HTMLInputElement>(null);
  const { success, error, info } = useToast();

  const [optimisticFiles, removeOptimistic] = useOptimistic(
    files,
    (current, removedPath: string) => current.filter((f) => f.path !== removedPath),
  );

  useEffect(() => {
    fetchFiles('');
  }, [fetchFiles]);

  const handleNavigate = (path: string) => {
    startTransition(() => navigateTo(path));
  };

  const handleUpload = async (fileList: FileList | null) => {
    if (readOnly || !fileList || fileList.length === 0) return;
    const { uploaded, error: err, cancelled } = await uploadFiles(fileList);
    if (cancelled) {
      info(t('home.uploadCancelled'));
      setPendingKey((k) => k + 1);
      return;
    }
    if (err) {
      error(err);
      setPendingKey((k) => k + 1);
      return;
    }
    success(t('home.uploaded', { n: uploaded }));
    setPendingKey((k) => k + 1);
    void refreshAccountUsage();
  };

  const refreshAccountUsage = async () => {
    if (!account) return;
    const result = await api.me();
    if (result.data) setUser(result.data);
  };

  const handleDelete = async (item: FileItem) => {
    setDeleteTarget(null);
    startTransition(async () => {
      removeOptimistic(item.path);
      const result = await deleteItem(item.path);
      if (result.error) {
        error(result.error);
        await fetchFiles(currentPath);
        return;
      }
      success(t('home.deleted', { name: item.name }));
    });
  };

  const handleDownload = async (item: FileItem) => {
    try {
      // Use streaming download to avoid buffering the entire file in memory.
      // Try File System Access API (Chrome/Edge) first for true streaming to disk.
      const url = buildUrl('/api/files/download', { path: item.path });
      const response = await fetch(url, { credentials: 'include' });
      if (!response.ok) {
        if (response.status === 401) {
          clearCredentials();
          throw new Error('Not authorized');
        }
        throw new Error(`Download failed (${response.status})`);
      }

      const suggestedName = item.isDir ? `${item.name}.zip` : item.name;

      // Immediate feedback so the user knows the download is starting.
      info(t('home.downloadStarting'));

      // Modern streaming save (Chromium-based browsers)
      if ('showSaveFilePicker' in window) {
        try {
          const handle = await (window as unknown as { showSaveFilePicker: (options: { suggestedName: string }) => Promise<FileSystemFileHandle> }).showSaveFilePicker({ suggestedName });
          const writable = await handle.createWritable();
          await response.body!.pipeTo(writable);
        } catch (e) {
          const err = e as DOMException;
          if (err && (err.name === 'AbortError' || err.name === 'NotAllowedError')) {
            return;
          }
          throw e;
        }
        success(t('home.downloadComplete'));
        return;
      }

      // Fallback: buffer into blob (Firefox/Safari)
      const blob = await response.blob();
      const blobUrl = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = blobUrl;
      a.download = suggestedName;
      a.click();
      // Keep the object URL alive until the download engine has started.
      window.setTimeout(() => URL.revokeObjectURL(blobUrl), 10_000);
      success(t('home.downloadComplete'));
    } catch (e) {
      error(e instanceof Error ? e.message : t('home.downloadFailed'));
    }
  };

  const handleOpen = (item: FileItem) => {
    if (item.isDir) {
      handleNavigate(item.path);
      return;
    }
    if (viewerKindFor(item.name)) {
      setPreviewItem(item);
    }
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setDragOver(false);
    if (readOnly) return;
    handleUpload(e.dataTransfer.files);
  };

  const toggleSort = (key: SortKey) => {
    if (sortKey === key) {
      setSortDir((d) => (d === 'asc' ? 'desc' : 'asc'));
    } else {
      setSortKey(key);
      setSortDir('asc');
    }
  };

  const filtered = optimisticFiles
    .filter((f) => f.name.toLowerCase().includes(searchQuery.toLowerCase()))
    .sort((a, b) => {
      if (a.isDir !== b.isDir) return a.isDir ? -1 : 1;
      const cmp =
        sortKey === 'name'
          ? a.name.localeCompare(b.name)
          : sortKey === 'size'
            ? a.size - b.size
            : a.modified.localeCompare(b.modified);
      return sortDir === 'asc' ? cmp : -cmp;
    });

  const totalSize = files.filter((f) => !f.isDir).reduce((acc, f) => acc + f.size, 0);
  const breadcrumbs = currentPath ? currentPath.split('/') : [];

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">
            <span className="text-gradient">{isShared ? t('nav.shared') : t('nav.myFiles')}</span>
          </h1>
          <p className="mt-1 text-sm text-slate-400">
            {t('home.items', { n: files.length })}
            {totalSize > 0 && ` · ${t('home.totalSize', { size: formatBytes(totalSize) })}`}
          </p>
          {account && account.quotaBytes && account.quotaBytes > 0 ? (
            <StorageUsage
              used={account.size ?? 0}
              quota={account.quotaBytes}
              files={account.files ?? 0}
            />
          ) : null}
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={() => fetchFiles(currentPath)}
            className="inline-flex items-center gap-2 rounded-xl border border-white/10 bg-white/5 px-3 py-2 text-sm font-medium text-slate-300 backdrop-blur transition-all duration-300 hover:border-cyan-400/40 hover:text-white"
          >
            <RefreshCw className={`h-4 w-4 ${isLoading ? 'animate-spin' : ''}`} />
            <span className="hidden sm:inline">{t('home.refresh')}</span>
          </button>
          {!readOnly && (
            <button
              onClick={() => fileInputRef.current?.click()}
              className="inline-flex items-center gap-2 rounded-xl bg-gradient-to-r from-cyan-500 to-violet-600 px-4 py-2 text-sm font-semibold text-white shadow-lg shadow-cyan-500/30 transition-all duration-300 hover:shadow-cyan-400/50 hover:brightness-110"
            >
              <Upload className="h-4 w-4" />
              {t('home.upload')}
            </button>
          )}
        </div>
      </div>

      {readOnly && (
        <div className="rounded-xl border border-amber-400/30 bg-amber-400/10 px-4 py-2.5 text-sm text-amber-200">
          {t('shared.readOnly')}
        </div>
      )}

      {!readOnly && !isUploading && uploadStatus === 'idle' && (
        <PendingUploads key={pendingKey} onFinished={() => setPendingKey((k) => k + 1)} />
      )}

      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex flex-wrap items-center gap-1 text-sm text-slate-400">
          <button
            onClick={() => handleNavigate('')}
            className="flex items-center gap-1 rounded-lg px-1.5 py-0.5 transition-colors hover:bg-white/10 hover:text-cyan-300"
          >
            <ArrowUp className="h-3.5 w-3.5" />
            {isShared ? t('nav.shared') : t('home.root')}
          </button>
          {breadcrumbs.map((crumb, i) => {
            const path = breadcrumbs.slice(0, i + 1).join('/');
            return (
              <span key={path} className="flex items-center gap-1">
                <ChevronRight className="h-3.5 w-3.5 text-slate-600" />
                <button
                  onClick={() => handleNavigate(path)}
                  className="rounded-lg px-1.5 py-0.5 transition-colors hover:bg-white/10 hover:text-cyan-300"
                >
                  {crumb}
                </button>
              </span>
            );
          })}
        </div>
        <div className="relative w-full sm:w-72">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-cyan-400/70" />
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder={t('home.filterPlaceholder')}
            className="w-full rounded-xl border border-white/10 bg-white/5 py-2 pl-9 pr-8 text-sm text-slate-200 placeholder-slate-500 outline-none backdrop-blur transition-colors focus:border-cyan-400/50 focus:ring-1 focus:ring-cyan-400/30"
          />
          {searchQuery && (
            <button
              onClick={() => setSearchQuery('')}
              className="absolute right-2 top-1/2 -translate-y-1/2 text-slate-500 hover:text-white"
              aria-label={t('home.clearSearch')}
            >
              <X className="h-4 w-4" />
            </button>
          )}
        </div>
      </div>

      <div
        className={`glass-panel overflow-hidden transition-all duration-300 ${
          dragOver
            ? 'border-cyan-400/60 shadow-[0_0_0_1px_rgba(34,211,238,0.4),0_0_40px_-12px_rgba(34,211,238,0.7)]'
            : 'card-hover'
        }`}
        onDragOver={(e) => {
          e.preventDefault();
          setDragOver(true);
        }}
        onDragLeave={() => setDragOver(false)}
        onDrop={handleDrop}
      >
        <div className="hidden grid-cols-12 gap-4 border-b border-white/10 px-4 py-2.5 text-[11px] font-semibold uppercase tracking-widest text-slate-500 md:grid">
          <button className="col-span-6 flex items-center gap-1 text-left hover:text-cyan-300" onClick={() => toggleSort('name')}>
            {t('home.name')}
            {sortKey === 'name' && <span className="text-cyan-400">{sortDir === 'asc' ? '↑' : '↓'}</span>}
          </button>
          <button className="col-span-2 flex items-center gap-1 text-left hover:text-cyan-300" onClick={() => toggleSort('size')}>
            {t('home.size')}
            {sortKey === 'size' && <span className="text-cyan-400">{sortDir === 'asc' ? '↑' : '↓'}</span>}
          </button>
          <button className="col-span-2 flex items-center gap-1 text-left hover:text-cyan-300" onClick={() => toggleSort('modified')}>
            {t('home.modified')}
            {sortKey === 'modified' && <span className="text-cyan-400">{sortDir === 'asc' ? '↑' : '↓'}</span>}
          </button>
          <span className="col-span-2 text-right">{t('home.actions')}</span>
        </div>
        {isLoading ? (
          <div className="flex flex-col items-center justify-center gap-3 py-20 text-slate-500">
            <Loader2 className="h-8 w-8 animate-spin text-cyan-400" />
            <p className="text-sm">{t('home.loadingFiles')}</p>
          </div>
        ) : loadError ? (
          <div className="flex flex-col items-center justify-center gap-3 py-20 text-center">
            <AlertTriangle className="h-8 w-8 text-red-400" />
            <p className="text-sm font-medium text-slate-300">{loadError}</p>
            <button
              onClick={() => fetchFiles(currentPath)}
              className="inline-flex items-center gap-2 rounded-xl border border-white/10 bg-white/5 px-4 py-2 text-sm font-medium text-slate-200 transition-colors hover:border-cyan-400/40 hover:text-white"
            >
              <RefreshCw className="h-4 w-4" />
              {t('common.retry')}
            </button>
          </div>
        ) : filtered.length === 0 ? (
          <div className="flex flex-col items-center justify-center gap-3 py-20 text-slate-500">
            <div className="relative flex h-16 w-16 items-center justify-center rounded-full bg-white/5">
              <FolderPlus className="h-8 w-8 text-cyan-300/70" />
              <span className="absolute inset-0 rounded-full bg-cyan-400/20 blur-md animate-glow" />
            </div>
            <p className="text-sm font-medium text-slate-400">
              {searchQuery ? t('home.noMatches') : t('home.folderEmpty')}
            </p>
            <p className="text-xs text-slate-600">
              {searchQuery ? t('home.tryDifferent') : t('home.uploadOrCreate')}
            </p>
          </div>
        ) : (
          <ul className="divide-y divide-white/5">
            {filtered.map((item) => (
              <li
                key={item.path}
                className="group grid grid-cols-12 items-center gap-4 px-4 py-3 transition-colors hover:bg-cyan-400/5"
                onDoubleClick={() => item.name && handleOpen(item)}
              >
                <div className="col-span-12 flex items-center gap-3 md:col-span-6">
                  <FileIcon name={item.name} isDir={item.isDir} />
                  <button
                    onClick={() => handleOpen(item)}
                    className="truncate text-sm font-medium text-slate-200 transition-colors hover:text-cyan-300"
                    title={item.name}
                  >
                    {item.name}
                  </button>
                  {!item.isDir && (
                    <span className={`rounded-md border px-1.5 py-0.5 text-[10px] font-semibold uppercase ${fileTypeColor(getExtension(item.name))}`}>
                      {getExtension(item.name) || t('home.file')}
                    </span>
                  )}
                </div>
                <div className="col-span-2 hidden text-sm text-slate-400 md:block">
                  {item.isDir ? '—' : formatBytes(item.size)}
                </div>
                <div className="col-span-2 hidden text-sm text-slate-500 md:block">{formatDate(item.modified)}</div>
                <div className="col-span-2 flex items-center justify-end gap-1 opacity-60 transition-opacity group-hover:opacity-100">
                  {!item.isDir && viewerKindFor(item.name) && (
                    <button
                      onClick={() => setPreviewItem(item)}
                      className="rounded-lg p-1.5 text-slate-400 transition-colors hover:bg-white/10 hover:text-cyan-300"
                      title={t('common.preview')}
                    >
                      <Eye className="h-4 w-4" />
                    </button>
                  )}
                  <button
                    onClick={() => handleDownload(item)}
                    className="rounded-lg p-1.5 text-slate-400 transition-colors hover:bg-white/10 hover:text-cyan-300"
                    title={item.isDir ? t('home.downloadZip') : t('home.download')}
                  >
                    <Download className="h-4 w-4" />
                  </button>
                  <button
                    onClick={() => setShareItem(item)}
                    className="rounded-lg p-1.5 text-slate-400 transition-colors hover:bg-white/10 hover:text-cyan-300"
                    title={t('home.share')}
                  >
                    <Link2 className="h-4 w-4" />
                  </button>
                  {!item.isDir && (
                    <button
                      onClick={() => setHistoryItem(item)}
                      className="rounded-lg p-1.5 text-slate-400 transition-colors hover:bg-white/10 hover:text-cyan-300"
                      title={t('home.history')}
                    >
                      <History className="h-4 w-4" />
                    </button>
                  )}
                  {!readOnly && (
                    <button
                      onClick={() => setDeleteTarget(item)}
                      className="rounded-lg p-1.5 text-slate-400 transition-colors hover:bg-red-500/10 hover:text-red-400"
                      title={t('common.delete')}
                    >
                      <Trash2 className="h-4 w-4" />
                    </button>
                  )}
                </div>
              </li>
            ))}
          </ul>
        )}
      </div>

      <input
        ref={fileInputRef}
        type="file"
        multiple
        className="hidden"
        onChange={(e) => {
          handleUpload(e.target.files);
          e.target.value = '';
        }}
      />

      <FilePreview
        item={previewItem}
        onClose={closePreview}
        onSaved={() => fetchFiles(currentPath)}
      />

      <ShareModal item={shareItem} onClose={() => setShareItem(null)} />

      <VersionHistory
        item={historyItem}
        readOnly={readOnly}
        onClose={() => setHistoryItem(null)}
        onRestored={() => fetchFiles(currentPath)}
      />

      <Modal
        open={deleteTarget !== null}
        onClose={() => setDeleteTarget(null)}
        title={t('home.deleteConfirmTitle')}
      >
        <div className="flex items-start gap-3">
          <div className="rounded-full bg-red-500/15 p-2 text-red-300">
            <AlertTriangle className="h-5 w-5" />
          </div>
          <div className="space-y-1">
            <p className="text-sm text-slate-300">
              {t('home.deleteConfirmMessage', { name: deleteTarget?.name ?? '' })}
            </p>
            {deleteTarget?.isDir && (
              <p className="text-xs text-slate-500">{t('home.deleteFolderNote')}</p>
            )}
          </div>
        </div>
        <div className="mt-6 flex justify-end gap-2">
          <button
            onClick={() => setDeleteTarget(null)}
            className="rounded-xl border border-white/10 bg-white/5 px-4 py-2 text-sm text-slate-200 transition-colors hover:bg-white/10"
          >
            {t('common.cancel')}
          </button>
          <button
            onClick={() => deleteTarget && handleDelete(deleteTarget)}
            className="rounded-xl bg-red-500/90 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-red-500"
          >
            {t('common.delete')}
          </button>
        </div>
      </Modal>
    </div>
  );
}

interface StorageUsageProps {
  used: number;
  quota: number;
  files: number;
}

/** Account-level storage gauge shown when the account has a quota. */
function StorageUsage({ used, quota, files }: StorageUsageProps) {
  const { t } = useI18n();
  const pct = quota > 0 ? Math.min(100, (used / quota) * 100) : 0;
  const nearLimit = pct >= 90;
  return (
    <div
      className="mt-3 max-w-xs animate-rise"
      title={`${formatBytes(used)} / ${formatBytes(quota)}`}
    >
      <div className="flex items-center justify-between gap-2 text-[11px] tabular-nums">
        <span className="min-w-0 flex-1 truncate text-slate-400">
          {t('home.storageUsed', {
            used: formatBytes(used),
            quota: formatBytes(quota),
            files,
          })}
        </span>
        <span
          className={`shrink-0 ${nearLimit ? 'font-semibold text-amber-300' : 'text-cyan-300'}`}
          dir="ltr"
        >
          {Math.round(pct)}%
        </span>
      </div>
      <div className="mt-1 h-1.5 w-full overflow-hidden rounded-full bg-white/10">
        <div
          className={`h-full rounded-full transition-[width] duration-500 ease-out ${
            nearLimit ? 'bg-amber-400' : 'bg-gradient-to-r from-cyan-300 via-sky-400 to-violet-400'
          }`}
          style={{ width: `${pct}%` }}
        />
      </div>
    </div>
  );
}