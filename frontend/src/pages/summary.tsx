import { useEffect, useMemo, useState } from 'react';
import { BarChart3, File, FileArchive, FileImage, FileText, FileVideo, Folder, HardDrive, Loader2, Music, RefreshCw } from 'lucide-react';
import { useI18n } from '../i18n';
import { useFileStore } from '../store/fileStore';
import { formatBytes, getExtension } from '../lib/utils';

interface TypeStat {
  label: string;
  count: number;
  size: number;
  icon: typeof File;
  color: string;
  bar: string;
}

const CARD_STYLES = [
  { icon: HardDrive, tile: 'from-cyan-500 to-blue-600', glow: 'shadow-cyan-500/40', text: 'text-cyan-300' },
  { icon: File, tile: 'from-emerald-500 to-teal-600', glow: 'shadow-emerald-500/40', text: 'text-emerald-300' },
  { icon: Folder, tile: 'from-amber-500 to-orange-600', glow: 'shadow-amber-500/40', text: 'text-amber-300' },
  { icon: BarChart3, tile: 'from-violet-500 to-fuchsia-600', glow: 'shadow-violet-500/40', text: 'text-violet-300' },
] as const;

export function Summary() {
  const { t } = useI18n();
  const rootFiles = useFileStore((s) => s.rootFiles);
  const refreshRoot = useFileStore((s) => s.refreshRoot);
  const [isLoading, setIsLoading] = useState(rootFiles.length === 0);
  const [error, setError] = useState<string | null>(null);
  const [attempt, setAttempt] = useState(0);

  useEffect(() => {
    let mounted = true;
    (async () => {
      setIsLoading(true);
      setError(null);
      const res = await refreshRoot();
      if (!mounted) return;
      if (!res.ok) setError(t('summary.loadFailed'));
      setIsLoading(false);
    })();
    return () => {
      mounted = false;
    };
  }, [refreshRoot, attempt, t]);

  const loadAll = () => setAttempt((v) => v + 1);

  const stats = useMemo(() => {
    const dirs = rootFiles.filter((f) => f.isDir);
    const fileList = rootFiles.filter((f) => !f.isDir);
    const totalSize = fileList.reduce((acc, f) => acc + f.size, 0);

    const typeMap = new Map<string, TypeStat>();
    const labelFor = (key: string): string =>
      ({
        images: t('summary.typeImages'),
        video: t('summary.typeVideo'),
        audio: t('summary.typeAudio'),
        archives: t('summary.typeArchives'),
        documents: t('summary.typeDocuments'),
        other: t('summary.typeOther'),
      })[key] ?? key;
    const addType = (key: string, icon: typeof File, color: string, bar: string, size: number, count: number) => {
      const existing = typeMap.get(key) || { label: labelFor(key), count: 0, size: 0, icon, color, bar };
      existing.count += count;
      existing.size += size;
      typeMap.set(key, existing);
    };

    for (const f of fileList) {
      const ext = getExtension(f.name);
      if (['jpg', 'jpeg', 'png', 'gif', 'webp', 'svg', 'bmp'].includes(ext)) {
        addType('images', FileImage, 'text-emerald-400', 'from-emerald-400 to-teal-500', f.size, 1);
      } else if (['mp4', 'mov', 'avi', 'mkv', 'webm'].includes(ext)) {
        addType('video', FileVideo, 'text-purple-400', 'from-purple-400 to-fuchsia-500', f.size, 1);
      } else if (['mp3', 'wav', 'ogg', 'flac'].includes(ext)) {
        addType('audio', Music, 'text-pink-400', 'from-pink-400 to-rose-500', f.size, 1);
      } else if (['zip', 'rar', '7z', 'tar', 'gz'].includes(ext)) {
        addType('archives', FileArchive, 'text-amber-400', 'from-amber-400 to-orange-500', f.size, 1);
      } else if (['txt', 'md', 'log', 'pdf', 'doc', 'docx'].includes(ext)) {
        addType('documents', FileText, 'text-sky-400', 'from-sky-400 to-cyan-500', f.size, 1);
      } else {
        addType('other', File, 'text-slate-400', 'from-slate-400 to-slate-500', f.size, 1);
      }
    }

    return {
      dirs: dirs.length,
      files: fileList.length,
      totalSize,
      types: Array.from(typeMap.values()).sort((a, b) => b.size - a.size),
    };
  }, [rootFiles, t]);

  const maxTypeSize = Math.max(...stats.types.map((t) => t.size), 1);

  if (isLoading) {
    return (
      <div className="flex flex-col items-center justify-center gap-3 py-32 text-slate-500">
        <Loader2 className="h-8 w-8 animate-spin text-cyan-400" />
        <p className="text-sm">{t('summary.loading')}</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex flex-col items-center justify-center gap-4 py-32 text-center">
        <p className="text-sm text-red-400">{error}</p>
        <button
          onClick={loadAll}
          className="inline-flex items-center gap-2 rounded-xl border border-white/10 bg-white/5 px-4 py-2 text-sm font-medium text-slate-300 transition-colors hover:border-cyan-400/40 hover:text-white"
        >
          <RefreshCw className="h-4 w-4" />
          {t('common.retry')}
        </button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">
          <span className="text-gradient">{t('summary.title')}</span>
        </h1>
        <p className="mt-1 text-sm text-slate-400">{t('summary.overview')}</p>
      </div>

      {/* Stat cards */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {[
          { value: formatBytes(stats.totalSize), label: t('summary.totalStorage') },
          { value: String(stats.files), label: t('summary.files') },
          { value: String(stats.dirs), label: t('summary.folders') },
          { value: String(stats.types.length), label: t('summary.fileTypes') },
        ].map((card, i) => {
          const style = CARD_STYLES[i];
          const Icon = style.icon;
          return (
            <div key={card.label} className="glass-panel card-hover p-5">
              <div className="flex items-center gap-3">
                <div className={`relative flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-gradient-to-br ${style.tile} shadow-lg ${style.glow}`}>
                  <Icon className={`h-5 w-5 ${style.text}`} />
                  <span className={`absolute inset-0 rounded-xl bg-gradient-to-br ${style.tile} blur-md opacity-40 animate-glow`} />
                </div>
                <div className="min-w-0">
                  <p className="truncate text-2xl font-bold tracking-tight">{card.value}</p>
                  <p className="text-xs text-slate-500">{card.label}</p>
                </div>
              </div>
            </div>
          );
        })}
      </div>

      {/* Storage breakdown */}
      <div className="glass-panel p-6">
        <h2 className="mb-4 text-lg font-semibold tracking-tight">{t('summary.storageByType')}</h2>
        {stats.types.length === 0 ? (
          <p className="py-8 text-center text-sm text-slate-500">{t('summary.noFilesYet')}</p>
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
                  <div className="h-2 w-full overflow-hidden rounded-full bg-white/5">
                    <div
                      className={`h-full rounded-full bg-gradient-to-r ${type.bar} shadow-[0_0_12px_rgba(34,211,238,0.5)] transition-all duration-700`}
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