import { useEffect, useState } from 'react';
import { Loader2, Pencil, Shield, Users } from 'lucide-react';
import { api, type AdminUser } from '../services/api';
import { formatBytes, formatDate } from '../lib/utils';
import { useI18n } from '../i18n';
import { useToast } from '../hooks/useToast';
import { Modal } from '../components/Modal';
import { useAuthStore } from '../store/authStore';

export function Admin() {
  const { t } = useI18n();
  const { success, error } = useToast();
  const isAdmin = useAuthStore((s) => s.user?.isAdmin ?? false);
  const [users, setUsers] = useState<AdminUser[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);
  const [editing, setEditing] = useState<AdminUser | null>(null);
  const [quotaInput, setQuotaInput] = useState('');
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      setIsLoading(true);
      setErrorMsg(null);
      const result = await api.adminUsers();
      if (cancelled) return;
      if (result.error) {
        setErrorMsg(result.error);
      } else {
        setUsers(result.data ?? []);
      }
      setIsLoading(false);
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  const openEdit = (user: AdminUser) => {
    setEditing(user);
    setQuotaInput(user.quotaBytes > 0 ? String(user.quotaBytes) : 'unlimited');
  };

  const handleSave = async () => {
    if (!editing) return;
    setSaving(true);
    const result = await api.setQuota(editing.username, quotaInput.trim());
    setSaving(false);
    if (result.error) {
      error(result.error);
      return;
    }
    if (result.data) {
      setUsers((prev) => prev.map((u) => (u.username === result.data!.username ? result.data! : u)));
    }
    setEditing(null);
    success(t('admin.quotaUpdated', { name: editing.username }));
  };

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
        ) : errorMsg ? (
          <div className="flex flex-col items-center justify-center gap-3 py-20 text-slate-500">
            <p className="text-sm font-medium text-red-300">{t('admin.loadFailed')}</p>
            <p className="text-xs text-slate-600">{errorMsg}</p>
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
                  <th className="px-4 py-2.5">{t('admin.quota')}</th>
                  <th className="px-4 py-2.5">{t('admin.createdAt')}</th>
                  <th className="px-4 py-2.5 text-right">{t('admin.actions')}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-white/5">
                {users.map((user) => {
                  const pct = user.quotaBytes > 0 ? Math.min(100, (user.size / user.quotaBytes) * 100) : 0;
                  return (
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
                      <td className="px-4 py-3">
                        {user.quotaBytes > 0 ? (
                          <div className="flex min-w-32 flex-col gap-1">
                            <span className="text-xs tabular-nums text-slate-400">
                              {formatBytes(user.size)} / {formatBytes(user.quotaBytes)}
                              <span className="ml-1 text-cyan-300">{Math.round(pct)}%</span>
                            </span>
                            <div className="h-1 w-full overflow-hidden rounded-full bg-white/10">
                              <div
                                className={`h-full rounded-full ${
                                  pct >= 90 ? 'bg-amber-400' : 'bg-gradient-to-r from-cyan-300 to-violet-400'
                                }`}
                                style={{ width: `${pct}%` }}
                              />
                            </div>
                          </div>
                        ) : (
                          <span className="text-xs text-slate-500">{t('admin.unlimited')}</span>
                        )}
                      </td>
                      <td className="px-4 py-3 text-slate-500">
                        {user.createdAt ? formatDate(user.createdAt) : '—'}
                      </td>
                      <td className="px-4 py-3 text-right">
                        {isAdmin && (
                          <button
                            onClick={() => openEdit(user)}
                            className="inline-flex items-center gap-1 rounded-lg border border-white/10 bg-white/5 px-2 py-1 text-[11px] text-slate-300 transition-colors hover:border-cyan-400/40 hover:text-white"
                            title={t('admin.editQuota')}
                          >
                            <Pencil className="h-3 w-3" />
                            {t('admin.editQuota')}
                          </button>
                        )}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>

      <Modal
        open={editing !== null}
        onClose={() => !saving && setEditing(null)}
        title={t('admin.setQuotaTitle', { name: editing?.username ?? '' })}
        size="sm"
      >
        <label className="mb-1 block text-xs font-semibold uppercase tracking-widest text-slate-500">
          {t('admin.quotaLabel')}
        </label>
        <input
          type="text"
          value={quotaInput}
          onChange={(e) => setQuotaInput(e.target.value)}
          placeholder={t('admin.quotaPlaceholder')}
          disabled={saving}
          className="w-full rounded-xl border border-white/10 bg-white/5 px-3 py-2 text-sm text-slate-200 placeholder-slate-500 outline-none backdrop-blur transition-colors focus:border-cyan-400/50 focus:ring-1 focus:ring-cyan-400/30 disabled:opacity-50"
          onKeyDown={(e) => {
            if (e.key === 'Enter' && !saving) void handleSave();
          }}
        />
        <p className="mt-2 text-xs text-slate-500">{t('admin.quotaPlaceholder')}</p>
        <div className="mt-6 flex justify-end gap-2">
          <button
            type="button"
            onClick={() => setEditing(null)}
            disabled={saving}
            className="rounded-xl border border-white/10 bg-white/5 px-4 py-2 text-sm text-slate-200 transition-colors hover:bg-white/10 disabled:opacity-50"
          >
            {t('common.cancel')}
          </button>
          <button
            type="button"
            onClick={() => void handleSave()}
            disabled={saving}
            className="rounded-xl bg-gradient-to-r from-cyan-500 to-violet-600 px-4 py-2 text-sm font-semibold text-white shadow-lg shadow-cyan-500/30 transition-colors hover:brightness-110 disabled:opacity-50"
          >
            {saving ? t('admin.saving') : t('admin.save')}
          </button>
        </div>
      </Modal>
    </div>
  );
}
