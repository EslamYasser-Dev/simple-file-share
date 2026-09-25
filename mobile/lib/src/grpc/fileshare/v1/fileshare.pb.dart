// This is a generated file - do not edit.
//
// Generated from fileshare/v1/fileshare.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:core' as $core;

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;
import 'package:protobuf/well_known_types/google/protobuf/timestamp.pb.dart'
    as $1;

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

class User extends $pb.GeneratedMessage {
  factory User({
    $core.String? username,
    $core.bool? isAdmin,
    $core.String? role,
    $core.bool? enabled,
    $fixnum.Int64? quotaBytes,
    $fixnum.Int64? size,
    $fixnum.Int64? files,
    $core.Iterable<$core.String>? permissions,
    $core.String? createdAt,
  }) {
    final result = User._();
    if (username != null) result.username = username;
    if (isAdmin != null) result.isAdmin = isAdmin;
    if (role != null) result.role = role;
    if (enabled != null) result.enabled = enabled;
    if (quotaBytes != null) result.quotaBytes = quotaBytes;
    if (size != null) result.size = size;
    if (files != null) result.files = files;
    if (permissions != null) result.permissions.addAll(permissions);
    if (createdAt != null) result.createdAt = createdAt;
    return result;
  }

  User._();

  factory User.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      User()..mergeFromBuffer(data, registry);
  factory User.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      User()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'User',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'fileshare.v1'),
      createEmptyInstance: User.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'username')
    ..aOB(2, _omitFieldNames ? '' : 'isAdmin')
    ..aOS(3, _omitFieldNames ? '' : 'role')
    ..aOB(4, _omitFieldNames ? '' : 'enabled')
    ..aInt64(5, _omitFieldNames ? '' : 'quotaBytes')
    ..aInt64(6, _omitFieldNames ? '' : 'size')
    ..aInt64(7, _omitFieldNames ? '' : 'files')
    ..pPS(8, _omitFieldNames ? '' : 'permissions')
    ..aOS(9, _omitFieldNames ? '' : 'createdAt')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  User clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  User copyWith(void Function(User) updates) =>
      super.copyWith((message) => updates(message as User)) as User;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use User() / User.new instead')
  static User create() => User._();
  static $pb.GeneratedMessage $_createMessage() => User._();
  @$core.override
  User createEmptyInstance() => User._();
  @$core.pragma('dart2js:noInline')
  static User getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<User>(User.$_createMessage);
  static User? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get username => $_getSZ(0);
  @$pb.TagNumber(1)
  set username($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasUsername() => $_has(0);
  @$pb.TagNumber(1)
  void clearUsername() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.bool get isAdmin => $_getBF(1);
  @$pb.TagNumber(2)
  set isAdmin($core.bool value) => $_setBool(1, value);
  @$pb.TagNumber(2)
  $core.bool hasIsAdmin() => $_has(1);
  @$pb.TagNumber(2)
  void clearIsAdmin() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get role => $_getSZ(2);
  @$pb.TagNumber(3)
  set role($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasRole() => $_has(2);
  @$pb.TagNumber(3)
  void clearRole() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.bool get enabled => $_getBF(3);
  @$pb.TagNumber(4)
  set enabled($core.bool value) => $_setBool(3, value);
  @$pb.TagNumber(4)
  $core.bool hasEnabled() => $_has(3);
  @$pb.TagNumber(4)
  void clearEnabled() => $_clearField(4);

  @$pb.TagNumber(5)
  $fixnum.Int64 get quotaBytes => $_getI64(4);
  @$pb.TagNumber(5)
  set quotaBytes($fixnum.Int64 value) => $_setInt64(4, value);
  @$pb.TagNumber(5)
  $core.bool hasQuotaBytes() => $_has(4);
  @$pb.TagNumber(5)
  void clearQuotaBytes() => $_clearField(5);

  @$pb.TagNumber(6)
  $fixnum.Int64 get size => $_getI64(5);
  @$pb.TagNumber(6)
  set size($fixnum.Int64 value) => $_setInt64(5, value);
  @$pb.TagNumber(6)
  $core.bool hasSize() => $_has(5);
  @$pb.TagNumber(6)
  void clearSize() => $_clearField(6);

  @$pb.TagNumber(7)
  $fixnum.Int64 get files => $_getI64(6);
  @$pb.TagNumber(7)
  set files($fixnum.Int64 value) => $_setInt64(6, value);
  @$pb.TagNumber(7)
  $core.bool hasFiles() => $_has(6);
  @$pb.TagNumber(7)
  void clearFiles() => $_clearField(7);

  @$pb.TagNumber(8)
  $pb.PbList<$core.String> get permissions => $_getList(7);

  @$pb.TagNumber(9)
  $core.String get createdAt => $_getSZ(8);
  @$pb.TagNumber(9)
  set createdAt($core.String value) => $_setString(8, value);
  @$pb.TagNumber(9)
  $core.bool hasCreatedAt() => $_has(8);
  @$pb.TagNumber(9)
  void clearCreatedAt() => $_clearField(9);
}

class AuthInfoResponse extends $pb.GeneratedMessage {
  factory AuthInfoResponse({
    $core.bool? signupEnabled,
  }) {
    final result = AuthInfoResponse._();
    if (signupEnabled != null) result.signupEnabled = signupEnabled;
    return result;
  }

  AuthInfoResponse._();

  factory AuthInfoResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AuthInfoResponse()..mergeFromBuffer(data, registry);
  factory AuthInfoResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AuthInfoResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AuthInfoResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'fileshare.v1'),
      createEmptyInstance: AuthInfoResponse.$_createMessage)
    ..aOB(1, _omitFieldNames ? '' : 'signupEnabled')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AuthInfoResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AuthInfoResponse copyWith(void Function(AuthInfoResponse) updates) =>
      super.copyWith((message) => updates(message as AuthInfoResponse))
          as AuthInfoResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use AuthInfoResponse() / AuthInfoResponse.new instead')
  static AuthInfoResponse create() => AuthInfoResponse._();
  static $pb.GeneratedMessage $_createMessage() => AuthInfoResponse._();
  @$core.override
  AuthInfoResponse createEmptyInstance() => AuthInfoResponse._();
  @$core.pragma('dart2js:noInline')
  static AuthInfoResponse getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<AuthInfoResponse>(
          AuthInfoResponse.$_createMessage);
  static AuthInfoResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.bool get signupEnabled => $_getBF(0);
  @$pb.TagNumber(1)
  set signupEnabled($core.bool value) => $_setBool(0, value);
  @$pb.TagNumber(1)
  $core.bool hasSignupEnabled() => $_has(0);
  @$pb.TagNumber(1)
  void clearSignupEnabled() => $_clearField(1);
}

class RegisterRequest extends $pb.GeneratedMessage {
  factory RegisterRequest({
    $core.String? username,
    $core.String? password,
  }) {
    final result = RegisterRequest._();
    if (username != null) result.username = username;
    if (password != null) result.password = password;
    return result;
  }

  RegisterRequest._();

