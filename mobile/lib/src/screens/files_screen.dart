import 'dart:async';
import 'dart:io';
import 'dart:typed_data';

import 'package:file_picker/file_picker.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:share_plus/share_plus.dart';

import '../format.dart';
import '../models.dart';
import '../state/auth_controller.dart';
import '../theme.dart';

class FilesScreen extends ConsumerStatefulWidget {
  const FilesScreen({super.key, this.onOpenShares});

  final VoidCallback? onOpenShares;

  @override
  ConsumerState<FilesScreen> createState() => _FilesScreenState();
}

class _FilesScreenState extends ConsumerState<FilesScreen> {
  String _path = '';
  List<FileItem> _items = const [];
  bool _loading = true;
  String? _error;
  int? _uploadPct;
  final _newFolderName = TextEditingController();
  StreamSubscription<ServerEvent>? _eventsSub;

  @override
  void initState() {
    super.initState();
    unawaited(_load(_path));
    _eventsSub = ref.read(eventsServiceProvider).stream.listen((_) {
      unawaited(_load(_path));
      unawaited(ref.read(authControllerProvider.notifier).refreshUser());
    });
  }

  @override
  void dispose() {
    _eventsSub?.cancel();
    _newFolderName.dispose();
    super.dispose();
  }

  Future<void> _load(String target) async {
    final res = await ref.read(apiClientProvider).listFiles(target);
    if (!mounted || target != _path) return;
    setState(() {
      if (res.error != null) {
        _error = res.error;
        _items = const [];
      } else {
        _error = null;
        final sorted = [...?res.data]
          ..sort((a, b) {
            if (a.isDir != b.isDir) return a.isDir ? -1 : 1;
            return a.name.toLowerCase().compareTo(b.name.toLowerCase());
          });
        _items = sorted;
      }
      _loading = false;
    });
  }

  Future<void> _refresh() async {
    await Future.wait([
      _load(_path),
      ref.read(authControllerProvider.notifier).refreshUser(),
    ]);
  }

  void _navigate(String path) {
    setState(() {
      _path = path;
      _loading = true;
      _items = const [];
      _error = null;
    });
    unawaited(_load(path));
  }

  void _goUp() {
    if (_path.isNotEmpty) _navigate(parentPath(_path));
  }

