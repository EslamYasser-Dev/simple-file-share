import { useState } from 'react';
import { AlertTriangle, ChevronDown, ChevronUp, X, Zap } from 'lucide-react';
import { Modal } from './Modal';
import { useI18n } from '../i18n';
import { useFileStore } from '../store/fileStore';
import { formatBytes, formatSpeed } from '../lib/utils';

/**
 * Floating upload indicator. It lives in the toast layer instead of the page
 * layout, so progress stays visible without moving surrounding content.
 */
export function UploadToast() {
  const { t } = useI18n();
  const isUploading = useFileStore((s) => s.isUploading);
  const activeUpload = useFileStore((s) => s.activeUpload);
  const uploadProgress = useFileStore((s) => s.uploadProgress);
  const uploadSpeed = useFileStore((s) => s.uploadSpeed);
  const cancelUpload = useFileStore((s) => s.cancelUpload);
  const [collapsed, setCollapsed] = useState(false);
  const [confirmingCancel, setConfirmingCancel] = useState(false);

  if (!isUploading || !activeUpload) {
    if (!confirmingCancel) return null;
  }

  const activeFiles = activeUpload ? activeUpload.files.filter((file) => file.status === 'active') : [];
  const queuedFiles = activeUpload ? activeUpload.files.filter((file) => file.status === 'queued') : [];
  const visibleFiles = [...activeFiles, ...queuedFiles].slice(0, 4);
  const hiddenCount = activeUpload ? Math.max(activeUpload.files.length - visibleFiles.length, 0) : 0;

  return (
    <>
      {isUploading && activeUpload && (
        <div
          role="status"
          aria-live="polite"
          className="pointer-events-auto relative w-[21rem] max-w-[calc(100vw-2rem)] overflow-hidden rounded-2xl border border-cyan-300/25 bg-[#050816]/90 p-4 shadow-[0_18px_60px_-20px_rgba(34,211,238,0.65)] backdrop-blur-xl animate-slide-in"
        >
          <div aria-hidden className="pointer-events-none absolute -right-12 -top-12 h-32 w-32 rounded-full bg-cyan-400/20 blur-2xl animate-glow" />
          <div aria-hidden className="pointer-events-none absolute -bottom-14 -left-10 h-32 w-32 rounded-full bg-violet-500/20 blur-2xl animate-glow" />
          <div aria-hidden className="absolute inset-x-0 top-0 h-0.5 bg-white/10">
            <div
              className="h-full bg-gradient-to-r from-cyan-300 via-sky-400 to-violet-400 shadow-[0_0_12px_rgba(34,211,238,0.9)] transition-[width] duration-300"
              style={{ width: `${uploadProgress}%` }}
            />
          </div>

          <div className="relative flex items-start gap-3">
            <div className="relative flex h-9 w-9 shrink-0 items-center justify-center rounded-xl border border-cyan-300/30 bg-cyan-400/10 text-cyan-200">
              <Zap className="h-4 w-4 animate-pulse" />
              <span aria-hidden className="absolute inset-0 rounded-xl bg-cyan-400/20 blur-md" />
            </div>
            <div className="min-w-0 flex-1">
              <div className="flex items-center justify-between gap-2">
                <p className="truncate text-sm font-semibold text-slate-100">{t('home.uploading')}</p>
                <span className="shrink-0 rounded-full border border-white/10 bg-white/5 px-2 py-0.5 text-[11px] tabular-nums text-slate-300">
                  {activeUpload.completedFiles}/{activeUpload.totalFiles}
                </span>
              </div>
              {!collapsed && visibleFiles.length > 0 && (
                <ul className="mt-1.5 max-h-24 space-y-1 overflow-y-auto">
                  {visibleFiles.map((file, index) => {
                    const percent =
                      file.size > 0 ? Math.min(100, Math.round((file.loaded / file.size) * 100)) : 0;
                    return (
                      <li
                        key={`${file.name}-${index}`}
                        className="flex items-center gap-2 text-[11px] tabular-nums text-slate-400"
                      >
                        <span
                          aria-hidden
                          className={`h-1.5 w-1.5 shrink-0 rounded-full ${
                            file.status === 'active' ? 'bg-cyan-300 shadow-[0_0_8px_rgba(34,211,238,0.9)]' : 'bg-slate-600'
                          }`}
                        />
                        <span title={file.name} className="min-w-0 flex-1 truncate">
                          {file.name}
                        </span>
                        <span className="shrink-0 text-slate-500">{percent}%</span>
                      </li>
                    );
                  })}
                  {hiddenCount > 0 && (
                    <li className="text-[11px] tabular-nums text-slate-500">+{hiddenCount}</li>
                  )}
                </ul>
              )}
            </div>
            <div className="flex shrink-0 items-center">
              <button
                type="button"
                onClick={() => setCollapsed((value) => !value)}
                className="rounded-lg p-1 text-slate-400 transition-colors hover:bg-white/10 hover:text-white"
                aria-label={collapsed ? t('home.expandUpload') : t('home.collapseUpload')}
                title={collapsed ? t('home.expandUpload') : t('home.collapseUpload')}
              >
                {collapsed ? <ChevronUp className="h-4 w-4" /> : <ChevronDown className="h-4 w-4" />}
              </button>
              <button
                type="button"
                onClick={() => setConfirmingCancel(true)}
                className="rounded-lg p-1 text-slate-400 transition-colors hover:bg-red-500/15 hover:text-red-300"
                aria-label={t('home.cancelUpload')}
                title={t('home.cancelUpload')}
              >
                <X className="h-4 w-4" />
              </button>
            </div>
          </div>

          {!collapsed && (
            <div className="relative mt-3">
              <div className="mb-1.5 flex items-center justify-between text-xs tabular-nums">
                <span className="font-semibold text-cyan-200">{uploadProgress}%</span>
                <span className="text-slate-400">{formatSpeed(uploadSpeed)}</span>
              </div>
              <div
                role="progressbar"
                aria-valuemin={0}
                aria-valuemax={100}
                aria-valuenow={uploadProgress}
                aria-label={t('home.uploading')}
                className="h-1.5 w-full overflow-hidden rounded-full bg-white/10"
              >
                <div
                  className="relative h-full rounded-full bg-gradient-to-r from-cyan-300 via-sky-400 to-violet-400 shadow-[0_0_14px_rgba(34,211,238,0.85)] transition-[width] duration-300"
                  style={{ width: `${uploadProgress}%` }}
                >
                  <span aria-hidden className="absolute inset-y-0 right-0 w-8 animate-pulse bg-white/30 blur-[3px]" />
                </div>
              </div>
              <p className="mt-1.5 text-[11px] tabular-nums text-slate-500">
                {formatBytes(activeUpload.loadedBytes)} / {formatBytes(activeUpload.totalBytes)}
              </p>
            </div>
          )}
        </div>
      )}

      <Modal
        open={confirmingCancel}
        onClose={() => setConfirmingCancel(false)}
        title={t('home.cancelUploadTitle')}
      >
        <div className="flex items-start gap-3">
          <div className="rounded-full bg-red-500/15 p-2 text-red-300">
            <AlertTriangle className="h-5 w-5" />
          </div>
          <p className="text-sm text-slate-300">{t('home.cancelUploadMessage')}</p>
        </div>
        <div className="mt-6 flex justify-end gap-2">
          <button
            type="button"
            onClick={() => setConfirmingCancel(false)}
            className="rounded-xl border border-white/10 bg-white/5 px-4 py-2 text-sm text-slate-200 transition-colors hover:bg-white/10"
          >
            {t('home.keepUploading')}
          </button>
          <button
            type="button"
            onClick={() => {
              setConfirmingCancel(false);
              cancelUpload();
            }}
            className="rounded-xl bg-red-500/90 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-red-500"
          >
            {t('home.cancelUpload')}
          </button>
        </div>
      </Modal>
    </>
  );
}
