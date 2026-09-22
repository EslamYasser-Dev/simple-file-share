import { useCallback, useEffect, useMemo, useState } from 'react';
import {
  Check,
  Copy,
  ExternalLink,
  Link2,
  Link2Off,
  Loader2,
} from 'lucide-react';
import { api, type FileItem, type ShareItem } from '../services/api';
import { Modal } from './Modal';
import { useToast } from '../hooks/useToast';
import { useI18n, type MessageKey } from '../i18n';
import { formatDate } from '../lib/utils';

/** Validity presets mapped to seconds (0 = no expiry). */
const VALIDITY_OPTIONS: { id: string; label: MessageKey; seconds: number }[] = [
  { id: '1h', label: 'share.valid1h', seconds: 3600 },
  { id: '24h', label: 'share.valid24h', seconds: 86400 },
  { id: '7d', label: 'share.valid7d', seconds: 604800 },
  { id: '30d', label: 'share.valid30d', seconds: 2592000 },
  { id: 'never', label: 'share.validNever', seconds: 0 },
];

interface ShareModalProps {
  item: FileItem | null;
  onClose: () => void;
}

/**
 * Creates and manages public share links for a single item. The link opens a
 * public download endpoint (no login), so the recipient can fetch the file
 * until the chosen validity window elapses or the owner revokes it.
 */
