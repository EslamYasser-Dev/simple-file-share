import { useCallback, useEffect, useState } from 'react';
import {
  ActivityIndicator,
  Alert,
  FlatList,
  Modal,
  Pressable,
  RefreshControl,
  Share,
  StyleSheet,
  Text,
  TextInput,
  View,
} from 'react-native';
import * as DocumentPicker from 'expo-document-picker';
import * as FileSystem from 'expo-file-system';
import * as Sharing from 'expo-sharing';
import { router } from 'expo-router';
import { api, type FileItem } from '../../services/api';
import { useAuth } from '../../services/auth';
import { subscribeEvents } from '../../services/events';
import { baseName, formatBytes, joinPath, parentPath } from '../../lib/format';
import { colors } from '../_layout';

export default function FilesScreen() {
  const { user, refreshUser } = useAuth();
  const [path, setPath] = useState('');
  const [items, setItems] = useState<FileItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [selected, setSelected] = useState<FileItem | null>(null);
  const [uploadPct, setUploadPct] = useState<number | null>(null);
  const [newFolderOpen, setNewFolderOpen] = useState(false);
  const [newFolderName, setNewFolderName] = useState('');
  const [sharePath, setSharePath] = useState<string | null>(null);
  const [shareUrl, setShareUrl] = useState<string | null>(null);
  const [shareBusy, setShareBusy] = useState(false);

  const load = useCallback(async (target: string) => {
    const res = await api.listFiles(target);
    if (res.error) {
      setError(res.error);
      setItems([]);
    } else {
      setError(null);
      setItems((res.data ?? []).slice().sort((a, b) => {
        if (a.isDir !== b.isDir) return a.isDir ? -1 : 1;
        return a.name.localeCompare(b.name);
      }));
    }
    setLoading(false);
    setRefreshing(false);
  }, []);

  useEffect(() => {
    let cancelled = false;
    const run = async () => {
      const res = await api.listFiles(path);
      if (cancelled) return;
      if (res.error) {
        setError(res.error);
        setItems([]);
      } else {
        setError(null);
        setItems((res.data ?? []).slice().sort((a, b) => {
          if (a.isDir !== b.isDir) return a.isDir ? -1 : 1;
          return a.name.localeCompare(b.name);
        }));
      }
      setLoading(false);
      setRefreshing(false);
    };
    void run();
    return () => {
      cancelled = true;
    };
  }, [path]);

  useEffect(() => {
    return subscribeEvents(() => {
      void load(path);
      void refreshUser();
    });
  }, [path, load, refreshUser]);

  const navigateInto = (item: FileItem) => {
    if (item.isDir) setPath(item.path);
    else setSelected(item);
  };

  const goUp = () => setPath(parentPath(path));

  const upload = async () => {
    const result = await DocumentPicker.getDocumentAsync({ copyToCacheDirectory: true, multiple: false });
    if (result.canceled || !result.assets?.length) return;
    const asset = result.assets[0];
    setUploadPct(0);
    const res = await api.uploadFile(
      asset.uri,
      asset.name,
      asset.mimeType ?? 'application/octet-stream',
      path,
      (loaded, total) => {
        if (total > 0) setUploadPct(Math.round((loaded / total) * 100));
      },
    );
    setUploadPct(null);
    if (res.error) {
      Alert.alert('Upload failed', res.error);
      return;
    }
    await load(path);
    await refreshUser();
  };

  const createFolder = async () => {
    const name = newFolderName.trim();
    if (!name) return;
    const full = joinPath(path, name);
    const res = await api.createDirectory(full);
    setNewFolderOpen(false);
    setNewFolderName('');
    if (res.error) {
      Alert.alert('Create failed', res.error);
      return;
    }
    await load(path);
  };

  const removeItem = (item: FileItem) => {
    Alert.alert('Delete', `Delete ${item.name}?`, [
      { text: 'Cancel', style: 'cancel' },
      {
        text: 'Delete',
        style: 'destructive',
        onPress: () => {
          void (async () => {
            const res = await api.deletePath(item.path);
            if (res.error) Alert.alert('Delete failed', res.error);
            setSelected(null);
            await load(path);
            await refreshUser();
          })();
        },
      },
    ]);
  };

  const downloadItem = async (item: FileItem) => {
    if (item.isDir) {
      Alert.alert('Folders', 'Folder download opens as a ZIP via the share link or web app.');
      return;
    }
    try {
      const res = await api.fetchDownload(item.path);
      if (res.error || !res.data) {
        Alert.alert('Download failed', res.error || 'Unknown error');
        return;
      }
      const { blob, name } = res.data;
      const bytes = await new Response(blob).arrayBuffer();
      const file = new FileSystem.File(FileSystem.Paths.cache, name);
      if (file.exists) file.delete();
      file.create({ overwrite: true });
      file.write(new Uint8Array(bytes));
      const canShare = await Sharing.isAvailableAsync();
      if (canShare) {
        await Sharing.shareAsync(file.uri, { dialogTitle: name, mimeType: blob.type || undefined });
      } else {
        Alert.alert('Saved', `Saved to cache as ${name}`);
      }
      setSelected(null);
    } catch (e) {
      Alert.alert('Download failed', e instanceof Error ? e.message : 'Network error');
    }
  };

  const openShare = async (item: FileItem) => {
    setSharePath(item.path);
    setShareUrl(null);
    setShareBusy(true);
    const res = await api.createShare(item.path, 0);
    setShareBusy(false);
    if (res.error || !res.data) {
      Alert.alert('Share failed', res.error || 'Unknown error');
      setSharePath(null);
      return;
    }
    setShareUrl(api.shareUrl(res.data.token));
  };

  const shareLink = async () => {
    if (!shareUrl) return;
    await Share.share({ message: shareUrl, url: shareUrl });
  };

  const quota = user?.quotaBytes ?? 0;
  const used = user?.size ?? 0;
  const pct = quota > 0 ? Math.min(100, Math.round((used / quota) * 100)) : 0;

  return (
    <View style={styles.root}>
      <View style={styles.toolbar}>
        <Pressable onPress={goUp} disabled={!path} style={[styles.toolBtn, !path && styles.disabled]}>
          <Text style={styles.toolBtnText}>↑</Text>
        </Pressable>
        <Text style={styles.crumb} numberOfLines={1}>
          /{path}
        </Text>
        <Pressable onPress={() => setNewFolderOpen(true)} style={styles.toolBtn}>
          <Text style={styles.toolBtnText}>+DIR</Text>
        </Pressable>
        <Pressable onPress={() => void upload()} style={[styles.toolBtn, styles.primaryBtn]}>
          <Text style={[styles.toolBtnText, styles.primaryBtnText]}>
            {uploadPct === null ? '↑ Upload' : `${uploadPct}%`}
          </Text>
        </Pressable>
      </View>

      {quota > 0 ? (
        <View style={styles.usage}>
          <Text style={styles.usageText}>
            {formatBytes(used)} of {formatBytes(quota)} used · {user?.files ?? 0} file(s)
          </Text>
          <View style={styles.usageTrack}>
            <View
              style={[
                styles.usageFill,
                { width: `${pct}%` },
                pct >= 90 ? styles.usageFillWarn : null,
              ]}
            />
          </View>
        </View>
      ) : null}

      {loading ? (
        <ActivityIndicator color={colors.accent} style={{ marginTop: 40 }} />
      ) : error ? (
        <View style={styles.center}>
          <Text style={styles.errorText}>{error}</Text>
          <Pressable onPress={() => void load(path)} style={styles.retryBtn}>
            <Text style={styles.primaryBtnText}>Retry</Text>
          </Pressable>
        </View>
      ) : (
        <FlatList
          data={items}
          keyExtractor={(item) => item.path}
          refreshControl={
            <RefreshControl
              refreshing={refreshing}
              tintColor={colors.accent}
              onRefresh={() => {
                setRefreshing(true);
                void load(path);
              }}
            />
          }
          ListEmptyComponent={
            <View style={styles.center}>
              <Text style={styles.muted}>Empty folder</Text>
            </View>
          }
          renderItem={({ item }) => (
            <Pressable style={styles.row} onPress={() => navigateInto(item)}>
              <Text style={styles.rowIcon}>{item.isDir ? '📁' : '📄'}</Text>
              <View style={styles.rowBody}>
                <Text style={styles.rowName} numberOfLines={1}>
                  {item.name}
                </Text>
                <Text style={styles.rowMeta}>
                  {item.isDir ? 'Folder' : formatBytes(item.size)}
                </Text>
              </View>
              {!item.isDir ? (
                <Pressable onPress={() => setSelected(item)} hitSlop={8} style={styles.moreBtn}>
                  <Text style={styles.moreText}>⋯</Text>
                </Pressable>
              ) : null}
            </Pressable>
          )}
        />
      )}

      <Modal visible={newFolderOpen} transparent animationType="fade">
        <View style={styles.modalBack}>
          <View style={styles.modalCard}>
            <Text style={styles.modalTitle}>New folder</Text>
            <TextInput
              style={styles.input}
              value={newFolderName}
              onChangeText={setNewFolderName}
              placeholder="Folder name"
              placeholderTextColor={colors.muted}
              autoCapitalize="none"
            />
            <View style={styles.modalActions}>
              <Pressable onPress={() => setNewFolderOpen(false)} style={styles.ghostBtn}>
                <Text style={styles.muted}>Cancel</Text>
              </Pressable>
              <Pressable onPress={() => void createFolder()} style={styles.primarySmall}>
                <Text style={styles.primaryBtnText}>Create</Text>
              </Pressable>
            </View>
          </View>
        </View>
      </Modal>

      <Modal visible={sharePath !== null} transparent animationType="fade">
        <View style={styles.modalBack}>
          <View style={styles.modalCard}>
            <Text style={styles.modalTitle}>Share {baseName(sharePath || '')}</Text>
            {shareBusy ? (
              <ActivityIndicator color={colors.accent} style={{ marginVertical: 16 }} />
            ) : shareUrl ? (
              <>
                <Text style={styles.shareUrl} selectable>
                  {shareUrl}
                </Text>
                <View style={styles.modalActions}>
                  <Pressable onPress={() => setSharePath(null)} style={styles.ghostBtn}>
                    <Text style={styles.muted}>Close</Text>
                  </Pressable>
                  <Pressable onPress={() => void shareLink()} style={styles.primarySmall}>
                    <Text style={styles.primaryBtnText}>Share</Text>
                  </Pressable>
                </View>
              </>
            ) : null}
          </View>
        </View>
      </Modal>

      <Modal visible={selected !== null} transparent animationType="fade">
        <View style={styles.modalBack}>
          <View style={styles.modalCard}>
            <Text style={styles.modalTitle} numberOfLines={2}>
              {selected?.name}
            </Text>
            {selected ? (
              <Text style={styles.muted}>{formatBytes(selected.size)}</Text>
            ) : null}
            <View style={styles.actionsCol}>
              {selected && !selected.isDir ? (
                <Pressable style={styles.actionBtn} onPress={() => void downloadItem(selected)}>
                  <Text style={styles.actionText}>Download & share</Text>
                </Pressable>
              ) : null}
              {selected ? (
                <Pressable style={styles.actionBtn} onPress={() => void openShare(selected)}>
                  <Text style={styles.actionText}>Create link</Text>
                </Pressable>
              ) : null}
              {selected ? (
                <Pressable
                  style={[styles.actionBtn, styles.dangerBtn]}
                  onPress={() => removeItem(selected)}
                >
                  <Text style={[styles.actionText, styles.dangerText]}>Delete</Text>
                </Pressable>
              ) : null}
              <Pressable style={[styles.actionBtn, styles.ghostBtn]} onPress={() => setSelected(null)}>
                <Text style={styles.muted}>Cancel</Text>
              </Pressable>
            </View>
          </View>
        </View>
      </Modal>

      <Pressable
        style={styles.fab}
        onPress={() => router.push('/(tabs)/shares')}
      >
        <Text style={styles.fabText}>🔗</Text>
      </Pressable>
    </View>
  );
}

