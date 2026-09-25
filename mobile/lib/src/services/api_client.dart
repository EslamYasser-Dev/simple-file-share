import 'dart:io';
import 'dart:typed_data';

import 'package:fixnum/fixnum.dart';
import 'package:grpc/grpc.dart';
import 'package:path_provider/path_provider.dart';

import '../config.dart';
import '../format.dart';
import '../grpc/fileshare/v1/fileshare.pbgrpc.dart';
import '../models.dart';
import 'grpc_connection.dart';
import 'token_store.dart';

class DownloadResult {
  const DownloadResult({required this.file, required this.name});

  final File file;
  final String name;
}

class ApiClient {
  ApiClient({GrpcConnection? connection, TokenStore? tokens})
      : _conn = connection ?? GrpcConnection(),
        _tokens = tokens ?? TokenStore();

  final GrpcConnection _conn;
  final TokenStore _tokens;

  void Function()? onUnauthorized;

  static const Duration _unaryTimeout = Duration(seconds: 30);
  static const int _uploadChunkSize = 64 * 1024;

  String buildUrl(String endpoint, [Map<String, Object?>? params]) {
    var uri = Uri.parse('$apiBaseUrl$endpoint');
    if (params != null && params.isNotEmpty) {
      uri = uri.replace(
        queryParameters: {
          ...uri.queryParameters,
          ...params.map((k, v) => MapEntry(k, v.toString())),
        },
      );
    }
    return uri.toString();
  }

  Future<void> _handle401() async {
    await _tokens.clear();
    onUnauthorized?.call();
  }

  Future<CallOptions> _options({
    bool authenticated = true,
    bool unary = true,
  }) async {
    final metadata = <String, String>{};
    if (authenticated) {
      final token = await _tokens.get();
      if (token != null && token.isNotEmpty) {
        metadata['authorization'] = 'Bearer $token';
      }
    }
    return CallOptions(
      metadata: metadata,
      timeout: unary ? _unaryTimeout : null,
    );
  }

  String _messageFrom(Object error) {
    if (error is GrpcError) {
      final message = error.message?.trim();
      if (message != null && message.isNotEmpty) return message;
      return 'Request failed (${error.code})';
    }
    return error.toString();
  }

  Future<ApiResult<T>> _fromError<T>(Object error) async {
    if (error is GrpcError && error.code == StatusCode.unauthenticated) {
      await _handle401();
      return ApiResult<T>(error: 'Not authorized', unauthorized: true);
    }
    return ApiResult<T>(error: _messageFrom(error));
  }

  Future<ApiResult<T>> _guard<T>(
    Future<T> Function(CallOptions options) run, {
    bool authenticated = true,
    bool unary = true,
  }) async {
    try {
      final options = await _options(authenticated: authenticated, unary: unary);
      return ApiResult<T>(data: await run(options));
    } catch (error) {
      return _fromError<T>(error);
    }
  }

  Future<ApiResult<void>> _guardVoid(
    Future<Object?> Function(CallOptions options) run, {
    bool authenticated = true,
    bool unary = true,
  }) async {
    try {
      final options = await _options(authenticated: authenticated, unary: unary);
      await run(options);
      return const ApiResult<void>();
    } catch (error) {
      return _fromError<void>(error);
    }
  }

  Future<ApiResult<TokenResponse>> login(
    String username,
    String password,
  ) async {
    final result = await _guard(
      (options) => _conn.auth.login(
        LoginRequest()
          ..username = username
          ..password = password,
        options: options,
      ),
      authenticated: false,
    );
    final response = result.data;
    if (response == null) {
      return ApiResult<TokenResponse>(
        error: result.error,
        unauthorized: result.unauthorized,
      );
    }
    if (response.accessToken.isNotEmpty) {
      await _tokens.set(response.accessToken);
    }
    return ApiResult<TokenResponse>(
      data: TokenResponse(
        accessToken: response.accessToken,
        tokenType: response.tokenType,
        expiresIn: response.expiresIn.toInt(),
      ),
    );
  }

  Future<ApiResult<void>> revoke() async {
    try {
      final options = await _options();
      await _conn.auth
          .logout(LogoutRequest(), options: options)
          .timeout(_unaryTimeout);
    } catch (_) {}
    await _tokens.clear();
    return const ApiResult<void>();
  }

  Future<ApiResult<AuthUser>> me() async {
    final result = await _guard(
      (options) => _conn.auth.me(MeRequest(), options: options),
    );
    final user = result.data;
    if (user == null) {
      return ApiResult<AuthUser>(
        error: result.error,
        unauthorized: result.unauthorized,
      );
    }
    return ApiResult<AuthUser>(data: _userFromProto(user));
  }

  Future<ApiResult<List<FileItem>>> listFiles([String path = '']) =>
      _guard((options) async {
        final response = await _conn.files.listFiles(
          ListFilesRequest()..path = path,
          options: options,
        );
        return response.files.map(_fileFromProto).toList();
      });

  Future<ApiResult<void>> createDirectory(String path) => _guardVoid(
        (options) => _conn.files.createDirectory(
          CreateDirectoryRequest()..path = path,
          options: options,
        ),
      );

