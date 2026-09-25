import 'package:flutter_test/flutter_test.dart';
import 'package:simplefileshare/src/grpc/fileshare/v1/fileshare.pbgrpc.dart';
import 'package:simplefileshare/src/services/events_service.dart';

void main() {
  group('serverEventFromProto', () {
    test('maps a full event', () {
      final event = serverEventFromProto(
        EventMessage()
          ..type = 'file.uploaded'
          ..path = 'docs/a.txt'
          ..user = 'eslam'
          ..at = '2024-01-02T03:04:05Z',
      );
      expect(event.type, 'file.uploaded');
      expect(event.path, 'docs/a.txt');
      expect(event.user, 'eslam');
      expect(event.at, '2024-01-02T03:04:05Z');
    });

    test('maps empty fields to null', () {
      final event = serverEventFromProto(EventMessage()..type = 'ping');
      expect(event.type, 'ping');
      expect(event.path, null);
      expect(event.user, null);
      expect(event.at, null);
    });
  });
}
