import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:share_plus/share_plus.dart';

import '../format.dart';
import '../models.dart';
import '../state/auth_controller.dart';
import '../theme.dart';

class SharesScreen extends ConsumerStatefulWidget {
  const SharesScreen({super.key});

  @override
  ConsumerState<SharesScreen> createState() => _SharesScreenState();
}

class _SharesScreenState extends ConsumerState<SharesScreen> {
  List<ShareItem> _items = const [];
  bool _loading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    unawaited(_load());
  }

  Future<void> _load() async {
    final res = await ref.read(apiClientProvider).listShares();
    if (!mounted) return;
    setState(() {
      if (res.error != null) {
        _error = res.error;
        _items = const [];
      } else {
        _error = null;
        _items = res.data ?? const [];
      }
      _loading = false;
    });
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

  Future<void> _revoke(ShareItem item) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Revoke link'),
        content: Text('Revoke the share for ${item.name}?'),
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
            child: const Text('Revoke'),
          ),
        ],
      ),
    );
    if (confirmed != true || !mounted) return;
    final res = await ref.read(apiClientProvider).revokeShare(item.token);
    if (!mounted) return;
    if (!res.ok) {
      _showError('Revoke failed', res.error ?? 'Unknown error');
      return;
    }
    unawaited(_load());
  }

  Widget _buildBody() {
    if (_loading) {
      return const Center(child: CircularProgressIndicator());
    }
    if (_error != null) {
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
                setState(() => _loading = true);
                unawaited(_load());
              },
              child: const Text('Retry'),
            ),
          ],
        ),
      );
    }
    if (_items.isEmpty) {
      return const Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text('No shared links yet', style: TextStyle(color: SfsColors.muted, fontSize: 15)),
            SizedBox(height: 6),
            Text(
              'Create one from a file’s menu.',
              style: TextStyle(color: SfsColors.muted, fontSize: 12),
            ),
          ],
        ),
      );
    }
    return RefreshIndicator(
      color: SfsColors.accent,
      backgroundColor: SfsColors.card,
      onRefresh: _load,
      child: ListView.builder(
        physics: const AlwaysScrollableScrollPhysics(),
        padding: const EdgeInsets.fromLTRB(12, 4, 12, 12),
        itemCount: _items.length,
        itemBuilder: (context, index) {
          final item = _items[index];
          return Container(
            margin: const EdgeInsets.only(bottom: 10),
            padding: const EdgeInsets.all(14),
            decoration: BoxDecoration(
              color: SfsColors.card,
              borderRadius: BorderRadius.circular(14),
              border: Border.all(color: SfsColors.border),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Expanded(
                      child: Text(
                        item.name,
                        style: const TextStyle(
                          color: SfsColors.text,
                          fontWeight: FontWeight.w700,
                          fontSize: 15,
                        ),
                      ),
                    ),
                    Text(
                      '/${item.path}',
                      style: const TextStyle(
                        color: SfsColors.muted,
                        fontSize: 12,
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 6),
                if (item.createdAt.isNotEmpty)
                  Text(
                    'by ${item.owner} · created ${formatDate(item.createdAt)}',
                    style: const TextStyle(color: SfsColors.muted, fontSize: 12),
                  )
                else if (item.owner.isNotEmpty)
                  Text(
                    'by ${item.owner}',
                    style: const TextStyle(color: SfsColors.muted, fontSize: 12),
                  ),
                if (item.expiresAt.isNotEmpty)
                  Text(
                    'Expires ${formatDate(item.expiresAt)}',
                    style: const TextStyle(color: SfsColors.warn, fontSize: 12),
                  ),
                const SizedBox(height: 8),
                Row(
                  mainAxisAlignment: MainAxisAlignment.end,
                  children: [
                    OutlinedButton.icon(
                      style: OutlinedButton.styleFrom(
                        foregroundColor: SfsColors.text,
                        side: const BorderSide(color: SfsColors.border),
                        padding: const EdgeInsets.symmetric(
                          horizontal: 12,
                          vertical: 6,
                        ),
                      ),
                      onPressed: () => SharePlus.instance.share(
                        ShareParams(text: ref.read(apiClientProvider).shareUrl(item.token)),
                      ),
                      icon: const Icon(Icons.share, size: 16),
                      label: const Text('Share'),
                    ),
                    const SizedBox(width: 8),
                    OutlinedButton.icon(
                      style: OutlinedButton.styleFrom(
                        foregroundColor: SfsColors.danger,
                        side: const BorderSide(color: SfsColors.dangerBorder),
                        padding: const EdgeInsets.symmetric(
                          horizontal: 12,
                          vertical: 6,
                        ),
                      ),
                      onPressed: () => unawaited(_revoke(item)),
                      icon: const Icon(Icons.link_off, size: 16),
                      label: const Text('Revoke'),
                    ),
                  ],
                ),
              ],
            ),
          );
        },
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: SfsColors.background,
      body: SafeArea(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Padding(
              padding: EdgeInsets.fromLTRB(16, 16, 16, 8),
              child: Text(
                'Shared links',
                style: TextStyle(
                  color: SfsColors.text,
                  fontSize: 20,
                  fontWeight: FontWeight.w800,
                ),
              ),
            ),
            Expanded(child: _buildBody()),
          ],
        ),
      ),
    );
  }
}