  factory RegisterRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RegisterRequest()..mergeFromBuffer(data, registry);
  factory RegisterRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RegisterRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RegisterRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'fileshare.v1'),
      createEmptyInstance: RegisterRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'username')
    ..aOS(2, _omitFieldNames ? '' : 'password')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RegisterRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RegisterRequest copyWith(void Function(RegisterRequest) updates) =>
      super.copyWith((message) => updates(message as RegisterRequest))
          as RegisterRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use RegisterRequest() / RegisterRequest.new instead')
  static RegisterRequest create() => RegisterRequest._();
  static $pb.GeneratedMessage $_createMessage() => RegisterRequest._();
  @$core.override
  RegisterRequest createEmptyInstance() => RegisterRequest._();
  @$core.pragma('dart2js:noInline')
  static RegisterRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<RegisterRequest>(
          RegisterRequest.$_createMessage);
  static RegisterRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get username => $_getSZ(0);
  @$pb.TagNumber(1)
  set username($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasUsername() => $_has(0);
  @$pb.TagNumber(1)
  void clearUsername() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get password => $_getSZ(1);
  @$pb.TagNumber(2)
  set password($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasPassword() => $_has(1);
  @$pb.TagNumber(2)
  void clearPassword() => $_clearField(2);
}

class AuthenticateRequest extends $pb.GeneratedMessage {
  factory AuthenticateRequest({
    $core.String? username,
    $core.String? password,
  }) {
    final result = AuthenticateRequest._();
    if (username != null) result.username = username;
    if (password != null) result.password = password;
    return result;
  }

  AuthenticateRequest._();

  factory AuthenticateRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AuthenticateRequest()..mergeFromBuffer(data, registry);
  factory AuthenticateRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AuthenticateRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AuthenticateRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'fileshare.v1'),
      createEmptyInstance: AuthenticateRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'username')
    ..aOS(2, _omitFieldNames ? '' : 'password')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AuthenticateRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AuthenticateRequest copyWith(void Function(AuthenticateRequest) updates) =>
      super.copyWith((message) => updates(message as AuthenticateRequest))
          as AuthenticateRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use AuthenticateRequest() / AuthenticateRequest.new instead')
  static AuthenticateRequest create() => AuthenticateRequest._();
  static $pb.GeneratedMessage $_createMessage() => AuthenticateRequest._();
  @$core.override
  AuthenticateRequest createEmptyInstance() => AuthenticateRequest._();
  @$core.pragma('dart2js:noInline')
  static AuthenticateRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<AuthenticateRequest>(
          AuthenticateRequest.$_createMessage);
  static AuthenticateRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get username => $_getSZ(0);
  @$pb.TagNumber(1)
  set username($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasUsername() => $_has(0);
  @$pb.TagNumber(1)
  void clearUsername() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get password => $_getSZ(1);
  @$pb.TagNumber(2)
  set password($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasPassword() => $_has(1);
  @$pb.TagNumber(2)
  void clearPassword() => $_clearField(2);
}

class MeRequest extends $pb.GeneratedMessage {
  factory MeRequest() => MeRequest._();

  MeRequest._();

  factory MeRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      MeRequest()..mergeFromBuffer(data, registry);
  factory MeRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      MeRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'MeRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'fileshare.v1'),
      createEmptyInstance: MeRequest.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MeRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MeRequest copyWith(void Function(MeRequest) updates) =>
      super.copyWith((message) => updates(message as MeRequest)) as MeRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use MeRequest() / MeRequest.new instead')
  static MeRequest create() => MeRequest._();
  static $pb.GeneratedMessage $_createMessage() => MeRequest._();
  @$core.override
  MeRequest createEmptyInstance() => MeRequest._();
  @$core.pragma('dart2js:noInline')
  static MeRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<MeRequest>(MeRequest.$_createMessage);
  static MeRequest? _defaultInstance;
}

class GetAuthInfoRequest extends $pb.GeneratedMessage {
  factory GetAuthInfoRequest() => GetAuthInfoRequest._();

  GetAuthInfoRequest._();

  factory GetAuthInfoRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetAuthInfoRequest()..mergeFromBuffer(data, registry);
  factory GetAuthInfoRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetAuthInfoRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetAuthInfoRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'fileshare.v1'),
      createEmptyInstance: GetAuthInfoRequest.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetAuthInfoRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetAuthInfoRequest copyWith(void Function(GetAuthInfoRequest) updates) =>
      super.copyWith((message) => updates(message as GetAuthInfoRequest))
          as GetAuthInfoRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use GetAuthInfoRequest() / GetAuthInfoRequest.new instead')
  static GetAuthInfoRequest create() => GetAuthInfoRequest._();
  static $pb.GeneratedMessage $_createMessage() => GetAuthInfoRequest._();
  @$core.override
  GetAuthInfoRequest createEmptyInstance() => GetAuthInfoRequest._();
  @$core.pragma('dart2js:noInline')
  static GetAuthInfoRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetAuthInfoRequest>(
          GetAuthInfoRequest.$_createMessage);
  static GetAuthInfoRequest? _defaultInstance;
}

class ListUsersRequest extends $pb.GeneratedMessage {
  factory ListUsersRequest() => ListUsersRequest._();

  ListUsersRequest._();

  factory ListUsersRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListUsersRequest()..mergeFromBuffer(data, registry);
  factory ListUsersRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListUsersRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListUsersRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'fileshare.v1'),
      createEmptyInstance: ListUsersRequest.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListUsersRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListUsersRequest copyWith(void Function(ListUsersRequest) updates) =>
      super.copyWith((message) => updates(message as ListUsersRequest))
          as ListUsersRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ListUsersRequest() / ListUsersRequest.new instead')
  static ListUsersRequest create() => ListUsersRequest._();
  static $pb.GeneratedMessage $_createMessage() => ListUsersRequest._();
  @$core.override
  ListUsersRequest createEmptyInstance() => ListUsersRequest._();
  @$core.pragma('dart2js:noInline')
  static ListUsersRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ListUsersRequest>(
          ListUsersRequest.$_createMessage);
  static ListUsersRequest? _defaultInstance;
}

class UserStats extends $pb.GeneratedMessage {
  factory UserStats({
    $core.String? username,
    $fixnum.Int64? fileCount,
    $fixnum.Int64? totalSize,
    $core.bool? isAdmin,
    $core.String? role,
    $core.bool? enabled,
    $fixnum.Int64? quotaBytes,
  }) {
    final result = UserStats._();
    if (username != null) result.username = username;
    if (fileCount != null) result.fileCount = fileCount;
    if (totalSize != null) result.totalSize = totalSize;
    if (isAdmin != null) result.isAdmin = isAdmin;
    if (role != null) result.role = role;
    if (enabled != null) result.enabled = enabled;
    if (quotaBytes != null) result.quotaBytes = quotaBytes;
    return result;
  }

  UserStats._();

  factory UserStats.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UserStats()..mergeFromBuffer(data, registry);
  factory UserStats.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UserStats()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UserStats',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'fileshare.v1'),
      createEmptyInstance: UserStats.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'username')
    ..aInt64(2, _omitFieldNames ? '' : 'fileCount')
    ..aInt64(3, _omitFieldNames ? '' : 'totalSize')
    ..aOB(4, _omitFieldNames ? '' : 'isAdmin')
    ..aOS(5, _omitFieldNames ? '' : 'role')
    ..aOB(6, _omitFieldNames ? '' : 'enabled')
    ..aInt64(7, _omitFieldNames ? '' : 'quotaBytes')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UserStats clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UserStats copyWith(void Function(UserStats) updates) =>
      super.copyWith((message) => updates(message as UserStats)) as UserStats;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use UserStats() / UserStats.new instead')
  static UserStats create() => UserStats._();
  static $pb.GeneratedMessage $_createMessage() => UserStats._();
  @$core.override
  UserStats createEmptyInstance() => UserStats._();
  @$core.pragma('dart2js:noInline')
  static UserStats getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UserStats>(UserStats.$_createMessage);
  static UserStats? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get username => $_getSZ(0);
  @$pb.TagNumber(1)
  set username($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasUsername() => $_has(0);
  @$pb.TagNumber(1)
  void clearUsername() => $_clearField(1);

  @$pb.TagNumber(2)
  $fixnum.Int64 get fileCount => $_getI64(1);
  @$pb.TagNumber(2)
  set fileCount($fixnum.Int64 value) => $_setInt64(1, value);
  @$pb.TagNumber(2)
  $core.bool hasFileCount() => $_has(1);
  @$pb.TagNumber(2)
  void clearFileCount() => $_clearField(2);

  @$pb.TagNumber(3)
  $fixnum.Int64 get totalSize => $_getI64(2);
  @$pb.TagNumber(3)
  set totalSize($fixnum.Int64 value) => $_setInt64(2, value);
  @$pb.TagNumber(3)
  $core.bool hasTotalSize() => $_has(2);
  @$pb.TagNumber(3)
  void clearTotalSize() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.bool get isAdmin => $_getBF(3);
  @$pb.TagNumber(4)
  set isAdmin($core.bool value) => $_setBool(3, value);
  @$pb.TagNumber(4)
  $core.bool hasIsAdmin() => $_has(3);
  @$pb.TagNumber(4)
  void clearIsAdmin() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get role => $_getSZ(4);
  @$pb.TagNumber(5)
  set role($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasRole() => $_has(4);
  @$pb.TagNumber(5)
  void clearRole() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.bool get enabled => $_getBF(5);
  @$pb.TagNumber(6)
  set enabled($core.bool value) => $_setBool(5, value);
  @$pb.TagNumber(6)
  $core.bool hasEnabled() => $_has(5);
  @$pb.TagNumber(6)
  void clearEnabled() => $_clearField(6);

  @$pb.TagNumber(7)
  $fixnum.Int64 get quotaBytes => $_getI64(6);
  @$pb.TagNumber(7)
  set quotaBytes($fixnum.Int64 value) => $_setInt64(6, value);
  @$pb.TagNumber(7)
  $core.bool hasQuotaBytes() => $_has(6);
  @$pb.TagNumber(7)
  void clearQuotaBytes() => $_clearField(7);
}

class ListUsersResponse extends $pb.GeneratedMessage {
  factory ListUsersResponse({
    $core.Iterable<UserStats>? users,
  }) {
    final result = ListUsersResponse._();
    if (users != null) result.users.addAll(users);
    return result;
  }

  ListUsersResponse._();

  factory ListUsersResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListUsersResponse()..mergeFromBuffer(data, registry);
  factory ListUsersResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListUsersResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListUsersResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'fileshare.v1'),
      createEmptyInstance: ListUsersResponse.$_createMessage)
    ..pPM<UserStats>(1, _omitFieldNames ? '' : 'users',
        subBuilder: UserStats.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListUsersResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListUsersResponse copyWith(void Function(ListUsersResponse) updates) =>
      super.copyWith((message) => updates(message as ListUsersResponse))
          as ListUsersResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ListUsersResponse() / ListUsersResponse.new instead')
  static ListUsersResponse create() => ListUsersResponse._();
  static $pb.GeneratedMessage $_createMessage() => ListUsersResponse._();
  @$core.override
  ListUsersResponse createEmptyInstance() => ListUsersResponse._();
  @$core.pragma('dart2js:noInline')
  static ListUsersResponse getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ListUsersResponse>(
          ListUsersResponse.$_createMessage);
  static ListUsersResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<UserStats> get users => $_getList(0);
}

