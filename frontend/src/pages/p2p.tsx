import { useCallback, useRef, useState } from 'react';
import { ArrowDownLeft, ArrowUpRight, Download, Loader2, Radio, Wifi, WifiOff } from 'lucide-react';
import { sendFileToPeer } from '../services/p2p';
import { useP2PStore, type P2PPeer } from '../store/p2pStore';
import { useI18n } from '../i18n';
import { useToast } from '../hooks/useToast';
import { formatBytes } from '../lib/utils';

export function P2P() {
  const { t } = useI18n();
  const { success, error, info } = useToast();
  const peerId = useP2PStore((s) => s.peerId);
  const connected = useP2PStore((s) => s.connected);
  const peers = useP2PStore((s) => s.peers);
  const transfers = useP2PStore((s) => s.transfers);
  const streamError = useP2PStore((s) => s.streamError);
  const removeTransfer = useP2PStore((s) => s.removeTransfer);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [selectedPeer, setSelectedPeer] = useState<string | null>(null);
  const [sending, setSending] = useState(false);

  const openPicker = useCallback(
    (peer: P2PPeer) => {
      setSelectedPeer(peer.id);
      fileInputRef.current?.click();
    },
    [],
  );

  const handleFiles = useCallback(
    async (list: FileList | null) => {
      if (!list || list.length === 0 || !selectedPeer) return;
      setSending(true);
      try {
        for (const file of Array.from(list)) {
          await sendFileToPeer(selectedPeer, file);
        }
        success(t('p2p.sent'));
      } catch (e) {
        error(e instanceof Error ? e.message : t('p2p.sendFailed'));
      } finally {
        setSending(false);
        if (fileInputRef.current) fileInputRef.current.value = '';
      }
    },
    [selectedPeer, success, error, t],
  );

  const sendTransfers = transfers.filter((x) => x.direction === 'send');
  const recvTransfers = transfers.filter((x) => x.direction === 'receive');

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">
            <span className="text-gradient">{t('p2p.title')}</span>
          </h1>
          <p className="mt-1 text-sm text-slate-400">{t('p2p.subtitle')}</p>
        </div>
        <div className="flex items-center gap-2 rounded-xl border border-white/10 bg-white/5 px-3 py-2 text-xs">
          {connected ? (
            <Wifi className="h-4 w-4 text-emerald-400" />
          ) : (
            <WifiOff className="h-4 w-4 text-slate-500" />
          )}
          <span className="text-slate-300">
            {connected ? t('p2p.online') : t('p2p.offline')}
          </span>
          {peerId && (
            <span className="max-w-[8rem] truncate font-mono text-[10px] text-slate-500" title={peerId}>
              {peerId.slice(0, 8)}
            </span>
          )}
        </div>
      </div>

      {!connected && (
        <div className="glass-panel flex items-center gap-3 border-amber-400/20 px-4 py-3 text-sm text-amber-200/90">
          <Loader2 className="h-4 w-4 animate-spin shrink-0" />
          <span>{streamError === 'reconnecting' ? t('p2p.reconnecting') : t('p2p.connecting')}</span>
        </div>
      )}

      <input
        ref={fileInputRef}
        type="file"
        multiple
        className="hidden"
        onChange={(e) => void handleFiles(e.target.files)}
      />

      {/* Peers */}
      <section className="glass-panel overflow-hidden">
        <div className="flex items-center justify-between border-b border-white/10 px-4 py-2.5">
          <div className="flex items-center gap-2 text-[11px] font-semibold uppercase tracking-widest text-slate-500">
            <Radio className="h-3.5 w-3.5" />
            {t('p2p.peers', { n: peers.length })}
          </div>
        </div>
        {peers.length === 0 ? (
          <div className="flex flex-col items-center justify-center gap-2 px-4 py-12 text-center">
            <Radio className="h-8 w-8 text-slate-600" />
            <p className="text-sm font-medium text-slate-400">{t('p2p.noPeers')}</p>
            <p className="max-w-sm text-xs text-slate-600">{t('p2p.noPeersHint')}</p>
          </div>
        ) : (
          <ul className="divide-y divide-white/5">
            {peers.map((peer) => (
              <li key={peer.id} className="flex flex-wrap items-center gap-3 px-4 py-3">
                <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-gradient-to-br from-cyan-400/30 to-violet-500/30 text-sm font-bold uppercase text-cyan-100">
                  {(peer.user || peer.id).charAt(0)}
                </div>
                <div className="min-w-0 flex-1">
                  <p className="truncate text-sm font-medium text-slate-200">
                    {peer.user || t('p2p.anonymous')}
                  </p>
                  <p className="truncate font-mono text-[10px] text-slate-600">{peer.id.slice(0, 12)}</p>
                </div>
                <button
                  type="button"
                  onClick={() => openPicker(peer)}
                  disabled={sending || !connected}
                  className="inline-flex items-center gap-1.5 rounded-xl bg-gradient-to-r from-cyan-500 to-violet-600 px-3 py-1.5 text-xs font-semibold text-white shadow-lg shadow-cyan-500/25 transition-all hover:brightness-110 disabled:cursor-not-allowed disabled:opacity-50"
                >
                  <ArrowUpRight className="h-3.5 w-3.5" />
                  {t('p2p.send')}
                </button>
              </li>
            ))}
          </ul>
        )}
      </section>

      {/* Transfers */}
      {(sendTransfers.length > 0 || recvTransfers.length > 0) && (
        <section className="glass-panel overflow-hidden">
          <div className="border-b border-white/10 px-4 py-2.5 text-[11px] font-semibold uppercase tracking-widest text-slate-500">
            {t('p2p.transfers')}
          </div>
          <ul className="divide-y divide-white/5">
            {[...recvTransfers, ...sendTransfers].map((tr) => {
              const percent = tr.size > 0 ? Math.min(100, Math.round((tr.loaded / tr.size) * 100)) : 0;
              const statusLabel =
                tr.status === 'connecting'
                  ? t('p2p.statusConnecting')
                  : tr.status === 'error'
                    ? tr.error || t('p2p.sendFailed')
                    : tr.status === 'done'
                      ? t('p2p.statusDone')
                      : `${percent}%`;
              return (
                <li key={tr.id} className="px-4 py-3">
                  <div className="flex items-center gap-3">
                    {tr.direction === 'receive' ? (
                      <ArrowDownLeft className="h-4 w-4 shrink-0 text-emerald-400" />
                    ) : (
                      <ArrowUpRight className="h-4 w-4 shrink-0 text-cyan-400" />
                    )}
                    <div className="min-w-0 flex-1">
                      <div className="flex items-center justify-between gap-2">
                        <p className="truncate text-sm text-slate-200" title={tr.name}>
                          {tr.name}
                        </p>
                        <span className="shrink-0 text-[11px] tabular-nums text-slate-500">
                          {formatBytes(tr.loaded)} / {formatBytes(tr.size)}
                        </span>
                      </div>
                      <div className="mt-1.5 flex items-center gap-2">
                        <div className="h-1 flex-1 overflow-hidden rounded-full bg-white/10">
                          <div
                            className={`h-full transition-[width] duration-300 ${
                              tr.status === 'error'
                                ? 'bg-red-400'
                                : tr.status === 'done'
                                  ? 'bg-emerald-400'
                                  : 'bg-gradient-to-r from-cyan-300 to-violet-400'
                            }`}
                            style={{ width: `${percent}%` }}
                          />
                        </div>
                        <span className="w-20 shrink-0 text-right text-[11px] text-slate-500">
                          {statusLabel}
                        </span>
                      </div>
                      {tr.peerLabel && (
                        <p className="mt-0.5 text-[10px] text-slate-600">
                          {tr.direction === 'receive' ? t('p2p.from') : t('p2p.to')} {tr.peerLabel}
                        </p>
                      )}
                    </div>
                    {tr.direction === 'receive' && tr.status === 'done' && tr.url && (
                      <a
                        href={tr.url}
                        download={tr.name}
                        onClick={() => info(t('p2p.downloading'))}
                        className="shrink-0 rounded-lg border border-white/10 bg-white/5 p-2 text-slate-300 transition-colors hover:border-cyan-400/40 hover:text-white"
                        aria-label={t('common.download')}
                        title={t('common.download')}
                      >
                        <Download className="h-4 w-4" />
                      </a>
                    )}
                    <button
                      type="button"
                      onClick={() => removeTransfer(tr.id)}
                      className="shrink-0 rounded-lg border border-white/10 bg-white/5 px-2 py-1.5 text-[11px] text-slate-400 transition-colors hover:border-red-400/40 hover:text-red-300"
                    >
                      {t('common.close')}
                    </button>
                  </div>
                </li>
              );
            })}
          </ul>
        </section>
      )}
    </div>
  );
}
