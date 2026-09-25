import 'dart:async';
import 'dart:convert';
import 'dart:math';

import 'package:dio/dio.dart';

import '../config.dart';
import '../models.dart';
import 'token_store.dart';

(List<ServerEvent>, String) parseSseChunk(String buffer) {
  final events = <ServerEvent>[];
  final parts = buffer.split('\n\n');
  final rest = parts.removeLast();
  for (final part in parts) {
    final lines = part.split('\n');
    var data = '';
    for (final line in lines) {
      if (line.startsWith(':')) continue;
      if (line.startsWith('data:')) {
        data += line.substring(5).trim();
      }
    }
    if (data.isEmpty) continue;
    try {
      final parsed = jsonDecode(data);
      if (parsed is Map<String, dynamic> && parsed['type'] is String) {
        events.add(ServerEvent.fromJson(parsed));
      }
    } catch (_) {}
  }
  return (events, rest);
}

class EventsService {
  EventsService({Dio? dio, TokenStore? tokens})
      : _tokens = tokens ?? TokenStore(),
        _dio = dio ?? Dio(BaseOptions(baseUrl: apiBaseUrl));

  final Dio _dio;
  final TokenStore _tokens;
  final StreamController<ServerEvent> _controller =
      StreamController<ServerEvent>.broadcast();

  bool _running = false;
  int _retryDelay = 1000;
  Timer? _retryTimer;
  CancelToken? _cancel;

  static const int _maxRetryMs = 30000;

  Stream<ServerEvent> get stream => _controller.stream;

  void start() {
    if (_running) return;
    _running = true;
    _retryDelay = 1000;
    unawaited(_open());
  }

  void stop() {
    _running = false;
    _retryTimer?.cancel();
    _retryTimer = null;
    _cancel?.cancel();
    _cancel = null;
  }

  void _emit(ServerEvent event) {
    if (_controller.isClosed) return;
    _controller.add(event);
  }

  void _scheduleRetry() {
    if (!_running) return;
    _retryTimer?.cancel();
    _retryTimer = Timer(Duration(milliseconds: _retryDelay), () {
      _retryTimer = null;
      unawaited(_open());
    });
    _retryDelay = min(_retryDelay * 2, _maxRetryMs);
  }

  Future<void> _open() async {
    if (!_running) return;
    final token = await _tokens.get();
    if (token == null) {
      _scheduleRetry();
      return;
    }
    _cancel = CancelToken();
    var buffer = '';
    try {
      final response = await _dio.get<ResponseBody>(
        '/api/events',
        options: Options(
          responseType: ResponseType.stream,
          headers: {
            'Authorization': 'Bearer $token',
            'Accept': 'text/event-stream',
          },
        ),
        cancelToken: _cancel,
      );
      final body = response.data;
      if (body == null) {
        _scheduleRetry();
        return;
      }
      _retryDelay = 1000;
      await for (final chunk
          in body.stream.cast<List<int>>().transform(utf8.decoder)) {
        if (!_running) break;
        buffer += chunk;
        final (events, rest) = parseSseChunk(buffer);
        buffer = rest;
        for (final event in events) {
          _emit(event);
        }
      }
    } catch (_) {
    } finally {
      _cancel = null;
      if (_running) _scheduleRetry();
    }
  }

  void dispose() {
    stop();
    _controller.close();
  }
}
