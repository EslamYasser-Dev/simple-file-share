// This is a generated file - do not edit.
//
// Generated from fileshare/v1/fileshare.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:grpc/service_api.dart' as $grpc;
import 'package:protobuf/protobuf.dart' as $pb;

import 'fileshare.pb.dart' as $0;

export 'fileshare.pb.dart';

/// AuthService exposes account lifecycle and identity operations.
@$pb.GrpcServiceName('fileshare.v1.AuthService')
class AuthServiceClient extends $grpc.Client {
  /// The hostname for this service.
  static const $core.String defaultHost = '';

  /// OAuth scopes needed for the client.
  static const $core.List<$core.String> oauthScopes = [
    '',
  ];

  AuthServiceClient(super.channel, {super.options, super.interceptors});

  /// Register creates a new account (public, gated by ENABLE_SIGNUP).
  $grpc.ResponseFuture<$0.User> register(
    $0.RegisterRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$register, request, options: options);
  }

  /// Authenticate validates credentials and returns the account. Subsequent
  /// calls must still carry Basic credentials in request metadata.
  $grpc.ResponseFuture<$0.User> authenticate(
    $0.AuthenticateRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$authenticate, request, options: options);
  }

  /// Me returns the authenticated caller's identity.
  $grpc.ResponseFuture<$0.User> me(
    $0.MeRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$me, request, options: options);
  }

  /// GetAuthInfo reports runtime auth capabilities (public).
  $grpc.ResponseFuture<$0.AuthInfoResponse> getAuthInfo(
    $0.GetAuthInfoRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$getAuthInfo, request, options: options);
  }

  /// ListUsers returns accounts with storage usage (admin only).
  $grpc.ResponseFuture<$0.ListUsersResponse> listUsers(
    $0.ListUsersRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$listUsers, request, options: options);
  }

  /// Login exchanges credentials for a bearer access token (public).
  $grpc.ResponseFuture<$0.LoginResponse> login(
    $0.LoginRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$login, request, options: options);
  }

  /// Logout revokes the bearer token carried in request metadata (public; an
  /// expired or missing token is a no-op so clients can always sign out).
  $grpc.ResponseFuture<$0.LogoutResponse> logout(
    $0.LogoutRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$logout, request, options: options);
  }

  // method descriptors

  static final _$register = $grpc.ClientMethod<$0.RegisterRequest, $0.User>(
      '/fileshare.v1.AuthService/Register',
      ($0.RegisterRequest value) => value.writeToBuffer(),
      $0.User.fromBuffer);
  static final _$authenticate =
      $grpc.ClientMethod<$0.AuthenticateRequest, $0.User>(
          '/fileshare.v1.AuthService/Authenticate',
          ($0.AuthenticateRequest value) => value.writeToBuffer(),
          $0.User.fromBuffer);
  static final _$me = $grpc.ClientMethod<$0.MeRequest, $0.User>(
      '/fileshare.v1.AuthService/Me',
      ($0.MeRequest value) => value.writeToBuffer(),
      $0.User.fromBuffer);
  static final _$getAuthInfo =
      $grpc.ClientMethod<$0.GetAuthInfoRequest, $0.AuthInfoResponse>(
          '/fileshare.v1.AuthService/GetAuthInfo',
          ($0.GetAuthInfoRequest value) => value.writeToBuffer(),
          $0.AuthInfoResponse.fromBuffer);
  static final _$listUsers =
      $grpc.ClientMethod<$0.ListUsersRequest, $0.ListUsersResponse>(
          '/fileshare.v1.AuthService/ListUsers',
          ($0.ListUsersRequest value) => value.writeToBuffer(),
          $0.ListUsersResponse.fromBuffer);
  static final _$login = $grpc.ClientMethod<$0.LoginRequest, $0.LoginResponse>(
      '/fileshare.v1.AuthService/Login',
      ($0.LoginRequest value) => value.writeToBuffer(),
      $0.LoginResponse.fromBuffer);
  static final _$logout =
      $grpc.ClientMethod<$0.LogoutRequest, $0.LogoutResponse>(
          '/fileshare.v1.AuthService/Logout',
          ($0.LogoutRequest value) => value.writeToBuffer(),
          $0.LogoutResponse.fromBuffer);
}

@$pb.GrpcServiceName('fileshare.v1.AuthService')
abstract class AuthServiceBase extends $grpc.Service {
  $core.String get $name => 'fileshare.v1.AuthService';

