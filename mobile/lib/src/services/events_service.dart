import 'dart:async';
import 'dart:math';

import 'package:grpc/grpc.dart';

import '../grpc/fileshare/v1/fileshare.pbgrpc.dart';
import '../models.dart';
import 'grpc_connection.dart';
import 'token_store.dart';

ServerEvent serverEventFromProto(EventMessage message) => ServerEvent(
      type: message.type,
      path: message.path.isEmpty ? null : message.path,
      user: message.user.isEmpty ? null : message.user,
      at: message.at.isEmpty ? null : message.at,
    );

class EventsService {
  EventsService({GrpcConnection? connection, TokenStore? tokens})
      : _conn = connection ?? GrpcConnection(),
        _tokens = tokens ?? TokenStore();

  final GrpcConnection _conn;
  final TokenStore _tokens;
  final StreamController<ServerEvent> _controller =
      StreamController<ServerEvent>.broadcast();

  bool _running = false;
  int _retryDelay = 1000;
  int _generation = 0;
  Timer? _retryTimer;
  ResponseStream<EventMessage>? _subscription;

  static const int _maxRetryMs = 30000;

  Stream<ServerEvent> get stream => _controller.stream;

  void start() {
    if (_running) return;
    _running = true;
    _retryDelay = 1000;
    unawaited(_open(_generation));
  }

  void stop() {
    _running = false;
    _generation++;
    _retryTimer?.cancel();
    _retryTimer = null;
    final subscription = _subscription;
    _subscription = null;
    if (subscription != null) {
      unawaited(subscription.cancel());
    }
  }

  void _emit(ServerEvent event) {
    if (_controller.isClosed) return;
    _controller.add(event);
  }

  void _scheduleRetry(int generation) {
    if (!_running || generation != _generation) return;
    _retryTimer?.cancel();
    _retryTimer = Timer(Duration(milliseconds: _retryDelay), () {
      _retryTimer = null;
      unawaited(_open(generation));
    });
    _retryDelay = min(_retryDelay * 2, _maxRetryMs);
  }

  Future<void> _open(int generation) async {
    if (!_running || generation != _generation) return;
    final token = await _tokens.get();
    if (!_running || generation != _generation) return;
    if (token == null || token.isEmpty) {
      _scheduleRetry(generation);
      return;
    }
    ResponseStream<EventMessage>? subscription;
    try {
      subscription = _conn.events.subscribe(
        SubscribeRequest(),
        options: CallOptions(metadata: {'authorization': 'Bearer $token'}),
      );
      if (generation != _generation) {
        unawaited(subscription.cancel());
        return;
      }
      _subscription = subscription;
      await for (final message in subscription) {
        if (!_running || generation != _generation) break;
        _retryDelay = 1000;
        _emit(serverEventFromProto(message));
      }
    } catch (_) {
    } finally {
      if (subscription != null && identical(_subscription, subscription)) {
        _subscription = null;
      }
    }
    _scheduleRetry(generation);
  }

  void dispose() {
    stop();
    _controller.close();
  }
}
