import { useCallback, useEffect, useState } from 'react';
import { Download, History, Loader2, RotateCcw } from 'lucide-react';
import { api, type FileItem } from '../services/api';
import { Modal } from './Modal';
import { useToast } from '../hooks/useToast';
import { useI18n } from '../i18n';
import { formatBytes, formatDate } from '../lib/utils';

interface VersionHistoryProps {
  item: FileItem | null;
  readOnly?: boolean;
  onClose: () => void;
  onRestored?: () => void;
}

export function VersionHistory({ item, readOnly, onClose, onRestored }: VersionHistoryProps) {
  const { t } = useI18n();
  const { success, error } = useToast();
  const [versions, setVersions] = useState<FileItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [restoring, setRestoring] = useState<number | null>(null);

  const itemPath = item?.path ?? '';

  const refresh = useCallback(async () => {
    if (!itemPath) return;
    setLoading(true);
    const result = await api.listVersions(itemPath);
    setLoading(false);
    if (result.error) {
      error(result.error);
      return;
    }
    setVersions(result.data ?? []);
  }, [itemPath, error]);

  useEffect(() => {
    if (!item) return;
    setVersions([]);
    setRestoring(null);
    void refresh();
  }, [item, refresh]);

  const handleDownload = async (v: FileItem) => {
    try {
      const blob = await api.downloadVersion(itemPath, v.version ?? 0);
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = item?.name ?? v.name;
      a.click();
      URL.revokeObjectURL(url);
    } catch (e) {
      error(e instanceof Error ? e.message : t('history.loadFailed'));
    }
  };

  const handleRestore = async (v: FileItem) => {
    if (restoring !== null) return;
    const n = v.version ?? 0;
    setRestoring(n);
    const result = await api.restoreVersion(itemPath, n);
    setRestoring(null);
    if (result.error) {
      error(result.error);
      return;
    }
    success(t('history.restored', { n }));
    void refresh();
    onRestored?.();
  };

  const newestFirst = [...versions].reverse();

  return (
    <Modal
      open={item !== null}
      onClose={onClose}
      title={t('history.title', { name: item?.name ?? '' })}
      size="lg"
    >
      <div className="space-y-4">
        <div className="flex items-center gap-3 rounded-xl border border-cyan-400/30 bg-cyan-400/10 px-3 py-2.5">
          <History className="h-4 w-4 shrink-0 text-cyan-300" />
          <div className="min-w-0 flex-1">
            <p className="text-sm font-semibold text-cyan-100">{t('history.current')}</p>
            <p className="text-xs text-slate-400">
              {item ? `${formatBytes(item.size)} · ${formatDate(item.modified)}` : ''}
            </p>
          </div>
        </div>

        <div className="flex items-center justify-between">
          <label className="text-xs font-semibold uppercase tracking-widest text-slate-500">
            {t('history.previous')}
          </label>
          {loading && <Loader2 className="h-3.5 w-3.5 animate-spin text-cyan-400" />}
        </div>

        {!loading && newestFirst.length === 0 && (
          <div className="rounded-xl border border-dashed border-white/10 px-4 py-6 text-center">
            <p className="text-sm text-slate-400">{t('history.empty')}</p>
            <p className="mt-1 text-xs text-slate-600">{t('history.emptyHint')}</p>
          </div>
        )}

        <ul className="space-y-2">
          {newestFirst.map((v) => (
            <li
              key={v.version}
              className="flex items-center gap-3 rounded-xl border border-white/10 bg-white/5 px-3 py-2.5"
            >
              <div className="min-w-0 flex-1">
                <p className="text-sm font-medium text-slate-200">
                  {t('history.version', { n: v.version ?? 0 })}
                </p>
                <p className="text-xs text-slate-500">
                  {formatBytes(v.size)} · {formatDate(v.modified)}
                </p>
              </div>
              <button
                onClick={() => void handleDownload(v)}
                className="shrink-0 rounded-lg p-1.5 text-slate-400 transition-colors hover:bg-white/10 hover:text-cyan-300"
                title={t('history.download')}
                aria-label={t('history.download')}
              >
                <Download className="h-4 w-4" />
              </button>
              {!readOnly && (
                <button
                  onClick={() => void handleRestore(v)}
                  disabled={restoring !== null}
                  className={`shrink-0 rounded-lg p-1.5 transition-colors disabled:opacity-50 ${
                    restoring === v.version
                      ? 'bg-amber-500/15 text-amber-300'
                      : 'text-slate-400 hover:bg-white/10 hover:text-emerald-300'
                  }`}
                  title={t('history.restore')}
                  aria-label={t('history.restore')}
                >
                  {restoring === v.version ? (
                    <Loader2 className="h-4 w-4 animate-spin" />
                  ) : (
                    <RotateCcw className="h-4 w-4" />
                  )}
                </button>
              )}
            </li>
          ))}
        </ul>
      </div>
    </Modal>
  );
}
