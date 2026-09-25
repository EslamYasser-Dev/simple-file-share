import 'dart:io' show Platform;

const String _envApiUrl = String.fromEnvironment('API_BASE_URL');

String get apiBaseUrl {
  final raw = _envApiUrl.trim();
  final fallback = Platform.isAndroid
      ? 'http://10.0.2.2:3000'
      : 'http://localhost:3000';
  final base = raw.isEmpty ? fallback : raw;
  return base.replaceAll(RegExp(r'/+$'), '');
}
