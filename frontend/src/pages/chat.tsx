import { useCallback, useEffect, useState } from 'react';
import { Activity, Loader2, Search, SearchX } from 'lucide-react';
import { api } from '../services/api';
import { FileIcon } from '../components/FileIcon';
import { formatBytes, timeAgo } from '../lib/utils';

interface FileItem {
  name: string;
  path: string;
  size: number;
  isDir: boolean;
  modified: string;
}

export function Chat() {
  const [query, setQuery] = useState('');
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
      setError(e instanceof Error ? e.message : 'Search failed');
      setResults([]);
      setSearched(true);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    const timer = setTimeout(() => {
      handleSearch(query);
    }, 300);
    return () => clearTimeout(timer);
  }, [query, handleSearch]);

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">Search</h1>
        <p className="text-sm text-slate-400">Find files and folders across your storage</p>
      </div>

      {/* Search input */}
      <div className="relative">
        <Search className="absolute left-3 top-1/2 h-5 w-5 -translate-y-1/2 text-slate-500" />
        <input
          type="text"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="Search by name or path..."
          autoFocus
          className="w-full rounded-xl border border-slate-800 bg-slate-900 py-3 pl-11 pr-4 text-sm text-slate-200 placeholder-slate-500 outline-none transition-colors focus:border-blue-500/50 focus:ring-1 focus:ring-blue-500/30"
        />
        {isLoading && (
          <Loader2 className="absolute right-3 top-1/2 h-5 w-5 -translate-y-1/2 animate-spin text-blue-500" />
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
          <p className="text-sm font-medium text-slate-400">No results found</p>
          <p className="text-xs text-slate-600">Try a different search term</p>
        </div>
      ) : results.length > 0 ? (
        <div className="overflow-hidden rounded-xl border border-slate-800 bg-slate-900/50">
          <div className="border-b border-slate-800 px-4 py-2.5 text-xs font-semibold uppercase tracking-wider text-slate-500">
            {results.length} result{results.length !== 1 ? 's' : ''}
          </div>
          <ul className="divide-y divide-slate-800/50">
            {results.map((item) => (
              <li key={item.path} className="flex items-center gap-3 px-4 py-3 transition-colors hover:bg-slate-800/30">
                <FileIcon name={item.name} isDir={item.isDir} />
                <div className="min-w-0 flex-1">
                  <p className="truncate text-sm font-medium text-slate-200">{item.name}</p>
                  <p className="truncate text-xs text-slate-500">{item.path}</p>
                </div>
                <div className="hidden text-right sm:block">
                  <p className="text-sm text-slate-400">{item.isDir ? 'Folder' : formatBytes(item.size)}</p>
                  <p className="text-xs text-slate-600">{timeAgo(item.modified)}</p>
                </div>
              </li>
            ))}
          </ul>
        </div>
      ) : (
        <div className="flex flex-col items-center justify-center gap-3 py-20 text-center">
          <Activity className="h-10 w-10 text-slate-600" />
          <p className="text-sm font-medium text-slate-400">Search your files</p>
          <p className="text-xs text-slate-600">Type a query above to find files and folders</p>
        </div>
      )}
    </div>
  );
}