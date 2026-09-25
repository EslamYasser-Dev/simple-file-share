import { useCallback, useEffect, useState } from 'react';
import {
  KeyRound,
  Loader2,
  Pencil,
  Plus,
  RefreshCw,
  Shield,
  Trash2,
  UserMinus,
  UserPlus,
  Users,
} from 'lucide-react';
import {
  api,
  type AdminUser,
  type AnalyticsBucket,
  type AnalyticsOverview,
  type AnalyticsTopFile,
  type RoleItem,
} from '../services/api';
import { formatBytes, formatDate } from '../lib/utils';
import { useI18n } from '../i18n';
import { useToast } from '../hooks/useToast';
import { Modal } from '../components/Modal';
import { useAuthStore } from '../store/authStore';

const ALL_PERMISSIONS = [
  'users.read',
  'users.create',
  'users.update',
  'users.delete',
  'quota.manage',
  'roles.manage',
  'shares.manage',
  'files.system',
] as const;

type Tab = 'users' | 'roles' | 'analytics' | 'password';

export function Admin() {
  const { t } = useI18n();
  const { success, error } = useToast();
  const me = useAuthStore((s) => s.user);
  const isAdmin = me?.isAdmin ?? false;
  const [tab, setTab] = useState<Tab>('users');
  const [users, setUsers] = useState<AdminUser[]>([]);
  const [roles, setRoles] = useState<RoleItem[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);
  const [editing, setEditing] = useState<AdminUser | null>(null);
  const [quotaInput, setQuotaInput] = useState('');
  const [showCreate, setShowCreate] = useState(false);
  const [showRoleEditor, setShowRoleEditor] = useState<RoleItem | 'new' | null>(null);
  const [showPassword, setShowPassword] = useState(false);
  const [saving, setSaving] = useState(false);
  const [confirmDelete, setConfirmDelete] = useState<string | null>(null);
  const [analyticsDays, setAnalyticsDays] = useState(30);
  const [overview, setOverview] = useState<AnalyticsOverview | null>(null);
  const [timeline, setTimeline] = useState<AnalyticsBucket[]>([]);
  const [topFiles, setTopFiles] = useState<AnalyticsTopFile[]>([]);
  const [analyticsLoading, setAnalyticsLoading] = useState(false);

  const loadUsers = useCallback(async () => {
    const result = await api.adminUsers();
    if (result.error) setErrorMsg(result.error);
    else setUsers(result.data ?? []);
  }, []);

  const loadRoles = useCallback(async () => {
    const result = await api.listRoles();
    if (!result.error) setRoles(result.data ?? []);
  }, []);

  const loadAnalytics = useCallback(async (days: number) => {
    setAnalyticsLoading(true);
    const [ov, tl, tf] = await Promise.all([
      api.analyticsOverview(days),
      api.analyticsTimeline(days),
      api.analyticsTopFiles(days, 10),
    ]);
    setOverview(ov.data ?? null);
    setTimeline(tl.data ?? []);
    setTopFiles(tf.data ?? []);
    setAnalyticsLoading(false);
  }, []);

  useEffect(() => {
    if (tab !== 'analytics') return;
    void (async () => {
      await loadAnalytics(analyticsDays);
    })();
  }, [tab, analyticsDays, loadAnalytics]);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      setIsLoading(true);
      setErrorMsg(null);
      await Promise.all([loadUsers(), loadRoles()]);
      if (cancelled) return;
      setIsLoading(false);
    })();
    return () => {
      cancelled = true;
    };
  }, [loadUsers, loadRoles]);

  const openEdit = (user: AdminUser) => {
    setEditing(user);
    // Seed with a human-readable size (or "unlimited") so the field matches
    // what admins type; raw byte counts are hard to read in the modal.
    setQuotaInput(user.quotaBytes > 0 ? formatBytes(user.quotaBytes) : 'unlimited');
  };

  const handleSaveQuota = async () => {
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

  const handleToggleEnabled = async (user: AdminUser) => {
    const next = !user.enabled;
    const result = await api.updateUser(user.username, { enabled: next });
    if (result.error) {
      error(result.error);
      return;
    }
    await loadUsers();
    success(next ? t('admin.userEnabled', { name: user.username }) : t('admin.userDisabled', { name: user.username }));
  };

  const handleDeleteUser = async (username: string) => {
    const result = await api.deleteUser(username);
    setConfirmDelete(null);
    if (result.error) {
      error(result.error);
      return;
    }
    await loadUsers();
    success(t('admin.userDeleted', { name: username }));
  };

  const handleChangeRole = async (user: AdminUser, role: string) => {
    if (role === user.role) return;
    const result = await api.updateUser(user.username, { role });
    if (result.error) {
      error(result.error);
      return;
    }
    await loadUsers();
    success(t('admin.roleChanged', { name: user.username }));
  };

  const handleResetPassword = async (username: string, password: string) => {
    setSaving(true);
    const result = await api.resetPassword(username, password);
    setSaving(false);
    if (result.error) {
      error(result.error);
      return;
    }
    success(t('admin.passwordReset', { name: username }));
    return true;
  };

  const tabBtn = (key: Tab, label: string) => (
    <button
      key={key}
      type="button"
      onClick={() => setTab(key)}
      className={`rounded-lg px-3 py-1.5 text-sm font-medium transition-colors ${
        tab === key ? 'bg-cyan-400/15 text-cyan-300' : 'text-slate-400 hover:bg-white/5 hover:text-white'
      }`}
    >
      {label}
    </button>
  );

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">
            <span className="text-gradient">{t('admin.title')}</span>
          </h1>
          <p className="mt-1 text-sm text-slate-400">{t('admin.subtitle')}</p>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <div className="flex rounded-xl border border-white/10 bg-white/5 p-1">
            {tabBtn('users', t('admin.tabUsers'))}
            {tabBtn('roles', t('admin.tabRoles'))}
            {tabBtn('analytics', t('admin.tabAnalytics'))}
            {tabBtn('password', t('admin.tabPassword'))}
          </div>
          <button
            type="button"
            onClick={() => void (tab === 'roles' ? loadRoles() : tab === 'analytics' ? loadAnalytics(analyticsDays) : loadUsers())}
            className="rounded-xl border border-white/10 bg-white/5 p-2 text-slate-300 transition-colors hover:border-cyan-400/40 hover:text-white"
            title={t('common.retry')}
          >
            <RefreshCw className="h-4 w-4" />
          </button>
        </div>
      </div>

      {tab === 'users' && (
        <div className="glass-panel overflow-hidden">
          <div className="flex items-center justify-between border-b border-white/10 px-4 py-3">
            <p className="text-xs font-semibold uppercase tracking-widest text-slate-500">{t('admin.tabUsers')}</p>
            {isAdmin && (
              <button
                type="button"
                onClick={() => setShowCreate(true)}
                className="inline-flex items-center gap-1 rounded-lg bg-gradient-to-r from-cyan-500 to-violet-600 px-3 py-1.5 text-xs font-semibold text-white shadow-lg shadow-cyan-500/20 transition hover:brightness-110"
              >
                <UserPlus className="h-3.5 w-3.5" />
                {t('admin.createUser')}
              </button>
            )}
          </div>
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
              {/* Desktop: full table. Mobile: stacked cards so actions stay reachable. */}
              <table className="hidden w-full text-start text-sm md:table">
                <thead>
                  <tr className="border-b border-white/10 text-[11px] font-semibold uppercase tracking-widest text-slate-500">
                    <th className="px-4 py-2.5">{t('admin.username')}</th>
                    <th className="px-4 py-2.5">{t('admin.role')}</th>
                    <th className="px-4 py-2.5">{t('admin.status')}</th>
                    <th className="px-4 py-2.5">{t('admin.files')}</th>
                    <th className="px-4 py-2.5">{t('admin.size')}</th>
                    <th className="px-4 py-2.5">{t('admin.quota')}</th>
                    <th className="px-4 py-2.5">{t('admin.createdAt')}</th>
                    <th className="px-4 py-2.5 text-end">{t('admin.actions')}</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-white/5">
                  {users.map((user) => {
                    const pct = user.quotaBytes > 0 ? Math.min(100, (user.size / user.quotaBytes) * 100) : 0;
                    return (
                      <tr key={user.username} className="transition-colors hover:bg-cyan-400/5">
                        <td className="px-4 py-3 font-medium text-slate-200">{user.username}</td>
                        <td className="px-4 py-3">
                          {isAdmin ? (
                            <select
                              value={user.role || (user.isAdmin ? 'admin' : 'member')}
                              onChange={(e) => void handleChangeRole(user, e.target.value)}
                              className="rounded-lg border border-white/10 bg-white/5 px-2 py-1 text-xs text-slate-200 outline-none focus:border-cyan-400/50"
                            >
                              {roles.map((r) => (
                                <option key={r.name} value={r.name} className="bg-slate-900">
                                  {r.name}
                                </option>
                              ))}
                              {!roles.some((r) => r.name === (user.role || 'member')) && (
                                <option value={user.role || 'member'} className="bg-slate-900">
                                  {user.role || 'member'}
                                </option>
                              )}
                            </select>
                          ) : user.isAdmin ? (
                            <span className="inline-flex items-center gap-1 rounded-md border border-cyan-400/30 bg-cyan-400/10 px-2 py-0.5 text-[11px] font-semibold text-cyan-300">
                              <Shield className="h-3 w-3" />
                              {t('admin.roleAdmin')}
                            </span>
                          ) : (
                            <span className="text-xs text-slate-400">{user.role || t('admin.roleUser')}</span>
                          )}
                        </td>
                        <td className="px-4 py-3">
                          <span
                            className={`inline-flex rounded-md px-2 py-0.5 text-[11px] font-semibold ${
                              user.enabled !== false
                                ? 'border border-emerald-400/30 bg-emerald-400/10 text-emerald-300'
                                : 'border border-amber-400/30 bg-amber-400/10 text-amber-300'
                            }`}
                          >
                            {user.enabled !== false ? t('admin.enabled') : t('admin.disabled')}
                          </span>
                        </td>
                        <td className="px-4 py-3 text-slate-400">{user.files}</td>
                        <td className="px-4 py-3 text-slate-400">
                          <bdi>{formatBytes(user.size)}</bdi>
                        </td>
                        <td className="px-4 py-3">
                          {user.quotaBytes > 0 ? (
                            <div className="flex min-w-32 flex-col gap-1">
                              <span className="flex flex-wrap items-center gap-x-1 text-xs tabular-nums text-slate-400">
                                <bdi>{formatBytes(user.size)}</bdi>
                                <span aria-hidden>/</span>
                                <bdi>{formatBytes(user.quotaBytes)}</bdi>
                                <span className="ms-1 text-cyan-300" dir="ltr">
                                  {Math.round(pct)}%
                                </span>
                              </span>
                              <div className="h-1 w-full overflow-hidden rounded-full bg-white/10">
                                <div
                                  className={`h-full rounded-full transition-[width] duration-500 ease-out ${
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
                        <td className="px-4 py-3 text-end">
                          {isAdmin && (
                            <div className="inline-flex flex-wrap justify-end gap-1">
                              <button
                                onClick={() => openEdit(user)}
                                className="inline-flex items-center gap-1 rounded-lg border border-white/10 bg-white/5 px-2 py-1 text-[11px] text-slate-300 transition-all hover:scale-105 hover:border-cyan-400/40 hover:text-white"
                                title={t('admin.editQuota')}
                              >
                                <Pencil className="h-3 w-3" />
                                {t('admin.editQuota')}
                              </button>
                              <button
                                onClick={() => setConfirmDelete(user.username)}
                                className="inline-flex items-center gap-1 rounded-lg border border-red-400/20 bg-red-400/10 px-2 py-1 text-[11px] text-red-300 transition-all hover:scale-105 hover:bg-red-400/20"
                                title={t('common.delete')}
                              >
                                <Trash2 className="h-3 w-3" />
                              </button>
                              <button
                                onClick={() => void handleToggleEnabled(user)}
                                className="inline-flex items-center gap-1 rounded-lg border border-white/10 bg-white/5 px-2 py-1 text-[11px] text-slate-300 transition-all hover:scale-105 hover:border-amber-400/40 hover:text-white"
                                title={user.enabled !== false ? t('admin.disable') : t('admin.enable')}
                              >
                                <UserMinus className="h-3 w-3" />
                              </button>
                              <button
                                onClick={() => {
                                  setEditing({ ...user, __reset: true } as AdminUser);
                                  setQuotaInput('');
                                  setShowPassword(true);
                                }}
                                className="inline-flex items-center gap-1 rounded-lg border border-white/10 bg-white/5 px-2 py-1 text-[11px] text-slate-300 transition-all hover:scale-105 hover:border-violet-400/40 hover:text-white"
                                title={t('admin.resetPassword')}
                              >
                                <KeyRound className="h-3 w-3" />
                              </button>
                            </div>
                          )}
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>

              {/* Mobile card list */}
              <ul className="divide-y divide-white/5 md:hidden">
                {users.map((user) => {
                  const pct = user.quotaBytes > 0 ? Math.min(100, (user.size / user.quotaBytes) * 100) : 0;
                  return (
                    <li key={user.username} className="space-y-2 px-4 py-3">
                      <div className="flex flex-wrap items-center justify-between gap-2">
                        <span className="font-medium text-slate-200">{user.username}</span>
                        <span
                          className={`inline-flex rounded-md px-2 py-0.5 text-[11px] font-semibold ${
                            user.enabled !== false
                              ? 'border border-emerald-400/30 bg-emerald-400/10 text-emerald-300'
                              : 'border border-amber-400/30 bg-amber-400/10 text-amber-300'
                          }`}
                        >
                          {user.enabled !== false ? t('admin.enabled') : t('admin.disabled')}
                        </span>
                      </div>
                      <div className="flex flex-wrap items-center gap-2 text-xs text-slate-400">
                        <span>{t('admin.role')}:</span>
                        {isAdmin ? (
                          <select
                            value={user.role || (user.isAdmin ? 'admin' : 'member')}
                            onChange={(e) => void handleChangeRole(user, e.target.value)}
                            className="rounded-lg border border-white/10 bg-white/5 px-2 py-1 text-xs text-slate-200 outline-none focus:border-cyan-400/50"
                          >
                            {roles.map((r) => (
                              <option key={r.name} value={r.name} className="bg-slate-900">
                                {r.name}
                              </option>
                            ))}
                            {!roles.some((r) => r.name === (user.role || 'member')) && (
                              <option value={user.role || 'member'} className="bg-slate-900">
                                {user.role || 'member'}
                              </option>
                            )}
                          </select>
                        ) : user.isAdmin ? (
                          <span className="inline-flex items-center gap-1 rounded-md border border-cyan-400/30 bg-cyan-400/10 px-2 py-0.5 text-[11px] font-semibold text-cyan-300">
                            <Shield className="h-3 w-3" />
                            {t('admin.roleAdmin')}
                          </span>
                        ) : (
                          <span>{user.role || t('admin.roleUser')}</span>
                        )}
                        <span className="ms-auto">
                          <bdi>{formatBytes(user.size)}</bdi>
                          {user.quotaBytes > 0 ? (
                            <>
                              {' / '}
                              <bdi>{formatBytes(user.quotaBytes)}</bdi>
                              <span className="ms-1 text-cyan-300" dir="ltr">
                                {Math.round(pct)}%
                              </span>
                            </>
                          ) : (
                            <span className="ms-1 text-slate-500">{t('admin.unlimited')}</span>
                          )}
                        </span>
                      </div>
                      {user.quotaBytes > 0 && (
                        <div className="h-1 w-full overflow-hidden rounded-full bg-white/10">
                          <div
                            className={`h-full rounded-full transition-[width] duration-500 ease-out ${
                              pct >= 90 ? 'bg-amber-400' : 'bg-gradient-to-r from-cyan-300 to-violet-400'
                            }`}
                            style={{ width: `${pct}%` }}
                          />
                        </div>
                      )}
                      {isAdmin && (
                        <div className="flex flex-wrap gap-1 pt-1">
                          <button
                            onClick={() => openEdit(user)}
                            className="inline-flex items-center gap-1 rounded-lg border border-white/10 bg-white/5 px-2.5 py-1.5 text-[11px] text-slate-300 transition hover:border-cyan-400/40 hover:text-white"
                          >
                            <Pencil className="h-3 w-3" />
                            {t('admin.editQuota')}
                          </button>
                          <button
                            onClick={() => setConfirmDelete(user.username)}
                            className="inline-flex items-center gap-1 rounded-lg border border-red-400/20 bg-red-400/10 px-2.5 py-1.5 text-[11px] text-red-300 transition hover:bg-red-400/20"
                          >
                            <Trash2 className="h-3 w-3" />
                            {t('common.delete')}
                          </button>
                          <button
                            onClick={() => void handleToggleEnabled(user)}
                            className="inline-flex items-center gap-1 rounded-lg border border-white/10 bg-white/5 px-2.5 py-1.5 text-[11px] text-slate-300 transition hover:border-amber-400/40 hover:text-white"
                          >
                            <UserMinus className="h-3 w-3" />
                            {user.enabled !== false ? t('admin.disable') : t('admin.enable')}
                          </button>
                          <button
                            onClick={() => {
                              setEditing({ ...user, __reset: true } as AdminUser);
                              setQuotaInput('');
                              setShowPassword(true);
                            }}
                            className="inline-flex items-center gap-1 rounded-lg border border-white/10 bg-white/5 px-2.5 py-1.5 text-[11px] text-slate-300 transition hover:border-violet-400/40 hover:text-white"
                          >
                            <KeyRound className="h-3 w-3" />
                            {t('admin.resetPassword')}
                          </button>
                        </div>
                      )}
                    </li>
                  );
                })}
              </ul>
            </div>
          )}
        </div>
      )}

      {tab === 'roles' && (
        <div className="glass-panel overflow-hidden">
          <div className="flex items-center justify-between border-b border-white/10 px-4 py-3">
            <p className="text-xs font-semibold uppercase tracking-widest text-slate-500">{t('admin.tabRoles')}</p>
            <button
              type="button"
              onClick={() => setShowRoleEditor('new')}
              className="inline-flex items-center gap-1 rounded-lg bg-gradient-to-r from-cyan-500 to-violet-600 px-3 py-1.5 text-xs font-semibold text-white shadow-lg shadow-cyan-500/20 transition hover:brightness-110"
            >
              <Plus className="h-3.5 w-3.5" />
              {t('admin.newRole')}
            </button>
          </div>
          <div className="divide-y divide-white/5">
            {roles.map((role) => (
              <div key={role.name} className="flex flex-wrap items-center justify-between gap-3 px-4 py-3">
                <div className="min-w-0">
                  <div className="flex items-center gap-2">
                    <span className="font-medium text-slate-200">{role.name}</span>
                    {role.builtIn && (
                      <span className="rounded-md border border-cyan-400/30 bg-cyan-400/10 px-2 py-0.5 text-[10px] font-semibold text-cyan-300">
                        {t('admin.builtIn')}
                      </span>
                    )}
                  </div>
                  {role.description && <p className="mt-0.5 text-xs text-slate-500">{role.description}</p>}
                  <div className="mt-1.5 flex flex-wrap gap-1">
                    {(role.permissions ?? []).length === 0 ? (
                      <span className="text-[11px] text-slate-600">{t('admin.noPermissions')}</span>
                    ) : (
                      role.permissions.map((p) => (
                        <span
                          key={p}
                          className="rounded bg-white/5 px-1.5 py-0.5 font-mono text-[10px] text-slate-400"
                        >
                          {p}
                        </span>
                      ))
                    )}
                  </div>
                </div>
                {!role.builtIn && (
                  <div className="flex gap-2">
                    <button
                      type="button"
                      onClick={() => setShowRoleEditor(role)}
                      className="rounded-lg border border-white/10 bg-white/5 px-3 py-1.5 text-xs text-slate-300 transition hover:border-cyan-400/40 hover:text-white"
                    >
                      {t('common.edit')}
                    </button>
                    <button
                      type="button"
                      onClick={async () => {
                        const result = await api.deleteRole(role.name);
                        if (result.error) error(result.error);
                        else {
                          await loadRoles();
                          success(t('admin.roleDeleted', { name: role.name }));
                        }
                      }}
                      className="rounded-lg border border-red-400/20 bg-red-400/10 px-3 py-1.5 text-xs text-red-300 transition hover:bg-red-400/20"
                    >
                      {t('common.delete')}
                    </button>
                  </div>
                )}
              </div>
            ))}
          </div>
        </div>
      )}

      {tab === 'analytics' && (
        <div className="space-y-6">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <p className="text-sm text-slate-400">{t('admin.analyticsSubtitle')}</p>
            <div className="flex items-center gap-2">
              <label className="text-xs font-semibold uppercase tracking-widest text-slate-500">
                {t('admin.analyticsDays')}
              </label>
              <select
                value={analyticsDays}
                onChange={(e) => setAnalyticsDays(Number(e.target.value))}
                className="rounded-xl border border-white/10 bg-white/5 px-3 py-1.5 text-sm text-slate-200 outline-none focus:border-cyan-400/50"
              >
                <option value={7}>{t('admin.analyticsDays7')}</option>
                <option value={30}>{t('admin.analyticsDays30')}</option>
                <option value={90}>{t('admin.analyticsDays90')}</option>
              </select>
            </div>
          </div>

          {analyticsLoading && !overview ? (
            <div className="glass-panel flex items-center justify-center py-16">
              <Loader2 className="h-5 w-5 animate-spin text-cyan-400" />
            </div>
          ) : (
            <>
              <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-6">
                {(
                  [
                    ['admin.analyticsUploads', overview?.uploads ?? 0],
                    ['admin.analyticsDownloads', overview?.downloads ?? 0],
                    ['admin.analyticsDeletes', overview?.deletes ?? 0],
                    ['admin.analyticsShares', overview?.shares ?? 0],
                    ['admin.analyticsLogins', overview?.logins ?? 0],
                    ['admin.analyticsActiveUsers', overview?.activeUsers ?? 0],
                  ] as const
                ).map(([label, value]) => (
                  <div key={label} className="glass-panel p-4">
                    <p className="text-[11px] font-semibold uppercase tracking-widest text-slate-500">
                      {t(label)}
                    </p>
                    <p className="mt-1 text-2xl font-bold text-slate-100">{value}</p>
                  </div>
                ))}
              </div>

              <div className="glass-panel p-4">
                <div className="mb-3 flex items-center justify-between">
                  <p className="text-xs font-semibold uppercase tracking-widest text-slate-500">
                    {t('admin.analyticsTimeline')}
                  </p>
                  <p className="text-xs text-slate-500">
                    {t('admin.analyticsBytes')}: {formatBytes(overview?.bytesUploaded ?? 0)}
                  </p>
                </div>
                {timeline.length === 0 ? (
                  <p className="py-8 text-center text-sm text-slate-500">{t('admin.analyticsEmpty')}</p>
                ) : (
                  <div className="flex h-36 items-end gap-1">
                    {timeline.map((bucket) => {
                      const max = Math.max(...timeline.map((b) => b.count), 1);
                      const height = Math.max(4, Math.round((bucket.count / max) * 100));
                      return (
                        <div
                          key={bucket.start}
                          className="min-w-0 flex-1 rounded-t bg-gradient-to-t from-violet-600/60 to-cyan-400/80 transition-all hover:from-violet-500 hover:to-cyan-300"
                          style={{ height: `${height}%` }}
                          title={`${formatDate(bucket.start)} — ${bucket.count} ${t('admin.analyticsHits')}`}
                        />
                      );
                    })}
                  </div>
                )}
              </div>

              <div className="glass-panel overflow-hidden">
                <div className="border-b border-white/10 px-4 py-3">
                  <p className="text-xs font-semibold uppercase tracking-widest text-slate-500">
                    {t('admin.analyticsTopFiles')}
                  </p>
                </div>
                {topFiles.length === 0 ? (
                  <p className="py-8 text-center text-sm text-slate-500">{t('admin.analyticsEmpty')}</p>
                ) : (
                  <div className="divide-y divide-white/5">
                    {topFiles.map((file) => (
                      <div key={file.path} className="flex items-center justify-between gap-3 px-4 py-2.5">
                        <span className="min-w-0 truncate font-mono text-sm text-slate-300">{file.path}</span>
                        <span className="shrink-0 text-xs text-slate-500">
                          {file.count} {t('admin.analyticsHits')}
                          {file.bytes > 0 ? ` · ${formatBytes(file.bytes)}` : ''}
                        </span>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </>
          )}
        </div>
      )}

      {tab === 'password' && (
        <div className="glass-panel max-w-md p-6">
          <h2 className="mb-4 text-lg font-semibold text-slate-100">{t('admin.changeMyPassword')}</h2>
          <ChangeOwnPasswordForm onDone={() => success(t('admin.passwordChanged'))} onError={error} />
        </div>
      )}

      <Modal
        open={editing !== null && !('__reset' in (editing ?? {}))}
        onClose={() => !saving && setEditing(null)}
        title={t('admin.setQuotaTitle', { name: editing?.username ?? '' })}
        size="sm"
      >
        <div className="space-y-4 animate-rise">
          {/* Live usage preview so admins see current fill before changing the cap */}
          {editing && editing.quotaBytes > 0 ? (
            <div className="rounded-xl border border-white/10 bg-white/5 p-3">
              <div className="flex items-center justify-between text-xs text-slate-400">
                <span>{t('admin.size')}</span>
                <span className="tabular-nums" dir="ltr">
                  <bdi>{formatBytes(editing.size)}</bdi> / <bdi>{formatBytes(editing.quotaBytes)}</bdi>
                </span>
              </div>
              <div className="mt-2 h-1.5 w-full overflow-hidden rounded-full bg-white/10">
                <div
                  className={`h-full rounded-full transition-[width] duration-500 ${
                    editing.quotaBytes > 0 && editing.size / editing.quotaBytes >= 0.9
                      ? 'bg-amber-400'
                      : 'bg-gradient-to-r from-cyan-300 to-violet-400'
                  }`}
                  style={{
                    width: `${Math.min(100, (editing.size / Math.max(editing.quotaBytes, 1)) * 100)}%`,
                  }}
                />
              </div>
            </div>
          ) : null}

          <div>
            <label
              className="mb-1.5 block text-xs font-semibold text-slate-500"
              dir="auto"
            >
              {t('admin.quotaLabel')}
            </label>
            <input
              type="text"
              dir="ltr"
              value={quotaInput}
              onChange={(e) => setQuotaInput(e.target.value)}
              placeholder={t('admin.quotaPlaceholder')}
              disabled={saving}
              className="w-full rounded-xl border border-white/10 bg-white/5 px-3 py-2.5 text-sm text-slate-200 placeholder-slate-500 outline-none backdrop-blur transition-all focus:border-cyan-400/50 focus:ring-2 focus:ring-cyan-400/20 disabled:opacity-50"
              onKeyDown={(e) => {
                if (e.key === 'Enter' && !saving) void handleSaveQuota();
              }}
            />
            <p className="mt-2 text-xs text-slate-500">{t('admin.quotaPlaceholder')}</p>
          </div>

          {/* Quick presets */}
          <div className="flex flex-wrap gap-2">
            {['100MB', '500MB', '1GB', '5GB', 'unlimited'].map((preset) => (
              <button
                key={preset}
                type="button"
                disabled={saving}
                onClick={() => setQuotaInput(preset)}
                className={`rounded-lg border px-2.5 py-1 text-[11px] font-medium transition-all hover:scale-105 disabled:opacity-50 ${
                  quotaInput.toLowerCase() === preset.toLowerCase()
                    ? 'border-cyan-400/50 bg-cyan-400/15 text-cyan-300'
                    : 'border-white/10 bg-white/5 text-slate-400 hover:border-cyan-400/40 hover:text-slate-200'
                }`}
              >
                <bdi>{preset}</bdi>
              </button>
            ))}
          </div>
        </div>

        <div className="mt-6 flex justify-end gap-2">
          <button
            type="button"
            onClick={() => setEditing(null)}
            disabled={saving}
            className="rounded-xl border border-white/10 bg-white/5 px-4 py-2 text-sm text-slate-200 transition-all hover:scale-105 hover:bg-white/10 disabled:opacity-50"
          >
            {t('common.cancel')}
          </button>
          <button
            type="button"
            onClick={() => void handleSaveQuota()}
            disabled={saving}
            className="inline-flex items-center gap-2 rounded-xl bg-gradient-to-r from-cyan-500 to-violet-600 px-4 py-2 text-sm font-semibold text-white shadow-lg shadow-cyan-500/30 transition-all hover:scale-105 hover:brightness-110 disabled:opacity-50"
          >
            {saving && <Loader2 className="h-4 w-4 animate-spin" />}
            {saving ? t('admin.saving') : t('admin.save')}
          </button>
        </div>
      </Modal>

      <CreateUserModal
        open={showCreate}
        roles={roles}
        onClose={() => setShowCreate(false)}
        onCreated={async () => {
          setShowCreate(false);
          await loadUsers();
          success(t('admin.userCreated'));
        }}
        onError={error}
      />

      <RoleEditorModal
        open={showRoleEditor !== null}
        role={showRoleEditor === 'new' ? null : showRoleEditor}
        onClose={() => setShowRoleEditor(null)}
        onSaved={async () => {
          setShowRoleEditor(null);
          await loadRoles();
          success(t('admin.roleSaved'));
        }}
        onError={error}
      />

      <ResetPasswordModal
        open={showPassword}
        username={editing?.username ?? ''}
        onClose={() => {
          setShowPassword(false);
          setEditing(null);
        }}
        onReset={handleResetPassword}
        saving={saving}
      />

      <Modal
        open={confirmDelete !== null}
        onClose={() => setConfirmDelete(null)}
        title={t('admin.confirmDeleteTitle', { name: confirmDelete ?? '' })}
        size="sm"
      >
        <p className="text-sm text-slate-400">{t('admin.confirmDeleteBody')}</p>
        <div className="mt-6 flex justify-end gap-2">
          <button
            type="button"
            onClick={() => setConfirmDelete(null)}
            className="rounded-xl border border-white/10 bg-white/5 px-4 py-2 text-sm text-slate-200 hover:bg-white/10"
          >
            {t('common.cancel')}
          </button>
          <button
            type="button"
            onClick={() => confirmDelete && void handleDeleteUser(confirmDelete)}
            className="rounded-xl bg-gradient-to-r from-red-500 to-rose-600 px-4 py-2 text-sm font-semibold text-white shadow-lg shadow-red-500/30 hover:brightness-110"
          >
            {t('common.delete')}
          </button>
        </div>
      </Modal>
    </div>
  );
}

const inputClass =
  'w-full rounded-xl border border-white/10 bg-white/5 px-3 py-2 text-sm text-slate-200 placeholder-slate-500 outline-none backdrop-blur transition-colors focus:border-cyan-400/50 focus:ring-1 focus:ring-cyan-400/30 disabled:opacity-50';

function CreateUserModal({
  open,
  roles,
  onClose,
  onCreated,
  onError,
}: {
  open: boolean;
  roles: RoleItem[];
  onClose: () => void;
  onCreated: () => void | Promise<void>;
  onError: (msg: string) => void;
}) {
  const { t } = useI18n();
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [role, setRole] = useState('member');
  const [enabled, setEnabled] = useState(true);
  const [saving, setSaving] = useState(false);

  const [prevCreateOpen, setPrevCreateOpen] = useState(open);
  if (prevCreateOpen !== open) {
    setPrevCreateOpen(open);
    if (open) {
      setUsername('');
      setPassword('');
      setRole('member');
      setEnabled(true);
    }
  }

  const submit = async () => {
    if (!username.trim() || password.length < 4) {
      onError(t('admin.createUserHint'));
      return;
    }
    setSaving(true);
    const result = await api.createUser({ username: username.trim(), password, role, enabled });
    setSaving(false);
    if (result.error) {
      onError(result.error);
      return;
    }
    await onCreated();
  };

  return (
    <Modal open={open} onClose={() => !saving && onClose()} title={t('admin.createUser')} size="sm">
      <div className="space-y-3">
        <div>
          <label className="mb-1 block text-xs font-semibold uppercase tracking-widest text-slate-500">
            {t('admin.username')}
          </label>
          <input className={inputClass} value={username} onChange={(e) => setUsername(e.target.value)} disabled={saving} />
        </div>
        <div>
          <label className="mb-1 block text-xs font-semibold uppercase tracking-widest text-slate-500">
            {t('auth.password')}
          </label>
          <input
            type="password"
            className={inputClass}
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            disabled={saving}
          />
        </div>
        <div>
          <label className="mb-1 block text-xs font-semibold uppercase tracking-widest text-slate-500">
            {t('admin.role')}
          </label>
          <select className={inputClass} value={role} onChange={(e) => setRole(e.target.value)} disabled={saving}>
            {roles.map((r) => (
              <option key={r.name} value={r.name} className="bg-slate-900">
                {r.name}
              </option>
            ))}
          </select>
        </div>
        <label className="flex items-center gap-2 text-sm text-slate-300">
          <input type="checkbox" checked={enabled} onChange={(e) => setEnabled(e.target.checked)} disabled={saving} />
          {t('admin.enabled')}
        </label>
        <div className="flex justify-end gap-2 pt-2">
          <button
            type="button"
            onClick={onClose}
            disabled={saving}
            className="rounded-xl border border-white/10 bg-white/5 px-4 py-2 text-sm text-slate-200 hover:bg-white/10 disabled:opacity-50"
          >
            {t('common.cancel')}
          </button>
          <button
            type="button"
            onClick={() => void submit()}
            disabled={saving}
            className="rounded-xl bg-gradient-to-r from-cyan-500 to-violet-600 px-4 py-2 text-sm font-semibold text-white shadow-lg shadow-cyan-500/30 hover:brightness-110 disabled:opacity-50"
          >
            {saving ? t('admin.saving') : t('common.create')}
          </button>
        </div>
      </div>
    </Modal>
  );
}

function RoleEditorModal({
  open,
  role,
  onClose,
  onSaved,
  onError,
}: {
  open: boolean;
  role: RoleItem | null;
  onClose: () => void;
  onSaved: () => void | Promise<void>;
  onError: (msg: string) => void;
}) {
  const { t } = useI18n();
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [perms, setPerms] = useState<string[]>([]);
  const [saving, setSaving] = useState(false);

  const [prevRoleEditor, setPrevRoleEditor] = useState({ open, role });
  if (prevRoleEditor.open !== open || prevRoleEditor.role !== role) {
    setPrevRoleEditor({ open, role });
    if (open) {
      setName(role?.name ?? '');
      setDescription(role?.description ?? '');
      setPerms(role?.permissions ?? []);
    }
  }

  const toggle = (p: string) => {
    setPerms((prev) => (prev.includes(p) ? prev.filter((x) => x !== p) : [...prev, p]));
  };

  const submit = async () => {
    if (!name.trim()) {
      onError(t('admin.roleNameRequired'));
      return;
    }
    setSaving(true);
    const result = await api.saveRole({ name: name.trim(), description: description.trim(), permissions: perms });
    setSaving(false);
    if (result.error) {
      onError(result.error);
      return;
    }
    await onSaved();
  };

  return (
    <Modal open={open} onClose={() => !saving && onClose()} title={role ? t('common.edit') : t('admin.newRole')} size="md">
      <div className="space-y-3">
        <div>
          <label className="mb-1 block text-xs font-semibold uppercase tracking-widest text-slate-500">
            {t('admin.roleName')}
          </label>
          <input className={inputClass} value={name} onChange={(e) => setName(e.target.value)} disabled={saving || Boolean(role)} />
        </div>
        <div>
          <label className="mb-1 block text-xs font-semibold uppercase tracking-widest text-slate-500">
            {t('admin.description')}
          </label>
          <input
            className={inputClass}
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            disabled={saving}
          />
        </div>
        <div>
          <label className="mb-2 block text-xs font-semibold uppercase tracking-widest text-slate-500">
            {t('admin.permissions')}
          </label>
          <div className="grid grid-cols-1 gap-2 sm:grid-cols-2">
            {ALL_PERMISSIONS.map((p) => (
              <label
                key={p}
                className="flex items-center gap-2 rounded-lg border border-white/10 bg-white/5 px-3 py-2 text-xs text-slate-300"
              >
                <input type="checkbox" checked={perms.includes(p)} onChange={() => toggle(p)} disabled={saving} />
                <span className="font-mono">{p}</span>
              </label>
            ))}
          </div>
        </div>
        <div className="flex justify-end gap-2 pt-2">
          <button
            type="button"
            onClick={onClose}
            disabled={saving}
            className="rounded-xl border border-white/10 bg-white/5 px-4 py-2 text-sm text-slate-200 hover:bg-white/10 disabled:opacity-50"
          >
            {t('common.cancel')}
          </button>
          <button
            type="button"
            onClick={() => void submit()}
            disabled={saving}
            className="rounded-xl bg-gradient-to-r from-cyan-500 to-violet-600 px-4 py-2 text-sm font-semibold text-white shadow-lg shadow-cyan-500/30 hover:brightness-110 disabled:opacity-50"
          >
            {saving ? t('admin.saving') : t('common.save')}
          </button>
        </div>
      </div>
    </Modal>
  );
}

function ResetPasswordModal({
  open,
  username,
  onClose,
  onReset,
  saving,
}: {
  open: boolean;
  username: string;
  onClose: () => void;
  onReset: (username: string, password: string) => Promise<boolean | undefined>;
  saving: boolean;
}) {
  const { t } = useI18n();
  const [password, setPassword] = useState('');

  const [prevResetOpen, setPrevResetOpen] = useState(open);
  if (prevResetOpen !== open) {
    setPrevResetOpen(open);
    if (open) setPassword('');
  }

  return (
    <Modal open={open} onClose={() => !saving && onClose()} title={t('admin.resetPasswordTitle', { name: username })} size="sm">
      <div className="space-y-3">
        <div>
          <label className="mb-1 block text-xs font-semibold uppercase tracking-widest text-slate-500">
            {t('admin.newPassword')}
          </label>
          <input
            type="password"
            className={inputClass}
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            disabled={saving}
          />
        </div>
        <div className="flex justify-end gap-2 pt-2">
          <button
            type="button"
            onClick={onClose}
            disabled={saving}
            className="rounded-xl border border-white/10 bg-white/5 px-4 py-2 text-sm text-slate-200 hover:bg-white/10 disabled:opacity-50"
          >
            {t('common.cancel')}
          </button>
          <button
            type="button"
            disabled={saving || password.length < 4}
            onClick={() => {
              void (async () => {
                const ok = await onReset(username, password);
                if (ok) onClose();
              })();
            }}
            className="rounded-xl bg-gradient-to-r from-cyan-500 to-violet-600 px-4 py-2 text-sm font-semibold text-white shadow-lg shadow-cyan-500/30 hover:brightness-110 disabled:opacity-50"
          >
            {saving ? t('admin.saving') : t('common.save')}
          </button>
        </div>
      </div>
    </Modal>
  );
}

function ChangeOwnPasswordForm({ onDone, onError }: { onDone: () => void; onError: (m: string) => void }) {
  const { t } = useI18n();
  const [current, setCurrent] = useState('');
  const [next, setNext] = useState('');
  const [confirm, setConfirm] = useState('');
  const [saving, setSaving] = useState(false);

  const submit = async () => {
    if (next !== confirm) {
      onError(t('auth.passwordsMismatch'));
      return;
    }
    setSaving(true);
    const result = await api.changePassword(current, next);
    setSaving(false);
    if (result.error) {
      onError(result.error);
      return;
    }
    setCurrent('');
    setNext('');
    setConfirm('');
    onDone();
  };

  return (
    <div className="space-y-3">
      <div>
        <label className="mb-1 block text-xs font-semibold uppercase tracking-widest text-slate-500">
          {t('admin.currentPassword')}
        </label>
        <input type="password" className={inputClass} value={current} onChange={(e) => setCurrent(e.target.value)} disabled={saving} />
      </div>
      <div>
        <label className="mb-1 block text-xs font-semibold uppercase tracking-widest text-slate-500">
          {t('admin.newPassword')}
        </label>
        <input type="password" className={inputClass} value={next} onChange={(e) => setNext(e.target.value)} disabled={saving} />
      </div>
      <div>
        <label className="mb-1 block text-xs font-semibold uppercase tracking-widest text-slate-500">
          {t('auth.confirmPassword')}
        </label>
        <input type="password" className={inputClass} value={confirm} onChange={(e) => setConfirm(e.target.value)} disabled={saving} />
      </div>
      <button
        type="button"
        onClick={() => void submit()}
        disabled={saving || !current || next.length < 4}
        className="w-full rounded-xl bg-gradient-to-r from-cyan-500 to-violet-600 px-4 py-2 text-sm font-semibold text-white shadow-lg shadow-cyan-500/30 hover:brightness-110 disabled:opacity-50"
      >
        {saving ? t('admin.saving') : t('common.save')}
      </button>
    </div>
  );
}
