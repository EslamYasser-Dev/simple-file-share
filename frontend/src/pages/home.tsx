import { useEffect, useOptimistic, useRef, useState, useTransition } from 'react';
import { ArrowUp, ChevronRight, Download, FolderPlus, Loader2, RefreshCw, Search, Trash2, Upload, X } from 'lucide-react';
import { api } from '../services/api';
import type { FileItem } from '../services/api';
import { FileIcon } from '../components/FileIcon';
import { useToast } from '../hooks/useToast';
import { useFileStore } from '../store/fileStore';
import { formatBytes, formatDate, getExtension, fileTypeColor } from '../lib/utils';

type SortKey = 'name' | 'size' | 'modified';
type SortDir = 'asc' | 'desc';

export function Home() {
  const files = useFileStore((s) => s.files);
  const currentPath = useFileStore((s) => s.currentPath);
  const isLoading = useFileStore((s) => s.isLoading);
  const isUploading = useFileStore((s) => s.isUploading);
  const uploadProgress = useFileStore((s) => s.uploadProgress);
  const fetchFiles = useFileStore((s) => s.fetchFiles);
  const navigateTo = useFileStore((s) => s.navigateTo);
  const uploadFiles = useFileStore((s) => s.uploadFiles);
  const deleteItem = useFileStore((s) => s.deleteItem);

  const [searchQuery, setSearchQuery] = useState('');
  const [sortKey, setSortKey] = useState<SortKey>('name');
  const [sortDir, setSortDir] = useState<SortDir>('asc');
  const [dragOver, setDragOver] = useState(false);
  const [, startTransition] = useTransition();
  const fileInputRef = useRef<HTMLInputElement>(null);
  const { success, error } = useToast();

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
    if (!fileList || fileList.length === 0) return;
    const { uploaded, error: err } = await uploadFiles(fileList);
    if (err) {
      error(err);
      return;
    }
    success(`Uploaded ${uploaded} file(s) successfully`);
  };

  const handleDelete = async (item: FileItem) => {
    if (!window.confirm(`Delete "${item.name}"? This cannot be undone.`)) return;
    startTransition(() => {
      removeOptimistic(item.path);
    });
    const result = await deleteItem(item.path);
    if (result.error) {
      error(result.error);
      await fetchFiles(currentPath);
      return;
    }
    success(`Deleted "${item.name}"`);
  };

  const handleDownload = async (item: FileItem) => {
    try {
      const blob = await api.downloadFile(item.isDir ? `${item.path}.zip` : item.path);
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = item.isDir ? `${item.name}.zip` : item.name;
      a.click();
      URL.revokeObjectURL(url);
    } catch (e) {
      error(e instanceof Error ? e.message : 'Download failed');
    }
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setDragOver(false);
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
      let cmp = 0;
      if (sortKey === 'name') cmp = a.name.localeCompare(b.name);
      else if (sortKey === 'size') cmp = a.size - b.size;
      else cmp = a.modified.localeCompare(b.modified);
      return sortDir === 'asc' ? cmp : -cmp;
    });

  const totalSize = files.filter((f) => !f.isDir).reduce((acc, f) => acc + f.size, 0);
  const breadcrumbs = currentPath ? currentPath.split('/') : [];

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">My Files</h1>
          <p className="text-sm text-slate-400">
            {files.length} item{files.length !== 1 ? 's' : ''}
            {totalSize > 0 && ` · ${formatBytes(totalSize)}`}
          </p>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={() => fetchFiles(currentPath)}
            className="inline-flex items-center gap-2 rounded-lg border border-slate-700 px-3 py-2 text-sm font-medium text-slate-300 transition-colors hover:bg-slate-800"
          >
            <RefreshCw className={`h-4 w-4 ${isLoading ? 'animate-spin' : ''}`} />
            <span className="hidden sm:inline">Refresh</span>
          </button>
          <button
            onClick={() => fileInputRef.current?.click()}
            className="inline-flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-semibold text-white transition-colors hover:bg-blue-500"
          >
            <Upload className="h-4 w-4" />
            Upload
          </button>
        </div>
      </div>
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex items-center gap-1 text-sm text-slate-400">
          <button
            onClick={() => handleNavigate('')}
            className="flex items-center gap-1 rounded px-1.5 py-0.5 transition-colors hover:bg-slate-800 hover:text-slate-200"
          >
            <ArrowUp className="h-3.5 w-3.5" />
            Root
          </button>
          {breadcrumbs.map((crumb, i) => {
            const path = breadcrumbs.slice(0, i + 1).join('/');
            return (
              <span key={path} className="flex items-center gap-1">
                <ChevronRight className="h-3.5 w-3.5 text-slate-600" />
                <button
                  onClick={() => handleNavigate(path)}
                  className="rounded px-1.5 py-0.5 transition-colors hover:bg-slate-800 hover:text-slate-200"
                >
                  {crumb}
                </button>
              </span>
            );
          })}
        </div>
        <div className="relative w-full sm:w-64">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-500" />
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder="Filter files..."
            className="w-full rounded-lg border border-slate-800 bg-slate-900 py-2 pl-9 pr-8 text-sm text-slate-200 placeholder-slate-500 outline-none transition-colors focus:border-blue-500/50 focus:ring-1 focus:ring-blue-500/30"
          />
          {searchQuery && (
            <button
              onClick={() => setSearchQuery('')}
              className="absolute right-2 top-1/2 -translate-y-1/2 text-slate-500 hover:text-slate-300"
              aria-label="Clear search"
            >
              <X className="h-4 w-4" />
            </button>
          )}
        </div>
      </div>

      {isUploading && (
        <div className="rounded-lg border border-blue-500/30 bg-blue-500/5 p-4">
          <div className="mb-2 flex items-center justify-between text-sm">
            <span className="flex items-center gap-2 text-blue-300">
              <Loader2 className="h-4 w-4 animate-spin" />
              Uploading...
            </span>
            <span className="text-blue-300">{uploadProgress}%</span>
          </div>
          <div className="h-1.5 w-full overflow-hidden rounded-full bg-slate-800">
            <div
              className="h-full rounded-full bg-blue-500 transition-all duration-300"
              style={{ width: `${uploadProgress}%` }}
            />
          </div>
        </div>
      )}

      <div
        className={`overflow-hidden rounded-xl border transition-colors ${
          dragOver ? 'border-blue-500/50 bg-blue-500/5' : 'border-slate-800 bg-slate-900/50'
        }`}
        onDragOver={(e) => {
          e.preventDefault();
          setDragOver(true);
        }}
        onDragLeave={() => setDragOver(false)}
        onDrop={handleDrop}
      >
        <div className="hidden grid-cols-12 gap-4 border-b border-slate-800 px-4 py-2.5 text-xs font-semibold uppercase tracking-wider text-slate-500 md:grid">
          <button className="col-span-6 flex items-center gap-1 text-left hover:text-slate-300" onClick={() => toggleSort('name')}>
            Name
            {sortKey === 'name' && <span className="text-blue-400">{sortDir === 'asc' ? '↑' : '↓'}</span>}
          </button>
          <button className="col-span-2 flex items-center gap-1 text-left hover:text-slate-300" onClick={() => toggleSort('size')}>
            Size
            {sortKey === 'size' && <span className="text-blue-400">{sortDir === 'asc' ? '↑' : '↓'}</span>}
          </button>
          <button className="col-span-2 flex items-center gap-1 text-left hover:text-slate-300" onClick={() => toggleSort('modified')}>
            Modified
            {sortKey === 'modified' && <span className="text-blue-400">{sortDir === 'asc' ? '↑' : '↓'}</span>}
          </button>
          <span className="col-span-2 text-right">Actions</span>
        </div>
        {isLoading ? (
          <div className="flex flex-col items-center justify-center gap-3 py-20 text-slate-500">
            <Loader2 className="h-8 w-8 animate-spin text-blue-500" />
            <p className="text-sm">Loading files...</p>
          </div>
        ) : filtered.length === 0 ? (
          <div className="flex flex-col items-center justify-center gap-3 py-20 text-slate-500">
            <div className="flex h-16 w-16 items-center justify-center rounded-full bg-slate-800/50">
              <FolderPlus className="h-8 w-8" />
            </div>
            <p className="text-sm font-medium text-slate-400">
              {searchQuery ? 'No files match your search' : 'This folder is empty'}
            </p>
            <p className="text-xs text-slate-600">
              {searchQuery ? 'Try a different search term' : 'Upload files or create a new folder to get started'}
            </p>
          </div>
        ) : (
          <ul className="divide-y divide-slate-800/50">
            {filtered.map((item) => (
              <li
                key={item.path}
                className="group grid grid-cols-12 items-center gap-4 px-4 py-3 transition-colors hover:bg-slate-800/30"
                onDoubleClick={() => item.isDir && handleNavigate(item.path)}
              >
                <div className="col-span-12 flex items-center gap-3 md:col-span-6">
                  <FileIcon name={item.name} isDir={item.isDir} />
                  <button
                    onClick={() => item.isDir && handleNavigate(item.path)}
                    className="truncate text-sm font-medium text-slate-200 hover:text-blue-400"
                    title={item.name}
                  >
                    {item.name}
                  </button>
                  {!item.isDir && (
                    <span className={`rounded border px-1.5 py-0.5 text-[10px] font-semibold uppercase ${fileTypeColor(getExtension(item.name))}`}>
                      {getExtension(item.name) || 'file'}
                    </span>
                  )}
                </div>
                <div className="col-span-2 hidden text-sm text-slate-400 md:block">
                  {item.isDir ? '—' : formatBytes(item.size)}
                </div>
                <div className="col-span-2 hidden text-sm text-slate-500 md:block">{formatDate(item.modified)}</div>
                <div className="col-span-2 flex items-center justify-end gap-1 opacity-0 transition-opacity group-hover:opacity-100">
                  <button
                    onClick={() => handleDownload(item)}
                    className="rounded p-1.5 text-slate-400 transition-colors hover:bg-slate-700 hover:text-blue-400"
                    title={item.isDir ? 'Download as ZIP' : 'Download'}
                  >
                    <Download className="h-4 w-4" />
                  </button>
                  <button
                    onClick={() => handleDelete(item)}
                    className="rounded p-1.5 text-slate-400 transition-colors hover:bg-red-500/10 hover:text-red-400"
                    title="Delete"
                  >
                    <Trash2 className="h-4 w-4" />
                  </button>
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
    </div>
  );
}
