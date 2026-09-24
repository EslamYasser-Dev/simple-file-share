import { useCallback, useEffect, useState } from 'react';
import {
  ActivityIndicator,
  Alert,
  FlatList,
  Pressable,
  Share,
  StyleSheet,
  Text,
  View,
} from 'react-native';
import { api, type ShareItem } from '../../services/api';
import { formatDate } from '../../lib/format';
import { colors } from '../_layout';

export default function SharesScreen() {
  const [items, setItems] = useState<ShareItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    const res = await api.listShares();
    if (res.error) setError(res.error);
    else {
      setError(null);
      setItems(res.data ?? []);
    }
    setLoading(false);
  }, []);

  useEffect(() => {
    let cancelled = false;
    const run = async () => {
      const res = await api.listShares();
      if (cancelled) return;
      if (res.error) setError(res.error);
      else {
        setError(null);
        setItems(res.data ?? []);
      }
      setLoading(false);
    };
    void run();
    return () => {
      cancelled = true;
    };
  }, []);

  const revoke = (token: string) => {
    Alert.alert('Revoke', 'Revoke this share link?', [
      { text: 'Cancel', style: 'cancel' },
      {
        text: 'Revoke',
        style: 'destructive',
        onPress: () => {
          void (async () => {
            const res = await api.revokeShare(token);
            if (res.error) Alert.alert('Revoke failed', res.error);
            await load();
          })();
        },
      },
    ]);
  };

  const copy = async (token: string) => {
    await Share.share({ message: api.shareUrl(token), url: api.shareUrl(token) });
  };

  if (loading) {
    return (
      <View style={styles.root}>
        <ActivityIndicator color={colors.accent} style={{ marginTop: 40 }} />
      </View>
    );
  }

  if (error) {
    return (
      <View style={styles.root}>
        <View style={styles.center}>
          <Text style={styles.error}>{error}</Text>
          <Pressable onPress={() => void load()} style={styles.retry}>
            <Text style={styles.retryText}>Retry</Text>
          </Pressable>
        </View>
      </View>
    );
  }

  return (
    <View style={styles.root}>
      <FlatList
        data={items}
        keyExtractor={(item) => item.token}
        contentContainerStyle={items.length === 0 ? styles.center : undefined}
        ListEmptyComponent={<Text style={styles.muted}>No share links yet</Text>}
        renderItem={({ item }) => (
          <View style={styles.card}>
            <Text style={styles.name} numberOfLines={1}>
              {item.name}
            </Text>
            <Text style={styles.meta} numberOfLines={1}>
              {item.path}
            </Text>
            <Text style={styles.metaSmall}>Expires: {formatDate(item.expiresAt)}</Text>
            <View style={styles.actions}>
              <Pressable onPress={() => void copy(item.token)} style={styles.btn}>
                <Text style={styles.btnText}>Share</Text>
              </Pressable>
              <Pressable onPress={() => revoke(item.token)} style={[styles.btn, styles.danger]}>
                <Text style={[styles.btnText, styles.dangerText]}>Revoke</Text>
              </Pressable>
            </View>
          </View>
        )}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  root: { flex: 1, backgroundColor: colors.background },
  center: { alignItems: 'center', justifyContent: 'center', marginTop: 64, gap: 16, padding: 24 },
  muted: { color: colors.muted },
  error: { color: '#F87171', textAlign: 'center' },
  retry: {
    backgroundColor: colors.accent,
    borderRadius: 10,
    paddingHorizontal: 18,
    paddingVertical: 10,
  },
  retryText: { color: '#0B1220', fontWeight: '800' },
  card: {
    marginHorizontal: 16,
    marginTop: 12,
    backgroundColor: colors.card,
    borderColor: colors.border,
    borderWidth: 1,
    borderRadius: 14,
    padding: 14,
  },
  name: { color: colors.text, fontWeight: '700', fontSize: 15 },
  meta: { color: colors.muted, marginTop: 4, fontSize: 13 },
  metaSmall: { color: colors.muted, marginTop: 2, fontSize: 11 },
  actions: { flexDirection: 'row', gap: 10, marginTop: 12 },
  btn: {
    flex: 1,
    borderRadius: 10,
    paddingVertical: 10,
    alignItems: 'center',
    backgroundColor: 'rgba(255,255,255,0.04)',
    borderWidth: 1,
    borderColor: colors.border,
  },
  btnText: { color: colors.text, fontWeight: '700', fontSize: 13 },
  danger: { borderColor: 'rgba(248,113,113,0.4)' },
  dangerText: { color: '#F87171' },
});
