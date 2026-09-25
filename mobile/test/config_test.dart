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
  });
}
