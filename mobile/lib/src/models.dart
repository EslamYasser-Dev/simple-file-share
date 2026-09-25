class FileItem {
  const FileItem({
    required this.name,
    required this.path,
    required this.size,
    required this.isDir,
    required this.modified,
    this.mimeType,
    this.version,
  });

  final String name;
  final String path;
  final int size;
  final bool isDir;
  final String modified;
  final String? mimeType;
  final int? version;

  factory FileItem.fromJson(Map<String, dynamic> json) => FileItem(
        name: json['name'] as String? ?? '',
        path: json['path'] as String? ?? '',
        size: (json['size'] as num?)?.toInt() ?? 0,
        isDir: json['isDir'] as bool? ?? false,
        modified: json['modified'] as String? ?? '',
        mimeType: json['mimeType'] as String?,
        version: (json['version'] as num?)?.toInt(),
      );
}

class AuthUser {
  const AuthUser({
    required this.username,
    required this.isAdmin,
    this.role,
    this.enabled,
    this.permissions,
    this.createdAt,
    this.quotaBytes,
    this.size,
    this.files,
  });

  final String username;
  final bool isAdmin;
  final String? role;
  final bool? enabled;
  final List<String>? permissions;
  final String? createdAt;
  final int? quotaBytes;
  final int? size;
  final int? files;

  factory AuthUser.fromJson(Map<String, dynamic> json) => AuthUser(
        username: json['username'] as String? ?? '',
        isAdmin: json['isAdmin'] as bool? ?? false,
        role: json['role'] as String?,
        enabled: json['enabled'] as bool?,
        permissions: (json['permissions'] as List?)
            ?.map((e) => e.toString())
            .toList(),
        createdAt: json['createdAt'] as String?,
        quotaBytes: (json['quotaBytes'] as num?)?.toInt(),
        size: (json['size'] as num?)?.toInt(),
        files: (json['files'] as num?)?.toInt(),
      );
}

class ShareItem {
  const ShareItem({
    required this.token,
    required this.path,
    required this.name,
    required this.owner,
    required this.createdAt,
    required this.expiresAt,
  });

  final String token;
  final String path;
  final String name;
  final String owner;
  final String createdAt;
  final String expiresAt;

  factory ShareItem.fromJson(Map<String, dynamic> json) => ShareItem(
        token: json['token'] as String? ?? '',
        path: json['path'] as String? ?? '',
        name: json['name'] as String? ?? '',
        owner: json['owner'] as String? ?? '',
        createdAt: json['createdAt'] as String? ?? '',
        expiresAt: json['expiresAt'] as String? ?? '',
      );
}

class TokenResponse {
  const TokenResponse({
    required this.accessToken,
    required this.tokenType,
    required this.expiresIn,
  });

  final String accessToken;
  final String tokenType;
  final int expiresIn;

  factory TokenResponse.fromJson(Map<String, dynamic> json) => TokenResponse(
        accessToken: json['accessToken'] as String? ?? '',
        tokenType: json['tokenType'] as String? ?? '',
        expiresIn: (json['expiresIn'] as num?)?.toInt() ?? 0,
      );
}

class ServerEvent {
  const ServerEvent({required this.type, this.path, this.user, this.at});

  final String type;
  final String? path;
  final String? user;
  final String? at;

  factory ServerEvent.fromJson(Map<String, dynamic> json) => ServerEvent(
        type: json['type'] as String? ?? '',
        path: json['path'] as String?,
        user: json['user'] as String?,
        at: json['at'] as String?,
      );
}

class UploadSessionInfo {
  const UploadSessionInfo({
    required this.id,
    required this.offset,
    required this.size,
    required this.chunkSize,
  });

  final String id;
  final int offset;
  final int size;
  final int chunkSize;

  factory UploadSessionInfo.fromJson(Map<String, dynamic> json) =>
      UploadSessionInfo(
        id: json['id'] as String? ?? '',
        offset: (json['offset'] as num?)?.toInt() ?? 0,
        size: (json['size'] as num?)?.toInt() ?? 0,
        chunkSize: (json['chunkSize'] as num?)?.toInt() ?? 0,
      );
}

class ApiResult<T> {
  const ApiResult({this.data, this.error, this.unauthorized = false});

  factory ApiResult.success(T? data) => ApiResult<T>(data: data);
  factory ApiResult.failure(String error) => ApiResult<T>(error: error);

  final T? data;
  final String? error;
  final bool unauthorized;

  bool get ok => error == null && !unauthorized;
}
