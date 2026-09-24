import { Pressable, StyleSheet, Text, View } from 'react-native';
import { router } from 'expo-router';
import { useAuth } from '../../services/auth';
import { formatBytes, formatDate } from '../../lib/format';
import { colors } from '../_layout';

export default function AccountScreen() {
  const { user, signOut } = useAuth();

  const onSignOut = async () => {
    await signOut();
    router.replace('/login');
  };

  const quota = user?.quotaBytes ?? 0;
  const used = user?.size ?? 0;
  const pct = quota > 0 ? Math.min(100, Math.round((used / quota) * 100)) : 0;

  return (
    <View style={styles.root}>
      <View style={styles.card}>
        <Text style={styles.label}>Signed in as</Text>
        <Text style={styles.username}>{user?.username ?? '—'}</Text>
        <Text style={styles.meta}>
          {user?.isAdmin ? 'Admin' : user?.role || 'Member'}
          {user?.createdAt ? ` · joined ${formatDate(user.createdAt)}` : ''}
        </Text>

        <View style={styles.stats}>
          <View style={styles.stat}>
            <Text style={styles.statValue}>{user?.files ?? 0}</Text>
            <Text style={styles.statLabel}>Files</Text>
          </View>
          <View style={styles.stat}>
            <Text style={styles.statValue}>{formatBytes(used)}</Text>
            <Text style={styles.statLabel}>Used</Text>
          </View>
          <View style={styles.stat}>
            <Text style={styles.statValue}>{quota > 0 ? formatBytes(quota) : '∞'}</Text>
            <Text style={styles.statLabel}>Quota</Text>
          </View>
        </View>

        {quota > 0 ? (
          <View style={styles.usageWrap}>
            <Text style={styles.usageText}>
              {formatBytes(used)} of {formatBytes(quota)} used ({pct}%)
            </Text>
            <View style={styles.usageTrack}>
              <View
                style={[
                  styles.usageFill,
                  { width: `${pct}%` },
                  pct >= 90 ? styles.usageWarn : null,
                ]}
              />
            </View>
          </View>
        ) : (
          <Text style={styles.usageText}>Unlimited storage</Text>
        )}
      </View>

      <Pressable style={styles.signOut} onPress={() => void onSignOut()}>
        <Text style={styles.signOutText}>Sign out</Text>
      </Pressable>

      <Text style={styles.footer}>Simple File Share · REST API</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  root: { flex: 1, backgroundColor: colors.background, padding: 16 },
  card: {
    backgroundColor: colors.card,
    borderColor: colors.border,
    borderWidth: 1,
    borderRadius: 16,
    padding: 18,
  },
  label: {
    color: colors.muted,
    fontSize: 11,
    fontWeight: '700',
    letterSpacing: 1.2,
    textTransform: 'uppercase',
  },
  username: { color: colors.text, fontSize: 24, fontWeight: '800', marginTop: 6 },
  meta: { color: colors.muted, marginTop: 4, fontSize: 13 },
  stats: { flexDirection: 'row', gap: 12, marginTop: 20 },
  stat: {
    flex: 1,
    backgroundColor: 'rgba(255,255,255,0.04)',
    borderRadius: 12,
    paddingVertical: 14,
    alignItems: 'center',
  },
  statValue: { color: colors.text, fontWeight: '800', fontSize: 16 },
  statLabel: { color: colors.muted, fontSize: 11, marginTop: 4 },
  usageWrap: { marginTop: 18 },
  usageText: { color: colors.muted, fontSize: 12, marginBottom: 8 },
  usageTrack: {
    height: 8,
    backgroundColor: 'rgba(255,255,255,0.08)',
    borderRadius: 999,
    overflow: 'hidden',
  },
  usageFill: { height: '100%', backgroundColor: colors.accent, borderRadius: 999 },
  usageWarn: { backgroundColor: '#FBBF24' },
  signOut: {
    marginTop: 20,
    borderRadius: 12,
    borderWidth: 1,
    borderColor: 'rgba(248,113,113,0.45)',
    paddingVertical: 14,
    alignItems: 'center',
  },
  signOutText: { color: '#F87171', fontWeight: '700' },
  footer: { color: colors.muted, textAlign: 'center', marginTop: 'auto', fontSize: 12 },
});
