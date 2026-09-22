import { useCallback, useDeferredValue, useEffect, useState } from 'react';
import { Activity, Loader2, Search, SearchX } from 'lucide-react';
import { api } from '../services/api';
import type { FileItem } from '../services/api';
import { FileIcon } from '../components/FileIcon';
import { useI18n } from '../i18n';
import { formatBytes, timeAgo } from '../lib/utils';

export function Chat() {
  const { t } = useI18n();
  const [query, setQuery] = useState('');
  const deferredQuery = useDeferredValue(query);
  const [results, setResults] = useState<FileItem[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [searched, setSearched] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSearch = useCallback(async (q: string) => {
    const trimmed = q.trim();
    if (!trimmed) {
      setResults([]);
      setSearched(false);
      return;
    }
    setIsLoading(true);
    setError(null);
    try {
      const { data, error: err } = await api.searchFiles(trimmed);
      if (err) throw new Error(err);
      setResults(data || []);
      setSearched(true);
    } catch (e) {
      setError(e instanceof Error ? e.message : t('chat.searchFailed'));
      setResults([]);
      setSearched(true);
    } finally {
      setIsLoading(false);
    }
  }, [t]);

  useEffect(() => {
    const timer = setTimeout(() => {
      handleSearch(deferredQuery);
    }, 300);
    return () => clearTimeout(timer);
  }, [deferredQuery, handleSearch]);

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">
          <span className="text-gradient">{t('chat.title')}</span>
        </h1>
        <p className="mt-1 text-sm text-slate-400">{t('chat.subtitle')}</p>
      </div>

      {/* Search input */}
      <div className="relative">
        <Search className="absolute left-3 top-1/2 h-5 w-5 -translate-y-1/2 text-cyan-400/70" />
        <input
          type="text"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder={t('chat.placeholder')}
          autoFocus
          className="w-full rounded-xl border border-white/10 bg-white/5 py-3 pl-11 pr-4 text-sm text-slate-200 placeholder-slate-500 outline-none backdrop-blur transition-all focus:border-cyan-400/50 focus:ring-1 focus:ring-cyan-400/30"
        />
        {isLoading && (
          <Loader2 className="absolute right-3 top-1/2 h-5 w-5 -translate-y-1/2 animate-spin text-cyan-400" />
        )}
      </div>

      {/* Results */}
      {error ? (
        <div className="flex flex-col items-center justify-center gap-3 py-20 text-center">
          <SearchX className="h-10 w-10 text-red-400" />
          <p className="text-sm text-red-400">{error}</p>
        </div>
      ) : searched && results.length === 0 ? (
        <div className="flex flex-col items-center justify-center gap-3 py-20 text-center">
          <SearchX className="h-10 w-10 text-slate-600" />
          <p className="text-sm font-medium text-slate-400">{t('chat.noResults')}</p>
          <p className="text-xs text-slate-600">{t('chat.tryDifferent')}</p>
        </div>
      ) : results.length > 0 ? (
        <div className="glass-panel overflow-hidden">
          <div className="border-b border-white/10 px-4 py-2.5 text-[11px] font-semibold uppercase tracking-widest text-slate-500">
            {t('chat.results', { n: results.length })}
          </div>
          <ul className="divide-y divide-white/5">
            {results.map((item) => (
              <li key={item.path} className="flex items-center gap-3 px-4 py-3 transition-colors hover:bg-cyan-400/5">
                <FileIcon name={item.name} isDir={item.isDir} />
                <div className="min-w-0 flex-1">
                  <p className="truncate text-sm font-medium text-slate-200">{item.name}</p>
                  <p className="truncate text-xs text-cyan-300/60">{item.path}</p>
                </div>
                <div className="hidden text-right sm:block">
                  <p className="text-sm text-slate-400">{item.isDir ? t('chat.folder') : formatBytes(item.size)}</p>
                  <p className="text-xs text-slate-600">{timeAgo(item.modified)}</p>
                </div>
              </li>
            ))}
          </ul>
        </div>
      ) : (
        <div className="flex flex-col items-center justify-center gap-3 py-20 text-center">
          <div className="relative flex h-14 w-14 items-center justify-center rounded-2xl bg-white/5">
            <Activity className="h-7 w-7 text-cyan-300/70" />
            <span className="absolute inset-0 rounded-2xl bg-cyan-400/15 blur-md animate-glow" />
          </div>
          <p className="text-sm font-medium text-slate-400">{t('chat.searchYourFiles')}</p>
          <p className="text-xs text-slate-600">{t('chat.typeQuery')}</p>
        </div>
      )}
    </div>
  );
}