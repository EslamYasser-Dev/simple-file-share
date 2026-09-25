import 'dart:typed_data';

import 'package:flutter_test/flutter_test.dart';
import 'package:simplefileshare/src/services/api_client.dart';
import 'package:simplefileshare/src/services/events_service.dart';
import 'package:simplefileshare/src/services/grpc_connection.dart';
import 'package:simplefileshare/src/services/token_store.dart';

const String _flag = String.fromEnvironment('RUN_GRPC_IT');
final bool _enabled = _flag == '1' || _flag == 'true';

class _MemTokenStore extends TokenStore {
  String? _token;

  @override
  Future<String?> get() async => _token;

  @override
  Future<void> set(String token) async => _token = token;

  @override
  Future<void> clear() async => _token = null;
}

void main() {
  test(
    'grpc client flow against a live backend',
    () async {
      final tokens = _MemTokenStore();
      final connection = GrpcConnection();
      final api = ApiClient(connection: connection, tokens: tokens);
      final events = EventsService(connection: connection, tokens: tokens);
      var eventsStarted = false;

      try {
        await api.deletePath('grpc-it');

        final login = await api.login('admin', 'admin');
        expect(login.ok, isTrue, reason: login.error);
        expect(login.data!.accessToken, isNotEmpty);

        final me = await api.me();
        expect(me.ok, isTrue, reason: me.error);
        expect(me.data!.username, 'admin');

        final firstEvent =
            events.stream.first.timeout(const Duration(seconds: 5));
        events.start();
        eventsStarted = true;

        final mkdir = await api.createDirectory('grpc-it');
        expect(mkdir.ok, isTrue, reason: mkdir.error);

        final event = await firstEvent;
        expect(event.type, 'mkdir');
        expect(event.path, 'grpc-it');
        events.stop();
        eventsStarted = false;

        const content = 'hello from grpc';
        final upload = await api.uploadFile(
          fileName: 'hello.txt',
          size: content.length,
          dirPath: 'grpc-it',
          bytes: Uint8List.fromList(content.codeUnits),
        );
        expect(upload.ok, isTrue, reason: upload.error);

        final files = await api.listFiles('grpc-it');
        expect(files.ok, isTrue, reason: files.error);
        expect(
          files.data!.map((f) => f.name),
          contains('hello.txt'),
        );
        expect(files.data!.first.size, content.length);

        final share = await api.createShare('grpc-it/hello.txt', 0);
        expect(share.ok, isTrue, reason: share.error);
        expect(share.data!.token, isNotEmpty);

        final shares = await api.listShares();
        expect(shares.ok, isTrue, reason: shares.error);
        expect(
          shares.data!.map((s) => s.token),
          contains(share.data!.token),
        );

        final revokeShare = await api.revokeShare(share.data!.token);
        expect(revokeShare.ok, isTrue, reason: revokeShare.error);

        final sharesAfter = await api.listShares();
        expect(sharesAfter.ok, isTrue, reason: sharesAfter.error);
        expect(
          sharesAfter.data!.map((s) => s.token),
          isNot(contains(share.data!.token)),
        );

        final remove = await api.deletePath('grpc-it');
        expect(remove.ok, isTrue, reason: remove.error);

        final logout = await api.revoke();
        expect(logout.ok, isTrue, reason: logout.error);

        final meAfter = await api.me();
        expect(meAfter.unauthorized, isTrue);
        expect(meAfter.data, isNull);
      } finally {
        if (eventsStarted) events.stop();
        events.dispose();
        await connection.shutdown();
      }
    },
    timeout: const Timeout(Duration(minutes: 2)),
    skip: _enabled
        ? false
        : 'set --dart-define=RUN_GRPC_IT=1 with a local backend on '
            'http://localhost:3000',
  );
}