class LoginRequest extends $pb.GeneratedMessage {
  factory LoginRequest({
    $core.String? username,
    $core.String? password,
  }) {
    final result = LoginRequest._();
    if (username != null) result.username = username;
    if (password != null) result.password = password;
    return result;
  }

  LoginRequest._();

  factory LoginRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      LoginRequest()..mergeFromBuffer(data, registry);
  factory LoginRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      LoginRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'LoginRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'fileshare.v1'),
      createEmptyInstance: LoginRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'username')
    ..aOS(2, _omitFieldNames ? '' : 'password')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LoginRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LoginRequest copyWith(void Function(LoginRequest) updates) =>
      super.copyWith((message) => updates(message as LoginRequest))
          as LoginRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use LoginRequest() / LoginRequest.new instead')
  static LoginRequest create() => LoginRequest._();
  static $pb.GeneratedMessage $_createMessage() => LoginRequest._();
  @$core.override
  LoginRequest createEmptyInstance() => LoginRequest._();
  @$core.pragma('dart2js:noInline')
  static LoginRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<LoginRequest>(
          LoginRequest.$_createMessage);
  static LoginRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get username => $_getSZ(0);
  @$pb.TagNumber(1)
  set username($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasUsername() => $_has(0);
  @$pb.TagNumber(1)
  void clearUsername() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get password => $_getSZ(1);
  @$pb.TagNumber(2)
  set password($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasPassword() => $_has(1);
  @$pb.TagNumber(2)
  void clearPassword() => $_clearField(2);
}

class LoginResponse extends $pb.GeneratedMessage {
  factory LoginResponse({
    $core.String? accessToken,
    $core.String? tokenType,
    $fixnum.Int64? expiresIn,
  }) {
    final result = LoginResponse._();
    if (accessToken != null) result.accessToken = accessToken;
    if (tokenType != null) result.tokenType = tokenType;
    if (expiresIn != null) result.expiresIn = expiresIn;
    return result;
  }

  LoginResponse._();

  factory LoginResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      LoginResponse()..mergeFromBuffer(data, registry);
  factory LoginResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      LoginResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'LoginResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'fileshare.v1'),
      createEmptyInstance: LoginResponse.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'accessToken')
    ..aOS(2, _omitFieldNames ? '' : 'tokenType')
    ..aInt64(3, _omitFieldNames ? '' : 'expiresIn')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LoginResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LoginResponse copyWith(void Function(LoginResponse) updates) =>
      super.copyWith((message) => updates(message as LoginResponse))
          as LoginResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use LoginResponse() / LoginResponse.new instead')
  static LoginResponse create() => LoginResponse._();
  static $pb.GeneratedMessage $_createMessage() => LoginResponse._();
  @$core.override
  LoginResponse createEmptyInstance() => LoginResponse._();
  @$core.pragma('dart2js:noInline')
  static LoginResponse getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<LoginResponse>(
          LoginResponse.$_createMessage);
  static LoginResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get accessToken => $_getSZ(0);
  @$pb.TagNumber(1)
  set accessToken($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAccessToken() => $_has(0);
  @$pb.TagNumber(1)
  void clearAccessToken() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get tokenType => $_getSZ(1);
  @$pb.TagNumber(2)
  set tokenType($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasTokenType() => $_has(1);
  @$pb.TagNumber(2)
  void clearTokenType() => $_clearField(2);

  @$pb.TagNumber(3)
  $fixnum.Int64 get expiresIn => $_getI64(2);
  @$pb.TagNumber(3)
  set expiresIn($fixnum.Int64 value) => $_setInt64(2, value);
  @$pb.TagNumber(3)
  $core.bool hasExpiresIn() => $_has(2);
  @$pb.TagNumber(3)
  void clearExpiresIn() => $_clearField(3);
}

class LogoutRequest extends $pb.GeneratedMessage {
  factory LogoutRequest() => LogoutRequest._();

  LogoutRequest._();

  factory LogoutRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      LogoutRequest()..mergeFromBuffer(data, registry);
  factory LogoutRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      LogoutRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'LogoutRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'fileshare.v1'),
      createEmptyInstance: LogoutRequest.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LogoutRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LogoutRequest copyWith(void Function(LogoutRequest) updates) =>
      super.copyWith((message) => updates(message as LogoutRequest))
          as LogoutRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use LogoutRequest() / LogoutRequest.new instead')
  static LogoutRequest create() => LogoutRequest._();
  static $pb.GeneratedMessage $_createMessage() => LogoutRequest._();
  @$core.override
  LogoutRequest createEmptyInstance() => LogoutRequest._();
  @$core.pragma('dart2js:noInline')
  static LogoutRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<LogoutRequest>(
          LogoutRequest.$_createMessage);
  static LogoutRequest? _defaultInstance;
}

class LogoutResponse extends $pb.GeneratedMessage {
  factory LogoutResponse({
    $core.String? status,
  }) {
    final result = LogoutResponse._();
    if (status != null) result.status = status;
    return result;
  }

  LogoutResponse._();

  factory LogoutResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      LogoutResponse()..mergeFromBuffer(data, registry);
  factory LogoutResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      LogoutResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'LogoutResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'fileshare.v1'),
      createEmptyInstance: LogoutResponse.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'status')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LogoutResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LogoutResponse copyWith(void Function(LogoutResponse) updates) =>
      super.copyWith((message) => updates(message as LogoutResponse))
          as LogoutResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use LogoutResponse() / LogoutResponse.new instead')
  static LogoutResponse create() => LogoutResponse._();
  static $pb.GeneratedMessage $_createMessage() => LogoutResponse._();
  @$core.override
  LogoutResponse createEmptyInstance() => LogoutResponse._();
  @$core.pragma('dart2js:noInline')
  static LogoutResponse getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<LogoutResponse>(
          LogoutResponse.$_createMessage);
  static LogoutResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get status => $_getSZ(0);
  @$pb.TagNumber(1)
  set status($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasStatus() => $_has(0);
  @$pb.TagNumber(1)
  void clearStatus() => $_clearField(1);
}

class CreateShareRequest extends $pb.GeneratedMessage {
  factory CreateShareRequest({
    $core.String? path,
    $fixnum.Int64? expiresInSeconds,
  }) {
    final result = CreateShareRequest._();
    if (path != null) result.path = path;
    if (expiresInSeconds != null) result.expiresInSeconds = expiresInSeconds;
    return result;
  }

  CreateShareRequest._();

  factory CreateShareRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateShareRequest()..mergeFromBuffer(data, registry);
  factory CreateShareRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateShareRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateShareRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'fileshare.v1'),
      createEmptyInstance: CreateShareRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'path')
    ..aInt64(2, _omitFieldNames ? '' : 'expiresInSeconds')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateShareRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateShareRequest copyWith(void Function(CreateShareRequest) updates) =>
      super.copyWith((message) => updates(message as CreateShareRequest))
          as CreateShareRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use CreateShareRequest() / CreateShareRequest.new instead')
  static CreateShareRequest create() => CreateShareRequest._();
  static $pb.GeneratedMessage $_createMessage() => CreateShareRequest._();
  @$core.override
  CreateShareRequest createEmptyInstance() => CreateShareRequest._();
  @$core.pragma('dart2js:noInline')
  static CreateShareRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateShareRequest>(
          CreateShareRequest.$_createMessage);
  static CreateShareRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get path => $_getSZ(0);
  @$pb.TagNumber(1)
  set path($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasPath() => $_has(0);
  @$pb.TagNumber(1)
  void clearPath() => $_clearField(1);

  @$pb.TagNumber(2)
  $fixnum.Int64 get expiresInSeconds => $_getI64(1);
  @$pb.TagNumber(2)
  set expiresInSeconds($fixnum.Int64 value) => $_setInt64(1, value);
  @$pb.TagNumber(2)
  $core.bool hasExpiresInSeconds() => $_has(1);
  @$pb.TagNumber(2)
  void clearExpiresInSeconds() => $_clearField(2);
}

class Share extends $pb.GeneratedMessage {
  factory Share({
    $core.String? token,
    $core.String? path,
    $core.String? name,
    $core.String? owner,
    $core.String? createdAt,
    $core.String? expiresAt,
  }) {
    final result = Share._();
    if (token != null) result.token = token;
    if (path != null) result.path = path;
    if (name != null) result.name = name;
    if (owner != null) result.owner = owner;
    if (createdAt != null) result.createdAt = createdAt;
    if (expiresAt != null) result.expiresAt = expiresAt;
    return result;
  }

  Share._();

  factory Share.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Share()..mergeFromBuffer(data, registry);
  factory Share.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Share()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Share',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'fileshare.v1'),
      createEmptyInstance: Share.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'token')
    ..aOS(2, _omitFieldNames ? '' : 'path')
    ..aOS(3, _omitFieldNames ? '' : 'name')
    ..aOS(4, _omitFieldNames ? '' : 'owner')
    ..aOS(5, _omitFieldNames ? '' : 'createdAt')
    ..aOS(6, _omitFieldNames ? '' : 'expiresAt')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Share clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Share copyWith(void Function(Share) updates) =>
      super.copyWith((message) => updates(message as Share)) as Share;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use Share() / Share.new instead')
  static Share create() => Share._();
  static $pb.GeneratedMessage $_createMessage() => Share._();
  @$core.override
  Share createEmptyInstance() => Share._();
  @$core.pragma('dart2js:noInline')
  static Share getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<Share>(Share.$_createMessage);
  static Share? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get token => $_getSZ(0);
  @$pb.TagNumber(1)
  set token($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasToken() => $_has(0);
  @$pb.TagNumber(1)
  void clearToken() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get path => $_getSZ(1);
  @$pb.TagNumber(2)
  set path($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasPath() => $_has(1);
  @$pb.TagNumber(2)
  void clearPath() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get name => $_getSZ(2);
  @$pb.TagNumber(3)
  set name($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasName() => $_has(2);
  @$pb.TagNumber(3)
  void clearName() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get owner => $_getSZ(3);
  @$pb.TagNumber(4)
  set owner($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasOwner() => $_has(3);
  @$pb.TagNumber(4)
  void clearOwner() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get createdAt => $_getSZ(4);
  @$pb.TagNumber(5)
  set createdAt($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasCreatedAt() => $_has(4);
  @$pb.TagNumber(5)
  void clearCreatedAt() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get expiresAt => $_getSZ(5);
  @$pb.TagNumber(6)
  set expiresAt($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasExpiresAt() => $_has(5);
  @$pb.TagNumber(6)
  void clearExpiresAt() => $_clearField(6);
}

class ListSharesRequest extends $pb.GeneratedMessage {
  factory ListSharesRequest() => ListSharesRequest._();

  ListSharesRequest._();

  factory ListSharesRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListSharesRequest()..mergeFromBuffer(data, registry);
  factory ListSharesRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListSharesRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListSharesRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'fileshare.v1'),
      createEmptyInstance: ListSharesRequest.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListSharesRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListSharesRequest copyWith(void Function(ListSharesRequest) updates) =>
      super.copyWith((message) => updates(message as ListSharesRequest))
          as ListSharesRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ListSharesRequest() / ListSharesRequest.new instead')
  static ListSharesRequest create() => ListSharesRequest._();
  static $pb.GeneratedMessage $_createMessage() => ListSharesRequest._();
  @$core.override
  ListSharesRequest createEmptyInstance() => ListSharesRequest._();
  @$core.pragma('dart2js:noInline')
  static ListSharesRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ListSharesRequest>(
          ListSharesRequest.$_createMessage);
  static ListSharesRequest? _defaultInstance;
}

class ListSharesResponse extends $pb.GeneratedMessage {
  factory ListSharesResponse({
    $core.Iterable<Share>? shares,
  }) {
    final result = ListSharesResponse._();
    if (shares != null) result.shares.addAll(shares);
    return result;
  }

  ListSharesResponse._();

  factory ListSharesResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListSharesResponse()..mergeFromBuffer(data, registry);
  factory ListSharesResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListSharesResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListSharesResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'fileshare.v1'),
      createEmptyInstance: ListSharesResponse.$_createMessage)
    ..pPM<Share>(1, _omitFieldNames ? '' : 'shares',
        subBuilder: Share.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListSharesResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListSharesResponse copyWith(void Function(ListSharesResponse) updates) =>
      super.copyWith((message) => updates(message as ListSharesResponse))
          as ListSharesResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ListSharesResponse() / ListSharesResponse.new instead')
  static ListSharesResponse create() => ListSharesResponse._();
  static $pb.GeneratedMessage $_createMessage() => ListSharesResponse._();
  @$core.override
  ListSharesResponse createEmptyInstance() => ListSharesResponse._();
  @$core.pragma('dart2js:noInline')
  static ListSharesResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListSharesResponse>(
          ListSharesResponse.$_createMessage);
  static ListSharesResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<Share> get shares => $_getList(0);
}

class RevokeShareRequest extends $pb.GeneratedMessage {
  factory RevokeShareRequest({
    $core.String? token,
  }) {
    final result = RevokeShareRequest._();
    if (token != null) result.token = token;
    return result;
  }

  RevokeShareRequest._();

  factory RevokeShareRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RevokeShareRequest()..mergeFromBuffer(data, registry);
  factory RevokeShareRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RevokeShareRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RevokeShareRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'fileshare.v1'),
      createEmptyInstance: RevokeShareRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'token')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RevokeShareRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RevokeShareRequest copyWith(void Function(RevokeShareRequest) updates) =>
      super.copyWith((message) => updates(message as RevokeShareRequest))
          as RevokeShareRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use RevokeShareRequest() / RevokeShareRequest.new instead')
  static RevokeShareRequest create() => RevokeShareRequest._();
  static $pb.GeneratedMessage $_createMessage() => RevokeShareRequest._();
  @$core.override
  RevokeShareRequest createEmptyInstance() => RevokeShareRequest._();
  @$core.pragma('dart2js:noInline')
  static RevokeShareRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RevokeShareRequest>(
          RevokeShareRequest.$_createMessage);
  static RevokeShareRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get token => $_getSZ(0);
  @$pb.TagNumber(1)
  set token($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasToken() => $_has(0);
  @$pb.TagNumber(1)
  void clearToken() => $_clearField(1);
}

class RevokeShareResponse extends $pb.GeneratedMessage {
  factory RevokeShareResponse({
    $core.String? message,
  }) {
    final result = RevokeShareResponse._();
    if (message != null) result.message = message;
    return result;
  }

  RevokeShareResponse._();

  factory RevokeShareResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RevokeShareResponse()..mergeFromBuffer(data, registry);
  factory RevokeShareResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RevokeShareResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RevokeShareResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'fileshare.v1'),
      createEmptyInstance: RevokeShareResponse.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'message')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RevokeShareResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RevokeShareResponse copyWith(void Function(RevokeShareResponse) updates) =>
      super.copyWith((message) => updates(message as RevokeShareResponse))
          as RevokeShareResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use RevokeShareResponse() / RevokeShareResponse.new instead')
  static RevokeShareResponse create() => RevokeShareResponse._();
  static $pb.GeneratedMessage $_createMessage() => RevokeShareResponse._();
  @$core.override
  RevokeShareResponse createEmptyInstance() => RevokeShareResponse._();
  @$core.pragma('dart2js:noInline')
  static RevokeShareResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RevokeShareResponse>(
          RevokeShareResponse.$_createMessage);
  static RevokeShareResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get message => $_getSZ(0);
  @$pb.TagNumber(1)
  set message($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasMessage() => $_has(0);
  @$pb.TagNumber(1)
  void clearMessage() => $_clearField(1);
}

class SubscribeRequest extends $pb.GeneratedMessage {
  factory SubscribeRequest() => SubscribeRequest._();

  SubscribeRequest._();

  factory SubscribeRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SubscribeRequest()..mergeFromBuffer(data, registry);
  factory SubscribeRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SubscribeRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SubscribeRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'fileshare.v1'),
      createEmptyInstance: SubscribeRequest.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SubscribeRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SubscribeRequest copyWith(void Function(SubscribeRequest) updates) =>
      super.copyWith((message) => updates(message as SubscribeRequest))
          as SubscribeRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use SubscribeRequest() / SubscribeRequest.new instead')
  static SubscribeRequest create() => SubscribeRequest._();
  static $pb.GeneratedMessage $_createMessage() => SubscribeRequest._();
  @$core.override
  SubscribeRequest createEmptyInstance() => SubscribeRequest._();
  @$core.pragma('dart2js:noInline')
  static SubscribeRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<SubscribeRequest>(
          SubscribeRequest.$_createMessage);
  static SubscribeRequest? _defaultInstance;
}

class EventMessage extends $pb.GeneratedMessage {
  factory EventMessage({
    $core.String? type,
    $core.String? path,
    $core.String? user,
    $core.String? at,
    $fixnum.Int64? bytes,
  }) {
    final result = EventMessage._();
    if (type != null) result.type = type;
    if (path != null) result.path = path;
    if (user != null) result.user = user;
    if (at != null) result.at = at;
    if (bytes != null) result.bytes = bytes;
    return result;
  }

  EventMessage._();

  factory EventMessage.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      EventMessage()..mergeFromBuffer(data, registry);
  factory EventMessage.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      EventMessage()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'EventMessage',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'fileshare.v1'),
      createEmptyInstance: EventMessage.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'type')
    ..aOS(2, _omitFieldNames ? '' : 'path')
    ..aOS(3, _omitFieldNames ? '' : 'user')
    ..aOS(4, _omitFieldNames ? '' : 'at')
    ..aInt64(5, _omitFieldNames ? '' : 'bytes')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  EventMessage clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  EventMessage copyWith(void Function(EventMessage) updates) =>
      super.copyWith((message) => updates(message as EventMessage))
          as EventMessage;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use EventMessage() / EventMessage.new instead')
  static EventMessage create() => EventMessage._();
  static $pb.GeneratedMessage $_createMessage() => EventMessage._();
  @$core.override
  EventMessage createEmptyInstance() => EventMessage._();
  @$core.pragma('dart2js:noInline')
  static EventMessage getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<EventMessage>(
          EventMessage.$_createMessage);
  static EventMessage? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get type => $_getSZ(0);
  @$pb.TagNumber(1)
  set type($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasType() => $_has(0);
  @$pb.TagNumber(1)
  void clearType() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get path => $_getSZ(1);
  @$pb.TagNumber(2)
  set path($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasPath() => $_has(1);
  @$pb.TagNumber(2)
  void clearPath() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get user => $_getSZ(2);
  @$pb.TagNumber(3)
  set user($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasUser() => $_has(2);
  @$pb.TagNumber(3)
  void clearUser() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get at => $_getSZ(3);
  @$pb.TagNumber(4)
  set at($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasAt() => $_has(3);
  @$pb.TagNumber(4)
  void clearAt() => $_clearField(4);

  @$pb.TagNumber(5)
  $fixnum.Int64 get bytes => $_getI64(4);
  @$pb.TagNumber(5)
  set bytes($fixnum.Int64 value) => $_setInt64(4, value);
  @$pb.TagNumber(5)
  $core.bool hasBytes() => $_has(4);
  @$pb.TagNumber(5)
  void clearBytes() => $_clearField(5);
}

class FileInfo extends $pb.GeneratedMessage {
  factory FileInfo({
    $core.String? name,
    $core.String? path,
    $fixnum.Int64? size,
    $core.bool? isDir,
    $1.Timestamp? modified,
  }) {
    final result = FileInfo._();
    if (name != null) result.name = name;
    if (path != null) result.path = path;
    if (size != null) result.size = size;
    if (isDir != null) result.isDir = isDir;
    if (modified != null) result.modified = modified;
    return result;
  }

  FileInfo._();

  factory FileInfo.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      FileInfo()..mergeFromBuffer(data, registry);
  factory FileInfo.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      FileInfo()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'FileInfo',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'fileshare.v1'),
      createEmptyInstance: FileInfo.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..aOS(2, _omitFieldNames ? '' : 'path')
    ..aInt64(3, _omitFieldNames ? '' : 'size')
    ..aOB(4, _omitFieldNames ? '' : 'isDir')
    ..aOM<$1.Timestamp>(5, _omitFieldNames ? '' : 'modified',
        subBuilder: $1.Timestamp.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  FileInfo clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  FileInfo copyWith(void Function(FileInfo) updates) =>
      super.copyWith((message) => updates(message as FileInfo)) as FileInfo;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use FileInfo() / FileInfo.new instead')
  static FileInfo create() => FileInfo._();
  static $pb.GeneratedMessage $_createMessage() => FileInfo._();
  @$core.override
  FileInfo createEmptyInstance() => FileInfo._();
  @$core.pragma('dart2js:noInline')
  static FileInfo getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<FileInfo>(FileInfo.$_createMessage);
  static FileInfo? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get path => $_getSZ(1);
  @$pb.TagNumber(2)
  set path($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasPath() => $_has(1);
  @$pb.TagNumber(2)
  void clearPath() => $_clearField(2);

  @$pb.TagNumber(3)
  $fixnum.Int64 get size => $_getI64(2);
  @$pb.TagNumber(3)
  set size($fixnum.Int64 value) => $_setInt64(2, value);
  @$pb.TagNumber(3)
  $core.bool hasSize() => $_has(2);
  @$pb.TagNumber(3)
  void clearSize() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.bool get isDir => $_getBF(3);
  @$pb.TagNumber(4)
  set isDir($core.bool value) => $_setBool(3, value);
  @$pb.TagNumber(4)
  $core.bool hasIsDir() => $_has(3);
  @$pb.TagNumber(4)
  void clearIsDir() => $_clearField(4);

  @$pb.TagNumber(5)
  $1.Timestamp get modified => $_getN(4);
  @$pb.TagNumber(5)
  set modified($1.Timestamp value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasModified() => $_has(4);
  @$pb.TagNumber(5)
  void clearModified() => $_clearField(5);
  @$pb.TagNumber(5)
  $1.Timestamp ensureModified() => $_ensure(4);
}

class ListFilesRequest extends $pb.GeneratedMessage {
  factory ListFilesRequest({
    $core.String? path,
  }) {
    final result = ListFilesRequest._();
    if (path != null) result.path = path;
    return result;
  }

  ListFilesRequest._();

  factory ListFilesRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListFilesRequest()..mergeFromBuffer(data, registry);
  factory ListFilesRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListFilesRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListFilesRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'fileshare.v1'),
      createEmptyInstance: ListFilesRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'path')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListFilesRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListFilesRequest copyWith(void Function(ListFilesRequest) updates) =>
      super.copyWith((message) => updates(message as ListFilesRequest))
          as ListFilesRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ListFilesRequest() / ListFilesRequest.new instead')
  static ListFilesRequest create() => ListFilesRequest._();
  static $pb.GeneratedMessage $_createMessage() => ListFilesRequest._();
  @$core.override
  ListFilesRequest createEmptyInstance() => ListFilesRequest._();
  @$core.pragma('dart2js:noInline')
  static ListFilesRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ListFilesRequest>(
          ListFilesRequest.$_createMessage);
  static ListFilesRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get path => $_getSZ(0);
  @$pb.TagNumber(1)
  set path($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasPath() => $_has(0);
  @$pb.TagNumber(1)
  void clearPath() => $_clearField(1);
}

class ListFilesResponse extends $pb.GeneratedMessage {
  factory ListFilesResponse({
    $core.Iterable<FileInfo>? files,
  }) {
    final result = ListFilesResponse._();
    if (files != null) result.files.addAll(files);
    return result;
  }

  ListFilesResponse._();

  factory ListFilesResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListFilesResponse()..mergeFromBuffer(data, registry);
  factory ListFilesResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListFilesResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListFilesResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'fileshare.v1'),
      createEmptyInstance: ListFilesResponse.$_createMessage)
    ..pPM<FileInfo>(1, _omitFieldNames ? '' : 'files',
        subBuilder: FileInfo.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListFilesResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListFilesResponse copyWith(void Function(ListFilesResponse) updates) =>
      super.copyWith((message) => updates(message as ListFilesResponse))
          as ListFilesResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ListFilesResponse() / ListFilesResponse.new instead')
  static ListFilesResponse create() => ListFilesResponse._();
  static $pb.GeneratedMessage $_createMessage() => ListFilesResponse._();
  @$core.override
  ListFilesResponse createEmptyInstance() => ListFilesResponse._();
  @$core.pragma('dart2js:noInline')
  static ListFilesResponse getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ListFilesResponse>(
          ListFilesResponse.$_createMessage);
  static ListFilesResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<FileInfo> get files => $_getList(0);
}

class GetFileInfoRequest extends $pb.GeneratedMessage {
  factory GetFileInfoRequest({
    $core.String? path,
  }) {
    final result = GetFileInfoRequest._();
    if (path != null) result.path = path;
    return result;
  }

  GetFileInfoRequest._();

  factory GetFileInfoRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetFileInfoRequest()..mergeFromBuffer(data, registry);
  factory GetFileInfoRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetFileInfoRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetFileInfoRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'fileshare.v1'),
      createEmptyInstance: GetFileInfoRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'path')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetFileInfoRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetFileInfoRequest copyWith(void Function(GetFileInfoRequest) updates) =>
      super.copyWith((message) => updates(message as GetFileInfoRequest))
          as GetFileInfoRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use GetFileInfoRequest() / GetFileInfoRequest.new instead')
  static GetFileInfoRequest create() => GetFileInfoRequest._();
  static $pb.GeneratedMessage $_createMessage() => GetFileInfoRequest._();
  @$core.override
  GetFileInfoRequest createEmptyInstance() => GetFileInfoRequest._();
  @$core.pragma('dart2js:noInline')
  static GetFileInfoRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetFileInfoRequest>(
          GetFileInfoRequest.$_createMessage);
  static GetFileInfoRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get path => $_getSZ(0);
  @$pb.TagNumber(1)
  set path($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasPath() => $_has(0);
  @$pb.TagNumber(1)
  void clearPath() => $_clearField(1);
}

class SearchFilesRequest extends $pb.GeneratedMessage {
  factory SearchFilesRequest({
    $core.String? query,
    $core.int? limit,
  }) {
    final result = SearchFilesRequest._();
    if (query != null) result.query = query;
    if (limit != null) result.limit = limit;
    return result;
  }

  SearchFilesRequest._();

  factory SearchFilesRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SearchFilesRequest()..mergeFromBuffer(data, registry);
  factory SearchFilesRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SearchFilesRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SearchFilesRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'fileshare.v1'),
      createEmptyInstance: SearchFilesRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'query')
    ..aI(2, _omitFieldNames ? '' : 'limit')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SearchFilesRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SearchFilesRequest copyWith(void Function(SearchFilesRequest) updates) =>
      super.copyWith((message) => updates(message as SearchFilesRequest))
          as SearchFilesRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use SearchFilesRequest() / SearchFilesRequest.new instead')
  static SearchFilesRequest create() => SearchFilesRequest._();
  static $pb.GeneratedMessage $_createMessage() => SearchFilesRequest._();
  @$core.override
  SearchFilesRequest createEmptyInstance() => SearchFilesRequest._();
  @$core.pragma('dart2js:noInline')
  static SearchFilesRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SearchFilesRequest>(
          SearchFilesRequest.$_createMessage);
  static SearchFilesRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get query => $_getSZ(0);
  @$pb.TagNumber(1)
  set query($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasQuery() => $_has(0);
  @$pb.TagNumber(1)
  void clearQuery() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.int get limit => $_getIZ(1);
  @$pb.TagNumber(2)
  set limit($core.int value) => $_setSignedInt32(1, value);
  @$pb.TagNumber(2)
  $core.bool hasLimit() => $_has(1);
  @$pb.TagNumber(2)
  void clearLimit() => $_clearField(2);
}

class SearchFilesResponse extends $pb.GeneratedMessage {
  factory SearchFilesResponse({
    $core.Iterable<FileInfo>? files,
  }) {
    final result = SearchFilesResponse._();
    if (files != null) result.files.addAll(files);
    return result;
  }

  SearchFilesResponse._();

  factory SearchFilesResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SearchFilesResponse()..mergeFromBuffer(data, registry);
  factory SearchFilesResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SearchFilesResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SearchFilesResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'fileshare.v1'),
      createEmptyInstance: SearchFilesResponse.$_createMessage)
    ..pPM<FileInfo>(1, _omitFieldNames ? '' : 'files',
        subBuilder: FileInfo.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SearchFilesResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SearchFilesResponse copyWith(void Function(SearchFilesResponse) updates) =>
      super.copyWith((message) => updates(message as SearchFilesResponse))
          as SearchFilesResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use SearchFilesResponse() / SearchFilesResponse.new instead')
  static SearchFilesResponse create() => SearchFilesResponse._();
  static $pb.GeneratedMessage $_createMessage() => SearchFilesResponse._();
  @$core.override
  SearchFilesResponse createEmptyInstance() => SearchFilesResponse._();
  @$core.pragma('dart2js:noInline')
  static SearchFilesResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SearchFilesResponse>(
          SearchFilesResponse.$_createMessage);
  static SearchFilesResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<FileInfo> get files => $_getList(0);
}

