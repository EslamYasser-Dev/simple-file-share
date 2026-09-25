import 'dart:convert';
import 'dart:io';
import 'dart:math' as math;
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:path_provider/path_provider.dart';

import '../config.dart';
import '../format.dart';
import '../models.dart';
import 'token_store.dart';

class DownloadResult {
  const DownloadResult({required this.file, required this.name});

  final File file;
  final String name;
}

class ApiClient {
  ApiClient({Dio? dio, TokenStore? tokens})
      : _tokens = tokens ?? TokenStore(),
        _dio = dio ??
            Dio(
              BaseOptions(
                baseUrl: apiBaseUrl,
                responseType: ResponseType.json,
                validateStatus: (s) => s != null && s >= 200 && s < 300,
              ),
            ) {
    _dio.interceptors.add(
      InterceptorsWrapper(
        onRequest: (options, handler) async {
          final token = await _tokens.get();
          if (token != null) {
            options.headers['Authorization'] = 'Bearer $token';
          }
          handler.next(options);
        },
      ),
    );
  }

  final Dio _dio;
  final TokenStore _tokens;

  void Function()? onUnauthorized;

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

  String _messageFrom(Object? body, int? status) {
    if (body is String && body.isNotEmpty) {
      try {
        final parsed = jsonDecode(body) as Map<String, dynamic>;
        final msg = (parsed['error'] ?? parsed['message']) as String?;
        if (msg != null && msg.isNotEmpty) return msg;
      } catch (_) {
        return body;
      }
      return body;
    }
    if (body is Map<String, dynamic>) {
      final msg = (body['error'] ?? body['message']) as String?;
      if (msg != null && msg.isNotEmpty) return msg;
    }
    return 'Request failed (${status ?? 0})';
  }

  bool _isNetworkError(DioException e) =>
      e.type == DioExceptionType.connectionError ||
      e.type == DioExceptionType.connectionTimeout ||
      e.type == DioExceptionType.sendTimeout ||
      e.type == DioExceptionType.receiveTimeout ||
      (e.type == DioExceptionType.unknown &&
          e.response?.statusCode == null);

  Future<ApiResult<T>> _guard<T>(
    Future<Response<dynamic>> Function() run, {
    T Function(Object? body)? parse,
  }) async {
    try {
      final response = await run();
      var body = response.data;
      if (body is String) {
        if (body.isEmpty) return ApiResult<T>(data: null);
        try {
          body = jsonDecode(body);
        } catch (_) {
          return ApiResult<T>(error: 'Failed to parse response');
        }
      }
      if (body == null) return ApiResult<T>(data: null);
      if (parse != null) {
        try {
          return ApiResult<T>(data: parse(body));
        } catch (_) {
          return ApiResult<T>(error: 'Failed to parse response');
        }
      }
      return ApiResult<T>(data: body as T);
    } on DioException catch (e) {
      final status = e.response?.statusCode;
      if (status == 401) {
        await _handle401();
        return ApiResult<T>(error: 'Not authorized', unauthorized: true);
      }
      if (_isNetworkError(e)) {
        return ApiResult<T>(error: e.message ?? 'Network request failed');
      }
      return ApiResult<T>(error: _messageFrom(e.response?.data, status));
    }
  }

  Future<ApiResult<TokenResponse>> login(
    String username,
    String password,
  ) async {
    final result = await _guard(
      () => _dio.post<dynamic>(
        '/api/auth/token',
        data: {'username': username, 'password': password},
      ),
      parse: (body) => TokenResponse.fromJson(body as Map<String, dynamic>),
    );
    final token = result.data?.accessToken;
    if (token != null && token.isNotEmpty) {
      await _tokens.set(token);
    }
    return result;
  }

  Future<ApiResult<void>> revoke() async {
    final result = await _guard<void>(
      () => _dio.post<dynamic>('/api/auth/revoke'),
    );
    await _tokens.clear();
    return result;
  }

  Future<ApiResult<AuthUser>> me() => _guard(
        () => _dio.get<dynamic>('/api/auth/me'),
        parse: (body) => AuthUser.fromJson(body as Map<String, dynamic>),
      );

