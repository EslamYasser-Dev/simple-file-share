import { useState } from 'react';
import { SlidersHorizontal } from 'lucide-react';
import { useConfigStore } from '../store/configStore';
import { useI18n } from '../i18n';
import { Modal } from './Modal';

interface ConfigPanelProps {
  open: boolean;
  onClose: () => void;
}

/**
 * Configuration panel for user-tunable behaviour: theme preference and the
 * concurrency limit applied to the parallel upload worker.
 */
export function ConfigPanel({ open, onClose }: ConfigPanelProps) {
  const { t } = useI18n();
  const theme = useConfigStore((s) => s.theme);
  const setTheme = useConfigStore((s) => s.setTheme);
  const maxUploads = useConfigStore((s) => s.maxConcurrentUploads);
  const setMaxUploads = useConfigStore((s) => s.setMaxConcurrentUploads);
  const lanShareEnabled = useConfigStore((s) => s.lanShareEnabled);
  const setLanShareEnabled = useConfigStore((s) => s.setLanShareEnabled);
  const [uploadDraft, setUploadDraft] = useState(String(maxUploads));
  const [prevMaxUploads, setPrevMaxUploads] = useState(maxUploads);
  if (prevMaxUploads !== maxUploads) {
    setPrevMaxUploads(maxUploads);
    setUploadDraft(String(maxUploads));
  }

  const clamp = (raw: string): number => {
    const n = Number.parseInt(raw, 10);
    if (!Number.isFinite(n)) return 1;
    return Math.min(Math.max(Math.trunc(n), 1), 8);
  };

  return (
    <Modal open={open} onClose={onClose} title={t('nav.config')}>
      <div className="space-y-5">
        {/* Theme */}
        <section>
          <label className="mb-2 block text-xs font-semibold uppercase tracking-wide text-slate-400">
            {t('nav.theme')}
          </label>
          <div className="grid grid-cols-3 gap-2">
            {(['system', 'dark', 'light'] as const).map((option) => (
              <button
                key={option}
                type="button"
                onClick={() => setTheme(option)}
                className={`rounded-xl border px-3 py-2 text-sm font-medium transition-colors ${
                  theme === option
                    ? 'border-cyan-400/60 bg-cyan-500/15 text-cyan-200'
                    : 'border-white/10 bg-white/5 text-slate-300 hover:border-cyan-400/40 hover:text-white'
                }`}
              >
                {t(`config.theme.${option}` as 'config.theme.system')}
              </button>
            ))}
          </div>
        </section>

        {/* Concurrency limit */}
        <section>
          <label
            htmlFor="cfg-uploads"
            className="mb-2 block text-xs font-semibold uppercase tracking-wide text-slate-400"
          >
            {t('config.maxUploads')}
          </label>
          <input
            id="cfg-uploads"
            type="number"
            min={1}
            max={8}
            value={uploadDraft}
            onChange={(e) => setUploadDraft(e.target.value)}
            onBlur={() => {
              const n = clamp(uploadDraft);
              setUploadDraft(String(n));
              setMaxUploads(n);
            }}
            className="w-full rounded-xl border border-white/10 bg-white/5 px-3 py-2.5 text-sm text-slate-100 outline-none transition-all focus:border-cyan-400/50 focus:ring-1 focus:ring-cyan-400/30"
          />
        </section>

        {/* LAN / P2P share toggle */}
        <section>
          <div className="flex items-start justify-between gap-3">
            <div className="min-w-0">
              <p className="text-xs font-semibold uppercase tracking-wide text-slate-400">
                {t('config.lanShare')}
              </p>
              <p className="mt-1 text-xs text-slate-500">{t('config.lanShareHint')}</p>
            </div>
            <button
              type="button"
              role="switch"
              aria-checked={lanShareEnabled}
              aria-label={t('config.lanShare')}
              onClick={() => setLanShareEnabled(!lanShareEnabled)}
              className={`relative h-6 w-11 shrink-0 rounded-full border transition-colors ${
                lanShareEnabled
                  ? 'border-cyan-400/50 bg-cyan-500/40'
                  : 'border-white/15 bg-white/10'
              }`}
            >
              <span
                className={`absolute top-0.5 rounded-full bg-white shadow transition-all ${
                  lanShareEnabled ? 'left-[calc(100%-1.125rem-0.125rem)]' : 'left-0.5'
                }`}
                style={{ height: '1.125rem', width: '1.125rem' }}
              />
            </button>
          </div>
          <p
            className={`mt-1.5 text-[11px] font-medium ${
              lanShareEnabled ? 'text-cyan-300/80' : 'text-slate-500'
            }`}
          >
            {lanShareEnabled ? t('config.lanShareOn') : t('config.lanShareOff')}
          </p>
        </section>

        <div className="flex items-center gap-2 text-xs text-slate-500">
          <SlidersHorizontal className="h-3.5 w-3.5" />
          <span>{t('config.hint')}</span>
        </div>
      </div>
    </Modal>
  );
}