class CreateDirectoryRequest extends $pb.GeneratedMessage {
  factory CreateDirectoryRequest({
    $core.String? path,
  }) {
    final result = CreateDirectoryRequest._();
    if (path != null) result.path = path;
    return result;
  }

  CreateDirectoryRequest._();

  factory CreateDirectoryRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateDirectoryRequest()..mergeFromBuffer(data, registry);
  factory CreateDirectoryRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateDirectoryRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateDirectoryRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'fileshare.v1'),
      createEmptyInstance: CreateDirectoryRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'path')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateDirectoryRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateDirectoryRequest copyWith(
          void Function(CreateDirectoryRequest) updates) =>
      super.copyWith((message) => updates(message as CreateDirectoryRequest))
          as CreateDirectoryRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateDirectoryRequest() / CreateDirectoryRequest.new instead')
  static CreateDirectoryRequest create() => CreateDirectoryRequest._();
  static $pb.GeneratedMessage $_createMessage() => CreateDirectoryRequest._();
  @$core.override
  CreateDirectoryRequest createEmptyInstance() => CreateDirectoryRequest._();
  @$core.pragma('dart2js:noInline')
  static CreateDirectoryRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateDirectoryRequest>(
          CreateDirectoryRequest.$_createMessage);
  static CreateDirectoryRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get path => $_getSZ(0);
  @$pb.TagNumber(1)
  set path($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasPath() => $_has(0);
  @$pb.TagNumber(1)
  void clearPath() => $_clearField(1);
}

