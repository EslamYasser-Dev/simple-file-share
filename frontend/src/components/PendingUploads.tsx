import { useCallback, useEffect, useRef, useState } from 'react';
import { AlertTriangle, FileUp, Loader2, RefreshCw, Trash2 } from 'lucide-react';
import { api, discardUploadSession, uploadFingerprint, type PendingUploadSession } from '../services/api';
import { useFileStore } from '../store/fileStore';
import { useI18n } from '../i18n';
import { useToast } from '../hooks/useToast';
import { formatBytes } from '../lib/utils';

/**
 * Interrupted resumable uploads from the server. After a reload the browser no
 * longer holds File handles, so resume asks the user to re-select the same
 * file; fingerprint matching continues from the staged offset.
 */
export function PendingUploads({ onFinished }: { onFinished?: () => void }) {
  const { t } = useI18n();
  const { success, error } = useToast();
  const uploadFiles = useFileStore((s) => s.uploadFiles);
  const isUploading = useFileStore((s) => s.isUploading);
  const [items, setItems] = useState<PendingUploadSession[]>([]);
  const [loading, setLoading] = useState(true);
  const [busyId, setBusyId] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const resumeTargetRef = useRef<PendingUploadSession | null>(null);

  const refresh = useCallback(async () => {
    setLoading(true);
    const result = await api.listUploadSessions();
    if (result.error) {
      setItems([]);
    } else {
      setItems(result.data ?? []);
    }
    setLoading(false);
  }, []);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  const openResume = (session: PendingUploadSession) => {
    resumeTargetRef.current = session;
    fileInputRef.current?.click();
  };

  const handleResumeFile = async (list: FileList | null) => {
    const session = resumeTargetRef.current;
    resumeTargetRef.current = null;
    if (!session || !list || list.length === 0) return;
    const file = list[0];
    if (fileInputRef.current) fileInputRef.current.value = '';

    const expected = uploadFingerprint(file, session.destination || '');
    if (session.fingerprint && session.fingerprint !== expected) {
      error(t('pending.fingerprintMismatch', { name: session.filename }));
      return;
    }
    if (file.name !== session.filename || file.size !== session.size) {
      error(t('pending.fileMismatch', { name: session.filename }));
      return;
    }

    setBusyId(session.id);
    const files = new DataTransfer();
    files.items.add(file);
    const result = await uploadFiles(files.files, session.destination || undefined);
    setBusyId(null);
    if (result.error) {
      error(result.error);
    } else if (result.cancelled) {
      /* user paused/cancelled mid-resume — list refresh below still useful */
    } else {
      success(t('pending.resumed', { name: session.filename }));
      onFinished?.();
    }
    await refresh();
  };

  const discard = async (session: PendingUploadSession) => {
    setBusyId(session.id);
    const result = await discardUploadSession(session.id);
    setBusyId(null);
    if (result.error) {
      error(result.error);
      return;
    }
    setItems((prev) => prev.filter((x) => x.id !== session.id));
  };

  if (loading) {
    return (
      <div className="glass-panel flex items-center gap-3 px-4 py-3 text-sm text-slate-400">
        <Loader2 className="h-4 w-4 animate-spin text-cyan-400" />
        {t('pending.loading')}
      </div>
    );
  }

  if (items.length === 0) return null;

  return (
    <section className="glass-panel overflow-hidden animate-rise">
      <div className="flex flex-wrap items-center justify-between gap-2 border-b border-white/10 px-4 py-2.5">
        <div className="flex items-center gap-2 text-[11px] font-semibold uppercase tracking-widest text-slate-500">
          <FileUp className="h-3.5 w-3.5" />
          {t('pending.title', { n: items.length })}
        </div>
        <button
          type="button"
          onClick={() => void refresh()}
          className="rounded-lg p-1 text-slate-400 transition-colors hover:bg-white/10 hover:text-white"
          aria-label={t('common.retry')}
          title={t('common.retry')}
        >
          <RefreshCw className="h-3.5 w-3.5" />
        </button>
      </div>
      <p className="border-b border-white/5 px-4 py-2 text-xs text-slate-500">{t('pending.hint')}</p>
      <ul className="divide-y divide-white/5">
        {items.map((session) => {
          const percent =
            session.size > 0 ? Math.min(100, Math.round((session.offset / session.size) * 100)) : 0;
          const busy = busyId === session.id || isUploading;
          return (
            <li key={session.id} className="px-4 py-3">
              <div className="flex flex-wrap items-center gap-3">
                <div className="min-w-0 flex-1">
                  <p className="truncate text-sm font-medium text-slate-200" title={session.filename}>
                    {session.filename}
                  </p>
                  <p className="truncate text-xs text-slate-500">
                    {session.destination ? `${session.destination}/` : ''}
                    {formatBytes(session.offset)} / {formatBytes(session.size)}
                    {session.expiresAt ? ` · ${t('pending.expires')}` : ''}
                  </p>
                  <div className="mt-1.5 h-1 w-full overflow-hidden rounded-full bg-white/10">
                    <div
                      className="h-full bg-gradient-to-r from-amber-300 to-cyan-400 transition-[width] duration-300"
                      style={{ width: `${percent}%` }}
                    />
                  </div>
                </div>
                <div className="flex shrink-0 items-center gap-2">
                  <span className="tabular-nums text-[11px] text-slate-500">{percent}%</span>
                  <button
                    type="button"
                    onClick={() => openResume(session)}
                    disabled={busy}
                    className="inline-flex items-center gap-1.5 rounded-xl bg-gradient-to-r from-cyan-500 to-violet-600 px-3 py-1.5 text-xs font-semibold text-white shadow-lg shadow-cyan-500/25 transition-all hover:brightness-110 disabled:cursor-not-allowed disabled:opacity-50"
                  >
                    {busyId === session.id ? (
                      <Loader2 className="h-3.5 w-3.5 animate-spin" />
                    ) : (
                      <FileUp className="h-3.5 w-3.5" />
                    )}
                    {t('pending.resume')}
                  </button>
                  <button
                    type="button"
                    onClick={() => void discard(session)}
                    disabled={busy}
                    className="rounded-lg border border-white/10 bg-white/5 p-2 text-slate-400 transition-colors hover:border-red-400/40 hover:text-red-300 disabled:opacity-50"
                    aria-label={t('pending.discard')}
                    title={t('pending.discard')}
                  >
                    <Trash2 className="h-3.5 w-3.5" />
                  </button>
                </div>
              </div>
            </li>
          );
        })}
      </ul>

      <input
        ref={fileInputRef}
        type="file"
        className="hidden"
        onChange={(e) => void handleResumeFile(e.target.files)}
      />

      {!isUploading && items.length > 0 && (
        <div className="flex items-start gap-2 border-t border-white/5 bg-amber-400/5 px-4 py-2.5 text-[11px] text-amber-200/80">
          <AlertTriangle className="mt-0.5 h-3.5 w-3.5 shrink-0" />
          <span>{t('pending.selectSameFile')}</span>
        </div>
      )}
    </section>
  );
}
