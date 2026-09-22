import { useEffect, useState } from 'react';
import { Download, ExternalLink, Loader2, PencilLine, Save, X, XCircle } from 'lucide-react';
import { api, type FileItem } from '../services/api';
import { contentTypeFor, formatBytes, viewerKindFor } from '../lib/utils';
import type { ViewerKind } from '../lib/utils';
import { useI18n } from '../i18n';
import { useToast } from '../hooks/useToast';

interface FilePreviewProps {
  item: FileItem | null;
  onClose: () => void;
  /** Called after a successful inline edit so the caller can refresh listings. */
  onSaved?: () => void;
}

async function fetchViewable(path: string): Promise<Blob> {
  try {
    return await api.viewFile(path);
  } catch {
    // Older servers without the inline endpoint can still stream via
    // the download endpoint; normalize the type so the browser renders it.
    try {
      return await api.downloadFile(path);
    } catch {
      throw new Error('VIEW_FAILED');
    }
  }
}

function withType(blob: Blob, name: string): Blob {
  const type = contentTypeFor(name);
  if (blob.type === type) return blob;
  const copy = blob.slice(0, blob.size);
  return new Blob([copy], { type });
}

const escapeHtml = (s: string) =>
  s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');

function inlineMarkdown(s: string): string {
  return s
    .replace(/`([^`]+)`/g, '<code>$1</code>')
    .replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>')
    .replace(/\*([^*]+)\*/g, '<em>$1</em>')
    .replace(/\[([^\]]+)\]\(([^)\s]+)\)/g, '<a class="text-cyan-300 underline" href="$2" target="_blank" rel="noopener">$1</a>');
}

/** Minimal, XSS-safe Markdown renderer for previewing .md files. */
function renderMarkdown(src: string): string {
  const lines = src.split('\n');
  const out: string[] = [];
  let inFence = false;
  let fenceLines: string[] = [];

  const flushFence = () => {
    out.push(`<pre class="dark-surface overflow-x-auto rounded-lg border border-white/10 bg-black/40 p-3 my-2 text-xs text-slate-100"><code>${escapeHtml(fenceLines.join('\n'))}</code></pre>`);
    fenceLines = [];
  };

  for (const raw of lines) {
    const line = raw.replace(/\r$/, '');

    if (line.trim().startsWith('```')) {
      if (inFence) {
        flushFence();
      }
      inFence = !inFence;
      continue;
    }
    if (inFence) {
      fenceLines.push(line);
      continue;
    }

    const esc = escapeHtml(line);
    const trimmed = line.trim();

    let m: RegExpMatchArray | null;
    if ((m = trimmed.match(/^(#{1,6})\s+(.*)$/))) {
      const level = m[1].length;
      out.push(`<h${level} class="font-semibold text-slate-100 mt-3 mb-1">${inlineMarkdown(escapeHtml(m[2]))}</h${level}>`);
    } else if (/^\s*([-*_])\1{2,}\s*$/.test(trimmed)) {
      out.push('<hr class="my-3 border-white/10" />');
    } else if ((m = line.match(/^>\s?(.*)$/))) {
      out.push(`<blockquote class="my-1 border-s border-cyan-400/50 pl-3 text-slate-400">${inlineMarkdown(escapeHtml(m[1]))}</blockquote>`);
    } else if ((m = line.match(/^\s*[-*+]\s+(.*)$/))) {
      out.push(`<li class="ml-4 list-disc">${inlineMarkdown(escapeHtml(m[1]))}</li>`);
    } else if ((m = line.match(/^\s*\d+\.\s+(.*)$/))) {
      out.push(`<li class="ml-4 list-decimal">${inlineMarkdown(escapeHtml(m[1]))}</li>`);
    } else if (trimmed === '') {
      out.push('<p class="h-2"></p>');
    } else {
      out.push(`<p class="my-1 leading-relaxed">${inlineMarkdown(esc)}</p>`);
    }
  }
  if (inFence && fenceLines.length > 0) flushFence();
  return out.join('\n');
}

export function FilePreview({ item, onClose, onSaved }: FilePreviewProps) {
  const { t } = useI18n();
  const { success, error } = useToast();
  const [objectUrl, setObjectUrl] = useState<string | null>(null);
  const [text, setText] = useState<string | null>(null);
  const [draft, setDraft] = useState<string>('');
  const [isEditing, setIsEditing] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [status, setStatus] = useState<'loading' | 'ready' | 'error'>('loading');
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const kind: ViewerKind = item ? viewerKindFor(item.name) : null;

  useEffect(() => {
    if (!item) return;
    let cancelled = false;
    let url: string | null = null;
    setStatus('loading');
    setIsEditing(false);
    setObjectUrl(null);
    setText(null);

    (async () => {
      try {
        const blob = await fetchViewable(item.path);
        if (cancelled) return;
        if (kind === 'markdown' || kind === 'text') {
          const content = await blob.text();
          setText(content);
          setDraft(content);
        } else {
          url = URL.createObjectURL(withType(blob, item.name));
          setObjectUrl(url);
        }
        setStatus('ready');
      } catch (e) {
        if (cancelled) return;
        setErrorMessage(e instanceof Error ? e.message : t('preview.loadFailed'));
        setStatus('error');
      }
    })();

    const handleKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && !isEditing) onClose();
    };
    window.addEventListener('keydown', handleKey);

    return () => {
      cancelled = true;
      window.removeEventListener('keydown', handleKey);
      if (url) URL.revokeObjectURL(url);
    };
  }, [item, kind, onClose, isEditing, t]);

  if (!item) return null;

  const handleDownload = async () => {
    try {
      const blob = await api.downloadFile(item.path);
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = item.name;
      a.click();
      URL.revokeObjectURL(url);
    } catch {
      error(t('preview.saveFailed'));
    }
  };

  const handleSave = async () => {
    if (!item) return;
    setIsSaving(true);
    const result = await api.updateFileText(item.path, draft);
    setIsSaving(false);
    if (result.error) {
      error(result.error);
      return;
    }
    setText(draft);
    setIsEditing(false);
    success(t('preview.saved'));
    onSaved?.();
  };

  const markdown = renderMarkdown(text ?? '');

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 p-4 backdrop-blur-md animate-slide-in"
      onClick={onClose}
      role="dialog"
      aria-modal="true"
      aria-label={t('preview.ariaLabel', { name: item.name })}
    >
      <div
        className="glass-panel flex h-full max-h-[90vh] w-full max-w-5xl flex-col overflow-hidden"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Toolbar */}
        <div className="flex items-center justify-between gap-3 border-b border-white/10 px-4 py-3">
          <div className="min-w-0">
            <h2 className="truncate text-sm font-semibold text-slate-100">{item.name}</h2>
            <p className="text-xs text-slate-500">{formatBytes(item.size)}</p>
          </div>
          <div className="flex shrink-0 items-center gap-1">
            {kind === 'markdown' && !isEditing && (
              <button
                onClick={() => setIsEditing(true)}
                className="inline-flex items-center gap-1.5 rounded-lg border border-cyan-400/30 bg-cyan-400/10 px-3 py-1.5 text-xs font-semibold text-cyan-200 transition-colors hover:bg-cyan-400/20"
              >
                <PencilLine className="h-3.5 w-3.5" />
                {t('preview.editMarkdown')}
              </button>
            )}
            {kind === 'markdown' && isEditing && (
              <button
                onClick={handleSave}
                disabled={isSaving}
                className="inline-flex items-center gap-1.5 rounded-lg bg-gradient-to-r from-cyan-500 to-violet-600 px-3 py-1.5 text-xs font-semibold text-white transition-all hover:brightness-110 disabled:opacity-50"
              >
                <Save className="h-3.5 w-3.5" />
                {t('common.save')}
              </button>
            )}
            {objectUrl && kind !== 'markdown' && (
              <a
                href={objectUrl}
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center gap-1.5 rounded-lg border border-white/10 bg-white/5 px-3 py-1.5 text-xs font-semibold text-slate-200 transition-colors hover:border-cyan-400/40 hover:text-white"
                title={t('common.openInNewTab')}
              >
                <ExternalLink className="h-3.5 w-3.5" />
              </a>
            )}
            <button
              onClick={handleDownload}
              className="inline-flex items-center gap-1.5 rounded-lg border border-white/10 bg-white/5 px-3 py-1.5 text-xs font-semibold text-slate-200 transition-colors hover:border-cyan-400/40 hover:bg-white/10 hover:text-white"
              title={t('common.download')}
            >
              <Download className="h-3.5 w-3.5" />
              {t('common.download')}
            </button>
            <button
              onClick={onClose}
              className="rounded-lg p-1.5 text-slate-400 transition-colors hover:bg-white/10 hover:text-white"
              title={t('common.close')}
            >
              <X className="h-5 w-5" />
            </button>
          </div>
        </div>

        {/* Body */}
        <div className="flex min-h-0 flex-1 items-center justify-center bg-[#030409] p-4">
          {status === 'loading' && (
            <div className="flex flex-col items-center gap-3 text-slate-500">
              <Loader2 className="h-8 w-8 animate-spin text-cyan-400" />
              <p className="text-sm">{t('preview.loading')}</p>
            </div>
          )}

          {status === 'error' && (
            <div className="glass-panel max-w-md space-y-3 p-6 text-center">
              <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-xl bg-red-500/10">
                <XCircle className="h-6 w-6 text-red-400" />
              </div>
              <p className="text-sm font-medium text-slate-200">
                {kind ? t('preview.couldNotOpen') : t('preview.noViewer')}
              </p>
              {errorMessage && errorMessage !== 'VIEW_FAILED' && (
                <p className="text-xs text-slate-500">{errorMessage}</p>
              )}
              <button
                onClick={handleDownload}
                className="inline-flex items-center gap-2 rounded-xl bg-gradient-to-r from-cyan-500 to-violet-600 px-4 py-2 text-sm font-semibold text-white transition-all hover:brightness-110"
              >
                <Download className="h-4 w-4" />
                {t('preview.downloadInstead')}
              </button>
            </div>
          )}

          {status === 'ready' && kind === 'pdf' && objectUrl && (
            <iframe src={objectUrl} title={item.name} className="h-full w-full rounded-lg border border-white/10 bg-white" />
          )}

          {status === 'ready' && kind === 'image' && objectUrl && (
            <div className="flex max-h-full max-w-full items-center justify-center">
              <img src={objectUrl} alt={item.name} className="max-h-[72vh] max-w-full rounded-lg border border-white/10 object-contain shadow-2xl" />
            </div>
          )}

          {status === 'ready' && kind === 'video' && objectUrl && (
            <video src={objectUrl} controls className="max-h-[72vh] max-w-full rounded-lg border border-white/10" />
          )}

          {status === 'ready' && kind === 'audio' && objectUrl && (
            <div className="w-full max-w-lg rounded-lg border border-white/10 bg-white/5 p-6">
              <audio src={objectUrl} controls className="w-full" />
            </div>
          )}

          {status === 'ready' && kind === 'markdown' && !isEditing && (
            <div className="h-full w-full overflow-auto rounded-lg border border-white/10 bg-[#04050c]/80 p-5 text-sm text-slate-300">
              <div className="prose prose-invert max-w-full" dangerouslySetInnerHTML={{ __html: markdown }} />
            </div>
          )}

          {status === 'ready' && kind === 'markdown' && isEditing && (
            <textarea
              value={draft}
              onChange={(e) => setDraft(e.target.value)}
              spellCheck={false}
              className="dark-surface h-full w-full resize-none rounded-lg border border-cyan-400/30 bg-[#05070f] p-4 font-mono text-xs leading-relaxed text-slate-200 outline-none focus:border-cyan-400/60 focus:ring-1 focus:ring-cyan-400/30"
            />
          )}

          {status === 'ready' && kind === 'text' && text !== null && (
            <pre className="h-full w-full overflow-auto rounded-lg border border-white/10 bg-[#04050c] p-4 text-xs leading-relaxed text-slate-300">
              {text.length > 200_000 ? `${text.slice(0, 200_000)}\n\n${t('preview.truncated')}` : text}
            </pre>
          )}
        </div>
      </div>
    </div>
  );
}