const styles = StyleSheet.create({
  root: { flex: 1, backgroundColor: colors.background },
  toolbar: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
    paddingHorizontal: 12,
    paddingVertical: 10,
    borderBottomWidth: 1,
    borderBottomColor: colors.border,
  },
  toolBtn: {
    backgroundColor: colors.card,
    borderColor: colors.border,
    borderWidth: 1,
    borderRadius: 10,
    paddingHorizontal: 10,
    paddingVertical: 8,
  },
  toolBtnText: { color: colors.text, fontSize: 12, fontWeight: '700' },
  primaryBtn: { backgroundColor: colors.accent, borderColor: colors.accent },
  primaryBtnText: { color: '#0B1220', fontWeight: '800' },
  disabled: { opacity: 0.4 },
  crumb: { flex: 1, color: colors.muted, fontSize: 13 },
  usage: { paddingHorizontal: 16, paddingTop: 10 },
  usageText: { color: colors.muted, fontSize: 11, marginBottom: 6 },
  usageTrack: {
    height: 6,
    backgroundColor: 'rgba(255,255,255,0.08)',
    borderRadius: 999,
    overflow: 'hidden',
  },
  usageFill: { height: '100%', backgroundColor: colors.accent, borderRadius: 999 },
  usageFillWarn: { backgroundColor: '#FBBF24' },
  row: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: 16,
    paddingVertical: 14,
    borderBottomWidth: 1,
    borderBottomColor: colors.border,
    gap: 12,
  },
  rowIcon: { fontSize: 20 },
  rowBody: { flex: 1 },
  rowName: { color: colors.text, fontSize: 15, fontWeight: '600' },
  rowMeta: { color: colors.muted, fontSize: 12, marginTop: 2 },
  moreBtn: { paddingHorizontal: 8 },
  moreText: { color: colors.muted, fontSize: 22 },
  center: { alignItems: 'center', marginTop: 48, gap: 16, paddingHorizontal: 24 },
  muted: { color: colors.muted, fontSize: 14 },
  errorText: { color: '#F87171', textAlign: 'center' },
  retryBtn: {
    backgroundColor: colors.accent,
    borderRadius: 10,
    paddingHorizontal: 18,
    paddingVertical: 10,
  },
  modalBack: {
    flex: 1,
    backgroundColor: 'rgba(0,0,0,0.55)',
    alignItems: 'center',
    justifyContent: 'center',
    padding: 20,
  },
  modalCard: {
    width: '100%',
    maxWidth: 360,
    backgroundColor: colors.card,
    borderColor: colors.border,
    borderWidth: 1,
    borderRadius: 16,
    padding: 18,
  },
  modalTitle: { color: colors.text, fontSize: 16, fontWeight: '700', marginBottom: 8 },
  input: {
    backgroundColor: 'rgba(255,255,255,0.05)',
    borderColor: colors.border,
    borderWidth: 1,
    borderRadius: 10,
    color: colors.text,
    paddingHorizontal: 12,
    paddingVertical: 10,
  },
  modalActions: {
    flexDirection: 'row',
    justifyContent: 'flex-end',
    gap: 10,
    marginTop: 16,
    alignItems: 'center',
  },
  ghostBtn: { paddingHorizontal: 12, paddingVertical: 10 },
  primarySmall: {
    backgroundColor: colors.accent,
    borderRadius: 10,
    paddingHorizontal: 14,
    paddingVertical: 10,
  },
  shareUrl: { color: colors.accent, fontSize: 13, marginBottom: 4 },
  actionsCol: { marginTop: 16, gap: 8 },
  actionBtn: {
    borderRadius: 10,
    borderWidth: 1,
    borderColor: colors.border,
    paddingVertical: 12,
    alignItems: 'center',
    backgroundColor: 'rgba(255,255,255,0.03)',
  },
  actionText: { color: colors.text, fontWeight: '600' },
  dangerBtn: { borderColor: 'rgba(248,113,113,0.4)' },
  dangerText: { color: '#F87171' },
  fab: {
    position: 'absolute',
    right: 16,
    bottom: 24,
    width: 52,
    height: 52,
    borderRadius: 26,
    backgroundColor: colors.accent,
    alignItems: 'center',
    justifyContent: 'center',
    elevation: 4,
  },
  fabText: { fontSize: 22 },
});