  Future<ApiResult<List<FileItem>>> listFiles([String path = '']) async {
    return _guard(
      () => _dio.get<dynamic>(
        '/api/files',
        queryParameters: path.isEmpty ? null : {'path': path},
      ),
      parse: (body) => (body as List<dynamic>)
          .map((e) => FileItem.fromJson(e as Map<String, dynamic>))
          .toList(),
    );
  }

  Future<ApiResult<void>> createDirectory(String path) => _guard(
        () => _dio.post<dynamic>('/api/directories', data: {'path': path}),
      );

  Future<ApiResult<void>> deletePath(String path) => _guard(
        () => _dio.delete<dynamic>('/api/files', data: {'path': path}),
      );

  Future<ApiResult<ShareItem>> createShare(
    String path,
    int expiresInSeconds,
  ) =>
      _guard(
        () => _dio.post<dynamic>(
          '/api/shares',
          data: {'path': path, 'expiresInSeconds': expiresInSeconds},
        ),
        parse: (body) => ShareItem.fromJson(body as Map<String, dynamic>),
      );

  Future<ApiResult<List<ShareItem>>> listShares() => _guard(
        () => _dio.get<dynamic>('/api/shares'),
        parse: (body) => (body as List<dynamic>)
            .map((e) => ShareItem.fromJson(e as Map<String, dynamic>))
            .toList(),
      );

  Future<ApiResult<void>> revokeShare(String token) => _guard(
        () => _dio.delete<dynamic>('/api/shares', data: {'token': token}),
      );

  String shareUrl(String token) => buildUrl('/api/share/$token');

  Future<ApiResult<UploadSessionInfo>> _createUploadSession({
    required String path,
    required String fileName,
    required int size,
  }) =>
      _guard(
        () => _dio.post<dynamic>(
          '/api/uploads',
          data: {
            'path': path,
            'filename': fileName,
            'size': size,
            'fingerprint': '$path|$fileName|$size',
          },
        ),
        parse: (body) =>
            UploadSessionInfo.fromJson(body as Map<String, dynamic>),
      );

  Future<ApiResult<void>> uploadFile({
    required String fileName,
    required int size,
    required String dirPath,
    File? file,
    Uint8List? bytes,
    void Function(int loaded, int total)? onProgress,
  }) async {
    if (size > 4 * 1024 * 1024 && (file != null || bytes != null)) {
      return _uploadResumable(
        fileName: fileName,
        size: size,
        dirPath: dirPath,
        file: file,
        bytes: bytes,
        onProgress: onProgress,
      );
    }
    return _uploadMultipart(
      fileName: fileName,
      dirPath: dirPath,
      file: file,
      bytes: bytes,
      onProgress: onProgress,
    );
  }

  Future<ApiResult<void>> _uploadResumable({
    required String fileName,
    required int size,
    required String dirPath,
    File? file,
    Uint8List? bytes,
    void Function(int loaded, int total)? onProgress,
  }) async {
    final created = await _createUploadSession(
      path: dirPath,
      fileName: fileName,
      size: size,
    );
    final session = created.data;
    if (created.error != null || session == null) {
      return ApiResult<void>(error: created.error);
    }
    final chunkSize =
        session.chunkSize > 0 ? session.chunkSize : 4 * 1024 * 1024;
    var offset = session.offset.clamp(0, size).toInt();
    onProgress?.call(offset, size);

    while (offset < size) {
      final base = offset;
      final end = math.min(base + chunkSize, size);
      final Stream<List<int>> chunkStream = file != null
          ? file.openRead(base, end)
          : Stream<List<int>>.value(bytes!.sublist(base, end));
      try {
        final response = await _dio.patch<dynamic>(
          '/api/uploads/${session.id}',
          data: chunkStream,
          options: Options(
            headers: {'Content-Type': 'application/octet-stream'},
          ),
          onSendProgress: (loaded, _) => onProgress?.call(base + loaded, size),
        );
        final body = response.data;
        var next = end;
        if (body is Map<String, dynamic>) {
          final reported = (body['offset'] as num?)?.toInt();
          if (reported != null && reported > offset) {
            next = reported.clamp(0, size).toInt();
          }
        }
        offset = next;
        onProgress?.call(offset, size);
      } on DioException catch (e) {
        final status = e.response?.statusCode;
        if (status == 401) {
          await _handle401();
          return const ApiResult<void>(
            error: 'Not authorized',
            unauthorized: true,
          );
        }
        if (status == 409) {
          final expected = _expectedOffset(e.response?.data);
          if (expected != null) {
            offset = expected.clamp(0, size).toInt();
            onProgress?.call(offset, size);
            continue;
          }
        }
        if (_isNetworkError(e)) {
          return ApiResult<void>(error: e.message ?? 'Network request failed');
        }
        return ApiResult<void>(error: _messageFrom(e.response?.data, status));
      }
    }

    return _guard<void>(
      () => _dio.post<dynamic>('/api/uploads/${session.id}/complete'),
    );
  }

