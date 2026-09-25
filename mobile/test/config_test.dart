import 'package:flutter_test/flutter_test.dart';
import 'package:simplefileshare/src/config.dart';

void main() {
  group('grpcTargetFrom', () {
    test('https defaults to port 443 and TLS', () {
      final target = grpcTargetFrom('https://shares.up.railway.app');
      expect(target.host, 'shares.up.railway.app');
      expect(target.port, 443);
      expect(target.secure, true);
    });

    test('http keeps an explicit port and is insecure', () {
      final target = grpcTargetFrom('http://localhost:3000');
      expect(target.host, 'localhost');
      expect(target.port, 3000);
      expect(target.secure, false);
    });

    test('http without port defaults to 80', () {
      final target = grpcTargetFrom('http://example.test');
      expect(target.port, 80);
      expect(target.secure, false);
    });

    test('trailing slashes are ignored', () {
      final target = grpcTargetFrom('https://example.test/');
      expect(target.host, 'example.test');
      expect(target.port, 443);
    });

    test('host and port overrides target the TCP proxy', () {
      final target = grpcTargetFrom(
        'https://shares.up.railway.app',
        host: 'shuttle.proxy.rlwy.net',
        port: 15140,
      );
      expect(target.host, 'shuttle.proxy.rlwy.net');
      expect(target.port, 15140);
      expect(target.secure, true);
    });

    test('empty host override falls back to the base URL host', () {
      final target = grpcTargetFrom('http://localhost:3000', host: '');
      expect(target.host, 'localhost');
      expect(target.port, 3000);
      expect(target.secure, false);
    });

    test('port override without explicit base port keeps TLS default', () {
      final target = grpcTargetFrom('https://example.test', port: 50051);
      expect(target.host, 'example.test');
      expect(target.port, 50051);
      expect(target.secure, true);
    });
  });
}