  AuthServiceBase() {
    $addMethod($grpc.ServiceMethod<$0.RegisterRequest, $0.User>(
        'Register',
        register_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.RegisterRequest.fromBuffer(value),
        ($0.User value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.AuthenticateRequest, $0.User>(
        'Authenticate',
        authenticate_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.AuthenticateRequest.fromBuffer(value),
        ($0.User value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.MeRequest, $0.User>(
        'Me',
        me_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.MeRequest.fromBuffer(value),
        ($0.User value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.GetAuthInfoRequest, $0.AuthInfoResponse>(
        'GetAuthInfo',
        getAuthInfo_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.GetAuthInfoRequest.fromBuffer(value),
        ($0.AuthInfoResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.ListUsersRequest, $0.ListUsersResponse>(
        'ListUsers',
        listUsers_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.ListUsersRequest.fromBuffer(value),
        ($0.ListUsersResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.LoginRequest, $0.LoginResponse>(
        'Login',
        login_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.LoginRequest.fromBuffer(value),
        ($0.LoginResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.LogoutRequest, $0.LogoutResponse>(
        'Logout',
        logout_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.LogoutRequest.fromBuffer(value),
        ($0.LogoutResponse value) => value.writeToBuffer()));
  }

  $async.Future<$0.User> register_Pre($grpc.ServiceCall $call,
      $async.Future<$0.RegisterRequest> $request) async {
    return register($call, await $request);
  }

  $async.Future<$0.User> register(
      $grpc.ServiceCall call, $0.RegisterRequest request);

  $async.Future<$0.User> authenticate_Pre($grpc.ServiceCall $call,
      $async.Future<$0.AuthenticateRequest> $request) async {
    return authenticate($call, await $request);
  }

  $async.Future<$0.User> authenticate(
      $grpc.ServiceCall call, $0.AuthenticateRequest request);

  $async.Future<$0.User> me_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.MeRequest> $request) async {
    return me($call, await $request);
  }

  $async.Future<$0.User> me($grpc.ServiceCall call, $0.MeRequest request);

  $async.Future<$0.AuthInfoResponse> getAuthInfo_Pre($grpc.ServiceCall $call,
      $async.Future<$0.GetAuthInfoRequest> $request) async {
    return getAuthInfo($call, await $request);
  }

  $async.Future<$0.AuthInfoResponse> getAuthInfo(
      $grpc.ServiceCall call, $0.GetAuthInfoRequest request);

  $async.Future<$0.ListUsersResponse> listUsers_Pre($grpc.ServiceCall $call,
      $async.Future<$0.ListUsersRequest> $request) async {
    return listUsers($call, await $request);
  }

  $async.Future<$0.ListUsersResponse> listUsers(
      $grpc.ServiceCall call, $0.ListUsersRequest request);

  $async.Future<$0.LoginResponse> login_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.LoginRequest> $request) async {
    return login($call, await $request);
  }

  $async.Future<$0.LoginResponse> login(
      $grpc.ServiceCall call, $0.LoginRequest request);

  $async.Future<$0.LogoutResponse> logout_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.LogoutRequest> $request) async {
    return logout($call, await $request);
  }

  $async.Future<$0.LogoutResponse> logout(
      $grpc.ServiceCall call, $0.LogoutRequest request);
}

/// ShareService mirrors the HTTP share API.
@$pb.GrpcServiceName('fileshare.v1.ShareService')
class ShareServiceClient extends $grpc.Client {
  /// The hostname for this service.
  static const $core.String defaultHost = '';

  /// OAuth scopes needed for the client.
  static const $core.List<$core.String> oauthScopes = [
    '',
  ];

  ShareServiceClient(super.channel, {super.options, super.interceptors});

  $grpc.ResponseFuture<$0.Share> createShare(
    $0.CreateShareRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$createShare, request, options: options);
  }

  $grpc.ResponseFuture<$0.ListSharesResponse> listShares(
    $0.ListSharesRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$listShares, request, options: options);
  }

  $grpc.ResponseFuture<$0.RevokeShareResponse> revokeShare(
    $0.RevokeShareRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$revokeShare, request, options: options);
  }

  // method descriptors

  static final _$createShare =
      $grpc.ClientMethod<$0.CreateShareRequest, $0.Share>(
          '/fileshare.v1.ShareService/CreateShare',
          ($0.CreateShareRequest value) => value.writeToBuffer(),
          $0.Share.fromBuffer);
  static final _$listShares =
      $grpc.ClientMethod<$0.ListSharesRequest, $0.ListSharesResponse>(
          '/fileshare.v1.ShareService/ListShares',
          ($0.ListSharesRequest value) => value.writeToBuffer(),
          $0.ListSharesResponse.fromBuffer);
  static final _$revokeShare =
      $grpc.ClientMethod<$0.RevokeShareRequest, $0.RevokeShareResponse>(
          '/fileshare.v1.ShareService/RevokeShare',
          ($0.RevokeShareRequest value) => value.writeToBuffer(),
          $0.RevokeShareResponse.fromBuffer);
}

@$pb.GrpcServiceName('fileshare.v1.ShareService')
abstract class ShareServiceBase extends $grpc.Service {
  $core.String get $name => 'fileshare.v1.ShareService';

  ShareServiceBase() {
    $addMethod($grpc.ServiceMethod<$0.CreateShareRequest, $0.Share>(
        'CreateShare',
        createShare_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.CreateShareRequest.fromBuffer(value),
        ($0.Share value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.ListSharesRequest, $0.ListSharesResponse>(
        'ListShares',
        listShares_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.ListSharesRequest.fromBuffer(value),
        ($0.ListSharesResponse value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$0.RevokeShareRequest, $0.RevokeShareResponse>(
            'RevokeShare',
            revokeShare_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $0.RevokeShareRequest.fromBuffer(value),
            ($0.RevokeShareResponse value) => value.writeToBuffer()));
  }

  $async.Future<$0.Share> createShare_Pre($grpc.ServiceCall $call,
      $async.Future<$0.CreateShareRequest> $request) async {
    return createShare($call, await $request);
  }

  $async.Future<$0.Share> createShare(
      $grpc.ServiceCall call, $0.CreateShareRequest request);

  $async.Future<$0.ListSharesResponse> listShares_Pre($grpc.ServiceCall $call,
      $async.Future<$0.ListSharesRequest> $request) async {
    return listShares($call, await $request);
  }

  $async.Future<$0.ListSharesResponse> listShares(
      $grpc.ServiceCall call, $0.ListSharesRequest request);

  $async.Future<$0.RevokeShareResponse> revokeShare_Pre($grpc.ServiceCall $call,
      $async.Future<$0.RevokeShareRequest> $request) async {
    return revokeShare($call, await $request);
  }

  $async.Future<$0.RevokeShareResponse> revokeShare(
      $grpc.ServiceCall call, $0.RevokeShareRequest request);
}

/// EventsService streams live application events for native clients
/// (the browser keeps using HTTP Server-Sent Events).
@$pb.GrpcServiceName('fileshare.v1.EventsService')
class EventsServiceClient extends $grpc.Client {
  /// The hostname for this service.
  static const $core.String defaultHost = '';

  /// OAuth scopes needed for the client.
  static const $core.List<$core.String> oauthScopes = [
    '',
  ];

  EventsServiceClient(super.channel, {super.options, super.interceptors});

  $grpc.ResponseStream<$0.EventMessage> subscribe(
    $0.SubscribeRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createStreamingCall(
        _$subscribe, $async.Stream.fromIterable([request]),
        options: options);
  }

  // method descriptors

  static final _$subscribe =
      $grpc.ClientMethod<$0.SubscribeRequest, $0.EventMessage>(
          '/fileshare.v1.EventsService/Subscribe',
          ($0.SubscribeRequest value) => value.writeToBuffer(),
          $0.EventMessage.fromBuffer);
}

@$pb.GrpcServiceName('fileshare.v1.EventsService')
abstract class EventsServiceBase extends $grpc.Service {
  $core.String get $name => 'fileshare.v1.EventsService';

  EventsServiceBase() {
    $addMethod($grpc.ServiceMethod<$0.SubscribeRequest, $0.EventMessage>(
        'Subscribe',
        subscribe_Pre,
        false,
        true,
        ($core.List<$core.int> value) => $0.SubscribeRequest.fromBuffer(value),
        ($0.EventMessage value) => value.writeToBuffer()));
  }

  $async.Stream<$0.EventMessage> subscribe_Pre($grpc.ServiceCall $call,
      $async.Future<$0.SubscribeRequest> $request) async* {
    yield* subscribe($call, await $request);
  }

  $async.Stream<$0.EventMessage> subscribe(
      $grpc.ServiceCall call, $0.SubscribeRequest request);
}

/// FileService mirrors the HTTP file API.
@$pb.GrpcServiceName('fileshare.v1.FileService')
class FileServiceClient extends $grpc.Client {
  /// The hostname for this service.
  static const $core.String defaultHost = '';

  /// OAuth scopes needed for the client.
  static const $core.List<$core.String> oauthScopes = [
    '',
  ];

  FileServiceClient(super.channel, {super.options, super.interceptors});

  $grpc.ResponseFuture<$0.ListFilesResponse> listFiles(
    $0.ListFilesRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$listFiles, request, options: options);
  }

  $grpc.ResponseFuture<$0.FileInfo> getFileInfo(
    $0.GetFileInfoRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$getFileInfo, request, options: options);
  }

  $grpc.ResponseFuture<$0.SearchFilesResponse> searchFiles(
    $0.SearchFilesRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$searchFiles, request, options: options);
  }

  $grpc.ResponseFuture<$0.CreateDirectoryResponse> createDirectory(
    $0.CreateDirectoryRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$createDirectory, request, options: options);
  }

  $grpc.ResponseFuture<$0.DeletePathResponse> deletePath(
    $0.DeletePathRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$deletePath, request, options: options);
  }

  $grpc.ResponseFuture<$0.UpdateFileContentResponse> updateFileContent(
    $0.UpdateFileContentRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$updateFileContent, request, options: options);
  }

  /// UploadFile accepts a metadata message followed by file chunks.
  $grpc.ResponseFuture<$0.UploadResponse> uploadFile(
    $async.Stream<$0.UploadRequest> request, {
    $grpc.CallOptions? options,
  }) {
    return $createStreamingCall(_$uploadFile, request, options: options).single;
  }

  /// DownloadFile streams a file (or a zipped directory) in chunks.
  $grpc.ResponseStream<$0.DownloadChunk> downloadFile(
    $0.DownloadFileRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createStreamingCall(
        _$downloadFile, $async.Stream.fromIterable([request]),
        options: options);
  }

  // method descriptors

  static final _$listFiles =
      $grpc.ClientMethod<$0.ListFilesRequest, $0.ListFilesResponse>(
          '/fileshare.v1.FileService/ListFiles',
          ($0.ListFilesRequest value) => value.writeToBuffer(),
          $0.ListFilesResponse.fromBuffer);
  static final _$getFileInfo =
      $grpc.ClientMethod<$0.GetFileInfoRequest, $0.FileInfo>(
          '/fileshare.v1.FileService/GetFileInfo',
          ($0.GetFileInfoRequest value) => value.writeToBuffer(),
          $0.FileInfo.fromBuffer);
  static final _$searchFiles =
      $grpc.ClientMethod<$0.SearchFilesRequest, $0.SearchFilesResponse>(
          '/fileshare.v1.FileService/SearchFiles',
          ($0.SearchFilesRequest value) => value.writeToBuffer(),
          $0.SearchFilesResponse.fromBuffer);
  static final _$createDirectory =
      $grpc.ClientMethod<$0.CreateDirectoryRequest, $0.CreateDirectoryResponse>(
          '/fileshare.v1.FileService/CreateDirectory',
          ($0.CreateDirectoryRequest value) => value.writeToBuffer(),
          $0.CreateDirectoryResponse.fromBuffer);
  static final _$deletePath =
      $grpc.ClientMethod<$0.DeletePathRequest, $0.DeletePathResponse>(
          '/fileshare.v1.FileService/DeletePath',
          ($0.DeletePathRequest value) => value.writeToBuffer(),
          $0.DeletePathResponse.fromBuffer);
  static final _$updateFileContent = $grpc.ClientMethod<
          $0.UpdateFileContentRequest, $0.UpdateFileContentResponse>(
      '/fileshare.v1.FileService/UpdateFileContent',
      ($0.UpdateFileContentRequest value) => value.writeToBuffer(),
      $0.UpdateFileContentResponse.fromBuffer);
  static final _$uploadFile =
      $grpc.ClientMethod<$0.UploadRequest, $0.UploadResponse>(
          '/fileshare.v1.FileService/UploadFile',
          ($0.UploadRequest value) => value.writeToBuffer(),
          $0.UploadResponse.fromBuffer);
  static final _$downloadFile =
      $grpc.ClientMethod<$0.DownloadFileRequest, $0.DownloadChunk>(
          '/fileshare.v1.FileService/DownloadFile',
          ($0.DownloadFileRequest value) => value.writeToBuffer(),
          $0.DownloadChunk.fromBuffer);
}

@$pb.GrpcServiceName('fileshare.v1.FileService')
abstract class FileServiceBase extends $grpc.Service {
  $core.String get $name => 'fileshare.v1.FileService';

  FileServiceBase() {
    $addMethod($grpc.ServiceMethod<$0.ListFilesRequest, $0.ListFilesResponse>(
        'ListFiles',
        listFiles_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.ListFilesRequest.fromBuffer(value),
        ($0.ListFilesResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.GetFileInfoRequest, $0.FileInfo>(
        'GetFileInfo',
        getFileInfo_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.GetFileInfoRequest.fromBuffer(value),
        ($0.FileInfo value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$0.SearchFilesRequest, $0.SearchFilesResponse>(
            'SearchFiles',
            searchFiles_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $0.SearchFilesRequest.fromBuffer(value),
            ($0.SearchFilesResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.CreateDirectoryRequest,
            $0.CreateDirectoryResponse>(
        'CreateDirectory',
        createDirectory_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.CreateDirectoryRequest.fromBuffer(value),
        ($0.CreateDirectoryResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.DeletePathRequest, $0.DeletePathResponse>(
        'DeletePath',
        deletePath_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.DeletePathRequest.fromBuffer(value),
        ($0.DeletePathResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.UpdateFileContentRequest,
            $0.UpdateFileContentResponse>(
        'UpdateFileContent',
        updateFileContent_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.UpdateFileContentRequest.fromBuffer(value),
        ($0.UpdateFileContentResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.UploadRequest, $0.UploadResponse>(
        'UploadFile',
        uploadFile,
        true,
        false,
        ($core.List<$core.int> value) => $0.UploadRequest.fromBuffer(value),
        ($0.UploadResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.DownloadFileRequest, $0.DownloadChunk>(
        'DownloadFile',
        downloadFile_Pre,
        false,
        true,
        ($core.List<$core.int> value) =>
            $0.DownloadFileRequest.fromBuffer(value),
        ($0.DownloadChunk value) => value.writeToBuffer()));
  }

  $async.Future<$0.ListFilesResponse> listFiles_Pre($grpc.ServiceCall $call,
      $async.Future<$0.ListFilesRequest> $request) async {
    return listFiles($call, await $request);
  }

  $async.Future<$0.ListFilesResponse> listFiles(
      $grpc.ServiceCall call, $0.ListFilesRequest request);

  $async.Future<$0.FileInfo> getFileInfo_Pre($grpc.ServiceCall $call,
      $async.Future<$0.GetFileInfoRequest> $request) async {
    return getFileInfo($call, await $request);
  }

  $async.Future<$0.FileInfo> getFileInfo(
      $grpc.ServiceCall call, $0.GetFileInfoRequest request);

  $async.Future<$0.SearchFilesResponse> searchFiles_Pre($grpc.ServiceCall $call,
      $async.Future<$0.SearchFilesRequest> $request) async {
    return searchFiles($call, await $request);
  }

  $async.Future<$0.SearchFilesResponse> searchFiles(
      $grpc.ServiceCall call, $0.SearchFilesRequest request);

  $async.Future<$0.CreateDirectoryResponse> createDirectory_Pre(
      $grpc.ServiceCall $call,
      $async.Future<$0.CreateDirectoryRequest> $request) async {
    return createDirectory($call, await $request);
  }

  $async.Future<$0.CreateDirectoryResponse> createDirectory(
      $grpc.ServiceCall call, $0.CreateDirectoryRequest request);

  $async.Future<$0.DeletePathResponse> deletePath_Pre($grpc.ServiceCall $call,
      $async.Future<$0.DeletePathRequest> $request) async {
    return deletePath($call, await $request);
  }

  $async.Future<$0.DeletePathResponse> deletePath(
      $grpc.ServiceCall call, $0.DeletePathRequest request);

  $async.Future<$0.UpdateFileContentResponse> updateFileContent_Pre(
      $grpc.ServiceCall $call,
      $async.Future<$0.UpdateFileContentRequest> $request) async {
    return updateFileContent($call, await $request);
  }

  $async.Future<$0.UpdateFileContentResponse> updateFileContent(
      $grpc.ServiceCall call, $0.UpdateFileContentRequest request);

  $async.Future<$0.UploadResponse> uploadFile(
      $grpc.ServiceCall call, $async.Stream<$0.UploadRequest> request);

  $async.Stream<$0.DownloadChunk> downloadFile_Pre($grpc.ServiceCall $call,
      $async.Future<$0.DownloadFileRequest> $request) async* {
    yield* downloadFile($call, await $request);
  }

  $async.Stream<$0.DownloadChunk> downloadFile(
      $grpc.ServiceCall call, $0.DownloadFileRequest request);
}
