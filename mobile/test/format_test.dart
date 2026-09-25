import 'package:flutter_test/flutter_test.dart';
import 'package:simplefileshare/src/format.dart';

void main() {
  group('formatBytes', () {
    test('zero', () {
      expect(formatBytes(0), '0 B');
    });

    test('small values', () {
      expect(formatBytes(5), '5 B');
      expect(formatBytes(1023), '1023 B');
    });

    test('kilobytes', () {
      expect(formatBytes(1024), '1 KB');
      expect(formatBytes(1536), '1.5 KB');
    });

    test('megabytes and gigabytes', () {
      expect(formatBytes(1048576), '1 MB');
      expect(formatBytes(1572864), '1.5 MB');
      expect(formatBytes(1073741824), '1 GB');
    });

    test('caps at TB', () {
      expect(formatBytes(1099511627776), '1 TB');
      expect(formatBytes(1125899906842624), '1024 TB');
    });
  });

  group('formatDate', () {
    test('empty input', () {
      expect(formatDate(''), '—');
      expect(formatDate('not-a-date'), '—');
    });

    test('morning', () {
      expect(formatDate('2024-01-02 03:04'), '1/2/2024, 03:04 AM');
    });

    test('midnight and noon', () {
      expect(formatDate('2024-01-02 00:04'), '1/2/2024, 12:04 AM');
      expect(formatDate('2024-12-25 12:00'), '12/25/2024, 12:00 PM');
    });

    test('evening with padding', () {
      expect(formatDate('2024-11-05 17:09'), '11/5/2024, 05:09 PM');
    });
  });

  group('baseName', () {
    test('extracts last segment', () {
      expect(baseName('a/b/c.txt'), 'c.txt');
      expect(baseName('file'), 'file');
      expect(baseName('dir/'), 'dir');
      expect(baseName(''), '');
    });
  });

  group('parentPath', () {
    test('drops last segment', () {
      expect(parentPath('a/b/c'), 'a/b');
      expect(parentPath('a/b/'), 'a');
      expect(parentPath('a'), '');
      expect(parentPath(''), '');
    });
  });

  group('joinPath', () {
    test('joins with single slash', () {
      expect(joinPath('', 'x'), 'x');
      expect(joinPath('a/b', 'c'), 'a/b/c');
      expect(joinPath('a/b/', 'c'), 'a/b/c');
    });
  });
}