export function ShareModal({ item, onClose }: ShareModalProps) {
  const { t } = useI18n();
  const { success, error } = useToast();

  const [links, setLinks] = useState<ShareItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [creating, setCreating] = useState(false);
  const [validity, setValidity] = useState('24h');
  const [pendingRevoke, setPendingRevoke] = useState<string | null>(null);
  const [copiedToken, setCopiedToken] = useState<string | null>(null);

  const itemPath = item?.path ?? '';

  const refresh = useCallback(async () => {
    if (!itemPath) return;
    setLoading(true);
    const result = await api.listShares();
    setLoading(false);
    if (result.error) {
      error(result.error);
      return;
    }
    // Links whose stored (virtual) path matches the item's path belong to it.
    setLinks((result.data ?? []).filter((share) => share.path === itemPath));
  }, [itemPath, error]);

  useEffect(() => {
    if (item) void refresh();
  }, [item, refresh]);

  const selected = useMemo(
    () => VALIDITY_OPTIONS.find((opt) => opt.id === validity) ?? VALIDITY_OPTIONS[1],
    [validity],
  );

  const handleCreate = async () => {
    if (!item || creating) return;
    setCreating(true);
    const result = await api.createShare(item.path, selected.seconds);
    setCreating(false);
    if (result.error) {
      error(t('share.createFailed'));
      return;
    }
    setLinks((prev) => [result.data as ShareItem, ...prev]);
    success(t('share.created'));
  };

  const handleRevoke = async (share: ShareItem) => {
    if (pendingRevoke !== share.token && share.token !== '') {
      setPendingRevoke(share.token);
      return;
    }
    setPendingRevoke(null);
    const result = await api.revokeShare(share.token);
    if (result.error) {
      error(result.error);
      return;
    }
    setLinks((prev) => prev.filter((s) => s.token !== share.token));
    success(t('share.revoked'));
  };

  const handleCopy = async (share: ShareItem) => {
    try {
      await navigator.clipboard.writeText(api.shareUrl(share.token));
      setCopiedToken(share.token);
      window.setTimeout(() => setCopiedToken(null), 1500);
      success(t('share.copied'));
    } catch {
      error('Failed to copy');
    }
  };

  const hint = item?.isDir ? t('share.linkHint') : t('share.hint');

  return (
    <Modal open={item !== null} onClose={onClose} title={t('share.title', { name: item?.name ?? '' })}>
      <div className="space-y-5">
        <p className="text-sm text-slate-400">{hint}</p>

        {/* Validity + create */}
        <div className="space-y-3">
          <label className="block text-xs font-semibold uppercase tracking-widest text-slate-500">
            {t('share.validity')}
          </label>
          <div className="flex flex-wrap gap-2">
            {VALIDITY_OPTIONS.map((opt) => (
              <button
                key={opt.id}
                onClick={() => setValidity(opt.id)}
                className={`rounded-xl border px-3 py-1.5 text-xs font-medium transition-colors ${
                  validity === opt.id
                    ? 'border-cyan-400/60 bg-cyan-400/10 text-cyan-300'
                    : 'border-white/10 bg-white/5 text-slate-300 hover:border-cyan-400/30 hover:text-white'
                }`}
              >
                {t(opt.label)}
              </button>
            ))}
          </div>
          <button
            onClick={() => void handleCreate()}
            disabled={creating}
            className="inline-flex w-full items-center justify-center gap-2 rounded-xl bg-gradient-to-r from-cyan-500 to-violet-600 px-4 py-2 text-sm font-semibold text-white shadow-lg shadow-cyan-500/30 transition-all duration-300 hover:brightness-110 disabled:cursor-not-allowed disabled:opacity-60"
          >
            {creating ? (
              <Loader2 className="h-4 w-4 animate-spin" />
            ) : (
              <Link2 className="h-4 w-4" />
            )}
            {creating ? t('share.creating') : t('share.create')}
          </button>
        </div>

        {/* Active links */}
        <div className="space-y-2">
          <div className="flex items-center justify-between">
            <label className="text-xs font-semibold uppercase tracking-widest text-slate-500">
              {t('share.active')}
            </label>
            {loading && <Loader2 className="h-3.5 w-3.5 animate-spin text-cyan-400" />}
          </div>

          {!loading && links.length === 0 && (
            <p className="rounded-xl border border-dashed border-white/10 px-4 py-3 text-center text-sm text-slate-500">
              {t('share.emptyLinksFor')}
            </p>
          )}

          <ul className="space-y-2">
            {links.map((share) => (
              <li
                key={share.token}
                className="flex flex-col gap-2 rounded-xl border border-white/10 bg-white/5 px-3 py-2.5"
              >
                <div className="flex items-center gap-2">
                  <Link2 className="h-3.5 w-3.5 shrink-0 text-cyan-400/70" />
                  <span className="truncate text-sm text-slate-200">
                    {t('share.link')}
                  </span>
                  <span className="ml-auto shrink-0 text-[11px] text-slate-500">
                    {share.expiresAt
                      ? t('share.expires', { date: formatDate(share.expiresAt) })
                      : t('share.never')}
                  </span>
                </div>
                <div className="flex items-center gap-2">
                  <code className="min-w-0 flex-1 truncate rounded-lg bg-black/40 px-2 py-1 text-[11px] text-slate-300">
                    {api.shareUrl(share.token)}
                  </code>
                  <button
                    onClick={() => void handleCopy(share)}
                    className="shrink-0 rounded-lg p-1.5 text-slate-400 transition-colors hover:bg-white/10 hover:text-cyan-300"
                    title={t('share.copy')}
                    aria-label={t('share.copy')}
                  >
                    {copiedToken === share.token ? (
                      <Check className="h-4 w-4 text-emerald-400" />
                    ) : (
                      <Copy className="h-4 w-4" />
                    )}
                  </button>
                  <a
                    href={api.shareUrl(share.token)}
                    target="_blank"
                    rel="noreferrer"
                    className="shrink-0 rounded-lg p-1.5 text-slate-400 transition-colors hover:bg-white/10 hover:text-cyan-300"
                    title={t('common.openInNewTab')}
                    aria-label={t('common.openInNewTab')}
                  >
                    <ExternalLink className="h-4 w-4" />
                  </a>
                  <button
                    onClick={() => void handleRevoke(share)}
                    className={`shrink-0 rounded-lg p-1.5 transition-colors ${
                      pendingRevoke === share.token
                        ? 'bg-red-500/15 text-red-400'
                        : 'text-slate-400 hover:bg-red-500/10 hover:text-red-400'
                    }`}
                    title={pendingRevoke === share.token ? t('share.confirmRevoke') : t('share.revoke')}
                    aria-label={t('share.revoke')}
                  >
                    {pendingRevoke === share.token ? (
                      <span className="px-0.5 text-[11px] font-semibold">{t('share.confirmRevoke')}</span>
                    ) : (
                      <Link2Off className="h-4 w-4" />
                    )}
                  </button>
                </div>
              </li>
            ))}
          </ul>
        </div>
      </div>
    </Modal>
  );
}