class CreateDirectoryResponse extends $pb.GeneratedMessage {
  factory CreateDirectoryResponse({
    $core.String? message,
  }) {
    final result = CreateDirectoryResponse._();
    if (message != null) result.message = message;
    return result;
  }

  CreateDirectoryResponse._();

  factory CreateDirectoryResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateDirectoryResponse()..mergeFromBuffer(data, registry);
  factory CreateDirectoryResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateDirectoryResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateDirectoryResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'fileshare.v1'),
      createEmptyInstance: CreateDirectoryResponse.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'message')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateDirectoryResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateDirectoryResponse copyWith(
          void Function(CreateDirectoryResponse) updates) =>
      super.copyWith((message) => updates(message as CreateDirectoryResponse))
          as CreateDirectoryResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateDirectoryResponse() / CreateDirectoryResponse.new instead')
  static CreateDirectoryResponse create() => CreateDirectoryResponse._();
  static $pb.GeneratedMessage $_createMessage() => CreateDirectoryResponse._();
  @$core.override
  CreateDirectoryResponse createEmptyInstance() => CreateDirectoryResponse._();
  @$core.pragma('dart2js:noInline')
  static CreateDirectoryResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateDirectoryResponse>(
          CreateDirectoryResponse.$_createMessage);
  static CreateDirectoryResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get message => $_getSZ(0);
  @$pb.TagNumber(1)
  set message($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasMessage() => $_has(0);
  @$pb.TagNumber(1)
  void clearMessage() => $_clearField(1);
}

