import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'src/screens/auth_gate.dart';
import 'src/theme.dart';

void main() {
  runApp(const ProviderScope(child: FileShareApp()));
}

class FileShareApp extends StatelessWidget {
  const FileShareApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'File Share',
      debugShowCheckedModeBanner: false,
      theme: buildAppTheme(),
      home: const AuthGate(),
    );
  }
}
