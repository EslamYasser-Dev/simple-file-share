import { useEffect, useMemo, useState } from 'react';
import { BarChart3, File, FileArchive, FileImage, FileText, FileVideo, Folder, HardDrive, Loader2, Music, RefreshCw } from 'lucide-react';
import { useFileStore } from '../store/fileStore';
import { formatBytes, getExtension } from '../lib/utils';

interface TypeStat {
  label: string;
  count: number;
  size: number;
  icon: typeof File;
  color: string;
}

export function Summary() {
  const rootFiles = useFileStore((s) => s.rootFiles);
  const refreshRoot = useFileStore((s) => s.refreshRoot);
  const [isLoading, setIsLoading] = useState(rootFiles.length === 0);
  const [error, setError] = useState<string | null>(null);
  const [attempt, setAttempt] = useState(0);

  useEffect(() => {
    let mounted = true;
    setIsLoading(true);
    setError(null);
    (async () => {
      const res = await refreshRoot();
      if (!mounted) return;
      if (!res.ok) setError('Failed to load storage summary');
      setIsLoading(false);
    })();
    return () => {
      mounted = false;
    };
  }, [refreshRoot, attempt]);

  const loadAll = () => setAttempt((v) => v + 1);

  const stats = useMemo(() => {
    const dirs = rootFiles.filter((f) => f.isDir);
    const fileList = rootFiles.filter((f) => !f.isDir);
    const totalSize = fileList.reduce((acc, f) => acc + f.size, 0);

    const typeMap = new Map<string, TypeStat>();
    const addType = (key: string, label: string, icon: typeof File, color: string, size: number, count: number) => {
      const existing = typeMap.get(key) || { label, count: 0, size: 0, icon, color };
      existing.count += count;
      existing.size += size;
      typeMap.set(key, existing);
    };

    for (const f of fileList) {
      const ext = getExtension(f.name);
      if (['jpg', 'jpeg', 'png', 'gif', 'webp', 'svg', 'bmp'].includes(ext)) {
        addType('images', 'Images', FileImage, 'text-emerald-400', f.size, 1);
      } else if (['mp4', 'mov', 'avi', 'mkv', 'webm'].includes(ext)) {
        addType('video', 'Video', FileVideo, 'text-purple-400', f.size, 1);
      } else if (['mp3', 'wav', 'ogg', 'flac'].includes(ext)) {
        addType('audio', 'Audio', Music, 'text-pink-400', f.size, 1);
      } else if (['zip', 'rar', '7z', 'tar', 'gz'].includes(ext)) {
        addType('archives', 'Archives', FileArchive, 'text-amber-400', f.size, 1);
      } else if (['txt', 'md', 'log', 'pdf', 'doc', 'docx'].includes(ext)) {
        addType('documents', 'Documents', FileText, 'text-sky-400', f.size, 1);
      } else {
        addType('other', 'Other', File, 'text-slate-400', f.size, 1);
      }
    }

    return {
      dirs: dirs.length,
      files: fileList.length,
      totalSize,
      types: Array.from(typeMap.values()).sort((a, b) => b.size - a.size),
    };
  }, [rootFiles]);

  const maxTypeSize = Math.max(...stats.types.map((t) => t.size), 1);

  if (isLoading) {
    return (
      <div className="flex flex-col items-center justify-center gap-3 py-32 text-slate-500">
        <Loader2 className="h-8 w-8 animate-spin text-blue-500" />
        <p className="text-sm">Loading summary...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex flex-col items-center justify-center gap-4 py-32 text-center">
        <p className="text-sm text-red-400">{error}</p>
        <button
          onClick={loadAll}
          className="inline-flex items-center gap-2 rounded-lg border border-slate-700 px-4 py-2 text-sm font-medium text-slate-300 transition-colors hover:bg-slate-800"
        >
          <RefreshCw className="h-4 w-4" />
          Retry
        </button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">Storage Summary</h1>
        <p className="text-sm text-slate-400">Overview of your files and storage usage</p>
      </div>

      {/* Stat cards */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <div className="rounded-xl border border-slate-800 bg-slate-900/50 p-5">
          <div className="flex items-center gap-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-blue-500/10">
              <HardDrive className="h-5 w-5 text-blue-400" />
            </div>
            <div>
              <p className="text-2xl font-bold">{formatBytes(stats.totalSize)}</p>
              <p className="text-xs text-slate-500">Total Storage</p>
            </div>
          </div>
        </div>
        <div className="rounded-xl border border-slate-800 bg-slate-900/50 p-5">
          <div className="flex items-center gap-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-emerald-500/10">
              <File className="h-5 w-5 text-emerald-400" />
            </div>
            <div>
              <p className="text-2xl font-bold">{stats.files}</p>
              <p className="text-xs text-slate-500">Files</p>
            </div>
          </div>
        </div>
        <div className="rounded-xl border border-slate-800 bg-slate-900/50 p-5">
          <div className="flex items-center gap-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-amber-500/10">
              <Folder className="h-5 w-5 text-amber-400" />
            </div>
            <div>
              <p className="text-2xl font-bold">{stats.dirs}</p>
              <p className="text-xs text-slate-500">Folders</p>
            </div>
          </div>
        </div>
        <div className="rounded-xl border border-slate-800 bg-slate-900/50 p-5">
          <div className="flex items-center gap-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-purple-500/10">
              <BarChart3 className="h-5 w-5 text-purple-400" />
            </div>
            <div>
              <p className="text-2xl font-bold">{stats.types.length}</p>
              <p className="text-xs text-slate-500">File Types</p>
            </div>
          </div>
        </div>
      </div>

      {/* Storage breakdown */}
      <div className="rounded-xl border border-slate-800 bg-slate-900/50 p-6">
        <h2 className="mb-4 text-lg font-semibold">Storage by Type</h2>
        {stats.types.length === 0 ? (
          <p className="py-8 text-center text-sm text-slate-500">No files uploaded yet</p>
        ) : (
          <div className="space-y-4">
            {stats.types.map((type) => {
              const Icon = type.icon;
              const pct = Math.round((type.size / maxTypeSize) * 100);
              return (
                <div key={type.label} className="space-y-1.5">
                  <div className="flex items-center justify-between text-sm">
                    <span className="flex items-center gap-2 font-medium text-slate-300">
                      <Icon className={`h-4 w-4 ${type.color}`} />
                      {type.label}
                      <span className="text-xs text-slate-500">({type.count})</span>
                    </span>
                    <span className="text-slate-400">{formatBytes(type.size)}</span>
                  </div>
                  <div className="h-1.5 w-full overflow-hidden rounded-full bg-slate-800">
                    <div
                      className="h-full rounded-full bg-gradient-to-r from-blue-500 to-purple-500 transition-all duration-500"
                      style={{ width: `${pct}%` }}
                    />
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
}