class DeletePathRequest extends $pb.GeneratedMessage {
  factory DeletePathRequest({
    $core.String? path,
  }) {
    final result = DeletePathRequest._();
    if (path != null) result.path = path;
    return result;
  }

  DeletePathRequest._();

  factory DeletePathRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeletePathRequest()..mergeFromBuffer(data, registry);
  factory DeletePathRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeletePathRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeletePathRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'fileshare.v1'),
      createEmptyInstance: DeletePathRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'path')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeletePathRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeletePathRequest copyWith(void Function(DeletePathRequest) updates) =>
      super.copyWith((message) => updates(message as DeletePathRequest))
          as DeletePathRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use DeletePathRequest() / DeletePathRequest.new instead')
  static DeletePathRequest create() => DeletePathRequest._();
  static $pb.GeneratedMessage $_createMessage() => DeletePathRequest._();
  @$core.override
  DeletePathRequest createEmptyInstance() => DeletePathRequest._();
  @$core.pragma('dart2js:noInline')
  static DeletePathRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<DeletePathRequest>(
          DeletePathRequest.$_createMessage);
  static DeletePathRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get path => $_getSZ(0);
  @$pb.TagNumber(1)
  set path($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasPath() => $_has(0);
  @$pb.TagNumber(1)
  void clearPath() => $_clearField(1);
}

