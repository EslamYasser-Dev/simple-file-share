import 'package:flutter_test/flutter_test.dart';
import 'package:simplefileshare/src/models.dart';

void main() {
  group('FileItem.fromJson', () {
    test('parses a full object', () {
      final item = FileItem.fromJson({
        'name': 'a.txt',
        'path': 'docs/a.txt',
        'size': 42.0,
        'isDir': false,
        'modified': '2024-01-02T03:04:00Z',
        'mimeType': 'text/plain',
        'version': 3.0,
      });
      expect(item.name, 'a.txt');
      expect(item.path, 'docs/a.txt');
      expect(item.size, 42);
      expect(item.isDir, false);
      expect(item.modified, '2024-01-02T03:04:00Z');
      expect(item.mimeType, 'text/plain');
      expect(item.version, 3);
    });

    test('applies defaults', () {
      final item = FileItem.fromJson(const {});
      expect(item.name, '');
      expect(item.path, '');
      expect(item.size, 0);
      expect(item.isDir, false);
      expect(item.mimeType, null);
      expect(item.version, null);
    });
  });

  group('AuthUser.fromJson', () {
    test('parses user fields', () {
      final user = AuthUser.fromJson({
        'username': 'eslam',
        'isAdmin': true,
        'role': 'admin',
        'enabled': true,
        'permissions': ['files.read'],
        'createdAt': '2024-01-01T00:00:00Z',
        'quotaBytes': 100.0,
        'size': 50.0,
        'files': 7.0,
      });
      expect(user.username, 'eslam');
      expect(user.isAdmin, true);
      expect(user.role, 'admin');
      expect(user.enabled, true);
      expect(user.permissions, ['files.read']);
      expect(user.quotaBytes, 100);
      expect(user.size, 50);
      expect(user.files, 7);
    });

    test('applies defaults', () {
      final user = AuthUser.fromJson(const {});
      expect(user.username, '');
      expect(user.isAdmin, false);
      expect(user.quotaBytes, null);
    });
  });

  group('ShareItem.fromJson', () {
    test('parses a share', () {
      final share = ShareItem.fromJson({
        'token': 'tok',
        'path': 'a.txt',
        'name': 'a.txt',
        'owner': 'eslam',
        'createdAt': 'c',
        'expiresAt': 'e',
      });
      expect(share.token, 'tok');
      expect(share.expiresAt, 'e');
    });
  });

  group('TokenResponse.fromJson', () {
    test('parses a token', () {
      final token = TokenResponse.fromJson({
        'accessToken': 'jwt',
        'tokenType': 'Bearer',
        'expiresIn': 900.0,
      });
      expect(token.accessToken, 'jwt');
      expect(token.expiresIn, 900);
    });
  });

  group('ApiResult', () {
    test('ok requires no error', () {
      expect(const ApiResult<int>(data: 1).ok, true);
      expect(const ApiResult<int>(error: 'x').ok, false);
      expect(const ApiResult<int>(unauthorized: true).ok, false);
    });
  });
}
