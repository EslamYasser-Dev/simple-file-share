import 'package:flutter_test/flutter_test.dart';
import 'package:simplefileshare/src/services/events_service.dart';

void main() {
  group('parseSseChunk', () {
    test('parses a complete frame', () {
      final (events, rest) =
          parseSseChunk('data: {"type":"change","path":"a.txt"}\n\n');
      expect(events, hasLength(1));
      expect(events.first.type, 'change');
      expect(events.first.path, 'a.txt');
      expect(rest, '');
    });

    test('keeps a partial frame in the remainder', () {
      final (events, rest) =
          parseSseChunk('data: {"type":"change"}\n\ndata: {"ty');
      expect(events, hasLength(1));
      expect(events.first.type, 'change');
      expect(rest, 'data: {"ty');
    });

    test('parses multiple frames in one chunk', () {
      final (events, rest) = parseSseChunk(
        'data: {"type":"created"}\n\ndata: {"type":"deleted"}\n\n',
      );
      expect(events.map((e) => e.type), ['created', 'deleted']);
      expect(rest, '');
    });

    test('skips comment lines', () {
      final (events, _) = parseSseChunk(': ping\n\n');
      expect(events, isEmpty);
    });

    test('ignores invalid json', () {
      final (events, _) = parseSseChunk('data: notjson\n\n');
      expect(events, isEmpty);
    });

    test('ignores objects without a string type', () {
      final (events, _) = parseSseChunk('data: {"foo":1}\n\n');
      expect(events, isEmpty);
    });

    test('joins split data lines', () {
      final (events, _) =
          parseSseChunk('data: {"type":\ndata: "upload"}\n\n');
      expect(events, hasLength(1));
      expect(events.first.type, 'upload');
    });
  });
}