class DeletePathResponse extends $pb.GeneratedMessage {
  factory DeletePathResponse({
    $core.String? message,
  }) {
    final result = DeletePathResponse._();
    if (message != null) result.message = message;
    return result;
  }

  DeletePathResponse._();

  factory DeletePathResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeletePathResponse()..mergeFromBuffer(data, registry);
  factory DeletePathResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeletePathResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeletePathResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'fileshare.v1'),
      createEmptyInstance: DeletePathResponse.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'message')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeletePathResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeletePathResponse copyWith(void Function(DeletePathResponse) updates) =>
      super.copyWith((message) => updates(message as DeletePathResponse))
          as DeletePathResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use DeletePathResponse() / DeletePathResponse.new instead')
  static DeletePathResponse create() => DeletePathResponse._();
  static $pb.GeneratedMessage $_createMessage() => DeletePathResponse._();
  @$core.override
  DeletePathResponse createEmptyInstance() => DeletePathResponse._();
  @$core.pragma('dart2js:noInline')
  static DeletePathResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeletePathResponse>(
          DeletePathResponse.$_createMessage);
  static DeletePathResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get message => $_getSZ(0);
  @$pb.TagNumber(1)
  set message($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasMessage() => $_has(0);
  @$pb.TagNumber(1)
  void clearMessage() => $_clearField(1);
}

class UpdateFileContentRequest extends $pb.GeneratedMessage {
  factory UpdateFileContentRequest({
    $core.String? path,
    $core.String? content,
  }) {
    final result = UpdateFileContentRequest._();
    if (path != null) result.path = path;
    if (content != null) result.content = content;
    return result;
  }

  UpdateFileContentRequest._();

  factory UpdateFileContentRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateFileContentRequest()..mergeFromBuffer(data, registry);
  factory UpdateFileContentRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateFileContentRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateFileContentRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'fileshare.v1'),
      createEmptyInstance: UpdateFileContentRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'path')
    ..aOS(2, _omitFieldNames ? '' : 'content')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateFileContentRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateFileContentRequest copyWith(
          void Function(UpdateFileContentRequest) updates) =>
      super.copyWith((message) => updates(message as UpdateFileContentRequest))
          as UpdateFileContentRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateFileContentRequest() / UpdateFileContentRequest.new instead')
  static UpdateFileContentRequest create() => UpdateFileContentRequest._();
  static $pb.GeneratedMessage $_createMessage() => UpdateFileContentRequest._();
  @$core.override
  UpdateFileContentRequest createEmptyInstance() =>
      UpdateFileContentRequest._();
  @$core.pragma('dart2js:noInline')
  static UpdateFileContentRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateFileContentRequest>(
          UpdateFileContentRequest.$_createMessage);
  static UpdateFileContentRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get path => $_getSZ(0);
  @$pb.TagNumber(1)
  set path($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasPath() => $_has(0);
  @$pb.TagNumber(1)
  void clearPath() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get content => $_getSZ(1);
  @$pb.TagNumber(2)
  set content($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasContent() => $_has(1);
  @$pb.TagNumber(2)
  void clearContent() => $_clearField(2);
}

class UpdateFileContentResponse extends $pb.GeneratedMessage {
  factory UpdateFileContentResponse({
    $core.String? message,
    $fixnum.Int64? size,
  }) {
    final result = UpdateFileContentResponse._();
    if (message != null) result.message = message;
    if (size != null) result.size = size;
    return result;
  }

  UpdateFileContentResponse._();

  factory UpdateFileContentResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateFileContentResponse()..mergeFromBuffer(data, registry);
  factory UpdateFileContentResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateFileContentResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateFileContentResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'fileshare.v1'),
      createEmptyInstance: UpdateFileContentResponse.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'message')
    ..aInt64(2, _omitFieldNames ? '' : 'size')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateFileContentResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateFileContentResponse copyWith(
          void Function(UpdateFileContentResponse) updates) =>
      super.copyWith((message) => updates(message as UpdateFileContentResponse))
          as UpdateFileContentResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateFileContentResponse() / UpdateFileContentResponse.new instead')
  static UpdateFileContentResponse create() => UpdateFileContentResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      UpdateFileContentResponse._();
  @$core.override
  UpdateFileContentResponse createEmptyInstance() =>
      UpdateFileContentResponse._();
  @$core.pragma('dart2js:noInline')
  static UpdateFileContentResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateFileContentResponse>(
          UpdateFileContentResponse.$_createMessage);
  static UpdateFileContentResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get message => $_getSZ(0);
  @$pb.TagNumber(1)
  set message($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasMessage() => $_has(0);
  @$pb.TagNumber(1)
  void clearMessage() => $_clearField(1);

  @$pb.TagNumber(2)
  $fixnum.Int64 get size => $_getI64(1);
  @$pb.TagNumber(2)
  set size($fixnum.Int64 value) => $_setInt64(1, value);
  @$pb.TagNumber(2)
  $core.bool hasSize() => $_has(1);
  @$pb.TagNumber(2)
  void clearSize() => $_clearField(2);
}

class UploadMetadata extends $pb.GeneratedMessage {
  factory UploadMetadata({
    $core.String? path,
    $core.String? filename,
  }) {
    final result = UploadMetadata._();
    if (path != null) result.path = path;
    if (filename != null) result.filename = filename;
    return result;
  }

  UploadMetadata._();

