import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../format.dart';
import '../state/auth_controller.dart';
import '../theme.dart';

class AccountScreen extends ConsumerWidget {
  const AccountScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final auth = ref.watch(authControllerProvider);
    final user = auth.user;
    if (user == null) {
      return const Scaffold(
        backgroundColor: SfsColors.background,
        body: Center(child: CircularProgressIndicator()),
      );
    }

    final role = user.isAdmin
        ? 'Admin'
        : ((user.role != null && user.role!.isNotEmpty)
            ? user.role!
            : 'Member');
    final joined = (user.createdAt != null && user.createdAt!.isNotEmpty)
        ? ' · joined ${formatDate(user.createdAt!)}'
        : '';
    final quota = user.quotaBytes ?? 0;
    final used = user.size ?? 0;
    final files = user.files ?? 0;

    return Scaffold(
      backgroundColor: SfsColors.background,
      body: SafeArea(
        child: ListView(
          padding: const EdgeInsets.all(16),
          children: [
            const Text(
              'Account',
              style: TextStyle(
                color: SfsColors.text,
                fontSize: 20,
                fontWeight: FontWeight.w800,
              ),
            ),
            const SizedBox(height: 16),
            Container(
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: SfsColors.card,
                borderRadius: BorderRadius.circular(16),
                border: Border.all(color: SfsColors.border),
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    user.username,
                    style: const TextStyle(
                      color: SfsColors.text,
                      fontSize: 18,
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    '$role$joined',
                    style: const TextStyle(color: SfsColors.muted, fontSize: 13),
                  ),
                  const SizedBox(height: 16),
                  Row(
                    children: [
                      Expanded(
                        child: _Stat(
                          label: 'Files',
                          value: '$files',
                        ),
                      ),
                      Expanded(
                        child: _Stat(
                          label: 'Used',
                          value: formatBytes(used),
                        ),
                      ),
                      Expanded(
                        child: _Stat(
                          label: 'Quota',
                          value: quota > 0 ? formatBytes(quota) : 'Unlimited',
                        ),
                      ),
                    ],
                  ),
                  if (quota > 0) ...[
                    const SizedBox(height: 14),
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
                    const SizedBox(height: 6),
                    Text(
                      '${formatBytes(used)} of ${formatBytes(quota)} used',
                      style: const TextStyle(
                        color: SfsColors.muted,
                        fontSize: 12,
                      ),
                    ),
                  ],
                ],
              ),
            ),
            const SizedBox(height: 20),
            OutlinedButton(
              style: OutlinedButton.styleFrom(
                foregroundColor: SfsColors.danger,
                side: const BorderSide(color: SfsColors.dangerBorder),
                padding: const EdgeInsets.symmetric(vertical: 14),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(12),
                ),
              ),
              onPressed: () =>
                  unawaited(ref.read(authControllerProvider.notifier).signOut()),
              child: const Text(
                'Sign out',
                style: TextStyle(fontWeight: FontWeight.w700),
              ),
            ),
            const SizedBox(height: 24),
            const Center(
              child: Text(
                'FileShare',
                style: TextStyle(color: SfsColors.muted, fontSize: 12),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _Stat extends StatelessWidget {
  const _Stat({required this.label, required this.value});

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          label.toUpperCase(),
          style: const TextStyle(
            color: SfsColors.muted,
            fontSize: 10,
            fontWeight: FontWeight.w700,
            letterSpacing: 1,
          ),
        ),
        const SizedBox(height: 2),
        Text(
          value,
          style: const TextStyle(
            color: SfsColors.text,
            fontSize: 16,
            fontWeight: FontWeight.w700,
          ),
        ),
      ],
    );
  }
}