  Future<ApiResult<void>> deletePath(String path) => _guardVoid(
        (options) => _conn.files.deletePath(
          DeletePathRequest()..path = path,
          options: options,
        ),
      );

  Future<ApiResult<ShareItem>> createShare(
    String path,
    int expiresInSeconds,
  ) =>
      _guard((options) async {
        final share = await _conn.shares.createShare(
          CreateShareRequest()
            ..path = path
            ..expiresInSeconds = Int64(expiresInSeconds),
          options: options,
        );
        return _shareFromProto(share);
      });

  Future<ApiResult<List<ShareItem>>> listShares() => _guard((options) async {
        final response = await _conn.shares.listShares(
          ListSharesRequest(),
          options: options,
        );
        return response.shares.map(_shareFromProto).toList();
      });

  Future<ApiResult<void>> revokeShare(String token) => _guardVoid(
        (options) => _conn.shares.revokeShare(
          RevokeShareRequest()..token = token,
          options: options,
        ),
      );

  String shareUrl(String token) => buildUrl('/api/share/$token');

  Future<ApiResult<void>> uploadFile({
    required String fileName,
    required int size,
    required String dirPath,
    File? file,
    Uint8List? bytes,
    void Function(int loaded, int total)? onProgress,
  }) =>
      _guardVoid(
        (options) => _conn.files.uploadFile(
          _uploadRequests(
            dirPath: dirPath,
            fileName: fileName,
            size: size,
            file: file,
            bytes: bytes,
            onProgress: onProgress,
          ),
          options: options,
        ),
        unary: false,
      );

  Stream<UploadRequest> _uploadRequests({
    required String dirPath,
    required String fileName,
    required int size,
    File? file,
    Uint8List? bytes,
    void Function(int loaded, int total)? onProgress,
  }) async* {
    yield UploadRequest()
      ..metadata = (UploadMetadata()
        ..path = dirPath
        ..filename = fileName);
    if (file == null && bytes == null) {
      throw StateError('No file data');
    }
    final source = file != null
        ? file.openRead()
        : Stream<List<int>>.value(bytes!);
    var buffer = <int>[];
    var sent = 0;
    await for (final piece in source) {
      buffer.addAll(piece);
      while (buffer.length >= _uploadChunkSize) {
        final chunk = Uint8List.fromList(
          buffer.sublist(0, _uploadChunkSize),
        );
        buffer = buffer.sublist(_uploadChunkSize);
        yield UploadRequest()..chunk = chunk;
        sent += chunk.length;
        onProgress?.call(sent, size);
      }
    }
    if (buffer.isNotEmpty) {
      final chunk = Uint8List.fromList(buffer);
      yield UploadRequest()..chunk = chunk;
      sent += chunk.length;
    }
    onProgress?.call(sent, size);
  }

  Future<ApiResult<DownloadResult>> download(String path) async {
    var savePath = '';
    var name = baseName(path);
    try {
      final dir = await getTemporaryDirectory();
      final suggested = baseName(path);
      final stamp = DateTime.now().millisecondsSinceEpoch;
      savePath = '${dir.path}/$stamp-$suggested';
      var first = true;
      final sink = File(savePath).openWrite();
      try {
        final chunks = _conn.files.downloadFile(
          DownloadFileRequest()..path = path,
          options: await _options(unary: false),
        );
        await for (final chunk in chunks) {
          if (first) {
            first = false;
            if (chunk.filename.isNotEmpty) name = chunk.filename;
          }
          if (chunk.data.isNotEmpty) sink.add(chunk.data);
        }
        await sink.flush();
      } finally {
        await sink.close();
      }
      var file = File(savePath);
      if (name != suggested) {
        final target = File(
          '${dir.path}/${DateTime.now().millisecondsSinceEpoch}-$name',
        );
        file = await file.rename(target.path);
      }
      return ApiResult(data: DownloadResult(file: file, name: name));
    } catch (error) {
      if (savePath.isNotEmpty) {
        try {
          await File(savePath).delete();
        } catch (_) {}
      }
      return _fromError<DownloadResult>(error);
    }
  }

  AuthUser _userFromProto(User user) => AuthUser(
        username: user.username,
        isAdmin: user.isAdmin,
        role: user.role.isEmpty ? null : user.role,
        enabled: user.enabled,
        permissions: user.permissions.toList(),
        createdAt: user.createdAt.isEmpty ? null : user.createdAt,
        quotaBytes: user.quotaBytes.toInt(),
        size: user.size.toInt(),
        files: user.files.toInt(),
      );

  FileItem _fileFromProto(FileInfo info) => FileItem(
        name: info.name,
        path: info.path,
        size: info.size.toInt(),
        isDir: info.isDir,
        modified: info.hasModified()
            ? info.modified.toDateTime().toUtc().toIso8601String()
            : '',
      );

  ShareItem _shareFromProto(Share share) => ShareItem(
        token: share.token,
        path: share.path,
        name: share.name,
        owner: share.owner,
        createdAt: share.createdAt,
        expiresAt: share.expiresAt,
      );
}