  int? _expectedOffset(Object? data) {
    if (data is Map<String, dynamic>) {
      return (data['expected'] as num?)?.toInt();
    }
    if (data is String) {
      try {
        return (jsonDecode(data) as Map<String, dynamic>)['expected'] as int?;
      } catch (_) {
        return null;
      }
    }
    return null;
  }

  Future<ApiResult<void>> _uploadMultipart({
    required String fileName,
    required String dirPath,
    File? file,
    Uint8List? bytes,
    void Function(int loaded, int total)? onProgress,
  }) async {
    try {
      final form = FormData.fromMap({
        if (dirPath.isNotEmpty) 'path': dirPath,
        'file': file != null
            ? await MultipartFile.fromFile(file.path, filename: fileName)
            : MultipartFile.fromBytes(bytes!, filename: fileName),
      });
      await _dio.post<dynamic>(
        '/api/upload',
        data: form,
        onSendProgress: onProgress,
      );
      return const ApiResult<void>();
    } on DioException catch (e) {
      final status = e.response?.statusCode;
      if (status == 401) {
        await _handle401();
        return const ApiResult<void>(
          error: 'Not authorized',
          unauthorized: true,
        );
      }
      if (_isNetworkError(e)) {
        return ApiResult<void>(error: e.message ?? 'Network request failed');
      }
      return ApiResult<void>(error: _messageFrom(e.response?.data, status));
    }
  }

  Future<ApiResult<DownloadResult>> download(String path) async {
    try {
      final dir = await getTemporaryDirectory();
      final suggested = baseName(path);
      final stamp = DateTime.now().millisecondsSinceEpoch;
      final savePath = '${dir.path}/$stamp-$suggested';
      final response = await _dio.download(
        buildUrl('/api/files/download', {'path': path}),
        savePath,
      );
      final disposition = response.headers.value('content-disposition') ?? '';
      final match = RegExp(
        r'''filename\*?=(?:UTF-8'')?"?([^";]+)"?''',
        caseSensitive: false,
      ).firstMatch(disposition);
      var name = suggested;
      if (match != null) {
        try {
          name = Uri.decodeComponent(match.group(1)!);
        } catch (_) {
          name = match.group(1)!;
        }
      }
      var file = File(savePath);
      if (name != suggested) {
        final target =
            File('${dir.path}/${DateTime.now().millisecondsSinceEpoch}-$name');
        file = await file.rename(target.path);
      }
      return ApiResult(data: DownloadResult(file: file, name: name));
    } on DioException catch (e) {
      final status = e.response?.statusCode;
      if (status == 401) {
        await _handle401();
        return const ApiResult<DownloadResult>(
          error: 'Not authorized',
          unauthorized: true,
        );
      }
      if (_isNetworkError(e)) {
        return ApiResult<DownloadResult>(
          error: e.message ?? 'Network request failed',
        );
      }
      return ApiResult<DownloadResult>(
        error: 'Download failed (${status ?? 0})',
      );
    }
  }
}