  void _showError(String title, String message) {
    showDialog<void>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(title, style: const TextStyle(color: SfsColors.danger)),
        content: Text(message),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('OK'),
          ),
        ],
      ),
    );
  }

  Future<void> _openNewFolder() async {
    _newFolderName.clear();
    final name = await showDialog<String>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('New folder'),
        content: TextField(
          controller: _newFolderName,
          autofocus: true,
          decoration: const InputDecoration(hintText: 'Folder name'),
          onSubmitted: (v) => Navigator.pop(context, v.trim()),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Cancel'),
          ),
          FilledButton(
            style: FilledButton.styleFrom(backgroundColor: SfsColors.accent),
            onPressed: () =>
                Navigator.pop(context, _newFolderName.text.trim()),
            child: const Text('Create'),
          ),
        ],
      ),
    );
    if (name == null || name.isEmpty || !mounted) return;
    final res = await ref
        .read(apiClientProvider)
        .createDirectory(joinPath(_path, name));
    if (!mounted) return;
    if (!res.ok) {
      _showError('Create failed', res.error ?? 'Unknown error');
      return;
    }
    unawaited(_load(_path));
    unawaited(ref.read(authControllerProvider.notifier).refreshUser());
  }

  Future<void> _upload() async {
    if (_uploadPct != null) return;
    final picked = await FilePicker.pickFiles();
    if (picked.isEmpty || !mounted) return;
    final pf = picked.first;
    File? file;
    Uint8List? bytes;
    int size;
    if (pf.path != null) {
      file = File(pf.path!);
      size = await file.length();
    } else {
      bytes = await pf.readAsBytes();
      size = bytes.length;
    }
    setState(() => _uploadPct = 0);
    final res = await ref.read(apiClientProvider).uploadFile(
          fileName: pf.name,
          size: size,
          dirPath: _path,
          file: file,
          bytes: bytes,
          onProgress: (loaded, total) {
            if (!mounted) return;
            final pct = total > 0 ? ((loaded / total) * 100).round() : 0;
            setState(() => _uploadPct = pct.clamp(0, 100));
          },
        );
    if (!mounted) return;
    setState(() => _uploadPct = null);
    if (!res.ok) {
      _showError('Upload failed', res.error ?? 'Unknown error');
      return;
    }
    await Future.wait([
      _load(_path),
      ref.read(authControllerProvider.notifier).refreshUser(),
    ]);
  }

  Future<void> _downloadAndShare(FileItem item) async {
    final res = await ref.read(apiClientProvider).download(item.path);
    if (!mounted) return;
    if (!res.ok || res.data == null) {
      _showError('Download failed', res.error ?? 'Unknown error');
      return;
    }
    await SharePlus.instance.share(ShareParams(
      files: [XFile(res.data!.file.path)],
      title: res.data!.name,
    ));
  }

  Future<void> _confirmDelete(FileItem item) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Delete'),
        content: Text('Delete ${item.name}?'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Cancel'),
          ),
          FilledButton(
            style: FilledButton.styleFrom(
              backgroundColor: SfsColors.danger,
              foregroundColor: Colors.white,
            ),
            onPressed: () => Navigator.pop(context, true),
            child: const Text('Delete'),
          ),
        ],
      ),
    );
    if (confirmed != true || !mounted) return;
    final res = await ref.read(apiClientProvider).deletePath(item.path);
    if (!mounted) return;
    if (!res.ok) {
      _showError('Delete failed', res.error ?? 'Unknown error');
      return;
    }
    unawaited(_load(_path));
    unawaited(ref.read(authControllerProvider.notifier).refreshUser());
  }

  void _openActions(FileItem item) {
    showDialog<void>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(item.name, maxLines: 1, overflow: TextOverflow.ellipsis),
        content: Text(formatBytes(item.size)),
        actions: [
          TextButton(
            onPressed: () {
              Navigator.pop(context);
              unawaited(_downloadAndShare(item));
            },
            child: const Text('Download & share'),
          ),
          TextButton(
            onPressed: () {
              Navigator.pop(context);
              showDialog<void>(
                context: this.context,
                builder: (context) => _ShareDialog(item: item),
              );
            },
            child: const Text('Create link'),
          ),
          TextButton(
            style: TextButton.styleFrom(foregroundColor: SfsColors.danger),
            onPressed: () {
              Navigator.pop(context);
              unawaited(_confirmDelete(item));
            },
            child: const Text('Delete'),
          ),
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Cancel'),
          ),
        ],
      ),
    );
  }

  Widget _buildBody() {
    if (_loading && _items.isEmpty && _error == null) {
      return const Center(child: CircularProgressIndicator());
    }
    if (_error != null && _items.isEmpty) {
      return Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 32),
              child: Text(
                _error!,
                textAlign: TextAlign.center,
                style: const TextStyle(color: SfsColors.danger),
              ),
            ),
            const SizedBox(height: 16),
            FilledButton(
              style: FilledButton.styleFrom(
                backgroundColor: SfsColors.accent,
                foregroundColor: SfsColors.onAccent,
              ),
              onPressed: () {
                setState(() {
                  _loading = true;
                  _error = null;
                });
                unawaited(_load(_path));
              },
              child: const Text('Retry'),
            ),
          ],
        ),
      );
    }
    return RefreshIndicator(
      color: SfsColors.accent,
      backgroundColor: SfsColors.card,
      onRefresh: _refresh,
      child: _items.isEmpty
          ? ListView(
              physics: const AlwaysScrollableScrollPhysics(),
              children: const [
                SizedBox(height: 160),
                Center(
                  child: Text(
                    'Empty folder',
                    style: TextStyle(color: SfsColors.muted, fontSize: 15),
                  ),
                ),
              ],
            )
          : ListView.separated(
              physics: const AlwaysScrollableScrollPhysics(),
              itemCount: _items.length,
              separatorBuilder: (_, _) => const Divider(
                height: 1,
                indent: 56,
                color: SfsColors.border,
              ),
              itemBuilder: (context, index) {
                final item = _items[index];
                return ListTile(
                  dense: true,
                  leading: Text(
                    item.isDir ? '📁' : '📄',
                    style: const TextStyle(fontSize: 20),
                  ),
                  title: Text(
                    item.name,
                    style: const TextStyle(
                      color: SfsColors.text,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                  subtitle: item.isDir
                      ? null
                      : Text(
                          formatBytes(item.size),
                          style: const TextStyle(
                            color: SfsColors.muted,
                            fontSize: 12,
                          ),
                        ),
                  trailing: item.isDir
                      ? const Icon(Icons.chevron_right, color: SfsColors.muted)
                      : IconButton(
                          icon: const Icon(
                            Icons.more_vert,
                            color: SfsColors.muted,
                            size: 20,
                          ),
                          onPressed: () => _openActions(item),
                        ),
                  onTap: () {
                    if (item.isDir) {
                      _navigate(item.path);
                    } else {
                      _openActions(item);
                    }
                  },
                );
              },
            ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final user = ref.watch(authControllerProvider).user;
    final quota = user?.quotaBytes ?? 0;
    final used = user?.size ?? 0;

    return Scaffold(
      backgroundColor: SfsColors.background,
      floatingActionButton: FloatingActionButton(
        backgroundColor: SfsColors.accent,
        foregroundColor: SfsColors.onAccent,
        onPressed: widget.onOpenShares,
        child: const Text('🔗', style: TextStyle(fontSize: 20)),
      ),
      body: SafeArea(
        child: Column(
          children: [
            Padding(
              padding: const EdgeInsets.fromLTRB(8, 8, 8, 4),
              child: Row(
                children: [
                  IconButton(
                    icon: const Icon(Icons.arrow_upward, size: 20),
                    color: SfsColors.muted,
                    onPressed: _path.isEmpty ? null : _goUp,
                  ),
                  Expanded(
                    child: Text(
                      _path.isEmpty ? '/' : '/$_path',
                      overflow: TextOverflow.ellipsis,
                      style: const TextStyle(
                        color: SfsColors.muted,
                        fontSize: 13,
                      ),
                    ),
                  ),
                  OutlinedButton(
                    style: OutlinedButton.styleFrom(
                      foregroundColor: SfsColors.text,
                      side: const BorderSide(color: SfsColors.border),
                      padding: const EdgeInsets.symmetric(
                        horizontal: 12,
                        vertical: 8,
                      ),
                    ),
                    onPressed: _openNewFolder,
                    child: const Text('+DIR'),
                  ),
                  const SizedBox(width: 8),
                  FilledButton(
                    style: FilledButton.styleFrom(
                      backgroundColor: SfsColors.accent,
                      foregroundColor: SfsColors.onAccent,
                      padding: const EdgeInsets.symmetric(
                        horizontal: 14,
                        vertical: 8,
                      ),
                    ),
                    onPressed: _uploadPct != null ? null : _upload,
                    child: Text(_uploadPct != null ? '$_uploadPct%' : 'Upload'),
                  ),
                ],
              ),
            ),
            if (quota > 0)
              Padding(
                padding: const EdgeInsets.fromLTRB(16, 4, 16, 4),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    ClipRRect(
                      borderRadius: BorderRadius.circular(999),
                      child: LinearProgressIndicator(
                        value: (used / quota).clamp(0.0, 1.0),
                        minHeight: 8,
                        backgroundColor: SfsColors.surfaceOverlay,
                        color: used >= quota
                            ? SfsColors.danger
                            : SfsColors.accent,
                      ),
                    ),
                    const SizedBox(height: 4),
                    Text(
                      '${formatBytes(used)} of ${formatBytes(quota)} used',
                      style: const TextStyle(
                        color: SfsColors.muted,
                        fontSize: 12,
                      ),
                    ),
                  ],
                ),
              ),
            Expanded(child: _buildBody()),
          ],
        ),
      ),
    );
  }
}