  factory UploadMetadata.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UploadMetadata()..mergeFromBuffer(data, registry);
  factory UploadMetadata.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UploadMetadata()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UploadMetadata',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'fileshare.v1'),
      createEmptyInstance: UploadMetadata.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'path')
    ..aOS(2, _omitFieldNames ? '' : 'filename')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UploadMetadata clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UploadMetadata copyWith(void Function(UploadMetadata) updates) =>
      super.copyWith((message) => updates(message as UploadMetadata))
          as UploadMetadata;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use UploadMetadata() / UploadMetadata.new instead')
  static UploadMetadata create() => UploadMetadata._();
  static $pb.GeneratedMessage $_createMessage() => UploadMetadata._();
  @$core.override
  UploadMetadata createEmptyInstance() => UploadMetadata._();
  @$core.pragma('dart2js:noInline')
  static UploadMetadata getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<UploadMetadata>(
          UploadMetadata.$_createMessage);
  static UploadMetadata? _defaultInstance;

  /// Destination directory (virtual, namespace-scoped).
  @$pb.TagNumber(1)
  $core.String get path => $_getSZ(0);
  @$pb.TagNumber(1)
  set path($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasPath() => $_has(0);
  @$pb.TagNumber(1)
  void clearPath() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get filename => $_getSZ(1);
  @$pb.TagNumber(2)
  set filename($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasFilename() => $_has(1);
  @$pb.TagNumber(2)
  void clearFilename() => $_clearField(2);
}

enum UploadRequest_Payload { metadata, chunk, notSet }

class UploadRequest extends $pb.GeneratedMessage {
  factory UploadRequest({
    UploadMetadata? metadata,
    $core.List<$core.int>? chunk,
  }) {
    final result = UploadRequest._();
    if (metadata != null) result.metadata = metadata;
    if (chunk != null) result.chunk = chunk;
    return result;
  }

  UploadRequest._();

  factory UploadRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UploadRequest()..mergeFromBuffer(data, registry);
  factory UploadRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UploadRequest()..mergeFromJson(json, registry);

  static const $core.Map<$core.int, UploadRequest_Payload>
      _UploadRequest_PayloadByTag = {
    1: UploadRequest_Payload.metadata,
    2: UploadRequest_Payload.chunk,
    0: UploadRequest_Payload.notSet
  };
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UploadRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'fileshare.v1'),
      createEmptyInstance: UploadRequest.$_createMessage)
    ..oo(0, [1, 2])
    ..aOM<UploadMetadata>(1, _omitFieldNames ? '' : 'metadata',
        subBuilder: UploadMetadata.$_createMessage)
    ..a<$core.List<$core.int>>(
        2, _omitFieldNames ? '' : 'chunk', $pb.PbFieldType.OY)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UploadRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UploadRequest copyWith(void Function(UploadRequest) updates) =>
      super.copyWith((message) => updates(message as UploadRequest))
          as UploadRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use UploadRequest() / UploadRequest.new instead')
  static UploadRequest create() => UploadRequest._();
  static $pb.GeneratedMessage $_createMessage() => UploadRequest._();
  @$core.override
  UploadRequest createEmptyInstance() => UploadRequest._();
  @$core.pragma('dart2js:noInline')
  static UploadRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<UploadRequest>(
          UploadRequest.$_createMessage);
  static UploadRequest? _defaultInstance;

  @$pb.TagNumber(1)
  @$pb.TagNumber(2)
  UploadRequest_Payload whichPayload() =>
      _UploadRequest_PayloadByTag[$_whichOneof(0)]!;
  @$pb.TagNumber(1)
  @$pb.TagNumber(2)
  void clearPayload() => $_clearField($_whichOneof(0));

  @$pb.TagNumber(1)
  UploadMetadata get metadata => $_getN(0);
  @$pb.TagNumber(1)
  set metadata(UploadMetadata value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasMetadata() => $_has(0);
  @$pb.TagNumber(1)
  void clearMetadata() => $_clearField(1);
  @$pb.TagNumber(1)
  UploadMetadata ensureMetadata() => $_ensure(0);

  @$pb.TagNumber(2)
  $core.List<$core.int> get chunk => $_getN(1);
  @$pb.TagNumber(2)
  set chunk($core.List<$core.int> value) => $_setBytes(1, value);
  @$pb.TagNumber(2)
  $core.bool hasChunk() => $_has(1);
  @$pb.TagNumber(2)
  void clearChunk() => $_clearField(2);
}

class UploadResponse extends $pb.GeneratedMessage {
  factory UploadResponse({
    $core.String? path,
    $fixnum.Int64? size,
  }) {
    final result = UploadResponse._();
    if (path != null) result.path = path;
    if (size != null) result.size = size;
    return result;
  }

  UploadResponse._();

  factory UploadResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UploadResponse()..mergeFromBuffer(data, registry);
  factory UploadResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UploadResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UploadResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'fileshare.v1'),
      createEmptyInstance: UploadResponse.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'path')
    ..aInt64(2, _omitFieldNames ? '' : 'size')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UploadResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UploadResponse copyWith(void Function(UploadResponse) updates) =>
      super.copyWith((message) => updates(message as UploadResponse))
          as UploadResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use UploadResponse() / UploadResponse.new instead')
  static UploadResponse create() => UploadResponse._();
  static $pb.GeneratedMessage $_createMessage() => UploadResponse._();
  @$core.override
  UploadResponse createEmptyInstance() => UploadResponse._();
  @$core.pragma('dart2js:noInline')
  static UploadResponse getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<UploadResponse>(
          UploadResponse.$_createMessage);
  static UploadResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get path => $_getSZ(0);
  @$pb.TagNumber(1)
  set path($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasPath() => $_has(0);
  @$pb.TagNumber(1)
  void clearPath() => $_clearField(1);

  @$pb.TagNumber(2)
  $fixnum.Int64 get size => $_getI64(1);
  @$pb.TagNumber(2)
  set size($fixnum.Int64 value) => $_setInt64(1, value);
  @$pb.TagNumber(2)
  $core.bool hasSize() => $_has(1);
  @$pb.TagNumber(2)
  void clearSize() => $_clearField(2);
}

class DownloadFileRequest extends $pb.GeneratedMessage {
  factory DownloadFileRequest({
    $core.String? path,
  }) {
    final result = DownloadFileRequest._();
    if (path != null) result.path = path;
    return result;
  }

  DownloadFileRequest._();

  factory DownloadFileRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DownloadFileRequest()..mergeFromBuffer(data, registry);
  factory DownloadFileRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DownloadFileRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DownloadFileRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'fileshare.v1'),
      createEmptyInstance: DownloadFileRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'path')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DownloadFileRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DownloadFileRequest copyWith(void Function(DownloadFileRequest) updates) =>
      super.copyWith((message) => updates(message as DownloadFileRequest))
          as DownloadFileRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use DownloadFileRequest() / DownloadFileRequest.new instead')
  static DownloadFileRequest create() => DownloadFileRequest._();
  static $pb.GeneratedMessage $_createMessage() => DownloadFileRequest._();
  @$core.override
  DownloadFileRequest createEmptyInstance() => DownloadFileRequest._();
  @$core.pragma('dart2js:noInline')
  static DownloadFileRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DownloadFileRequest>(
          DownloadFileRequest.$_createMessage);
  static DownloadFileRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get path => $_getSZ(0);
  @$pb.TagNumber(1)
  set path($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasPath() => $_has(0);
  @$pb.TagNumber(1)
  void clearPath() => $_clearField(1);
}

class DownloadChunk extends $pb.GeneratedMessage {
  factory DownloadChunk({
    $core.String? filename,
    $core.String? contentType,
    $core.List<$core.int>? data,
  }) {
    final result = DownloadChunk._();
    if (filename != null) result.filename = filename;
    if (contentType != null) result.contentType = contentType;
    if (data != null) result.data = data;
    return result;
  }

  DownloadChunk._();

  factory DownloadChunk.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DownloadChunk()..mergeFromBuffer(data, registry);
  factory DownloadChunk.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DownloadChunk()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DownloadChunk',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'fileshare.v1'),
      createEmptyInstance: DownloadChunk.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'filename')
    ..aOS(2, _omitFieldNames ? '' : 'contentType')
    ..a<$core.List<$core.int>>(
        3, _omitFieldNames ? '' : 'data', $pb.PbFieldType.OY)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DownloadChunk clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DownloadChunk copyWith(void Function(DownloadChunk) updates) =>
      super.copyWith((message) => updates(message as DownloadChunk))
          as DownloadChunk;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use DownloadChunk() / DownloadChunk.new instead')
  static DownloadChunk create() => DownloadChunk._();
  static $pb.GeneratedMessage $_createMessage() => DownloadChunk._();
  @$core.override
  DownloadChunk createEmptyInstance() => DownloadChunk._();
  @$core.pragma('dart2js:noInline')
  static DownloadChunk getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<DownloadChunk>(
          DownloadChunk.$_createMessage);
  static DownloadChunk? _defaultInstance;

  /// Set on the first chunk only.
  @$pb.TagNumber(1)
  $core.String get filename => $_getSZ(0);
  @$pb.TagNumber(1)
  set filename($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasFilename() => $_has(0);
  @$pb.TagNumber(1)
  void clearFilename() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get contentType => $_getSZ(1);
  @$pb.TagNumber(2)
  set contentType($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasContentType() => $_has(1);
  @$pb.TagNumber(2)
  void clearContentType() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.List<$core.int> get data => $_getN(2);
  @$pb.TagNumber(3)
  set data($core.List<$core.int> value) => $_setBytes(2, value);
  @$pb.TagNumber(3)
  $core.bool hasData() => $_has(2);
  @$pb.TagNumber(3)
  void clearData() => $_clearField(3);
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
