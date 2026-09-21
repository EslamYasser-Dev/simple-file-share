import { useEffect, useState } from 'react';
import { Loader2, Shield, Users } from 'lucide-react';
import { api, type AdminUser } from '../services/api';
import { formatBytes, formatDate } from '../lib/utils';
import { useI18n } from '../i18n';

export function Admin() {
  const { t } = useI18n();
  const [users, setUsers] = useState<AdminUser[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      setIsLoading(true);
      setError(null);
      const result = await api.adminUsers();
      if (cancelled) return;
      if (result.error) {
        setError(result.error);
      } else {
        setUsers(result.data ?? []);
      }
      setIsLoading(false);
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">
          <span className="text-gradient">{t('admin.title')}</span>
        </h1>
        <p className="mt-1 text-sm text-slate-400">{t('admin.subtitle')}</p>
      </div>

      <div className="glass-panel overflow-hidden">
        {isLoading ? (
          <div className="flex flex-col items-center justify-center gap-3 py-20 text-slate-500">
            <Loader2 className="h-8 w-8 animate-spin text-cyan-400" />
            <p className="text-sm">{t('admin.loading')}</p>
          </div>
        ) : error ? (
          <div className="flex flex-col items-center justify-center gap-3 py-20 text-slate-500">
            <p className="text-sm font-medium text-red-300">{t('admin.loadFailed')}</p>
            <p className="text-xs text-slate-600">{error}</p>
          </div>
        ) : users.length === 0 ? (
          <div className="flex flex-col items-center justify-center gap-3 py-20 text-slate-500">
            <Users className="h-8 w-8 text-cyan-300/70" />
            <p className="text-sm">{t('admin.empty')}</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-sm">
              <thead>
                <tr className="border-b border-white/10 text-[11px] font-semibold uppercase tracking-widest text-slate-500">
                  <th className="px-4 py-2.5">{t('admin.username')}</th>
                  <th className="px-4 py-2.5">{t('admin.role')}</th>
                  <th className="px-4 py-2.5">{t('admin.files')}</th>
                  <th className="px-4 py-2.5">{t('admin.size')}</th>
                  <th className="px-4 py-2.5">{t('admin.createdAt')}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-white/5">
                {users.map((user) => (
                  <tr key={user.username} className="transition-colors hover:bg-cyan-400/5">
                    <td className="px-4 py-3 font-medium text-slate-200">{user.username}</td>
                    <td className="px-4 py-3">
                      {user.isAdmin ? (
                        <span className="inline-flex items-center gap-1 rounded-md border border-cyan-400/30 bg-cyan-400/10 px-2 py-0.5 text-[11px] font-semibold text-cyan-300">
                          <Shield className="h-3 w-3" />
                          {t('admin.roleAdmin')}
                        </span>
                      ) : (
                        <span className="text-xs text-slate-400">{t('admin.roleUser')}</span>
                      )}
                    </td>
                    <td className="px-4 py-3 text-slate-400">{user.files}</td>
                    <td className="px-4 py-3 text-slate-400">{formatBytes(user.size)}</td>
                    <td className="px-4 py-3 text-slate-500">
                      {user.createdAt ? formatDate(user.createdAt) : '—'}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}
