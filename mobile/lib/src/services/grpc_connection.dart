import 'dart:io' show HttpClient;

import 'package:grpc/grpc.dart';

import '../config.dart';
import '../grpc/fileshare/v1/fileshare.pbgrpc.dart';

class GrpcConnection {
  GrpcConnection({GrpcTarget? target, ClientChannel? channel})
      : _target = target ?? grpcTarget,
        _providedChannel = channel;

  final GrpcTarget _target;
  final ClientChannel? _providedChannel;

  Future<ClientChannel>? _pendingChannel;

  Future<ClientChannel> get _channel {
    final provided = _providedChannel;
    if (provided != null) return Future<ClientChannel>.value(provided);
    return _pendingChannel ??= _openChannel();
  }

  Future<ClientChannel> _openChannel() async {
    final pinnedCert = _target.secure ? await _fetchCertPem() : null;
    return ClientChannel(
      _target.host,
      port: _target.port,
      options: ChannelOptions(
        credentials: pinnedCert != null
            ? ChannelCredentials.secure(certificates: pinnedCert)
            : _target.secure
                ? const ChannelCredentials.secure()
                : const ChannelCredentials.insecure(),
        keepAlive: const ClientKeepAliveOptions(
          pingInterval: Duration(seconds: 30),
          permitWithoutCalls: true,
        ),
      ),
    );
  }

  /// Fetches the server's gRPC certificate over the trusted HTTPS API so the
  /// channel can pin it. Returns null (fall back to the system trust store)
  /// on any failure, including the endpoint being absent.
  Future<List<int>?> _fetchCertPem() async {
    try {
      final client = HttpClient();
      try {
        final request = await client
            .getUrl(Uri.parse('$apiBaseUrl/api/grpc/cert'))
            .timeout(const Duration(seconds: 10));
        final response =
            await request.close().timeout(const Duration(seconds: 10));
        if (response.statusCode != 200) return null;
        final bytes = await response.expand((chunk) => chunk).toList();
        return bytes.isEmpty ? null : bytes;
      } finally {
        client.close(force: true);
      }
    } catch (_) {
      return null;
    }
  }

  late final Future<AuthServiceClient> auth =
      _channel.then((channel) => AuthServiceClient(channel));
  late final Future<FileServiceClient> files =
      _channel.then((channel) => FileServiceClient(channel));
  late final Future<ShareServiceClient> shares =
      _channel.then((channel) => ShareServiceClient(channel));
  late final Future<EventsServiceClient> events =
      _channel.then((channel) => EventsServiceClient(channel));

  Future<void> shutdown() async {
    final provided = _providedChannel;
    if (provided != null) {
      await provided.shutdown();
      return;
    }
    final pending = _pendingChannel;
    if (pending == null) return;
    await (await pending).shutdown();
  }
}
