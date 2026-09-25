import 'dart:async';

import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../models.dart';
import '../services/api_client.dart';
import '../services/events_service.dart';
import '../services/token_store.dart';

enum AuthStatus { loading, signedOut, signedIn }

class AuthState {
  const AuthState(this.status, [this.user]);

  final AuthStatus status;
  final AuthUser? user;
}

final tokenStoreProvider = Provider<TokenStore>((ref) => TokenStore());

final apiClientProvider = Provider<ApiClient>((ref) => ApiClient());

final eventsServiceProvider = Provider<EventsService>((ref) {
  final service = EventsService();
  ref.onDispose(service.dispose);
  return service;
});

class AuthController extends Notifier<AuthState> {
  @override
  AuthState build() {
    final api = ref.read(apiClientProvider);
    api.onUnauthorized = () {
      ref.read(eventsServiceProvider).stop();
      state = const AuthState(AuthStatus.signedOut);
    };
    unawaited(_restore());
    return const AuthState(AuthStatus.loading);
  }

  Future<void> _restore() async {
    final token = await ref.read(tokenStoreProvider).get();
    if (token == null) {
      state = const AuthState(AuthStatus.signedOut);
      return;
    }
    final res = await ref.read(apiClientProvider).me();
    if (res.data != null) {
      state = AuthState(AuthStatus.signedIn, res.data);
      ref.read(eventsServiceProvider).start();
    } else {
      state = const AuthState(AuthStatus.signedOut);
    }
  }

  Future<String?> signIn(String username, String password) async {
    final api = ref.read(apiClientProvider);
    final login = await api.login(username, password);
    if (!login.ok || login.data == null) {
      return login.error ?? 'Sign-in failed';
    }
    final me = await api.me();
    if (!me.ok || me.data == null) {
      return me.error ?? 'Unable to load account';
    }
    state = AuthState(AuthStatus.signedIn, me.data);
    ref.read(eventsServiceProvider).start();
    return null;
  }

  Future<void> signOut() async {
    ref.read(eventsServiceProvider).stop();
    await ref.read(apiClientProvider).revoke();
    state = const AuthState(AuthStatus.signedOut);
  }

  Future<void> refreshUser() async {
    final res = await ref.read(apiClientProvider).me();
    if (res.data != null && state.status == AuthStatus.signedIn) {
      state = AuthState(AuthStatus.signedIn, res.data);
    }
  }
}

final authControllerProvider =
    NotifierProvider<AuthController, AuthState>(AuthController.new);
