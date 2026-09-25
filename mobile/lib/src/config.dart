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

class GrpcTarget {
  const GrpcTarget({
    required this.host,
    required this.port,
    required this.secure,
  });

  final String host;
  final int port;
  final bool secure;
}

GrpcTarget grpcTargetFrom(String baseUrl) {
  final uri = Uri.parse(baseUrl.replaceAll(RegExp(r'/+$'), ''));
  final secure = uri.scheme != 'http';
  return GrpcTarget(
    host: uri.host,
    port: uri.hasPort ? uri.port : (secure ? 443 : 80),
    secure: secure,
  );
}

GrpcTarget get grpcTarget => grpcTargetFrom(apiBaseUrl);
