import 'package:grpc/grpc.dart';

import '../config.dart';
import '../grpc/fileshare/v1/fileshare.pbgrpc.dart';

class GrpcConnection {
  GrpcConnection({GrpcTarget? target, ClientChannel? channel})
      : _channel = channel ?? _connect(target ?? grpcTarget);

  static ClientChannel _connect(GrpcTarget target) => ClientChannel(
        target.host,
        port: target.port,
        options: ChannelOptions(
          credentials: target.secure
              ? const ChannelCredentials.secure()
              : const ChannelCredentials.insecure(),
          keepAlive: const ClientKeepAliveOptions(
            pingInterval: Duration(seconds: 30),
            permitWithoutCalls: true,
          ),
        ),
      );

  final ClientChannel _channel;

  late final AuthServiceClient auth = AuthServiceClient(_channel);
  late final FileServiceClient files = FileServiceClient(_channel);
  late final ShareServiceClient shares = ShareServiceClient(_channel);
  late final EventsServiceClient events = EventsServiceClient(_channel);

  Future<void> shutdown() => _channel.shutdown();
}