class _ShareDialog extends ConsumerStatefulWidget {
  const _ShareDialog({required this.item});

  final FileItem item;

  @override
  ConsumerState<_ShareDialog> createState() => _ShareDialogState();
}

class _ShareDialogState extends ConsumerState<_ShareDialog> {
  bool _busy = true;
  String? _url;
  String? _error;

  @override
  void initState() {
    super.initState();
    _create();
  }

  Future<void> _create() async {
    final api = ref.read(apiClientProvider);
    final res = await api.createShare(widget.item.path, 0);
    if (!mounted) return;
    if (!res.ok || res.data == null) {
      setState(() {
        _busy = false;
        _error = res.error ?? 'Unknown error';
      });
      return;
    }
    setState(() {
      _busy = false;
      _url = api.shareUrl(res.data!.token);
    });
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: Text(
        'Share ${widget.item.name}',
        maxLines: 1,
        overflow: TextOverflow.ellipsis,
      ),
      content: _busy
          ? const SizedBox(
              width: 40,
              height: 40,
              child: Center(child: CircularProgressIndicator()),
            )
          : _error != null
              ? Text(_error!, style: const TextStyle(color: SfsColors.danger))
              : SelectableText(
                  _url!,
                  style: const TextStyle(
                    color: SfsColors.accent,
                    fontSize: 14,
                    fontWeight: FontWeight.w600,
                  ),
                ),
      actions: _busy
          ? null
          : _error != null
              ? [
                  TextButton(
                    onPressed: () => Navigator.pop(context),
                    child: const Text('Close'),
                  ),
                ]
              : [
                  TextButton(
                    onPressed: () => Navigator.pop(context),
                    child: const Text('Close'),
                  ),
                  FilledButton(
                    style: FilledButton.styleFrom(
                      backgroundColor: SfsColors.accent,
                      foregroundColor: SfsColors.onAccent,
                    ),
                    onPressed: () async {
                      final url = _url;
                      Navigator.pop(context);
                      if (url != null) {
                        await SharePlus.instance
                            .share(ShareParams(text: url));
                      }
                    },
                    child: const Text('Share'),
                  ),
                ],
    );
  }
}
